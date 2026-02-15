package executor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/runner"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// RemoteBashExecutor executes remote-bash steps on Docker/SSH runners
type RemoteBashExecutor struct {
	templateEngine template.TemplateEngine
}

// NewRemoteBashExecutor creates a new remote bash executor
func NewRemoteBashExecutor(engine template.TemplateEngine) *RemoteBashExecutor {
	return &RemoteBashExecutor{
		templateEngine: engine,
	}
}

// Name returns the executor name for logging/debugging
func (e *RemoteBashExecutor) Name() string {
	return "remote-bash"
}

// StepTypes returns the step types this executor handles
func (e *RemoteBashExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeRemoteBash}
}

// Execute executes a remote-bash step
func (e *RemoteBashExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
	}

	timeout, err := step.Timeout.Duration()
	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Validate step_runner is set for remote-bash
	if step.StepRunner == "" || step.StepRunner == core.RunnerTypeHost {
		err := fmt.Errorf("remote-bash step '%s' requires step_runner to be 'docker' or 'ssh'", step.Name)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Create runner based on step_runner and step_runner_config
	r, err := e.createRunner(step.StepRunner, step.StepRunnerConfig)
	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Setup the runner (fresh connection for each step)
	if err := r.Setup(ctx); err != nil {
		result.Status = core.StepStatusFailed
		result.Error = fmt.Errorf("runner setup failed: %w", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, result.Error
	}

	// Ensure cleanup happens
	defer func() {
		cleanupCtx := context.Background() // Use fresh context for cleanup
		_ = r.Cleanup(cleanupCtx)
	}()

	// Extract binaries path for fallback resolution
	binariesPath := ""
	if bp, ok := execCtx.GetVariable("Binaries"); ok {
		if bpStr, ok := bp.(string); ok {
			binariesPath = bpStr
		}
	}

	// Execute command(s) using the runner
	var output string
	if len(step.ParallelCommands) > 0 {
		output, err = e.executeParallel(ctx, r, step.ParallelCommands, timeout, binariesPath)
	} else if len(step.Commands) > 0 {
		output, err = e.executeSequential(ctx, r, step.Commands, timeout, binariesPath)
	} else if step.Command != "" {
		// Assemble command with structured args if present
		finalCmd := assembleCommand(step.Command, step.SpeedArgs, step.ConfigArgs, step.InputArgs, step.OutputArgs)
		output, err = e.executeCommandWithFallback(ctx, r, finalCmd, timeout, binariesPath)
	} else {
		err = fmt.Errorf("no command specified")
	}

	result.Output = output
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Write stdout/stderr to file if std_file is specified
	if step.StdFile != "" {
		if writeErr := writeStdFile(step.StdFile, output); writeErr != nil {
			// Log warning but don't fail the step
			execCtx.Logger.Warn("Failed to write std_file",
				zap.String("path", step.StdFile),
				zap.Error(writeErr))
		}
	}

	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		return result, err
	}

	result.Status = core.StepStatusSuccess

	// Copy remote file to host if specified (before cleanup)
	if step.StepRemoteFile != "" && step.HostOutputFile != "" {
		if copyErr := r.CopyFromRemote(ctx, step.StepRemoteFile, step.HostOutputFile); copyErr != nil {
			// Log warning but don't fail the step
			execCtx.Logger.Warn("Failed to copy remote file",
				zap.String("remote", step.StepRemoteFile),
				zap.String("local", step.HostOutputFile),
				zap.Error(copyErr))
		} else {
			execCtx.Logger.Debug("Copied remote file to host",
				zap.String("remote", step.StepRemoteFile),
				zap.String("local", step.HostOutputFile))
		}
	}

	return result, nil
}

// createRunner creates a runner based on step_runner type and step_runner_config
func (e *RemoteBashExecutor) createRunner(runnerType core.RunnerType, cfg *core.StepRunnerConfig) (runner.Runner, error) {
	// Get the embedded RunnerConfig (or create empty one)
	runnerCfg := &core.RunnerConfig{}
	if cfg != nil && cfg.RunnerConfig != nil {
		runnerCfg = cfg.RunnerConfig
	}

	// Pass empty string for binaryPath since we only execute shell commands,
	// not the osmedeus binary itself
	switch runnerType {
	case core.RunnerTypeDocker:
		return runner.NewDockerRunner(runnerCfg, "")
	case core.RunnerTypeSSH:
		return runner.NewSSHRunner(runnerCfg, "")
	default:
		return nil, fmt.Errorf("unsupported step_runner for remote-bash: %s (must be 'docker' or 'ssh')", runnerType)
	}
}

// executeCommand executes a single command on the remote runner
func (e *RemoteBashExecutor) executeCommand(ctx context.Context, r runner.Runner, command string, timeout time.Duration) (string, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	cmdResult, err := r.Execute(ctx, command)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			output := ""
			if cmdResult != nil {
				output = cmdResult.Output
			}
			return output, fmt.Errorf("command timed out after %s", timeout)
		}
		output := ""
		if cmdResult != nil {
			output = cmdResult.Output
		}
		return output, fmt.Errorf("command failed: %w", err)
	}

	if cmdResult.ExitCode != 0 {
		return cmdResult.Output, newExitCodeErrorf(cmdResult.ExitCode, "command exited with code %d", cmdResult.ExitCode)
	}

	return strings.TrimSpace(cmdResult.Output), nil
}

// executeCommandWithFallback wraps executeCommand with automatic retry on exit code 127.
// Fallback 1: strip timeout prefix. Fallback 2: prepend binariesPath to the binary.
func (e *RemoteBashExecutor) executeCommandWithFallback(ctx context.Context, r runner.Runner, command string, timeout time.Duration, binariesPath string) (string, error) {
	output, err := e.executeCommand(ctx, r, command, timeout)
	if err == nil {
		return output, nil
	}

	// Only attempt fallback on exit code 127 (command not found)
	var ecErr *exitCodeError
	if !errors.As(err, &ecErr) || ecErr.code != 127 {
		return output, err
	}

	// Don't retry if context is already cancelled
	if ctx.Err() != nil {
		return output, err
	}

	currentCmd := command

	// Fallback 1: strip timeout prefix
	if result := stripTimeoutPrefix(currentCmd); result.stripped && result.command != "" {
		// Use parsed duration from timeout prefix as fallback if step timeout is not set
		retryTimeout := timeout
		if retryTimeout == 0 && result.duration > 0 {
			retryTimeout = result.duration
		}
		output, err = e.executeCommand(ctx, r, result.command, retryTimeout)
		if err == nil {
			return output, nil
		}
		if !errors.As(err, &ecErr) || ecErr.code != 127 {
			return output, err
		}
		currentCmd = result.command
	}

	// Fallback 2: prepend binaries path
	if prepended, ok := prependBinariesPath(currentCmd, binariesPath); ok {
		output, err = e.executeCommand(ctx, r, prepended, timeout)
		return output, err
	}

	return output, err
}

// executeSequential executes commands sequentially
func (e *RemoteBashExecutor) executeSequential(ctx context.Context, r runner.Runner, commands []string, timeout time.Duration, binariesPath string) (string, error) {
	var outputs []string

	for _, cmd := range commands {
		output, err := e.executeCommandWithFallback(ctx, r, cmd, timeout, binariesPath)
		outputs = append(outputs, output)
		if err != nil {
			return strings.Join(outputs, "\n"), err
		}
	}

	return strings.Join(outputs, "\n"), nil
}

// executeParallel executes commands in parallel
func (e *RemoteBashExecutor) executeParallel(ctx context.Context, r runner.Runner, commands []string, timeout time.Duration, binariesPath string) (string, error) {
	type cmdResult struct {
		index  int
		output string
		err    error
	}

	results := make(chan cmdResult, len(commands))
	var wg sync.WaitGroup

	for i, cmd := range commands {
		wg.Add(1)
		go func(idx int, command string) {
			defer wg.Done()
			output, err := e.executeCommandWithFallback(ctx, r, command, timeout, binariesPath)
			results <- cmdResult{index: idx, output: output, err: err}
		}(i, cmd)
	}

	// Wait for all commands to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	outputs := make([]string, len(commands))
	var firstError error

	for res := range results {
		outputs[res.index] = res.output
		if res.err != nil && firstError == nil {
			firstError = res.err
		}
	}

	return strings.Join(outputs, "\n"), firstError
}

// CanHandle returns true if this executor can handle the given step type
func (e *RemoteBashExecutor) CanHandle(stepType core.StepType) bool {
	return stepType == core.StepTypeRemoteBash
}
