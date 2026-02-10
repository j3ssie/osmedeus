package core

// AgentToolDef defines a tool available to the agent.
// Supports two styles:
//   - Preset: references a built-in osmedeus function by name (schema auto-generated)
//   - Custom: explicit name, description, parameters schema, and handler expression
type AgentToolDef struct {
	// Preset tool — name of an osmedeus function (schema auto-generated from PresetToolRegistry)
	Preset string `yaml:"preset,omitempty" json:"preset,omitempty"`

	// Custom tool fields
	Name        string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Parameters  map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`

	// Handler is a JS expression executed when the tool is called.
	// The parsed arguments are available as the `args` object.
	Handler string `yaml:"handler,omitempty" json:"handler,omitempty"`
}

// IsPreset returns true if this is a preset tool definition
func (t *AgentToolDef) IsPreset() bool {
	return t.Preset != ""
}

// AgentMemoryConfig configures conversation memory for agent steps
type AgentMemoryConfig struct {
	// MaxMessages is the sliding window size for conversation messages.
	// When exceeded, oldest non-system messages are dropped.
	// 0 = unlimited (keep all messages in context).
	MaxMessages int `yaml:"max_messages,omitempty" json:"max_messages,omitempty"`

	// SummarizeOnTruncate enables LLM-based summarization of dropped messages
	// instead of silently discarding them when the sliding window is exceeded.
	SummarizeOnTruncate bool `yaml:"summarize_on_truncate,omitempty" json:"summarize_on_truncate,omitempty"`

	// PersistPath is the file path to save the conversation JSON after completion.
	PersistPath string `yaml:"persist_path,omitempty" json:"persist_path,omitempty"`

	// ResumePath is the file path to load a prior conversation from on start.
	ResumePath string `yaml:"resume_path,omitempty" json:"resume_path,omitempty"`
}

// DefaultMaxAgentDepth is the maximum nesting depth for sub-agent spawning.
const DefaultMaxAgentDepth = 3

// SubAgentDef defines an inline sub-agent that can be spawned by a parent agent.
type SubAgentDef struct {
	Name          string             `yaml:"name" json:"name"`
	Description   string             `yaml:"description,omitempty" json:"description,omitempty"`
	SystemPrompt  string             `yaml:"system_prompt,omitempty" json:"system_prompt,omitempty"`
	AgentTools    []AgentToolDef     `yaml:"agent_tools,omitempty" json:"agent_tools,omitempty"`
	MaxIterations int                `yaml:"max_iterations,omitempty" json:"max_iterations,omitempty"`
	Models        []string           `yaml:"models,omitempty" json:"models,omitempty"`
	LLMConfig     *LLMStepConfig     `yaml:"llm_config,omitempty" json:"llm_config,omitempty"`
	OutputSchema  string             `yaml:"output_schema,omitempty" json:"output_schema,omitempty"`
	Memory        *AgentMemoryConfig `yaml:"memory,omitempty" json:"memory,omitempty"`
	StopCondition string             `yaml:"stop_condition,omitempty" json:"stop_condition,omitempty"`
	SubAgents     []SubAgentDef      `yaml:"sub_agents,omitempty" json:"sub_agents,omitempty"` // Recursive nesting
	OnToolStart   string             `yaml:"on_tool_start,omitempty" json:"on_tool_start,omitempty"`
	OnToolEnd     string             `yaml:"on_tool_end,omitempty" json:"on_tool_end,omitempty"`
}

// DeepCopy creates an independent deep copy of the SubAgentDef.
func (s *SubAgentDef) DeepCopy() SubAgentDef {
	cp := *s

	if len(s.AgentTools) > 0 {
		cp.AgentTools = make([]AgentToolDef, len(s.AgentTools))
		copy(cp.AgentTools, s.AgentTools)
	}
	if len(s.Models) > 0 {
		cp.Models = make([]string, len(s.Models))
		copy(cp.Models, s.Models)
	}
	if s.LLMConfig != nil {
		cfg := *s.LLMConfig
		cp.LLMConfig = &cfg
	}
	if s.Memory != nil {
		mem := *s.Memory
		cp.Memory = &mem
	}
	if len(s.SubAgents) > 0 {
		cp.SubAgents = make([]SubAgentDef, len(s.SubAgents))
		for i, sa := range s.SubAgents {
			cp.SubAgents[i] = sa.DeepCopy()
		}
	}

	return cp
}
