package core

// Trigger defines when a workflow should execute
type Trigger struct {
	Name     string       `yaml:"name"`
	On       TriggerType  `yaml:"on"`
	Schedule string       `yaml:"schedule,omitempty"` // cron expression (for cron triggers)
	Event    *EventConfig `yaml:"event,omitempty"`    // event configuration (for event triggers)
	Path     string       `yaml:"path,omitempty"`     // watch path (for watch triggers)
	Input    TriggerInput `yaml:"input,omitempty"`
	Enabled  bool         `yaml:"enabled"`
}

// EventConfig holds event trigger configuration
type EventConfig struct {
	Topic   string   `yaml:"topic"`             // e.g., "webhook.received", "assets.new"
	Filters []string `yaml:"filters,omitempty"` // JS expressions: ["event.name == 'discovered'"]
}

// TriggerInput defines the input source for trigger
type TriggerInput struct {
	Type     string `yaml:"type"`               // file, event_data, function, param
	Path     string `yaml:"path,omitempty"`     // for file type
	Field    string `yaml:"field,omitempty"`    // for event_data type
	Function string `yaml:"function,omitempty"` // for function type (e.g., jq("{{event.data}}", ".url"))
	Name     string `yaml:"name,omitempty"`     // parameter name to set
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

// MatchesTopic checks if the trigger's event topic matches the given topic
func (t *Trigger) MatchesTopic(topic string) bool {
	if !t.IsEvent() || t.Event == nil {
		return false
	}
	// Empty topic matches all events
	if t.Event.Topic == "" {
		return true
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
