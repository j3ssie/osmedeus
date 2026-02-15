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

  ` + terminal.Green("# Run module from URL") + `
  osmedeus run ` + terminal.Yellow("--module-url") + ` https://example.com/module.yaml ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Run module from GitHub (public)") + `
  osmedeus run ` + terminal.Yellow("--module-url") + ` https://raw.githubusercontent.com/user/repo/main/module.yaml ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Run module from private GitHub repo (requires GH_TOKEN or GITHUB_API_KEY)") + `
  osmedeus run ` + terminal.Yellow("--module-url") + ` https://github.com/user/private-repo/blob/main/module.yaml ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Load parameters from YAML/JSON file") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--params-file") + ` params.yaml

  ` + terminal.Green("# Custom workspace path") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--workspace") + ` /custom/workspace

  ` + terminal.Green("# Skip heuristics checks") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--heuristics-check") + ` none

  ` + terminal.Green("# Concurrent targets from file") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--concurrency") + ` 5

  ` + terminal.Green("# View chunk info for target file") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--chunk-size") + ` 100

  ` + terminal.Green("# Run specific chunk (0-indexed)") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--chunk-size") + ` 100 ` + terminal.Yellow("--chunk-part") + ` 2

  ` + terminal.Green("# Split into 4 equal chunks and run chunk 0") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--chunk-count") + ` 4 ` + terminal.Yellow("--chunk-part") + ` 0

  ` + terminal.Green("# Distributed processing across machines") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--chunk-size") + ` 250 ` + terminal.Yellow("--chunk-part") + ` 0  ` + terminal.Gray("# Machine 1") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--chunk-size") + ` 250 ` + terminal.Yellow("--chunk-part") + ` 1  ` + terminal.Gray("# Machine 2") + `

  ` + terminal.Green("# Queue a run for later processing") + `
  osmedeus run ` + terminal.Yellow("--queue") + ` ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Queue with file target") + `
  osmedeus run ` + terminal.Yellow("--queue") + ` ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt

  ` + terminal.Green("# Process queued tasks (alias for 'osmedeus worker queue run')") + `
  osmedeus run ` + terminal.Yellow("--queue-run") + `

  ` + terminal.Green("# Process queued tasks with concurrency") + `
  osmedeus run ` + terminal.Yellow("--queue-run") + ` ` + terminal.Yellow("--concurrency") + ` 3

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

  ` + terminal.Green("# Start server without queue polling") + `
  osmedeus serve ` + terminal.Yellow("--no-queue-polling") + `

` + docsFooter()
}

// UsageWorkflow returns the Long description for the workflow command
func UsageWorkflow() string {
	return terminal.BoldCyan("◆ Description") + `
  Commands for listing, viewing, and validating workflows.

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("list") + `      - List available workflows (alias: ls)
  • ` + terminal.Yellow("show") + `      - Show workflow details (alias: view)
  • ` + terminal.Yellow("validate") + `  - Validate a workflow (alias: val)

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

` + terminal.BoldCyan("▷ Bulk Processing") + `
  ` + terminal.Green("# Process multiple targets from file") + `
  osmedeus func e 'log_info("Processing: " + target)' ` + terminal.Yellow("-T") + ` targets.txt

  ` + terminal.Green("# Function from file with targets") + `
  osmedeus func e ` + terminal.Yellow("--function-file") + ` check.js ` + terminal.Yellow("-T") + ` targets.txt

  ` + terminal.Green("# With concurrency") + `
  osmedeus func e 'httpGet("https://" + target)' ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("-c") + ` 10

  ` + terminal.Green("# Combined with params") + `
  osmedeus func e 'log_info(prefix + target)' ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--params") + ` 'prefix=test_' ` + terminal.Yellow("-c") + ` 5

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
  • ` + terminal.Yellow("status") + `  - Show worker pool status (alias: ls); use ` + terminal.Yellow("--json") + ` for JSON output
  • ` + terminal.Yellow("set") + `     - Update a worker field (alias, public-ip, ssh-enabled, ssh-keys-path)
  • ` + terminal.Yellow("eval") + `    - Evaluate a function expression with distributed hooks
  • ` + terminal.Yellow("queue") + `   - Manage and process queued tasks (list, new, run)

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

  ` + terminal.Green("# Join and auto-detect public IP") + `
  osmedeus worker join ` + terminal.Yellow("--get-public-ip") + `

` + docsFooter()
}

// UsageWorkerStatus returns the Long description for the worker status command
func UsageWorkerStatus() string {
	return terminal.BoldCyan("◆ Description") + `
  Display the status of all workers connected to the Redis server.

  Use ` + terminal.Yellow("--json") + ` to output worker info as JSON for scripting and automation.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Show worker status as a table") + `
  osmedeus worker status

  ` + terminal.Green("# Output worker info as JSON") + `
  osmedeus worker status ` + terminal.Yellow("--json") + `

` + docsFooter()
}

// UsageWorkerEval returns the Long description for the worker eval command
func UsageWorkerEval() string {
	return terminal.BoldCyan("◆ Description") + `
  Evaluate a utility function expression with distributed hooks registered.

  This connects to Redis and registers run_on_master() hooks so that
  expressions can route calls to the master node. Useful for one-shot
  operations from a worker context (e.g., inside Docker or CI pipelines)
  without running a full worker loop.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Simple function eval with distributed hooks") + `
  osmedeus worker eval 'log_info("hello from worker eval")' ` + terminal.Yellow("--redis-url") + ` redis://localhost:6379

  ` + terminal.Green("# Route a call to the master node") + `
  osmedeus worker eval 'run_on_master("func", "log_info(\"routed via redis\")")' ` + terminal.Yellow("--redis-url") + ` redis://localhost:6379

  ` + terminal.Green("# With target variable") + `
  osmedeus worker eval 'log_info("hello")' ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("--redis-url") + ` redis://localhost:6379

  ` + terminal.Green("# Read script from stdin") + `
  echo 'run_on_master("func", "db_import_sarif(\"ws\", \"/path/f.sarif\")")' | osmedeus worker eval ` + terminal.Yellow("--stdin") + ` ` + terminal.Yellow("--redis-url") + ` redis://localhost:6379

` + docsFooter()
}

// UsageWorkerSet returns the Long description for the worker set command
func UsageWorkerSet() string {
	return terminal.BoldCyan("◆ Description") + `
  Update a field on a registered worker. The worker can be identified by its
  ID or alias.

` + terminal.BoldCyan("▶ Valid Fields") + `
  • ` + terminal.Yellow("alias") + `          - Human-friendly name for the worker
  • ` + terminal.Yellow("public-ip") + `      - Public IP address
  • ` + terminal.Yellow("ssh-enabled") + `    - Whether SSH is enabled (true/false)
  • ` + terminal.Yellow("ssh-keys-path") + `  - Path to SSH keys

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Set an alias for a worker") + `
  osmedeus worker set <worker-id> alias scanner-1

  ` + terminal.Green("# Set public IP") + `
  osmedeus worker set scanner-1 public-ip 203.0.113.10

  ` + terminal.Green("# Enable SSH") + `
  osmedeus worker set scanner-1 ssh-enabled true

  ` + terminal.Green("# With custom Redis URL") + `
  osmedeus worker set <worker-id> alias prod-1 ` + terminal.Yellow("--redis-url") + ` redis://localhost:6379

` + docsFooter()
}

// UsageWorkerQueue returns the Long description for the worker queue command
func UsageWorkerQueue() string {
	return terminal.BoldCyan("◆ Description") + `
  Manage and process queued tasks. Tasks can be queued via the ` + terminal.Yellow("--queue") + ` flag
  on ` + terminal.Yellow("osmedeus run") + ` or via ` + terminal.Yellow("osmedeus worker queue new") + `.

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("list") + `  - List all queued tasks (alias: ls)
  • ` + terminal.Yellow("new") + `   - Queue a new task for later processing
  • ` + terminal.Yellow("run") + `   - Process queued tasks (polls DB and Redis)

` + docsFooter()
}

// UsageWorkerQueueList returns the Long description for the worker queue list command
func UsageWorkerQueueList() string {
	return terminal.BoldCyan("◆ Description") + `
  List all queued tasks from the database.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# List all queued tasks") + `
  osmedeus worker queue list

  ` + terminal.Green("# Output as JSON") + `
  osmedeus worker queue list ` + terminal.Yellow("--json") + `

` + docsFooter()
}

// UsageWorkerQueueNew returns the Long description for the worker queue new command
func UsageWorkerQueueNew() string {
	return terminal.BoldCyan("◆ Description") + `
  Queue a new task for later processing. Creates a DB record and optionally
  pushes to Redis if configured.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Queue a module run") + `
  osmedeus worker queue new ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Queue a flow run with targets from file") + `
  osmedeus worker queue new ` + terminal.Yellow("-f") + ` general ` + terminal.Yellow("-T") + ` targets.txt

  ` + terminal.Green("# Queue with parameters") + `
  osmedeus worker queue new ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-t") + ` example.com ` + terminal.Yellow("-p") + ` 'threads=20'

` + docsFooter()
}

// UsageWorkerQueueRun returns the Long description for the worker queue run command
func UsageWorkerQueueRun() string {
	return terminal.BoldCyan("◆ Description") + `
  Process queued tasks by polling both the database and Redis (if configured).
  Uses a shared channel with deduplication to prevent double-execution.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Process queued tasks with default concurrency") + `
  osmedeus worker queue run

  ` + terminal.Green("# Process with higher concurrency") + `
  osmedeus worker queue run ` + terminal.Yellow("--concurrency") + ` 3

  ` + terminal.Green("# Process with custom Redis URL") + `
  osmedeus worker queue run ` + terminal.Yellow("--redis-url") + ` redis://localhost:6379

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
  ` + terminal.Yellow("notification.provider") + `          Notification provider (telegram, webhook)
  ` + terminal.Yellow("notification.telegram.enabled") + `  Enable Telegram notifications (true/false)
  ` + terminal.Yellow("notification.telegram.bot_token") + ` Telegram bot token from @BotFather
  ` + terminal.Yellow("notification.telegram.chat_id") + `  Telegram chat ID to send messages to
  ` + terminal.Yellow("notification.webhooks.0.url") + `    Webhook URL (use index 0, 1, 2... for multiple)
  ` + terminal.Yellow("notification.webhooks.0.enabled") + ` Enable webhook (true/false)
  ` + terminal.Yellow("notification.webhooks.0.timeout") + ` Webhook timeout in seconds
  ` + terminal.Yellow("environments.external_binaries_path") + ` Binaries directory
  ` + terminal.Yellow("storage.enabled") + `                Enable cloud storage (true/false)

` + docsFooter()
}

func UsageConfigView() string {
	return terminal.BoldCyan("◆ Description") + `
  View a configuration value using dot notation.
  Supports wildcard patterns with --force flag.

` + terminal.BoldCyan("▷ Syntax") + `
  osmedeus config view <key>
  osmedeus config view '<pattern>' --force

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Exact key lookup") + `
  ` + terminal.Green("osmedeus config view server.port") + `
  ` + terminal.Green("osmedeus config view server.username") + `
  ` + terminal.Green("osmedeus config view server.password") + `
  ` + terminal.Green("osmedeus config view server.jwt.secret_signing_key") + `
  ` + terminal.Green("osmedeus config view server.jwt.secret_signing_key --redact") + `

  ` + terminal.Green("# Wildcard pattern search (requires --force)") + `
  ` + terminal.Green("osmedeus config view 'server.*' --force") + `
  ` + terminal.Green("osmedeus config view 'database.*' --force") + `
  ` + terminal.Green("osmedeus config view '*password*' --force") + `
  ` + terminal.Green("osmedeus config view 'server.*' --force --redact") + `

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
  Use --clean-ws to also remove workspace data (e.g. ~/workspaces-osmedeus).

  ` + terminal.Yellow("WARNING:") + ` This is a destructive operation that cannot be undone.
  Use the --force flag to skip the confirmation prompt.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("osmedeus db clean --force") + `
  ` + terminal.Green("osmedeus db clean --force --clean-ws") + `

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

  ` + terminal.Green("# Run module from URL") + `
  osmedeus run ` + terminal.Yellow("--module-url") + ` https://example.com/module.yaml ` + terminal.Yellow("-t") + ` example.com

  ` + terminal.Green("# Run module from private GitHub repo") + `
  osmedeus run ` + terminal.Yellow("--module-url") + ` https://github.com/user/private/blob/main/module.yaml ` + terminal.Yellow("-t") + ` example.com

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

  ` + terminal.Green("# View chunk info") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--chunk-size") + ` 100

  ` + terminal.Green("# Run specific chunk") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--chunk-size") + ` 100 ` + terminal.Yellow("--chunk-part") + ` 2

  ` + terminal.Green("# Split into 4 equal chunks and run chunk 0") + `
  osmedeus run ` + terminal.Yellow("-m") + ` recon ` + terminal.Yellow("-T") + ` targets.txt ` + terminal.Yellow("--chunk-count") + ` 4 ` + terminal.Yellow("--chunk-part") + ` 0

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

  ` + terminal.Green("# Search workflows by name or description") + `
  osmedeus workflow ls recon
  osmedeus workflow ls --search subdomain

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

  ` + terminal.Green("# Evaluate function with distributed hooks (one-shot)") + `
  osmedeus worker eval 'run_on_master("func", "log_info(\"hello\")")' ` + terminal.Yellow("--redis-url") + ` redis://localhost:6379

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
  ` + terminal.Yellow("--module-url") + `           URL to fetch module YAML from (supports GitHub private repos)

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

` + terminal.Cyan("  Chunk Mode:") + `
  ` + terminal.Yellow("--chunk-size") + `           Split targets into chunks of N targets each (0 = disabled)
  ` + terminal.Yellow("--chunk-count") + `          Split targets into N equal chunks (0 = disabled)
  ` + terminal.Yellow("--chunk-part") + `           Execute only chunk M (0-indexed, requires --chunk-size or --chunk-count)
  ` + terminal.Yellow("--chunk-threads") + `        Override concurrency within chunk (0 = use -c value)

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
  osmedeus workflow list              List available workflows (alias: ls)
  osmedeus workflow ls <search>       Search workflows by name or description
  osmedeus workflow show <name>       Show workflow details (alias: view)
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
  ` + terminal.Yellow("-T, --targets") + `          File containing targets (one per line)
  ` + terminal.Yellow("--function-file") + `        File containing the function/script to execute
  ` + terminal.Yellow("-c, --concurrency") + `      Number of concurrent executions (default: 1)

` + terminal.BoldYellow("WORKER COMMAND") + ` - Distributed worker management
` + terminal.Gray("───────────────────────────────────────────────────────────────────") + `
  osmedeus worker join                Join the distributed worker pool
  osmedeus worker status              Show worker pool status
  osmedeus worker eval <script>       Evaluate function with distributed hooks

` + terminal.Cyan("  Join Flags:") + `
  ` + terminal.Yellow("--redis-url") + `            Redis connection URL
  ` + terminal.Yellow("--workers") + `              Number of concurrent workers (default: 5)

` + terminal.Cyan("  Status Flags:") + `
  ` + terminal.Yellow("--redis-url") + `            Redis connection URL

` + terminal.Cyan("  Eval Flags:") + `
  ` + terminal.Yellow("--redis-url") + `            Redis connection URL
  ` + terminal.Yellow("-e, --eval") + `             Script to evaluate
  ` + terminal.Yellow("-t, --target") + `           Target value for {{target}} variable
  ` + terminal.Yellow("--params") + `               Additional parameters (key=value format)
  ` + terminal.Yellow("--stdin") + `                Read script from stdin

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

// UsageClient returns the Long description for the client command
func UsageClient() string {
	return terminal.BoldCyan("◆ Description") + `
  Interact with a remote osmedeus server via REST API.

` + terminal.BoldCyan("▶ Environment Variables") + `
  ` + terminal.Yellow("OSM_REMOTE_URL") + `      Remote server URL (e.g., http://localhost:8002)
  ` + terminal.Yellow("OSM_REMOTE_AUTH_KEY") + ` API authentication key for x-osm-api-key header

` + terminal.BoldCyan("▶ Subcommands") + `
  • ` + terminal.Yellow("fetch") + `  - Fetch data from server (runs, assets, vulns, etc.)
  • ` + terminal.Yellow("run") + `    - Create or cancel a run
  • ` + terminal.Yellow("exec") + `   - Execute a function remotely

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Configure via environment") + `
  export OSM_REMOTE_URL="http://localhost:8002"
  export OSM_REMOTE_AUTH_KEY="your-api-key"

  ` + terminal.Green("# Fetch data from different tables") + `
  osmedeus client fetch ` + terminal.Yellow("--table") + ` assets
  osmedeus client fetch ` + terminal.Yellow("-t") + ` runs
  osmedeus client fetch ` + terminal.Yellow("-t") + ` vulnerabilities ` + terminal.Yellow("--severity") + ` critical

  ` + terminal.Green("# Create a run") + `
  osmedeus client run ` + terminal.Yellow("-f") + ` basic-recon ` + terminal.Yellow("-T") + ` example.com

  ` + terminal.Green("# Cancel a run") + `
  osmedeus client run ` + terminal.Yellow("--cancel") + ` abc123-run-uuid

  ` + terminal.Green("# Execute a function") + `
  osmedeus client exec 'log_info("Hello from remote")'

` + docsFooter()
}

// UsageClientFetch returns the Long description for the client fetch command
func UsageClientFetch() string {
	return terminal.BoldCyan("◆ Description") + `
  Fetch data from the remote osmedeus server.

` + terminal.BoldCyan("▶ Supported Tables") + `
  • ` + terminal.Yellow("runs") + `             - Workflow execution runs
  • ` + terminal.Yellow("step_results") + `     - Step execution results
  • ` + terminal.Yellow("artifacts") + `        - Output artifacts from runs
  • ` + terminal.Yellow("assets") + `           - HTTP assets discovered during scans (default)
  • ` + terminal.Yellow("event_logs") + `       - System event logs
  • ` + terminal.Yellow("schedules") + `        - Scheduled workflow executions
  • ` + terminal.Yellow("workspaces") + `       - Scan workspaces
  • ` + terminal.Yellow("vulnerabilities") + `  - Discovered vulnerabilities
  • ` + terminal.Yellow("asset_diffs") + `      - Asset diff snapshots
  • ` + terminal.Yellow("vuln_diffs") + `       - Vulnerability diff snapshots

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Fetch assets (default)") + `
  osmedeus client fetch
  osmedeus client fetch ` + terminal.Yellow("-t") + ` assets ` + terminal.Yellow("-w") + ` example.com

  ` + terminal.Green("# Fetch runs") + `
  osmedeus client fetch ` + terminal.Yellow("--table") + ` runs
  osmedeus client fetch ` + terminal.Yellow("-t") + ` runs ` + terminal.Yellow("--status") + ` running

  ` + terminal.Green("# Fetch vulnerabilities with severity filter") + `
  osmedeus client fetch ` + terminal.Yellow("-t") + ` vulnerabilities ` + terminal.Yellow("--severity") + ` critical

  ` + terminal.Green("# Fetch step results") + `
  osmedeus client fetch ` + terminal.Yellow("-t") + ` step_results

  ` + terminal.Green("# Pagination") + `
  osmedeus client fetch ` + terminal.Yellow("-t") + ` assets ` + terminal.Yellow("--limit") + ` 50 ` + terminal.Yellow("--offset") + ` 100

  ` + terminal.Green("# JSON output") + `
  osmedeus client ` + terminal.Yellow("--json") + ` fetch ` + terminal.Yellow("-t") + ` runs

` + docsFooter()
}

// UsageClientRun returns the Long description for the client run command
func UsageClientRun() string {
	return terminal.BoldCyan("◆ Description") + `
  Create or cancel a workflow run on the remote server.

` + terminal.BoldCyan("▶ Create Mode Flags") + `
  ` + terminal.Yellow("-T, --target") + `  Target to run against (required)
  One of:
  ` + terminal.Yellow("-f, --flow") + `    Flow workflow name
  ` + terminal.Yellow("-m, --module") + `  Module workflow name

` + terminal.BoldCyan("▶ Cancel Mode") + `
  ` + terminal.Yellow("--cancel") + `      Run ID to cancel (switches to cancel mode)

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Create a flow run") + `
  osmedeus client run ` + terminal.Yellow("-f") + ` basic-recon ` + terminal.Yellow("-T") + ` example.com

  ` + terminal.Green("# Create a module run") + `
  osmedeus client run ` + terminal.Yellow("-m") + ` subdomain ` + terminal.Yellow("-T") + ` example.com

  ` + terminal.Green("# Cancel a run by ID") + `
  osmedeus client run ` + terminal.Yellow("--cancel") + ` abc123-run-uuid

  ` + terminal.Green("# JSON output") + `
  osmedeus client ` + terminal.Yellow("--json") + ` run ` + terminal.Yellow("-f") + ` recon ` + terminal.Yellow("-T") + ` example.com

` + docsFooter()
}

// UsageClientExec returns the Long description for the client exec command
func UsageClientExec() string {
	return terminal.BoldCyan("◆ Description") + `
  Execute a utility function on the remote server.

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Execute a simple function") + `
  osmedeus client exec 'log_info("Hello from remote")'

  ` + terminal.Green("# With target variable") + `
  osmedeus client exec ` + terminal.Yellow("-t") + ` example.com 'fileExists("{{target}}/output.txt")'

  ` + terminal.Green("# Using --script flag") + `
  osmedeus client exec ` + terminal.Yellow("-s") + ` 'trim("  hello  ")'

  ` + terminal.Green("# JSON output") + `
  osmedeus client ` + terminal.Yellow("--json") + ` exec 'trim("  test  ")'

` + docsFooter()
}

// UsageUninstall returns the Long description for the uninstall command
func UsageUninstall() string {
	return terminal.BoldCyan("◆ Description") + `
  Remove Osmedeus installation including base folder, configuration,
  and optionally workspace data.

  ` + terminal.BoldRed("WARNING: This is a destructive and irreversible operation!") + `

` + terminal.BoldCyan("▶ What Gets Removed") + `
  • ` + terminal.Yellow("~/osmedeus-base") + `         - Settings, workflows, binaries, data
  • ` + terminal.Yellow("~/.osmedeus") + `             - Initialization marker
  • ` + terminal.Yellow("osmedeus binary") + `         - Removed from PATH (up to 3 locations)

  With ` + terminal.Yellow("--clean") + `:
  • ` + terminal.Yellow("~/workspaces-osmedeus") + `   - All scan results and workspace data

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Preview what will be removed (no --force = confirmation prompt)") + `
  osmedeus uninstall

  ` + terminal.Green("# Uninstall without workspaces (keeps scan results)") + `
  osmedeus uninstall ` + terminal.Yellow("--force") + `

  ` + terminal.Green("# Full uninstall including all scan data") + `
  osmedeus uninstall ` + terminal.Yellow("--force") + ` ` + terminal.Yellow("--clean") + `

` + docsFooter()
}

// docsFooter returns the documentation footer
func docsFooter() string {
	return terminal.HiCyan("📖 Documentation: ") + terminal.HiWhite(core.DOCS) + "\n"
}
