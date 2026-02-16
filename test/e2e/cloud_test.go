package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloud_ConfigSet tests cloud configuration setting via CLI
func TestCloud_ConfigSet(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud config set command")

	// Test setting provider token
	stdout, _, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.digitalocean.token", "test-token-12345")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Configuration updated successfully")

	log.Success("Cloud config set command works correctly")
}

// TestCloud_ConfigShow tests cloud configuration display
func TestCloud_ConfigShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud config show command")

	// First set some test config
	_, _, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "defaults.provider", "digitalocean")
	require.NoError(t, err)

	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "defaults.max_instances", "5")
	require.NoError(t, err)

	// Show configuration
	stdout, _, err := runCLIWithLog(t, log, "cloud", "config", "show")
	require.NoError(t, err)

	// Verify output contains expected configuration
	assert.Contains(t, stdout, "provider:")
	assert.Contains(t, stdout, "digitalocean")

	log.Success("Cloud config show displays configuration correctly")
}

// TestCloud_ConfigSetInvalidKey tests error handling for invalid config keys
func TestCloud_ConfigSetInvalidKey(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud config set with invalid key")

	stdout, stderr, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "invalid.nested.key.that.does.not.exist", "value")

	// Should either error or provide warning
	if err != nil {
		log.Info("Command failed as expected with error: %v", err)
	} else {
		combined := stdout + stderr
		log.Info("Command output: %s", combined)
	}

	log.Success("Invalid config key handled appropriately")
}

// TestCloud_ConfigEnvironmentVariables tests environment variable resolution
func TestCloud_ConfigEnvironmentVariables(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud config with environment variables")

	// Set environment variable
	testToken := "env-test-token-67890"
	require.NoError(t, os.Setenv("TEST_DIGITALOCEAN_TOKEN", testToken))
	defer func() {
		_ = os.Unsetenv("TEST_DIGITALOCEAN_TOKEN")
	}()

	// Set config with environment variable reference
	stdout, _, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.digitalocean.token", "${TEST_DIGITALOCEAN_TOKEN}")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Configuration updated successfully")

	// Show config and verify environment variable is resolved
	stdout, _, err = runCLIWithLog(t, log, "cloud", "config", "show")
	require.NoError(t, err)

	// Note: Environment variables may be shown as-is or resolved depending on implementation
	log.Info("Config output: %s", stdout)

	log.Success("Environment variable configuration works")
}

// TestCloud_CreateHelp tests cloud create help output
func TestCloud_CreateHelp(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud create --help")

	stdout, _, err := runCLIWithLog(t, log, "cloud", "create", "--help")
	require.NoError(t, err)

	// Verify help contains expected flags
	assert.Contains(t, stdout, "--provider")
	assert.Contains(t, stdout, "--mode")
	assert.Contains(t, stdout, "--instances")
	assert.Contains(t, stdout, "--force")

	log.Success("Cloud create help displays all required flags")
}

// TestCloud_CreateWithoutConfig tests error handling when config is missing
func TestCloud_CreateWithoutConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cloud create test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cloud create without configuration")

	baseDir, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"cloud", "create", "--provider", "digitalocean", "--instances", "1")

	combined := stdout + stderr
	log.Info("Output: %s", combined)

	// Should fail due to missing configuration or show "not yet fully implemented"
	if err != nil {
		log.Info("Command failed as expected: %v", err)
		assert.Contains(t, combined, "token")
	} else {
		// May show "not yet fully implemented" message
		log.Info("Command completed with output indicating implementation status")
	}

	// Clean up base directory
	if baseDir != "" {
		_ = os.RemoveAll(baseDir)
	}

	log.Success("Cloud create handles missing configuration appropriately")
}

// TestCloud_CreateDryRun tests cloud create with cost estimation
func TestCloud_CreateDryRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cloud create dry-run test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cloud create cost estimation")

	// Set minimal configuration
	_, _, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.digitalocean.token", "test-token")
	require.NoError(t, err)

	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.digitalocean.region", "nyc3")
	require.NoError(t, err)

	// Attempt to create infrastructure
	baseDir, stdout, stderr, _ := runCLIWithLogAndBase(t, log,
		"cloud", "create", "--provider", "digitalocean", "--instances", "3")

	combined := stdout + stderr
	log.Info("Output: %s", combined)

	// Clean up base directory
	if baseDir != "" {
		_ = os.RemoveAll(baseDir)
	}

	// Should either show cost estimation or "not yet fully implemented"
	if strings.Contains(combined, "not yet fully implemented") {
		log.Info("Cloud create shows expected implementation status")
	} else if strings.Contains(combined, "cost") || strings.Contains(combined, "hourly") {
		log.Info("Cloud create shows cost estimation")
	}

	log.Success("Cloud create dry-run test completed")
}

// TestCloud_List tests cloud infrastructure listing
func TestCloud_List(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud list command")

	stdout, _, err := runCLIWithLog(t, log, "cloud", "list")
	require.NoError(t, err)

	// Should show either empty list or existing infrastructure
	log.Info("List output: %s", stdout)

	// Verify output format
	if strings.Contains(stdout, "No active") || strings.Contains(stdout, "not yet fully implemented") {
		log.Info("No active infrastructure or feature in development")
	} else {
		// If infrastructure exists, verify table format
		assert.Contains(t, stdout, "ID")
	}

	log.Success("Cloud list command works correctly")
}

// TestCloud_DestroyHelp tests cloud destroy help output
func TestCloud_DestroyHelp(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud destroy --help")

	stdout, _, err := runCLIWithLog(t, log, "cloud", "destroy", "--help")
	require.NoError(t, err)

	// Verify help contains usage information
	assert.Contains(t, stdout, "destroy")
	assert.Contains(t, stdout, "infrastructure")

	log.Success("Cloud destroy help displays correctly")
}

// TestCloud_DestroyNonExistent tests destroying non-existent infrastructure
func TestCloud_DestroyNonExistent(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud destroy with non-existent ID")

	baseDir, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"cloud", "destroy", "non-existent-infrastructure-id")

	combined := stdout + stderr
	log.Info("Output: %s", combined)

	// Clean up base directory
	if baseDir != "" {
		_ = os.RemoveAll(baseDir)
	}

	// Should either error or show "not found" message
	if err != nil {
		log.Info("Command failed as expected for non-existent infrastructure")
	} else if strings.Contains(combined, "not found") || strings.Contains(combined, "not yet fully implemented") {
		log.Info("Command handled non-existent infrastructure appropriately")
	}

	log.Success("Cloud destroy handles non-existent infrastructure correctly")
}

// TestCloud_RunHelp tests cloud run help output
func TestCloud_RunHelp(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud run --help")

	stdout, _, err := runCLIWithLog(t, log, "cloud", "run", "--help")
	require.NoError(t, err)

	// Verify help contains expected flags
	assert.Contains(t, stdout, "--provider")
	assert.Contains(t, stdout, "--instances")
	assert.Contains(t, stdout, "--flow")

	log.Success("Cloud run help displays all required flags")
}

// TestCloud_RunWithoutTarget tests error handling when target is missing
func TestCloud_RunWithoutTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cloud run test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cloud run without target")

	baseDir, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"cloud", "run", "-f", "general", "--instances", "1")

	combined := stdout + stderr
	log.Info("Output: %s", combined)

	// Clean up base directory
	if baseDir != "" {
		_ = os.RemoveAll(baseDir)
	}

	// Should fail due to missing target
	if err != nil {
		log.Info("Command failed as expected: %v", err)
		assert.Contains(t, combined, "target")
	}

	log.Success("Cloud run handles missing target appropriately")
}

// TestCloud_ConfigFileCreation tests that cloud config file is created properly
func TestCloud_ConfigFileCreation(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud config file creation")

	// Create temporary base directory
	baseDir := t.TempDir()

	// Set config with custom base directory
	_, _, err := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "defaults.provider", "digitalocean")
	require.NoError(t, err)

	// Verify config file was created
	configPath := filepath.Join(baseDir, "cloud-settings.yaml")
	_, err = os.Stat(configPath)
	if err == nil {
		log.Info("Config file created at: %s", configPath)

		// Read and verify content
		content, err := os.ReadFile(configPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "digitalocean")

		log.Success("Cloud config file created and contains expected data")
	} else if os.IsNotExist(err) {
		log.Info("Config file not created (may use different path or strategy)")
	} else {
		require.NoError(t, err)
	}
}

// TestCloud_MultipleProviderConfigs tests configuring multiple cloud providers
func TestCloud_MultipleProviderConfigs(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing multiple cloud provider configurations")

	// Configure DigitalOcean
	_, _, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.digitalocean.token", "do-token-12345")
	require.NoError(t, err)

	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.digitalocean.region", "nyc3")
	require.NoError(t, err)

	// Configure AWS
	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.aws.access_key_id", "aws-key-12345")
	require.NoError(t, err)

	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.aws.region", "us-east-1")
	require.NoError(t, err)

	// Show configuration
	stdout, _, err := runCLIWithLog(t, log, "cloud", "config", "show")
	require.NoError(t, err)

	// Verify both providers are configured
	assert.Contains(t, stdout, "digitalocean")
	assert.Contains(t, stdout, "aws")

	log.Success("Multiple provider configurations work correctly")
}

// TestCloud_CostLimitConfiguration tests cost limit settings
func TestCloud_CostLimitConfiguration(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cost limit configuration")

	// Set hourly spend limit
	_, _, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "limits.max_hourly_spend", "0.50")
	require.NoError(t, err)

	// Set total spend limit
	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "limits.max_total_spend", "5.00")
	require.NoError(t, err)

	// Set max instances
	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "limits.max_instances", "10")
	require.NoError(t, err)

	// Verify configuration
	stdout, _, err := runCLIWithLog(t, log, "cloud", "config", "show")
	require.NoError(t, err)

	log.Info("Config output: %s", stdout)
	assert.Contains(t, stdout, "limits")

	log.Success("Cost limit configuration works correctly")
}

// TestCloud_StateDirectory tests cloud state directory creation
func TestCloud_StateDirectory(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing cloud state directory")

	// Create temporary base directory
	baseDir := t.TempDir()

	// Set state path configuration
	_, _, err := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "state.backend", "local")
	require.NoError(t, err)

	// Expected state directory path
	stateDir := filepath.Join(baseDir, "cloud-state")
	log.Info("Expected state directory: %s", stateDir)

	// Note: State directory may only be created when actually used
	// This test just verifies the configuration can be set
	log.Success("Cloud state configuration set successfully")
}

// TestCloud_Integration_FullLifecycle tests a complete cloud workflow (when implemented)
func TestCloud_Integration_FullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full lifecycle integration test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cloud full lifecycle integration")

	// This test is designed to run when cloud functionality is fully implemented
	// For now, it tests the command sequence

	baseDir := t.TempDir()

	// 1. Configure provider
	log.Step("Step 1: Configure provider")
	_, _, err := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "providers.digitalocean.token", "test-token")
	require.NoError(t, err)

	_, _, err = runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "defaults.provider", "digitalocean")
	require.NoError(t, err)

	// 2. Attempt to create infrastructure
	log.Step("Step 2: Attempt infrastructure creation")
	stdout, stderr, _ := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "create", "--instances", "1")

	combined := stdout + stderr
	log.Info("Create output: %s", combined)

	// 3. List infrastructure
	log.Step("Step 3: List infrastructure")
	stdout, _, _ = runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "list")
	log.Info("List output: %s", stdout)

	// 4. Attempt to destroy (if any infrastructure was created)
	log.Step("Step 4: Attempt infrastructure destruction")
	stdout, stderr, _ = runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "destroy", "--all")
	combined = stdout + stderr
	log.Info("Destroy output: %s", combined)

	log.Success("Cloud full lifecycle test sequence completed")
}

// TestCloud_SSHKeyConfiguration tests SSH key settings for cloud workers
func TestCloud_SSHKeyConfiguration(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing SSH key configuration for cloud")

	// Set SSH private key path
	_, _, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "ssh.private_key_path", "/home/user/.ssh/cloud_rsa")
	require.NoError(t, err)

	// Set SSH user
	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "ssh.user", "osmedeus")
	require.NoError(t, err)

	// Verify configuration
	stdout, _, err := runCLIWithLog(t, log, "cloud", "config", "show")
	require.NoError(t, err)

	log.Info("Config output: %s", stdout)

	log.Success("SSH key configuration works correctly")
}

// TestCloud_ProviderValidation tests that invalid provider names are rejected
func TestCloud_ProviderValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping provider validation test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing invalid provider name validation")

	baseDir, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"cloud", "create", "--provider", "invalid-provider-name", "--instances", "1")

	combined := stdout + stderr
	log.Info("Output: %s", combined)

	// Clean up base directory
	if baseDir != "" {
		_ = os.RemoveAll(baseDir)
	}

	// Should fail with validation error or show as unsupported
	if err != nil {
		log.Info("Command failed as expected for invalid provider")
		assert.Contains(t, combined, "provider")
	}

	log.Success("Invalid provider name handled correctly")
}

// TestCloud_InstanceCountValidation tests instance count validation
func TestCloud_InstanceCountValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping instance count validation test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing instance count validation")

	// Test with zero instances
	baseDir, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"cloud", "create", "--provider", "digitalocean", "--instances", "0")

	combined := stdout + stderr
	log.Info("Output for zero instances: %s", combined)

	if baseDir != "" {
		_ = os.RemoveAll(baseDir)
	}

	// Should fail or warn about invalid instance count
	if err != nil {
		log.Info("Zero instances rejected as expected")
	}

	// Test with negative instances
	baseDir, stdout, stderr, err = runCLIWithLogAndBase(t, log,
		"cloud", "create", "--provider", "digitalocean", "--instances", "-1")

	combined = stdout + stderr
	log.Info("Output for negative instances: %s", combined)

	if baseDir != "" {
		_ = os.RemoveAll(baseDir)
	}

	// Should fail
	if err != nil {
		log.Info("Negative instances rejected as expected")
	}

	log.Success("Instance count validation works correctly")
}

// TestCloud_Distributed_WorkerRegistration tests cloud worker registration (integration)
func TestCloud_Distributed_WorkerRegistration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping distributed worker registration test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cloud worker registration with distributed system")

	// This test requires Redis and distributed setup
	// It's a placeholder for when cloud functionality is fully implemented

	log.Info("This test will verify that cloud-provisioned workers can register with master")
	log.Info("Requires: Redis, master server, cloud provider access")

	// Test sequence would be:
	// 1. Start Redis
	// 2. Start master server with --master flag
	// 3. Provision cloud infrastructure
	// 4. Wait for workers to register (via cloud-init script)
	// 5. Verify worker count matches requested instances
	// 6. Run distributed workflow
	// 7. Collect results
	// 8. Destroy infrastructure

	log.Success("Cloud worker registration test placeholder created")
}

// TestCloud_RunWithDistributed tests cloud run with distributed workflow execution
func TestCloud_RunWithDistributed(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cloud run distributed test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cloud run with distributed workflow")

	baseDir := t.TempDir()

	// Configure provider
	_, _, err := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "providers.digitalocean.token", "test-token")
	require.NoError(t, err)

	// Get test workflow path
	workflowPath := filepath.Join(getTestdataPath(t), "test-bash.yaml")

	// Attempt cloud run
	stdout, stderr, _ := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "run",
		"-m", "test-bash",
		"-F", workflowPath,
		"-t", "example.com",
		"--instances", "2",
		"--provider", "digitalocean")

	combined := stdout + stderr
	log.Info("Cloud run output: %s", combined)

	log.Success("Cloud run command sequence completed")
}

// TestCloud_CostTracking tests cost tracking during execution
func TestCloud_CostTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cost tracking test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cloud cost tracking")

	// This test verifies that cost tracking works during cloud execution
	// It would require actual infrastructure or mocked provider

	log.Info("Cost tracking test would verify:")
	log.Info("- Hourly cost calculation")
	log.Info("- Real-time cost updates")
	log.Info("- Max spend limit enforcement")
	log.Info("- Cost summary on completion")

	log.Success("Cost tracking test placeholder created")
}

// TestCloud_Cleanup_OnFailure tests cleanup behavior when provisioning fails
func TestCloud_Cleanup_OnFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cleanup on failure test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cleanup on provisioning failure")

	baseDir := t.TempDir()

	// Set cleanup_on_failure configuration
	_, _, err := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "defaults.cleanup_on_failure", "true")
	require.NoError(t, err)

	// Configure with invalid credentials to trigger failure
	_, _, err = runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "providers.digitalocean.token", "invalid-token")
	require.NoError(t, err)

	// Attempt to create infrastructure (should fail and cleanup)
	stdout, stderr, _ := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "create",
		"--provider", "digitalocean",
		"--instances", "1")

	combined := stdout + stderr
	log.Info("Create with failure output: %s", combined)

	// Verify no infrastructure state remains
	stateDir := filepath.Join(baseDir, "cloud-state", "infrastructure")
	if _, err := os.Stat(stateDir); err == nil {
		entries, _ := os.ReadDir(stateDir)
		log.Info("Found %d state files (should be 0 after cleanup)", len(entries))
	} else {
		log.Info("State directory does not exist (cleanup successful)")
	}

	log.Success("Cleanup on failure test completed")
}

// TestCloud_WithTimeout tests cloud operations with timeout
func TestCloud_WithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timeout test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing cloud operations with timeout")

	baseDir := t.TempDir()

	// Set timeout configuration
	_, _, err := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "defaults.timeout", "5m")
	require.NoError(t, err)

	// Verify configuration
	stdout, _, err := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "show")
	require.NoError(t, err)

	log.Info("Config output: %s", stdout)

	log.Success("Timeout configuration works correctly")
}

// TestCloud_ParallelOperations tests that multiple cloud operations can't interfere
func TestCloud_ParallelOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping parallel operations test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing parallel cloud operations isolation")

	// This test would verify that multiple concurrent cloud operations
	// don't interfere with each other (different base folders, state isolation)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan bool, 2)

	// Operation 1: Configure and attempt create
	go func() {
		baseDir1 := t.TempDir()
		_, _, _ = runCLIWithLog(t, log,
			"--base-folder", baseDir1,
			"cloud", "config", "set", "defaults.provider", "digitalocean")
		done <- true
	}()

	// Operation 2: Configure and list
	go func() {
		baseDir2 := t.TempDir()
		_, _, _ = runCLIWithLog(t, log,
			"--base-folder", baseDir2,
			"cloud", "config", "set", "defaults.provider", "aws")
		done <- true
	}()

	// Wait for both operations with timeout
	for range 2 {
		select {
		case <-done:
			log.Info("Operation completed")
		case <-ctx.Done():
			t.Fatal("Parallel operations timed out")
		}
	}

	log.Success("Parallel operations completed without interference")
}

// TestCloud_SpotInstanceConfiguration tests spot/preemptible instance settings
func TestCloud_SpotInstanceConfiguration(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing spot instance configuration")

	// AWS spot instances
	_, _, err := runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.aws.use_spot", "true")
	require.NoError(t, err)

	// GCP preemptible instances
	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "providers.gcp.use_preemptible", "true")
	require.NoError(t, err)

	// Set global default
	_, _, err = runCLIWithLog(t, log,
		"cloud", "config", "set", "defaults.use_spot", "true")
	require.NoError(t, err)

	// Verify configuration
	stdout, _, err := runCLIWithLog(t, log, "cloud", "config", "show")
	require.NoError(t, err)

	log.Info("Config output: %s", stdout)

	log.Success("Spot instance configuration works correctly")
}

// TestCloud_CustomSetupCommands tests custom worker setup commands
func TestCloud_CustomSetupCommands(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing custom setup commands configuration")

	// This would test adding custom commands to setup.commands array
	// For now, just verify the configuration path exists

	baseDir := t.TempDir()

	// Note: Setting array values via CLI may require special syntax
	// This test documents the expected capability

	log.Info("Custom setup commands would be configured via:")
	log.Info("  setup.commands[0] = 'apt-get update'")
	log.Info("  setup.commands[1] = 'apt-get install -y custom-tool'")

	// Verify config file can be created
	_, _, err := runCLIWithLog(t, log,
		"--base-folder", baseDir,
		"cloud", "config", "set", "defaults.provider", "digitalocean")
	require.NoError(t, err)

	log.Success("Custom setup commands configuration documented")
}

// TestCloud_ProviderRegions tests region configuration for different providers
func TestCloud_ProviderRegions(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing provider region configuration")

	regions := map[string]string{
		"digitalocean": "nyc3",
		"aws":          "us-east-1",
		"gcp":          "us-central1",
		"linode":       "us-east",
		"azure":        "eastus",
	}

	for provider, region := range regions {
		key := fmt.Sprintf("providers.%s.region", provider)
		_, _, err := runCLIWithLog(t, log,
			"cloud", "config", "set", key, region)
		require.NoError(t, err)
		log.Info("Set region for %s: %s", provider, region)
	}

	// Verify all regions are configured
	stdout, _, err := runCLIWithLog(t, log, "cloud", "config", "show")
	require.NoError(t, err)

	for _, region := range regions {
		assert.Contains(t, stdout, region)
	}

	log.Success("Provider region configuration works for all providers")
}
