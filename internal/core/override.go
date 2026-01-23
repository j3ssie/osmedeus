package core

import "fmt"

// OverrideMode specifies how to merge steps or modules during inheritance
type OverrideMode string

const (
	// OverrideModeReplace completely replaces parent items with child items
	OverrideModeReplace OverrideMode = "replace"
	// OverrideModePrepend adds child items before parent items
	OverrideModePrepend OverrideMode = "prepend"
	// OverrideModeAppend adds child items after parent items (default)
	OverrideModeAppend OverrideMode = "append"
	// OverrideModeMerge matches items by name: replaces matching, appends new, removes specified
	OverrideModeMerge OverrideMode = "merge"
)

// WorkflowOverride contains all override specifications for workflow inheritance
type WorkflowOverride struct {
	// Params override specific parameter properties
	Params map[string]*ParamOverride `yaml:"params,omitempty"`

	// Steps override for module workflows
	Steps *StepsOverride `yaml:"steps,omitempty"`

	// Modules override for flow workflows
	Modules *ModulesOverride `yaml:"modules,omitempty"`

	// Triggers override - replaces parent triggers entirely if set
	Triggers []Trigger `yaml:"triggers,omitempty"`

	// Dependencies override - merged with parent dependencies
	Dependencies *Dependencies `yaml:"dependencies,omitempty"`

	// Preferences override - child values override parent values
	Preferences *Preferences `yaml:"preferences,omitempty"`

	// RunnerConfig override - child values override parent values
	RunnerConfig *RunnerConfig `yaml:"runner_config,omitempty"`

	// Runner type override
	Runner *RunnerType `yaml:"runner,omitempty"`
}

// ParamOverride allows overriding specific properties of a parameter
type ParamOverride struct {
	// Default overrides the default value
	Default any `yaml:"default,omitempty"`

	// Type overrides the parameter type
	Type *string `yaml:"type,omitempty"`

	// Required overrides whether the parameter is required
	Required *bool `yaml:"required,omitempty"`

	// Generator overrides the generator function
	Generator *string `yaml:"generator,omitempty"`
}

// UnmarshalYAML supports shorthand syntax where a scalar value is treated as the default:
//
//	param-a: "value"     # shorthand: scalar becomes default
//	param-b: false       # shorthand: bool becomes default
//
// And verbose syntax with full struct:
//
//	param-a:
//	  default: "value"
//	  type: "string"
//	  required: true
func (p *ParamOverride) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try scalar first (shorthand syntax)
	var scalar interface{}
	if err := unmarshal(&scalar); err == nil {
		switch v := scalar.(type) {
		case string, bool, float64:
			p.Default = v
			return nil
		case int:
			p.Default = v
			return nil
		case int64:
			p.Default = int(v)
			return nil
		case uint64:
			// goccy/go-yaml returns uint64 for positive integers
			p.Default = int(v)
			return nil
		case nil:
			// nil value is valid, means no default
			return nil
		case map[string]interface{}:
			// Fall through to map handling below
		default:
			// Unknown type, try map handling
		}
	}

	// Try as map (verbose syntax)
	var m map[string]interface{}
	if err := unmarshal(&m); err != nil {
		return fmt.Errorf("invalid param override: expected scalar or map")
	}

	if v, ok := m["default"]; ok {
		// Normalize integer types
		switch val := v.(type) {
		case int64:
			p.Default = int(val)
		case uint64:
			p.Default = int(val)
		default:
			p.Default = v
		}
	}
	if v, ok := m["type"].(string); ok {
		p.Type = &v
	}
	if v, ok := m["required"].(bool); ok {
		p.Required = &v
	}
	if v, ok := m["generator"].(string); ok {
		p.Generator = &v
	}
	return nil
}

// StepsOverride specifies how to override steps in a module workflow
type StepsOverride struct {
	// Mode specifies the merge strategy: replace, prepend, append, merge
	// Default is "append"
	Mode OverrideMode `yaml:"mode,omitempty"`

	// Steps to add (used with prepend, append modes)
	// or to match and replace (used with merge mode)
	Steps []Step `yaml:"steps,omitempty"`

	// Remove lists step names to remove (only used with merge mode)
	Remove []string `yaml:"remove,omitempty"`

	// Replace lists steps that replace existing steps by name (only used with merge mode)
	Replace []Step `yaml:"replace,omitempty"`
}

// ModulesOverride specifies how to override modules in a flow workflow
type ModulesOverride struct {
	// Mode specifies the merge strategy: replace, prepend, append, merge
	// Default is "append"
	Mode OverrideMode `yaml:"mode,omitempty"`

	// Modules to add (used with prepend, append modes)
	// or to match and replace (used with merge mode)
	Modules []ModuleRef `yaml:"modules,omitempty"`

	// Remove lists module names to remove (only used with merge mode)
	Remove []string `yaml:"remove,omitempty"`

	// Replace lists modules that replace existing modules by name (only used with merge mode)
	Replace []ModuleRef `yaml:"replace,omitempty"`
}

// IsValidOverrideMode checks if the mode is a valid override mode
func IsValidOverrideMode(mode OverrideMode) bool {
	switch mode {
	case OverrideModeReplace, OverrideModePrepend, OverrideModeAppend, OverrideModeMerge, "":
		return true
	default:
		return false
	}
}

// GetEffectiveMode returns the effective mode, defaulting to append if empty
func (s *StepsOverride) GetEffectiveMode() OverrideMode {
	if s.Mode == "" {
		return OverrideModeAppend
	}
	return s.Mode
}

// GetEffectiveMode returns the effective mode, defaulting to append if empty
func (m *ModulesOverride) GetEffectiveMode() OverrideMode {
	if m.Mode == "" {
		return OverrideModeAppend
	}
	return m.Mode
}
