package core

import "fmt"

// Param represents a workflow parameter
type Param struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`    // "string", "bool", "int" (default: "string")
	Default   any    `yaml:"default"` // Supports string, bool, int from YAML
	Required  bool   `yaml:"required"`
	Generator string `yaml:"generator"` // e.g., uuid(), currentDate(), getEnvVar("KEY")
}

// HasDefault returns true if the parameter has a default value
func (p *Param) HasDefault() bool {
	return p.Default != nil && p.DefaultString() != ""
}

// HasGenerator returns true if the parameter has a generator function
func (p *Param) HasGenerator() bool {
	return p.Generator != ""
}

// IsRequired returns true if the parameter is required
func (p *Param) IsRequired() bool {
	return p.Required
}

// IsBool returns true if the parameter type is bool
func (p *Param) IsBool() bool {
	return p.Type == "bool"
}

// IsInt returns true if the parameter type is int
func (p *Param) IsInt() bool {
	return p.Type == "int"
}

// DefaultString returns the default value as a string
func (p *Param) DefaultString() string {
	if p.Default == nil {
		return ""
	}
	return fmt.Sprintf("%v", p.Default)
}

// DefaultBool returns the default value as a bool
func (p *Param) DefaultBool() bool {
	if p.Default == nil {
		return false
	}
	switch v := p.Default.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	default:
		return false
	}
}
