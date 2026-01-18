# Osmedeus Library SDK

Use Osmedeus as a Go library to programmatically execute workflows and evaluate utility functions.

## Installation

```bash
go get github.com/j3ssie/osmedeus/v5/lib
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/j3ssie/osmedeus/v5/lib"
)

func main() {
    // Define a simple workflow
    workflowYAML := `
name: simple-scan
kind: module
steps:
  - name: echo-target
    type: bash
    command: echo "Scanning {{target}}"
`

    // Run the workflow
    result, err := lib.Run("example.com", workflowYAML, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Status: %s\n", result.Status)
    fmt.Printf("Duration: %v\n", result.Duration)
}
```

## API Reference

### Workflow Execution

#### `Run(target, workflowYAML, opts) (*RunResult, error)`

Execute a module workflow against a target.

```go
result, err := lib.Run("example.com", workflowYAML, nil)
```

#### `RunWithContext(ctx, target, workflowYAML, opts) (*RunResult, error)`

Execute with context support for cancellation and timeout.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := lib.RunWithContext(ctx, "example.com", workflowYAML, nil)
```

#### `RunModule(target, workflowYAML) (*RunResult, error)`

Convenience wrapper with default options.

```go
result, err := lib.RunModule("example.com", workflowYAML)
```

#### `RunModuleWithParams(target, workflowYAML, params) (*RunResult, error)`

Convenience wrapper with custom parameters.

```go
result, err := lib.RunModuleWithParams("example.com", workflowYAML, map[string]string{
    "threads": "20",
    "timeout": "30",
})
```

### Function Evaluation

#### `Eval(expression, opts) (interface{}, error)`

Evaluate a JavaScript expression with optional context.

```go
// Simple expression
result, err := lib.Eval(`1 + 1`, nil)

// With context variables
result, err := lib.Eval(`trim(input)`, &lib.EvalOptions{
    Context: map[string]interface{}{"input": "  hello  "},
})
```

#### `EvalCondition(condition, opts) (bool, error)`

Evaluate a boolean condition.

```go
ok, err := lib.EvalCondition(`len(items) > 0`, &lib.EvalOptions{
    Context: map[string]interface{}{"items": []string{"a", "b"}},
})
```

#### `EvalFunction(expression) (interface{}, error)`

Convenience wrapper for Eval without options.

```go
result, err := lib.EvalFunction(`uuid()`)
```

#### `EvalFunctionWithContext(expression, ctx) (interface{}, error)`

Convenience wrapper for Eval with context.

```go
result, err := lib.EvalFunctionWithContext(`split(text, ",")`, map[string]interface{}{
    "text": "a,b,c",
})
```

### Workflow Validation

#### `ParseWorkflow(workflowYAML) (*core.Workflow, error)`

Parse a workflow YAML string without executing.

```go
workflow, err := lib.ParseWorkflow(workflowYAML)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Workflow: %s (%s)\n", workflow.Name, workflow.Kind)
```

#### `ValidateWorkflow(workflowYAML) error`

Parse and validate a workflow. Returns nil if valid.

```go
if err := lib.ValidateWorkflow(workflowYAML); err != nil {
    fmt.Printf("Invalid workflow: %v\n", err)
}
```

## Configuration Options

### RunOptions

```go
opts := &lib.RunOptions{
    // Parameters passed to the workflow
    Params: map[string]string{
        "threads": "20",
        "custom":  "value",
    },

    // Scan aggressiveness: "aggressive", "default", or "gently"
    Tactic: "default",

    // Show commands without executing
    DryRun: false,

    // Enable detailed output (shows step stdout)
    Verbose: false,

    // Suppress step output (default: true for library mode)
    Silent: true,

    // Custom configuration (nil uses DefaultConfig)
    Config: nil,

    // Override base folder path
    BaseFolder: "",

    // Override workspaces output directory
    WorkspacesPath: "",

    // Disable writing workflow state files
    DisableWorkflowState: false,

    // Skip database operations (default: true for library mode)
    DisableDatabase: true,
}
```

### EvalOptions

```go
opts := &lib.EvalOptions{
    // Variables accessible in the expression
    Context: map[string]interface{}{
        "input": "value",
        "count": 42,
    },

    // Convenience: sets ctx["target"]
    Target: "example.com",
}
```

## Result Types

### RunResult

```go
type RunResult struct {
    WorkflowName string
    RunID        string
    Target       string
    Status       string              // "completed", "failed", "cancelled", "skipped"
    StartTime    time.Time
    EndTime      time.Time
    Duration     time.Duration
    Steps        []*StepResult
    Exports      map[string]interface{}
    Artifacts    []string
    Error        error
    OutputPath   string
    Message      string
}

// Helper methods
result.IsSuccess()           // true if completed
result.IsFailed()            // true if failed
result.IsCancelled()         // true if cancelled
result.IsSkipped()           // true if skipped
result.GetExport("name")     // get exported variable
result.GetExportString("x")  // get as string
result.GetExportBool("x")    // get as bool
result.SuccessfulSteps()     // count of successful steps
result.FailedSteps()         // count of failed steps
result.SkippedSteps()        // count of skipped steps
```

### StepResult

```go
type StepResult struct {
    Name     string
    Type     string
    Status   string              // "success", "failed", "skipped"
    Output   string
    Duration time.Duration
    Error    error
    Exports  map[string]interface{}
}

// Helper methods
step.IsSuccess()
step.IsFailed()
step.IsSkipped()
step.GetExport("name")
```

## Error Handling

The library provides typed errors for different failure scenarios:

```go
result, err := lib.Run(target, yaml, nil)
if err != nil {
    switch {
    case errors.Is(err, lib.ErrEmptyTarget):
        fmt.Println("Target cannot be empty")
    case errors.Is(err, lib.ErrEmptyWorkflow):
        fmt.Println("Workflow YAML cannot be empty")
    case errors.Is(err, lib.ErrNotModule):
        fmt.Println("Only module workflows are supported")
    case lib.IsParseError(err):
        fmt.Println("YAML parsing failed:", err)
    case lib.IsValidationError(err):
        fmt.Println("Workflow validation failed:", err)
    case lib.IsExecutionError(err):
        fmt.Println("Execution failed:", err)
    default:
        fmt.Println("Error:", err)
    }
}
```

### Error Types

| Error | Description |
|-------|-------------|
| `ErrEmptyTarget` | Target cannot be empty |
| `ErrEmptyWorkflow` | Workflow content cannot be empty |
| `ErrEmptyExpression` | Expression cannot be empty |
| `ErrNotModule` | Workflow must be kind 'module' |
| `ParseError` | YAML parsing failed |
| `ValidationError` | Workflow validation failed |
| `ExecutionError` | Step execution failed |

## Available Functions

The `Eval` functions have access to all Osmedeus utility functions:

### File Operations

| Function | Description |
|----------|-------------|
| `fileExists(path)` | Check if file exists |
| `fileLength(path)` | Get file line count |
| `dirLength(path)` | Get directory entry count |
| `readFile(path)` | Read file contents |
| `readLines(path, n)` | Read first n lines |
| `removeFile(path)` | Delete a file |
| `createFolder(path)` | Create directory |
| `appendFile(path, data)` | Append to file |
| `glob(pattern)` | Find files matching pattern |

### String Operations

| Function | Description |
|----------|-------------|
| `trim(s)` | Remove whitespace |
| `split(s, sep)` | Split string |
| `join(arr, sep)` | Join array |
| `replace(s, old, new)` | Replace substring |
| `contains(s, substr)` | Check substring |
| `startsWith(s, prefix)` | Check prefix |
| `endsWith(s, suffix)` | Check suffix |
| `toLowerCase(s)` | Convert to lowercase |
| `toUpperCase(s)` | Convert to uppercase |
| `match(s, pattern)` | Regex match |
| `regexExtract(s, pattern)` | Extract regex groups |

### Type Conversion

| Function | Description |
|----------|-------------|
| `parseInt(s)` | Parse integer |
| `parseFloat(s)` | Parse float |
| `toString(v)` | Convert to string |
| `toBoolean(v)` | Convert to boolean |

### Utilities

| Function | Description |
|----------|-------------|
| `len(v)` | Get length |
| `isEmpty(v)` | Check if empty |
| `isNotEmpty(v)` | Check if not empty |
| `uuid()` | Generate UUID |
| `randomString(n)` | Generate random string |
| `base64Encode(s)` | Base64 encode |
| `base64Decode(s)` | Base64 decode |

### Logging

| Function | Description |
|----------|-------------|
| `log_info(msg)` | Log info message |
| `log_warn(msg)` | Log warning |
| `log_error(msg)` | Log error |
| `log_debug(msg)` | Log debug |

## Examples

### Run a Reconnaissance Workflow

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/j3ssie/osmedeus/v5/lib"
)

func main() {
    workflow := `
name: recon
kind: module
params:
  - name: threads
    default: "10"
steps:
  - name: subdomain-enum
    type: bash
    command: subfinder -d {{target}} -t {{threads}} -o {{Output}}/subdomains.txt
    exports:
      subdomains_file: "{{Output}}/subdomains.txt"

  - name: check-results
    type: function
    function: |
      log_info("Found " + fileLength("{{subdomains_file}}") + " subdomains")
`

    // Run with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()

    result, err := lib.RunWithContext(ctx, "example.com", workflow, &lib.RunOptions{
        Params: map[string]string{"threads": "20"},
        Tactic: "aggressive",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Status: %s\n", result.Status)
    fmt.Printf("Duration: %v\n", result.Duration)
    fmt.Printf("Output: %s\n", result.OutputPath)

    // Check exports
    if file, ok := result.GetExportString("subdomains_file"); ok {
        fmt.Printf("Subdomains file: %s\n", file)
    }
}
```

### Validate Workflow Before Execution

```go
package main

import (
    "fmt"
    "log"

    "github.com/j3ssie/osmedeus/v5/lib"
)

func main() {
    workflow := `
name: my-workflow
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
`

    // Validate first
    if err := lib.ValidateWorkflow(workflow); err != nil {
        log.Fatalf("Invalid workflow: %v", err)
    }

    // Parse to inspect
    w, _ := lib.ParseWorkflow(workflow)
    fmt.Printf("Workflow: %s\n", w.Name)
    fmt.Printf("Steps: %d\n", len(w.Steps))

    // Then execute
    result, err := lib.Run("target.com", workflow, nil)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Result: %s\n", result.Status)
}
```

### Use Functions for File Processing

```go
package main

import (
    "fmt"
    "log"

    "github.com/j3ssie/osmedeus/v5/lib"
)

func main() {
    // Check if file exists
    exists, _ := lib.Eval(`fileExists("/tmp/results.txt")`, nil)
    fmt.Printf("File exists: %v\n", exists)

    // Read and process file
    if exists.(bool) {
        lineCount, _ := lib.Eval(`fileLength("/tmp/results.txt")`, nil)
        fmt.Printf("Line count: %v\n", lineCount)
    }

    // String processing
    result, _ := lib.Eval(`split(trim(input), ",")`, &lib.EvalOptions{
        Context: map[string]interface{}{
            "input": "  a, b, c  ",
        },
    })
    fmt.Printf("Split result: %v\n", result)

    // Conditional logic
    hasResults, err := lib.EvalCondition(`fileExists(path) && fileLength(path) > 0`, &lib.EvalOptions{
        Context: map[string]interface{}{
            "path": "/tmp/results.txt",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Has results: %v\n", hasResults)
}
```

### Custom Configuration

```go
package main

import (
    "log"

    "github.com/j3ssie/osmedeus/v5/internal/config"
    "github.com/j3ssie/osmedeus/v5/lib"
)

func main() {
    // Load custom config
    cfg, err := config.Load("/path/to/osmedeus-base")
    if err != nil {
        // Fall back to default
        cfg = config.DefaultConfig()
    }

    workflow := `
name: custom-scan
kind: module
steps:
  - name: scan
    type: bash
    command: echo "Using custom config"
`

    result, err := lib.Run("target.com", workflow, &lib.RunOptions{
        Config:         cfg,
        WorkspacesPath: "/custom/output/path",
        Verbose:        true,
        Silent:         false,
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Completed: %s", result.Status)
}
```

