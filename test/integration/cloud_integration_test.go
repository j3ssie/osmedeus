package integration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/cloud"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloudConfig_LoadAndSave tests cloud configuration persistence
func TestCloudConfig_LoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cloud-settings.yaml")

	// Create a test cloud config
	cfg := &config.CloudConfigs{
		Providers: config.Providers{
			DigitalOcean: config.DigitalOceanConfig{
				Token:  "test-token-12345",
				Region: "nyc3",
				Size:   "s-2vcpu-4gb",
				Image:  "ubuntu-22-04-x64",
			},
		},
		Defaults: config.Defaults{
			Provider:         "digitalocean",
			Mode:             "vm",
			MaxInstances:     5,
			Timeout:          "10m",
			CleanupOnFailure: true,
		},
		Limits: config.Limits{
			MaxHourlySpend: 1.00,
			MaxTotalSpend:  10.00,
			MaxInstances:   20,
		},
		State: config.State{
			Backend: "local",
			Path:    filepath.Join(tempDir, "cloud-state"),
		},
	}

	// Save configuration
	err := cloud.SaveCloudConfig(cfg, configPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load configuration back
	loadedCfg, err := cloud.LoadCloudConfig(configPath)
	require.NoError(t, err)

	// Verify loaded config matches original
	assert.Equal(t, cfg.Providers.DigitalOcean.Token, loadedCfg.Providers.DigitalOcean.Token)
	assert.Equal(t, cfg.Providers.DigitalOcean.Region, loadedCfg.Providers.DigitalOcean.Region)
	assert.Equal(t, cfg.Defaults.Provider, loadedCfg.Defaults.Provider)
	assert.Equal(t, cfg.Defaults.MaxInstances, loadedCfg.Defaults.MaxInstances)
	assert.Equal(t, cfg.Limits.MaxHourlySpend, loadedCfg.Limits.MaxHourlySpend)
}

// TestCloudConfig_EnvironmentVariableResolution tests environment variable substitution
func TestCloudConfig_EnvironmentVariableResolution(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cloud-settings.yaml")

	// Set test environment variables
	testToken := "env-resolved-token-67890"
	require.NoError(t, os.Setenv("TEST_DO_TOKEN", testToken))
	defer func() {
		_ = os.Unsetenv("TEST_DO_TOKEN")
	}()

	// Create config with environment variable reference
	cfg := &config.CloudConfigs{
		Providers: config.Providers{
			DigitalOcean: config.DigitalOceanConfig{
				Token:  "${TEST_DO_TOKEN}",
				Region: "nyc3",
			},
		},
		Defaults: config.Defaults{
			Provider: "digitalocean",
		},
	}

	// Save configuration
	err := cloud.SaveCloudConfig(cfg, configPath)
	require.NoError(t, err)

	// Load and resolve environment variables
	loadedCfg, err := cloud.LoadCloudConfig(configPath)
	require.NoError(t, err)

	// Verify environment variable was resolved
	assert.Equal(t, testToken, loadedCfg.Providers.DigitalOcean.Token)
}

// TestCloudConfig_Validation tests configuration validation
func TestCloudConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.CloudConfigs
		shouldError bool
		errorMsg    string
	}{
		{
			name: "valid_config",
			config: &config.CloudConfigs{
				Providers: config.Providers{
					DigitalOcean: config.DigitalOceanConfig{
						Token:  "valid-token",
						Region: "nyc3",
					},
				},
				Defaults: config.Defaults{
					Provider:     "digitalocean",
					MaxInstances: 5,
				},
				Limits: config.Limits{
					MaxHourlySpend: 10.0,
					MaxTotalSpend:  100.0,
					MaxInstances:   20,
				},
			},
			shouldError: false,
		},
		{
			name: "missing_provider_token",
			config: &config.CloudConfigs{
				Defaults: config.Defaults{
					Provider:     "digitalocean",
					MaxInstances: 5,
				},
				Limits: config.Limits{
					MaxHourlySpend: 10.0,
					MaxTotalSpend:  100.0,
					MaxInstances:   20,
				},
			},
			shouldError: true,
			errorMsg:    "token",
		},
		{
			name: "invalid_provider",
			config: &config.CloudConfigs{
				Defaults: config.Defaults{
					Provider:     "invalid-provider",
					MaxInstances: 5,
				},
				Limits: config.Limits{
					MaxHourlySpend: 10.0,
					MaxTotalSpend:  100.0,
					MaxInstances:   20,
				},
			},
			shouldError: true,
			errorMsg:    "provider",
		},
		{
			name: "invalid_max_instances",
			config: &config.CloudConfigs{
				Providers: config.Providers{
					DigitalOcean: config.DigitalOceanConfig{
						Token:  "valid-token",
						Region: "nyc3",
					},
				},
				Defaults: config.Defaults{
					Provider:     "digitalocean",
					MaxInstances: 0,
				},
				Limits: config.Limits{
					MaxHourlySpend: 10.0,
					MaxTotalSpend:  100.0,
					MaxInstances:   20,
				},
			},
			shouldError: true,
			errorMsg:    "instances",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cloud.ValidateCloudConfig(tt.config)
			if tt.shouldError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestCostTracking tests cost tracking functionality
func TestCostTracking(t *testing.T) {
	hourlyRate := 0.02232 // s-2vcpu-4gb DigitalOcean droplet
	maxSpend := 10.00

	tracker := cloud.NewCostTracker(hourlyRate, maxSpend)

	// Initial cost should be very close to 0
	initialCost := tracker.GetCurrentCost()
	assert.InDelta(t, 0.0, initialCost, 0.001)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Cost should have increased
	cost := tracker.GetCurrentCost()
	assert.Greater(t, cost, 0.0)

	// Elapsed time should be positive
	elapsed := tracker.GetElapsedTime()
	assert.Greater(t, elapsed, time.Duration(0))

	// Check limits (should pass with low spend)
	err := tracker.CheckLimits()
	assert.NoError(t, err)

	// Get summary
	summary := tracker.GetCostSummary()
	assert.Contains(t, summary, "Elapsed")
	assert.Contains(t, summary, "Cost")
	assert.Contains(t, summary, "Hourly Rate")

	// Test estimated final cost
	estimatedCost := tracker.GetEstimatedFinalCost(1 * time.Hour)
	assert.Equal(t, hourlyRate, estimatedCost)
}

// TestCostLimitCheck tests cost limit validation
func TestCostLimitCheck(t *testing.T) {
	// Use very high hourly rate to trigger limit quickly
	hourlyRate := 100000.0 // $100k/hour
	maxSpend := 0.01       // $0.01 limit

	tracker := cloud.NewCostTracker(hourlyRate, maxSpend)

	// Wait for cost to accumulate
	time.Sleep(100 * time.Millisecond)

	// Should exceed limit
	err := tracker.CheckLimits()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum total spend")
}

// TestInfrastructureState tests infrastructure state management
func TestInfrastructureState(t *testing.T) {
	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, "cloud-state")

	// Create a test infrastructure
	infra := &cloud.Infrastructure{
		ID:            "test-infra-12345",
		Provider:      "digitalocean",
		Mode:          "vm",
		CreatedAt:     time.Now(),
		PulumiStackID: "test-stack",
		Resources: []cloud.Resource{
			{
				Type:      "vm",
				ID:        "droplet-123",
				PublicIP:  "203.0.113.1",
				PrivateIP: "10.0.0.1",
				Status:    "active",
			},
		},
		Metadata: map[string]interface{}{
			"region": "nyc3",
			"size":   "s-2vcpu-4gb",
		},
	}

	// Save infrastructure state
	err := cloud.SaveInfrastructureState(infra, stateDir)
	require.NoError(t, err)

	// Verify state file was created
	stateFile := filepath.Join(stateDir, "infrastructure", infra.ID+".json")
	_, err = os.Stat(stateFile)
	require.NoError(t, err)

	// Check if infrastructure exists
	exists := cloud.InfrastructureExists(infra.ID, stateDir)
	assert.True(t, exists)

	// Load infrastructure state
	loadedInfra, err := cloud.LoadInfrastructureState(infra.ID, stateDir)
	require.NoError(t, err)

	// Verify loaded state matches original
	assert.Equal(t, infra.ID, loadedInfra.ID)
	assert.Equal(t, infra.Provider, loadedInfra.Provider)
	assert.Equal(t, infra.Mode, loadedInfra.Mode)
	assert.Equal(t, len(infra.Resources), len(loadedInfra.Resources))
	assert.Equal(t, infra.Resources[0].ID, loadedInfra.Resources[0].ID)

	// List all infrastructure states
	states, err := cloud.ListInfrastructures(stateDir)
	require.NoError(t, err)
	assert.Len(t, states, 1)
	assert.Equal(t, infra.ID, states[0].ID)

	// Remove infrastructure state
	err = cloud.RemoveInfrastructureState(infra.ID, stateDir)
	require.NoError(t, err)

	// Verify state file was deleted
	exists = cloud.InfrastructureExists(infra.ID, stateDir)
	assert.False(t, exists)

	_, err = os.Stat(stateFile)
	assert.True(t, os.IsNotExist(err))
}

// TestMultipleInfrastructureStates tests managing multiple infrastructure states
func TestMultipleInfrastructureStates(t *testing.T) {
	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, "cloud-state")

	// Create multiple infrastructure states
	infras := []*cloud.Infrastructure{
		{
			ID:       "infra-1",
			Provider: "digitalocean",
			Mode:     "vm",
		},
		{
			ID:       "infra-2",
			Provider: "aws",
			Mode:     "vm",
		},
		{
			ID:       "infra-3",
			Provider: "gcp",
			Mode:     "serverless",
		},
	}

	// Save all states
	for _, infra := range infras {
		err := cloud.SaveInfrastructureState(infra, stateDir)
		require.NoError(t, err)
	}

	// List all states
	states, err := cloud.ListInfrastructures(stateDir)
	require.NoError(t, err)
	assert.Len(t, states, 3)

	// Verify each state can be loaded
	for _, infra := range infras {
		loaded, err := cloud.LoadInfrastructureState(infra.ID, stateDir)
		require.NoError(t, err)
		assert.Equal(t, infra.ID, loaded.ID)
		assert.Equal(t, infra.Provider, loaded.Provider)
	}

	// Remove one state
	err = cloud.RemoveInfrastructureState("infra-2", stateDir)
	require.NoError(t, err)

	// Verify only 2 states remain
	states, err = cloud.ListInfrastructures(stateDir)
	require.NoError(t, err)
	assert.Len(t, states, 2)
}

// TestDefaultCloudConfig tests default configuration generation
func TestDefaultCloudConfig(t *testing.T) {
	cfg := config.DefaultCloudConfigs()

	// Verify defaults are set
	assert.Equal(t, "local", cfg.State.Backend)
	assert.Equal(t, "vm", cfg.Defaults.Mode)
	assert.Equal(t, "digitalocean", cfg.Defaults.Provider)
	assert.Greater(t, cfg.Defaults.MaxInstances, 0)
	assert.Greater(t, cfg.Limits.MaxInstances, 0)
	assert.NotEmpty(t, cfg.Defaults.Timeout)

	// Verify provider configs have placeholder values
	assert.Contains(t, cfg.Providers.DigitalOcean.Token, "${")
	assert.Contains(t, cfg.Providers.AWS.AccessKeyID, "${")
	assert.Contains(t, cfg.Providers.GCP.ProjectID, "${")

	// Verify sensible default regions
	assert.NotEmpty(t, cfg.Providers.DigitalOcean.Region)
	assert.NotEmpty(t, cfg.Providers.AWS.Region)
	assert.NotEmpty(t, cfg.Providers.GCP.Region)
}

// TestInfrastructureStateEmpty tests loading from empty state directory
func TestInfrastructureStateEmpty(t *testing.T) {
	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, "cloud-state")

	// List infrastructures from non-existent directory
	states, err := cloud.ListInfrastructures(stateDir)
	require.NoError(t, err)
	assert.Len(t, states, 0)

	// Check non-existent infrastructure
	exists := cloud.InfrastructureExists("non-existent-id", stateDir)
	assert.False(t, exists)

	// Try to load non-existent infrastructure
	_, err = cloud.LoadInfrastructureState("non-existent-id", stateDir)
	assert.Error(t, err)
}

// TestCloudConfigMultipleProviders tests configuring multiple cloud providers
func TestCloudConfigMultipleProviders(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "cloud-settings.yaml")

	// Create config with multiple providers
	cfg := &config.CloudConfigs{
		Providers: config.Providers{
			DigitalOcean: config.DigitalOceanConfig{
				Token:  "do-token-12345",
				Region: "nyc3",
				Size:   "s-2vcpu-4gb",
			},
			AWS: config.AWSConfig{
				AccessKeyID:     "aws-key-12345",
				SecretAccessKey: "aws-secret-12345",
				Region:          "us-east-1",
				InstanceType:    "t3.medium",
			},
			GCP: config.GCPConfig{
				ProjectID:       "gcp-project-12345",
				CredentialsFile: "/path/to/credentials.json",
				Region:          "us-central1",
				MachineType:     "n1-standard-2",
			},
		},
		Defaults: config.Defaults{
			Provider:     "digitalocean",
			MaxInstances: 10,
		},
		Limits: config.Limits{
			MaxHourlySpend: 5.0,
			MaxTotalSpend:  50.0,
			MaxInstances:   20,
		},
	}

	// Save configuration
	err := cloud.SaveCloudConfig(cfg, configPath)
	require.NoError(t, err)

	// Load configuration back
	loadedCfg, err := cloud.LoadCloudConfig(configPath)
	require.NoError(t, err)

	// Verify all providers are configured
	assert.Equal(t, cfg.Providers.DigitalOcean.Token, loadedCfg.Providers.DigitalOcean.Token)
	assert.Equal(t, cfg.Providers.AWS.AccessKeyID, loadedCfg.Providers.AWS.AccessKeyID)
	assert.Equal(t, cfg.Providers.GCP.ProjectID, loadedCfg.Providers.GCP.ProjectID)
}

// TestInfrastructureMetadata tests infrastructure metadata handling
func TestInfrastructureMetadata(t *testing.T) {
	tempDir := t.TempDir()
	stateDir := filepath.Join(tempDir, "cloud-state")

	// Create infrastructure with metadata
	infra := &cloud.Infrastructure{
		ID:       "test-infra-meta",
		Provider: "digitalocean",
		Mode:     "vm",
		Metadata: map[string]interface{}{
			"region":        "nyc3",
			"size":          "s-2vcpu-4gb",
			"worker_count":  3,
			"tags":          []string{"production", "web"},
			"custom_config": map[string]string{"key": "value"},
		},
	}

	// Save and reload
	err := cloud.SaveInfrastructureState(infra, stateDir)
	require.NoError(t, err)

	loadedInfra, err := cloud.LoadInfrastructureState(infra.ID, stateDir)
	require.NoError(t, err)

	// Verify metadata is preserved
	assert.Equal(t, "nyc3", loadedInfra.Metadata["region"])
	assert.Equal(t, "s-2vcpu-4gb", loadedInfra.Metadata["size"])
	assert.Equal(t, float64(3), loadedInfra.Metadata["worker_count"])
}

// TestCloudConfigValidation_AWS tests AWS-specific validation
func TestCloudConfigValidation_AWS(t *testing.T) {
	cfg := &config.CloudConfigs{
		Providers: config.Providers{
			AWS: config.AWSConfig{
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				Region:          "us-east-1",
			},
		},
		Defaults: config.Defaults{
			Provider:     "aws",
			MaxInstances: 5,
		},
		Limits: config.Limits{
			MaxHourlySpend: 10.0,
			MaxTotalSpend:  100.0,
			MaxInstances:   20,
		},
	}

	err := cloud.ValidateCloudConfig(cfg)
	assert.NoError(t, err)

	// Test with missing credentials
	cfg.Providers.AWS.AccessKeyID = ""
	err = cloud.ValidateCloudConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AWS credentials")
}

// TestCloudConfigValidation_GCP tests GCP-specific validation
func TestCloudConfigValidation_GCP(t *testing.T) {
	cfg := &config.CloudConfigs{
		Providers: config.Providers{
			GCP: config.GCPConfig{
				ProjectID:       "test-project",
				CredentialsFile: "/path/to/creds.json",
				Region:          "us-central1",
			},
		},
		Defaults: config.Defaults{
			Provider:     "gcp",
			MaxInstances: 5,
		},
		Limits: config.Limits{
			MaxHourlySpend: 10.0,
			MaxTotalSpend:  100.0,
			MaxInstances:   20,
		},
	}

	err := cloud.ValidateCloudConfig(cfg)
	assert.NoError(t, err)

	// Test with missing credentials
	cfg.Providers.GCP.ProjectID = ""
	err = cloud.ValidateCloudConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GCP credentials")
}

// TestCostEstimation tests cost estimation for different durations
func TestCostEstimation(t *testing.T) {
	hourlyRate := 0.02232 // DigitalOcean s-2vcpu-4gb
	maxSpend := 100.0

	tracker := cloud.NewCostTracker(hourlyRate, maxSpend)

	tests := []struct {
		duration     time.Duration
		expectedCost float64
	}{
		{1 * time.Hour, 0.02232},
		{24 * time.Hour, 0.53568},  // Daily
		{168 * time.Hour, 3.74976}, // Weekly (7 * 24 hours)
	}

	for _, tt := range tests {
		cost := tracker.GetEstimatedFinalCost(tt.duration)
		assert.InDelta(t, tt.expectedCost, cost, 0.00001)
	}
}
