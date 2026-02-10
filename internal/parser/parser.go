package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// Parser handles workflow YAML parsing
type Parser struct{}

// NewParser creates a new workflow parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses a workflow from a file path
func (p *Parser) Parse(path string) (*core.Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	workflow, err := p.ParseContent(data)
	if err != nil {
		return nil, err
	}

	// Set metadata
	workflow.FilePath = path
	workflow.Checksum = p.calculateChecksum(data)

	return workflow, nil
}

// ParseContent parses workflow content from bytes
func (p *Parser) ParseContent(content []byte) (*core.Workflow, error) {
	var workflow core.Workflow
	if err := yaml.Unmarshal(content, &workflow); err != nil {
		// Use go-yaml's built-in formatter for detailed error with line numbers
		// FormatError(err, colored, includeSource) returns formatted error with position info
		formatted := yaml.FormatError(err, false, true)
		return nil, fmt.Errorf("YAML parse error:\n%s", formatted)
	}

	return &workflow, nil
}

// Validate validates a parsed workflow
func (p *Parser) Validate(w *core.Workflow) error {
	// Validate kind
	if w.Kind != core.KindModule && w.Kind != core.KindFlow {
		return &ValidationError{
			Field:   "kind",
			Message: fmt.Sprintf("invalid kind: %s, must be 'module' or 'flow'", w.Kind),
		}
	}

	// Validate name
	if w.Name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "workflow name is required",
		}
	}

	// Validate preferences if present
	if err := p.validatePreferences(w.Preferences); err != nil {
		return err
	}

	// Validate override if present
	if err := p.validateOverride(w); err != nil {
		return err
	}

	// Validate based on kind
	if w.IsModule() {
		return p.validateModule(w)
	}
	return p.validateFlow(w)
}

// validateModule validates module-specific fields
func (p *Parser) validateModule(w *core.Workflow) error {
	// Modules must have at least one step, unless they extend another workflow
	// (in which case they will inherit steps after resolution)
	if len(w.Steps) == 0 && w.Extends == "" {
		return &ValidationError{
			Field:   "steps",
			Message: "module must have at least one step",
		}
	}

	// Validate each step
	for i, step := range w.Steps {
		if err := p.validateStep(&step, i); err != nil {
			return err
		}
	}

	return nil
}

// validateFlow validates flow-specific fields
func (p *Parser) validateFlow(w *core.Workflow) error {
	// Flows must have at least one module, unless they extend another workflow
	// (in which case they will inherit modules after resolution)
	if len(w.Modules) == 0 && w.Extends == "" {
		return &ValidationError{
			Field:   "modules",
			Message: "flow must have at least one module reference",
		}
	}

	// Validate each module reference
	for i, mod := range w.Modules {
		if mod.Name == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("modules[%d].name", i),
				Message: "module reference name is required",
			}
		}
		// Path is required only for external modules (not inline modules)
		if mod.Path == "" && !mod.IsInline() {
			return &ValidationError{
				Field:   fmt.Sprintf("modules[%d].path", i),
				Message: "module reference path is required (or define inline steps)",
			}
		}
		// Inline modules must have at least one step
		if mod.IsInline() && len(mod.Steps) == 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("modules[%d].steps", i),
				Message: "inline module must have at least one step",
			}
		}
	}

	return nil
}

// validatePreferences validates workflow preferences
func (p *Parser) validatePreferences(prefs *core.Preferences) error {
	if prefs == nil {
		return nil
	}

	// Validate heuristics_check value if set
	if prefs.HeuristicsCheck != nil {
		validValues := map[string]bool{"none": true, "basic": true, "advanced": true}
		if !validValues[*prefs.HeuristicsCheck] {
			return &ValidationError{
				Field:   "preferences.heuristics_check",
				Message: fmt.Sprintf("invalid value: %s, must be 'none', 'basic', or 'advanced'", *prefs.HeuristicsCheck),
			}
		}
	}

	return nil
}

// validateOverride validates the override section of a workflow
func (p *Parser) validateOverride(w *core.Workflow) error {
	if w.Override == nil {
		return nil
	}

	// Validate steps override mode
	if w.Override.Steps != nil {
		if !core.IsValidOverrideMode(w.Override.Steps.Mode) {
			return &ValidationError{
				Field:   "override.steps.mode",
				Message: fmt.Sprintf("invalid mode: %s, must be 'replace', 'prepend', 'append', or 'merge'", w.Override.Steps.Mode),
			}
		}

		// Validate that steps override is only used with module workflows
		if w.IsFlow() {
			return &ValidationError{
				Field:   "override.steps",
				Message: "steps override can only be used with module workflows",
			}
		}

		// Validate steps in override
		for i, step := range w.Override.Steps.Steps {
			if err := p.validateStep(&step, i); err != nil {
				return &ValidationError{
					Field:   fmt.Sprintf("override.steps.steps[%d]", i),
					Message: err.Error(),
				}
			}
		}

		// Validate replacement steps
		for i, step := range w.Override.Steps.Replace {
			if err := p.validateStep(&step, i); err != nil {
				return &ValidationError{
					Field:   fmt.Sprintf("override.steps.replace[%d]", i),
					Message: err.Error(),
				}
			}
		}
	}

	// Validate modules override mode
	if w.Override.Modules != nil {
		if !core.IsValidOverrideMode(w.Override.Modules.Mode) {
			return &ValidationError{
				Field:   "override.modules.mode",
				Message: fmt.Sprintf("invalid mode: %s, must be 'replace', 'prepend', 'append', or 'merge'", w.Override.Modules.Mode),
			}
		}

		// Validate that modules override is only used with flow workflows
		if w.IsModule() {
			return &ValidationError{
				Field:   "override.modules",
				Message: "modules override can only be used with flow workflows",
			}
		}
	}

	// Validate preferences in override
	if w.Override.Preferences != nil {
		if err := p.validatePreferences(w.Override.Preferences); err != nil {
			return &ValidationError{
				Field:   "override.preferences",
				Message: err.Error(),
			}
		}
	}

	return nil
}

// validateStep validates a step definition
func (p *Parser) validateStep(step *core.Step, index int) error {
	if step.Name == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("steps[%d].name", index),
			Message: "step name is required",
		}
	}

	if step.Type == "" {
		return &ValidationError{
			Field:   fmt.Sprintf("steps[%d].type", index),
			Message: "step type is required",
		}
	}

	// Validate step type
	switch step.Type {
	case core.StepTypeBash:
		if step.Command == "" && len(step.Commands) == 0 && len(step.ParallelCommands) == 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d]", index),
				Message: "bash step must have command, commands, or parallel_commands",
			}
		}
	case core.StepTypeFunction:
		if step.Function == "" && len(step.Functions) == 0 && len(step.ParallelFunctions) == 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d]", index),
				Message: "function step must have function, functions, or parallel_functions",
			}
		}
	case core.StepTypeParallel:
		if len(step.ParallelSteps) == 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d]", index),
				Message: "parallel step must have parallel_steps",
			}
		}
	case core.StepTypeForeach:
		if step.Input == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].input", index),
				Message: "foreach step must have input",
			}
		}
		if step.Variable == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].variable", index),
				Message: "foreach step must have variable name",
			}
		}
		if step.Step == nil {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].step", index),
				Message: "foreach step must have inner step",
			}
		}
	case core.StepTypeRemoteBash:
		// Validate step_runner is set and valid
		if step.StepRunner == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].step_runner", index),
				Message: "remote-bash step must have step_runner set to 'docker' or 'ssh'",
			}
		}
		if step.StepRunner != core.RunnerTypeDocker && step.StepRunner != core.RunnerTypeSSH {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].step_runner", index),
				Message: fmt.Sprintf("invalid step_runner: %s (must be 'docker' or 'ssh')", step.StepRunner),
			}
		}
		// Validate has command
		if step.Command == "" && len(step.Commands) == 0 && len(step.ParallelCommands) == 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d]", index),
				Message: "remote-bash step must have command, commands, or parallel_commands",
			}
		}
	case core.StepTypeHTTP:
		// Validate HTTP step has URL
		if step.URL == "" {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].url", index),
				Message: "http step must have url",
			}
		}
	case core.StepTypeLLM:
		// Validate LLM step has messages or embedding_input
		if len(step.Messages) == 0 && len(step.EmbeddingInput) == 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d]", index),
				Message: "llm step must have messages or embedding_input",
			}
		}
		// Validate embedding step has is_embedding flag when using embedding_input
		if len(step.EmbeddingInput) > 0 && !step.IsEmbedding {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].is_embedding", index),
				Message: "llm step with embedding_input should have is_embedding: true",
			}
		}
		// Validate message roles if messages are provided
		for i, msg := range step.Messages {
			if msg.Role == "" {
				return &ValidationError{
					Field:   fmt.Sprintf("steps[%d].messages[%d].role", index, i),
					Message: "message role is required",
				}
			}
			validRoles := map[core.LLMMessageRole]bool{
				core.LLMRoleSystem:    true,
				core.LLMRoleUser:      true,
				core.LLMRoleAssistant: true,
				core.LLMRoleTool:      true,
			}
			if !validRoles[msg.Role] {
				return &ValidationError{
					Field:   fmt.Sprintf("steps[%d].messages[%d].role", index, i),
					Message: fmt.Sprintf("invalid message role: %s (must be system, user, assistant, or tool)", msg.Role),
				}
			}
		}
	case core.StepTypeAgent:
		// Validate agent step has query or queries (not both)
		if step.Query == "" && len(step.Queries) == 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].query", index),
				Message: "agent step must have 'query' or 'queries'",
			}
		}
		if step.Query != "" && len(step.Queries) > 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d]", index),
				Message: "agent step cannot have both 'query' and 'queries'",
			}
		}
		// Validate queries are not empty strings
		for i, q := range step.Queries {
			if q == "" {
				return &ValidationError{
					Field:   fmt.Sprintf("steps[%d].queries[%d]", index, i),
					Message: "query in queries list must not be empty",
				}
			}
		}
		// Validate agent step has max_iterations
		if step.MaxIterations <= 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].max_iterations", index),
				Message: "agent step must have max_iterations > 0",
			}
		}
		// Validate agent step has agent_tools
		if len(step.AgentTools) == 0 {
			return &ValidationError{
				Field:   fmt.Sprintf("steps[%d].agent_tools", index),
				Message: "agent step must have at least one agent_tool",
			}
		}
		// Validate individual agent tools
		seen := make(map[string]bool)
		for i, tool := range step.AgentTools {
			if tool.IsPreset() {
				// Validate preset tool exists
				if _, ok := core.GetPresetTool(tool.Preset); !ok {
					return &ValidationError{
						Field:   fmt.Sprintf("steps[%d].agent_tools[%d].preset", index, i),
						Message: fmt.Sprintf("unknown preset tool: %s", tool.Preset),
					}
				}
				if seen[tool.Preset] {
					return &ValidationError{
						Field:   fmt.Sprintf("steps[%d].agent_tools[%d].preset", index, i),
						Message: fmt.Sprintf("duplicate tool name: %s", tool.Preset),
					}
				}
				seen[tool.Preset] = true
			} else {
				// Custom tool must have name and description
				if tool.Name == "" {
					return &ValidationError{
						Field:   fmt.Sprintf("steps[%d].agent_tools[%d].name", index, i),
						Message: "custom agent tool requires 'name' field",
					}
				}
				if tool.Description == "" {
					return &ValidationError{
						Field:   fmt.Sprintf("steps[%d].agent_tools[%d].description", index, i),
						Message: "custom agent tool requires 'description' field",
					}
				}
				if seen[tool.Name] {
					return &ValidationError{
						Field:   fmt.Sprintf("steps[%d].agent_tools[%d].name", index, i),
						Message: fmt.Sprintf("duplicate tool name: %s", tool.Name),
					}
				}
				seen[tool.Name] = true
			}
		}
	default:
		return &ValidationError{
			Field:   fmt.Sprintf("steps[%d].type", index),
			Message: fmt.Sprintf("invalid step type: %s", step.Type),
		}
	}

	return nil
}

// calculateChecksum calculates SHA256 checksum of content
func (p *Parser) calculateChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

// DefaultParser is the global parser instance
var DefaultParser = NewParser()

// Parse parses a workflow file using the default parser
func Parse(path string) (*core.Workflow, error) {
	return DefaultParser.Parse(path)
}

// ParseContent parses workflow content using the default parser
func ParseContent(content []byte) (*core.Workflow, error) {
	return DefaultParser.ParseContent(content)
}

// Validate validates a workflow using the default parser
func Validate(w *core.Workflow) error {
	return DefaultParser.Validate(w)
}
