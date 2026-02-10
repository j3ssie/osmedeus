package core

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
)

func TestAgentToolDef_IsPreset(t *testing.T) {
	t.Run("preset tool", func(t *testing.T) {
		tool := AgentToolDef{Preset: "bash"}
		assert.True(t, tool.IsPreset())
	})

	t.Run("custom tool", func(t *testing.T) {
		tool := AgentToolDef{Name: "my_tool", Description: "A custom tool"}
		assert.False(t, tool.IsPreset())
	})

	t.Run("empty tool", func(t *testing.T) {
		tool := AgentToolDef{}
		assert.False(t, tool.IsPreset())
	})
}

func TestAgentToolDef_YAMLUnmarshal(t *testing.T) {
	t.Run("preset tool", func(t *testing.T) {
		yamlData := `preset: bash`
		var tool AgentToolDef
		err := yaml.Unmarshal([]byte(yamlData), &tool)
		assert.NoError(t, err)
		assert.Equal(t, "bash", tool.Preset)
		assert.True(t, tool.IsPreset())
	})

	t.Run("custom tool with handler", func(t *testing.T) {
		yamlData := `
name: nuclei_scan
description: "Run nuclei scanner"
parameters:
  type: object
  properties:
    url:
      type: string
      description: "Target URL"
  required:
    - url
handler: 'exec_cmd("nuclei -u " + args.url)'
`
		var tool AgentToolDef
		err := yaml.Unmarshal([]byte(yamlData), &tool)
		assert.NoError(t, err)
		assert.Equal(t, "nuclei_scan", tool.Name)
		assert.Equal(t, "Run nuclei scanner", tool.Description)
		assert.NotNil(t, tool.Parameters)
		assert.Equal(t, `exec_cmd("nuclei -u " + args.url)`, tool.Handler)
		assert.False(t, tool.IsPreset())
	})
}

func TestAgentMemoryConfig_YAMLUnmarshal(t *testing.T) {
	t.Run("full memory config", func(t *testing.T) {
		yamlData := `
max_messages: 50
persist_path: "/tmp/agent/conv.json"
resume_path: "/tmp/agent/prev.json"
`
		var mem AgentMemoryConfig
		err := yaml.Unmarshal([]byte(yamlData), &mem)
		assert.NoError(t, err)
		assert.Equal(t, 50, mem.MaxMessages)
		assert.Equal(t, "/tmp/agent/conv.json", mem.PersistPath)
		assert.Equal(t, "/tmp/agent/prev.json", mem.ResumePath)
	})

	t.Run("empty memory config", func(t *testing.T) {
		yamlData := `{}`
		var mem AgentMemoryConfig
		err := yaml.Unmarshal([]byte(yamlData), &mem)
		assert.NoError(t, err)
		assert.Equal(t, 0, mem.MaxMessages)
		assert.Empty(t, mem.PersistPath)
		assert.Empty(t, mem.ResumePath)
	})
}

func TestStep_IsAgentStep(t *testing.T) {
	t.Run("agent step", func(t *testing.T) {
		step := Step{Type: StepTypeAgent}
		assert.True(t, step.IsAgentStep())
	})

	t.Run("non-agent step", func(t *testing.T) {
		step := Step{Type: StepTypeBash}
		assert.False(t, step.IsAgentStep())
	})
}

func TestStep_AgentFields_YAMLUnmarshal(t *testing.T) {
	yamlData := `
name: test-agent
type: agent
query: "Analyze {{Target}}"
system_prompt: "You are a helpful assistant."
max_iterations: 10
stop_condition: 'contains(agent_content, "DONE")'
parallel_tool_calls: false
agent_tools:
  - preset: bash
  - preset: read_file
  - name: custom_tool
    description: "A custom tool"
    parameters:
      type: object
      properties:
        input:
          type: string
    handler: 'trim(args.input)'
memory:
  max_messages: 50
  persist_path: "/tmp/conv.json"
llm_config:
  model: "gpt-4o"
  temperature: 0.2
exports:
  result: "{{agent_content}}"
`
	var step Step
	err := yaml.Unmarshal([]byte(yamlData), &step)
	assert.NoError(t, err)

	assert.Equal(t, "test-agent", step.Name)
	assert.Equal(t, StepTypeAgent, step.Type)
	assert.Equal(t, "Analyze {{Target}}", step.Query)
	assert.Equal(t, "You are a helpful assistant.", step.SystemPrompt)
	assert.Equal(t, 10, step.MaxIterations)
	assert.Equal(t, `contains(agent_content, "DONE")`, step.StopCondition)
	assert.NotNil(t, step.ParallelToolCalls)
	assert.False(t, *step.ParallelToolCalls)

	// Agent tools
	assert.Len(t, step.AgentTools, 3)
	assert.Equal(t, "bash", step.AgentTools[0].Preset)
	assert.Equal(t, "read_file", step.AgentTools[1].Preset)
	assert.Equal(t, "custom_tool", step.AgentTools[2].Name)
	assert.Equal(t, `trim(args.input)`, step.AgentTools[2].Handler)

	// Memory
	assert.NotNil(t, step.Memory)
	assert.Equal(t, 50, step.Memory.MaxMessages)
	assert.Equal(t, "/tmp/conv.json", step.Memory.PersistPath)

	// LLM Config
	assert.NotNil(t, step.LLMConfig)
	assert.Equal(t, "gpt-4o", step.LLMConfig.Model)

	// Exports
	assert.Equal(t, "{{agent_content}}", step.Exports["result"])
}

func TestSubAgentDef_DeepCopy(t *testing.T) {
	t.Run("deep copy independence", func(t *testing.T) {
		original := SubAgentDef{
			Name:         "recon",
			Description:  "Recon agent",
			SystemPrompt: "You are a recon specialist",
			AgentTools: []AgentToolDef{
				{Preset: "bash"},
				{Preset: "http_get"},
			},
			Models:        []string{"gpt-4o", "claude-3"},
			MaxIterations: 5,
			LLMConfig:     &LLMStepConfig{Model: "gpt-4o"},
			Memory:        &AgentMemoryConfig{MaxMessages: 20, PersistPath: "/tmp/conv.json"},
			StopCondition: `contains(agent_content, "DONE")`,
			SubAgents: []SubAgentDef{
				{
					Name:       "nested",
					AgentTools: []AgentToolDef{{Preset: "read_file"}},
				},
			},
			OnToolStart: `log_info("start")`,
			OnToolEnd:   `log_info("end")`,
		}

		cp := original.DeepCopy()

		// Verify values match
		assert.Equal(t, "recon", cp.Name)
		assert.Equal(t, "Recon agent", cp.Description)
		assert.Equal(t, "You are a recon specialist", cp.SystemPrompt)
		assert.Len(t, cp.AgentTools, 2)
		assert.Len(t, cp.Models, 2)
		assert.Equal(t, "gpt-4o", cp.LLMConfig.Model)
		assert.Equal(t, 20, cp.Memory.MaxMessages)
		assert.Len(t, cp.SubAgents, 1)
		assert.Equal(t, "nested", cp.SubAgents[0].Name)

		// Modify original, verify copy is independent
		original.AgentTools[0].Preset = "modified"
		assert.Equal(t, "bash", cp.AgentTools[0].Preset)

		original.Models[0] = "modified"
		assert.Equal(t, "gpt-4o", cp.Models[0])

		original.LLMConfig.Model = "modified"
		assert.Equal(t, "gpt-4o", cp.LLMConfig.Model)

		original.Memory.MaxMessages = 999
		assert.Equal(t, 20, cp.Memory.MaxMessages)

		original.SubAgents[0].Name = "modified"
		assert.Equal(t, "nested", cp.SubAgents[0].Name)
	})

	t.Run("nil pointers", func(t *testing.T) {
		original := SubAgentDef{
			Name: "minimal",
		}
		cp := original.DeepCopy()
		assert.Equal(t, "minimal", cp.Name)
		assert.Nil(t, cp.LLMConfig)
		assert.Nil(t, cp.Memory)
		assert.Nil(t, cp.SubAgents)
	})
}

func TestSubAgentDef_YAMLUnmarshal(t *testing.T) {
	yamlData := `
name: recon_agent
description: "Specialized agent for recon"
system_prompt: "You are a recon specialist"
max_iterations: 5
agent_tools:
  - preset: bash
  - preset: http_get
sub_agents:
  - name: port_scanner
    description: "Scans ports"
    agent_tools:
      - preset: bash
`
	var sa SubAgentDef
	err := yaml.Unmarshal([]byte(yamlData), &sa)
	assert.NoError(t, err)
	assert.Equal(t, "recon_agent", sa.Name)
	assert.Equal(t, "Specialized agent for recon", sa.Description)
	assert.Equal(t, "You are a recon specialist", sa.SystemPrompt)
	assert.Equal(t, 5, sa.MaxIterations)
	assert.Len(t, sa.AgentTools, 2)
	assert.Len(t, sa.SubAgents, 1)
	assert.Equal(t, "port_scanner", sa.SubAgents[0].Name)
}

func TestStep_SubAgents_YAMLUnmarshal(t *testing.T) {
	yamlData := `
name: orchestrator
type: agent
query: "Coordinate analysis"
system_prompt: "You are an orchestrator"
max_iterations: 10
max_agent_depth: 2
agent_tools:
  - preset: bash
sub_agents:
  - name: recon_agent
    description: "Recon specialist"
    system_prompt: "You do recon"
    max_iterations: 5
    agent_tools:
      - preset: bash
      - preset: http_get
  - name: vuln_scanner
    description: "Vulnerability scanner"
    agent_tools:
      - preset: bash
`
	var step Step
	err := yaml.Unmarshal([]byte(yamlData), &step)
	assert.NoError(t, err)
	assert.Equal(t, "orchestrator", step.Name)
	assert.Equal(t, 2, step.MaxAgentDepth)
	assert.Len(t, step.SubAgents, 2)
	assert.Equal(t, "recon_agent", step.SubAgents[0].Name)
	assert.Equal(t, "Recon specialist", step.SubAgents[0].Description)
	assert.Len(t, step.SubAgents[0].AgentTools, 2)
	assert.Equal(t, "vuln_scanner", step.SubAgents[1].Name)
}

func TestStep_Clone_AgentFields(t *testing.T) {
	ptc := false
	step := Step{
		Name:              "agent-step",
		Type:              StepTypeAgent,
		Query:             "test query",
		SystemPrompt:      "test prompt",
		MaxIterations:     10,
		StopCondition:     "test condition",
		ParallelToolCalls: &ptc,
		AgentTools: []AgentToolDef{
			{Preset: "bash"},
			{Name: "custom", Handler: "handler()"},
		},
		Memory: &AgentMemoryConfig{
			MaxMessages: 50,
			PersistPath: "/tmp/conv.json",
		},
		SubAgents: []SubAgentDef{
			{Name: "sub1", AgentTools: []AgentToolDef{{Preset: "bash"}}},
		},
		MaxAgentDepth: 2,
	}

	cloned := step.Clone()

	// Verify values are copied
	assert.Equal(t, "agent-step", cloned.Name)
	assert.Equal(t, "test query", cloned.Query)
	assert.Equal(t, "test prompt", cloned.SystemPrompt)
	assert.Equal(t, 10, cloned.MaxIterations)
	assert.Equal(t, "test condition", cloned.StopCondition)
	assert.NotNil(t, cloned.ParallelToolCalls)
	assert.False(t, *cloned.ParallelToolCalls)

	// Agent tools are deep copied
	assert.Len(t, cloned.AgentTools, 2)
	assert.Equal(t, "bash", cloned.AgentTools[0].Preset)

	// Modify original, verify clone is independent
	step.AgentTools[0].Preset = "modified"
	assert.Equal(t, "bash", cloned.AgentTools[0].Preset)

	// Memory is deep copied
	assert.NotNil(t, cloned.Memory)
	assert.Equal(t, 50, cloned.Memory.MaxMessages)
	step.Memory.MaxMessages = 100
	assert.Equal(t, 50, cloned.Memory.MaxMessages)

	// SubAgents are deep copied
	assert.Len(t, cloned.SubAgents, 1)
	assert.Equal(t, "sub1", cloned.SubAgents[0].Name)
	step.SubAgents[0].Name = "modified"
	assert.Equal(t, "sub1", cloned.SubAgents[0].Name)
	assert.Equal(t, 2, cloned.MaxAgentDepth)
}
