package executor

import (
	"context"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// StepExecutorPlugin defines the interface for step type plugins.
// Any executor that handles step types should implement this interface.
type StepExecutorPlugin interface {
	// Name returns the plugin name for logging/debugging
	Name() string

	// StepTypes returns the step types this plugin handles
	StepTypes() []core.StepType

	// Execute runs the step and returns the result
	Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error)
}

// PluginRegistry manages registered step executor plugins.
// It maps step types to their corresponding plugin implementations.
type PluginRegistry struct {
	plugins map[core.StepType]StepExecutorPlugin
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[core.StepType]StepExecutorPlugin),
	}
}

// Register adds a plugin to the registry.
// The plugin will be registered for all step types it reports via StepTypes().
func (r *PluginRegistry) Register(plugin StepExecutorPlugin) {
	for _, stepType := range plugin.StepTypes() {
		r.plugins[stepType] = plugin
	}
}

// Get returns the plugin for a step type, or nil if not found
func (r *PluginRegistry) Get(stepType core.StepType) (StepExecutorPlugin, bool) {
	plugin, ok := r.plugins[stepType]
	return plugin, ok
}

// Has checks if a step type is registered
func (r *PluginRegistry) Has(stepType core.StepType) bool {
	_, ok := r.plugins[stepType]
	return ok
}

// ListStepTypes returns all registered step types
func (r *PluginRegistry) ListStepTypes() []core.StepType {
	types := make([]core.StepType, 0, len(r.plugins))
	for t := range r.plugins {
		types = append(types, t)
	}
	return types
}
