package parser

import (
	"fmt"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/utils"
)

// DependencyChecker validates workflow dependencies
type DependencyChecker struct{}

// NewDependencyChecker creates a new dependency checker
func NewDependencyChecker() *DependencyChecker {
	return &DependencyChecker{}
}

// CheckCommands checks if required commands are available in binariesPath or system PATH
func (c *DependencyChecker) CheckCommands(commands []string, binariesPath string) error {
	var missing []string
	for _, cmd := range commands {
		if _, err := utils.LookPathWithBinaries(cmd, binariesPath); err != nil {
			missing = append(missing, cmd)
		}
	}

	if len(missing) > 0 {
		return &DependencyError{
			Type:    "commands",
			Missing: missing,
		}
	}
	return nil
}

// CheckFiles checks if required files exist
func (c *DependencyChecker) CheckFiles(files []string, ctx map[string]interface{}) error {
	// Files would need template rendering first
	// This is a placeholder for the actual implementation
	return nil
}

// CheckVariables validates required variables are present and valid
func (c *DependencyChecker) CheckVariables(deps []core.VariableDep, ctx map[string]interface{}) error {
	var errors []string

	for _, dep := range deps {
		if !dep.Required {
			continue
		}

		value, ok := ctx[dep.Name]
		if !ok || value == nil || value == "" {
			errors = append(errors, fmt.Sprintf("required variable '%s' is missing", dep.Name))
			continue
		}

		// Validate type
		if err := c.validateType(dep.Name, value, dep.Type); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return &DependencyError{
			Type:   "variables",
			Errors: errors,
		}
	}
	return nil
}

// validateType validates a value against a variable type.
// Supports comma-separated types (e.g., "domain,url") where matching any type is sufficient.
func (c *DependencyChecker) validateType(name string, value interface{}, varType core.VariableType) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("variable '%s' must be a string", name)
	}

	// Use MatchesAnyVariableType to support comma-separated types
	ok, err := core.MatchesAnyVariableType(str, varType)
	if err != nil {
		return fmt.Errorf("variable '%s' has invalid type: %w", name, err)
	}
	if !ok {
		return fmt.Errorf("variable '%s' does not match type '%s'", name, varType)
	}
	return nil
}

// CheckAll performs all dependency checks
// Note: FunctionsConditions are checked at runtime by the executor, not here
func (c *DependencyChecker) CheckAll(deps *core.Dependencies, ctx map[string]interface{}, binariesPath string) error {
	if deps == nil {
		return nil
	}

	if len(deps.Commands) > 0 {
		if err := c.CheckCommands(deps.Commands, binariesPath); err != nil {
			return err
		}
	}

	if len(deps.Files) > 0 {
		if err := c.CheckFiles(deps.Files, ctx); err != nil {
			return err
		}
	}

	if len(deps.Variables) > 0 {
		if err := c.CheckVariables(deps.Variables, ctx); err != nil {
			return err
		}
	}

	return nil
}

// DependencyError represents a dependency check failure
type DependencyError struct {
	Type    string
	Missing []string
	Errors  []string
}

func (e *DependencyError) Error() string {
	if len(e.Missing) > 0 {
		return fmt.Sprintf("missing %s: %v", e.Type, e.Missing)
	}
	return fmt.Sprintf("%s validation errors: %v", e.Type, e.Errors)
}

// isValidDomain checks if a string is a valid domain
// DefaultDependencyChecker is the global dependency checker
var DefaultDependencyChecker = NewDependencyChecker()

// CheckDependencies checks all dependencies using the default checker
func CheckDependencies(deps *core.Dependencies, ctx map[string]interface{}, binariesPath string) error {
	return DefaultDependencyChecker.CheckAll(deps, ctx, binariesPath)
}
