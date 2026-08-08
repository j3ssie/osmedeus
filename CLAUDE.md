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
make test-e2e-cloud     # Cloud E2E tests (cloud CLI commands)
make test-sudo          # Sudo-aware E2E tests (requires interactive sudo prompt)
make test-cloud         # Cloud integration tests (internal cloud package)
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
make update-ui          # Build platform/osmedeus-dashboard and refresh public/ui/
make update-ui DASHBOARD_SKIP_BUILD=1   # Reuse an existing dashboard build

# Platform sub-projects
make sync-platform                        # Publish platform/* OUT to standalone repos
make sync-platform PLATFORM=osmedeus-registry   # Just one
make sync-platform PLATFORM_COMMIT=1      # Also commit and push

# Agent Skills
make sync-skills                                # Push public/skills/ out to ../osmedeus-skills
make sync-skills SKILLS_DEST=/path/to/checkout  # Push to a checkout elsewhere

# Release
make bump-version       # Bump the patch in constants.go (v5.1.0 -> v5.1.1)
make bump-version PART=minor      # v5.1.0 -> v5.2.0 (also: major|pre|release)
make bump-version SET=v5.1.0-rc.1 # Set an explicit version
make bump-version DRY_RUN=1       # Preview without writing
make github-release     # Build and publish GitHub release (GoReleaser)
make npm-binaries       # Cross-compile the 4 npm target platforms
make npm-build          # Stage @j3ssie/osmedeus packages into build/dist-npm/
make npm-pack           # npm-build + inspectable .tgz tarballs
make npm-publish        # Publish to npm (needs NPM_TOKEN); DRY_RUN=1 to preview
```

## Architecture Overview

Osmedeus is a workflow engine for security automation. It executes YAML-defined workflows with support for multiple execution environments.

### Layered Architecture

```
CLI/API (pkg/cli, pkg/server)
         ↓
Executor (internal/executor) - coordinates workflow execution
         ↓
StepDispatcher - routes to: BashExecutor, FunctionExecutor, ForeachExecutor, ParallelExecutor, RemoteBashExecutor, HTTPExecutor, LLMExecutor, AgentExecutor, ACPExecutor
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
| `internal/cloud` | Cloud infrastructure provisioning (DigitalOcean, AWS, GCP, Linode, Azure) |
| `public` | Embedded assets: UI, presets, base example, and coding-agent skills |
| `platform` | Vendored sub-projects: dashboard, registry, workflow (not Go code) |

### Key Types

```go
WorkflowKind: "module" | "flow"  // module = single unit, flow = orchestrates modules
StepType: "bash" | "function" | "parallel-steps" | "foreach" | "remote-bash" | "http" | "llm" | "agent" | "agent-acp"
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
- Functions evaluated via Goja JS runtime: `file_exists()`, `file_length()`, `trim()`, `exec_python()`, `exec_ts()`, `detect_language()`, `extract_to()`, `db_import_sarif()`, `nmap_to_jsonl()`, `tmux_run()`, `ssh_exec()`, `skip()`, etc.

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

Preset tools: `bash`, `read_file`, `read_lines`, `file_exists`, `file_length`, `append_file`, `save_content`, `glob`, `grep_string`, `grep_regex`, `http_get`, `http_request`, `jq`, `exec_python`, `exec_python_file`, `exec_ts`, `exec_ts_file`, `run_module`, `run_flow`

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

### Agent-ACP Step Type

The `agent-acp` step type spawns an external AI coding agent as a subprocess and communicates via the Agent Communication Protocol (ACP). Unlike the `agent` step type (which uses the internal LLM loop), `agent-acp` delegates to real agent binaries.

Built-in agents (defined in `internal/executor/acp_executor.go`):
- `claude-code` — `npx -y @zed-industries/claude-code-acp@latest`
- `codex` — `npx -y @zed-industries/codex-acp`
- `opencode` — `opencode acp`
- `gemini` — `gemini --experimental-acp`

Key YAML fields:
- `agent` - Built-in agent name (required unless `acp_config.command` is set)
- `messages` - Conversation messages (role + content) used as the prompt
- `cwd` - Working directory for the ACP session
- `allowed_paths` - Restrict file reads to these directories
- `acp_config.command` - Custom agent command (overrides built-in registry)
- `acp_config.args` - Custom agent command arguments
- `acp_config.env` - Environment variables for the agent process
- `acp_config.write_enabled` - Allow file writes (default: false)

Available exports: `acp_output`, `acp_stderr`, `acp_agent`

```yaml
steps:
  - name: acp-agent
    type: agent-acp
    agent: claude-code
    cwd: "{{Output}}"
    allowed_paths:
      - "{{Output}}"
    acp_config:
      env:
        CUSTOM_VAR: "hello"
      write_enabled: true
    messages:
      - role: system
        content: "You are a security analyst."
      - role: user
        content: "Analyze the scan results in {{Output}} and create a summary."
    exports:
      analysis: "{{acp_output}}"
```

### Agent CLI Command

Run an ACP agent interactively from the terminal:
```bash
osmedeus agent "your message here"              # Run with claude-code (default)
osmedeus agent --agent codex "your message"     # Use a specific agent
osmedeus agent --list                            # List available agents
osmedeus agent --cwd /path/to/project "msg"     # Set working directory
osmedeus agent --timeout 1h "msg"               # Custom timeout (default: 30m)
echo "message" | osmedeus agent --stdin          # Read from stdin
```

### Org (Tenant) Layer

Orgs group multiple workspaces so assets, findings and runs can be queried across
all of them at once — a company with many root domains gets one org spanning every
workspace.

- **Model**: `Org` in `internal/database/models.go` (UUID pk, unique name). `org_uuid`
  is a denormalized column on `workspaces`, `assets`, `vulnerabilities` and `runs`
  (`orgScopedTables` in `internal/database/database.go`), so cross-workspace queries
  need no join.
- **Default org**: `DefaultOrgUUID` = `00000000-0000-0000-0000-000000000001`, seeded
  on migrate. Cannot be deleted or renamed.
- **Two different empty semantics** — this is the backward-compatibility contract:
  - *Read*: empty org means **no filter**, so queries span every org exactly as they
    did before orgs existed.
  - *Write*: empty org is coerced to the default org, never stored blank.
- **Migration**: `addOrgUUIDColumns` adds `org_uuid TEXT NOT NULL DEFAULT '<default>'`,
  so every pre-existing row is attributed to the default org automatically.
  `backfillOrgUUID` then reclaims rows holding an explicit empty string (Bun writes
  the Go zero value, bypassing the column DEFAULT).
- **Attribution is automatic**: `BeforeAppendModel` hooks in `internal/database/org_hooks.go`
  resolve an unset `org_uuid` from the row's workspace on insert *and* update, so
  importers need no org awareness and a re-scan never evicts a row from its org.
  Writes that deliberately set `org_uuid` use raw SQL without a model, so the hooks
  do not fire on them.
- **Resolution order** (`pkg/cli/org_resolve.go`): `--org` flag → `$OSMEDEUS_ORG_UUID`
  → `$OSMEDEUS_ORG` → `{{base_folder}}/.active-org` → unset (no filter).
  `--org` accepts a name or a UUID.
- **CLI**: `pkg/cli/org.go`. **API**: `pkg/server/handlers/orgs.go`, plus `?org=` on
  the assets, vulnerabilities, runs and workspaces endpoints.

Re-scanning a workspace without `--org` deliberately leaves its org alone;
only an explicit `--org` moves it (`EnsureWorkspaceRuntimeWithOrg`).

### Platform Sub-Projects

`platform/` holds the non-Go sub-projects that ship alongside the engine, vendored
into this repo so they version together with the code they talk to.

| Directory | Standalone repo | Purpose |
|-----------|-----------------|---------|
| `platform/osmedeus-dashboard` | `osmedeus/osmedeus-dashboard` | Next.js UI; built into `public/ui/` and `go:embed`ed |
| `platform/osmedeus-registry` | `osmedeus/osmedeus-registry` | Binary registry metadata and install scripts |
| `platform/osmedeus-workflow` | `osmedeus/osmedeus-workflow` | Public workflow collection |

- **This repo is the source of truth.** Edit under `platform/`, then
  `make sync-platform` publishes to the standalone repos. It writes files only —
  review and commit there yourself unless you pass `PLATFORM_COMMIT=1`.
- **No nested `.git`.** The sub-projects are plain directories here; their
  history lives in the standalone repos.
- **Build outputs are gitignored** (`node_modules/`, `.next/`, `build/`, `out/`).
  `.dockerignore` excludes `platform/` wholesale so an installed `node_modules`
  never enters the Docker build context.
- `.agents/` is un-ignored under `platform/` (`!platform/*/.agents/`) because the
  dashboard tracks its agent skills upstream — ignoring them here would make
  `sync-platform --delete` wipe them.
- **The dashboard is not the embedded UI.** `public/ui/` is the *built* output and
  is what gets embedded; `make update-ui` rebuilds it from `platform/`.

Targets that consume `platform/` rather than a sibling checkout:

| Target / file | Reads from |
|---------------|-----------|
| `make update-ui` | `platform/osmedeus-dashboard` (builds, then copies to `public/ui/`) |
| `make snapshot-release` | `$(REGISTRY_DIR)` = `platform/osmedeus-registry` for `registry-metadata-direct-fetch.json` and `install.sh` |
| `build/docker/docker-compose.canary.yaml` | `platform/osmedeus-workflow` mounted as the canary's workflows |

Adding a new consumer? Point it at `platform/<name>/`, never `../<name>/` — the
sibling checkout may not exist on a fresh clone.

Note `make sync-skills` is separate and stays that way: skills live in
`public/skills/` because they are `go:embed`ed, so they are not a `platform/`
sub-project. Both sync targets default to publishing into the directory that
holds this repo.

### Bundled Agent Skills

Skill bundles that teach an AI coding agent how to write osmedeus workflows and
drive the CLI are embedded in the binary and installed via `osmedeus skills install`.
Because they ship inside the binary, an installed skill always matches the running version.

- **Content**: `public/skills/<bundle>/` — `SKILL.md` (YAML frontmatter: `name`, `description`) plus optional `references/*.md`
- **Embed**: `//go:embed all:skills` in `public/embed.go`, read from `public.EmbedFS` under `skills/`
- **Implementation**: `pkg/cli/skills.go` — `list`, `get`, `install` subcommands
- **Discovery is filesystem-driven**: any directory under `public/skills/` containing a `SKILL.md` is a bundle, so adding one needs no code change
- **Install destinations**: `--agent claude` → `.claude/skills/`, `--agent codex|agents` → `.agents/skills/`; `--scope project` (cwd) or `global` (home)
- **`osmedeus install skills`** is a thin alias sharing `RunSkillsInstall`, mirroring the `workflow install` / `RunInstallWorkflow` pattern (a cobra command has one parent)

`public/skills/` is **authored in-tree** — it is the source of truth, so a skill
change ships in the same commit as the code it documents. The standalone repo
https://github.com/osmedeus/osmedeus-skills is a published mirror (it lets agents
install a skill without osmedeus); `make sync-skills` pushes this directory out to
a local checkout of it, writing bundle directories only — never the destination's
`README.md`, which is its own public landing page.

Installing into a project's `.claude/skills/` also benefits `osmedeus agent`, which
spawns claude-code/codex via ACP in that directory.

### npm Distribution

`npm install -g @j3ssie/osmedeus` ships the Go binary through npm. Everything
lives under `build/npm/` and is driven by the `npm-*` make targets.

- **One npm name, version-suffixed platform builds** (codex-style): the launcher
  publishes as `@j3ssie/osmedeus@<version>`, each platform build as
  `@j3ssie/osmedeus@<version>-<tag>` for the four tags `linux-x64`,
  `linux-arm64`, `darwin-x64`, `darwin-arm64`. The launcher pulls its own build
  in as an **aliased optionalDependency**
  (`"@j3ssie/osmedeus-linux-x64": "npm:@j3ssie/osmedeus@<version>-linux-x64"`),
  so an install downloads exactly one binary.
- **The binary ships gzipped** (`vendor/<tag>/osmedeus.gz`, ~50MB vs ~240MB raw)
  because the UI, presets and skills are embedded. `bin/osmedeus.js`
  decompresses it on first run into `~/.osmedeus/npm-bin/<version>/<tag>/`
  (override with `$OSMEDEUS_NPM_HOME`) — version-scoped, so an upgrade can never
  exec a stale binary — then `spawn`s it, forwarding args, stdio, signals and
  the exit status.
- **Publish order matters**: platform packages first, then the launcher, else its
  optionalDependencies do not resolve. `npm-publish` does this, then pins and
  verifies the `latest` dist-tag (retrying through registry cache lag).
- **Version source** is `VERSION` in `internal/core/constants.go`, minus the `v`
  — bump it with `make bump-version` (`build/scripts/bump-version.sh`), which is
  the only thing that should rewrite that constant. npm versions are immutable,
  so `build.mjs` refuses to pack unless the `.build-version` stamp
  `make npm-binaries` writes matches the version being published. Never verify a
  build by grepping the binary for a version string — the embedded
  docs/presets/UI mention other versions and it false-matches.
- **Files**: `build/npm/build.mjs` (staging), `build/npm/bin/osmedeus.js`
  (launcher — kept out of the `bin/` gitignore rule by a `!build/npm/bin/`
  negation), output in `build/dist-npm-bin/` (binaries) and `build/dist-npm/`
  (staged packages), both gitignored.

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
osmedeus worker join                             # Join as distributed worker (ID: wosm-<uuid8>)
osmedeus worker join --get-public-ip             # Join with public IP detection (alias: wosm-<ip>)
osmedeus install binary --name <name>            # Install specific binary
osmedeus install binary --all                    # Install all binaries
osmedeus install binary --name <name> --check    # Check if binary is installed
osmedeus install binary --all --check            # Check all binaries status
osmedeus install binary --nix-build-install      # Install binaries via Nix
osmedeus install binary --nix-installation       # Install Nix package manager
osmedeus install binary --list-registry-nix-build      # List Nix binaries
osmedeus install binary --list-registry-direct-fetch   # List direct-fetch binaries
osmedeus install base --preset                   # Install base from preset repository
osmedeus install base --preset --keep-setting    # Install base, restore previous osm-settings.yaml
osmedeus install workflow --preset               # Install workflows from preset repository
osmedeus install validate --preset               # Validate/install ready-to-use base
osmedeus install env                             # Add binaries to PATH (auto-detects shell)
osmedeus install env --all                       # Add to all shell configs
osmedeus skills                                  # List bundled coding-agent skills (alias: skill)
osmedeus skills list --json                      # List skills as JSON
osmedeus skills get <name>                       # Print a skill's SKILL.md to stdout
osmedeus skills get <name> --full                # Include reference files
osmedeus skills get --all                        # Print every bundled skill
osmedeus skills install                          # Install default skill into ./.claude/skills/
osmedeus skills install --scope global           # Install into ~/.claude/skills/
osmedeus skills install --agent codex            # Install into .agents/skills/ instead
osmedeus skills install --all --force            # Install every bundle, overwriting
osmedeus skills install --dir <path>             # Install to an explicit directory
osmedeus install skills                          # Alias for 'osmedeus skills install'
osmedeus update                                  # Self-update to latest version
osmedeus update --check                          # Check for updates without installing
osmedeus snapshot export <workspace>             # Export workspace as ZIP
osmedeus snapshot import <source>                # Import from file or URL
osmedeus snapshot list                           # List available snapshots
osmedeus run -m <module> -t <target> -G          # Run with progress bar (shorthand)
osmedeus run -f <flow> -t <target> -x <module>   # Exclude module(s) from flow
osmedeus run -f <flow> -t <target> -X <substr>   # Fuzzy-exclude modules by substring
osmedeus worker status                           # Show registered workers
osmedeus worker eval -e '<expr>'                 # Evaluate function with distributed hooks
osmedeus worker set <id> <field> <value>         # Update worker metadata
osmedeus worker queue list                       # List queued tasks
osmedeus worker queue new -f <flow> -t <target>  # Queue task for delayed execution
osmedeus worker queue run --concurrency 5        # Process queued tasks
osmedeus assets                                  # List discovered assets
osmedeus assets -w <workspace>                   # Filter assets by workspace
osmedeus assets --source httpx --type web        # Filter by source and type
osmedeus assets --stats                          # Show asset statistics
osmedeus assets --stats -w <workspace>           # Stats filtered by workspace
osmedeus assets --columns url,title,status_code  # Custom columns
osmedeus assets --limit 100 --offset 50          # Pagination
osmedeus assets --json                           # JSON output
osmedeus assets --org acme                       # Assets across every workspace in an org
osmedeus org                                     # List orgs (alias: orgs, tenant)
osmedeus org create acme -d "ACME Corp"          # Create an org
osmedeus org show acme                           # Org details, counts and workspaces
osmedeus org assign acme -w acme.com -w acme.io  # Group existing workspaces into an org
osmedeus org use acme                            # Set the active org (eval $(...) to export)
osmedeus org use --clear                         # Clear the active org
osmedeus org rename acme acme-corp               # Rename an org
osmedeus org delete acme                         # Delete org, data moves to default org
osmedeus org delete acme --purge                 # Delete org and all its data
osmedeus run -m <module> -t <target> --org acme  # Attribute a scan to an org
osmedeus agent "your prompt"                     # Run ACP agent (default: claude-code)
osmedeus agent --agent codex "your prompt"       # Use a specific agent
osmedeus agent --list                            # List available agents
osmedeus agent --cwd /path/to/project "prompt"   # Set working directory
osmedeus agent --timeout 1h "prompt"             # Custom timeout (default: 30m)
echo "prompt" | osmedeus agent --stdin           # Read from stdin
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
- **Orgs**: Group workspaces under a tenant; full CRUD + workspace assignment. `?org=` filters assets, vulnerabilities, runs and workspaces
- **Event Logs**: Query execution events
- **Functions**: Execute utility functions via API
- **Snapshots**: Export/import workspace archives
- **LLM**: OpenAI-compatible chat completions and embeddings
- **Agent ACP**: OpenAI-compatible endpoint that spawns local ACP agent subprocesses (`POST /osm/api/agent/chat/completions`)
- **Install**: Binary registry and installation management

## Cloud Documentation

Cloud infrastructure enables distributed security scanning across multiple providers (DigitalOcean, AWS, GCP, Linode, Azure):

- **Usage Examples**: `docs/cloud-usage-examples.md` - Comprehensive examples with copy-paste commands for all cloud operations
- **Quick Reference**: `docs/cloud-quick-reference.md` - Fast lookup for common commands and configurations
- **Cheatsheet**: `docs/cloud-cheatsheet.md` - Single-page printable reference card
- **Architecture**: `docs/cloud-usage-guide.md` - Cloud feature architecture and design
- **Test Documentation**: `test/e2e/CLOUD_TESTS_README.md` - Cloud test coverage and patterns
- **Config Template**: `public/presets/cloud-settings.example.yaml` - Full configuration example

Key cloud commands:
```bash
osmedeus cloud config set <key> <value>     # Configure cloud provider
osmedeus cloud create --instances N          # Provision infrastructure
osmedeus cloud list                          # List active infrastructure
osmedeus cloud run -f <flow> -t <target> --instances N  # Run distributed workflow
osmedeus cloud destroy <id>                  # Destroy infrastructure
```

## Workflow Hooks

Workflows support pre/post execution hooks via the `hooks` field:
```yaml
hooks:
  pre_scan_steps:
    - name: setup
      type: bash
      command: echo "Starting scan"
  post_scan_steps:
    - name: cleanup
      type: bash
      command: echo "Scan complete"
```
Hooks are defined using `WorkflowHooks` in `internal/core/workflow.go`. Pre-scan steps run before the main steps, post-scan steps run after completion.

## Queue System

Delayed task execution via database and Redis queues:
- `osmedeus worker queue new -f <flow> -t <target>` - Queue a task (creates Run with `is_queued=true`)
- `osmedeus worker queue run --concurrency N` - Process queued tasks with configurable parallelism
- Dual-source polling: database (every 5s) + Redis BRPOP (optional)
- Deduplication via runUUID tracking
- Implementation in `pkg/cli/worker_queue.go`

## Nmap Integration

Utility functions for nmap port scanning and result processing:
- `nmap_to_jsonl(input_path, output_path)` - Convert nmap XML/gnmap to JSONL format
- `run_nmap(target, flags?, output?)` - Execute nmap and auto-convert results to JSONL
- `db_import_port_assets(workspace, file_path, source?)` - Import port scan JSONL into database

## Tmux Session Management

Functions for managing long-running background processes:
- `tmux_run(command, session_name?)` - Create detached tmux session (auto-generates `bosm-<random8>` name)
- `tmux_capture(session_name)` - Capture pane output (pass `"all"` for all sessions)
- `tmux_send(session_name, command)` - Send keystrokes to session
- `tmux_kill(session_name)` / `tmux_list()` - Kill session / list all sessions

## SSH & Distributed Sync Functions

Functions for remote execution and file synchronization:
- `ssh_exec(host, command, user?, key_path?, password?, port?)` - Execute command via SSH (pooled connections)
- `ssh_rsync(host, src, dest, user?, key_path?, password?, port?)` - Copy files via rsync+SSH
- `sync_from_master(src, dest)` - Pull files from master node (falls back to local cp)
- `sync_from_worker(identifier, ip, src, dest)` / `rsync_to_worker(...)` - Sync with specific workers

## TypeScript Execution

- `exec_ts(code)` - Run inline TypeScript code via `bun -e`
- `exec_ts_file(path)` - Run a TypeScript file via `bun run`

## Skip Module Control

- `skip(message?)` - Abort remaining steps in current module; flow continues to next module
- Raises `ErrSkipModule` / `SkipModuleError` (defined in `internal/functions/constants.go`)

## Module Exclusion

- `-x, --exclude <module>` - Exclude module(s) from flow execution (exact match, repeatable)
- `-X, --fuzzy-exclude <substr>` - Exclude modules whose name contains substring (repeatable)

## Webhook Triggers

API endpoints for triggering runs via webhooks:
- `GET /osm/api/webhook-runs` - List webhook-enabled runs
- `GET|POST /osm/api/webhook-runs/{uuid}/trigger` - Trigger run via webhook UUID (unauthenticated)
- Runs store `webhook_uuid` and optional `webhook_auth_key` for authentication

## CDN/WAF Asset Classification

Assets now include CDN/WAF classification fields derived from httpx JSON data:
- `is_cdn` - Asset is behind a CDN (`cdn=true` or `cdn_name` non-empty in httpx)
- `is_cloud` - CDN name matches a cloud provider (AWS, GCP, Azure, etc.)
- `is_waf` - `cdn_type` equals "waf" in httpx data

## Adding New Features

**New Step Type**: Add constant in `core/types.go`, create executor implementing `StepExecutor` interface in `internal/executor/`, register in `PluginRegistry` via `dispatcher.go`

**New Runner**: Implement Runner interface in `internal/runner/`, add type constant, register in runner factory

**New CLI Command**: Create in `pkg/cli/`, add to `rootCmd` in `init()`

**New API Endpoint**: Add handler in `pkg/server/handlers/`, register route in `server.go`, document in `docs/api/`

**New Utility Function**: Add Go implementation in `internal/functions/`, register in `goja_runtime.go`, add constant in `constants.go`

**New Agent Preset Tool**: Add to `PresetToolRegistry` in `internal/core/agent_tool_presets.go`, add case in `buildPresetCallExpr()` in `internal/executor/agent_executor.go`

**New Bundled Skill**: Add a directory with a `SKILL.md` under `public/skills/`, then run `make sync-skills` to mirror it out to the skills repo. No Go changes needed — discovery finds any such directory.

## Architecture Notes

- **Executor**: Fresh instances created per target/request - no global singleton
- **Step Dispatcher**: Uses plugin registry pattern for extensible step type handling
- **Scheduler**: File watching uses fsnotify for instant inotify-based notifications
- **Decision Routing**: Uses switch/case syntax for conditional workflow branching
- **Run Registry**: Tracks active runs with PID management for cancellation support
- **Write Coordinator**: Batches database writes (step results, progress, artifacts) reducing I/O by ~70%
- **Install Base Backup**: `InstallBase()` automatically backs up `osm-settings.yaml` to `backup-osm-settings.yaml`; `--keep-setting` flag restores the previous settings after installation
- **Worker Identity**: Worker IDs use `wosm-<uuid8>` format; default alias is `wosm-<public-ip>` or `wosm-<local-ip>` when no `--alias` is provided
- **Execute Hooks**: Distributed coordination via `RegisterExecuteHooks()` in `internal/functions/execute_hooks.go` - avoids circular imports between functions and distributed packages
- **Queue System**: Dual-source polling (DB + Redis) with deduplication and configurable concurrency in `pkg/cli/worker_queue.go`
- **Command Fallback**: `internal/executor/cmd_fallback.go` handles timeout prefix stripping and custom binary path prepending
- **Bundled Skills**: `public/skills/` is the source of truth, embedded via `//go:embed all:skills`; edit in place, then `make sync-skills` pushes the bundles out to the standalone osmedeus-skills repo (a published mirror) for review and commit there
- **npm Distribution**: run `make bump-version` before `make npm-publish` — npm versions are immutable, so a version can only be superseded, never re-published. `make npm-binaries` restamps `build/dist-npm-bin/.build-version`; the guard in `build.mjs` fails closed if it does not match
- **Org Attribution**: never set `org_uuid` manually in a new import path — the `BeforeAppendModel` hooks derive it from the row's workspace. If you add an org-scoped table, add it to `orgScopedTables` and give it a `BeforeAppendModel` hook, or its rows will be invisible to every org query

## SARIF Integration

Utility functions for parsing SARIF (Static Analysis Results Interchange Format) output from SAST tools:
- `db_import_sarif(workspace, file_path)` - Import vulnerabilities from SARIF into database (supports Semgrep, Trivy, Kingfisher, Bearer)
- `convert_sarif_to_markdown(input_path, output_path)` - Convert SARIF to readable markdown tables
- `detect_language(path)` - Detect dominant programming language of a source folder (26+ languages)
- `extract_to(source, dest)` - Auto-detect archive format (.zip, .tar.gz, .tar.bz2, .tar.xz, .tgz) and extract
- `nmap_to_jsonl(input, output)` - Convert nmap XML/gnmap to JSONL
- `db_import_port_assets(workspace, file_path, source?)` - Import nmap JSONL into database as IP assets

## Performance Optimizations

- **Compiled JS caching**: Loop conditions compiled once and cached (60-80% faster)
- **Parallel shard rendering**: Template rendering uses parallel shards (20-40% faster startup)
- **Memory-mapped I/O**: Large files (>1MB) use mmap for 40-60% faster line counting
- **Efficient output buffering**: Runners use optimized buffer combining
