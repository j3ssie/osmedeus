package cloud

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"gopkg.in/yaml.v3"
)

// LoadCloudConfig loads and parses the cloud configuration from the specified path
func LoadCloudConfig(configPath string) (*config.CloudConfigs, error) {
	// Expand home directory if needed
	if configPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cloud config: %w", err)
	}

	// Parse YAML
	var cfg config.CloudConfigs
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse cloud config: %w", err)
	}

	// Resolve environment variables in credentials
	if err := resolveEnvVars(&cfg); err != nil {
		return nil, fmt.Errorf("failed to resolve environment variables: %w", err)
	}

	return &cfg, nil
}

// SaveCloudConfig saves the cloud configuration to the specified path
func SaveCloudConfig(cfg *config.CloudConfigs, configPath string) error {
	// Expand home directory if needed
	if configPath[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal cloud config: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cloud config: %w", err)
	}

	return nil
}

// resolveEnvVars expands environment variables in configuration values
func resolveEnvVars(cfg *config.CloudConfigs) error {
	// AWS
	cfg.Providers.AWS.AccessKeyID = os.ExpandEnv(cfg.Providers.AWS.AccessKeyID)
	cfg.Providers.AWS.SecretAccessKey = os.ExpandEnv(cfg.Providers.AWS.SecretAccessKey)
	cfg.Providers.AWS.Region = os.ExpandEnv(cfg.Providers.AWS.Region)

	// GCP
	cfg.Providers.GCP.ProjectID = os.ExpandEnv(cfg.Providers.GCP.ProjectID)
	cfg.Providers.GCP.CredentialsFile = os.ExpandEnv(cfg.Providers.GCP.CredentialsFile)
	cfg.Providers.GCP.Region = os.ExpandEnv(cfg.Providers.GCP.Region)

	// DigitalOcean
	cfg.Providers.DigitalOcean.Token = os.ExpandEnv(cfg.Providers.DigitalOcean.Token)
	cfg.Providers.DigitalOcean.Region = os.ExpandEnv(cfg.Providers.DigitalOcean.Region)

	// Linode
	cfg.Providers.Linode.Token = os.ExpandEnv(cfg.Providers.Linode.Token)
	cfg.Providers.Linode.Region = os.ExpandEnv(cfg.Providers.Linode.Region)

	// Azure
	cfg.Providers.Azure.SubscriptionID = os.ExpandEnv(cfg.Providers.Azure.SubscriptionID)
	cfg.Providers.Azure.TenantID = os.ExpandEnv(cfg.Providers.Azure.TenantID)
	cfg.Providers.Azure.ClientID = os.ExpandEnv(cfg.Providers.Azure.ClientID)
	cfg.Providers.Azure.ClientSecret = os.ExpandEnv(cfg.Providers.Azure.ClientSecret)
	cfg.Providers.Azure.Location = os.ExpandEnv(cfg.Providers.Azure.Location)

	// SSH
	cfg.SSH.PrivateKeyPath = os.ExpandEnv(cfg.SSH.PrivateKeyPath)
	cfg.SSH.PrivateKeyContent = os.ExpandEnv(cfg.SSH.PrivateKeyContent)

	// State
	cfg.State.Path = os.ExpandEnv(cfg.State.Path)

	return nil
}

// ValidateCloudConfig validates the cloud configuration
func ValidateCloudConfig(cfg *config.CloudConfigs) error {
	// Validate provider credentials based on default provider
	switch cfg.Defaults.Provider {
	case "aws":
		if cfg.Providers.AWS.AccessKeyID == "" || cfg.Providers.AWS.SecretAccessKey == "" {
			return fmt.Errorf("AWS credentials not configured")
		}
	case "gcp":
		if cfg.Providers.GCP.ProjectID == "" || cfg.Providers.GCP.CredentialsFile == "" {
			return fmt.Errorf("GCP credentials not configured")
		}
	case "digitalocean":
		if cfg.Providers.DigitalOcean.Token == "" {
			return fmt.Errorf("DigitalOcean token not configured")
		}
	case "linode":
		if cfg.Providers.Linode.Token == "" {
			return fmt.Errorf("linode token not configured")
		}
	case "azure":
		if cfg.Providers.Azure.SubscriptionID == "" || cfg.Providers.Azure.ClientID == "" {
			return fmt.Errorf("azure credentials not configured")
		}
	default:
		return fmt.Errorf("invalid provider: %s", cfg.Defaults.Provider)
	}

	// Validate limits
	if cfg.Limits.MaxHourlySpend < 0 {
		return fmt.Errorf("max_hourly_spend must be non-negative")
	}
	if cfg.Limits.MaxTotalSpend < 0 {
		return fmt.Errorf("max_total_spend must be non-negative")
	}
	if cfg.Limits.MaxInstances < 1 {
		return fmt.Errorf("max_instances must be at least 1")
	}

	// Validate defaults
	if cfg.Defaults.MaxInstances < 1 {
		return fmt.Errorf("defaults.max_instances must be at least 1")
	}

	return nil
}
