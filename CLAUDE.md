# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test Commands

```bash
# Build
make build              # Build to bin/osmedeus
make build-all          # Cross-platform builds (linux, darwin, windows)

# Test
make test-unit          # Fast unit tests (no external dependencies)
make test-integration   # Integration tests (requires Docker)
make test-e2e           # E2E CLI tests (requires binary build)
make test-e2e-ssh       # SSH E2E tests (module & step level SSH runner)
make test-e2e-api       # API E2E tests (all endpoints with Redis + seeded DB)
make test-distributed   # Distributed run e2e tests (requires Docker for Redis)
make test-docker        # Docker runner tests
make test-ssh           # SSH runner unit tests (starts test SSH container)
make test-canary-all    # Canary tests: real scans in Docker (30-60min)
make test-canary-repo   # Canary: SAST scan on juice-shop (~25min)
make test-canary-domain # Canary: domain recon on hackerone.com (~20min)
make test-canary-ip     # Canary: CIDR scan on IP list (~25min)
make test-canary-general # Canary: domain-list-recon on hackerone.com subdomains (~40min)
go test -v ./internal/functions/...  # Run tests for specific package
go test -v -run TestName ./...       # Run single test by name

# Development
make fmt                # Format code
make lint               # Run golangci-lint
make tidy               # go mod tidy
make run                # Build and run

# Installation
make install            # Install to $GOBIN (or $GOPATH/bin)
make swagger            # Generate Swagger documentation

# Docker Toolbox
make docker-toolbox       # Build toolbox image (all tools pre-installed)
make docker-toolbox-run   # Start toolbox container
make docker-toolbox-shell # Enter toolbox container shell

# Docker Canary (real-world scan testing)
make canary-up            # Build & start canary container
make canary-down          # Stop & cleanup canary container

# UI
make update-ui          # Update embedded UI from dashboard build
```

## Architecture Overview

Osmedeus is a workflow engine for security automation. It executes YAML-defined workflows with support for multiple execution environments.

### Layered Architecture

```
CLI/API (pkg/cli, pkg/server)
         ↓
Executor (internal/executor) - coordinates workflow execution
         ↓
StepDispatcher - routes to: BashExecutor, FunctionExecutor, ForeachExecutor, ParallelExecutor, RemoteBashExecutor, HTTPExecutor, LLMExecutor, AgentExecutor
         ↓
Runner (internal/runner) - executes commands via: HostRunner, DockerRunner, SSHRunner
```

### Core Packages

| Package | Purpose |
|---------|---------|
| `internal/core` | Type definitions: Workflow, Step, Trigger, RunnerConfig, ExecutionContext |
| `internal/parser` | YAML parsing, validation, and caching (Loader) |
| `internal/executor` | Workflow execution engine with step dispatching |
| `internal/runner` | Execution environments implementing Runner interface |
| `internal/template` | `{{Variable}}` interpolation engine |
| `internal/functions` | Utility functions via Goja JavaScript VM |
| `internal/scheduler` | Cron, event, and file-watch triggers (fsnotify-based) |
| `internal/database` | SQLite/PostgreSQL via Bun ORM |
| `pkg/cli` | Cobra CLI commands |
| `pkg/server` | Fiber REST API |
| `internal/snapshot` | Workspace export/import as compressed ZIP archives |
| `internal/installer` | Binary installation (direct-fetch and Nix modes) |
| `internal/state` | Run state export for debugging and sharing |
| `internal/updater` | Self-update functionality via GitHub releases |

### Key Types

```go
WorkflowKind: "module" | "flow"  // module = single unit, flow = orchestrates modules
StepType: "bash" | "function" | "parallel-steps" | "foreach" | "remote-bash" | "http" | "llm" | "agent"
RunnerType: "host" | "docker" | "ssh"
TriggerType: "cron" | "event" | "watch" | "manual"
```

### Decision Routing

Steps support conditional branching via `decision` field with switch/case syntax:
```yaml
decision:
  switch: "{{variable}}"
  cases:
    "value1": { goto: step-a }
    "value2": { goto: step-b }
  default: { goto: fallback }
```
Use `goto: _end` to terminate workflow.

### Workflow Execution Flow

1. CLI parses args ▷ loads config from `~/osmedeus-base/osm-settings.yaml`
2. Parser loads YAML workflow, validates, caches in Loader
3. Executor initializes context with built-in variables (`{{Target}}`, `{{Output}}`, etc.)
4. StepDispatcher routes each step to appropriate executor
5. Runner executes commands, captures output
6. Exports propagate to subsequent steps

### Template System

- `{{Variable}}` - standard template variables (Target, Output, threads, etc.)
- `[[variable]]` - foreach loop variables (to avoid conflicts)
- Functions evaluated via Goja JS runtime: `file_exists()`, `file_length()`, `trim()`, `exec_python()`, `detect_language()`, `extract_to()`, `db_import_sarif()`, etc.

### Platform Variables

Built-in variables for environment detection:
- `{{PlatformOS}}` - Operating system (linux, darwin, windows)
- `{{PlatformArch}}` - CPU architecture (amd64, arm64)
- `{{PlatformInDocker}}` - "true" if running in Docker container
- `{{PlatformInKubernetes}}` - "true" if running in Kubernetes pod
- `{{PlatformCloudProvider}}` - Cloud provider (aws, gcp, azure, local)

### Agent Step Type

The `agent` step type provides an agentic LLM execution loop with tool calling, sub-agent orchestration, and memory management.

Key YAML fields:
- `query` / `queries` - Task prompt (single or multi-goal)
- `agent_tools` - List of preset or custom tools available to the agent
- `max_iterations` - Maximum tool-calling loop iterations (required, > 0)
- `system_prompt` - System prompt for the agent
- `sub_agents` - Inline sub-agents spawnable via `spawn_agent` tool call
- `memory` - Sliding window config (`max_messages`, `summarize_on_truncate`, `persist_path`, `resume_path`)
- `models` - Preferred models tried in order before falling back to default
- `output_schema` - JSON schema enforced on final output
- `plan_prompt` - Optional planning stage prompt run before the main loop
- `stop_condition` - JS expression evaluated after each iteration
- `on_tool_start` / `on_tool_end` - JS hook expressions for tool call tracing
- `parallel_tool_calls` - Enable/disable parallel tool execution (default: true)

Preset tools: `bash`, `read_file`, `read_lines`, `file_exists`, `file_length`, `append_file`, `save_content`, `glob`, `grep_string`, `grep_regex`, `http_get`, `http_request`, `jq`, `exec_python`, `exec_python_file`, `run_module`, `run_flow`

Available exports: `agent_content`, `agent_history`, `agent_iterations`, `agent_total_tokens`, `agent_prompt_tokens`, `agent_completion_tokens`, `agent_tool_results`, `agent_plan`, `agent_goal_results`

```yaml
steps:
  - name: analyze-target
    type: agent
    query: "Enumerate subdomains of {{Target}} and summarize findings."
    system_prompt: "You are a security reconnaissance agent."
    max_iterations: 10
    agent_tools:
      - preset: bash
      - preset: read_file
      - preset: save_content
    memory:
      max_messages: 30
      persist_path: "{{Output}}/agent/conversation.json"
    exports:
      findings: "{{agent_content}}"
```

## CLI Commands

```bash
osmedeus run -f <flow> -t <target>              # Run flow workflow
osmedeus run -m <module> -t <target>            # Run module workflow
osmedeus run -m <m1> -m <m2> -t <target>        # Run multiple modules in sequence
osmedeus run -m <module> -t <target> --timeout 2h   # With timeout
osmedeus run -m <module> -t <target> --repeat       # Repeat continuously
osmedeus run -m <module> -T targets.txt -c 5    # Concurrent target scanning
osmedeus run -m <module> -t <target> -P params.yaml  # With params file
osmedeus workflow list                           # List available workflows
osmedeus workflow show <name>                    # Show workflow details
osmedeus workflow validate <name>                # Validate workflow YAML
osmedeus func list                               # List utility functions
osmedeus func e 'log_info("{{target}}")'         # Evaluate function
osmedeus --usage-example                         # Show all usage examples
osmedeus server                                  # Start REST API (see docs/api/ for endpoints)
osmedeus server --master                         # Start as distributed master
osmedeus worker join                             # Join as distributed worker
osmedeus install binary --name <name>            # Install specific binary
osmedeus install binary --all                    # Install all binaries
osmedeus install binary --name <name> --check    # Check if binary is installed
osmedeus install binary --all --check            # Check all binaries status
osmedeus install binary --nix-build-install      # Install binaries via Nix
osmedeus install binary --nix-installation       # Install Nix package manager
osmedeus install binary --list-registry-nix-build      # List Nix binaries
osmedeus install binary --list-registry-direct-fetch   # List direct-fetch binaries
osmedeus install base --preset                   # Install base from preset repository
osmedeus install workflow --preset               # Install workflows from preset repository
osmedeus install validate --preset               # Validate/install ready-to-use base
osmedeus install env                             # Add binaries to PATH (auto-detects shell)
osmedeus install env --all                       # Add to all shell configs
osmedeus update                                  # Self-update to latest version
osmedeus update --check                          # Check for updates without installing
osmedeus snapshot export <workspace>             # Export workspace as ZIP
osmedeus snapshot import <source>                # Import from file or URL
osmedeus snapshot list                           # List available snapshots
osmedeus run -m <module> -t <target> -G          # Run with progress bar (shorthand)
```

### Event Trigger Input Syntax

Event triggers support two syntaxes for extracting variables:

**New exports-style syntax (multiple variables):**
```yaml
triggers:
  - name: on-new-asset
    on: event
    event:
      topic: assets.new
    input:
      target: event_data.url
      description: trim(event_data.desc)
      source: event.source
```

**Legacy syntax (single input):**
```yaml
input:
  type: event_data
  field: url
  name: target
```

## API Documentation

REST API documentation with curl examples is in `docs/api/`. Key endpoint categories:
- **Runs**: Create, list, cancel (with PID termination), get steps/artifacts
- **Workflows**: List, get details, refresh index
- **Schedules**: Full CRUD + enable/disable/trigger
- **Assets/Workspaces**: Query discovered data
- **Event Logs**: Query execution events
- **Functions**: Execute utility functions via API
- **Snapshots**: Export/import workspace archives
- **LLM**: OpenAI-compatible chat completions and embeddings
- **Install**: Binary registry and installation management

## Adding New Features

**New Step Type**: Add constant in `core/types.go`, create executor implementing `StepExecutor` interface in `internal/executor/`, register in `PluginRegistry` via `dispatcher.go`

**New Runner**: Implement Runner interface in `internal/runner/`, add type constant, register in runner factory

**New CLI Command**: Create in `pkg/cli/`, add to `rootCmd` in `init()`

**New API Endpoint**: Add handler in `pkg/server/handlers/`, register route in `server.go`, document in `docs/api/`

**New Utility Function**: Add Go implementation in `internal/functions/`, register in `goja_runtime.go`, add constant in `constants.go`

**New Agent Preset Tool**: Add to `PresetToolRegistry` in `internal/core/agent_tool_presets.go`, add case in `buildPresetCallExpr()` in `internal/executor/agent_executor.go`

## Architecture Notes

- **Executor**: Fresh instances created per target/request - no global singleton
- **Step Dispatcher**: Uses plugin registry pattern for extensible step type handling
- **Scheduler**: File watching uses fsnotify for instant inotify-based notifications
- **Decision Routing**: Uses switch/case syntax for conditional workflow branching
- **Run Registry**: Tracks active runs with PID management for cancellation support
- **Write Coordinator**: Batches database writes (step results, progress, artifacts) reducing I/O by ~70%

## SARIF Integration

Utility functions for parsing SARIF (Static Analysis Results Interchange Format) output from SAST tools:
- `db_import_sarif(workspace, file_path)` - Import vulnerabilities from SARIF into database (supports Semgrep, Trivy, Kingfisher, Bearer)
- `convert_sarif_to_markdown(input_path, output_path)` - Convert SARIF to readable markdown tables
- `detect_language(path)` - Detect dominant programming language of a source folder (26+ languages)
- `extract_to(source, dest)` - Auto-detect archive format (.zip, .tar.gz, .tar.bz2, .tar.xz, .tgz) and extract

## Performance Optimizations

- **Compiled JS caching**: Loop conditions compiled once and cached (60-80% faster)
- **Parallel shard rendering**: Template rendering uses parallel shards (20-40% faster startup)
- **Memory-mapped I/O**: Large files (>1MB) use mmap for 40-60% faster line counting
- **Efficient output buffering**: Runners use optimized buffer combining
