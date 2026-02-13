package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/metrics"
	"github.com/j3ssie/osmedeus/v5/internal/runner"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// BashExecutor executes bash steps
type BashExecutor struct {
	templateEngine template.TemplateEngine
	runner         runner.Runner
}

// NewBashExecutor creates a new bash executor
func NewBashExecutor(engine template.TemplateEngine) *BashExecutor {
	return &BashExecutor{
		templateEngine: engine,
	}
}

// Name returns the executor name for logging/debugging
func (e *BashExecutor) Name() string {
	return "bash"
}

// StepTypes returns the step types this executor handles
func (e *BashExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeBash}
}

// SetRunner sets the runner for command execution
func (e *BashExecutor) SetRunner(r runner.Runner) {
	e.runner = r
}

// assembleCommand joins the command with structured args in order:
// command + speed_args + config_args + input_args + output_args
func assembleCommand(command, speedArgs, configArgs, inputArgs, outputArgs string) string {
	parts := []string{command}
	if speedArgs != "" {
		parts = append(parts, speedArgs)
	}
	if configArgs != "" {
		parts = append(parts, configArgs)
	}
	if inputArgs != "" {
		parts = append(parts, inputArgs)
	}
	if outputArgs != "" {
		parts = append(parts, outputArgs)
	}
	return strings.Join(parts, " ")
}

// writeStdFile writes command output to the specified file
func writeStdFile(path, content string) error {
	// Ensure parent directory exists
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// extractToolName extracts the tool/binary name from a command string.
// Returns the base name of the first word in the command.
func extractToolName(command string) string {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "unknown"
	}
	return filepath.Base(parts[0])
}

// Execute executes a bash step
func (e *BashExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
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

	var output string

	// Determine execution mode
	if len(step.ParallelCommands) > 0 {
		output, err = e.executeParallel(ctx, step.ParallelCommands, timeout)
	} else if len(step.Commands) > 0 {
		output, err = e.executeSequential(ctx, step.Commands, timeout)
	} else if step.Command != "" {
		// Assemble command with structured args if present
		finalCmd := assembleCommand(step.Command, step.SpeedArgs, step.ConfigArgs, step.InputArgs, step.OutputArgs)
		output, err = e.executeCommand(ctx, finalCmd, timeout)
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
	return result, nil
}

// executeCommand executes a single command

func (e *BashExecutor) executeCommand(ctx context.Context, command string, timeout time.Duration) (string, error) {
	// Track execution timing for metrics
	startTime := time.Now()
	toolName := extractToolName(command)

	// Apply timeout if specified
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Use runner if available, otherwise fall back to local execution
	if e.runner != nil {
		result, err := e.runner.Execute(ctx, command)
		duration := time.Since(startTime).Seconds()

		// Check for context cancellation first (Ctrl+C or timeout)
		// This must be checked before other errors because runners return nil error
		// but set result.ExitCode to -1 when context is cancelled
		if ctx.Err() != nil {
			if ctx.Err() == context.DeadlineExceeded {
				metrics.RecordToolExecution(toolName, "timeout", duration)
				return result.Output, fmt.Errorf("command timed out after %s", timeout)
			}
			metrics.RecordToolExecution(toolName, "cancelled", duration)
			return result.Output, ctx.Err()
		}

		if err != nil {
			metrics.RecordToolExecution(toolName, "error", duration)
			return result.Output, fmt.Errorf("command failed: %w", err)
		}
		if result.ExitCode != 0 {
			metrics.RecordToolExecution(toolName, "failed", duration)
			return result.Output, fmt.Errorf("command exited with code %d", result.ExitCode)
		}
		metrics.RecordToolExecution(toolName, "success", duration)
		return strings.TrimSpace(result.Output), nil
	}

	// Fallback to local execution
	// @NOTE: yes yes, I know this is a security risk. This is the intended behavior as it suppose to be that flexible so you should think twice of what you will run or design your workflow better
	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	stdout := runner.NewLimitedBuffer(runner.MaxOutputSize)
	stderr := runner.NewLimitedBuffer(runner.MaxStderrSize)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	duration := time.Since(startTime).Seconds()

	output := string(stdout.Bytes())
	if stderr.Len() > 0 {
		output += "\n" + string(stderr.Bytes())
	}
	if stdout.Overflow() || stderr.Overflow() {
		output += "\n[output truncated]"
	}

	// Check for context cancellation first (Ctrl+C or timeout)
	if ctx.Err() != nil {
		if ctx.Err() == context.DeadlineExceeded {
			metrics.RecordToolExecution(toolName, "timeout", duration)
			return output, fmt.Errorf("command timed out after %s", timeout)
		}
		metrics.RecordToolExecution(toolName, "cancelled", duration)
		return output, ctx.Err()
	}

	if err != nil {
		metrics.RecordToolExecution(toolName, "error", duration)
		return output, fmt.Errorf("command failed: %w\nstderr: %s", err, string(stderr.Bytes()))
	}

	metrics.RecordToolExecution(toolName, "success", duration)
	return strings.TrimSpace(output), nil
}

// executeSequential executes commands sequentially
func (e *BashExecutor) executeSequential(ctx context.Context, commands []string, timeout time.Duration) (string, error) {
	var outputs []string

	for _, cmd := range commands {
		output, err := e.executeCommand(ctx, cmd, timeout)
		outputs = append(outputs, output)
		if err != nil {
			return strings.Join(outputs, "\n"), err
		}
	}

	return strings.Join(outputs, "\n"), nil
}

// executeParallel executes commands in parallel with bounded concurrency.
// Uses a worker pool capped at runtime.NumCPU()*2 to prevent unbounded goroutine/memory growth.
func (e *BashExecutor) executeParallel(ctx context.Context, commands []string, timeout time.Duration) (string, error) {
	type result struct {
		index  int
		output string
		err    error
	}

	maxWorkers := runtime.NumCPU() * 2
	if maxWorkers > len(commands) {
		maxWorkers = len(commands)
	}

	type workItem struct {
		index   int
		command string
	}

	workQueue := make(chan workItem, maxWorkers)
	results := make(chan result, len(commands))
	var wg sync.WaitGroup

	// Start bounded worker pool
	for range maxWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workQueue {
				output, err := e.executeCommand(ctx, work.command, timeout)
				results <- result{index: work.index, output: output, err: err}
			}
		}()
	}

	// Feed work items
	go func() {
		for i, cmd := range commands {
			workQueue <- workItem{index: i, command: cmd}
		}
		close(workQueue)
	}()

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	outputs := make([]string, len(commands))
	var firstError error

	for r := range results {
		outputs[r.index] = r.output
		if r.err != nil && firstError == nil {
			firstError = r.err
		}
	}

	return strings.Join(outputs, "\n"), firstError
}

// CanHandle returns true if this executor can handle the given step type
func (e *BashExecutor) CanHandle(stepType core.StepType) bool {
	return stepType == core.StepTypeBash
}
