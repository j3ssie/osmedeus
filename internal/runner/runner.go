package runner

import (
	"context"
	"fmt"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// CommandResult holds the output of a command execution
type CommandResult struct {
	Output   string // Combined stdout and stderr
	ExitCode int    // Exit code of the command
	Error    error  // Error if execution failed
}

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
