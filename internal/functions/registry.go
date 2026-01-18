package functions

import (
	"fmt"
	"sync"
)

// Registry manages utility functions
type Registry struct {
	runtime *OttoRuntime
	mu      sync.RWMutex
}

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	r := &Registry{
		runtime: NewOttoRuntime(),
	}
	return r
}

// Execute executes a function expression and returns the result
func (r *Registry) Execute(expr string, ctx map[string]interface{}) (interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.runtime.Execute(expr, ctx)
}

// EvaluateCondition evaluates a condition expression and returns true/false
func (r *Registry) EvaluateCondition(condition string, ctx map[string]interface{}) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.runtime.EvaluateCondition(condition, ctx)
}

// EvaluateExports evaluates export expressions and returns the results
func (r *Registry) EvaluateExports(exports map[string]string, ctx map[string]interface{}) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]interface{}, len(exports))
	for name, expr := range exports {
		value, err := r.runtime.Execute(expr, ctx)
		if err != nil {
			return nil, fmt.Errorf("error evaluating export %s: %w", name, err)
		}
		results[name] = value
	}
	return results, nil
}

// Register registers a custom function
func (r *Registry) Register(name string, fn interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.runtime.Register(name, fn)
}

// Reset creates a fresh runtime instance
func (r *Registry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.runtime = NewOttoRuntime()
}

// GetRuntime returns the Otto runtime (for testing)
func (r *Registry) GetRuntime() *OttoRuntime {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.runtime
}

// DefaultRegistry is the global function registry
var DefaultRegistry = NewRegistry()

// Execute executes a function using the default registry
func Execute(expr string, ctx map[string]interface{}) (interface{}, error) {
	return DefaultRegistry.Execute(expr, ctx)
}

// EvaluateCondition evaluates a condition using the default registry
func EvaluateCondition(condition string, ctx map[string]interface{}) (bool, error) {
	return DefaultRegistry.EvaluateCondition(condition, ctx)
}

// EvaluateExports evaluates exports using the default registry
func EvaluateExports(exports map[string]string, ctx map[string]interface{}) (map[string]interface{}, error) {
	return DefaultRegistry.EvaluateExports(exports, ctx)
}
