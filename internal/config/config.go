package config

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// exampleConfigYAML contains the default configuration template
// Source of truth: public/presets/osm-settings.example.yaml
var exampleConfigYAML = []byte(`# Osmedeus Configuration File
# This file contains all available configuration options for osmedeus.
# Copy this file to ~/osmedeus-base/osm-settings.yaml and customize as needed.

# =============================================================================
# Base Folder
# =============================================================================
# Root directory for all osmedeus data (workflows, binaries, data, etc.)
# Environment variables like $HOME are automatically expanded
base_folder: $HOME/osmedeus-base

# =============================================================================
# Environment Paths
# =============================================================================
# Directory paths for various osmedeus components
# Use {{base_folder}} to reference the base_folder value above
environments:
  # Path to binary executables (tools like nmap, ffuf, etc.)
  external_binaries_path: "{{base_folder}}/external-binaries"

  # Data directory for storing assets, wordlists, etc.
  external_data: "{{base_folder}}/external-data"

  # External configuration files (nuclei templates, etc.)
  external_configs: "{{base_folder}}/external-configs"

  # Output directory for scan workspaces
  # Each target gets its own subdirectory here
  workspaces: "$HOME/workspaces-osmedeus"

  # Directory containing workflow YAML files
  # Subdirectories: flows/, modules/
  workflows: "{{base_folder}}/workflows"

  # Directory for workspace snapshots (zip archives)
  # Used by the snapshot-download API endpoint
  snapshot: "{{base_folder}}/snapshot"

  # Directory for markdown report templates
  # Used by render_markdown_report() function
  markdown_report_templates: "{{base_folder}}/markdown-report-templates"

  # Directory for external agent configurations
  # Used for LLM Agent commands, skills, and related configurations
  external_agent_configs: "{{base_folder}}/external-agent-configs"

  # Directory for external utility scripts
  # Used for storing custom scripts and utilities
  external_scripts: "{{base_folder}}/external-scripts"

# =============================================================================
# Database Configuration
# =============================================================================
# Osmedeus supports SQLite (default) and PostgreSQL
database:
  # Database engine: "sqlite" or "postgresql"
  db_engine: sqlite

  # SQLite: Path to the database file
  # Ignored when using PostgreSQL
  db_path: "{{base_folder}}/database-osm.sqlite"

  # PostgreSQL connection settings
  # Only used when db_engine is "postgresql"
  host: localhost
  port: 5432
  username: osmedeus
  password: osmedeus
  db_name: osmedeus

  # Connection timeout in seconds
  connection_timeout: 60

  # PostgreSQL SSL mode: disable, require, verify-ca, verify-full
  ssl_mode: disable

# =============================================================================
# Server Configuration
# =============================================================================
# REST API server settings for the web interface
server:
  # Host to bind the server to
  # Use "0.0.0.0" to listen on all interfaces
  # Use "127.0.0.1" to listen only on localhost
  host: "0.0.0.0"

  # Port number for the API server
  port: 8002

  # Path to serve static UI files
  # Default: {{base_folder}}/ui/ - if this directory exists, it will be served at /ui
  # Set to empty string to disable UI serving
  ui_path: "{{base_folder}}/ui/"

  # Random prefix for workspace static files (auto-generated 16 chars if empty)
  # Used as URL path segment for direct access to workspaces folder
  workspace_prefix_key: ""

  # Authentication credentials (map of username:password)
  # Supports multiple users (password auto-generated if empty)
  simple_user_map_key:
    osmedeus: ""

  # JWT (JSON Web Token) settings
  jwt:
    # Secret key for signing JWT tokens (auto-generated if empty)
    secret_signing_key: ""

    # Token expiration time in minutes
    expiration_minutes: 60

  # License type shown in HTTP Server header and /server-info endpoint
  license: "open-source"

  # Enable Prometheus metrics endpoint at /metrics (default: true)
  # Set to false to disable metrics collection and endpoint
  enable_metrics: true

  # CORS allowed origins (default: "*" allows all origins)
  # Use comma-separated list for multiple origins: "https://example.com,https://app.example.com"
  cors_allowed_origins: "*"

  # API Key Authentication (alternative to JWT login flow)
  # When enabled, all API requests must include header: x-osm-api-key: <your-key>
  # This takes priority over JWT authentication when enabled
  enabled_auth_api: true
  auth_api_key: ""

# =============================================================================
# Scan Tactic Configuration
# =============================================================================
# Thread counts for different scan intensity levels
# Higher values = faster but more aggressive scans
# Lower values = slower but gentler on target systems
scan_tactic:
  # Aggressive/fast mode - maximum parallelism
  # Used with: osmedeus scan -t target --tactic aggressive
  aggressive: 40

  # Default/normal mode - balanced approach
  # Used when no tactic is specified
  default: 10

  # Gentle/thorough mode - minimal parallelism
  # Used with: osmedeus scan -t target --tactic gently
  gently: 5

# =============================================================================
# Redis Configuration (Optional)
# =============================================================================
# Redis is required for distributed scanning mode
# Leave host empty to disable Redis
redis:
  # Redis server hostname
  # Leave empty to disable distributed mode
  host: ""

  # Redis server port
  port: 6379

  # Redis authentication (if required)
  username: ""
  password: ""

  # Redis database number (0-15)
  db: 0

  # Connection timeout in seconds
  connection_timeout: 60

# =============================================================================
# Global Variables
# =============================================================================
# User-defined variables available in workflows via {{VARIABLE_NAME}}
# Variables can optionally be exported to environment variables
# Use _API_KEY suffix for secrets to indicate sensitive values
#
# Format:
#   VARIABLE_NAME:
#     value: "the-value"
#     as_env: true  # Optional: export as env var (default: true)
#
# Example usage in workflows:
#   - bash: "echo {{GITHUB_API_KEY}}"
#   - bash: "shodan search $SHODAN_API_KEY"  # Uses env var
global_vars:
  # GitHub personal access token for API access
  GITHUB_API_KEY:
    value: ""
    as_env: true  # Exports as GITHUB_API_KEY

  # Shodan API key for passive reconnaissance
  SHODAN_API_KEY:
    value: ""
    as_env: true  # Exports as SHODAN_API_KEY

  # Censys API key for certificate/host search
  CENSYS_API_KEY:
    value: ""
    as_env: true  # Exports as CENSYS_API_KEY

  # PassiveTotal API key for passive DNS/WHOIS
  PASSIVETOTAL_API_KEY:
    value: ""
    as_env: true  # Exports as PASSIVETOTAL_API_KEY

  # Add more API keys as needed (use _API_KEY suffix for secrets)

# =============================================================================
# Notification Configuration
# =============================================================================
# Send notifications when scans complete or find interesting results
notification:
  # Notification provider: "telegram" (future: slack, discord, webhook)
  provider: telegram

  # Master switch to enable/disable all notifications
  enabled: false

  # Telegram bot settings
  # Create a bot via @BotFather and get the token
  # Get your chat ID by messaging @userinfobot
  telegram:
    # Bot token from @BotFather
    bot_token: ""

    # Chat ID to send messages to (can be user or group)
    chat_id: 0

    # Enable Telegram notifications
    enabled: false

# =============================================================================
# Cloud Storage Configuration (Optional)
# =============================================================================
# S3-compatible storage for backing up scan results
# Supports AWS S3, MinIO, Cloudflare R2, Google Cloud Storage, DigitalOcean Spaces, Oracle OCI
storage:
  # Storage provider: "s3", "minio", "r2", "gcs", "spaces", "oci"
  provider: s3

  # Storage endpoint URL (auto-resolved for most providers)
  # Leave empty to auto-resolve based on provider and region/account_id
  endpoint: ""

  # Access credentials
  access_key_id: ""
  secret_access_key: ""

  # Bucket name for storing results
  bucket: ""

  # Cloud region (e.g., us-east-1, eu-west-1)
  region: us-east-1

  # Account ID or Namespace (R2: account ID, OCI: namespace)
  account_id: ""

  # Use SSL/TLS for connections
  use_ssl: true

  # Force path-style URLs
  path_style: false

  # Default presigned URL expiry (e.g., "1h", "30m")
  presign_expiry: "1h"

  # Enable cloud storage uploads
  enabled: false

# =============================================================================
# LLM Configuration (Optional)
# =============================================================================
# Large Language Model settings for AI-powered features
# Supports providers like Ollama, OpenAI, Anthropic, etc.
# Multiple providers can be configured for automatic rotation on error/rate limit
llm_config:
  # List of LLM providers (rotates to next on error/rate limit)
  llm_providers:
    # Primary provider (used first)
    - provider: ollama
      base_url: "http://localhost:11434/v1/chat/completions"
      auth_token: ""
      model: "gpt-oss:120b-cloud"
    # Backup provider example (uncomment to enable rotation)
    # - provider: openai
    #   base_url: "https://api.openai.com/v1/chat/completions"
    #   auth_token: "sk-your-api-key"
    #   model: "gpt-4"

  # Enable LLM tool call features
  enabled_tool_call: false

  # Maximum number of tokens to generate
  max_tokens: 1000

  # Temperature for sampling
  temperature: 0.7

  # Top-k sampling
  top_k: 50

  # Top-p sampling
  top_p: 0.9

  # Number of completions to generate
  n: 1

  # Maximum number of retries for failed requests
  max_retries: 3

  # Timeout for API requests
  timeout: 120s

  # Enable streaming responses
  stream: false

  # Enable structured JSON output format
  structured_json_format: false

  # System prompt for the LLM
  system_prompt: ""

  # Custom headers for API requests
  custom_headers: ""
`)

// generateRandomString generates a random alphanumeric string of the given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

// GlobalVar represents a single global variable with optional env export
type GlobalVar struct {
	Value string `yaml:"value"`
	AsEnv *bool  `yaml:"as_env,omitempty"` // pointer to distinguish unset (defaults true) from false
}

// IsAsEnv returns true if this variable should be exported to environment
// Defaults to true if not explicitly set
func (g GlobalVar) IsAsEnv() bool {
	if g.AsEnv == nil {
		return true // default
	}
	return *g.AsEnv
}

// GlobalVarsConfig holds all global variables
type GlobalVarsConfig map[string]GlobalVar

// NotificationConfig for multi-provider notifications
type NotificationConfig struct {
	Provider string          `yaml:"provider"` // telegram, webhook
	Enabled  bool            `yaml:"enabled"`
	Telegram TelegramConfig  `yaml:"telegram,omitempty"`
	Webhooks []WebhookConfig `yaml:"webhooks,omitempty"` // Multiple webhook endpoints
}

var (
	globalConfig *Config
	configMu     sync.RWMutex
)

// Config holds the complete application configuration
type Config struct {
	BaseFolder   string             `yaml:"base_folder"`
	Environments EnvironmentConfig  `yaml:"environments"`
	Database     DatabaseConfig     `yaml:"database"`
	Server       ServerConfig       `yaml:"server"`
	ScanTactic   ScanTacticConfig   `yaml:"scan_tactic"`
	Redis        RedisConfig        `yaml:"redis"`
	GlobalVars   GlobalVarsConfig   `yaml:"global_vars"`
	Notification NotificationConfig `yaml:"notification"`
	Storage      StorageConfig      `yaml:"storage"`
	LLM          LLMConfig          `yaml:"llm_config"`

	// Runtime paths (resolved from templates)
	BinariesPath                string `yaml:"-"`
	DataPath                    string `yaml:"-"`
	ConfigsPath                 string `yaml:"-"`
	WorkspacesPath              string `yaml:"-"`
	WorkflowsPath               string `yaml:"-"`
	UIPath                      string `yaml:"-"` // Resolved UI static files path
	SnapshotPath                string `yaml:"-"` // Resolved snapshot directory path
	MarkdownReportTemplatesPath string `yaml:"-"` // Resolved markdown report templates path
	ExternalAgentConfigsPath    string `yaml:"-"` // Resolved external agent configs path
	ExternalScriptsPath         string `yaml:"-"` // Resolved external scripts path
}

// EnvironmentConfig holds environment path configurations
type EnvironmentConfig struct {
	ExternalBinariesPath    string `yaml:"external_binaries_path"`
	ExternalData            string `yaml:"external_data"`
	ExternalConfigs         string `yaml:"external_configs"`
	Workspaces              string `yaml:"workspaces"`
	Workflows               string `yaml:"workflows"`
	Snapshot                string `yaml:"snapshot"`
	MarkdownReportTemplates string `yaml:"markdown_report_templates"`
	ExternalAgentConfigs    string `yaml:"external_agent_configs"`
	ExternalScripts         string `yaml:"external_scripts"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	DBEngine          string `yaml:"db_engine"` // sqlite, postgresql
	DBPath            string `yaml:"db_path"`   // SQLite file path
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	DBName            string `yaml:"db_name"`
	ConnectionTimeout int    `yaml:"connection_timeout"`
	SSLMode           string `yaml:"ssl_mode"`
}

// ServerConfig holds API server settings
type ServerConfig struct {
	Host               string            `yaml:"host"`
	Port               int               `yaml:"port"`
	UIPath             string            `yaml:"ui_path"`              // Path to serve static UI files
	WorkspacePrefixKey string            `yaml:"workspace_prefix_key"` // Random prefix for workspace static files (16 chars)
	SimpleUserMapKey   map[string]string `yaml:"simple_user_map_key"`  // Map of username:password for authentication
	JWT                JWTConfig         `yaml:"jwt"`                  // JWT settings
	License            string            `yaml:"license"`              // License type shown in ServerHeader and /server-info
	EnabledAuthAPI     bool              `yaml:"enabled_auth_api"`     // Enable API key authentication (default: false)
	AuthAPIKey         string            `yaml:"auth_api_key"`         // API key for x-osm-api-key header authentication
	EnableMetrics      *bool             `yaml:"enable_metrics,omitempty"`       // Enable Prometheus metrics endpoint (default: true)
	CORSAllowedOrigins string            `yaml:"cors_allowed_origins,omitempty"` // CORS allowed origins (default: "*")
	EventReceiverURL   string            `yaml:"event_receiver_url,omitempty"`   // URL for event receiver (auto-resolved from host:port if empty)
}

// IsMetricsEnabled returns true if the metrics endpoint should be enabled.
// Defaults to true if not explicitly set.
func (c *ServerConfig) IsMetricsEnabled() bool {
	if c.EnableMetrics == nil {
		return true
	}
	return *c.EnableMetrics
}

// GetCORSAllowedOrigins returns the configured CORS allowed origins.
// Defaults to "*" (all origins) if not explicitly set.
func (c *ServerConfig) GetCORSAllowedOrigins() string {
	if c.CORSAllowedOrigins == "" {
		return "*"
	}
	return c.CORSAllowedOrigins
}

// GetEventReceiverURL returns the event receiver URL.
// If EventReceiverURL is explicitly set, returns that.
// Otherwise, constructs URL from Host and Port if both are set.
// Returns empty string if neither option is available.
func (c *ServerConfig) GetEventReceiverURL() string {
	if c.EventReceiverURL != "" {
		return c.EventReceiverURL
	}
	if c.Host == "" || c.Port == 0 {
		return ""
	}
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

// ScanTacticConfig holds scan aggressiveness levels
type ScanTacticConfig struct {
	Aggressive int `yaml:"aggressive"`
	Default    int `yaml:"default"`
	Gently     int `yaml:"gently"`
}

// JWTConfig holds JWT settings
type JWTConfig struct {
	SecretSigningKey  string `yaml:"secret_signing_key"`
	ExpirationMinutes int    `yaml:"expiration_minutes"`
}

// RedisConfig holds Redis connection settings for distributed mode
type RedisConfig struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	DB                int    `yaml:"db"`
	ConnectionTimeout int    `yaml:"connection_timeout"`
}

// TelegramConfig holds Telegram bot settings
type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   int64  `yaml:"chat_id"`
	Enabled  bool   `yaml:"enabled"`
}

// WebhookConfig holds configuration for a single webhook endpoint
type WebhookConfig struct {
	URL           string            `yaml:"url"`
	Enabled       bool              `yaml:"enabled"`
	Headers       map[string]string `yaml:"headers,omitempty"`
	Timeout       int               `yaml:"timeout,omitempty"`         // seconds, default 30
	RetryCount    int               `yaml:"retry_count,omitempty"`     // default 3
	SkipTLSVerify bool              `yaml:"skip_tls_verify,omitempty"` // default false
	Events        []string          `yaml:"events,omitempty"`          // scan_complete, scan_failed, step_failed, etc.
}

// StorageConfig holds cloud storage settings (S3-compatible)
type StorageConfig struct {
	Provider        string `yaml:"provider"` // s3, minio, gcs, r2, spaces, oci
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Bucket          string `yaml:"bucket"`
	Region          string `yaml:"region"`
	UseSSL          bool   `yaml:"use_ssl"`
	Enabled         bool   `yaml:"enabled"`

	// New fields for provider-specific configuration
	AccountID     string `yaml:"account_id,omitempty"`     // For R2 (account ID) or OCI (namespace)
	PathStyle     bool   `yaml:"path_style,omitempty"`     // Force path-style URLs (needed for some providers)
	PresignExpiry string `yaml:"presign_expiry,omitempty"` // Default presign expiry (e.g., "1h", "30m")
}

// ProviderEndpoint holds endpoint template and defaults for a storage provider
type ProviderEndpoint struct {
	EndpointTemplate string
	UseSSL           bool
	PathStyle        bool
}

// ProviderEndpoints maps provider names to their endpoint configurations
var ProviderEndpoints = map[string]ProviderEndpoint{
	"r2":     {"%s.r2.cloudflarestorage.com", true, true},        // account_id
	"gcs":    {"storage.googleapis.com", true, false},            // HMAC keys
	"spaces": {"%s.digitaloceanspaces.com", true, false},         // region
	"oci":    {"%s.compat.objectstorage.%s.oraclecloud.com", true, true}, // namespace, region
	"s3":     {"s3.%s.amazonaws.com", true, false},               // region
	"minio":  {"", false, true},                                  // user-provided
}

// ResolveEndpoint resolves the endpoint URL based on provider type
func (c *StorageConfig) ResolveEndpoint() string {
	// If endpoint is explicitly set, use it
	if c.Endpoint != "" {
		return c.Endpoint
	}

	providerInfo, ok := ProviderEndpoints[c.Provider]
	if !ok {
		return c.Endpoint
	}

	switch c.Provider {
	case "r2":
		if c.AccountID != "" {
			return fmt.Sprintf(providerInfo.EndpointTemplate, c.AccountID)
		}
	case "gcs":
		return providerInfo.EndpointTemplate
	case "spaces":
		if c.Region != "" {
			return fmt.Sprintf(providerInfo.EndpointTemplate, c.Region)
		}
	case "oci":
		if c.AccountID != "" && c.Region != "" {
			return fmt.Sprintf(providerInfo.EndpointTemplate, c.AccountID, c.Region)
		}
	case "s3":
		if c.Region != "" {
			return fmt.Sprintf(providerInfo.EndpointTemplate, c.Region)
		}
	}

	return c.Endpoint
}

// GetPresignExpiry returns the presign expiry duration with default of 1 hour
func (c *StorageConfig) GetPresignExpiry() time.Duration {
	if c.PresignExpiry == "" {
		return time.Hour
	}

	d, err := time.ParseDuration(c.PresignExpiry)
	if err != nil {
		return time.Hour
	}
	return d
}

// ShouldUseSSL returns the SSL setting, considering provider defaults
func (c *StorageConfig) ShouldUseSSL() bool {
	// If explicitly set in config, use that
	if c.UseSSL {
		return true
	}

	// Check provider defaults
	if providerInfo, ok := ProviderEndpoints[c.Provider]; ok {
		return providerInfo.UseSSL
	}

	return c.UseSSL
}

// ShouldUsePathStyle returns whether path-style URLs should be used
func (c *StorageConfig) ShouldUsePathStyle() bool {
	// If explicitly set in config, use that
	if c.PathStyle {
		return true
	}

	// Check provider defaults
	if providerInfo, ok := ProviderEndpoints[c.Provider]; ok {
		return providerInfo.PathStyle
	}

	return c.PathStyle
}

// TemplateEngineConfig holds configuration for the template engine
type TemplateEngineConfig struct {
	// UseShardedEngine enables the sharded template engine for better concurrency
	// under high parallelism (foreach loops with multiple workers, parallel steps)
	// Default: true
	UseShardedEngine *bool `yaml:"use_sharded_engine,omitempty"`

	// ShardCount is the number of cache shards (must be power of 2)
	// Higher values reduce lock contention but increase memory usage
	// Default: 16
	ShardCount int `yaml:"shard_count,omitempty"`

	// ShardCacheSize is the LRU cache size per shard
	// Total cache capacity = ShardCount * ShardCacheSize
	// Default: 64 (total 1024 templates)
	ShardCacheSize int `yaml:"shard_cache_size,omitempty"`

	// EnablePooling enables sync.Pool for context map reuse
	// Reduces GC pressure during high-throughput rendering
	// Default: true
	EnablePooling *bool `yaml:"enable_pooling,omitempty"`

	// EnableBatch enables batch template rendering optimization
	// Groups templates by shard to minimize lock acquisitions
	// Default: true
	EnableBatch *bool `yaml:"enable_batch,omitempty"`
}

// IsShardedEngineEnabled returns whether sharded engine should be used
// Defaults to true if not explicitly set
func (c *TemplateEngineConfig) IsShardedEngineEnabled() bool {
	if c.UseShardedEngine == nil {
		return true
	}
	return *c.UseShardedEngine
}

// IsPoolingEnabled returns whether context pooling is enabled
// Defaults to true if not explicitly set
func (c *TemplateEngineConfig) IsPoolingEnabled() bool {
	if c.EnablePooling == nil {
		return true
	}
	return *c.EnablePooling
}

// IsBatchEnabled returns whether batch rendering is enabled
// Defaults to true if not explicitly set
func (c *TemplateEngineConfig) IsBatchEnabled() bool {
	if c.EnableBatch == nil {
		return true
	}
	return *c.EnableBatch
}

// GetShardCount returns the shard count with default
func (c *TemplateEngineConfig) GetShardCount() int {
	if c.ShardCount <= 0 {
		return 16
	}
	return c.ShardCount
}

// GetShardCacheSize returns the shard cache size with default
func (c *TemplateEngineConfig) GetShardCacheSize() int {
	if c.ShardCacheSize <= 0 {
		return 64
	}
	return c.ShardCacheSize
}

// LLMProvider holds configuration for a single LLM provider endpoint
type LLMProvider struct {
	Provider  string `yaml:"provider"`   // ollama, openai, anthropic, custom, etc.
	BaseURL   string `yaml:"base_url"`   // API endpoint URL
	AuthToken string `yaml:"auth_token"` // Authentication token (can be blank for local Ollama)
	Model     string `yaml:"model"`      // Model name/ID
}

// LLMConfig holds LLM (Large Language Model) settings
type LLMConfig struct {
	LLMProviders         []LLMProvider `yaml:"llm_providers"`          // List of LLM providers for rotation
	EnabledToolCall      bool          `yaml:"enabled_tool_call"`      // Enable LLM tool call features
	MaxTokens            int           `yaml:"max_tokens"`             // Maximum number of tokens to generate
	Temperature          float64       `yaml:"temperature"`            // Temperature for sampling
	TopK                 int           `yaml:"top_k"`                  // Top-k sampling
	TopP                 float64       `yaml:"top_p"`                  // Top-p sampling
	N                    int           `yaml:"n"`                      // Number of completions to generate
	MaxRetries           int           `yaml:"max_retries"`            // Maximum number of retries for failed requests
	Timeout              string        `yaml:"timeout"`                // Timeout for API requests
	Stream               bool          `yaml:"stream"`                 // Enable streaming responses
	StructuredJSONFormat bool          `yaml:"structured_json_format"` // Enable structured JSON output format
	SystemPrompt         string        `yaml:"system_prompt"`          // System prompt for the LLM
	CustomHeaders        string        `yaml:"custom_headers"`         // Custom headers for API requests

	// Internal fields for provider rotation (not serialized)
	currentIndex int        // Current provider index
	mu           sync.Mutex // Mutex for thread-safe rotation
}

// Load loads configuration from the specified base folder
func Load(baseFolder string) (*Config, error) {
	settingsPath := filepath.Join(baseFolder, "osm-settings.yaml")
	cfg, err := LoadFromFile(settingsPath)
	if err != nil {
		return nil, err
	}

	// Override base folder if provided
	if baseFolder != "" {
		cfg.BaseFolder = baseFolder
	}

	// Resolve environment paths
	cfg.ResolvePaths()

	return cfg, nil
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseConfig(data)
}

// LoadFromBytes loads configuration from raw YAML bytes
func LoadFromBytes(data []byte) (*Config, error) {
	return ParseConfig(data)
}

// ResolvePaths resolves template variables in environment paths.
// This method should be called after changing BaseFolder to recalculate all derived paths.
func (c *Config) ResolvePaths() {
	baseFolder := c.resolveEnvVars(c.BaseFolder)
	c.BaseFolder = baseFolder

	// Resolve each path with base_folder substitution
	c.BinariesPath = c.resolvePath(c.Environments.ExternalBinariesPath, baseFolder)
	c.DataPath = c.resolvePath(c.Environments.ExternalData, baseFolder)
	c.ConfigsPath = c.resolvePath(c.Environments.ExternalConfigs, baseFolder)
	c.WorkspacesPath = c.resolvePath(c.Environments.Workspaces, baseFolder)
	c.WorkflowsPath = c.resolvePath(c.Environments.Workflows, baseFolder)

	// Resolve server paths
	c.UIPath = c.resolvePath(c.Server.UIPath, baseFolder)

	// Resolve snapshot path
	c.SnapshotPath = c.resolvePath(c.Environments.Snapshot, baseFolder)

	// Resolve markdown report templates path
	c.MarkdownReportTemplatesPath = c.resolvePath(c.Environments.MarkdownReportTemplates, baseFolder)

	// Resolve external agent configs path
	c.ExternalAgentConfigsPath = c.resolvePath(c.Environments.ExternalAgentConfigs, baseFolder)

	// Resolve external scripts path
	c.ExternalScriptsPath = c.resolvePath(c.Environments.ExternalScripts, baseFolder)
}

// resolvePath resolves a single path with variable substitution
func (c *Config) resolvePath(path, baseFolder string) string {
	if path == "" {
		return ""
	}

	// Replace {{base_folder}} template variable
	resolved := strings.ReplaceAll(path, "{{base_folder}}", baseFolder)

	// Resolve environment variables like $HOME
	resolved = c.resolveEnvVars(resolved)

	return resolved
}

// resolveEnvVars resolves environment variables in a string
func (c *Config) resolveEnvVars(s string) string {
	return os.ExpandEnv(s)
}

// GetWorkflowsDir returns the workflows directory path
func (c *Config) GetWorkflowsDir() string {
	return c.WorkflowsPath
}

// GetModulesDir returns the modules directory path
func (c *Config) GetModulesDir() string {
	return filepath.Join(c.WorkflowsPath, "modules")
}

// GetWorkspacesDir returns the workspaces directory path
func (c *Config) GetWorkspacesDir() string {
	return c.WorkspacesPath
}

// GetDBPath returns the resolved database file path for SQLite
func (c *Config) GetDBPath() string {
	if c.Database.DBPath == "" {
		return filepath.Join(c.BaseFolder, "database-osm.sqlite")
	}
	return c.resolvePath(c.Database.DBPath, c.BaseFolder)
}

// IsSQLite returns true if the database engine is SQLite
func (c *Config) IsSQLite() bool {
	return c.Database.DBEngine == "" || c.Database.DBEngine == "sqlite"
}

// IsPostgres returns true if the database engine is PostgreSQL
func (c *Config) IsPostgres() bool {
	return c.Database.DBEngine == "postgresql" || c.Database.DBEngine == "postgres"
}

// IsRedisConfigured returns true if Redis is configured
func (c *Config) IsRedisConfigured() bool {
	return c.Redis.Host != "" && c.Redis.Port > 0
}

// IsDistributedMode returns true if Redis is configured (distributed mode enabled)
func (c *Config) IsDistributedMode() bool {
	return c.IsRedisConfigured()
}

// GetRedisAddr returns the Redis address in host:port format
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// =============================================================================
// Distributed Mode State
// =============================================================================

// WorkerModeState tracks whether we're running as a distributed worker
var workerModeState struct {
	isWorker bool
	workerID string
}

// SetWorkerMode sets the distributed worker mode flag
func SetWorkerMode(isWorker bool, workerID string) {
	workerModeState.isWorker = isWorker
	workerModeState.workerID = workerID
}

// IsWorkerMode returns true if running as a distributed worker
func IsWorkerMode() bool {
	return workerModeState.isWorker
}

// GetWorkerID returns the worker ID if in worker mode
func GetWorkerID() string {
	return workerModeState.workerID
}

// ShouldUseRedisDataQueues returns true if database writes should be routed to Redis queues.
// This is true when: Redis is configured AND we're running in worker mode.
func ShouldUseRedisDataQueues() bool {
	cfg := Get()
	if cfg == nil {
		return false
	}
	return cfg.IsDistributedMode() && IsWorkerMode()
}

// GetDSN returns the PostgreSQL connection string
func (c *Config) GetDSN() string {
	sslMode := c.Database.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.Username, c.Database.Password,
		c.Database.Host, c.Database.Port,
		c.Database.DBName, sslMode)
}

// IsNotificationConfigured returns true if any notification provider is configured and enabled
func (c *Config) IsNotificationConfigured() bool {
	if !c.Notification.Enabled {
		return false
	}
	switch c.Notification.Provider {
	case "telegram":
		return c.Notification.Telegram.BotToken != "" && c.Notification.Telegram.ChatID != 0
	default:
		return false
	}
}

// IsTelegramConfigured returns true if Telegram is configured and enabled
// Deprecated: Use IsNotificationConfigured instead
func (c *Config) IsTelegramConfigured() bool {
	return c.Notification.Provider == "telegram" &&
		c.Notification.Enabled &&
		c.Notification.Telegram.BotToken != "" &&
		c.Notification.Telegram.ChatID != 0
}

// GetGlobalVar returns a global variable value by name
func (c *Config) GetGlobalVar(name string) (string, bool) {
	if c.GlobalVars == nil {
		return "", false
	}
	if v, ok := c.GlobalVars[name]; ok {
		return v.Value, true
	}
	return "", false
}

// ExportGlobalVarsToEnv exports variables with as_env=true to environment
// Variable names are converted to UPPERCASE_WITH_UNDERSCORES
func (c *Config) ExportGlobalVarsToEnv() {
	if c.GlobalVars == nil {
		return
	}
	for name, v := range c.GlobalVars {
		if v.IsAsEnv() && v.Value != "" {
			envName := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
			_ = os.Setenv(envName, v.Value)
		}
	}
}

// GetAllGlobalVars returns all global vars as map[string]string for templates
func (c *Config) GetAllGlobalVars() map[string]string {
	result := make(map[string]string)
	if c.GlobalVars == nil {
		return result
	}
	for name, v := range c.GlobalVars {
		result[name] = v.Value
	}
	return result
}

// IsStorageConfigured returns true if cloud storage is configured and enabled
func (c *Config) IsStorageConfigured() bool {
	return c.Storage.Enabled && c.Storage.Endpoint != "" && c.Storage.Bucket != ""
}

// IsLLMConfigured returns true if LLM is configured and enabled
func (c *Config) IsLLMConfigured() bool {
	return c.LLM.EnabledToolCall && len(c.LLM.LLMProviders) > 0
}

// GetCurrentProvider returns the current active LLM provider (thread-safe)
// Returns nil if no providers are configured
func (l *LLMConfig) GetCurrentProvider() *LLMProvider {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.LLMProviders) == 0 {
		return nil
	}
	return &l.LLMProviders[l.currentIndex]
}

// RotateProvider advances to the next LLM provider (thread-safe, wraps around)
// Returns the new current provider, or nil if no providers are configured
func (l *LLMConfig) RotateProvider() *LLMProvider {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.LLMProviders) == 0 {
		return nil
	}
	l.currentIndex = (l.currentIndex + 1) % len(l.LLMProviders)
	return &l.LLMProviders[l.currentIndex]
}

// ResetProviderIndex resets the current provider to the first one (thread-safe)
func (l *LLMConfig) ResetProviderIndex() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.currentIndex = 0
}

// GetProviderCount returns the number of configured LLM providers
func (l *LLMConfig) GetProviderCount() int {
	return len(l.LLMProviders)
}

// GetCurrentProviderIndex returns the current provider index (thread-safe)
func (l *LLMConfig) GetCurrentProviderIndex() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.currentIndex
}

// ResolveServerCredentials adds credentials from environment variables
// OSM_USERNAME and OSM_PASSWORD are added to the user map if both are set
func (c *Config) ResolveServerCredentials() {
	envUser := os.Getenv("OSM_USERNAME")
	envPass := os.Getenv("OSM_PASSWORD")
	if envUser != "" && envPass != "" {
		if c.Server.SimpleUserMapKey == nil {
			c.Server.SimpleUserMapKey = make(map[string]string)
		}
		c.Server.SimpleUserMapKey[envUser] = envPass
	}
}

// GetThreads returns thread counts for the given scan tactic
// Returns (threads, baseThreads) where baseThreads is half of threads
func (c *Config) GetThreads(tactic string) (int, int) {
	var threads int
	switch tactic {
	case "aggressive", "fast":
		threads = c.ScanTactic.Aggressive
	case "gently", "thorough":
		threads = c.ScanTactic.Gently
	default: // normal, default
		threads = c.ScanTactic.Default
	}
	if threads <= 0 {
		threads = 10 // fallback default
	}
	baseThreads := threads / 2
	if baseThreads < 1 {
		baseThreads = 1
	}
	return threads, baseThreads
}

// Set sets the global configuration
func Set(cfg *Config) {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = cfg
}

// Get returns the global configuration
func Get() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	baseFolder := filepath.Join(homeDir, "osmedeus-base")

	return &Config{
		BaseFolder: baseFolder,
		Environments: EnvironmentConfig{
			ExternalBinariesPath: "{{base_folder}}/external-binaries",
			ExternalData:         "{{base_folder}}/external-data",
			ExternalConfigs:      "{{base_folder}}/external-configs",
			Workspaces:           "{{base_folder}}/workspaces",
			Workflows:            "{{base_folder}}/workflows",
			Snapshot:             "{{base_folder}}/snapshot",
			ExternalScripts:      "{{base_folder}}/external-scripts",
		},
		Database: DatabaseConfig{
			DBEngine:          "sqlite",
			DBPath:            "{{base_folder}}/database-osm.sqlite",
			Host:              "localhost",
			Port:              5432,
			Username:          "osmedeus",
			Password:          "osmedeus",
			DBName:            "osmedeus",
			ConnectionTimeout: 60,
			SSLMode:           "disable",
		},
		Server: ServerConfig{
			Host:               "0.0.0.0",
			Port:               8002,
			UIPath:             "{{base_folder}}/ui/",
			WorkspacePrefixKey: generateRandomString(16),
			SimpleUserMapKey: map[string]string{
				"osmedeus": generateRandomString(12),
			},
			JWT: JWTConfig{
				SecretSigningKey:  generateRandomString(64),
				ExpirationMinutes: 60,
			},
			License:        "open-source",
			EnabledAuthAPI: true,
			AuthAPIKey:     generateRandomString(32),
		},
		ScanTactic: ScanTacticConfig{
			Aggressive: 40,
			Default:    10,
			Gently:     5,
		},
		Redis: RedisConfig{
			Host:              "",
			Port:              6379,
			Username:          "",
			Password:          "",
			DB:                0,
			ConnectionTimeout: 60,
		},
		GlobalVars: GlobalVarsConfig{
			"GITHUB_API_KEY":       {Value: ""},
			"SHODAN_API_KEY":       {Value: ""},
			"CENSYS_API_KEY":       {Value: ""},
			"PASSIVETOTAL_API_KEY": {Value: ""},
			// Add more default placeholders as needed
		},
		Notification: NotificationConfig{
			Provider: "telegram",
			Enabled:  false,
			Telegram: TelegramConfig{
				BotToken: "",
				ChatID:   0,
				Enabled:  false,
			},
		},
		Storage: StorageConfig{
			Provider:        "s3",
			Endpoint:        "",
			AccessKeyID:     "",
			SecretAccessKey: "",
			Bucket:          "",
			Region:          "us-east-1",
			UseSSL:          true,
			Enabled:         false,
		},
		LLM: LLMConfig{
			LLMProviders: []LLMProvider{
				{
					Provider:  "ollama",
					BaseURL:   "http://localhost:11434/v1/chat/completions",
					AuthToken: "",
					Model:     "gpt-oss:120b-cloud",
				},
			},
			EnabledToolCall:      false,
			MaxTokens:            1000,
			Temperature:          0.7,
			TopK:                 50,
			TopP:                 0.9,
			N:                    1,
			MaxRetries:           3,
			Timeout:              "120s",
			Stream:               false,
			StructuredJSONFormat: false,
			SystemPrompt:         "",
			CustomHeaders:        "",
		},
	}
}

// EnsureConfigExists creates osm-settings.yaml if it doesn't exist
// Uses the embedded example configuration file as template
func EnsureConfigExists(baseFolder string) error {
	settingsPath := filepath.Join(baseFolder, "osm-settings.yaml")

	// Check if file already exists
	if _, err := os.Stat(settingsPath); err == nil {
		return nil // File exists, nothing to do
	}

	// Create base folder if needed
	if err := os.MkdirAll(baseFolder, 0755); err != nil {
		return err
	}

	// Generate random values and replace blank placeholders in template
	configContent := string(exampleConfigYAML)
	configContent = strings.Replace(configContent,
		"workspace_prefix_key: \"\"",
		fmt.Sprintf("workspace_prefix_key: \"%s\"", generateRandomString(16)),
		1)

	// Generate random auth_api_key (32 chars)
	configContent = strings.Replace(configContent,
		"auth_api_key: \"\"",
		fmt.Sprintf("auth_api_key: \"%s\"", generateRandomString(32)),
		1)

	// Generate random secret_signing_key (64 chars)
	configContent = strings.Replace(configContent,
		"secret_signing_key: \"\"",
		fmt.Sprintf("secret_signing_key: \"%s\"", generateRandomString(64)),
		1)

	// Generate random password for default osmedeus user (12 chars)
	configContent = strings.Replace(configContent,
		"osmedeus: \"\"",
		fmt.Sprintf("osmedeus: \"%s\"", generateRandomString(12)),
		1)

	// Write the config file with generated values
	return os.WriteFile(settingsPath, []byte(configContent), 0644)
}
