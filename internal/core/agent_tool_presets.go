package core

import (
	"fmt"
	"strings"
)

// PresetToolDef holds a preset tool's metadata for auto-generating LLMTool schemas
type PresetToolDef struct {
	Description string
	Parameters  map[string]interface{}
}

// PresetToolRegistry maps preset names to their tool definitions.
// These are used to auto-generate OpenAI-compatible function schemas
// when an agent step references a tool by preset name.
var PresetToolRegistry = map[string]PresetToolDef{
	"bash": {
		Description: "Execute a shell command and return its output",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"command": map[string]interface{}{
					"type":        "string",
					"description": "The shell command to execute",
				},
			},
			"required": []string{"command"},
		},
	},
	"read_file": {
		Description: "Read the contents of a file",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read",
				},
			},
			"required": []string{"path"},
		},
	},
	"read_lines": {
		Description: "Read a file and return its contents as an array of lines",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file to read",
				},
			},
			"required": []string{"path"},
		},
	},
	"file_exists": {
		Description: "Check if a file exists at the given path",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to check",
				},
			},
			"required": []string{"path"},
		},
	},
	"file_length": {
		Description: "Count the number of non-empty lines in a file",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the file",
				},
			},
			"required": []string{"path"},
		},
	},
	"append_file": {
		Description: "Append content from source file to destination file",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"dest": map[string]interface{}{
					"type":        "string",
					"description": "Destination file path",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Source file path to append from",
				},
			},
			"required": []string{"dest", "content"},
		},
	},
	"save_content": {
		Description: "Write string content to a file (overwrites if exists)",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"content": map[string]interface{}{
					"type":        "string",
					"description": "Content to write",
				},
				"path": map[string]interface{}{
					"type":        "string",
					"description": "File path to write to",
				},
			},
			"required": []string{"content", "path"},
		},
	},
	"glob": {
		Description: "Find files matching a glob pattern",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Glob pattern (e.g., '*.txt', '/path/**/*.json')",
				},
			},
			"required": []string{"pattern"},
		},
	},
	"grep_string": {
		Description: "Search a file for lines containing a string",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "File path to search in",
				},
				"str": map[string]interface{}{
					"type":        "string",
					"description": "String to search for",
				},
			},
			"required": []string{"source", "str"},
		},
	},
	"grep_regex": {
		Description: "Search a file for lines matching a regex pattern",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source": map[string]interface{}{
					"type":        "string",
					"description": "File path to search in",
				},
				"pattern": map[string]interface{}{
					"type":        "string",
					"description": "Regex pattern to match",
				},
			},
			"required": []string{"source", "pattern"},
		},
	},
	"http_get": {
		Description: "Make an HTTP GET request and return the response",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL to send GET request to",
				},
			},
			"required": []string{"url"},
		},
	},
	"http_request": {
		Description: "Make an HTTP request with specified method, headers, and body",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"method": map[string]interface{}{
					"type":        "string",
					"description": "HTTP method (GET, POST, PUT, DELETE, etc.)",
				},
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL to send request to",
				},
				"body": map[string]interface{}{
					"type":        "string",
					"description": "Request body (for POST/PUT)",
				},
				"headers": map[string]interface{}{
					"type":        "string",
					"description": "Headers as JSON string",
				},
			},
			"required": []string{"url", "method"},
		},
	},
	"jq": {
		Description: "Query JSON data using jq expression syntax",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"json_data": map[string]interface{}{
					"type":        "string",
					"description": "JSON string to query",
				},
				"expression": map[string]interface{}{
					"type":        "string",
					"description": "jq expression (e.g., '.name', '.items[].id')",
				},
			},
			"required": []string{"json_data", "expression"},
		},
	},
	"exec_python": {
		Description: "Run inline Python code and return stdout (prefers python3)",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "Python code to execute",
				},
			},
			"required": []string{"code"},
		},
	},
	"exec_python_file": {
		Description: "Run a Python file and return stdout (prefers uv, then python3)",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the Python file to execute",
				},
			},
			"required": []string{"path"},
		},
	},
	"exec_ts": {
		Description: "Run inline TypeScript code via bun and return stdout",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type":        "string",
					"description": "TypeScript code to execute",
				},
			},
			"required": []string{"code"},
		},
	},
	"exec_ts_file": {
		Description: "Run a TypeScript file via bun and return stdout",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "Path to the TypeScript file to execute",
				},
			},
			"required": []string{"path"},
		},
	},
	"run_module": {
		Description: "Run an osmedeus module as a subprocess",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"module": map[string]interface{}{
					"type":        "string",
					"description": "Module name to run",
				},
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Target to scan",
				},
				"params": map[string]interface{}{
					"type":        "string",
					"description": "Optional comma-separated key=value params (e.g., 'threads=10,deep=true')",
				},
			},
			"required": []string{"module", "target"},
		},
	},
	"run_flow": {
		Description: "Run an osmedeus flow as a subprocess",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"flow": map[string]interface{}{
					"type":        "string",
					"description": "Flow name to run",
				},
				"target": map[string]interface{}{
					"type":        "string",
					"description": "Target to scan",
				},
				"params": map[string]interface{}{
					"type":        "string",
					"description": "Optional comma-separated key=value params (e.g., 'threads=10,deep=true')",
				},
			},
			"required": []string{"flow", "target"},
		},
	},
}

// GetPresetTool returns the LLMTool schema for a preset tool name.
// Returns the tool and true if found, zero value and false otherwise.
func GetPresetTool(presetName string) (LLMTool, bool) {
	preset, ok := PresetToolRegistry[presetName]
	if !ok {
		return LLMTool{}, false
	}

	return LLMTool{
		Type: "function",
		Function: LLMToolFunction{
			Name:        presetName,
			Description: preset.Description,
			Parameters:  preset.Parameters,
		},
	}, true
}

// ResolveAgentTools converts a list of AgentToolDef into OpenAI-compatible LLMTool schemas.
// Preset tools are looked up from PresetToolRegistry; custom tools are converted directly.
func ResolveAgentTools(defs []AgentToolDef) ([]LLMTool, error) {
	tools := make([]LLMTool, 0, len(defs))

	for _, def := range defs {
		if def.IsPreset() {
			tool, ok := GetPresetTool(def.Preset)
			if !ok {
				return nil, fmt.Errorf("unknown preset tool: %s", def.Preset)
			}
			tools = append(tools, tool)
		} else {
			// Custom tool
			if def.Name == "" {
				return nil, fmt.Errorf("custom agent tool requires 'name' field")
			}
			tool := LLMTool{
				Type: "function",
				Function: LLMToolFunction{
					Name:        def.Name,
					Description: def.Description,
					Parameters:  def.Parameters,
				},
			}
			tools = append(tools, tool)
		}
	}

	return tools, nil
}

// SpawnAgentToolName is the tool name used for sub-agent spawning.
const SpawnAgentToolName = "spawn_agent"

// BuildSpawnAgentTool dynamically generates the spawn_agent tool schema
// with an enum of available sub-agent names and their descriptions.
func BuildSpawnAgentTool(subAgents []SubAgentDef) LLMTool {
	// Build enum of sub-agent names
	names := make([]interface{}, 0, len(subAgents))
	var descParts []string
	for _, sa := range subAgents {
		names = append(names, sa.Name)
		desc := sa.Name
		if sa.Description != "" {
			desc += ": " + sa.Description
		}
		descParts = append(descParts, desc)
	}

	description := "Spawn a sub-agent to handle a specialized task. Available agents: " + strings.Join(descParts, "; ")

	return LLMTool{
		Type: "function",
		Function: LLMToolFunction{
			Name:        SpawnAgentToolName,
			Description: description,
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent": map[string]interface{}{
						"type":        "string",
						"description": "Name of the sub-agent to spawn",
						"enum":        names,
					},
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The task or query to delegate to the sub-agent",
					},
				},
				"required": []string{"agent", "query"},
			},
		},
	}
}

// ResolveAgentToolsWithSubAgents resolves agent tools and appends the spawn_agent
// tool if sub-agents are defined.
func ResolveAgentToolsWithSubAgents(defs []AgentToolDef, subAgents []SubAgentDef) ([]LLMTool, error) {
	tools, err := ResolveAgentTools(defs)
	if err != nil {
		return nil, err
	}

	if len(subAgents) > 0 {
		tools = append(tools, BuildSpawnAgentTool(subAgents))
	}

	return tools, nil
}
