package runner

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// MaxOutputSize limits the combined output size to prevent memory issues
// with very large command outputs (10MB default)
const MaxOutputSize = 10 * 1024 * 1024

// MaxStderrSize limits stderr capture (1MB) since stderr is typically small
const MaxStderrSize = 1 * 1024 * 1024

// LimitedBuffer wraps bytes.Buffer and silently discards writes after maxSize.
// This prevents unbounded memory growth when a tool produces GBs of output.
type LimitedBuffer struct {
	buf      []byte
	maxSize  int
	overflow bool
	mu       sync.Mutex
}

// NewLimitedBuffer creates a LimitedBuffer with the given max size.
func NewLimitedBuffer(maxSize int) *LimitedBuffer {
	return &LimitedBuffer{
		maxSize: maxSize,
	}
}

// Write implements io.Writer. Writes beyond maxSize are silently discarded.
func (lb *LimitedBuffer) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.buf) >= lb.maxSize {
		lb.overflow = true
		return len(p), nil // pretend we wrote it all
	}

	remaining := lb.maxSize - len(lb.buf)
	if len(p) > remaining {
		lb.buf = append(lb.buf, p[:remaining]...)
		lb.overflow = true
		return len(p), nil
	}

	lb.buf = append(lb.buf, p...)
	return len(p), nil
}

// Bytes returns the buffered bytes.
func (lb *LimitedBuffer) Bytes() []byte {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.buf
}

// Len returns the number of buffered bytes.
func (lb *LimitedBuffer) Len() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return len(lb.buf)
}

// Overflow returns true if any writes were discarded.
func (lb *LimitedBuffer) Overflow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.overflow
}

// combineOutput efficiently combines stdout and stderr from LimitedBuffers.
// Appends "[output truncated]" if either buffer overflowed.
func combineOutput(stdout, stderr *LimitedBuffer) string {
	stdoutBytes := stdout.Bytes()
	stderrBytes := stderr.Bytes()
	totalLen := len(stdoutBytes) + len(stderrBytes)

	if totalLen == 0 {
		return ""
	}

	truncated := stdout.Overflow() || stderr.Overflow()

	var sb strings.Builder
	if truncated {
		sb.Grow(totalLen + 20)
	} else {
		sb.Grow(totalLen)
	}
	sb.Write(stdoutBytes)
	sb.Write(stderrBytes)
	if truncated {
		sb.WriteString("\n[output truncated]")
	}
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
