# Hacking on Osmedeus

This document describes the technical architecture and development practices for Osmedeus. It's intended for developers who want to understand, modify, or extend the codebase.

## Table of Contents

- [Project Structure](#project-structure)
- [Architecture Overview](#architecture-overview)
- [Core Components](#core-components)
- [Workflow Engine](#workflow-engine)
- [Execution Pipeline](#execution-pipeline)
- [Runner System](#runner-system)
- [Authentication Middleware](#authentication-middleware)
- [Template Engine](#template-engine)
- [Function Registry](#function-registry)
- [Scheduler System](#scheduler-system)
- [Database Layer](#database-layer)
- [Testing](#testing)
- [Adding New Features](#adding-new-features)

## Project Structure

```
osmedeus-ng/
├── cmd/
│   └── osmedeus/
│       └── main.go              # Application entry point
├── internal/                    # Private packages
│   ├── config/                  # Configuration management
│   │   ├── config.go            # Config loading and defaults
│   │   ├── settings.go          # Settings struct definitions
│   │   └── secrets.go           # Secret management
│   ├── core/                    # Core types and interfaces
│   │   ├── workflow.go          # Workflow and RunnerConfig structs
│   │   ├── step.go              # Step definition
│   │   ├── trigger.go           # Trigger types and events
│   │   ├── types.go             # Status types, constants
│   │   ├── context.go           # Execution context
│   │   └── constants.go         # Project metadata
│   ├── parser/                  # Workflow parsing
│   │   ├── parser.go            # YAML parser
│   │   ├── loader.go            # Workflow loader with caching
│   │   └── validator.go         # Workflow validation
│   ├── executor/                # Workflow execution
│   │   ├── executor.go          # Main executor
│   │   ├── dispatcher.go        # Step dispatcher
│   │   ├── bash_executor.go     # Bash step handler
│   │   ├── function_executor.go # Function step handler
│   │   ├── foreach_executor.go  # Foreach step handler
│   │   ├── parallel_executor.go # Parallel step handler
│   │   ├── remote_bash_executor.go # Remote-bash step handler
│   │   ├── http_executor.go     # HTTP request handler
│   │   └── llm_executor.go      # LLM step handler
│   ├── runner/                  # Execution environments
│   │   ├── runner.go            # Runner interface
│   │   ├── host_runner.go       # Local execution
│   │   ├── docker_runner.go     # Docker execution
│   │   └── ssh_runner.go        # SSH remote execution
│   ├── template/                # Template rendering
│   │   ├── engine.go            # Template engine
│   │   ├── generators.go        # Value generators
│   │   └── context.go           # Template context
│   ├── functions/               # Utility functions
│   │   ├── registry.go          # Function registry
│   │   ├── otto_runtime.go      # JavaScript runtime
│   │   ├── file_functions.go    # File operations
│   │   ├── string_functions.go  # String operations
│   │   ├── util_functions.go    # Utility functions
│   │   └── jq.go                # JSON query functions
│   ├── scheduler/               # Trigger scheduling
│   │   └── scheduler.go         # Cron, event, watch triggers
│   ├── database/                # Data persistence
│   │   ├── database.go          # Connection management
│   │   ├── models.go            # Data models
│   │   ├── jsonl.go             # JSONL import/export
│   │   └── repository/          # Data access layer
│   ├── heuristics/              # Target analysis
│   │   ├── heuristics.go        # Target type detection
│   │   ├── url.go               # URL parsing
│   │   └── domain.go            # Domain analysis
│   ├── workspace/               # Workspace management
│   │   └── workspace.go         # Workspace creation
│   ├── snapshot/                # Workspace snapshots
│   │   └── snapshot.go          # Export/import ZIP archives
│   ├── installer/               # Binary installation
│   │   ├── installer.go         # Core installer logic
│   │   ├── nix.go               # Nix package manager support
│   │   └── registry.go          # Binary registry management
│   ├── state/                   # Run state export
│   │   ├── export.go            # State export functionality
│   │   └── types.go             # State types
│   ├── updater/                 # Self-update functionality
│   │   ├── updater.go           # Update logic
│   │   └── github_source.go     # GitHub releases source
│   ├── terminal/                # Terminal UI
│   │   ├── printer.go           # Output formatting
│   │   ├── colors.go            # ANSI colors
│   │   ├── symbols.go           # Unicode symbols
│   │   ├── spinner.go           # Loading spinners
│   │   └── table.go             # Table rendering
│   └── logger/                  # Structured logging
│       └── logger.go            # Zap logger wrapper
├── pkg/                         # Public packages
│   ├── cli/                     # CLI commands
│   │   ├── root.go              # Root command
│   │   ├── scan.go              # Scan command
│   │   ├── workflow.go          # Workflow command
│   │   ├── function.go          # Function command
│   │   └── server.go            # Server command
│   └── server/                  # REST API
│       ├── server.go            # Server setup
│       ├── handlers/            # Request handlers
│       └── middleware/          # Auth middleware (JWT, API Key)
├── test/                        # Test suites
│   ├── e2e/                     # E2E CLI tests
│   ├── integration/             # Integration tests
│   └── testdata/
│       ├── workflows/           # Test workflow fixtures
│       └── complex-workflows/   # Complex workflow examples
└── build/                       # Build artifacts
    ├── docker/                  # Docker files
    └── DEPLOYMENT.md            # Deployment guide
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
│  │             │ │              │ │   foreach, parallel-steps│ │
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
    Type             StepType      // bash, function, foreach, parallel-steps, remote-bash, http, llm
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
             ┌──────────────┐            ┌──────────────┐
             │ HTTPExecutor │            │ LLMExecutor  │
             └──────────────┘            └──────────────┘
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
    execCtx.SetVariable("TaskID", execCtx.RunID)
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

### Otto JavaScript Runtime

Functions are implemented in Go and exposed to an Otto JavaScript VM:

```go
// internal/functions/otto_runtime.go

type OttoRuntime struct {
    vm *otto.Otto
}

func NewOttoRuntime() *OttoRuntime {
    vm := otto.New()
    runtime := &OttoRuntime{vm: vm}
    runtime.registerFunctions()
    return runtime
}

func (r *OttoRuntime) registerFunctions() {
    r.vm.Set("fileExists", r.fileExists)
    r.vm.Set("fileLength", r.fileLength)
    r.vm.Set("trim", r.trim)
    // ... register all functions
}
```

### Adding New Functions

1. Add the Go implementation in the appropriate file:

```go
// internal/functions/file_functions.go

func (r *OttoRuntime) myNewFunction(call otto.FunctionCall) otto.Value {
    arg := call.Argument(0).String()
    // ... implementation
    result, _ := r.vm.ToValue(output)
    return result
}
```

2. Register in `registerFunctions()`:

```go
r.vm.Set("myNewFunction", r.myNewFunction)
```

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

### Event Filtering

Events are matched using JavaScript expressions:

```go
func (s *Scheduler) evaluateFilters(filters []string, event *core.Event) bool {
    vm := otto.New()
    vm.Set("event", eventObj)

    for _, filter := range filters {
        result, _ := vm.Run(filter)
        if !result.ToBoolean() {
            return false
        }
    }
    return true
}
```

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
    WorkspacePath  string
    StartedAt      time.Time
    CompletedAt    time.Time
    ErrorMessage   string
    ScheduleID     string
    TriggerType    string    // "manual", "cron", "event", "api"
    TriggerName    string
    TotalSteps     int
    CompletedSteps int
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
    InputConfig  map[string]string  // JSON params
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
└── api_test.go                          # API endpoint e2e tests (all routes)
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
```
