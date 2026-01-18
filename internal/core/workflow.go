package core

import "strings"

// TagList is a comma-separated list of tags that parses to []string
type TagList []string

// UnmarshalYAML implements custom YAML unmarshaling for comma-separated tags
func (t *TagList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if s == "" {
		*t = []string{}
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	*t = parts
	return nil
}

// Workflow represents either a Module or Flow
type Workflow struct {
	Kind         WorkflowKind  `yaml:"kind"`
	Name         string        `yaml:"name"`
	Description  string        `yaml:"description"`
	Tags         TagList       `yaml:"tags,omitempty"`
	Params       []Param       `yaml:"params"`
	Triggers     []Trigger     `yaml:"trigger"`
	Dependencies *Dependencies `yaml:"dependencies"`
	Reports      []Report      `yaml:"reports"`

	// Execution preferences (optional, can be overridden by CLI flags)
	Preferences *Preferences `yaml:"preferences,omitempty"`

	// Runner configuration (module-kind only)
	Runner       RunnerType    `yaml:"runner,omitempty"`
	RunnerConfig *RunnerConfig `yaml:"runner_config,omitempty"`

	// Module-specific fields
	Steps []Step `yaml:"steps,omitempty"`

	// Flow-specific fields
	Modules []ModuleRef `yaml:"modules,omitempty"`

	// Internal metadata
	FilePath string `yaml:"-"`
	Checksum string `yaml:"-"`
}

// RunnerConfig holds configuration for different runner types
type RunnerConfig struct {
	// Docker configuration
	Image      string            `yaml:"image,omitempty"`      // Docker image e.g., "ubuntu:latest"
	Env        map[string]string `yaml:"env,omitempty"`        // Environment variables
	Volumes    []string          `yaml:"volumes,omitempty"`    // Volume mounts e.g., "/host:/container"
	Network    string            `yaml:"network,omitempty"`    // Network mode e.g., "host", "bridge"
	Persistent bool              `yaml:"persistent,omitempty"` // true=reuse container, false=ephemeral

	// SSH configuration
	Host     string `yaml:"host,omitempty"`     // SSH hostname or IP
	Port     int    `yaml:"port,omitempty"`     // SSH port (default 22)
	User     string `yaml:"user,omitempty"`     // SSH username
	KeyFile  string `yaml:"key_file,omitempty"` // Path to SSH private key
	Password string `yaml:"password,omitempty"` // SSH password (prefer key_file)

	// Common configuration
	WorkDir string `yaml:"workdir,omitempty"` // Working directory on remote/container
}

// ModuleRef references a module in a flow
type ModuleRef struct {
	Name      string            `yaml:"name"`
	Path      string            `yaml:"path"`
	Params    map[string]string `yaml:"params"`
	DependsOn []string          `yaml:"depends_on"`
	Condition string            `yaml:"condition"`
	OnSuccess []Action          `yaml:"on_success"`
	OnError   []Action          `yaml:"on_error"`
	Decision  *DecisionConfig   `yaml:"decision"`
}

// IsModule returns true if the workflow is a module
func (w *Workflow) IsModule() bool {
	return w.Kind == KindModule
}

// IsFlow returns true if the workflow is a flow
func (w *Workflow) IsFlow() bool {
	return w.Kind == KindFlow
}

// GetRequiredParams returns all required parameters
func (w *Workflow) GetRequiredParams() []Param {
	var required []Param
	for _, p := range w.Params {
		if p.Required {
			required = append(required, p)
		}
	}
	return required
}

// HasTriggers returns true if the workflow has any triggers defined
func (w *Workflow) HasTriggers() bool {
	return len(w.Triggers) > 0
}

// IsManualExecutionAllowed checks if manual (CLI) execution is allowed
// Returns true if:
// - No triggers are defined (default behavior allows manual)
// - A manual trigger exists and is enabled
// - No manual trigger is explicitly defined (default is enabled)
func (w *Workflow) IsManualExecutionAllowed() bool {
	// If no triggers defined, manual is allowed (default behavior)
	if len(w.Triggers) == 0 {
		return true
	}

	// Look for explicit manual trigger
	for _, t := range w.Triggers {
		if t.On == TriggerManual {
			return t.Enabled // Use explicit setting
		}
	}

	// No manual trigger defined among other triggers, default to true
	return true
}

// GetEventTriggers returns all event-type triggers
func (w *Workflow) GetEventTriggers() []Trigger {
	var triggers []Trigger
	for _, t := range w.Triggers {
		if t.On == TriggerEvent && t.Enabled {
			triggers = append(triggers, t)
		}
	}
	return triggers
}
