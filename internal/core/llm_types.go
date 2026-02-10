package core

import (
	"fmt"

	"github.com/j3ssie/osmedeus/v5/internal/json"
)

// ParseOutputSchema converts a JSON string into an LLMResponseFormat suitable
// for the OpenAI response_format parameter. The schemaJSON should be a valid
// JSON schema object, e.g. '{"type":"object","properties":{...}}'.
func ParseOutputSchema(schemaJSON string) (*LLMResponseFormat, error) {
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return nil, fmt.Errorf("invalid output_schema JSON: %w", err)
	}
	return &LLMResponseFormat{
		Type: "json_schema",
		JSONSchema: map[string]interface{}{
			"name":   "output_schema",
			"schema": schema,
		},
	}, nil
}

// LLMMessageRole represents the role of a message sender
type LLMMessageRole string

const (
	LLMRoleSystem    LLMMessageRole = "system"
	LLMRoleUser      LLMMessageRole = "user"
	LLMRoleAssistant LLMMessageRole = "assistant"
	LLMRoleTool      LLMMessageRole = "tool"
)

// LLMContentType represents the type of content in a message part
type LLMContentType string

const (
	LLMContentTypeText     LLMContentType = "text"
	LLMContentTypeImageURL LLMContentType = "image_url"
)

// LLMImageURL represents an image URL with optional detail level
type LLMImageURL struct {
	URL    string `yaml:"url" json:"url"`
	Detail string `yaml:"detail,omitempty" json:"detail,omitempty"` // "low", "high", "auto"
}

// LLMContentPart represents a single content part (text or image)
type LLMContentPart struct {
	Type     LLMContentType `yaml:"type" json:"type"`
	Text     string         `yaml:"text,omitempty" json:"text,omitempty"`
	ImageURL *LLMImageURL   `yaml:"image_url,omitempty" json:"image_url,omitempty"`
}

// LLMMessage represents a single message in the conversation
// Content can be either a string or []LLMContentPart for multimodal
type LLMMessage struct {
	Role       LLMMessageRole `yaml:"role" json:"role"`
	Content    interface{}    `yaml:"content" json:"content"` // string or []LLMContentPart
	Name       string         `yaml:"name,omitempty" json:"name,omitempty"`
	ToolCallID string         `yaml:"tool_call_id,omitempty" json:"tool_call_id,omitempty"`
	ToolCalls  []LLMToolCall  `yaml:"tool_calls,omitempty" json:"tool_calls,omitempty"`
}

// LLMToolFunction defines a function that the LLM can call
type LLMToolFunction struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Parameters  map[string]interface{} `yaml:"parameters,omitempty" json:"parameters,omitempty"`
}

// LLMTool represents a tool available to the LLM
type LLMTool struct {
	Type     string          `yaml:"type" json:"type"` // "function"
	Function LLMToolFunction `yaml:"function" json:"function"`
}

// LLMToolCallFunction represents the function details in a tool call
type LLMToolCallFunction struct {
	Name      string `yaml:"name" json:"name"`
	Arguments string `yaml:"arguments" json:"arguments"` // JSON string
}

// LLMToolCall represents a tool call made by the LLM in a response
type LLMToolCall struct {
	ID       string              `yaml:"id" json:"id"`
	Type     string              `yaml:"type" json:"type"` // "function"
	Function LLMToolCallFunction `yaml:"function" json:"function"`
}

// LLMResponseFormat specifies the output format
type LLMResponseFormat struct {
	Type       string                 `yaml:"type" json:"type"` // "text", "json_object", "json_schema"
	JSONSchema map[string]interface{} `yaml:"json_schema,omitempty" json:"json_schema,omitempty"`
}

// LLMStepConfig holds step-level LLM configuration overrides
// These override the global llm_config settings
type LLMStepConfig struct {
	// Provider override (use specific provider by name instead of rotation)
	Provider string `yaml:"provider,omitempty"`

	// Model override
	Model string `yaml:"model,omitempty"`

	// Generation parameters (using pointers to distinguish between unset and zero values)
	MaxTokens   *int     `yaml:"max_tokens,omitempty"`
	Temperature *float64 `yaml:"temperature,omitempty"`
	TopK        *int     `yaml:"top_k,omitempty"`
	TopP        *float64 `yaml:"top_p,omitempty"`
	N           *int     `yaml:"n,omitempty"`

	// Request settings
	Timeout    string `yaml:"timeout,omitempty"`
	MaxRetries *int   `yaml:"max_retries,omitempty"`
	Stream     *bool  `yaml:"stream,omitempty"`

	// Response format
	ResponseFormat *LLMResponseFormat `yaml:"response_format,omitempty"`

	// Custom headers (merged with global)
	CustomHeaders map[string]string `yaml:"custom_headers,omitempty"`
}
