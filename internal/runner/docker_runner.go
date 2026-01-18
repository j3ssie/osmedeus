package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// DockerRunner executes commands inside a Docker container
type DockerRunner struct {
	config      *core.RunnerConfig
	binaryPath  string
	containerID string
	remoteDir   string
}

// NewDockerRunner creates a new Docker runner
func NewDockerRunner(config *core.RunnerConfig, binaryPath string) (*DockerRunner, error) {
	if config.Image == "" {
		return nil, fmt.Errorf("docker runner requires image to be specified")
	}

	return &DockerRunner{
		config:     config,
		binaryPath: binaryPath,
		remoteDir:  "/root/.osmedeus-remote",
	}, nil
}

// Setup prepares the Docker container
func (r *DockerRunner) Setup(ctx context.Context) error {
	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found in PATH: %w", err)
	}

	// Pull image if needed (ignore errors if image already exists)
	pullCmd := exec.CommandContext(ctx, "docker", "pull", r.config.Image)
	_ = pullCmd.Run() // Ignore error, image might already exist locally

	if r.config.Persistent {
		// Create and start a persistent container
		return r.startPersistentContainer(ctx)
	}

	return nil
}

// startPersistentContainer creates and starts a long-running container
func (r *DockerRunner) startPersistentContainer(ctx context.Context) error {
	args := []string{"run", "-d", "--rm"}

	// Add environment variables
	for k, v := range r.config.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add volume mounts
	for _, vol := range r.config.Volumes {
		args = append(args, "-v", vol)
	}

	// Add network mode
	if r.config.Network != "" {
		args = append(args, "--network", r.config.Network)
	}

	// Add working directory
	if r.config.WorkDir != "" {
		args = append(args, "-w", r.config.WorkDir)
	}

	// Add image and keep-alive command
	args = append(args, r.config.Image, "tail", "-f", "/dev/null")

	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w, stderr: %s", err, stderr.String())
	}

	r.containerID = strings.TrimSpace(stdout.String())
	if r.containerID == "" {
		return fmt.Errorf("failed to get container ID")
	}

	// Copy binary to container if it exists
	if r.binaryPath != "" {
		if err := r.copyBinaryToContainer(ctx); err != nil {
			// Non-fatal, log and continue
			fmt.Printf("Warning: failed to copy binary to container: %v\n", err)
		}
	}

	return nil
}

// copyBinaryToContainer copies the osmedeus binary to the container
func (r *DockerRunner) copyBinaryToContainer(ctx context.Context) error {
	if r.containerID == "" {
		return fmt.Errorf("no container running")
	}

	// Create remote directory
	mkdirCmd := exec.CommandContext(ctx, "docker", "exec", r.containerID, "mkdir", "-p", r.remoteDir)
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Copy binary
	destPath := filepath.Join(r.remoteDir, "osmedeus")
	copyCmd := exec.CommandContext(ctx, "docker", "cp", r.binaryPath, fmt.Sprintf("%s:%s", r.containerID, destPath))
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make it executable
	chmodCmd := exec.CommandContext(ctx, "docker", "exec", r.containerID, "chmod", "+x", destPath)
	if err := chmodCmd.Run(); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	return nil
}

// Execute runs a command in the Docker container
func (r *DockerRunner) Execute(ctx context.Context, command string) (*CommandResult, error) {
	if r.config.Persistent && r.containerID != "" {
		return r.execInContainer(ctx, command)
	}
	return r.runEphemeral(ctx, command)
}

// execInContainer runs a command in an existing container using docker exec
func (r *DockerRunner) execInContainer(ctx context.Context, command string) (*CommandResult, error) {
	args := []string{"exec"}

	// Add working directory if specified
	if r.config.WorkDir != "" {
		args = append(args, "-w", r.config.WorkDir)
	}

	args = append(args, r.containerID, "sh", "-c", command)

	cmd := exec.Command("docker", args...)

	// Create new process group so we can kill all children on interrupt
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		return &CommandResult{
			Output:   "",
			ExitCode: -1,
			Error:    err,
		}, nil
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
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		<-done
		return &CommandResult{
			Output:   stdout.String() + stderr.String(),
			ExitCode: -1,
			Error:    ctx.Err(),
		}, nil

	case err := <-done:
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		return &CommandResult{
			Output:   stdout.String() + stderr.String(),
			ExitCode: exitCode,
			Error:    err,
		}, nil
	}
}

// runEphemeral runs a command in a new ephemeral container (docker run --rm)
func (r *DockerRunner) runEphemeral(ctx context.Context, command string) (*CommandResult, error) {
	args := []string{"run", "--rm"}

	// Add environment variables
	for k, v := range r.config.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add volume mounts
	for _, vol := range r.config.Volumes {
		args = append(args, "-v", vol)
	}

	// Add network mode
	if r.config.Network != "" {
		args = append(args, "--network", r.config.Network)
	}

	// Add working directory
	if r.config.WorkDir != "" {
		args = append(args, "-w", r.config.WorkDir)
	}

	// Add image and command
	args = append(args, r.config.Image, "sh", "-c", command)

	cmd := exec.Command("docker", args...)

	// Create new process group so we can kill all children on interrupt
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment from host if needed
	cmd.Env = os.Environ()

	// Start the command
	if err := cmd.Start(); err != nil {
		return &CommandResult{
			Output:   "",
			ExitCode: -1,
			Error:    err,
		}, nil
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
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		<-done
		return &CommandResult{
			Output:   stdout.String() + stderr.String(),
			ExitCode: -1,
			Error:    ctx.Err(),
		}, nil

	case err := <-done:
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			}
		}
		return &CommandResult{
			Output:   stdout.String() + stderr.String(),
			ExitCode: exitCode,
			Error:    err,
		}, nil
	}
}

// Cleanup stops and removes the container
func (r *DockerRunner) Cleanup(ctx context.Context) error {
	if r.config.Persistent && r.containerID != "" {
		// Stop the container (--rm flag will auto-remove it)
		stopCmd := exec.CommandContext(ctx, "docker", "stop", r.containerID)
		_ = stopCmd.Run() // Ignore error
		r.containerID = ""
	}
	return nil
}

// CopyFromRemote copies a file from the container to the local host using docker cp
func (r *DockerRunner) CopyFromRemote(ctx context.Context, remotePath, localPath string) error {
	if r.containerID == "" {
		return fmt.Errorf("no container ID available for file copy")
	}

	// Ensure parent directory exists
	if dir := filepath.Dir(localPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// docker cp <containerID>:<remotePath> <localPath>
	cmd := exec.CommandContext(ctx, "docker", "cp",
		fmt.Sprintf("%s:%s", r.containerID, remotePath), localPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker cp failed: %w, output: %s", err, string(output))
	}
	return nil
}

// Type returns the runner type
func (r *DockerRunner) Type() core.RunnerType {
	return core.RunnerTypeDocker
}

// IsRemote returns true since Docker container is considered remote
func (r *DockerRunner) IsRemote() bool {
	return true
}

// GetContainerID returns the current container ID (for persistent mode)
func (r *DockerRunner) GetContainerID() string {
	return r.containerID
}
