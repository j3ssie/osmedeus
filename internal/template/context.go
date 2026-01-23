package template

import (
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// BuildContext builds a template context from an execution context
func BuildContext(execCtx *core.ExecutionContext) map[string]interface{} {
	ctx := make(map[string]interface{})

	// Add built-in constants
	ctx["DefaultUA"] = core.DefaultUA
	ctx["Version"] = core.VERSION

	// Add all variables from execution context
	for k, v := range execCtx.GetVariables() {
		ctx[k] = v
	}

	// Add standard variables
	ctx["workflow"] = execCtx.WorkflowName
	ctx["run_uuid"] = execCtx.RunUUID
	ctx["target"] = execCtx.Target
	ctx["workspace"] = execCtx.WorkspacePath
	ctx["base_folder"] = execCtx.BaseFolder

	return ctx
}

// MergeContext merges multiple contexts, later contexts override earlier ones
func MergeContext(contexts ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, ctx := range contexts {
		for k, v := range ctx {
			result[k] = v
		}
	}
	return result
}

// ContextFromParams builds a context from parameter key-value pairs
func ContextFromParams(params map[string]string) map[string]interface{} {
	ctx := make(map[string]interface{}, len(params))
	for k, v := range params {
		ctx[k] = v
	}
	return ctx
}

// ContextBuilder provides a fluent interface for building contexts
type ContextBuilder struct {
	ctx map[string]interface{}
}

// NewContextBuilder creates a new context builder
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		ctx: make(map[string]interface{}),
	}
}

// Set sets a value in the context
func (b *ContextBuilder) Set(key string, value interface{}) *ContextBuilder {
	b.ctx[key] = value
	return b
}

// SetAll sets multiple values from a map
func (b *ContextBuilder) SetAll(values map[string]interface{}) *ContextBuilder {
	for k, v := range values {
		b.ctx[k] = v
	}
	return b
}

// SetParams sets values from string parameters
func (b *ContextBuilder) SetParams(params map[string]string) *ContextBuilder {
	for k, v := range params {
		b.ctx[k] = v
	}
	return b
}

// SetFromExecution sets values from an execution context
func (b *ContextBuilder) SetFromExecution(execCtx *core.ExecutionContext) *ContextBuilder {
	return b.SetAll(BuildContext(execCtx))
}

// Build returns the built context
func (b *ContextBuilder) Build() map[string]interface{} {
	// Return a copy to prevent modifications
	result := make(map[string]interface{}, len(b.ctx))
	for k, v := range b.ctx {
		result[k] = v
	}
	return result
}
