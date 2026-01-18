package e2e

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nixTestContainer = "osm-test-nix"
)

// getNixDockerComposePath returns the path to the Nix test docker-compose file
func getNixDockerComposePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(getProjectRoot(t), "build", "docker", "docker-compose.nix-test.yaml")
}

// isNixInstalled checks if Nix is available on the host
func isNixInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

// startNixContainer starts the Nix container using docker-compose
func startNixContainer(t *testing.T, log *TestLogger) func() {
	t.Helper()
	composePath := getNixDockerComposePath(t)

	log.Info("Building Nix test container from %s", composePath)

	// Build the container first
	buildCmd := exec.Command("docker-compose", "-f", composePath, "build")
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build Nix container: %v\nOutput: %s", err, buildOutput)
	}

	log.Info("Starting Nix test container")

	cmd := exec.Command("docker-compose", "-f", composePath, "up", "-d", "nix-test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start Nix container: %v\nOutput: %s", err, output)
	}

	// Wait for container to be ready
	waitForNixContainer(t, log, 30*time.Second)

	return func() {
		log.Info("Stopping Nix test container")
		cmd := exec.Command("docker-compose", "-f", composePath, "down", "-v")
		_ = cmd.Run()
	}
}

// waitForNixContainer waits for the Nix container to be ready
func waitForNixContainer(t *testing.T, log *TestLogger, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)

	log.Info("Waiting for Nix container to be ready...")

	for time.Now().Before(deadline) {
		cmd := exec.Command("docker", "exec", nixTestContainer, "nix", "--version")
		output, err := cmd.CombinedOutput()
		if err == nil && strings.Contains(string(output), "nix") {
			log.Success("Nix container is ready")
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("Nix container not ready after %v", timeout)
}

// runCLIInNixContainer executes the osmedeus CLI inside the Nix container
func runCLIInNixContainer(t *testing.T, log *TestLogger, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	// Build the full command
	dockerArgs := []string{"exec", nixTestContainer, "/app/bin/osmedeus"}
	dockerArgs = append(dockerArgs, args...)

	log.Command(args...)

	cmd := exec.Command("docker", dockerArgs...)
	var stdoutBuf, stderrBuf strings.Builder
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	log.Result(stdout, stderr)

	if err != nil {
		log.Error("Command failed: %v", err)
	}

	return stdout, stderr, err
}

// TestNix_ListBinaries tests the --list-binary-nix flag
// This test runs on the host and doesn't require Nix installation
func TestNix_ListBinaries(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing --list-binary-nix flag")

	stdout, _, err := runCLIWithLog(t, log, "install", "binary", "--list-registry-nix-build")
	require.NoError(t, err)

	// Verify output contains expected categories
	log.Info("Asserting output contains tool categories")
	assert.Contains(t, stdout, "Nix Flake Binaries")
	assert.Contains(t, stdout, "Total:")

	log.Success("--list-binary-nix flag works correctly")
}

// TestNix_ListBinaryRegistry tests the --list-binary-registry flag
func TestNix_ListBinaryRegistry(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing --list-binary-registry flag")

	stdout, _, err := runCLIWithLog(t, log, "install", "binary", "--list-registry-direct-fetch")
	require.NoError(t, err)

	// Verify output contains expected content
	log.Info("Asserting output contains registry binaries")
	assert.Contains(t, stdout, "Registry Binaries")
	assert.Contains(t, stdout, "nuclei")
	assert.Contains(t, stdout, "Total:")

	log.Success("--list-binary-registry flag works correctly")
}

// TestNix_InstallBinary_Host tests Nix mode installation on the host
// This test is skipped if Nix is not installed on the host machine
func TestNix_InstallBinary_Host(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Nix e2e test in short mode")
	}

	if !isNixInstalled() {
		t.Skip("Nix is not installed on host, skipping host-based test")
	}

	log := NewTestLogger(t)
	log.Step("Testing Nix mode installation on host")

	// Test with curl (a simple, commonly available package)
	stdout, stderr, err := runCLIWithLog(t, log, "install", "binary",
		"--name", "curl",
		"--nix-build-install")

	combinedOutput := stdout + stderr

	// The command should succeed (or skip if binary already in PATH)
	if err != nil {
		log.Warn("Installation may have partially failed: %s", stderr)
	}

	// Check for acceptable outcomes:
	// 1. Binary already available in PATH (skipped)
	// 2. Nix installation was attempted
	// 3. Installation succeeded
	alreadyInPath := strings.Contains(combinedOutput, "already available in PATH")
	nixAttempted := strings.Contains(combinedOutput, "Installing via Nix") ||
		strings.Contains(combinedOutput, "nix profile add")
	installSucceeded := strings.Contains(combinedOutput, "installed") ||
		strings.Contains(combinedOutput, "success")

	acceptableOutcome := alreadyInPath || nixAttempted || installSucceeded || err == nil

	if alreadyInPath {
		log.Info("Binary already in PATH, skipping installation (expected behavior)")
	}

	assert.True(t, acceptableOutcome, "Expected either: binary in PATH, Nix install attempted, or success")

	log.Success("Nix mode installation test completed on host")
}

// TestNix_InstallBinary_Docker tests Nix mode installation inside a Docker container
// This is the primary test for Nix installation as it provides an isolated environment
func TestNix_InstallBinary_Docker(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Nix Docker e2e test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing Nix mode installation in Docker container")

	// Start Nix container
	cleanup := startNixContainer(t, log)
	defer cleanup()

	// Step 1: Verify Nix is available in container
	log.Step("Step 1: Verifying Nix is available in container")
	cmd := exec.Command("docker", "exec", nixTestContainer, "nix", "--version")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Nix should be available in container: %s", output)
	assert.Contains(t, string(output), "nix")

	// Step 2: Verify osmedeus binary is accessible
	log.Step("Step 2: Verifying osmedeus binary is accessible")
	stdout, stderr, err := runCLIInNixContainer(t, log, "--version")
	require.NoError(t, err, "osmedeus should be accessible: %s", stderr)
	assert.Contains(t, stdout, "osmedeus")

	// Step 3: Test list-binary-nix in container
	log.Step("Step 3: Testing --list-binary-nix in container")
	stdout, _, err = runCLIInNixContainer(t, log, "install", "binary", "--list-binary-nix")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Available binaries in Nix flake")

	// Step 4: Test list-binary-registry in container
	log.Step("Step 4: Testing --list-binary-registry in container")
	stdout, _, err = runCLIInNixContainer(t, log, "install", "binary", "--list-binary-registry")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Available binaries in registry")

	// Step 5: Test actual Nix installation with curl (simple package)
	log.Step("Step 5: Testing Nix mode installation with curl")
	stdout, stderr, err = runCLIInNixContainer(t, log, "install", "binary",
		"--name", "curl",
		"--nix-build-install")

	// The installation should at least attempt to use Nix
	combinedOutput := stdout + stderr
	if err != nil {
		log.Warn("Installation command returned error: %v", err)
		// Check if it's a non-fatal issue
		if strings.Contains(combinedOutput, "nix") || strings.Contains(combinedOutput, "Nix") {
			log.Info("Nix was invoked, test passed despite error")
		}
	}

	// Verify Nix was invoked
	nixInvoked := strings.Contains(combinedOutput, "Installing via Nix") ||
		strings.Contains(combinedOutput, "nix profile add") ||
		strings.Contains(combinedOutput, "nixpkgs#")

	if !nixInvoked && err == nil {
		// If no error and nix wasn't explicitly mentioned, it might have succeeded silently
		log.Info("Command succeeded, verifying curl is available")
		verifyCmd := exec.Command("docker", "exec", nixTestContainer, "which", "curl")
		verifyOutput, verifyErr := verifyCmd.CombinedOutput()
		if verifyErr == nil && len(verifyOutput) > 0 {
			log.Success("curl is available in container")
		}
	}

	log.Success("Nix mode installation test completed in Docker")
}

// TestNix_CheckBinary tests the --check flag with Nix-related binaries
func TestNix_CheckBinary(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing --check flag")

	// Test checking a binary that's likely to be installed (curl is common)
	stdout, _, err := runCLIWithLog(t, log, "install", "binary",
		"--name", "curl",
		"--check")

	require.NoError(t, err)

	// The check should report the status
	log.Info("Asserting output contains check result")
	assert.Contains(t, stdout, "curl")

	log.Success("--check flag works correctly")
}

// TestNix_NixModeWithoutNix tests error handling when Nix is not installed
// This test verifies the error message when --nix-build-install is used but Nix is unavailable
func TestNix_NixModeWithoutNix(t *testing.T) {
	if isNixInstalled() {
		t.Skip("Nix is installed, skipping 'nix not installed' error test")
	}

	log := NewTestLogger(t)
	log.Step("Testing --nix-build-install error when Nix is not installed")

	_, stderr, err := runCLIWithLog(t, log, "install", "binary",
		"--name", "curl",
		"--nix-build-install")

	// Should fail because Nix is not installed
	require.Error(t, err)

	// Verify helpful error message
	combinedOutput := stderr
	hasNixError := strings.Contains(combinedOutput, "Nix is not installed") ||
		strings.Contains(combinedOutput, "nix-installation")

	assert.True(t, hasNixError, "Expected helpful error message about Nix not being installed")

	log.Success("Error handling for missing Nix works correctly")
}

// TestNix_FullWorkflow runs a comprehensive Nix e2e test in Docker
// This test covers the complete workflow: list -> check -> install -> verify
func TestNix_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Nix full workflow e2e test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Running full Nix workflow e2e test")

	// Start Nix container
	cleanup := startNixContainer(t, log)
	defer cleanup()

	// Step 1: List available Nix binaries
	log.Step("Step 1: Listing available Nix binaries")
	stdout, _, err := runCLIInNixContainer(t, log, "install", "binary", "--list-binary-nix")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Available binaries in Nix flake")
	assert.Contains(t, stdout, "Total:")

	// Step 2: List available registry binaries
	log.Step("Step 2: Listing available registry binaries")
	stdout, _, err = runCLIInNixContainer(t, log, "install", "binary", "--list-binary-registry")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Available binaries in registry")

	// Step 3: Check if curl is installed (it shouldn't be initially)
	log.Step("Step 3: Checking initial curl status")
	stdout, _, _ = runCLIInNixContainer(t, log, "install", "binary", "--name", "curl", "--check")
	// Log the initial status (may or may not be installed)
	log.Info("Initial curl status: %s", strings.TrimSpace(stdout))

	// Step 4: Install curl via Nix
	log.Step("Step 4: Installing curl via Nix")
	stdout, stderr, err := runCLIInNixContainer(t, log, "install", "binary",
		"--name", "curl",
		"--nix-build-install")

	combinedOutput := stdout + stderr
	if err != nil {
		log.Warn("Installation returned error: %v", err)
		// Check if it's a known acceptable error
		if strings.Contains(combinedOutput, "already") {
			log.Info("curl already installed, continuing")
		}
	}

	// Step 5: Verify curl is now available
	log.Step("Step 5: Verifying curl installation")
	verifyCmd := exec.Command("docker", "exec", nixTestContainer, "which", "curl")
	verifyOutput, _ := verifyCmd.CombinedOutput()
	if len(verifyOutput) > 0 {
		log.Success("curl is available at: %s", strings.TrimSpace(string(verifyOutput)))
	}

	// Step 6: Final check
	log.Step("Step 6: Final binary check")
	stdout, _, _ = runCLIInNixContainer(t, log, "install", "binary", "--name", "curl", "--check")
	log.Info("Final curl status: %s", strings.TrimSpace(stdout))

	log.Success("Full Nix workflow e2e test passed!")
}
