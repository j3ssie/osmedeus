package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// HostRunner executes commands on the local machine
type HostRunner struct {
	binariesPath string
	onPIDStart   PIDCallback
	onPIDEnd     PIDCallback
}

// NewHostRunner creates a new host runner
func NewHostRunner(binariesPath string) *HostRunner {
	return &HostRunner{binariesPath: binariesPath}
}

// Execute runs a command on the local machine
func (r *HostRunner) Execute(ctx context.Context, command string) (*CommandResult, error) {
	// @NOTE: This is intentional - the workflow engine is designed to execute arbitrary
	// commands from YAML workflow definitions. The command input comes from trusted
	// workflow files, not from untrusted user input.
	cmd := exec.Command("sh", "-c", command)

	// Create new process group so we can kill all children on interrupt
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Prepend binaries path to PATH if configured
	if r.binariesPath != "" {
		env := os.Environ()
		pathUpdated := false
		for i, e := range env {
			if strings.HasPrefix(e, "PATH=") {
				env[i] = "PATH=" + r.binariesPath + ":" + strings.TrimPrefix(e, "PATH=")
				pathUpdated = true
				break
			}
		}
		if !pathUpdated {
			env = append(env, "PATH="+r.binariesPath)
		}
		cmd.Env = env
	}

	stdout := NewLimitedBuffer(MaxOutputSize)
	stderr := NewLimitedBuffer(MaxStderrSize)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return &CommandResult{
			Output:   "",
			ExitCode: -1,
			Error:    err,
		}, nil
	}

	// Track PID for cancellation support
	pid := cmd.Process.Pid
	if r.onPIDStart != nil {
		r.onPIDStart(pid)
	}

	// Wait for command completion or context cancellation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// Context cancelled - kill entire process group
		if cmd.Process != nil {
			// Kill process group (negative PID)
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		<-done // Wait for process to exit

		// Notify PID ended
		if r.onPIDEnd != nil {
			r.onPIDEnd(pid)
		}

		return &CommandResult{
			Output:   combineOutput(stdout, stderr),
			ExitCode: -1,
			Error:    ctx.Err(),
		}, nil

	case err := <-done:
		// Normal completion - notify PID ended
		if r.onPIDEnd != nil {
			r.onPIDEnd(pid)
		}

		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		return &CommandResult{
			Output:   combineOutput(stdout, stderr),
			ExitCode: exitCode,
			Error:    err,
		}, nil
	}
}

// SetPIDCallbacks sets callbacks for process lifecycle events
func (r *HostRunner) SetPIDCallbacks(onStart, onEnd PIDCallback) {
	r.onPIDStart = onStart
	r.onPIDEnd = onEnd
}

// Setup is a no-op for host runner
func (r *HostRunner) Setup(ctx context.Context) error {
	return nil
}

// Cleanup is a no-op for host runner
func (r *HostRunner) Cleanup(ctx context.Context) error {
	return nil
}

// CopyFromRemote is not supported for host runner (local execution doesn't need file copy)
func (r *HostRunner) CopyFromRemote(ctx context.Context, remotePath, localPath string) error {
	return fmt.Errorf("CopyFromRemote not supported for host runner")
}

// Type returns the runner type
func (r *HostRunner) Type() core.RunnerType {
	return core.RunnerTypeHost
}

// IsRemote returns false since host runner runs locally
func (r *HostRunner) IsRemote() bool {
	return false
}
