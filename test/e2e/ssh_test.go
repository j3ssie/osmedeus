package e2e

import (
	"net"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testSSHPort     = "2222"
	testSSHUser     = "testuser"
	testSSHPassword = "testpass"
)

// getSSHDockerComposePath returns the path to the SSH test docker-compose file
func getSSHDockerComposePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(getProjectRoot(t), "build", "docker", "docker-compose.test.yaml")
}

// startSSHServer starts the SSH container using docker-compose
func startSSHServer(t *testing.T, log *TestLogger) func() {
	t.Helper()
	composePath := getSSHDockerComposePath(t)

	log.Info("Starting SSH server container from %s", composePath)

	cmd := exec.Command("docker-compose", "-f", composePath, "up", "-d", "ssh-server")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start SSH container: %v\nOutput: %s", err, output)
	}

	// Wait for SSH to be ready
	waitForSSH(t, log, testSSHPort, 30*time.Second)

	return func() {
		log.Info("Stopping SSH server container")
		cmd := exec.Command("docker-compose", "-f", composePath, "down", "-v")
		_ = cmd.Run()
	}
}

// waitForSSH waits for SSH server to accept connections
// The linuxserver/openssh-server container may open port before SSH daemon is ready
func waitForSSH(t *testing.T, log *TestLogger, port string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)

	log.Info("Waiting for SSH on port %s...", port)

	// First, wait for port to be open
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", "localhost:"+port, time.Second)
		if err == nil {
			_ = conn.Close()
			break
		}
		time.Sleep(time.Second)
	}

	// Give SSH daemon extra time to fully initialize after port is open
	log.Info("Port open, waiting for SSH daemon to initialize...")
	time.Sleep(5 * time.Second)

	// Verify SSH connection works by attempting a real connection
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", "localhost:"+port, 2*time.Second)
		if err == nil {
			// Try to read the SSH banner to confirm daemon is ready
			_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			buf := make([]byte, 256)
			n, err := conn.Read(buf)
			_ = conn.Close()
			if err == nil && n > 0 && string(buf[:3]) == "SSH" {
				log.Success("SSH server is ready")
				return
			}
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("SSH server not ready after %v", timeout)
}

// TestSSH_ModuleRunner tests the module-level SSH runner (runner: ssh)
// This test starts a real Docker SSH container and executes a workflow
// that runs all steps on the remote SSH server
func TestSSH_ModuleRunner(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH e2e test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing module-level SSH runner")

	// Start SSH server
	cleanupSSH := startSSHServer(t, log)
	defer cleanupSSH()

	// Run run with SSH runner workflow
	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, stderr, err := runCLIWithLog(t, log, "run",
		"-m", "test-ssh-runner",
		"-t", "ssh-e2e-test",
		"-F", workflowPath)

	require.NoError(t, err, "SSH runner run failed: %s", stderr)
	// Check for successful completion - the CLI doesn't print raw command output
	assert.Contains(t, stdout, "check-remote")
	assert.Contains(t, stdout, "Status: completed")

	log.Success("Module-level SSH runner test passed")
}

// TestSSH_StepRunner tests the step-level SSH runner (step_runner: ssh)
// This test uses the remote-bash step type with SSH runner configuration
// to execute individual steps on a remote SSH server
func TestSSH_StepRunner(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH e2e test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing step-level SSH runner (remote-bash)")

	// Start SSH server
	cleanupSSH := startSSHServer(t, log)
	defer cleanupSSH()

	// Run run with remote-bash SSH workflow
	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, stderr, err := runCLIWithLog(t, log, "run",
		"-m", "test-remote-bash-ssh",
		"-t", "remote-bash-e2e-test",
		"-p", "ssh_host=localhost",
		"-p", "ssh_user=testuser",
		"-p", "ssh_password=testpass",
		"-F", workflowPath)

	require.NoError(t, err, "remote-bash SSH run failed: %s", stderr)
	// Check for successful completion - the CLI doesn't print raw command output
	assert.Contains(t, stdout, "check-ssh-connection")
	assert.Contains(t, stdout, "Status: completed")

	log.Success("Step-level SSH runner test passed")
}

// TestSSH_FullWorkflow is a comprehensive E2E test that verifies:
// 1. SSH container startup
// 2. Module-level SSH runner execution
// 3. Step-level SSH runner execution
// 4. Command output verification
// 5. Proper cleanup
func TestSSH_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH e2e test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Running full SSH workflow e2e test")

	// Step 1: Start SSH server
	log.Step("Step 1: Starting SSH container")
	cleanupSSH := startSSHServer(t, log)
	defer cleanupSSH()

	workflowPath := getTestdataPath(t)

	// Step 2: Test module-level SSH runner
	log.Step("Step 2: Testing module-level SSH runner")
	stdout, stderr, err := runCLIWithLog(t, log, "run",
		"-m", "test-ssh-runner",
		"-t", "full-ssh-test",
		"-F", workflowPath)

	require.NoError(t, err, "Module-level SSH test failed: %s", stderr)
	assert.Contains(t, stdout, "Status: completed", "Expected successful completion")
	assert.Contains(t, stdout, "check-remote")

	// Step 3: Test step-level SSH runner
	log.Step("Step 3: Testing step-level SSH runner")
	stdout, stderr, err = runCLIWithLog(t, log, "run",
		"-m", "test-remote-bash-ssh",
		"-t", "full-remote-bash-test",
		"-p", "ssh_host=localhost",
		"-p", "ssh_user=testuser",
		"-p", "ssh_password=testpass",
		"-F", workflowPath)

	require.NoError(t, err, "Step-level SSH test failed: %s", stderr)
	assert.Contains(t, stdout, "Status: completed", "Expected successful completion")

	log.Success("Full SSH workflow e2e test passed!")
}

// TestSSH_DryRun tests SSH runner in dry-run mode
// NOTE: SSH runner still attempts connection in dry-run mode (runner.Setup is called)
// so this test requires the SSH container to be running
func TestSSH_DryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH e2e test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing SSH runner in dry-run mode")

	// Start SSH server (dry-run still connects for validation)
	cleanupSSH := startSSHServer(t, log)
	defer cleanupSSH()

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	// Module-level SSH runner dry-run
	stdout, _, err := runCLIWithLog(t, log, "run",
		"-m", "test-ssh-runner",
		"-t", "ssh-dryrun-test",
		"--dry-run",
		"-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "DRY-RUN")

	// Step-level SSH runner dry-run
	stdout, _, err = runCLIWithLog(t, log, "run",
		"-m", "test-remote-bash-ssh",
		"-t", "remote-bash-dryrun-test",
		"--dry-run",
		"-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "DRY-RUN")

	log.Success("SSH dry-run test passed")
}
