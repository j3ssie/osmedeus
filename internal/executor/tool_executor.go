package executor

import (
	"context"
	"fmt"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// ToolExecutor defines the interface for executing agent tool calls.
// Each implementation handles a specific tool (preset or custom).
type ToolExecutor interface {
	// Name returns the tool name (matches the function name in LLM tool calls)
	Name() string
	// Execute runs the tool with the given arguments and returns the result string
	Execute(ctx context.Context, args map[string]interface{}, execCtx *core.ExecutionContext) (string, error)
}

// ToolExecutorRegistry manages a collection of ToolExecutor instances
type ToolExecutorRegistry struct {
	executors map[string]ToolExecutor
}

// NewToolExecutorRegistry creates an empty registry
func NewToolExecutorRegistry() *ToolExecutorRegistry {
	return &ToolExecutorRegistry{
		executors: make(map[string]ToolExecutor),
	}
}

// Register adds a ToolExecutor to the registry
func (r *ToolExecutorRegistry) Register(te ToolExecutor) {
	r.executors[te.Name()] = te
}

// Get returns the ToolExecutor for the given name
func (r *ToolExecutorRegistry) Get(name string) (ToolExecutor, bool) {
	te, ok := r.executors[name]
	return te, ok
}

// Execute dispatches a tool call to the appropriate executor
func (r *ToolExecutorRegistry) Execute(ctx context.Context, name string, args map[string]interface{}, execCtx *core.ExecutionContext) (string, error) {
	te, ok := r.executors[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return te.Execute(ctx, args, execCtx)
}

// PresetToolExecutor wraps a preset tool and executes it via the function registry
type PresetToolExecutor struct {
	name     string
	registry *functions.Registry
}

// NewPresetToolExecutor creates a preset tool executor
func NewPresetToolExecutor(name string, registry *functions.Registry) *PresetToolExecutor {
	return &PresetToolExecutor{name: name, registry: registry}
}

// Name returns the tool name
func (e *PresetToolExecutor) Name() string {
	return e.name
}

// Execute runs the preset tool by building a function call expression
func (e *PresetToolExecutor) Execute(_ context.Context, args map[string]interface{}, execCtx *core.ExecutionContext) (string, error) {
	expr := buildPresetCallExpr(e.name, args)
	vars := execCtx.GetVariables()
	result, err := e.registry.Execute(expr, vars)
	if err != nil {
		return "", fmt.Errorf("preset tool '%s' failed: %w", e.name, err)
	}
	return formatToolResult(result), nil
}

// CustomToolExecutor wraps a custom tool with a JS handler expression
type CustomToolExecutor struct {
	name     string
	handler  string
	registry *functions.Registry
}

// NewCustomToolExecutor creates a custom tool executor
func NewCustomToolExecutor(name, handler string, registry *functions.Registry) *CustomToolExecutor {
	return &CustomToolExecutor{name: name, handler: handler, registry: registry}
}

// Name returns the tool name
func (e *CustomToolExecutor) Name() string {
	return e.name
}

// Execute runs the custom tool handler JS expression
func (e *CustomToolExecutor) Execute(_ context.Context, args map[string]interface{}, execCtx *core.ExecutionContext) (string, error) {
	vars := execCtx.GetVariables()
	vars["args"] = args
	result, err := e.registry.Execute(e.handler, vars)
	if err != nil {
		return "", fmt.Errorf("custom handler '%s' failed: %w", e.name, err)
	}
	return formatToolResult(result), nil
}

// BuildToolRegistry constructs a ToolExecutorRegistry from the resolved agent tool definitions
func BuildToolRegistry(toolDefs []core.AgentToolDef, funcRegistry *functions.Registry) *ToolExecutorRegistry {
	log := logger.Get()
	reg := NewToolExecutorRegistry()

	for _, def := range toolDefs {
		if def.IsPreset() {
			reg.Register(NewPresetToolExecutor(def.Preset, funcRegistry))
		} else if def.Handler != "" {
			reg.Register(NewCustomToolExecutor(def.Name, def.Handler, funcRegistry))
		} else {
			log.Debug("Tool definition has no handler, registering as preset fallback",
				zap.String("tool", def.Name),
			)
			reg.Register(NewPresetToolExecutor(def.Name, funcRegistry))
		}
	}

	return reg
}

// SubAgentToolExecutor handles spawn_agent tool calls by creating
// a child AgentExecutor and running the specified sub-agent.
type SubAgentToolExecutor struct {
	subAgents      map[string]core.SubAgentDef
	templateEngine template.TemplateEngine
	funcRegistry   *functions.Registry
	config         *config.Config
	silent         bool
	currentDepth   int
	maxDepth       int
	parentState    *agentState // for token merging
}

// Name returns the tool name
func (e *SubAgentToolExecutor) Name() string {
	return core.SpawnAgentToolName
}

// Execute spawns a child agent and returns its output as the tool result.
func (e *SubAgentToolExecutor) Execute(ctx context.Context, args map[string]interface{}, execCtx *core.ExecutionContext) (string, error) {
	log := logger.Get()

	agentName, _ := args["agent"].(string)
	query, _ := args["query"].(string)

	if agentName == "" {
		return "", fmt.Errorf("spawn_agent requires 'agent' parameter")
	}
	if query == "" {
		return "", fmt.Errorf("spawn_agent requires 'query' parameter")
	}

	// Look up sub-agent definition
	saDef, ok := e.subAgents[agentName]
	if !ok {
		return "", fmt.Errorf("unknown sub-agent: %s", agentName)
	}

	// Check depth limit
	childDepth := e.currentDepth + 1
	if childDepth > e.maxDepth {
		return "", fmt.Errorf("sub-agent depth limit exceeded (depth %d > max %d)", childDepth, e.maxDepth)
	}

	log.Debug("Spawning sub-agent",
		zap.String("agent", agentName),
		zap.Int("depth", childDepth),
		zap.Int("max_depth", e.maxDepth),
	)

	// Build synthetic step from SubAgentDef
	syntheticStep := buildSyntheticStep(&saDef, query)

	// Create child AgentExecutor
	childExec := NewAgentExecutor(e.templateEngine, e.funcRegistry)
	childExec.SetConfig(e.config)
	childExec.SetSilent(e.silent)
	childExec.SetDepthContext(childDepth, e.maxDepth)

	// Execute child agent
	result, err := childExec.Execute(ctx, syntheticStep, execCtx)
	if err != nil {
		// Return error as string tool result (don't crash parent)
		errMsg := fmt.Sprintf("Sub-agent '%s' failed: %s", agentName, err.Error())
		log.Warn("Sub-agent execution failed",
			zap.String("agent", agentName),
			zap.Error(err),
		)
		return errMsg, nil
	}

	// Merge child tokens into parent
	if e.parentState != nil {
		childTotal, _ := result.Exports["agent_total_tokens"].(int)
		childPrompt, _ := result.Exports["agent_prompt_tokens"].(int)
		childCompletion, _ := result.Exports["agent_completion_tokens"].(int)
		e.parentState.MergeTokens(childTotal, childPrompt, childCompletion)
	}

	return result.Output, nil
}

// buildSyntheticStep creates a Step from a SubAgentDef and query.
func buildSyntheticStep(sa *core.SubAgentDef, query string) *core.Step {
	maxIterations := sa.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 10
	}

	step := &core.Step{
		Name:          "sub-agent-" + sa.Name,
		Type:          core.StepTypeAgent,
		Query:         query,
		SystemPrompt:  sa.SystemPrompt,
		AgentTools:    sa.AgentTools,
		MaxIterations: maxIterations,
		Models:        sa.Models,
		LLMConfig:     sa.LLMConfig,
		OutputSchema:  sa.OutputSchema,
		Memory:        sa.Memory,
		StopCondition: sa.StopCondition,
		SubAgents:     sa.SubAgents,
		OnToolStart:   sa.OnToolStart,
		OnToolEnd:     sa.OnToolEnd,
	}

	return step
}

// BuildToolRegistryWithSubAgents constructs a ToolExecutorRegistry that includes
// both standard tools and the spawn_agent tool for sub-agent delegation.
func BuildToolRegistryWithSubAgents(
	toolDefs []core.AgentToolDef,
	funcRegistry *functions.Registry,
	engine template.TemplateEngine,
	cfg *config.Config,
	silent bool,
	currentDepth int,
	maxDepth int,
	parentState *agentState,
	subAgents []core.SubAgentDef,
) *ToolExecutorRegistry {
	reg := BuildToolRegistry(toolDefs, funcRegistry)

	if len(subAgents) > 0 {
		// Build name → def map
		saMap := make(map[string]core.SubAgentDef, len(subAgents))
		for _, sa := range subAgents {
			saMap[sa.Name] = sa
		}

		reg.Register(&SubAgentToolExecutor{
			subAgents:      saMap,
			templateEngine: engine,
			funcRegistry:   funcRegistry,
			config:         cfg,
			silent:         silent,
			currentDepth:   currentDepth,
			maxDepth:       maxDepth,
			parentState:    parentState,
		})
	}

	return reg
}

// executeToolCallViaRegistry parses tool call arguments and dispatches to the registry
func executeToolCallViaRegistry(
	ctx context.Context,
	tc core.LLMToolCall,
	toolRegistry *ToolExecutorRegistry,
	execCtx *core.ExecutionContext,
	log *zap.Logger,
) (string, error) {
	funcName := tc.Function.Name
	argsJSON := tc.Function.Arguments

	log.Debug("Executing tool call via registry",
		zap.String("tool", funcName),
		zap.String("args", argsJSON),
	)

	var args map[string]interface{}
	if argsJSON != "" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", fmt.Errorf("failed to parse tool arguments for %s: %w", funcName, err)
		}
	}
	if args == nil {
		args = make(map[string]interface{})
	}

	return toolRegistry.Execute(ctx, funcName, args, execCtx)
}
