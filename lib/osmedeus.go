package lib

import (
	"context"
	"path/filepath"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/template"
)

// Run executes a workflow from YAML content against a target.
// This is the main entry point for programmatic workflow execution.
//
// Example:
//
//	result, err := lib.Run("example.com", workflowYAML, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Status: %s\n", result.Status)
func Run(target string, workflowYAML string, opts *RunOptions) (*RunResult, error) {
	return RunWithContext(context.Background(), target, workflowYAML, opts)
}

// RunWithContext executes a workflow with context support for cancellation/timeout.
//
// Example with timeout:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	result, err := lib.RunWithContext(ctx, "example.com", workflowYAML, nil)
func RunWithContext(ctx context.Context, target string, workflowYAML string, opts *RunOptions) (*RunResult, error) {
	// Validate inputs
	if target == "" {
		return nil, ErrEmptyTarget
	}
	if workflowYAML == "" {
		return nil, ErrEmptyWorkflow
	}

	// Use default options if nil
	if opts == nil {
		opts = DefaultRunOptions()
	}

	// Parse workflow
	workflow, err := parser.ParseContent([]byte(workflowYAML))
	if err != nil {
		return nil, NewParseError("failed to parse workflow YAML", err)
	}

	// Validate workflow
	if err := parser.Validate(workflow); err != nil {
		return nil, &ValidationError{
			Message: err.Error(),
			Err:     err,
		}
	}

	// Check that it's a module (flows not supported in library mode)
	if !workflow.IsModule() {
		return nil, ErrNotModule
	}

	// Build configuration
	cfg := opts.Config
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Override paths if specified
	if opts.BaseFolder != "" {
		cfg.BaseFolder = opts.BaseFolder
	}
	if opts.WorkspacesPath != "" {
		cfg.Environments.Workspaces = opts.WorkspacesPath
	}

	// Resolve all paths (expands templates and env vars)
	cfg.ResolvePaths()

	// Build params map
	params := make(map[string]string)
	params["target"] = target
	if opts.Tactic != "" {
		params["tactic"] = opts.Tactic
	} else {
		params["tactic"] = "default"
	}

	// Merge user-provided params
	for k, v := range opts.Params {
		params[k] = v
	}

	// Create executor
	exec := executor.NewExecutor()
	exec.SetDryRun(opts.DryRun)
	exec.SetVerbose(opts.Verbose)
	exec.SetSilent(opts.Silent)
	exec.SetDisableWorkflowState(opts.DisableWorkflowState)

	// Execute workflow
	result, err := exec.ExecuteModule(ctx, workflow, params, cfg)

	// Build output path
	outputPath := ""
	if result != nil {
		// Use target as workspace name (sanitized internally by executor)
		outputPath = filepath.Join(cfg.WorkspacesPath, target)
	}

	// Convert to lib result
	runResult := fromWorkflowResult(result, outputPath)

	// If execution failed with an error, wrap it
	if err != nil && runResult != nil {
		runResult.Error = err
	}

	return runResult, err
}

// RunModule is a convenience wrapper that executes a module workflow with default options.
//
// Example:
//
//	result, err := lib.RunModule("example.com", workflowYAML)
func RunModule(target, workflowYAML string) (*RunResult, error) {
	return Run(target, workflowYAML, nil)
}

// RunModuleWithParams is a convenience wrapper that executes a module workflow with custom params.
//
// Example:
//
//	result, err := lib.RunModuleWithParams("example.com", workflowYAML, map[string]string{
//	    "threads": "20",
//	    "timeout": "30",
//	})
func RunModuleWithParams(target, workflowYAML string, params map[string]string) (*RunResult, error) {
	opts := DefaultRunOptions()
	opts.Params = params
	return Run(target, workflowYAML, opts)
}

// Eval evaluates a JavaScript expression with optional context.
// Uses the Otto JavaScript runtime with built-in utility functions.
//
// Example:
//
//	// Check if a file exists
//	exists, err := lib.Eval(`fileExists("/etc/passwd")`, nil)
//
//	// Use context variables
//	result, err := lib.Eval(`trim(input)`, &lib.EvalOptions{
//	    Context: map[string]interface{}{"input": "  hello  "},
//	})
func Eval(expression string, opts *EvalOptions) (interface{}, error) {
	if expression == "" {
		return nil, ErrEmptyExpression
	}

	// Build context
	ctx := make(map[string]interface{})
	if opts != nil {
		// Copy user context
		for k, v := range opts.Context {
			ctx[k] = v
		}
		// Override target if specified
		if opts.Target != "" {
			ctx["target"] = opts.Target
		}
	}

	// Create template engine for variable rendering
	engine := template.NewEngine()

	// Render template variables in the expression
	rendered, err := engine.Render(expression, ctx)
	if err != nil {
		// If template rendering fails, try with original expression
		rendered = expression
	}

	// Create function registry and execute
	registry := functions.NewRegistry()
	return registry.Execute(rendered, ctx)
}

// EvalCondition evaluates a boolean condition expression.
// Returns true/false based on the expression result.
//
// Example:
//
//	ok, err := lib.EvalCondition(`len(items) > 0`, &lib.EvalOptions{
//	    Context: map[string]interface{}{"items": []string{"a", "b"}},
//	})
func EvalCondition(condition string, opts *EvalOptions) (bool, error) {
	if condition == "" {
		return false, ErrEmptyExpression
	}

	// Build context
	ctx := make(map[string]interface{})
	if opts != nil {
		for k, v := range opts.Context {
			ctx[k] = v
		}
		if opts.Target != "" {
			ctx["target"] = opts.Target
		}
	}

	// Create template engine for variable rendering
	engine := template.NewEngine()

	// Render template variables in the condition
	rendered, err := engine.Render(condition, ctx)
	if err != nil {
		rendered = condition
	}

	// Create function registry and evaluate condition
	registry := functions.NewRegistry()
	return registry.EvaluateCondition(rendered, ctx)
}

// EvalFunction is a convenience wrapper for Eval without options.
//
// Example:
//
//	result, err := lib.EvalFunction(`uuid()`)
func EvalFunction(expression string) (interface{}, error) {
	return Eval(expression, nil)
}

// EvalFunctionWithContext is a convenience wrapper for Eval with context.
//
// Example:
//
//	result, err := lib.EvalFunctionWithContext(`split(text, ",")`, map[string]interface{}{
//	    "text": "a,b,c",
//	})
func EvalFunctionWithContext(expression string, ctx map[string]interface{}) (interface{}, error) {
	return Eval(expression, &EvalOptions{Context: ctx})
}

// ParseWorkflow parses a workflow YAML string and returns the parsed workflow.
// Useful for validation or inspection before execution.
//
// Example:
//
//	workflow, err := lib.ParseWorkflow(workflowYAML)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Workflow: %s (%s)\n", workflow.Name, workflow.Kind)
func ParseWorkflow(workflowYAML string) (*core.Workflow, error) {
	if workflowYAML == "" {
		return nil, ErrEmptyWorkflow
	}

	workflow, err := parser.ParseContent([]byte(workflowYAML))
	if err != nil {
		return nil, NewParseError("failed to parse workflow YAML", err)
	}

	return workflow, nil
}

// ValidateWorkflow parses and validates a workflow YAML string.
// Returns nil if the workflow is valid.
//
// Example:
//
//	if err := lib.ValidateWorkflow(workflowYAML); err != nil {
//	    fmt.Printf("Invalid workflow: %v\n", err)
//	}
func ValidateWorkflow(workflowYAML string) error {
	workflow, err := ParseWorkflow(workflowYAML)
	if err != nil {
		return err
	}

	if err := parser.Validate(workflow); err != nil {
		return &ValidationError{
			Message: err.Error(),
			Err:     err,
		}
	}

	return nil
}
