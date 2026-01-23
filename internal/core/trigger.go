package core

import (
	"path"
	"strings"
	"time"
)

// containsWildcard checks if pattern has glob wildcards
func containsWildcard(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}

// Trigger defines when a workflow should execute
type Trigger struct {
	Name     string       `yaml:"name"`
	On       TriggerType  `yaml:"on"`
	Schedule string       `yaml:"schedule,omitempty"` // cron expression (for cron triggers)
	Event    *EventConfig `yaml:"event,omitempty"`    // event configuration (for event triggers)
	Path     string       `yaml:"path,omitempty"`     // watch path (for watch triggers)
	Debounce string       `yaml:"debounce,omitempty"` // debounce duration for watch triggers (e.g., "500ms", "1s")
	Input    TriggerInput `yaml:"input,omitempty"`
	Enabled  bool         `yaml:"enabled"`

	// ScheduleID links this trigger to a database Schedule record (for API-created schedules)
	// When set, the trigger handler will look up the Schedule to get Target, Params, etc.
	ScheduleID string `yaml:"-" json:"schedule_id,omitempty"`
}

// EventConfig holds event trigger configuration
type EventConfig struct {
	Topic           string   `yaml:"topic"`                      // e.g., "webhook.received", "assets.new"
	Filters         []string `yaml:"filters,omitempty"`          // JS expressions: ["event.name == 'discovered'"]
	FilterFunctions []string `yaml:"filter_functions,omitempty"` // JS with utility functions: ["contains(event.data.url, '/api/')"]
	DedupeKey       string   `yaml:"dedupe_key,omitempty"`       // template for deduplication key (e.g., "{{event.source}}-{{event.data.url}}")
	DedupeWindow    string   `yaml:"dedupe_window,omitempty"`    // duration to ignore duplicates (e.g., "5s", "1m")
}

// TriggerInput defines the input source for trigger.
// Supports two syntaxes:
//
// Legacy syntax (single input):
//
//	input:
//	  type: event_data
//	  field: url
//	  name: target
//
// New exports-style syntax (multiple variables):
//
//	input:
//	  target: event_data.url
//	  description: trim(event_data.desc)
//	  source: event.source
type TriggerInput struct {
	// New exports-style syntax (map of variable name -> expression)
	Vars map[string]string `yaml:"-"` // Custom unmarshal handles this

	// Legacy fields (for backward compatibility)
	Type     string `yaml:"type,omitempty"`     // file, event_data, function, param
	Path     string `yaml:"path,omitempty"`     // for file type
	Field    string `yaml:"field,omitempty"`    // for event_data type
	Function string `yaml:"function,omitempty"` // for function type (e.g., jq("{{event.data}}", ".url"))
	Name     string `yaml:"name,omitempty"`     // parameter name to set
}

// HasVars returns true if the new Vars syntax is used
func (ti *TriggerInput) HasVars() bool {
	return len(ti.Vars) > 0
}

// UnmarshalYAML handles both legacy and new syntax.
// Uses the goccy/go-yaml unmarshaler signature for compatibility with the parser.
func (ti *TriggerInput) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// First, try to decode to raw map to inspect keys
	var raw map[string]interface{}
	if err := unmarshal(&raw); err != nil {
		return err
	}

	// Check for legacy keys
	_, hasType := raw["type"]
	_, hasField := raw["field"]
	_, hasName := raw["name"]
	_, hasPath := raw["path"]
	_, hasFunction := raw["function"]

	if hasType || hasField || hasName || hasPath || hasFunction {
		// Legacy syntax - unmarshal to a temp struct to avoid recursion
		type legacyInput struct {
			Type     string `yaml:"type"`
			Path     string `yaml:"path"`
			Field    string `yaml:"field"`
			Function string `yaml:"function"`
			Name     string `yaml:"name"`
		}
		var legacy legacyInput
		if err := unmarshal(&legacy); err != nil {
			return err
		}
		ti.Type = legacy.Type
		ti.Path = legacy.Path
		ti.Field = legacy.Field
		ti.Function = legacy.Function
		ti.Name = legacy.Name
		return nil
	}

	// New syntax - all keys are variable names with expression values
	ti.Vars = make(map[string]string)
	for k, v := range raw {
		if str, ok := v.(string); ok {
			ti.Vars[k] = str
		}
	}
	return nil
}

// IsCron returns true if this is a cron trigger
func (t *Trigger) IsCron() bool {
	return t.On == TriggerCron
}

// IsEvent returns true if this is an event trigger
func (t *Trigger) IsEvent() bool {
	return t.On == TriggerEvent
}

// IsWatch returns true if this is a file watch trigger
func (t *Trigger) IsWatch() bool {
	return t.On == TriggerWatch
}

// IsManual returns true if this is a manual trigger
func (t *Trigger) IsManual() bool {
	return t.On == TriggerManual
}

// IsEnabled returns true if the trigger is enabled
func (t *Trigger) IsEnabled() bool {
	return t.Enabled
}

// MatchesTopic checks if the trigger's event topic matches the given topic.
// Supports glob patterns:
//   - "*" matches everything
//   - "test*" matches topics starting with "test"
//   - "*.new" matches topics ending with ".new"
//   - "assets.*.created" matches "assets.subdomain.created"
func (t *Trigger) MatchesTopic(topic string) bool {
	if !t.IsEvent() || t.Event == nil {
		return false
	}
	// Empty topic matches all events
	if t.Event.Topic == "" {
		return true
	}

	// Use glob matching if pattern contains wildcards
	if containsWildcard(t.Event.Topic) {
		matched, err := path.Match(t.Event.Topic, topic)
		if err != nil {
			// Invalid pattern, fall back to exact match
			return t.Event.Topic == topic
		}
		return matched
	}

	return t.Event.Topic == topic
}

// HasFilters returns true if the event trigger has filters defined
func (t *Trigger) HasFilters() bool {
	return t.IsEvent() && t.Event != nil && len(t.Event.Filters) > 0
}

// GetFilters returns the filter expressions for the event trigger
func (t *Trigger) GetFilters() []string {
	if t.Event == nil {
		return nil
	}
	return t.Event.Filters
}

// HasFilterFunctions returns true if the event trigger has filter functions defined
func (t *Trigger) HasFilterFunctions() bool {
	return t.IsEvent() && t.Event != nil && len(t.Event.FilterFunctions) > 0
}

// GetFilterFunctions returns the filter function expressions for the event trigger
func (t *Trigger) GetFilterFunctions() []string {
	if t.Event == nil {
		return nil
	}
	return t.Event.FilterFunctions
}

// GetDebounceDuration parses and returns the debounce duration for watch triggers
func (t *Trigger) GetDebounceDuration() time.Duration {
	if t.Debounce == "" {
		return 0
	}
	d, err := time.ParseDuration(t.Debounce)
	if err != nil {
		return 0
	}
	return d
}

// HasDebounce returns true if the trigger has debounce configured
func (t *Trigger) HasDebounce() bool {
	return t.GetDebounceDuration() > 0
}

// GetDedupeWindow parses and returns the deduplication window duration
func (e *EventConfig) GetDedupeWindow() time.Duration {
	if e == nil || e.DedupeWindow == "" {
		return 0
	}
	d, err := time.ParseDuration(e.DedupeWindow)
	if err != nil {
		return 0
	}
	return d
}

// HasDeduplication returns true if the event config has deduplication configured
func (e *EventConfig) HasDeduplication() bool {
	return e != nil && e.DedupeKey != "" && e.GetDedupeWindow() > 0
}
