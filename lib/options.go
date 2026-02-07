package lib

import "github.com/j3ssie/osmedeus/v5/internal/config"

// RunOptions configures workflow execution
type RunOptions struct {
	// Params are key-value parameters passed to the workflow
	// These are merged with workflow default params
	Params map[string]string

	// Tactic sets the scan aggressiveness: "aggressive", "default", or "gently"
	// Controls thread counts and parallelism levels
	Tactic string

	// DryRun shows commands without executing them
	DryRun bool

	// Verbose enables detailed output (shows step stdout)
	Verbose bool

	// Silent suppresses step output (default: true for library mode)
	// When true, only errors are reported
	Silent bool

	// Config provides a custom configuration
	// If nil, DefaultConfig() is used
	Config *config.Config

	// BaseFolder overrides the base folder path
	// Default: ~/osmedeus-base
	BaseFolder string

	// WorkspacesPath overrides the output directory for workspaces
	// Default: ~/workspaces-osmedeus (or from config)
	WorkspacesPath string

	// DisableWorkflowState disables writing workflow state files to output directory
	// Useful for ephemeral/in-memory execution
	DisableWorkflowState bool

	// SkipWorkspace skips creating workspace/output directory
	// Useful for empty-target mode where no real target exists
	SkipWorkspace bool

	// DisableDatabase skips all database operations
	// Default: true for library mode
	DisableDatabase bool
}

// DefaultRunOptions returns default options suitable for library usage
func DefaultRunOptions() *RunOptions {
	return &RunOptions{
		Params:               make(map[string]string),
		Tactic:               "default",
		DryRun:               false,
		Verbose:              false,
		Silent:               true, // Library mode defaults to silent
		DisableWorkflowState: false,
		DisableDatabase:      true, // Library mode skips DB by default
	}
}

// WithParams returns a copy of options with the given params
func (o *RunOptions) WithParams(params map[string]string) *RunOptions {
	copy := *o
	copy.Params = params
	return &copy
}

// WithTactic returns a copy of options with the given tactic
func (o *RunOptions) WithTactic(tactic string) *RunOptions {
	copy := *o
	copy.Tactic = tactic
	return &copy
}

// WithDryRun returns a copy of options with dry-run enabled
func (o *RunOptions) WithDryRun(dryRun bool) *RunOptions {
	copy := *o
	copy.DryRun = dryRun
	return &copy
}

// WithVerbose returns a copy of options with verbose enabled
func (o *RunOptions) WithVerbose(verbose bool) *RunOptions {
	copy := *o
	copy.Verbose = verbose
	return &copy
}

// WithSilent returns a copy of options with silent mode
func (o *RunOptions) WithSilent(silent bool) *RunOptions {
	copy := *o
	copy.Silent = silent
	return &copy
}

// WithConfig returns a copy of options with the given config
func (o *RunOptions) WithConfig(cfg *config.Config) *RunOptions {
	copy := *o
	copy.Config = cfg
	return &copy
}

// WithBaseFolder returns a copy of options with the given base folder
func (o *RunOptions) WithBaseFolder(folder string) *RunOptions {
	copy := *o
	copy.BaseFolder = folder
	return &copy
}

// WithWorkspacesPath returns a copy of options with the given workspaces path
func (o *RunOptions) WithWorkspacesPath(path string) *RunOptions {
	copy := *o
	copy.WorkspacesPath = path
	return &copy
}

// EvalOptions configures function evaluation
type EvalOptions struct {
	// Context provides variables for the expression
	// Variables are accessible by name in JavaScript expressions
	Context map[string]interface{}

	// Target is a convenience field that sets ctx["target"]
	// If both Target and Context["target"] are set, Target takes precedence
	Target string
}

// DefaultEvalOptions returns default options for function evaluation
func DefaultEvalOptions() *EvalOptions {
	return &EvalOptions{
		Context: make(map[string]interface{}),
	}
}

// WithContext returns a copy of options with the given context
func (o *EvalOptions) WithContext(ctx map[string]interface{}) *EvalOptions {
	copy := *o
	copy.Context = ctx
	return &copy
}

// WithTarget returns a copy of options with the given target
func (o *EvalOptions) WithTarget(target string) *EvalOptions {
	copy := *o
	copy.Target = target
	return &copy
}
