package core

import "context"

// WorkflowParser parses workflow files
type WorkflowParser interface {
	// Parse parses a workflow file and returns the workflow
	Parse(path string) (*Workflow, error)

	// ParseContent parses workflow content from bytes
	ParseContent(content []byte) (*Workflow, error)

	// Validate validates a parsed workflow
	Validate(w *Workflow) error
}

// TemplateEngine renders templates with variable substitution
type TemplateEngine interface {
	// Render renders a template string with the given context
	Render(template string, ctx map[string]interface{}) (string, error)

	// RenderStep renders all template fields in a step
	RenderStep(step *Step, ctx map[string]interface{}) (*Step, error)

	// ExecuteGenerator executes a generator function and returns the result
	ExecuteGenerator(expr string) (string, error)
}

// FunctionRegistry manages and executes utility functions
type FunctionRegistry interface {
	// Register registers a function with the given name
	Register(name string, fn interface{}) error

	// Execute executes a function expression and returns the result
	Execute(expr string, ctx map[string]interface{}) (interface{}, error)

	// EvaluateCondition evaluates a condition expression and returns true/false
	EvaluateCondition(condition string, ctx map[string]interface{}) (bool, error)

	// EvaluateExports evaluates export expressions and returns the results
	EvaluateExports(exports map[string]string, ctx map[string]interface{}) (map[string]interface{}, error)
}

// StepExecutor executes individual steps
type StepExecutor interface {
	// Execute executes a step and returns the result
	Execute(ctx context.Context, step *Step, execCtx *ExecutionContext) (*StepResult, error)

	// CanHandle returns true if this executor can handle the given step type
	CanHandle(stepType StepType) bool
}

// WorkflowExecutor executes complete workflows
type WorkflowExecutor interface {
	// ExecuteModule executes a module workflow
	ExecuteModule(ctx context.Context, module *Workflow, params map[string]string) (*WorkflowResult, error)

	// ExecuteFlow executes a flow workflow
	ExecuteFlow(ctx context.Context, flow *Workflow, params map[string]string) (*WorkflowResult, error)
}

// Scheduler manages workflow triggers and scheduling
type Scheduler interface {
	// RegisterTrigger registers a workflow trigger
	RegisterTrigger(workflow *Workflow, trigger *Trigger) error

	// UnregisterTrigger removes a trigger by name
	UnregisterTrigger(name string) error

	// Start starts the scheduler
	Start() error

	// Stop stops the scheduler
	Stop() error

	// EmitEvent emits a named event with payload
	EmitEvent(name string, payload map[string]interface{}) error
}

// WorkflowLoader loads workflows from disk
type WorkflowLoader interface {
	// LoadWorkflow loads a single workflow by name
	LoadWorkflow(name string) (*Workflow, error)

	// LoadAllWorkflows loads all workflows from the configured directory
	LoadAllWorkflows() ([]*Workflow, error)

	// ReloadWorkflows reloads all workflows from disk
	ReloadWorkflows() error

	// GetWorkflow returns a cached workflow by name
	GetWorkflow(name string) (*Workflow, bool)
}

// DependencyChecker validates workflow dependencies
type DependencyChecker interface {
	// CheckCommands checks if required commands are available
	CheckCommands(commands []string) error

	// CheckFiles checks if required files exist
	CheckFiles(files []string, ctx map[string]interface{}) error

	// CheckVariables validates required variables are present
	CheckVariables(deps []VariableDep, ctx map[string]interface{}) error

	// CheckAll performs all dependency checks
	CheckAll(deps *Dependencies, ctx map[string]interface{}) error
}
