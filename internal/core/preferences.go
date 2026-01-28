package core

// Preferences defines workflow-level execution preferences.
// All fields are pointers to distinguish "not set" (nil) from "explicitly false".
// When a preference is set in a workflow, it serves as a default that can be
// overridden by explicit CLI flags.
type Preferences struct {
	// DisableNotifications turns off all notifications (--disable-notification)
	DisableNotifications *bool `yaml:"disable_notifications,omitempty"`

	// DisableLogging turns off all logging output (--disable-logging)
	DisableLogging *bool `yaml:"disable_logging,omitempty"`

	// HeuristicsCheck sets the heuristics check level: "none", "basic", "advanced" (--heuristics-check)
	HeuristicsCheck *string `yaml:"heuristics_check,omitempty"`

	// CIOutputFormat outputs results in JSON format for CI pipelines (--ci-output-format)
	CIOutputFormat *bool `yaml:"ci_output_format,omitempty"`

	// Silent suppresses all output except errors (--silent)
	Silent *bool `yaml:"silent,omitempty"`

	// Repeat enables repeat mode after completion (--repeat)
	Repeat *bool `yaml:"repeat,omitempty"`

	// RepeatWaitTime sets wait time between repeats, e.g., "60s", "1h" (--repeat-wait-time)
	RepeatWaitTime *string `yaml:"repeat_wait_time,omitempty"`

	// EmptyTarget allows running without a target (--empty-target)
	EmptyTarget *bool `yaml:"empty_target,omitempty"`
}

// Helper functions to safely get values with defaults

// GetDisableNotifications returns the disable_notifications preference or the default value
func (p *Preferences) GetDisableNotifications(defaultVal bool) bool {
	if p == nil || p.DisableNotifications == nil {
		return defaultVal
	}
	return *p.DisableNotifications
}

// GetDisableLogging returns the disable_logging preference or the default value
func (p *Preferences) GetDisableLogging(defaultVal bool) bool {
	if p == nil || p.DisableLogging == nil {
		return defaultVal
	}
	return *p.DisableLogging
}

// GetHeuristicsCheck returns the heuristics_check preference or the default value
func (p *Preferences) GetHeuristicsCheck(defaultVal string) string {
	if p == nil || p.HeuristicsCheck == nil {
		return defaultVal
	}
	return *p.HeuristicsCheck
}

// GetCIOutputFormat returns the ci_output_format preference or the default value
func (p *Preferences) GetCIOutputFormat(defaultVal bool) bool {
	if p == nil || p.CIOutputFormat == nil {
		return defaultVal
	}
	return *p.CIOutputFormat
}

// GetSilent returns the silent preference or the default value
func (p *Preferences) GetSilent(defaultVal bool) bool {
	if p == nil || p.Silent == nil {
		return defaultVal
	}
	return *p.Silent
}

// GetRepeat returns the repeat preference or the default value
func (p *Preferences) GetRepeat(defaultVal bool) bool {
	if p == nil || p.Repeat == nil {
		return defaultVal
	}
	return *p.Repeat
}

// GetRepeatWaitTime returns the repeat_wait_time preference or the default value
func (p *Preferences) GetRepeatWaitTime(defaultVal string) string {
	if p == nil || p.RepeatWaitTime == nil {
		return defaultVal
	}
	return *p.RepeatWaitTime
}

// GetEmptyTarget returns the empty_target preference or the default value
func (p *Preferences) GetEmptyTarget(defaultVal bool) bool {
	if p == nil || p.EmptyTarget == nil {
		return defaultVal
	}
	return *p.EmptyTarget
}
