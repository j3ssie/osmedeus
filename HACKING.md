# Hacking on Osmedeus

This document describes the technical architecture and development practices for Osmedeus. It's intended for developers who want to understand, modify, or extend the codebase.

## Table of Contents

- [Project Structure](#project-structure)
- [Architecture Overview](#architecture-overview)
- [Core Components](#core-components)
- [Workflow Engine](#workflow-engine)
- [Agent Step Type](#agent-step-type)
- [Execution Pipeline](#execution-pipeline)
- [Runner System](#runner-system)
- [Authentication Middleware](#authentication-middleware)
- [Template Engine](#template-engine)
- [Function Registry](#function-registry)
- [Scheduler System](#scheduler-system)
- [Workflow Linter](#workflow-linter)
- [Database Layer](#database-layer)
- [SARIF Integration](#sarif-integration)
- [Testing](#testing)
- [Canary Testing](#canary-testing)
- [Adding New Features](#adding-new-features)
- [CLI Shortcuts and Tips](#cli-shortcuts-and-tips)

## Project Structure

```
osmedeus/
├── cmd/osmedeus/           # Application entry point
├── internal/               # Private packages
│   ├── client/             # Remote API client
│   ├── config/             # Configuration management
│   ├── console/            # Console output capture
│   ├── core/               # Core types (Workflow, Step, Trigger, etc.)
│   ├── database/           # SQLite/PostgreSQL via Bun ORM
│   ├── distributed/        # Distributed execution (master/worker)
│   ├── executor/           # Workflow execution engine
│   ├── fileio/             # High-performance file I/O (mmap)
│   ├── functions/          # Utility functions (Goja JS runtime)
│   ├── heuristics/         # Target type detection
│   ├── installer/          # Binary installation (direct/Nix)
│   ├── linter/             # Workflow linting and validation
│   ├── logger/             # Structured logging (Zap)
│   ├── parser/             # YAML parsing and caching
│   ├── runner/             # Execution environments (host/docker/ssh)
│   ├── scheduler/          # Trigger scheduling (cron/event/watch)
│   ├── snapshot/           # Workspace export/import
│   ├── state/              # Run state export
│   ├── template/           # {{Variable}} interpolation engine
│   ├── terminal/           # Terminal UI (colors, tables, spinners)
│   ├── updater/            # Self-update via GitHub releases
│   └── workspace/          # Workspace management
├── lib/                    # Shared library utilities
├── pkg/                    # Public packages
│   ├── cli/                # Cobra CLI commands
│   └── server/             # Fiber REST API server
│       ├── handlers/       # Request handlers
│       └── middleware/     # Auth middleware (JWT, API Key)
├── public/                 # Public assets (examples, presets, UI)
├── test/                   # Test suites
│   ├── e2e/                # E2E CLI tests
│   ├── integration/        # Integration tests
│   └── testdata/           # Test workflow fixtures
├── docs/                   # API documentation
└── build/                  # Build artifacts and Docker files
```

## Architecture Overview

Osmedeus follows a layered architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI / API                            │
│              (pkg/cli, pkg/server)                          │
├─────────────────────────────────────────────────────────────┤
│                      Executor Layer                          │
│  ┌─────────────┐ ┌──────────────┐ ┌────────────────────┐   │
│  │  Executor   │ │  Dispatcher  │ │  Step Executors    │   │
│  │             │ │              │ │  (bash, function,  │   │
│  │             │ │              │ │   foreach, etc.)   │   │
│  └─────────────┘ └──────────────┘ └────────────────────┘   │
├─────────────────────────────────────────────────────────────┤
│                      Runner Layer                            │
│  ┌──────────────┐ ┌───────────────┐ ┌─────────────────┐    │
│  │ Host Runner  │ │ Docker Runner │ │   SSH Runner    │    │
│  └──────────────┘ └───────────────┘ └─────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                     Support Systems                          │
│  ┌──────────────┐ ┌───────────────┐ ┌─────────────────┐    │
│  │   Template   │ │   Functions   │ │   Scheduler     │    │
│  │   Engine     │ │   Registry    │ │   (triggers)    │    │
│  └──────────────┘ └───────────────┘ └─────────────────┘    │
├─────────────────────────────────────────────────────────────┤
│                      Data Layer                              │
│  ┌──────────────┐ ┌───────────────┐ ┌─────────────────┐    │
│  │   Parser/    │ │   Database    │ │   Workspace     │    │
│  │   Loader     │ │   (SQLite/PG) │ │   Manager       │    │
│  └──────────────┘ └───────────────┘ └─────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### Workflow Types

```go
// internal/core/workflow.go

type Workflow struct {
    Kind         WorkflowKind  // "module" or "flow"
    Name         string
    Description  string
    Params       []Param
    Triggers     []Trigger
    Runner       RunnerType
    RunnerConfig *RunnerConfig
    Steps        []Step        // For modules
    Modules      []ModuleRef   // For flows
}
```

**Module**: Single execution unit with sequential steps
**Flow**: Orchestrates multiple modules with dependency management

### Step Types

```go
// internal/core/step.go

type Step struct {
    Name             string
    Type             StepType      // bash, function, foreach, parallel-steps, remote-bash, http, llm, agent
    PreCondition     string        // Skip condition
    Command          string        // For bash/remote-bash
    Commands         []string      // Multiple commands
    Function         string        // For function type
    Input            string        // For foreach
    Variable         string        // Foreach variable name
    Threads          int           // Foreach parallelism
    Step             *Step         // Nested step for foreach
    ParallelSteps    []Step        // For parallel-steps type
    StepRunner       RunnerType    // For remote-bash: docker or ssh
    StepRunnerConfig *StepRunnerConfig // Runner config for remote-bash

    // Agent step fields
    Query             string             // Task prompt for the agent
    Queries           []string           // Multiple queries (multi-goal mode)
    SystemPrompt      string             // System prompt for the agent
    AgentTools        []AgentToolDef     // Preset or custom tools
    MaxIterations     int                // Max tool-calling loop iterations
    Models            []string           // Preferred models (tried in order)
    SubAgents         []SubAgentDef      // Inline sub-agents spawnable via spawn_agent
    MaxAgentDepth     int                // Max nesting depth for sub-agents (default: 3)
    Memory            *AgentMemoryConfig // Sliding window, summarization, persistence
    OutputSchema      string             // JSON schema for structured output
    StopCondition     string             // JS expression evaluated after each iteration
    PlanPrompt        string             // Planning stage prompt
    OnToolStart       string             // JS hook before each tool call
    OnToolEnd         string             // JS hook after each tool call

    Exports          map[string]string
    OnSuccess        []Action
    OnError          []Action
    Decision         *DecisionConfig   // Conditional branching (switch/case)
}
```

#### remote-bash Step Type

The `remote-bash` step type allows per-step Docker or SSH execution, independent of the module-level runner:

```yaml
steps:
  - name: docker-scan
    type: remote-bash
    step_runner: docker
    step_runner_config:
      image: alpine:latest
      volumes:
        - /data:/data
    command: nmap -sV {{target}}

  - name: ssh-scan
    type: remote-bash
    step_runner: ssh
    step_runner_config:
      host: "{{ssh_host}}"
      port: 22
      user: "{{ssh_user}}"
      key_file: ~/.ssh/id_rsa
    command: whoami && hostname
```

#### Decision Routing (Conditional Branching)

Steps can include decision routing to jump to different steps based on switch/case matching:

```yaml
steps:
  - name: detect-type
    type: bash
    command: echo "{{target_type}}"
    exports:
      detected_type: "output"
    decision:
      switch: "{{detected_type}}"
      cases:
        "domain":
          goto: subdomain-enum
        "ip":
          goto: port-scan
        "cidr":
          goto: network-scan
      default:
        goto: generic-recon

  - name: subdomain-enum
    type: bash
    command: subfinder -d {{target}}
    decision:
      switch: "always"
      cases:
        "always":
          goto: _end  # Special value to end workflow
```

The `_end` special value terminates workflow execution from the current step.

## Agent Step Type

The `agent` step type implements an agentic LLM execution loop. It sends a query to the LLM with available tools, executes tool calls returned by the LLM, feeds results back, and repeats until the LLM responds without tool calls or `max_iterations` is reached.

### Execution Flow

```
1. Planning stage (optional)     ──▶  LLM generates a plan from plan_prompt
2. Initialize conversation       ──▶  system_prompt + query + plan (if any)
3. Main agent loop:
   a. Send conversation to LLM (with tools)
   b. If no tool_calls → done
   c. Execute tool calls (parallel or sequential)
   d. Append tool results to conversation
   e. Evaluate stop_condition (if defined)
   f. Apply memory window (if configured)
   g. Repeat until max_iterations
4. Structured output (optional)  ──▶  Final LLM call with output_schema
5. Persist conversation (if memory.persist_path set)
```

### YAML Structure

```yaml
steps:
  - name: my-agent
    type: agent
    query: "Analyze {{Target}} and report findings."
    system_prompt: "You are a security analyst."
    max_iterations: 10
    agent_tools:
      - preset: bash
      - preset: read_file
      - preset: http_get
      - name: custom_tool
        description: "My custom tool"
        parameters:
          type: object
          properties:
            input:
              type: string
          required: [input]
        handler: 'process(args.input)'
    models:
      - gpt-4o
      - claude-sonnet-4-20250514
    memory:
      max_messages: 30
      summarize_on_truncate: true
      persist_path: "{{Output}}/agent/conversation.json"
      resume_path: "{{Output}}/agent/conversation.json"
    stop_condition: 'contains(agent_content, "DONE")'
    output_schema: '{"type":"object","properties":{"summary":{"type":"string"}}}'
    plan_prompt: "Create a plan for analyzing the target."
    on_tool_start: 'log_info("Tool: " + tool_name)'
    on_tool_end: 'log_info("Result: " + result)'
    parallel_tool_calls: true
    exports:
      findings: "{{agent_content}}"
```

### Preset Tools

All preset tools are defined in `PresetToolRegistry` (`internal/core/agent_tool_presets.go`):

| Preset | Description | Parameters |
|--------|-------------|------------|
| `bash` | Execute a shell command | `command` |
| `read_file` | Read file contents | `path` |
| `read_lines` | Read file as array of lines | `path` |
| `file_exists` | Check if a file exists | `path` |
| `file_length` | Count non-empty lines in a file | `path` |
| `append_file` | Append content from source to dest | `dest`, `content` |
| `save_content` | Write string content to a file | `content`, `path` |
| `glob` | Find files matching a glob pattern | `pattern` |
| `grep_string` | Search file for lines containing a string | `source`, `str` |
| `grep_regex` | Search file for lines matching a regex | `source`, `pattern` |
| `http_get` | Make an HTTP GET request | `url` |
| `http_request` | Make an HTTP request with method/headers/body | `url`, `method`, `body`?, `headers`? |
| `jq` | Query JSON data using jq syntax | `json_data`, `expression` |
| `exec_python` | Run inline Python code | `code` |
| `exec_python_file` | Run a Python file | `path` |
| `run_module` | Run an osmedeus module | `module`, `target`, `params`? |
| `run_flow` | Run an osmedeus flow | `flow`, `target`, `params`? |

### Custom Tool Definition

Custom tools use a JS handler expression. The parsed arguments are available as the `args` object:

```yaml
agent_tools:
  - name: check_domain
    description: "Validate if a string is a valid domain"
    parameters:
      type: object
      properties:
        domain:
          type: string
      required: [domain]
    handler: 'contains(args.domain, ".")'
```

### Sub-Agent Orchestration

Agents can delegate to sub-agents via the `spawn_agent` tool (automatically added when `sub_agents` is defined):

```yaml
steps:
  - name: orchestrator
    type: agent
    query: "Analyze {{Target}} by coordinating specialists"
    system_prompt: "You are an orchestrator. Delegate tasks to sub-agents."
    max_iterations: 10
    max_agent_depth: 3
    agent_tools:
      - preset: bash
    sub_agents:
      - name: recon_agent
        description: "Specialized agent for reconnaissance"
        system_prompt: "You are a recon specialist"
        max_iterations: 5
        agent_tools:
          - preset: bash
          - preset: http_get
      - name: vuln_scanner
        description: "Specialized agent for vulnerability scanning"
        max_iterations: 5
        agent_tools:
          - preset: bash
          - preset: read_file
```

Sub-agents are implemented via `SubAgentToolExecutor` in `internal/executor/tool_executor.go`. Child token counts are merged into the parent via `agentState.MergeTokens()`.

### Memory Management

- **Sliding window**: `max_messages` limits conversation history (system message always kept)
- **Summarization**: `summarize_on_truncate: true` uses LLM to summarize dropped messages
- **Persistence**: `persist_path` saves conversation JSON after completion
- **Resume**: `resume_path` loads a prior conversation on start

### Available Exports

| Export | Description |
|--------|-------------|
| `agent_content` | Final text output from the agent |
| `agent_history` | Full conversation history as JSON |
| `agent_iterations` | Number of iterations completed |
| `agent_total_tokens` | Total tokens used (including sub-agents) |
| `agent_prompt_tokens` | Prompt tokens used |
| `agent_completion_tokens` | Completion tokens used |
| `agent_tool_results` | All tool call results as JSON |
| `agent_plan` | Plan generated by planning stage (if used) |
| `agent_goal_results` | Results per query in multi-goal mode (JSON) |

### Tool Hooks

Hook expressions receive these variables:
- `tool_name` - Name of the tool being called
- `tool_args` - JSON string of tool arguments
- `result` - Tool result (empty in `on_tool_start`)
- `duration` - Execution time in ms (0 in `on_tool_start`)
- `iteration` - Current agent iteration number
- `error` - Error string (empty if no error)

### Execution Context



```go
// internal/core/context.go

type ExecutionContext struct {
    WorkflowName string
    WorkflowKind WorkflowKind
    RunID        string
    Target       string
    Variables    map[string]interface{}
    Params       map[string]string
    Exports      map[string]interface{}
    StepIndex    int
    Logger       *zap.Logger
}
```

The context is passed through the execution pipeline and accumulates state:
- Variables are set by the executor (built-in variables)
- Params are user-provided
- Exports are step outputs that propagate to subsequent steps

## Workflow Engine

### Parser

The parser (`internal/parser/parser.go`) handles YAML parsing:

```go
type Parser struct{}

func (p *Parser) Parse(path string) (*core.Workflow, error)
func (p *Parser) Validate(workflow *core.Workflow) error
```

### Loader

The loader (`internal/parser/loader.go`) provides caching and lookup:

```go
type Loader struct {
    workflowsDir string
    modulesDir   string
    cache        map[string]*core.Workflow
}

func (l *Loader) LoadWorkflow(name string) (*core.Workflow, error)
func (l *Loader) ListFlows() ([]string, error)
func (l *Loader) ListModules() ([]string, error)
```

Lookup order:
1. Check cache
2. Try `workflows/<name>.yaml`
3. Try `workflows/<name>-flow.yaml`
4. Try `workflows/modules/<name>.yaml`
5. Try `workflows/modules/<name>-module.yaml`

## Execution Pipeline

### Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   CLI/API    │────▶│   Executor   │────▶│  Dispatcher  │
└──────────────┘     └──────────────┘     └──────────────┘
                                                 │
                    ┌────────────────────────────┼────────────────────────────┐
                    │                            │                            │
                    ▼                            ▼                            ▼
             ┌──────────────┐            ┌──────────────┐            ┌──────────────┐
             │ BashExecutor │            │FunctionExec  │            │ForeachExec   │
             └──────────────┘            └──────────────┘            └──────────────┘
             ┌──────────────┐            ┌──────────────┐            ┌──────────────┐
             │ HTTPExecutor │            │ LLMExecutor  │            │AgentExecutor │
             └──────────────┘            └──────────────┘            └──────────────┘
                    │                            │                            │
                    └────────────────────────────┼────────────────────────────┘
                                                 ▼
                                          ┌──────────────┐
                                          │    Runner    │
                                          └──────────────┘
```

### Executor

```go
// internal/executor/executor.go

type Executor struct {
    templateEngine   *template.Engine
    functionRegistry *functions.Registry
    stepDispatcher   *StepDispatcher
}

func (e *Executor) ExecuteModule(ctx context.Context, module *core.Workflow,
                                  params map[string]string, cfg *config.Config) (*core.WorkflowResult, error)
func (e *Executor) ExecuteFlow(ctx context.Context, flow *core.Workflow,
                                params map[string]string, cfg *config.Config) (*core.WorkflowResult, error)
```

Key responsibilities:
1. Initialize execution context with built-in variables
2. Create and setup the appropriate runner
3. Iterate through steps, dispatching to appropriate handler
4. Handle pre-conditions, exports, and decision routing
5. Process on_success/on_error actions

### Step Dispatcher

The dispatcher uses a plugin registry pattern for extensible step type handling:

```go
// internal/executor/dispatcher.go

type StepDispatcher struct {
    registry         *PluginRegistry     // Extensible executor registry
    templateEngine   *template.Engine
    functionRegistry *functions.Registry
    bashExecutor     *BashExecutor       // Registered as plugin
    llmExecutor      *LLMExecutor        // Registered as plugin
    runner           runner.Runner
}

// PluginRegistry manages step type executors
type PluginRegistry struct {
    executors map[core.StepType]StepExecutor
}

// StepExecutor interface for all step type handlers
type StepExecutor interface {
    CanHandle(stepType core.StepType) bool
    Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext, runner runner.Runner) (*core.StepResult, error)
}

func (d *StepDispatcher) Dispatch(ctx context.Context, step *core.Step,
                                   execCtx *core.ExecutionContext) (*core.StepResult, error)
```

Built-in executors registered at startup:
- `BashExecutor` - handles `bash` steps
- `FunctionExecutor` - handles `function` steps
- `ForeachExecutor` - handles `foreach` steps
- `ParallelExecutor` - handles `parallel-steps` steps
- `RemoteBashExecutor` - handles `remote-bash` steps
- `HTTPExecutor` - handles `http` steps
- `LLMExecutor` - handles `llm` steps
- `AgentExecutor` - handles `agent` steps (agentic LLM loop with tool calling)

## Run Control Plane

The run control plane tracks active workflow executions for cancellation support:

```go
// internal/executor/run_control_plane.go

type RunControlPlane struct {
    mu   sync.RWMutex
    runs map[string]*ActiveRun
}

type ActiveRun struct {
    RunUUID   string
    Cancel    context.CancelFunc
    PIDs      *sync.Map  // Currently running process IDs
    StartedAt time.Time
}

// Key operations
func (r *RunControlPlane) Register(runUUID string, cancel context.CancelFunc) *ActiveRun
func (r *RunControlPlane) Cancel(runUUID string) ([]int, error)  // Returns killed PIDs
func (r *RunControlPlane) AddPID(runUUID string, pid int)
func (r *RunControlPlane) RemovePID(runUUID string, pid int)
```

The control plane is accessed via `GetRunControlPlane()` singleton. When a run is cancelled:
1. Context is cancelled to stop new operations
2. All tracked PIDs are killed via SIGKILL (including process groups)

## Runner System

### Interface

```go
// internal/runner/runner.go

type Runner interface {
    Execute(ctx context.Context, command string) (*CommandResult, error)
    Setup(ctx context.Context) error
    Cleanup(ctx context.Context) error
    Type() core.RunnerType
    IsRemote() bool
}

type CommandResult struct {
    Output   string
    ExitCode int
    Error    error
}
```

### Host Runner

Simple local execution using `os/exec`:

```go
func (r *HostRunner) Execute(ctx context.Context, command string) (*CommandResult, error) {
    cmd := exec.CommandContext(ctx, "sh", "-c", command)
    // ... execute and capture output
}
```

### Docker Runner

Supports both ephemeral (`docker run --rm`) and persistent (`docker exec`) modes:

```go
type DockerRunner struct {
    config      *core.RunnerConfig
    containerID string  // For persistent mode
}

func (r *DockerRunner) Execute(ctx context.Context, command string) (*CommandResult, error) {
    if r.config.Persistent && r.containerID != "" {
        return r.execInContainer(ctx, command)
    }
    return r.runEphemeral(ctx, command)
}
```

### SSH Runner

Uses `golang.org/x/crypto/ssh` for remote execution:

```go
type SSHRunner struct {
    config *core.RunnerConfig
    client *ssh.Client
}

func (r *SSHRunner) Setup(ctx context.Context) error {
    // Build auth methods (key or password)
    // Establish SSH connection
    // Optionally copy binary to remote
}
```

## Authentication Middleware

### Auth Types

The server supports two authentication methods:

| Method | Header | Description |
|--------|--------|-------------|
| API Key | `x-osm-api-key` | Simple token-based auth |
| JWT | `Authorization: Bearer <token>` | Token from `/osm/api/login` |

### Priority Logic

```go
// pkg/server/server.go - setupRoutes()

if s.config.Server.EnabledAuthAPI {
    api.Use(middleware.APIKeyAuth(s.config))
} else if !s.options.NoAuth {
    api.Use(middleware.JWTAuth(s.config))
}
```

Priority order:
1. **API Key Auth** - If `EnabledAuthAPI` is true
2. **JWT Auth** - If API key auth disabled and NoAuth is false
3. **No Auth** - If NoAuth option is true

### APIKeyAuth Implementation

```go
// pkg/server/middleware/auth.go

func APIKeyAuth(cfg *config.Config) fiber.Handler {
    return func(c *fiber.Ctx) error {
        apiKey := c.Get("x-osm-api-key")
        if !isValidAPIKey(apiKey, cfg.Server.AuthAPIKey) {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error":   true,
                "message": "Invalid or missing API key",
            })
        }
        return c.Next()
    }
}
```

Security features:
- Case-sensitive exact matching
- Rejects empty/whitespace-only keys
- Rejects placeholder values ("null", "undefined", "nil")

## Template Engine

### Variable Resolution

The template engine (`internal/template/engine.go`) handles `{{variable}}` interpolation:

```go
type Engine struct{}

func (e *Engine) Render(template string, ctx map[string]interface{}) (string, error)
```

Resolution order:
1. Check context variables
2. Check environment variables (optional)
3. Return empty string if not found

### Built-in Variable Injection

```go
// internal/executor/executor.go

func (e *Executor) injectBuiltinVariables(cfg *config.Config, params map[string]string,
                                           execCtx *core.ExecutionContext) {
    execCtx.SetVariable("BaseFolder", cfg.BaseFolder)
    execCtx.SetVariable("Target", params["target"])
    execCtx.SetVariable("Output", filepath.Join(workspacesPath, targetSpace))
    execCtx.SetVariable("threads", threads)
    execCtx.SetVariable("RunUUID", execCtx.RunUUID)
    // ... more variables
}
```

### Foreach Variable Syntax

Foreach uses `[[variable]]` syntax (double brackets) to avoid conflicts with template variables:

```yaml
- name: process-items
  type: foreach
  input: "/path/to/items.txt"
  variable: item
  step:
    command: echo [[item]]  # Replaced during foreach iteration
```

## Function Registry

### Goja JavaScript Runtime

Functions are implemented in Go and exposed to a Goja JavaScript VM with pooling for performance:

```go
// internal/functions/goja_runtime.go

type GojaRuntime struct {
    pool *GojaPool  // Pool of pre-warmed VMs
}

func NewGojaRuntime() *GojaRuntime {
    pool := NewGojaPool(4)  // Pool size
    return &GojaRuntime{pool: pool}
}

// Pool provides thread-safe VM reuse with compiled program caching
type GojaPool struct {
    vms           []*gojaVM
    compiledProgs sync.Map  // Cached compiled JS programs
}
```

### Adding New Functions

1. Add the Go implementation in the appropriate file:

```go
// internal/functions/file_functions.go

func (vf *vmFunc) myNewFunction(call goja.FunctionCall) goja.Value {
    arg := call.Argument(0).String()
    // ... implementation
    return vf.vm.ToValue(output)
}
```

2. Register in `goja_runtime.go`:

```go
_ = vm.Set("my_new_function", vf.myNewFunction)
```

3. Add constant in `constants.go`:

```go
const FnMyNewFunction = "my_new_function"
```

Notable utility functions include `exec_python(code)` and `exec_python_file(path)` for running Python code, and `run_module(module, target, params)` / `run_flow(flow, target, params)` for launching osmedeus workflows as subprocesses.

### Output and Control Functions

These functions provide output and execution control within workflows:

```go
// internal/functions/util_functions.go

// printf prints a message to stdout
func (r *OttoRuntime) printf(call otto.FunctionCall) otto.Value

// catFile prints file content to stdout
func (r *OttoRuntime) catFile(call otto.FunctionCall) otto.Value

// exit exits the scan with given code (0=success, non-zero=error)
func (r *OttoRuntime) exit(call otto.FunctionCall) otto.Value
```

Usage in workflows:
```yaml
steps:
  - name: print-status
    type: function
    function: printf("Scan completed for {{Target}}")

  - name: show-results
    type: function
    function: cat_file("{{Output}}/results.txt")
```

### Event Functions

These functions enable event-driven workflows by generating and emitting events:

```go
// internal/functions/event_functions.go

// generate_event emits a single structured event
// Usage: generate_event(workspace, topic, source, data_type, data)
func (vf *vmFunc) generateEvent(call goja.FunctionCall) goja.Value

// generate_event_from_file emits an event for each line in a file
// Usage: generate_event_from_file(workspace, topic, source, data_type, filePath)
func (vf *vmFunc) generateEventFromFile(call goja.FunctionCall) goja.Value
```

Usage in workflows:
```yaml
steps:
  - name: emit-single-event
    type: function
    function: |
      generate_event("{{Workspace}}", "assets.new", "scanner", "subdomain", "api.example.com")

  - name: emit-from-file
    type: function
    function: |
      generate_event_from_file("{{Workspace}}", "assets.new", "recon", "subdomain", "{{Output}}/subdomains.txt")
```

Event delivery uses a fallback chain:
1. **Server API** - POST to `/osm/api/events/emit` if server configured
2. **Redis Pub/Sub** - Publish to `osm:events:{topic}` in distributed mode
3. **Database Queue** - Store in `event_logs` table with `processed=false`
4. **Webhooks** - Send to configured webhook endpoints

### SARIF Functions

Functions for parsing SARIF (Static Analysis Results Interchange Format) output from SAST tools:

```go
// internal/functions/sarif_functions.go

// db_import_sarif imports vulnerabilities from a SARIF file into the database
// Supports output from: Semgrep, Trivy, Kingfisher, Bearer
// Usage: db_import_sarif(workspace, file_path) -> {new, updated, unchanged, errors, total}
func (vf *vmFunc) dbImportSARIF(call goja.FunctionCall) goja.Value

// convert_sarif_to_markdown converts a SARIF file to a markdown table
// Usage: convert_sarif_to_markdown(input_path, output_path) -> bool
func (vf *vmFunc) convertSARIFToMarkdown(call goja.FunctionCall) goja.Value
```

Usage in workflows:
```yaml
steps:
  - name: import-sarif
    type: function
    function: |
      db_import_sarif("{{Workspace}}", "{{Output}}/semgrep.sarif")

  - name: sarif-report
    type: function
    function: |
      convert_sarif_to_markdown("{{Output}}/trivy.sarif", "{{Output}}/trivy-report.md")
```

SARIF severity mapping: `error` → high, `warning` → medium, `note` → low, `none` → info.

### Type Detection and Archive Functions

```go
// internal/functions/type_functions.go

// detect_language detects the dominant programming language in a directory
// Supports 26+ languages via file extension and shebang analysis
// Skips non-source dirs: node_modules, vendor, .git, __pycache__, etc.
// Usage: detect_language(path) -> string ("golang", "python", "javascript", etc.)
func (vf *vmFunc) detectLanguage(call goja.FunctionCall) goja.Value
```

```go
// internal/functions/file_functions.go

// extract_to auto-detects archive format and extracts to destination
// Supports: .zip, .tar.gz, .tgz, .tar.bz2, .tar.xz
// Removes destination directory first (idempotent)
// Usage: extract_to(source, dest) -> bool
func (vf *vmFunc) extractTo(call goja.FunctionCall) goja.Value
```

Usage in workflows:
```yaml
steps:
  - name: detect-lang
    type: function
    function: |
      detect_language("{{Output}}/repo")
    exports:
      lang: "output"

  - name: extract-repo
    type: function
    function: |
      extract_to("/tmp/repo.tar.gz", "{{Output}}/repo")
```

### Function Execution

```go
// internal/functions/registry.go

func (r *Registry) Execute(expr string, ctx map[string]interface{}) (interface{}, error) {
    return r.runtime.Execute(expr, ctx)
}

func (r *Registry) EvaluateCondition(condition string, ctx map[string]interface{}) (bool, error) {
    return r.runtime.EvaluateCondition(condition, ctx)
}
```

## Scheduler System

### Trigger Types

```go
// internal/core/trigger.go

type TriggerType string

const (
    TriggerManual TriggerType = "manual"
    TriggerCron   TriggerType = "cron"
    TriggerEvent  TriggerType = "event"
    TriggerWatch  TriggerType = "watch"
)
```

### Scheduler

The scheduler manages workflow triggers using gocron for cron jobs and fsnotify for file watching:

```go
// internal/scheduler/scheduler.go

type Scheduler struct {
    scheduler  gocron.Scheduler
    triggers   map[string]*RegisteredTrigger
    handlers   map[string]TriggerHandler
    events     chan *core.Event

    // File watcher (fsnotify-based)
    watcher    *fsnotify.Watcher
    watchPaths map[string][]*RegisteredTrigger  // path → triggers mapping
}

func (s *Scheduler) RegisterTrigger(workflow *core.Workflow, trigger *core.Trigger) error
func (s *Scheduler) EmitEvent(event *core.Event) error
func (s *Scheduler) Start() error   // Starts cron scheduler, file watcher, and event listener
func (s *Scheduler) Stop() error    // Stops all and closes watcher
```

File watching uses fsnotify for instant inotify-based notifications (sub-millisecond latency) instead of polling.

### Event Trigger Input Syntax

Event triggers support two syntaxes for extracting variables from events:

**New exports-style syntax (recommended for multiple variables):**
```yaml
triggers:
  - name: on-new-asset
    on: event
    event:
      topic: assets.new
      filters:
        - "event.data_type == 'subdomain'"
      filter_functions:
        - "contains(event_data.url, '/api/')"  # Utility functions available
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

The `filter_functions` field allows using utility functions (like `contains()`, `starts_with()`, etc.) in filters, while `filters` uses plain JavaScript expressions.

### Event Envelope

The full event context is available in triggered workflows via the `event` object:
- `event.topic` - Event topic
- `event.source` - Event source
- `event.data_type` - Data type
- `event_data.*` - Event data fields (shorthand for `event.data.*`)

### Event Filtering

Events are matched using JavaScript expressions with Goja runtime:

```go
func (s *Scheduler) evaluateFilters(filters []string, event *core.Event) bool {
    vm := goja.New()
    vm.Set("event", eventObj)
    vm.Set("event_data", event.Data)  // Shorthand access

    for _, filter := range filters {
        result, _ := vm.RunString(filter)
        if !result.ToBoolean() {
            return false
        }
    }
    return true
}
```

## Workflow Linter

The workflow linter (`internal/linter/`) provides static analysis of workflow YAML files to catch common issues before execution.

### Usage

```bash
# Lint a single workflow
osmedeus workflow lint my-workflow.yaml

# Lint by workflow name (searches in workflows path)
osmedeus workflow lint my-workflow

# Lint all workflows in a directory
osmedeus workflow lint /path/to/workflows/

# Output formats
osmedeus workflow lint my-workflow.yaml --format pretty   # Default, colored output
osmedeus workflow lint my-workflow.yaml --format json     # Machine-readable JSON
osmedeus workflow lint my-workflow.yaml --format github   # GitHub Actions annotations

# Filter by severity
osmedeus workflow lint my-workflow.yaml --severity warning  # Show warnings and above
osmedeus workflow lint my-workflow.yaml --severity error    # Show only errors

# Disable specific rules
osmedeus workflow lint my-workflow.yaml --disable unused-variable,empty-step

# CI mode (exit with error code if issues found)
osmedeus workflow lint my-workflow.yaml --check
```

### Severity Levels

| Severity | Description | Exit Code |
|----------|-------------|-----------|
| **info** | Best practice suggestions (e.g., unused exports) | 0 |
| **warning** | Potential issues that may cause problems | 0 |
| **error** | Critical issues that will likely cause failures | 1 (with --check) |

### Built-in Rules

| Rule | Severity | Description |
|------|----------|-------------|
| `missing-required-field` | warning | Detects missing required fields (name, kind, type) |
| `duplicate-step-name` | warning | Detects multiple steps with the same name |
| `empty-step` | warning | Detects steps with no executable content |
| `unused-variable` | info | Detects exports that are never referenced |
| `invalid-goto` | warning | Detects decision goto references to non-existent steps |
| `invalid-depends-on` | warning | Detects depends_on references to non-existent steps |
| `circular-dependency` | warning | Detects circular references in step dependencies |

**Note**: The `undefined-variable` rule is available but not enabled by default as it can produce false positives for dynamically-injected variables.

### Built-in Variables

The linter recognizes all runtime-injected variables to avoid false positives. These include:

**Path Variables**: `BaseFolder`, `Binaries`, `Data`, `ExternalData`, `ExternalConfigs`, `Workflows`, `Workspaces`, etc.

**Target Variables**: `Target`, `target`, `TargetFile`, `TargetSpace`

**Output Variables**: `Output`, `output`, `Workspace`, `workspace`

**Metadata Variables**: `Version`, `RunUUID`, `TaskDate`, `TimeStamp`, `Today`, `RandomString`

**Heuristic Variables**: `TargetType`, `TargetRootDomain`, `TargetTLD`, `Org`, `TargetHost`, `TargetPort`, etc.

**Chunk Variables**: `ChunkIndex`, `ChunkSize`, `TotalChunks`, `ChunkStart`, `ChunkEnd`

### Linter Architecture

```go
// internal/linter/linter.go

type Linter struct {
    rules   []LinterRule
    options LinterOptions
}

// LinterRule interface for all lint rules
type LinterRule interface {
    Name() string
    Description() string
    Severity() Severity
    Check(ast *WorkflowAST) []LintIssue
}

func (l *Linter) Lint(path string) (*LintResult, error)
func (l *Linter) LintContent(content []byte, filename string) (*LintResult, error)
```

### Adding a New Lint Rule

1. Create the rule in `internal/linter/rules.go`:

```go
type MyNewRule struct{}

func (r *MyNewRule) Name() string        { return "my-new-rule" }
func (r *MyNewRule) Description() string { return "Detects my issue" }
func (r *MyNewRule) Severity() Severity  { return SeverityWarning }

func (r *MyNewRule) Check(wast *WorkflowAST) []LintIssue {
    var issues []LintIssue
    // ... implementation
    return issues
}
```

2. Register in `GetDefaultRules()`:

```go
func GetDefaultRules() []LinterRule {
    return []LinterRule{
        // ... existing rules
        &MyNewRule{},
    }
}
```

## Write Coordinator

The write coordinator batches database operations to reduce I/O by ~70%:

```go
// internal/database/write_coordinator.go

type WriteCoordinator struct {
    runID          int64
    stepResults    []*StepResult      // Buffered step results
    progressDelta  int                // Accumulated progress updates
    artifacts      []*Artifact        // Buffered artifacts
    flushThreshold int                // Flush after N step results (default: 10)
    flushInterval  time.Duration      // Flush every interval (default: 5s)
}

// Usage
wc := NewWriteCoordinator(runID, runUUID, nil)  // nil uses defaults
defer wc.Close()  // Final flush

wc.AddStepResult(stepName, stepType, status, command, output, ...)
wc.IncrementProgress(1)
wc.AddArtifact(path, artifactType)
```

## Platform Detection

Platform detection functions for environment-aware workflows:

```go
// internal/executor/platform.go

func DetectDocker() bool       // Checks /.dockerenv and /proc/1/cgroup
func DetectKubernetes() bool   // Checks service account directory
func DetectCloudProvider() string  // Returns: aws, gcp, azure, or local
```

These are exposed as template variables:
- `{{PlatformOS}}` - runtime.GOOS
- `{{PlatformArch}}` - runtime.GOARCH
- `{{PlatformInDocker}}` - "true" or "false"
- `{{PlatformInKubernetes}}` - "true" or "false"
- `{{PlatformCloudProvider}}` - aws/gcp/azure/local

## Database Layer

### Multi-Engine Support

```go
// internal/database/database.go

func Connect(cfg *config.Config) (*bun.DB, error) {
    switch {
    case cfg.IsPostgres():
        return connectPostgres(cfg)
    case cfg.IsSQLite():
        return connectSQLite(cfg)
    default:
        return nil, fmt.Errorf("unsupported database engine")
    }
}
```

### Models

```go
// internal/database/models.go

type Run struct {
    ID             string
    RunID          string
    WorkflowName   string
    WorkflowKind   string    // "flow" or "module"
    Target         string
    Params         map[string]string
    Status         string    // "pending", "running", "completed", "failed"
    Workspace      string    // Logical workspace name (same as TargetSpace)
    StartedAt      time.Time
    CompletedAt    time.Time
    ErrorMessage   string
    ScheduleID     string
    TriggerType    string    // "manual", "cron", "event", "api"
    TriggerName    string
    TotalSteps     int
    CompletedSteps int
    RunPriority    string    // "low", "normal", "high", "critical"
    RunMode        string    // "local", "distributed", "cloud"
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type Asset struct {
    ID            int64
    Workspace     string
    AssetValue    string    // Primary identifier (hostname)
    URL           string
    Input         string
    Scheme        string    // "http", "https"
    Method        string
    Path          string
    StatusCode    int
    ContentType   string
    ContentLength int64
    Title         string
    Words         int
    Lines         int
    HostIP        string
    A             []string  // DNS A records (JSON)
    TLS           string
    AssetType     string
    Tech          []string  // Technologies (JSON)
    Time          string    // Response time
    Remarks       string    // Labels
    Source        string    // Discovery source
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type Workspace struct {
    ID              int64
    Name            string
    LocalPath       string
    TotalAssets     int
    TotalSubdomains int
    TotalURLs       int
    TotalVulns      int
    VulnCritical    int
    VulnHigh        int
    VulnMedium      int
    VulnLow         int
    VulnPotential   int
    RiskScore       float64
    Tags            []string  // JSON array
    LastRun         time.Time
    RunWorkflow     string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type EventLog struct {
    ID           int64
    Topic        string    // "run.started", "run.completed", "asset.discovered", etc.
    EventID      string
    Name         string
    Source       string    // "executor", "scheduler", "api"
    DataType     string
    Data         string    // JSON payload
    Workspace    string
    RunID        string
    WorkflowName string
    Processed    bool
    ProcessedAt  time.Time
    Error        string
    CreatedAt    time.Time
}

type Schedule struct {
    ID           string
    Name         string
    WorkflowName string
    WorkflowPath string
    TriggerName  string
    TriggerType  string    // "cron", "event", "watch"
    Schedule     string    // Cron expression
    EventTopic   string
    WatchPath    string
    Target       string    // Default target for scheduled runs
    Workspace    string    // Default workspace
    Params       map[string]string  // Additional parameters (JSON)
    InputConfig  map[string]string  // JSON params (deprecated, use Params)
    IsEnabled    bool
    LastRun      time.Time
    NextRun      time.Time
    RunCount     int
    CreatedAt    time.Time
    UpdatedAt    time.Time
}
```

### Repository Pattern

```go
// internal/database/repository/asset_repo.go

type AssetRepository struct {
    db *bun.DB
}

func (r *AssetRepository) Create(ctx context.Context, asset *database.Asset) error
func (r *AssetRepository) Search(ctx context.Context, query AssetQuery) ([]*database.Asset, int, error)
func (r *AssetRepository) Upsert(ctx context.Context, asset *database.Asset) error
```

### Schedule Operations

```go
// internal/database/seed.go

func ListSchedules(ctx context.Context, offset, limit int) (*ScheduleResult, error)
func GetScheduleByID(ctx context.Context, id string) (*Schedule, error)
func CreateSchedule(ctx context.Context, input CreateScheduleInput) (*Schedule, error)
func UpdateSchedule(ctx context.Context, id string, input UpdateScheduleInput) (*Schedule, error)
func DeleteSchedule(ctx context.Context, id string) error
func UpdateScheduleLastRun(ctx context.Context, id string) error
```

### JSONL Import

```go
// internal/database/jsonl.go

type JSONLImporter struct {
    db        *bun.DB
    batchSize int
}

func (i *JSONLImporter) ImportAssets(ctx context.Context, filePath, workspace, source string) (*ImportResult, error)
```

### Event Log Management

The `db_reset_event_logs` utility function enables event reprocessing:

```go
// internal/functions/db_functions.go

// db_reset_event_logs(workspace?, topic_pattern?) -> {reset: int, total: int}
// Resets processed event logs back to unprocessed state

// Examples:
db_reset_event_logs()                          // Reset all processed events
db_reset_event_logs("example.com")             // Reset events for workspace
db_reset_event_logs("", "db.*")                // Reset events matching topic pattern (glob)
db_reset_event_logs("example.com", "assets.*") // Both filters
```

Topic patterns use glob syntax (`*` matches any characters, `?` matches single character).

## SARIF Integration

Osmedeus supports importing and analyzing results from SAST (Static Application Security Testing) tools that produce SARIF output.

### Supported Tools

| Tool | Type | SARIF Output |
|------|------|--------------|
| Semgrep | Code analysis | `semgrep --sarif -o results.sarif` |
| Trivy | Container/FS scanning | `trivy fs --format sarif -o results.sarif` |
| Kingfisher | Dependency checks | Native SARIF output |
| Bearer | API key detection | Native SARIF output |

### Import Pipeline

```
SARIF File → Parse runs/results/rules → Map severity → Upsert into database
                                                          ↓
                                               {new, updated, unchanged, errors, total}
```

The import function:
1. Parses the SARIF JSON structure (runs → results → rules/locations)
2. Maps SARIF severity levels to osmedeus severity (error→high, warning→medium, note→low)
3. Upserts findings into the database with deduplication
4. Marks assets with `asset_type='repo'` for code-level analysis
5. Returns stats: `{new, updated, unchanged, errors, total}`

### Markdown Reporting

`convert_sarif_to_markdown()` generates severity-sorted tables with:
- Severity counts summary
- Location (file:line)
- Rule ID, title, and description

## Testing

### Test Structure

```
internal/functions/registry_test.go      # Function unit tests
internal/parser/loader_test.go           # Parser/loader unit tests
internal/runner/runner_test.go           # Runner unit tests
internal/executor/executor_test.go       # Executor unit tests
internal/scheduler/scheduler_test.go     # Scheduler unit tests
pkg/server/handlers/handlers_test.go     # API handler unit tests
test/integration/workflow_test.go        # Workflow integration tests
test/e2e/                                # E2E CLI tests
├── e2e_test.go                          # Common test helpers
├── version_test.go                      # Version command tests
├── health_test.go                       # Health command tests
├── workflow_test.go                     # Workflow command tests
├── function_test.go                     # Function command tests
├── scan_test.go                         # Scan command tests
├── server_test.go                       # Server command tests
├── worker_test.go                       # Worker command tests
├── distributed_test.go                  # Distributed scan e2e tests
├── ssh_test.go                          # SSH runner e2e tests (module & step level)
├── api_test.go                          # API endpoint e2e tests (all routes)
├── agent_test.go                        # Agent step e2e tests
└── canary_test.go                       # Canary tests (real-world scans in Docker)
```

### Running Tests

```bash
# All unit tests (fast, no external dependencies)
make test-unit

# Integration tests (requires Docker)
make test-integration

# E2E CLI tests (requires binary build)
make test-e2e

# SSH E2E tests - full workflow tests with SSH runner
# Tests both module-level (runner: ssh) and step-level (step_runner: ssh)
# Uses linuxserver/openssh-server Docker container
make test-e2e-ssh

# API E2E tests - tests all API endpoints
# Starts Redis, seeds database, starts server, tests all routes
make test-e2e-api

# Distributed scan e2e tests (requires Docker for Redis)
make test-distributed

# Docker runner tests
make test-docker

# SSH runner unit tests (using linuxserver/openssh-server)
make test-ssh

# Canary tests (real-world scans in Docker, 20-60 min each)
make test-canary-all      # All canary scenarios (60-90min)
make test-canary-repo     # SAST on juice-shop (~25min)
make test-canary-domain   # Domain recon on hackerone.com (~20min)
make test-canary-ip       # CIDR scanning (~25min)
make test-canary-general  # Domain-list-recon on hackerone.com subdomains (~40min)
make canary-up            # Build & start canary container (shared setup)
make canary-down          # Teardown canary container

# All tests with coverage
make test-coverage
```

### Writing Tests

Use testify for assertions:

```go
func TestMyFeature(t *testing.T) {
    // Arrange
    tmpDir := t.TempDir()

    // Act
    result, err := myFunction(tmpDir)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

For integration tests, use build tags:

```go
func TestDockerRunner_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // ...
}
```

## Canary Testing

Canary tests are real-world integration tests that run actual security scans inside a Docker container. They verify the full pipeline from workflow execution through database persistence and API reporting.

### Architecture

```
1. Build canary Docker image (multi-stage: Go 1.25 builder → Ubuntu 24.04 runtime)
2. Compile osmedeus from current source (not released binaries)
3. Layer onto toolbox image with pre-installed SAST tools (Trivy, Semgrep, Kingfisher)
4. Start API server in background on :8002
5. Run scan workflows against real targets
6. Verify: filesystem artifacts + API responses + database records
```

### Test Scenarios

| Test | Target | Duration | What It Tests |
|------|--------|----------|---------------|
| `TestCanary_Repo` | juice-shop | ~25min | SAST scanning, SARIF import, vulnerability DB |
| `TestCanary_Domain` | hackerone.com | ~20min | DNS enumeration, subdomain discovery |
| `TestCanary_CIDR` | Public IPs | ~25min | Network range scanning |
| `TestCanary_General` | hackerone.com subdomains | ~40min | Domain-list-recon (probing, fingerprinting, scanning) |
| `TestCanary_FullSuite` | All above | ~90min | Complete lifecycle with container management |

### Running Canary Tests

```bash
# Full suite (builds container → runs all 3 → cleans up)
make test-canary-all

# Individual scenarios (each handles container lifecycle)
make test-canary-repo
make test-canary-domain
make test-canary-ip

# Manual container management for development
make canary-up            # Build & start container
make canary-down          # Stop & cleanup
```

### Verification Layers

Canary tests assert across three layers:
1. **Filesystem**: SARIF files, markdown reports, text outputs exist
2. **API**: Runs, assets, vulnerabilities accessible via REST endpoints
3. **Database**: Workspace records, total counts, vulnerability severity breakdown

## Preset Installation

Osmedeus supports installing base folders and workflows from curated preset repositories for reproducible deployments.

### CLI Commands

```bash
# Install base folder from preset repository
osmedeus install base --preset

# Install workflows from preset repository
osmedeus install workflow --preset

# Validate and install ready-to-use base
osmedeus install validate --preset
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OSM_PRESET_URL` | Default base repo | Override base preset source |
| `OSM_WORKFLOW_URL` | Default workflow repo | Override workflow preset source |

Preset installation is useful for Docker images and CI/CD pipelines where reproducible, stable deployments from tested configurations are preferred over manual source specification.

## Adding New Features

### Adding a New Step Type

1. Define the type in `internal/core/types.go`:

```go
const StepTypeMyNew StepType = "mynew"
```

2. Create executor in `internal/executor/mynew_executor.go`:

```go
type MyNewExecutor struct {
    templateEngine *template.Engine
}

func (e *MyNewExecutor) Execute(ctx context.Context, step *core.Step,
                                  execCtx *core.ExecutionContext) (*core.StepResult, error) {
    // Implementation
}
```

3. Register in dispatcher (`internal/executor/dispatcher.go`):

```go
func (d *StepDispatcher) Dispatch(...) (*core.StepResult, error) {
    switch step.Type {
    // ...
    case core.StepTypeMyNew:
        return d.myNewExecutor.Execute(ctx, step, execCtx)
    }
}
```

### Adding a New Runner

1. Create runner in `internal/runner/myrunner.go`:

```go
type MyRunner struct {
    config *core.RunnerConfig
}

func (r *MyRunner) Execute(ctx context.Context, command string) (*CommandResult, error)
func (r *MyRunner) Setup(ctx context.Context) error
func (r *MyRunner) Cleanup(ctx context.Context) error
func (r *MyRunner) Type() core.RunnerType
func (r *MyRunner) IsRemote() bool
```

2. Add type in `internal/core/types.go`:

```go
const RunnerTypeMy RunnerType = "myrunner"
```

3. Register in factory (`internal/runner/runner.go`):

```go
func NewRunnerFromType(runnerType core.RunnerType, ...) (Runner, error) {
    switch runnerType {
    case core.RunnerTypeMy:
        return NewMyRunner(config, binaryPath)
    }
}
```

### Adding a New Installer Mode

1. Create installer in `internal/installer/mymode.go`:

```go
func InstallBinaryViaMyMode(name, pkg, binariesFolder string) error {
    // Implementation
}
```

2. Add flag in `pkg/cli/install.go`:

```go
installBinaryCmd.Flags().BoolVar(&myModeInstall, "my-mode-install", false, "use MyMode to install")
```

3. Register in `runInstallBinary()` switch statement.

See `internal/installer/nix.go` for a complete example.

### Adding a New API Endpoint

1. Add handler in `pkg/server/handlers/handlers.go`:

```go
func MyHandler(cfg *config.Config) fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Implementation
        return c.JSON(fiber.Map{"data": result})
    }
}
```

2. Register route in `pkg/server/server.go`:

```go
func (s *Server) setupRoutes() {
    // ...
    api.Get("/my-endpoint", handlers.MyHandler(s.config))
}
```

### Adding a New CLI Command

1. Create command file in `pkg/cli/mycommand.go`:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Implementation
    },
}

func init() {
    myCmd.Flags().StringVarP(&myFlag, "flag", "f", "", "description")
}
```

2. Register in `pkg/cli/root.go`:

```go
func init() {
    rootCmd.AddCommand(myCmd)
}
```

## CLI Shortcuts and Tips

### Command Aliases

- `osmedeus func` - alias for `osmedeus function`
- `osmedeus func e` - alias for `osmedeus function eval`
- `osmedeus db ls` - alias for `osmedeus db list`

### Database CLI Commands

Query and manage database tables directly from the CLI:

```bash
# List all tables with row counts
osmedeus db list

# Query specific table (default columns shown)
osmedeus db list --table event_logs

# List available columns for a table
osmedeus db list --table event_logs --list-columns

# Filter by specific columns
osmedeus db list --table event_logs --columns topic,source,data_type,data

# Show all columns including hidden ones (id, timestamps)
osmedeus db list --table event_logs --all

# Filter by field value
osmedeus db list --table event_logs --where topic=assets.new
osmedeus db list --table event_logs --where processed=false

# Search across all columns
osmedeus db list --table event_logs --search "nuclei"

# Output as JSON for scripting
osmedeus db list --table event_logs --json

# Pagination
osmedeus db list --table event_logs --offset 50 --limit 100
```

Default columns per table:
- `runs`: run_id, job_id, workflow_name, target, status, started_at
- `event_logs`: topic, source, processed, data_type, workspace, data
- `assets`: asset_value, host_ip, title, status_code, last_seen_at, technologies
- `schedules`: name, workflow_name, trigger_type, schedule, is_enabled, run_count

### Function Evaluation CLI

Evaluate utility functions from the command line with bulk processing support:

```bash
# Single expression evaluation
osmedeus func eval 'log_info("hello")'
osmedeus func eval -e 'fileLength("/path/to/file.txt")'

# With target variable
osmedeus func eval -e 'httpGet("https://" + target)' -t example.com

# Bulk processing from file (target variable available in script)
osmedeus func eval -e 'log_info("Processing: " + target)' -T targets.txt

# Bulk processing with concurrency
osmedeus func eval -e 'httpGet("https://" + target)' -T targets.txt -c 10

# Using function files for reusable logic
osmedeus func eval --function-file check-host.js -T targets.txt -c 5

# Additional parameters
osmedeus func eval -e 'log_info(target + " in " + ws)' -T targets.txt --params ws=production

# Function name with arguments
osmedeus func eval log_info "hello world"
osmedeus func eval -f httpGet "https://example.com"

# Read script from stdin
echo 'log_info("hello")' | osmedeus func eval --stdin

# List available functions
osmedeus func list
osmedeus func list event  # Filter by category
```

### New Scan Flags

- `-c, --concurrency` - Number of targets to scan concurrently
- `--timeout` - Scan timeout (e.g., `2h`, `3h`, `1d`)
- `--repeat` - Repeat scan after completion
- `--repeat-wait-time` - Wait time between repeats (e.g., `30m`, `1h`, `1d`)
- `-m` can be specified multiple times to run modules in sequence

### Debugging Tips

- Use `osmedeus --usage-example` to see comprehensive examples for all commands
- Use `--verbose` or `--debug` for detailed logging
- Use `--dry-run` to preview scan execution without running commands
- Use `--log-file-tmp` to create timestamped log files for debugging

## Code Style

- Use `go fmt` and `golangci-lint`
- Follow Go naming conventions
- Use structured logging with zap
- Return errors, don't panic
- Use context for cancellation
- Write tests for new features

## Useful Commands

```bash
# Build
make build

# Test
make test-unit

# Format
make fmt

# Lint
make lint

# Tidy dependencies
make tidy

# Generate (if needed)
make generate

# Generate Swagger docs
make swagger

# Update embedded UI from dashboard build
make update-ui

# Install to $GOBIN
make install

# Docker Toolbox (all tools pre-installed)
make docker-toolbox          # Build toolbox image
make docker-toolbox-run      # Start toolbox container
make docker-toolbox-shell    # Enter container shell

# Canary tests (real-world scans in Docker)
make test-canary-all         # All scenarios (30-60min)
make canary-up               # Start canary container
make canary-down             # Cleanup canary container
```
