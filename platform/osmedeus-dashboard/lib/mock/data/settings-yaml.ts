export const mockSettingsYaml = `# =============================================================================
# Osmedeus Configuration File
# =============================================================================

base_folder: ~/osmedeus-base

# -----------------------------------------------------------------------------
# Environment Configuration
# -----------------------------------------------------------------------------
environments:
  binaries_path: "{{base_folder}}/binaries"
  data_path: "{{base_folder}}/data"
  workflows_path: "{{base_folder}}/workflows"
  storages_path: "{{base_folder}}/storages"
  clouds_path: "{{base_folder}}/clouds"
  logs_path: "{{base_folder}}/logs"

# -----------------------------------------------------------------------------
# Server Configuration
# -----------------------------------------------------------------------------
server:
  host: "0.0.0.0"
  port: 8002
  enable_cors: true
  cors_origins:
    - "http://localhost:3000"
    - "http://127.0.0.1:3000"
  workspace_prefix_key: "[REDACTED]"
  simple_user_map_key: "[REDACTED]"
  jwt:
    secret_signing_key: "[REDACTED]"
    expiration_minutes: 180
    refresh_expiration_days: 7

# -----------------------------------------------------------------------------
# Database Configuration
# -----------------------------------------------------------------------------
database:
  host: "localhost"
  port: 5432
  name: "osmedeus"
  username: "[REDACTED]"
  password: "[REDACTED]"
  ssl_mode: "disable"
  max_idle_connections: 10
  max_open_connections: 100
  connection_max_lifetime: 3600

# -----------------------------------------------------------------------------
# Notification Configuration
# -----------------------------------------------------------------------------
notifications:
  slack:
    webhook_url: "[REDACTED]"
    channel: "#security-alerts"
    enabled: false
  discord:
    webhook_url: "[REDACTED]"
    enabled: false
  telegram:
    bot_token: "[REDACTED]"
    chat_id: "[REDACTED]"
    enabled: false

# -----------------------------------------------------------------------------
# Cloud Provider Configuration
# -----------------------------------------------------------------------------
cloud:
  provider: "digitalocean"
  digitalocean:
    api_token: "[REDACTED]"
    region: "nyc1"
    size: "s-2vcpu-4gb"
    image: "ubuntu-22-04-x64"
  aws:
    access_key_id: "[REDACTED]"
    secret_access_key: "[REDACTED]"
    region: "us-east-1"
  linode:
    api_token: "[REDACTED]"
    region: "us-east"

# -----------------------------------------------------------------------------
# Scan Configuration
# -----------------------------------------------------------------------------
scan:
  default_workflow: "general"
  max_concurrent_scans: 5
  timeout_minutes: 1440
  retry_failed_modules: true
  max_retries: 3

# -----------------------------------------------------------------------------
# LLM Configuration
# -----------------------------------------------------------------------------
llm:
  provider: "openai"
  openai:
    api_key: "[REDACTED]"
    model: "gpt-4"
    max_tokens: 4096
  anthropic:
    api_key: "[REDACTED]"
    model: "claude-3-sonnet-20240229"

# -----------------------------------------------------------------------------
# Logging Configuration
# -----------------------------------------------------------------------------
logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "{{base_folder}}/logs/osmedeus.log"
  max_size_mb: 100
  max_backups: 5
  max_age_days: 30
`;
