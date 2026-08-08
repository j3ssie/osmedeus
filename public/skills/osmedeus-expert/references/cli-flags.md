# CLI Flags Reference

## Global Flags (All Commands)

```
--settings-file           Path to osm-settings.yaml
--base-folder, -b         Base folder with workflows
--workflow-folder, -F     Custom workflow folder
--verbose, -v             Verbose output
--debug                   Debug mode (verbose + debug logging)
--silent, -q              Suppress all output except errors
--log-file                Path to log file
--log-file-tmp            Create temporary log file
--disable-logging         Disable all logging
--disable-color           Disable colored output
--disable-notification    Disable notifications
--spinner                 Show spinner animations
--ci-output-format        JSON output for CI pipelines
--usage-example, -H       Show usage examples
--full-usage-example      Show full usage in pager
--json                    Output in JSON format
--width                   Max column width for tables (default: 80)
--force                   Skip confirmation prompts
--disable-db              Disable database connection
--skip-auto-setup         Skip automatic setup
```

## Run Command Flags

### Workflow Selection

```
-f, --flow NAME           Flow workflow name
-m, --module NAME         Module workflow(s) (repeatable: -m mod1 -m mod2)
```

### Target Input

```
-t, --target TARGET       Target(s) (repeatable)
-T, --target-file FILE    File with targets (one per line)
--empty-target            Run without target
--convert-to-file         Write targets to temp file, use file path as target
--convert-file-to-line    Expand file target, each line becomes separate target
```

### Parameters

```
-p, --params KEY=VALUE    Parameters (repeatable: -p k1=v1 -p k2=v2)
-P, --params-file FILE    Parameters file (YAML or JSON)
-S, --space NAME          Override {{TargetSpace}}
-W, --workspaces-folder   Override {{Workspaces}}
-w, --workspace PATH      Custom workspace path
```

### Execution Control

```
-c, --concurrency N       Concurrent targets (default: 1)
-B, --tactic TACTIC       Run tactic: aggressive, default, gently
--threads-hold N          Override thread count (0 = use tactic)
--timeout DURATION        Run timeout (e.g., 2h, 30m, 1d)
--repeat                  Repeat after completion
--repeat-wait-time DUR    Wait between repeats (default: 1m)
--dry-run                 Show execution plan without running
--skip-validation         Skip target type validation
--sudo-aware              Authenticate sudo once and keep credentials alive
```

### Module Selection

```
-x, --exclude MODULE      Exclude module(s) (repeatable)
-X, --fuzzy-exclude STR   Exclude modules matching substring (repeatable)
--std-module              Read module YAML from stdin
--module-url URL          Fetch module YAML from URL
```

### Heuristics

```
--heuristics-check LEVEL  none, basic (default), advanced
```

### Chunking (Large Target Lists)

```
--chunk-size N            Split targets into chunks of N
--chunk-count N           Split into N equal chunks
--chunk-part M            Execute only chunk M (0-indexed)
--chunk-threads N         Override concurrency within chunk
```

### Distributed & Queue

```
-D, --distributed-run     Submit to worker queue (requires Redis)
--redis-url URL           Redis connection URL override
--queue                   Queue for later processing
--queue-run               Process queued tasks
-G, --progress-bar        Show progress bar
--disable-workflow-state  Don't save workflow YAML to output
```

### Server & Webhooks

```
--server-url URL          Server URL for cron registration
--run-priority PRIORITY   low, normal, high, critical
--as-webhook              Register webhook trigger
--webhook-auth-key KEY    Webhook authentication key
--as-cron SCHEDULE        Create a cron schedule instead of executing (e.g., '0 2 * * *')
```

## Assets Command Flags

```
[search]                  Search term for filtering (positional argument)
-w, --workspace NAME      Filter by workspace name
--source SOURCE           Filter by source field (e.g., httpx, subfinder)
--type TYPE               Filter by asset_type field (e.g., web, subdomain)
--stats                   Show asset statistics (technologies, sources, types)
--limit N                 Maximum records to return (default: 50)
--offset N                Records to skip for pagination (default: 0)
--columns COLS            Comma-separated columns to display
--exclude-columns COLS    Comma-separated columns to exclude
--all                     Show all columns including hidden ones (id, timestamps)
--json                    Output in JSON format
```

## Worker Command Flags

### Worker Status

```
--columns COLS            Comma-separated columns to display
--exclude-columns COLS    Comma-separated columns to exclude
-s, --search TERM         Filter workers by substring (case-insensitive)
--redis-url URL           Redis connection URL override
```

### Worker Eval

```
-e, --eval SCRIPT         Script to evaluate
-t, --target TARGET       Target value for {{target}} variable
--params KEY=VALUE        Additional parameters (repeatable)
--stdin                   Read script from stdin
--redis-url URL           Redis connection URL override
```

### Worker Set

```
Usage: osmedeus worker set <worker-id-or-alias> <field> <value>
Valid fields: alias, public-ip, ssh-enabled, ssh-keys-path
```

### Worker Queue

```
# queue list
--redis-url URL           Redis connection URL override

# queue new
-f, --flow NAME           Flow workflow name
-m, --module NAME         Module workflow name
-t, --target TARGET       Target(s) to queue (repeatable)
-T, --target-file FILE    File containing targets
-p, --params KEY=VALUE    Additional parameters (repeatable)
--redis-url URL           Redis connection URL override

# queue run
--concurrency N           Concurrent task executors (default: 1)
--redis-url URL           Redis connection URL override
```

## Cloud Command Flags

### Cloud Config

```
# cloud config set <key> <value>
# cloud config list (alias: ls)
--show-secrets            Show sensitive values (default: hidden)
```

### Cloud Create

```
-p, --provider PROVIDER   Cloud provider (aws, gcp, digitalocean, linode, azure)
-m, --mode MODE           Execution mode (vm, serverless)
-n, --instances N         Number of instances to create
-f, --force               Force recreation of existing infrastructure
```

### Cloud Run

```
-p, --provider PROVIDER   Cloud provider
-m, --mode MODE           Execution mode
-n, --instances N         Number of instances
```

### Cloud Destroy

```
Usage: osmedeus cloud destroy [infrastructure-id]
```

### Cloud Setup

```
Usage: osmedeus cloud setup <ip> [ip2] [ip3] ...
--verbose-setup           Show full setup output
--ansible                 Use Ansible playbook for setup
```

## Query Command Flags

### Shared Flags (All Subcommands)

```
--limit N                 Maximum records to return (default: 50)
--offset N                Records to skip for pagination (default: 0)
--columns COLS            Comma-separated columns to display
--exclude-columns COLS    Comma-separated columns to exclude
--all                     Show all columns including hidden ones
--where KEY=VALUE         Filter by column (repeatable)
--search TERM             Search all columns (case-insensitive)
--json                    Output in JSON format (inherited global flag)
```

### Query Vulns

```
-w, --workspace NAME      Filter by workspace
--severity LEVEL          Filter: critical, high, medium, low, info
--confidence LEVEL        Filter: confirmed, firm, tentative
--asset VALUE             Filter by asset value (substring)
```

### Query Runs

```
-w, --workspace NAME      Filter by workspace
--status STATUS           Filter: pending, running, completed, failed, cancelled
--workflow NAME           Filter by workflow name
--target VALUE            Filter by target (substring)
```

### Query Steps

```
-r, --run UUID            Run UUID (required)
```

## Uninstall Command Flags

```
--clean                   Also remove workspaces data (~/workspaces-osmedeus)
```
