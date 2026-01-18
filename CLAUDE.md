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
StepDispatcher - routes to: BashExecutor, FunctionExecutor, ForeachExecutor, ParallelExecutor, RemoteBashExecutor, HTTPExecutor, LLMExecutor
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
| `internal/functions` | Utility functions via Otto JavaScript VM |
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
WorkflowKind: "module" | "flow"      // module = single unit, flow = orchestrates modules
StepType: "bash" | "function" | "parallel-steps" | "foreach" | "remote-bash" | "http" | "llm"
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
- Functions evaluated via Otto JS runtime: `fileExists()`, `fileLength()`, `trim()`, etc.

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
osmedeus install env                             # Add binaries to PATH (auto-detects shell)
osmedeus install env --all                       # Add to all shell configs
osmedeus update                                  # Self-update to latest version
osmedeus update --check                          # Check for updates without installing
osmedeus snapshot export <workspace>             # Export workspace as ZIP
osmedeus snapshot import <source>                # Import from file or URL
osmedeus snapshot list                           # List available snapshots
osmedeus run -m <module> -t <target> -G          # Run with progress bar (shorthand)
```

## API Documentation

REST API documentation with curl examples is in `docs/api/`. Key endpoint categories:
- **Runs**: Create, list, cancel, get steps/artifacts
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

**New Utility Function**: Add Go implementation in `internal/functions/`, register in `otto_runtime.go`

## Architecture Notes

- **Executor**: Fresh instances created per target/request - no global singleton
- **Step Dispatcher**: Uses plugin registry pattern for extensible step type handling
- **Scheduler**: File watching uses fsnotify for instant inotify-based notifications
- **Decision Routing**: Uses switch/case syntax for conditional workflow branching
