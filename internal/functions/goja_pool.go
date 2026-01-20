package functions

import (
	"regexp"
	"sync"

	"github.com/dop251/goja"
)

// Variable extraction for lazy loading optimization
var (
	// varRefCache caches parsed variable references per expression
	varRefCache sync.Map // expr -> []string

	// Pattern to match variable identifiers (excludes JS keywords)
	varPattern = regexp.MustCompile(`\b([a-zA-Z_][a-zA-Z0-9_]*)\b`)

	// JavaScript keywords and built-in objects to exclude from variable extraction
	jsKeywords = map[string]bool{
		// Keywords
		"true": true, "false": true, "null": true, "undefined": true,
		"if": true, "else": true, "for": true, "while": true, "do": true,
		"switch": true, "case": true, "default": true, "break": true, "continue": true,
		"function": true, "return": true, "var": true, "let": true, "const": true,
		"new": true, "delete": true, "typeof": true, "instanceof": true,
		"this": true, "void": true, "in": true, "of": true,
		"try": true, "catch": true, "finally": true, "throw": true,
		"class": true, "extends": true, "super": true, "import": true, "export": true,
		"async": true, "await": true, "yield": true,
		// Built-in objects
		"Math": true, "String": true, "Number": true, "Boolean": true,
		"Array": true, "Object": true, "JSON": true, "Date": true,
		"RegExp": true, "Error": true, "console": true, "parseInt": true, "parseFloat": true,
		"isNaN": true, "isFinite": true, "encodeURI": true, "decodeURI": true,
		"encodeURIComponent": true, "decodeURIComponent": true,
	}
)

// extractVariables returns variable names referenced in expression.
// Results are cached for repeated expressions (common in loop iterations).
func extractVariables(expr string) []string {
	if cached, ok := varRefCache.Load(expr); ok {
		return cached.([]string)
	}

	matches := varPattern.FindAllStringSubmatch(expr, -1)
	seen := make(map[string]bool)
	var vars []string

	for _, match := range matches {
		name := match[1]
		if !seen[name] && !jsKeywords[name] {
			seen[name] = true
			vars = append(vars, name)
		}
	}

	varRefCache.Store(expr, vars)
	return vars
}

// vmContextRegistry maps Goja VMs to their execution context.
// This allows functions to find their context via the VM reference.
var vmContextRegistry sync.Map // map[*goja.Runtime]*VMContext

// getVMContext retrieves the execution context for a given Goja VM.
// Returns nil if no context is registered (shouldn't happen in normal use).
func getVMContext(vm *goja.Runtime) *VMContext {
	if ctx, ok := vmContextRegistry.Load(vm); ok {
		return ctx.(*VMContext)
	}
	return nil
}

// VMContext wraps a Goja VM with per-execution context fields.
// This allows parallel execution by giving each goroutine its own VM+context.
type VMContext struct {
	vm *goja.Runtime

	// Context fields for function execution (previously on GojaRuntime)
	workspaceName string
	stateFile     string
	scanID        string
	workflowName  string
	workflowKind  string
	target        string
	workspacePath string

	// RuntimeVars stores variables set via set_var() for retrieval with get_var()
	RuntimeVars map[string]string
}

// VMRegistrationFunc is called to register functions on a new VM
type VMRegistrationFunc func(vm *goja.Runtime)

// VMPool provides goroutine-safe access to configured Goja VMs.
// Uses sync.Pool for efficient VM reuse.
type VMPool struct {
	pool       sync.Pool
	registerFn VMRegistrationFunc
}

// NewVMPool creates a new VM pool with a function registration callback.
// The registerFn is called once per VM to register built-in functions.
func NewVMPool(registerFn VMRegistrationFunc) *VMPool {
	p := &VMPool{
		registerFn: registerFn,
	}
	p.pool = sync.Pool{
		New: func() interface{} {
			ctx := newVMContext()
			// Register functions on the new VM
			if p.registerFn != nil {
				p.registerFn(ctx.vm)
			}
			return ctx
		},
	}
	return p
}

// Get retrieves a VM context from the pool and registers it
func (p *VMPool) Get() *VMContext {
	ctx := p.pool.Get().(*VMContext)
	// Register the VM->context mapping so functions can find their context
	vmContextRegistry.Store(ctx.vm, ctx)
	return ctx
}

// Put clears context fields, unregisters the VM, and returns context to pool
func (p *VMPool) Put(ctx *VMContext) {
	// Unregister the VM->context mapping
	vmContextRegistry.Delete(ctx.vm)

	// Clear context fields to prevent data leakage between executions
	ctx.workspaceName = ""
	ctx.stateFile = ""
	ctx.scanID = ""
	ctx.workflowName = ""
	ctx.workflowKind = ""
	ctx.target = ""
	ctx.workspacePath = ""
	ctx.RuntimeVars = nil

	p.pool.Put(ctx)
}

// newVMContext creates a new Goja VM (functions will be registered by caller)
func newVMContext() *VMContext {
	return &VMContext{
		vm: goja.New(),
	}
}

// VM returns the underlying Goja VM for function registration
func (v *VMContext) VM() *goja.Runtime {
	return v.vm
}

// SetContext sets the execution context from a template context map
func (v *VMContext) SetContext(ctx map[string]interface{}) {
	// Extract workspace from context
	if ws, ok := ctx["Workspace"].(string); ok {
		v.workspaceName = ws
	} else if ws, ok := ctx["TargetSpace"].(string); ok {
		v.workspaceName = ws
	}

	// Extract state file path
	if sf, ok := ctx["StateFile"].(string); ok {
		v.stateFile = sf
	}

	// Extract scan ID
	if sid, ok := ctx["TaskID"].(string); ok {
		v.scanID = sid
	}

	// Extract workflow name and kind
	if wn, ok := ctx["WorkflowName"].(string); ok {
		v.workflowName = wn
	}
	if wk, ok := ctx["WorkflowKind"].(string); ok {
		v.workflowKind = wk
	}

	// Extract target
	if t, ok := ctx["Target"].(string); ok {
		v.target = t
	}

	// Extract workspace path (Output directory)
	if op, ok := ctx["Output"].(string); ok {
		v.workspacePath = op
	}
}

// SetVariables sets context variables on the VM
func (v *VMContext) SetVariables(ctx map[string]interface{}) error {
	for k, val := range ctx {
		if err := v.vm.Set(k, val); err != nil {
			return err
		}
	}
	return nil
}

// SetVariablesLazy sets only variables referenced in the expression.
// This is an optimization for expressions that use only a few variables
// from a large context (50-80% faster for typical pre_condition checks).
func (v *VMContext) SetVariablesLazy(ctx map[string]interface{}, expr string) error {
	referenced := extractVariables(expr)

	for _, name := range referenced {
		if val, ok := ctx[name]; ok {
			if err := v.vm.Set(name, val); err != nil {
				return err
			}
		}
	}
	return nil
}

// Run executes a JavaScript expression
func (v *VMContext) Run(expr string) (goja.Value, error) {
	return v.vm.RunString(expr)
}

// ToValue converts a Go value to a Goja value
func (v *VMContext) ToValue(val interface{}) goja.Value {
	return v.vm.ToValue(val)
}
