package config

import (
	"github.com/goccy/go-yaml"
)

// ParseConfig parses configuration from YAML bytes
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ParseConfigStrict parses configuration with strict validation
func ParseConfigStrict(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.UnmarshalWithOptions(data, &cfg, yaml.Strict()); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ToYAML serializes the config to YAML bytes
func (c *Config) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate required fields
	if c.BaseFolder == "" {
		return &ConfigError{Field: "base_folder", Message: "base_folder is required"}
	}

	// Validate database config if server mode is expected
	if c.Database.Host == "" {
		c.Database.Host = "localhost"
	}
	if c.Database.Port == 0 {
		c.Database.Port = 5432
	}

	// Validate server config
	if c.Server.Port == 0 {
		c.Server.Port = 8002
	}

	// Set defaults for scan tactics
	if c.ScanTactic.Default == 0 {
		c.ScanTactic.Default = 10
	}
	if c.ScanTactic.Aggressive == 0 {
		c.ScanTactic.Aggressive = 40
	}
	if c.ScanTactic.Gently == 0 {
		c.ScanTactic.Gently = 5
	}

	// Set defaults for JWT (1 day = 1440 minutes)
	if c.Server.JWT.ExpirationMinutes == 0 {
		c.Server.JWT.ExpirationMinutes = 1440
	}

	// Set default for snapshot path
	if c.Environments.Snapshot == "" {
		c.Environments.Snapshot = "{{base_folder}}/snapshot"
	}

	return nil
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e *ConfigError) Error() string {
	return "config error: " + e.Field + " - " + e.Message
}
