package runner

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// MaxOutputSize limits the combined output size to prevent memory issues
// with very large command outputs (10MB default)
const MaxOutputSize = 10 * 1024 * 1024

// combineOutput efficiently combines stdout and stderr using a single allocation.
// This reduces memory allocations from 3 to 1 compared to stdout.String() + stderr.String().
// Outputs exceeding MaxOutputSize are truncated with a warning message.
func combineOutput(stdout, stderr *bytes.Buffer) string {
	totalLen := stdout.Len() + stderr.Len()
	if totalLen == 0 {
		return ""
	}

	if totalLen > MaxOutputSize {
		// Truncate with message
		var sb strings.Builder
		sb.Grow(MaxOutputSize + 30)
		limit := min(MaxOutputSize, stdout.Len())
		sb.Write(stdout.Bytes()[:limit])
		if remaining := MaxOutputSize - limit; remaining > 0 && stderr.Len() > 0 {
			sb.Write(stderr.Bytes()[:min(remaining, stderr.Len())])
		}
		sb.WriteString("\n[output truncated]")
		return sb.String()
	}

	var sb strings.Builder
	sb.Grow(totalLen)
	sb.Write(stdout.Bytes())
	sb.Write(stderr.Bytes())
	return sb.String()
}

// CommandResult holds the output of a command execution
type CommandResult struct {
	Output   string // Combined stdout and stderr
	ExitCode int    // Exit code of the command
	Error    error  // Error if execution failed
}

// PIDCallback is called when a process starts or ends
type PIDCallback func(pid int)

// Runner interface for executing commands in different environments
type Runner interface {
	// Execute runs a command and returns the result
	Execute(ctx context.Context, command string) (*CommandResult, error)

	// Setup prepares the runner (e.g., copy binary, start container, establish SSH)
	Setup(ctx context.Context) error

	// Cleanup tears down resources (e.g., stop container, close SSH connection)
	Cleanup(ctx context.Context) error

	// Type returns the runner type
	Type() core.RunnerType

	// IsRemote returns true if commands run on a remote machine
	IsRemote() bool

	// CopyFromRemote copies a file from the remote environment to the local host
	// For Docker: uses docker cp, for SSH: uses rsync
	CopyFromRemote(ctx context.Context, remotePath, localPath string) error

	// SetPIDCallbacks sets callbacks for process lifecycle events.
	// onStart is called when a process starts (with the PID)
	// onEnd is called when the process ends (with the PID)
	// This enables tracking of running processes for cancellation support.
	SetPIDCallbacks(onStart, onEnd PIDCallback)
}

// NewRunner creates a runner based on workflow configuration
func NewRunner(workflow *core.Workflow, binaryPath string) (Runner, error) {
	runnerType := workflow.Runner
	if runnerType == "" {
		runnerType = core.RunnerTypeHost
	}

	config := workflow.RunnerConfig
	if config == nil {
		config = &core.RunnerConfig{}
	}

	switch runnerType {
	case core.RunnerTypeHost:
		return NewHostRunner(binaryPath), nil
	case core.RunnerTypeDocker:
		return NewDockerRunner(config, binaryPath)
	case core.RunnerTypeSSH:
		return NewSSHRunner(config, binaryPath)
	default:
		return nil, fmt.Errorf("unknown runner type: %s", runnerType)
	}
}

// NewRunnerFromType creates a runner from type and config directly
func NewRunnerFromType(runnerType core.RunnerType, config *core.RunnerConfig, binaryPath string) (Runner, error) {
	if config == nil {
		config = &core.RunnerConfig{}
	}

	switch runnerType {
	case core.RunnerTypeHost, "":
		return NewHostRunner(binaryPath), nil
	case core.RunnerTypeDocker:
		return NewDockerRunner(config, binaryPath)
	case core.RunnerTypeSSH:
		return NewSSHRunner(config, binaryPath)
	default:
		return nil, fmt.Errorf("unknown runner type: %s", runnerType)
	}
}
