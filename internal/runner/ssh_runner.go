package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// sshPidCounter generates unique remote pidfile names per command invocation.
var sshPidCounter int64

// SSHRunner executes commands on a remote machine via SSH
type SSHRunner struct {
	config     *core.RunnerConfig
	binaryPath string
	client     *ssh.Client
	remoteDir  string
	pooled     bool       // true if connection is from pool
	poolKey    SSHPoolKey // key for releasing back to pool
	onPIDStart PIDCallback
	onPIDEnd   PIDCallback
}

// NewSSHRunner creates a new SSH runner
func NewSSHRunner(config *core.RunnerConfig, binaryPath string) (*SSHRunner, error) {
	if config.Host == "" {
		return nil, fmt.Errorf("SSH runner requires host to be specified")
	}
	if config.User == "" {
		return nil, fmt.Errorf("SSH runner requires user to be specified")
	}

	return &SSHRunner{
		config:     config,
		binaryPath: binaryPath,
		remoteDir:  "~/.osmedeus-remote",
	}, nil
}

// Setup establishes SSH connection and copies the binary
// Uses connection pooling to reuse SSH connections across steps
func (r *SSHRunner) Setup(ctx context.Context) error {
	// Get connection from pool
	pool := GetSSHPool()
	client, poolKey, err := pool.Get(ctx, r.config)
	if err != nil {
		return err
	}

	r.client = client
	r.pooled = true
	r.poolKey = poolKey

	// Create remote directory
	if _, err := r.runCommand(ctx, "mkdir -p "+r.remoteDir); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Copy binary to remote machine if it exists
	if r.binaryPath != "" {
		if err := r.copyBinary(); err != nil {
			// Non-fatal, log and continue
			fmt.Printf("Warning: failed to copy binary to remote: %v\n", err)
		}
	}

	return nil
}

// expandPath expands ~ to home directory
func (r *SSHRunner) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// copyBinary copies the osmedeus binary to the remote machine using SFTP
func (r *SSHRunner) copyBinary() error {
	if r.client == nil {
		return fmt.Errorf("SSH client not connected")
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(r.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer func() { _ = sftpClient.Close() }()

	// Expand remote directory
	homeDir, err := r.getRemoteHome()
	if err != nil {
		return fmt.Errorf("failed to get remote home: %w", err)
	}
	remotePath := strings.Replace(r.remoteDir, "~", homeDir, 1)

	// Create remote directory
	_ = sftpClient.MkdirAll(remotePath)

	// Open local file
	localFile, err := os.Open(r.binaryPath)
	if err != nil {
		return fmt.Errorf("failed to open local binary: %w", err)
	}
	defer func() { _ = localFile.Close() }()

	// Get local file info for size
	localInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local binary: %w", err)
	}

	// Create remote file
	destPath := filepath.Join(remotePath, "osmedeus")
	remoteFile, err := sftpClient.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer func() { _ = remoteFile.Close() }()

	// Copy file
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Make executable
	if err := sftpClient.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to chmod binary: %w", err)
	}

	fmt.Printf("Copied binary to remote (%d bytes)\n", localInfo.Size())
	return nil
}

// getRemoteHome gets the home directory on the remote machine
func (r *SSHRunner) getRemoteHome() (string, error) {
	// Use background context for internal utility calls
	output, err := r.runCommand(context.Background(), "echo $HOME")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// runCommand executes a command on the remote machine with context support.
// Adds workdir prefix and a `timeout(1)` wrapper when ctx has a deadline, then
// delegates to ExecuteSSHCommand for the cancel-aware exec path.
func (r *SSHRunner) runCommand(ctx context.Context, command string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("SSH client not connected")
	}

	// Set working directory if specified
	if r.config.WorkDir != "" {
		command = fmt.Sprintf("cd %s && %s", r.config.WorkDir, command)
	}

	// Enforce context deadline using timeout command on remote host
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining > 0 {
			secs := int(remaining.Seconds())
			if secs > 0 {
				command = fmt.Sprintf("timeout %ds %s", secs, command)
			}
		}
	}

	return ExecuteSSHCommand(ctx, r.client, command)
}

// ExecuteSSHCommand runs `command` on an existing *ssh.Client and returns
// combined output. When ctx is cancelled mid-flight, a watcher goroutine
// opens a second SSH session and kills the remote process group (TERM then
// KILL). This is the shared cancel-aware code path used by SSHRunner and by
// ad-hoc helpers (e.g. the ssh_exec workflow function).
//
// Why this matters: closing the local SSH session does NOT kill remote
// processes — sshd just detaches them. The pidfile + process-group kill is
// the only reliable way to terminate scanning still running on the remote.
func ExecuteSSHCommand(ctx context.Context, client *ssh.Client, command string) (string, error) {
	if client == nil {
		return "", fmt.Errorf("SSH client not connected")
	}

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer func() { _ = session.Close() }()

	// Pin the pidfile to /tmp: callers (e.g. SSHRunner.Setup) may invoke this
	// before any remote scratch directory exists; /tmp is always writable.
	pidFile := fmt.Sprintf("/tmp/.osm-pid-%d-%d.tmp",
		time.Now().UnixNano(), atomic.AddInt64(&sshPidCounter, 1))

	// No `exec` here: keeping the wrapper sh alive means pipelines (which
	// run in subshells) still inherit the wrapper's process group, so the
	// negative-PID kill in killRemoteProcessGroup reaches every child.
	wrapped := fmt.Sprintf("echo $$ > %s; %s", pidFile, command)

	// Watch for cancellation and kill the remote process group out-of-band.
	cancelWatchDone := make(chan struct{})
	defer close(cancelWatchDone)
	if ctx != nil && ctx.Done() != nil {
		go func() {
			select {
			case <-ctx.Done():
				killRemoteProcessGroup(client, pidFile)
			case <-cancelWatchDone:
			}
		}()
	}

	output, err := session.CombinedOutput(wrapped)

	// Best-effort pidfile cleanup. Uses a short-lived session so a stuck
	// kill goroutine cannot indefinitely tie up cleanup.
	removeRemoteFile(client, pidFile)

	return string(output), err
}

// killRemoteProcessGroup opens a separate SSH session and sends SIGTERM then
// SIGKILL to the process group whose leader PID is recorded in pidFile.
// All operations are best-effort — failures are intentional silent because
// this runs from a cancellation path that should not block the caller.
func killRemoteProcessGroup(client *ssh.Client, pidFile string) {
	if client == nil {
		return
	}
	session, err := client.NewSession()
	if err != nil {
		return
	}
	defer func() { _ = session.Close() }()

	// Race-tolerant kill: the wrapper shell may not have written the pidfile
	// yet at the moment cancellation fires. Retry briefly, then escalate
	// from TERM to KILL on the negative PID (process-group signal).
	script := fmt.Sprintf(`for i in 1 2 3 4 5 6 7 8 9 10; do
  if [ -s %s ]; then
    P=$(cat %s 2>/dev/null)
    if [ -n "$P" ]; then
      kill -TERM -- -"$P" 2>/dev/null
      sleep 1
      kill -KILL -- -"$P" 2>/dev/null
      kill -KILL -- "$P" 2>/dev/null
      break
    fi
  fi
  sleep 1
done
rm -f %s 2>/dev/null`, pidFile, pidFile, pidFile)
	_ = session.Run(script)
}

// removeRemoteFile best-effort deletes a file on the remote host.
func removeRemoteFile(client *ssh.Client, path string) {
	if client == nil {
		return
	}
	session, err := client.NewSession()
	if err != nil {
		return
	}
	defer func() { _ = session.Close() }()
	_ = session.Run(fmt.Sprintf("rm -f %s", path))
}

// Execute runs a command on the remote machine
func (r *SSHRunner) Execute(ctx context.Context, command string) (*CommandResult, error) {
	output, err := r.runCommand(ctx, command)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		}
	}

	return &CommandResult{
		Output:   output,
		ExitCode: exitCode,
		Error:    err,
	}, nil
}

// Cleanup releases the SSH connection back to the pool
func (r *SSHRunner) Cleanup(ctx context.Context) error {
	if r.client != nil {
		if r.pooled {
			// Release back to pool for reuse
			GetSSHPool().Release(r.poolKey)
		} else {
			// Direct close for non-pooled connections
			_ = r.client.Close()
		}
		r.client = nil
	}
	return nil
}

// getPort returns the SSH port, defaulting to 22
func (r *SSHRunner) getPort() int {
	if r.config.Port == 0 {
		return 22
	}
	return r.config.Port
}

// CopyFromRemote copies a file from the SSH host to local using rsync
func (r *SSHRunner) CopyFromRemote(ctx context.Context, remotePath, localPath string) error {
	// Ensure parent directory exists
	if dir := filepath.Dir(localPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Build rsync source: user@host:path
	source := fmt.Sprintf("%s@%s:%s", r.config.User, r.config.Host, remotePath)

	// Build rsync command with SSH options
	args := []string{"-avz", "-e"}
	if r.config.KeyFile != "" {
		keyPath := r.expandPath(r.config.KeyFile)
		args = append(args, fmt.Sprintf("ssh -i %s -p %d -o StrictHostKeyChecking=no", keyPath, r.getPort()))
	} else {
		args = append(args, fmt.Sprintf("ssh -p %d -o StrictHostKeyChecking=no", r.getPort()))
	}
	args = append(args, source, localPath)

	cmd := exec.CommandContext(ctx, "rsync", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsync failed: %w, output: %s", err, string(output))
	}
	return nil
}

// CopyFromRemoteSFTP copies a file from the SSH host to local using SFTP (no rsync dependency).
func (r *SSHRunner) CopyFromRemoteSFTP(remotePath, localPath string) error {
	if r.client == nil {
		return fmt.Errorf("SSH client not connected")
	}

	sftpClient, err := sftp.NewClient(r.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer func() { _ = sftpClient.Close() }()

	remoteFile, err := sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file %s: %w", remotePath, err)
	}
	defer func() { _ = remoteFile.Close() }()

	if dir := filepath.Dir(localPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %w", err)
		}
	}

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer func() { _ = localFile.Close() }()

	if _, err := io.Copy(localFile, remoteFile); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	return nil
}

// Type returns the runner type
func (r *SSHRunner) Type() core.RunnerType {
	return core.RunnerTypeSSH
}

// IsRemote returns true since SSH runner runs on a remote machine
func (r *SSHRunner) IsRemote() bool {
	return true
}

// GetRemoteDir returns the remote directory where binary is stored
func (r *SSHRunner) GetRemoteDir() string {
	return r.remoteDir
}

// SetPIDCallbacks records process lifecycle callbacks. Remote PIDs are NOT
// forwarded to the local run control plane (the control plane's PID kill is
// a local syscall, so a remote PID could collide with an unrelated local
// process). Remote cancellation is instead driven by ctx.Done() inside
// runCommand, which kills the remote process group out-of-band via SSH.
func (r *SSHRunner) SetPIDCallbacks(onStart, onEnd PIDCallback) {
	r.onPIDStart = onStart
	r.onPIDEnd = onEnd
}
