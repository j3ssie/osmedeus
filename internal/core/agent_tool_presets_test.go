package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPresetTool(t *testing.T) {
	t.Run("known preset", func(t *testing.T) {
		tool, ok := GetPresetTool("bash")
		assert.True(t, ok)
		assert.Equal(t, "function", tool.Type)
		assert.Equal(t, "bash", tool.Function.Name)
		assert.NotEmpty(t, tool.Function.Description)
		assert.NotNil(t, tool.Function.Parameters)
	})

	t.Run("all presets have valid schemas", func(t *testing.T) {
		for name := range PresetToolRegistry {
			tool, ok := GetPresetTool(name)
			assert.True(t, ok, "preset %s should exist", name)
			assert.Equal(t, "function", tool.Type)
			assert.Equal(t, name, tool.Function.Name)
			assert.NotEmpty(t, tool.Function.Description, "preset %s should have description", name)
			assert.NotNil(t, tool.Function.Parameters, "preset %s should have parameters", name)

			// Verify parameters have 'type' and 'properties'
			params := tool.Function.Parameters
			assert.Equal(t, "object", params["type"], "preset %s parameters should be object type", name)
			props, ok := params["properties"]
			assert.True(t, ok, "preset %s should have properties", name)
			assert.NotNil(t, props, "preset %s properties should not be nil", name)
		}
	})

	t.Run("unknown preset", func(t *testing.T) {
		_, ok := GetPresetTool("nonexistent")
		assert.False(t, ok)
	})
}

func TestGetPresetTool_SpecificPresets(t *testing.T) {
	presets := []string{
		"bash", "read_file", "read_lines", "file_exists", "file_length",
		"append_file", "save_content", "glob", "grep_string", "grep_regex",
		"http_get", "http_request", "jq",
		"exec_python", "exec_python_file", "run_module", "run_flow",
	}

	for _, name := range presets {
		t.Run(name, func(t *testing.T) {
			tool, ok := GetPresetTool(name)
			require.True(t, ok, "preset %s should be registered", name)
			assert.Equal(t, name, tool.Function.Name)

			// Verify required fields are present
			params := tool.Function.Parameters
			required, hasRequired := params["required"]
			if hasRequired {
				reqSlice, ok := required.([]string)
				assert.True(t, ok, "required should be []string for %s", name)
				assert.NotEmpty(t, reqSlice, "required should not be empty for %s", name)
			}
		})
	}
}

func TestResolveAgentTools(t *testing.T) {
	t.Run("preset tools only", func(t *testing.T) {
		defs := []AgentToolDef{
			{Preset: "bash"},
			{Preset: "read_file"},
			{Preset: "file_exists"},
		}

		tools, err := ResolveAgentTools(defs)
		require.NoError(t, err)
		assert.Len(t, tools, 3)
		assert.Equal(t, "bash", tools[0].Function.Name)
		assert.Equal(t, "read_file", tools[1].Function.Name)
		assert.Equal(t, "file_exists", tools[2].Function.Name)
	})

	t.Run("custom tool", func(t *testing.T) {
		defs := []AgentToolDef{
			{
				Name:        "my_tool",
				Description: "My custom tool",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input": map[string]interface{}{
							"type": "string",
						},
					},
				},
				Handler: "exec_cmd(args.input)",
			},
		}

		tools, err := ResolveAgentTools(defs)
		require.NoError(t, err)
		assert.Len(t, tools, 1)
		assert.Equal(t, "function", tools[0].Type)
		assert.Equal(t, "my_tool", tools[0].Function.Name)
		assert.Equal(t, "My custom tool", tools[0].Function.Description)
	})

	t.Run("mixed preset and custom", func(t *testing.T) {
		defs := []AgentToolDef{
			{Preset: "bash"},
			{
				Name:        "custom",
				Description: "Custom tool",
				Parameters:  map[string]interface{}{"type": "object"},
			},
			{Preset: "read_file"},
		}

		tools, err := ResolveAgentTools(defs)
		require.NoError(t, err)
		assert.Len(t, tools, 3)
		assert.Equal(t, "bash", tools[0].Function.Name)
		assert.Equal(t, "custom", tools[1].Function.Name)
		assert.Equal(t, "read_file", tools[2].Function.Name)
	})

	t.Run("unknown preset fails", func(t *testing.T) {
		defs := []AgentToolDef{
			{Preset: "nonexistent_tool"},
		}

		_, err := ResolveAgentTools(defs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown preset tool")
	})

	t.Run("custom tool without name fails", func(t *testing.T) {
		defs := []AgentToolDef{
			{Description: "no name tool"},
		}

		_, err := ResolveAgentTools(defs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires 'name' field")
	})

	t.Run("empty list", func(t *testing.T) {
		tools, err := ResolveAgentTools([]AgentToolDef{})
		require.NoError(t, err)
		assert.Empty(t, tools)
	})
}

func TestBuildSpawnAgentTool(t *testing.T) {
	subAgents := []SubAgentDef{
		{Name: "recon_agent", Description: "Recon specialist"},
		{Name: "vuln_scanner", Description: "Vulnerability scanner"},
	}

	tool := BuildSpawnAgentTool(subAgents)

	assert.Equal(t, "function", tool.Type)
	assert.Equal(t, SpawnAgentToolName, tool.Function.Name)
	assert.Contains(t, tool.Function.Description, "recon_agent")
	assert.Contains(t, tool.Function.Description, "vuln_scanner")
	assert.Contains(t, tool.Function.Description, "Recon specialist")

	// Verify parameters schema
	params := tool.Function.Parameters
	assert.Equal(t, "object", params["type"])

	props, ok := params["properties"].(map[string]interface{})
	require.True(t, ok)

	// Verify agent param has enum
	agentParam, ok := props["agent"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", agentParam["type"])
	enumVals, ok := agentParam["enum"].([]interface{})
	require.True(t, ok)
	assert.Len(t, enumVals, 2)
	assert.Equal(t, "recon_agent", enumVals[0])
	assert.Equal(t, "vuln_scanner", enumVals[1])

	// Verify query param
	queryParam, ok := props["query"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "string", queryParam["type"])

	// Verify required
	required, ok := params["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "agent")
	assert.Contains(t, required, "query")
}

func TestResolveAgentToolsWithSubAgents(t *testing.T) {
	t.Run("no sub-agents — same as ResolveAgentTools", func(t *testing.T) {
		defs := []AgentToolDef{{Preset: "bash"}}
		tools, err := ResolveAgentToolsWithSubAgents(defs, nil)
		require.NoError(t, err)
		assert.Len(t, tools, 1)
		assert.Equal(t, "bash", tools[0].Function.Name)
	})

	t.Run("with sub-agents — spawn_agent appended", func(t *testing.T) {
		defs := []AgentToolDef{{Preset: "bash"}, {Preset: "read_file"}}
		subAgents := []SubAgentDef{
			{Name: "recon", Description: "Recon agent"},
		}

		tools, err := ResolveAgentToolsWithSubAgents(defs, subAgents)
		require.NoError(t, err)
		assert.Len(t, tools, 3) // bash + read_file + spawn_agent

		assert.Equal(t, "bash", tools[0].Function.Name)
		assert.Equal(t, "read_file", tools[1].Function.Name)
		assert.Equal(t, SpawnAgentToolName, tools[2].Function.Name)
	})

	t.Run("error propagated from ResolveAgentTools", func(t *testing.T) {
		defs := []AgentToolDef{{Preset: "nonexistent"}}
		_, err := ResolveAgentToolsWithSubAgents(defs, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown preset tool")
	})
}
