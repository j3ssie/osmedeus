package cli

import (
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
)

// UsageRoot returns the Long description for the root command
func UsageRoot() string {
	return terminal.BoldCyan("◆ Description") + `
  Osmedeus is a powerful workflow engine for executing automated
  reconnaissance and security assessment workflows.

  It supports both module (single execution units) and flow (multi-module
  orchestration) workflows with parallel and sequential execution patterns.

` + terminal.BoldCyan("▶ Key Features") + `
  • Execute YAML-defined security workflows
  • Support for parallel and sequential execution
  • Distributed scanning with master/worker architecture
  • Template variables and utility functions

` + terminal.BoldCyan("▷ Quick Start") + `
  ` + terminal.Green("# Run a module workflow") + `
  osmedeus run ` + terminal.Yellow("-m") + ` simple-module ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Run a flow workflow") + `
  osmedeus run ` + terminal.Yellow("-f") + ` recon-workflow ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Evaluate a utility function") + `
  osmedeus func e 'log_info("Hello {{target}}")' ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# List available workflows") + `
  osmedeus workflow list

  ` + terminal.Green("# Show all usage examples") + `
  osmedeus ` + terminal.Yellow("--usage-example") + `

` + docsFooter()
}

// UsageRun returns the Long description for the run command
func UsageRun() string {
	return terminal.BoldCyan("◆ Description") + `
  Execute a workflow against one or more targets.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Run against a single target") + `
  osmedeus run ` + terminal.Yellow("-f") + ` recon-workflow ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Run against multiple targets") + `
  osmedeus run ` + terminal.Yellow("-m") + ` simple-module ` + terminal.Yellow("-t") + ` target1.com ` + terminal.Yellow("-t") + ` target2.com

  ` + terminal.Green("# Run with stdin input with concurrency") + `
  cat list-of-urls.txt | osmedeus run ` + terminal.Yellow("-m") + ` simple-module ` + terminal.Yellow("--concurrency") + ` 10

  ` + terminal.Green("# Combine multiple input methods") + `
  echo "extra.com" | osmedeus run ` + terminal.Yellow("-m") + ` simple-module ` + terminal.Yellow("-t") + ` main.com ` + terminal.Yellow("-T") + ` more-targets.txt

  ` + terminal.Green("# Run with custom parameters") + `
  osmedeus run ` + terminal.Yellow("-m") + ` simple-module ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--params") + ` 'threads=20'

  ` + terminal.Green("# Run with custom base folder") + `
  osmedeus run ` + terminal.Yellow("--base-folder") + ` /opt/osmedeus-base ` + terminal.Yellow("-f") + ` recon-workflow ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Run with timeout (cancel if exceeds)") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--timeout") + ` 2h

  ` + terminal.Green("# Repeat run every hour continuously") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--repeat") + ` ` + terminal.Yellow("--repeat-wait-time") + ` 1h

  ` + terminal.Green("# Run multiple modules in sequence") + `
  osmedeus run ` + terminal.Yellow("-m") + ` subdomain ` + terminal.Yellow("-m") + ` portscan ` + terminal.Yellow("-m") + ` vulnscan ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Combine timeout with repeat mode") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--timeout") + ` 3h ` + terminal.Yellow("--repeat") + ` ` + terminal.Yellow("--repeat-wait-time") + ` 30m

  ` + terminal.Green("# Dry-run mode (preview without executing)") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--dry-run") + `

  ` + terminal.Green("# Run module from stdin (pipe YAML)") + `
  cat module.yaml | osmedeus run ` + terminal.Yellow("--std-module") + ` ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Load parameters from YAML/JSON file") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--params-file") + ` params.yaml

  ` + terminal.Green("# Custom workspace path") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--workspace") + ` /custom/workspace

  ` + terminal.Green("# Skip heuristics checks") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--heuristics-check") + ` none

  ` + terminal.Green("# Concurrent targets from file") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--concurrency") + ` 5

` + docsFooter()
}

// UsageServe returns the Long description for the serve command
func UsageServe() string {
	return terminal.BoldCyan("◆ Description") + `
  Start the Osmedeus web server that provides REST API endpoints.

` + terminal.BoldCyan("▶ Features") + `
  • REST API for managing runs
  • Workflow listing and management
  • Real-time run progress via WebSocket
  • Settings management

  Use ` + terminal.Yellow("--master") + ` to run as a distributed master node that coordinates
  workers connected via Redis.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Start server with default settings") + `
  osmedeus serve

  ` + terminal.Green("# Start server on custom port") + `
  osmedeus serve ` + terminal.Yellow("--port") + ` 8080

  ` + terminal.Green("# Start server without authentication (development only)") + `
  osmedeus serve ` + terminal.Yellow("-A") + `

  ` + terminal.Green("# Start server on specific host without auth") + `
  osmedeus serve ` + terminal.Yellow("--host") + ` 127.0.0.1 ` + terminal.Yellow("--port") + ` 8811 ` + terminal.Yellow("-A") + `

  ` + terminal.Green("# Start as distributed master node") + `
  osmedeus serve ` + terminal.Yellow("--master") + `

` + docsFooter()
}

// UsageWorkflow returns the Long description for the workflow command
func UsageWorkflow() string {
	return terminal.BoldCyan("◆ Description") + `
  Commands for listing, viewing, and validating workflows.

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("list") + `      - List available workflows
  • ` + terminal.Yellow("show") + `      - Show workflow details
  • ` + terminal.Yellow("validate") + `  - Validate a workflow

` + terminal.BoldCyan("▶ Workflow Preferences") + `
  Workflows can define execution preferences in YAML that act as defaults.
  CLI flags always take precedence over workflow preferences.

  ` + terminal.Yellow("preferences:") + `
    ` + terminal.Gray("disable_notifications: false") + `   # --disable-notification
    ` + terminal.Gray("disable_logging: true") + `          # --disable-logging
    ` + terminal.Gray("heuristics_check: 'basic'") + `      # --heuristics-check
    ` + terminal.Gray("ci_output_format: true") + `         # --ci-output-format
    ` + terminal.Gray("silent: true") + `                   # --silent
    ` + terminal.Gray("repeat: true") + `                   # --repeat
    ` + terminal.Gray("repeat_wait_time: '60s'") + `        # --repeat-wait-time

` + docsFooter()
}

// UsageFunction returns the Long description for the function command
func UsageFunction() string {
	return terminal.BoldCyan("◆ Description") + `
  Execute and test utility functions available in workflows.

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("list") + `       - List all available functions
  • ` + terminal.Yellow("eval (e)") + `   - Evaluate scripts with template rendering

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# List all available functions") + `
  osmedeus func list

  ` + terminal.Green("# Evaluate a simple function") + `
  osmedeus func eval 'trim("  hello  ")'

  ` + terminal.Green("# Short alias for eval") + `
  osmedeus func e 'log_info("Hello World")'

  ` + terminal.Green("# Use with target variable") + `
  osmedeus func e 'fileExists("{{target}}")' ` + terminal.Yellow("-t") + ` /tmp/test.txt

  ` + terminal.Green("# Print markdown file with syntax highlighting") + `
  osmedeus func e 'print_markdown_from_file("README.md")'

  ` + terminal.Green("# Multi-line script with variable") + `
  osmedeus func e 'var x = trim("  test  "); log_info(x); x'

  ` + terminal.Green("# Make HTTP request") + `
  osmedeus func e 'httpRequest("https://api.example.com", "GET", {}, "")'

  ` + terminal.Green("# With custom params") + `
  osmedeus func e 'log_info("{{host}}:{{port}}")' ` + terminal.Yellow("--params") + ` 'host=localhost' ` + terminal.Yellow("--params") + ` 'port=8080'

  ` + terminal.Green("# Use -f flag for shell path autocompletion on file arguments") + `
  osmedeus func e ` + terminal.Yellow("-f") + ` trim "  hello world  "
  osmedeus func e ` + terminal.Yellow("-f") + ` fileExists /tmp
  osmedeus func e ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("-f") + ` log_info "Processing {{target}}"

  ` + terminal.Green("# Query database with SQL") + `
  osmedeus func e 'db_select("SELECT severity, COUNT(*) FROM vulnerabilities GROUP BY severity", "markdown")'

  ` + terminal.Green("# Query filtered assets from database") + `
  osmedeus func e 'db_select_assets_filtered("example.com", 200, "subdomain", "jsonl")'

  ` + terminal.Green("# Read script from stdin") + `
  echo 'log_info("hello")' | osmedeus func e ` + terminal.Yellow("--stdin") + `

  ` + terminal.Green("# Alternative stdin syntax") + `
  echo 'trim("  test  ")' | osmedeus func e -

` + docsFooter()
}

// UsageFunctionEval returns the Long description for the function eval command
func UsageFunctionEval() string {
	return terminal.BoldCyan("◆ Description") + `
  Evaluate a script with template rendering and function execution.

` + terminal.BoldCyan("▶ Processing Phases") + `
  1. Template variables ({{target}}, {{custom}}) are rendered
  2. The result is executed as JavaScript with access to utility functions

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Print markdown file with syntax highlighting") + `
  osmedeus func e 'print_markdown_from_file("README.md")'

  ` + terminal.Green("# Log a message with INFO prefix") + `
  osmedeus func e 'log_info("Scan completed for {{target}}")' ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Save content to file") + `
  osmedeus func e 'save_content(render_markdown_from_file("README.md"), "/tmp/output.txt")'

  ` + terminal.Green("# Use with variable") + `
  osmedeus func e 'var content = "sample"; save_content(content, "/tmp/output.txt")'


  ` + terminal.Green("# Sort a file using Unix sort") + `
  osmedeus func e 'sortUnix("/tmp/input.txt", "/tmp/sorted.txt")'

  ` + terminal.Green("# Make HTTP request") + `
  osmedeus func e 'httpRequest("https://api.example.com", "GET", {}, "")'

  ` + terminal.Green("# With custom params") + `
  osmedeus func e 'log_info("{{host}}:{{port}}")' ` + terminal.Yellow("--params") + ` 'host=localhost' ` + terminal.Yellow("--params") + ` 'port=8080'

  ` + terminal.Green("# Use -f flag for shell path autocompletion on file arguments") + `
  osmedeus func e ` + terminal.Yellow("-f") + ` trim "  hello world  "
  osmedeus func e ` + terminal.Yellow("-f") + ` fileExists /tmp
  osmedeus func e ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("-f") + ` log_info "Processing {{target}}"

  ` + terminal.Green("# Query database - get vulnerability counts by severity") + `
  osmedeus func e 'db_select("SELECT severity, COUNT(*) as count FROM vulnerabilities GROUP BY severity", "markdown")'

  ` + terminal.Green("# Query database - get filtered assets as JSONL") + `
  osmedeus func e 'db_select_assets_filtered("example.com", 200, "subdomain", "jsonl")'

  ` + terminal.Green("# Query database - get all vulnerabilities for a workspace") + `
  osmedeus func e 'db_select_vulnerabilities("example.com", "markdown")'

  ` + terminal.Green("# Read script from stdin") + `
  echo 'print_markdown_from_file("README.md")' | osmedeus func e --stdin

  ` + terminal.Green("# Alternative stdin syntax") + `
  echo 'log_info("hello")' | osmedeus func e -

` + docsFooter()
}

// UsageHealth returns the Long description for the health command
func UsageHealth() string {
	return terminal.BoldCyan("◆ Description") + `
  Check the Osmedeus environment for issues and fix them.
  ` + terminal.Gray("This command is an alias for 'osmedeus install validate'.") + `

` + terminal.BoldCyan("✔ Checks Performed") + `
  • Base folder, workspaces, workflows folders exist (creates if missing)
  • Configuration file is valid (osm-settings.yaml)
  • All workflows are valid

` + terminal.BoldCyan("▷ Examples") + `
  osmedeus health                 # using alias
  osmedeus install validate       # primary command

` + docsFooter()
}

// UsageWorker returns the Long description for the worker command
func UsageWorker() string {
	return terminal.BoldCyan("◆ Description") + `
  Commands for managing worker nodes in distributed mode.

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("join") + `    - Join the distributed worker pool
  • ` + terminal.Yellow("status") + `  - Show worker pool status

` + docsFooter()
}

// UsageWorkerJoin returns the Long description for the worker join command
func UsageWorkerJoin() string {
	return terminal.BoldCyan("◆ Description") + `
  Join the distributed worker pool and start processing tasks.

  The worker will connect to Redis and wait for tasks from the master node.
  Tasks are executed using the local workflow engine.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Join using settings from osm-settings.yaml") + `
  osmedeus worker join

  ` + terminal.Green("# Join using a specific Redis URL") + `
  osmedeus worker join ` + terminal.Yellow("--redis-url") + ` redis://user:pass@localhost:6379/0

` + docsFooter()
}

// UsageWorkerStatus returns the Long description for the worker status command
func UsageWorkerStatus() string {
	return terminal.BoldCyan("◆ Description") + `
  Display the status of all workers connected to the Redis server.

` + docsFooter()
}

// UsageConfig returns the Long description for the config command
func UsageConfig() string {
	return terminal.BoldCyan("◆ Description") + `
  Manage osmedeus configuration settings.

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("clean") + `  - Reset configuration to defaults
  • ` + terminal.Yellow("set") + `    - Set a configuration value
  • ` + terminal.Yellow("view") + `   - View a configuration value
  • ` + terminal.Yellow("list") + `   - List configuration values

` + docsFooter()
}

// UsageConfigClean returns the Long description for the config clean command
func UsageConfigClean() string {
	return terminal.BoldCyan("◆ Description") + `
  Reset the configuration file to default values.
  Backs up the existing config to osm-settings.yaml.backup before overwriting.

` + terminal.BoldCyan("▷ Example") + `
  ` + terminal.Green("osmedeus config clean") + `

` + docsFooter()
}

// UsageConfigSet returns the Long description for the config set command
func UsageConfigSet() string {
	return terminal.BoldCyan("◆ Description") + `
  Set a configuration value using dot notation.

` + terminal.BoldCyan("▷ Syntax") + `
  osmedeus config set <key> <value>

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("osmedeus config set server.port 9000") + `
  ` + terminal.Green("osmedeus config set server.username admin") + `
  ` + terminal.Green("osmedeus config set server.password \"d8506b99a052e797f73d1dab\"") + `
  ` + terminal.Green("osmedeus config set server.jwt.secret_signing_key \"d8506b99a052e797f73d1dab\"") + `
  ` + terminal.Green("osmedeus config set scan_tactic.default 20") + `
  ` + terminal.Green("osmedeus config set global_vars.github_token ghp_xxx") + `
  ` + terminal.Green("osmedeus config set notification.enabled true") + `

` + terminal.BoldCyan("▷ Available Keys") + `
  ` + terminal.Yellow("base_folder") + `                    Base directory path
  ` + terminal.Yellow("server.host") + `                    Server bind host
  ` + terminal.Yellow("server.port") + `                    Server port number
  ` + terminal.Yellow("server.username") + `                Auth username
  ` + terminal.Yellow("server.password") + `                Auth password
  ` + terminal.Yellow("server.simple_user_map_key.<username>") + ` Auth user password by username
  ` + terminal.Yellow("server.jwt.secret_signing_key") + `    JWT secret signing key
  ` + terminal.Yellow("server.jwt.expiration_minutes") + `    JWT expiration time in minutes
  ` + terminal.Yellow("server.ui_path") + `                 UI static files path
  ` + terminal.Yellow("server.enabled_auth_api") + `        Enable API key auth (true/false)
  ` + terminal.Yellow("server.auth_api_key") + `            API key for x-osm-api-key header
  ` + terminal.Yellow("database.db_engine") + `             sqlite or postgresql
  ` + terminal.Yellow("database.host") + `                  Database host
  ` + terminal.Yellow("database.port") + `                  Database port
  ` + terminal.Yellow("scan_tactic.aggressive") + `         Aggressive mode threads
  ` + terminal.Yellow("scan_tactic.default") + `            Default mode threads
  ` + terminal.Yellow("scan_tactic.gently") + `             Gentle mode threads
  ` + terminal.Yellow("redis.host") + `                     Redis host
  ` + terminal.Yellow("redis.port") + `                     Redis port
  ` + terminal.Yellow("global_vars.<name>") + `             Set a global variable
  ` + terminal.Yellow("notification.enabled") + `           Enable notifications (true/false)
  ` + terminal.Yellow("notification.telegram.bot_token") + ` Telegram bot token
  ` + terminal.Yellow("environments.external_binaries_path") + ` Binaries directory
  ` + terminal.Yellow("storage.enabled") + `                Enable cloud storage (true/false)

` + docsFooter()
}

func UsageConfigView() string {
	return terminal.BoldCyan("◆ Description") + `
  View a configuration value using dot notation.

` + terminal.BoldCyan("▷ Syntax") + `
  osmedeus config view <key>

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("osmedeus config view server.port") + `
  ` + terminal.Green("osmedeus config view server.username") + `
  ` + terminal.Green("osmedeus config view server.password") + `
  ` + terminal.Green("osmedeus config view server.jwt.secret_signing_key") + `
  ` + terminal.Green("osmedeus config view server.jwt.secret_signing_key --redact") + `

` + docsFooter()
}

func UsageConfigList() string {
	return terminal.BoldCyan("◆ Description") + `
  List configuration values in dot notation.

` + terminal.BoldCyan("▷ Syntax") + `
  osmedeus config list

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("osmedeus config list") + `
  ` + terminal.Green("osmedeus config list --show-secrets") + `

` + docsFooter()
}

// UsageDB returns the Long description for the db command
func UsageDB() string {
	return terminal.BoldCyan("◆ Description") + `
  Database management commands for seeding and cleaning data.

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("list") + `     - List database tables and row counts
  • ` + terminal.Yellow("seed") + `     - Seed database with sample data
  • ` + terminal.Yellow("clean") + `    - Remove all data from database
  • ` + terminal.Yellow("migrate") + `  - Run database migrations

` + docsFooter()
}

// UsageDBSeed returns the Long description for the db seed command
func UsageDBSeed() string {
	return terminal.BoldCyan("◆ Description") + `
  Seed the database with sample data for development and testing.

  This command populates the database with realistic sample records including:
  • Runs (completed, running, failed examples)
  • Step results (subfinder, httpx, nuclei, etc.)
  • Artifacts (subdomains.txt, alive-hosts.txt, etc.)
  • Assets (HTTP endpoints with status codes and tech stacks)
  • Event logs (run events, asset discoveries)
  • Schedules (daily recon, weekly vuln scan)

` + terminal.BoldCyan("▷ Example") + `
  ` + terminal.Green("osmedeus db seed") + `

` + docsFooter()
}

// UsageDBClean returns the Long description for the db clean command
func UsageDBClean() string {
	return terminal.BoldCyan("◆ Description") + `
  Remove all data from all database tables.

  ` + terminal.Yellow("WARNING:") + ` This is a destructive operation that cannot be undone.
  Use the --force flag to skip the confirmation prompt.

` + terminal.BoldCyan("▷ Example") + `
  ` + terminal.Green("osmedeus db clean --force") + `

` + docsFooter()
}

// UsageDBMigrate returns the Long description for the db migrate command
func UsageDBMigrate() string {
	return terminal.BoldCyan("◆ Description") + `
  Run database migrations to create or update tables.

  This command ensures all required tables exist with the correct schema.
  Safe to run multiple times (uses IF NOT EXISTS).

` + terminal.BoldCyan("▷ Example") + `
  ` + terminal.Green("osmedeus db migrate") + `

` + docsFooter()
}

// UsageDBList returns the Long description for the db list command
func UsageDBList() string {
	return terminal.BoldCyan("◆ Description") + `
  List all database tables with their row counts, or list records from a
  specific table with pagination support.

` + terminal.BoldCyan("▶ Options") + `
  ` + terminal.Yellow("-t, --table") + `         Table name to list records from
  ` + terminal.Yellow("--offset") + `            Number of records to skip (default: 0)
  ` + terminal.Yellow("--limit") + `             Maximum records to return (default: 20, max: 100)
  ` + terminal.Yellow("--list-columns") + `      List all available columns for the specified table
  ` + terminal.Yellow("--exclude-columns") + `   Comma-separated column names to exclude from output

` + terminal.BoldCyan("▶ Valid Tables") + `
  runs, step_results, artifacts, assets, event_logs, schedules

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# List all tables with row counts") + `
  osmedeus db list

  ` + terminal.Green("# List records from runs table") + `
  osmedeus db list ` + terminal.Yellow("-t") + ` runs

  ` + terminal.Green("# List available columns for assets table") + `
  osmedeus db list ` + terminal.Yellow("-t") + ` assets ` + terminal.Yellow("--list-columns") + `

  ` + terminal.Green("# List assets excluding specific columns") + `
  osmedeus db list ` + terminal.Yellow("-t") + ` assets ` + terminal.Yellow("--exclude-columns") + ` id,created_at,updated_at

  ` + terminal.Green("# List assets with pagination") + `
  osmedeus db list ` + terminal.Yellow("-t") + ` assets ` + terminal.Yellow("--offset") + ` 0 ` + terminal.Yellow("--limit") + ` 10

  ` + terminal.Green("# Get next page of results") + `
  osmedeus db list ` + terminal.Yellow("-t") + ` assets ` + terminal.Yellow("--offset") + ` 10 ` + terminal.Yellow("--limit") + ` 10

` + docsFooter()
}

// UsageInstall returns the Long description for the install command
func UsageInstall() string {
	return terminal.BoldCyan("◆ Description") + `
  Install workflows, base folder, or binaries from various sources.

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("workflow") + `  - Install workflows from git URL, zip URL, or local zip
  • ` + terminal.Yellow("base") + `      - Install base folder (backs up and restores database)
  • ` + terminal.Yellow("binary") + `    - Install binaries from registry
  • ` + terminal.Yellow("env") + `       - Add binaries path to shell configuration
  • ` + terminal.Yellow("validate") + `  - Check and fix environment health

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# List available binaries (direct-fetch mode)") + `
  osmedeus install binary ` + terminal.Yellow("--list-registry-direct-fetch") + `

  ` + terminal.Green("# List available binaries (nix-build mode)") + `
  osmedeus install binary ` + terminal.Yellow("--list-registry-nix-build") + `

  ` + terminal.Green("# Install specific binaries") + `
  osmedeus install binary ` + terminal.Yellow("--name") + ` nuclei ` + terminal.Yellow("--name") + ` httpx

  ` + terminal.Green("# Install all required binaries") + `
  osmedeus install binary ` + terminal.Yellow("--all") + `

  ` + terminal.Green("# Install all binaries including optional ones") + `
  osmedeus install binary ` + terminal.Yellow("--all") + ` ` + terminal.Yellow("--install-optional") + `

  ` + terminal.Green("# Check if binaries are installed") + `
  osmedeus install binary ` + terminal.Yellow("--all") + ` ` + terminal.Yellow("--check") + `

  ` + terminal.Green("# Install Nix package manager") + `
  osmedeus install binary ` + terminal.Yellow("--nix-installation") + `

  ` + terminal.Green("# Install binary via Nix") + `
  osmedeus install binary ` + terminal.Yellow("--name") + ` nuclei ` + terminal.Yellow("--nix-build-install") + `

  ` + terminal.Green("# Install all binaries via Nix") + `
  osmedeus install binary ` + terminal.Yellow("--all") + ` ` + terminal.Yellow("--nix-build-install") + `

  ` + terminal.Green("# Install workflows from git or from a zip URL or from a local zip file") + `
  osmedeus install workflow https://github.com/user/osmedeus-workflows.git
  osmedeus install workflow http://<custom-host>/workflow-osmedeus.zip
  osmedeus install workflow local-file-workflow-osmedeus.zip

  ` + terminal.Green("# Install base folder from git") + `
  osmedeus install base https://github.com/user/osmedeus-base.git
  osmedeus install base http://<custom-host>/osmedeus-base.zip
  osmedeus install base local-file-osmedeus-base.zip

` + docsFooter()
}

// UsageAllExamples returns comprehensive usage examples for all commands
func UsageAllExamples() string {
	return terminal.BoldCyan("▶ Run Examples") + `
  ` + terminal.Green("# Basic module run") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Flow workflow run") + `
  osmedeus run ` + terminal.Yellow("-f") + ` general ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Multiple targets") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` target1.com ` + terminal.Yellow("-t") + ` target2.com

  ` + terminal.Green("# Stdin input") + `
  cat urls.txt | osmedeus run ` + terminal.Yellow("-m") + ` recon

  ` + terminal.Green("# Run with stdin input with concurrency") + `
  cat list-of-urls.txt | osmedeus run ` + terminal.Yellow("-m") + ` simple-module ` + terminal.Yellow("--concurrency") + ` 10

  ` + terminal.Green("# With custom parameters") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--params") + ` 'threads=50'

  ` + terminal.Green("# Parameters from YAML file") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--params-file") + ` params.yaml

  ` + terminal.Green("# Dry-run mode") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--dry-run") + `

  ` + terminal.Green("# Run module from stdin YAML") + `
  cat module.yaml | osmedeus run ` + terminal.Yellow("--std-module") + ` ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Custom workspace") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--workspace") + ` /path/to/workspace

  ` + terminal.Green("# With timeout") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--timeout") + ` 2h

  ` + terminal.Green("# Repeat run continuously") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--repeat") + ` ` + terminal.Yellow("--repeat-wait-time") + ` 1h

  ` + terminal.Green("# Run multiple modules in sequence") + `
  osmedeus run ` + terminal.Yellow("-m") + ` subdomain ` + terminal.Yellow("-m") + ` portscan ` + terminal.Yellow("-m") + ` vuln ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Skip heuristics checks") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--heuristics-check") + ` none

  ` + terminal.Green("# Concurrent targets") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--concurrency") + ` 5

` + terminal.BoldYellow("★ Function Eval (Powerful Scripting)") + `
  ` + terminal.Green("# Print markdown file") + `
  osmedeus func e 'print_markdown_from_file("README.md")'

  ` + terminal.Green("# Log with variable substitution") + `
  osmedeus func e 'log_info("Scanning {{target}}")' ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Save content to file") + `
  osmedeus func e 'save_content("data", "/tmp/out.txt")'

  ` + terminal.Green("# Make HTTP request") + `
  osmedeus func e 'httpRequest("https://api.example.com", "GET", {}, "")'

  ` + terminal.Green("# Sort file using Unix sort") + `
  osmedeus func e 'sortUnix("/tmp/input.txt", "/tmp/sorted.txt")'

  ` + terminal.Green("# Read from stdin") + `
  echo 'log_info("hello")' | osmedeus func e -

` + terminal.BoldCyan("▶ Server Examples") + `
  ` + terminal.Green("# Start server") + `
  osmedeus serve

  ` + terminal.Green("# Custom port") + `
  osmedeus serve ` + terminal.Yellow("--port") + ` 8080

  ` + terminal.Green("# No authentication (dev mode)") + `
  osmedeus serve ` + terminal.Yellow("-A") + `

  ` + terminal.Green("# Distributed master mode") + `
  osmedeus serve ` + terminal.Yellow("--master") + `

` + terminal.BoldCyan("▶ Workflow Examples") + `
  ` + terminal.Green("# List all workflows") + `
  osmedeus workflow list

  ` + terminal.Green("# Show workflow details") + `
  osmedeus workflow show recon

  ` + terminal.Green("# Validate a workflow") + `
  osmedeus workflow validate my-workflow

` + terminal.BoldCyan("▶ Worker Examples (Distributed Mode)") + `
  ` + terminal.Green("# Join worker pool") + `
  osmedeus worker join

  ` + terminal.Green("# With custom Redis URL") + `
  osmedeus worker join ` + terminal.Yellow("--redis-url") + ` redis://localhost:6379/0

  ` + terminal.Green("# Check worker status") + `
  osmedeus worker status

` + terminal.BoldCyan("▶ Install Examples") + `
  ` + terminal.Green("# Install binary") + `
  osmedeus install binary ` + terminal.Yellow("--name") + ` nuclei

  ` + terminal.Green("# Install multiple binaries") + `
  osmedeus install binary ` + terminal.Yellow("--name") + ` nuclei ` + terminal.Yellow("--name") + ` httpx

  ` + terminal.Green("# Install all binaries") + `
  osmedeus install binary ` + terminal.Yellow("--all") + `

  ` + terminal.Green("# Install Nix package manager") + `
  osmedeus install binary ` + terminal.Yellow("--nix-installation") + `

  ` + terminal.Green("# Install binary via Nix") + `
  osmedeus install binary ` + terminal.Yellow("--name") + ` nuclei ` + terminal.Yellow("--nix-build-install") + `

  ` + terminal.Green("# Install all binaries via Nix") + `
  osmedeus install binary ` + terminal.Yellow("--all") + ` ` + terminal.Yellow("--nix-build-install") + `

  ` + terminal.Green("# Install workflows from git") + `
  osmedeus install workflow https://github.com/user/workflows.git

` + terminal.BoldCyan("▶ Utility Examples") + `
  ` + terminal.Green("# Health check") + `
  osmedeus health

  ` + terminal.Green("# Reset config") + `
  osmedeus config clean

  ` + terminal.Green("# Set config value") + `
  osmedeus config set server.port 9000

  ` + terminal.Green("# Database commands") + `
  osmedeus db list
  osmedeus db seed
  osmedeus db clean ` + terminal.Yellow("--force") + `

` + docsFooter()
}

// UsageFullExample returns comprehensive usage with all flags for pager display
func UsageFullExample() string {
	return terminal.BoldCyan("═══════════════════════════════════════════════════════════════════") + `
` + terminal.BoldCyan("                     OSMEDEUS FULL USAGE REFERENCE") + `
` + terminal.BoldCyan("═══════════════════════════════════════════════════════════════════") + `

` + terminal.BoldYellow("GLOBAL FLAGS") + ` (available for all commands)
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  ` + terminal.Yellow("--settings-file") + `        Path to settings file (default: $HOME/osmedeus-base/osm-settings.yaml)
  ` + terminal.Yellow("-b, --base-folder") + `      Base folder containing workflows and settings
  ` + terminal.Yellow("-F, --workflow-folder") + `  Custom workflow folder path
  ` + terminal.Yellow("-v, --verbose") + `          Enable verbose output
  ` + terminal.Yellow("--debug") + `                Enable debug mode (verbose + debug logging)
  ` + terminal.Yellow("-q, --silent") + `           Silent mode - suppress all output except errors
  ` + terminal.Yellow("--log-file") + `             Path to log file (logs to both console and file)
  ` + terminal.Yellow("--log-file-tmp") + `         Create temporary log file osmedeus-log-<timestamp>.log
  ` + terminal.Yellow("-H, --usage-example") + `    Show comprehensive usage examples
  ` + terminal.Yellow("--full-usage-example") + `   Show this full usage reference (pager mode)
  ` + terminal.Yellow("--spinner") + `              Show spinner animations during execution
  ` + terminal.Yellow("--disable-logging") + `      Disable all logging output
  ` + terminal.Yellow("--disable-color") + `        Disable colored output
  ` + terminal.Yellow("--disable-notification") + ` Disable all notifications
  ` + terminal.Yellow("--disable-db") + `           Disable database connection (lightweight mode)
  ` + terminal.Yellow("--ci-output-format") + `     Output results in JSON format for CI pipelines

` + terminal.BoldYellow("RUN COMMAND") + ` - Execute workflows
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus run [flags]

` + terminal.Cyan("  Workflow Selection:") + `
  ` + terminal.Yellow("-f, --flow") + `             Flow workflow name to execute
  ` + terminal.Yellow("-m, --module") + `           Module workflow(s) to execute (can specify multiple)
  ` + terminal.Yellow("--std-module") + `           Read module YAML from stdin

` + terminal.Cyan("  Target Selection:") + `
  ` + terminal.Yellow("-t, --target") + `           Target(s) to run against (can specify multiple)
  ` + terminal.Yellow("-T, --target-file") + `      File containing targets (one per line)
  ` + terminal.Yellow("--empty-target") + `         Run without target (generates placeholder)

` + terminal.Cyan("  Parameters:") + `
  ` + terminal.Yellow("-p, --params") + `           Additional parameters (key=value format)
  ` + terminal.Yellow("-P, --params-file") + `      File containing parameters (JSON or YAML)
  ` + terminal.Yellow("-B, --tactic") + `           Run tactic: aggressive, default, gently
  ` + terminal.Yellow("--threads-hold") + `         Override thread count (0 = use tactic default)

` + terminal.Cyan("  Execution Control:") + `
  ` + terminal.Yellow("-c, --concurrency") + `      Number of targets to run concurrently (default: 1)
  ` + terminal.Yellow("--timeout") + `              Run timeout (e.g., 2h, 3h, 1d)
  ` + terminal.Yellow("--repeat") + `               Repeat run after completion
  ` + terminal.Yellow("--repeat-wait-time") + `     Wait time between repeats (default: 1h)
  ` + terminal.Yellow("--dry-run") + `              Show what would be executed without running
  ` + terminal.Yellow("-G, --progress-bar") + `     Show progress bar during execution

` + terminal.Cyan("  Workspace:") + `
  ` + terminal.Yellow("-w, --workspace") + `        Custom workspace path
  ` + terminal.Yellow("-W, --workspaces-folder") + ` Override {{Workspaces}} variable
  ` + terminal.Yellow("-S, --space") + `            Override {{TargetSpace}} variable

` + terminal.Cyan("  Filtering:") + `
  ` + terminal.Yellow("-x, --exclude") + `          Module(s) to exclude from execution
  ` + terminal.Yellow("--heuristics-check") + `     Heuristics check level: none, basic, advanced

` + terminal.Cyan("  Distributed Mode:") + `
  ` + terminal.Yellow("-D, --distributed-run") + `  Submit run to distributed worker queue
  ` + terminal.Yellow("--redis-url") + `            Redis connection URL for distributed mode

` + terminal.BoldYellow("SERVE COMMAND") + ` - Start REST API server
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus serve [flags]

  ` + terminal.Yellow("--host") + `                 Host to bind the server to (default: from config)
  ` + terminal.Yellow("--port") + `                 Port number for the API server
  ` + terminal.Yellow("-A, --no-auth") + `          Disable authentication (development only)
  ` + terminal.Yellow("--master") + `               Run as distributed master node
  ` + terminal.Yellow("--redis-url") + `            Redis connection URL for master mode

` + terminal.BoldYellow("WORKFLOW COMMAND") + ` - Manage workflows
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus workflow list              List available workflows
  osmedeus workflow show <name>       Show workflow details
  osmedeus workflow validate <name>   Validate a workflow (alias: val)

` + terminal.Cyan("  List Flags:") + `
  ` + terminal.Yellow("--tags") + `                 Filter workflows by tags (comma-separated)
  ` + terminal.Yellow("--show-tags") + `            Show tags column in output

` + terminal.Cyan("  Show Flags:") + `
  ` + terminal.Yellow("-v, --verbose") + `          Show detailed variable descriptions
  ` + terminal.Yellow("--table") + `                Show metadata table instead of YAML

` + terminal.BoldYellow("FUNCTION COMMAND") + ` - Execute utility functions
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus func list                  List all available functions (alias: ls)
  osmedeus func eval <script>         Evaluate a script (alias: e)

` + terminal.Cyan("  Eval Flags:") + `
  ` + terminal.Yellow("-e, --eval") + `             Script to evaluate
  ` + terminal.Yellow("-t, --target") + `           Target value for {{target}} variable
  ` + terminal.Yellow("--params") + `               Additional parameters (key=value format)
  ` + terminal.Yellow("--stdin") + `                Read script from stdin

` + terminal.BoldYellow("WORKER COMMAND") + ` - Distributed worker management
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus worker join                Join the distributed worker pool
  osmedeus worker status              Show worker pool status

` + terminal.Cyan("  Join Flags:") + `
  ` + terminal.Yellow("--redis-url") + `            Redis connection URL
  ` + terminal.Yellow("--workers") + `              Number of concurrent workers (default: 5)

` + terminal.Cyan("  Status Flags:") + `
  ` + terminal.Yellow("--redis-url") + `            Redis connection URL

` + terminal.BoldYellow("DATABASE COMMAND") + ` - Database management
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus db list                    List tables with row counts (alias: ls)
  osmedeus db list -t <table>         List records from a table
  osmedeus db seed                    Seed database with sample data
  osmedeus db clean --force           Remove all data from database
  osmedeus db migrate                 Run database migrations
  osmedeus db index workflow          Index workflows from filesystem to database

` + terminal.Cyan("  List Flags:") + `
  ` + terminal.Yellow("-t, --table") + `            Table name to list records from
  ` + terminal.Yellow("--offset") + `               Number of records to skip (default: 0)
  ` + terminal.Yellow("--limit") + `                Maximum records to return (default: 50)
  ` + terminal.Yellow("--json") + `                 Output records as JSON only (bypasses TUI)
  ` + terminal.Yellow("--no-tui") + `               Disable interactive TUI mode, use plain text
  ` + terminal.Yellow("--where") + `                Filter records (key=value, can be repeated)
  ` + terminal.Yellow("--columns") + `              Comma-separated columns to display
  ` + terminal.Yellow("--search") + `               Search all columns for substring
  ` + terminal.Yellow("--width") + `                Max column width for table display (default: 30)
  ` + terminal.Yellow("--all") + `                  Show all columns including hidden ones

` + terminal.Cyan("  Clean Flags:") + `
  ` + terminal.Yellow("--force") + `                Skip confirmation prompt

` + terminal.Cyan("  Index Workflow Flags:") + `
  ` + terminal.Yellow("--force") + `                Force re-index all workflows regardless of checksum

` + terminal.BoldYellow("CONFIG COMMAND") + ` - Configuration management
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus config clean               Reset configuration to defaults
  osmedeus config set <key> <value>   Set a configuration value
  osmedeus config view <key>          View a configuration value
  osmedeus config list                List configuration values

` + terminal.BoldYellow("INSTALL COMMAND") + ` - Install components
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus install workflow <source>  Install workflows from git/zip
  osmedeus install base <source>      Install base folder
  osmedeus install binary             Install binaries from registry
  osmedeus install validate           Check and fix environment health (alias: val)
  osmedeus install env                Display environment paths

` + terminal.Cyan("  Binary Flags:") + `
  ` + terminal.Yellow("-n, --name") + `             Binary name(s) to install (can be repeated)
  ` + terminal.Yellow("--all") + `                  Install all binaries from registry
  ` + terminal.Yellow("--check") + `                Check if binaries are installed
  ` + terminal.Yellow("-r, --registry") + `         Custom registry JSON file path or URL
  ` + terminal.Yellow("--nix-pkgs") + `             Nix package(s) to add (repeatable)
  ` + terminal.Yellow("--nix-build-install") + `    Use Nix to install binaries instead of direct downloads
  ` + terminal.Yellow("--nix-installation") + `     Install Nix package manager (Determinate Systems installer)

` + terminal.BoldYellow("HEALTH COMMAND") + ` - Environment health check
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus health                     Check environment for issues

` + terminal.BoldCyan("═══════════════════════════════════════════════════════════════════") + `
` + docsFooter()
}

// docsFooter returns the documentation footer
func docsFooter() string {
	return terminal.HiCyan("📖 Documentation: ") + terminal.HiWhite(core.DOCS) + "\n"
}
