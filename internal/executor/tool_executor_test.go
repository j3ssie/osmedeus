package executor

import (
	"context"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolExecutorRegistry_RegisterAndGet(t *testing.T) {
	reg := NewToolExecutorRegistry()

	preset := NewPresetToolExecutor("bash", nil)
	reg.Register(preset)

	got, ok := reg.Get("bash")
	assert.True(t, ok)
	assert.Equal(t, "bash", got.Name())

	_, ok = reg.Get("nonexistent")
	assert.False(t, ok)
}

func TestToolExecutorRegistry_ExecuteUnknown(t *testing.T) {
	reg := NewToolExecutorRegistry()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	_, err := reg.Execute(context.Background(), "unknown_tool", nil, execCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown tool")
}

func TestPresetToolExecutor_Name(t *testing.T) {
	preset := NewPresetToolExecutor("bash", nil)
	assert.Equal(t, "bash", preset.Name())
}

func TestPresetToolExecutor_Execute(t *testing.T) {
	funcRegistry := functions.NewRegistry()
	preset := NewPresetToolExecutor("bash", funcRegistry)
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	args := map[string]interface{}{"command": "echo hello"}
	result, err := preset.Execute(context.Background(), args, execCtx)
	require.NoError(t, err)
	assert.Contains(t, result, "hello")
}

func TestCustomToolExecutor_Name(t *testing.T) {
	custom := NewCustomToolExecutor("greet", `"Hello " + args.name`, nil)
	assert.Equal(t, "greet", custom.Name())
}

func TestCustomToolExecutor_Execute(t *testing.T) {
	funcRegistry := functions.NewRegistry()
	custom := NewCustomToolExecutor("greet", `"Hello " + args.name`, funcRegistry)
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	args := map[string]interface{}{"name": "World"}
	result, err := custom.Execute(context.Background(), args, execCtx)
	require.NoError(t, err)
	assert.Equal(t, "Hello World", result)
}

func TestBuildToolRegistry_PresetOnly(t *testing.T) {
	funcRegistry := functions.NewRegistry()
	toolDefs := []core.AgentToolDef{
		{Preset: "bash"},
		{Preset: "read_file"},
	}

	reg := BuildToolRegistry(toolDefs, funcRegistry)

	_, ok := reg.Get("bash")
	assert.True(t, ok)

	_, ok = reg.Get("read_file")
	assert.True(t, ok)

	_, ok = reg.Get("nonexistent")
	assert.False(t, ok)
}

func TestBuildToolRegistry_CustomOnly(t *testing.T) {
	funcRegistry := functions.NewRegistry()
	toolDefs := []core.AgentToolDef{
		{
			Name:        "greet",
			Description: "Greet someone",
			Handler:     `"Hello " + args.name`,
		},
	}

	reg := BuildToolRegistry(toolDefs, funcRegistry)

	te, ok := reg.Get("greet")
	assert.True(t, ok)
	assert.Equal(t, "greet", te.Name())
}

func TestBuildToolRegistry_Mixed(t *testing.T) {
	funcRegistry := functions.NewRegistry()
	toolDefs := []core.AgentToolDef{
		{Preset: "bash"},
		{
			Name:        "custom_tool",
			Description: "A custom tool",
			Handler:     `"result"`,
		},
	}

	reg := BuildToolRegistry(toolDefs, funcRegistry)

	_, ok := reg.Get("bash")
	assert.True(t, ok)

	_, ok = reg.Get("custom_tool")
	assert.True(t, ok)
}

func TestSubAgentToolExecutor_Name(t *testing.T) {
	exec := &SubAgentToolExecutor{}
	assert.Equal(t, core.SpawnAgentToolName, exec.Name())
}

func TestSubAgentToolExecutor_Execute_UnknownAgent(t *testing.T) {
	exec := &SubAgentToolExecutor{
		subAgents: map[string]core.SubAgentDef{
			"recon": {Name: "recon"},
		},
	}
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	_, err := exec.Execute(context.Background(), map[string]interface{}{
		"agent": "nonexistent",
		"query": "do something",
	}, execCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown sub-agent")
}

func TestSubAgentToolExecutor_Execute_MissingQuery(t *testing.T) {
	exec := &SubAgentToolExecutor{
		subAgents: map[string]core.SubAgentDef{
			"recon": {Name: "recon"},
		},
	}
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	_, err := exec.Execute(context.Background(), map[string]interface{}{
		"agent": "recon",
	}, execCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires 'query'")
}

func TestSubAgentToolExecutor_Execute_MissingAgent(t *testing.T) {
	exec := &SubAgentToolExecutor{
		subAgents: map[string]core.SubAgentDef{},
	}
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	_, err := exec.Execute(context.Background(), map[string]interface{}{
		"query": "do something",
	}, execCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires 'agent'")
}

func TestSubAgentToolExecutor_Execute_DepthLimit(t *testing.T) {
	exec := &SubAgentToolExecutor{
		subAgents: map[string]core.SubAgentDef{
			"recon": {Name: "recon", AgentTools: []core.AgentToolDef{{Preset: "bash"}}},
		},
		currentDepth: 3,
		maxDepth:     3,
	}
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	_, err := exec.Execute(context.Background(), map[string]interface{}{
		"agent": "recon",
		"query": "scan ports",
	}, execCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "depth limit exceeded")
}

func TestBuildToolRegistryWithSubAgents(t *testing.T) {
	funcRegistry := functions.NewRegistry()
	toolDefs := []core.AgentToolDef{{Preset: "bash"}}
	subAgents := []core.SubAgentDef{
		{Name: "recon", Description: "Recon agent", AgentTools: []core.AgentToolDef{{Preset: "bash"}}},
	}

	reg := BuildToolRegistryWithSubAgents(
		toolDefs, funcRegistry, nil, nil, true,
		0, 3, nil, subAgents,
	)

	// Should have bash and spawn_agent
	_, ok := reg.Get("bash")
	assert.True(t, ok)

	_, ok = reg.Get(core.SpawnAgentToolName)
	assert.True(t, ok)
}

func TestBuildToolRegistryWithSubAgents_NoSubAgents(t *testing.T) {
	funcRegistry := functions.NewRegistry()
	toolDefs := []core.AgentToolDef{{Preset: "bash"}}

	reg := BuildToolRegistryWithSubAgents(
		toolDefs, funcRegistry, nil, nil, true,
		0, 3, nil, nil,
	)

	// Should have bash only (no spawn_agent)
	_, ok := reg.Get("bash")
	assert.True(t, ok)

	_, ok = reg.Get(core.SpawnAgentToolName)
	assert.False(t, ok)
}

func TestBuildSyntheticStep(t *testing.T) {
	sa := &core.SubAgentDef{
		Name:          "recon",
		SystemPrompt:  "You are a recon specialist",
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}, {Preset: "http_get"}},
		MaxIterations: 5,
		Models:        []string{"gpt-4o"},
		StopCondition: `contains(agent_content, "DONE")`,
		SubAgents: []core.SubAgentDef{
			{Name: "nested", AgentTools: []core.AgentToolDef{{Preset: "bash"}}},
		},
	}

	step := buildSyntheticStep(sa, "scan ports on target")
	assert.Equal(t, "sub-agent-recon", step.Name)
	assert.Equal(t, core.StepTypeAgent, step.Type)
	assert.Equal(t, "scan ports on target", step.Query)
	assert.Equal(t, "You are a recon specialist", step.SystemPrompt)
	assert.Len(t, step.AgentTools, 2)
	assert.Equal(t, 5, step.MaxIterations)
	assert.Equal(t, []string{"gpt-4o"}, step.Models)
	assert.Len(t, step.SubAgents, 1)
}

func TestBuildSyntheticStep_DefaultMaxIterations(t *testing.T) {
	sa := &core.SubAgentDef{
		Name:       "minimal",
		AgentTools: []core.AgentToolDef{{Preset: "bash"}},
	}
	step := buildSyntheticStep(sa, "test")
	assert.Equal(t, 10, step.MaxIterations)
}

func TestToolExecutorRegistry_FullFlow(t *testing.T) {
	funcRegistry := functions.NewRegistry()
	toolDefs := []core.AgentToolDef{
		{Preset: "bash"},
		{
			Name:        "echo_tool",
			Description: "Echo something",
			Handler:     `"echoed: " + args.text`,
		},
	}

	reg := BuildToolRegistry(toolDefs, funcRegistry)
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	// Test preset execution
	result, err := reg.Execute(context.Background(), "bash", map[string]interface{}{"command": "echo test"}, execCtx)
	require.NoError(t, err)
	assert.Contains(t, result, "test")

	// Test custom execution
	result, err = reg.Execute(context.Background(), "echo_tool", map[string]interface{}{"text": "hello"}, execCtx)
	require.NoError(t, err)
	assert.Equal(t, "echoed: hello", result)
}
