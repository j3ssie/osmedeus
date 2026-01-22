package core

import (
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

// ExecutionContext holds runtime state for workflow execution
type ExecutionContext struct {
	WorkflowName  string
	WorkflowKind  WorkflowKind
	RunUUID       string
	Target        string
	WorkspacePath string
	BaseFolder    string

	// Params are the input parameters (immutable after init)
	Params map[string]interface{}

	// Exports are variables exported by steps (mutable)
	Exports map[string]interface{}

	// Variables combines Params and Exports for template rendering
	Variables map[string]interface{}

	// Logger for this execution
	Logger *zap.Logger

	// StepIndex tracks the current step number (for display purposes)
	StepIndex int

	// WorkspaceName is the workspace identifier for database operations
	WorkspaceName string

	// mu protects concurrent access to Exports and Variables
	mu sync.RWMutex

	// variablesSnapshot provides O(1) read access for GetVariables()
	// Updated atomically on SetVariable/SetExport/MergeExports/SetParam
	variablesSnapshot atomic.Value // map[string]interface{}
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(workflowName string, kind WorkflowKind, runUUID, target string) *ExecutionContext {
	return &ExecutionContext{
		WorkflowName: workflowName,
		WorkflowKind: kind,
		RunUUID:      runUUID,
		Target:       target,
		Params:       make(map[string]interface{}),
		Exports:      make(map[string]interface{}),
		Variables:    make(map[string]interface{}),
	}
}

// updateSnapshot creates an immutable copy of Variables for fast reads
// Must be called with c.mu held (Lock, not RLock)
func (c *ExecutionContext) updateSnapshot() {
	snapshot := make(map[string]interface{}, len(c.Variables))
	for k, v := range c.Variables {
		snapshot[k] = v
	}
	c.variablesSnapshot.Store(snapshot)
}

// SetParam sets a parameter value
func (c *ExecutionContext) SetParam(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Params[key] = value
	c.Variables[key] = value
	c.updateSnapshot()
}

// GetParam gets a parameter value
func (c *ExecutionContext) GetParam(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.Params[key]
	return v, ok
}

// SetExport sets an exported variable
func (c *ExecutionContext) SetExport(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Exports[key] = value
	c.Variables[key] = value
	c.updateSnapshot()
}

// GetExport gets an exported variable
func (c *ExecutionContext) GetExport(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.Exports[key]
	return v, ok
}

// GetVariable gets a variable (param or export)
func (c *ExecutionContext) GetVariable(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.Variables[key]
	return v, ok
}

// SetVariable sets a variable
func (c *ExecutionContext) SetVariable(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Variables[key] = value
	c.updateSnapshot()
}

// GetVariables returns all variables for template rendering.
// Uses atomic snapshot for O(1) read performance.
func (c *ExecutionContext) GetVariables() map[string]interface{} {
	// Fast path: return cached snapshot (no lock needed)
	if snapshot := c.variablesSnapshot.Load(); snapshot != nil {
		return snapshot.(map[string]interface{})
	}
	// Fallback for uninitialized contexts (shouldn't happen in normal use)
	c.mu.RLock()
	defer c.mu.RUnlock()
	vars := make(map[string]interface{}, len(c.Variables))
	for k, v := range c.Variables {
		vars[k] = v
	}
	return vars
}

// MergeExports merges exports from a step result
func (c *ExecutionContext) MergeExports(exports map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range exports {
		c.Exports[k] = v
		c.Variables[k] = v
	}
	c.updateSnapshot()
}

// Clone creates a shallow copy of the context for child execution
func (c *ExecutionContext) Clone() *ExecutionContext {
	c.mu.RLock()
	defer c.mu.RUnlock()

	clone := &ExecutionContext{
		WorkflowName:  c.WorkflowName,
		WorkflowKind:  c.WorkflowKind,
		RunUUID:       c.RunUUID,
		Target:        c.Target,
		WorkspacePath: c.WorkspacePath,
		BaseFolder:    c.BaseFolder,
		WorkspaceName: c.WorkspaceName,
		Params:        make(map[string]interface{}, len(c.Params)),
		Exports:       make(map[string]interface{}, len(c.Exports)),
		Variables:     make(map[string]interface{}, len(c.Variables)),
		Logger:        c.Logger,
	}

	for k, v := range c.Params {
		clone.Params[k] = v
	}
	for k, v := range c.Exports {
		clone.Exports[k] = v
	}
	for k, v := range c.Variables {
		clone.Variables[k] = v
	}

	// Initialize snapshot for fast GetVariables() reads
	clone.updateSnapshot()

	return clone
}

// CloneForLoop creates an optimized clone for foreach/parallel iterations.
// Key optimizations:
//   - Shares Params reference (documented as immutable after init)
//   - Pre-sets loop variables to avoid separate SetVariable calls
//   - Reduces map copy overhead by ~33% (skips Params copy)
func (c *ExecutionContext) CloneForLoop(loopVar string, loopValue interface{}, iterID int) *ExecutionContext {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Estimate capacity: parent variables + 2 loop variables
	varCapacity := len(c.Variables) + 2

	clone := &ExecutionContext{
		WorkflowName:  c.WorkflowName,
		WorkflowKind:  c.WorkflowKind,
		RunUUID:       c.RunUUID,
		Target:        c.Target,
		WorkspacePath: c.WorkspacePath,
		BaseFolder:    c.BaseFolder,
		WorkspaceName: c.WorkspaceName,
		Logger:        c.Logger,
		// Share immutable Params reference (no copy needed)
		Params: c.Params,
		// Fresh exports map for this iteration
		Exports: make(map[string]interface{}, 4),
		// Variables map with pre-allocated capacity
		Variables: make(map[string]interface{}, varCapacity),
	}

	// Copy parent Variables for template rendering
	for k, v := range c.Variables {
		clone.Variables[k] = v
	}

	// Pre-set loop variables (avoids separate SetVariable calls)
	if loopVar != "" {
		clone.Variables[loopVar] = loopValue
	}
	clone.Variables["_id_"] = iterID

	// Initialize snapshot for fast GetVariables() reads
	clone.updateSnapshot()

	return clone
}
