package core

import (
	"fmt"
	"regexp"
	"strings"
)

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

// IsToggleParam detects boolean/toggle parameters by:
// - Type == "bool"
// - Name patterns: enableX, enable_X, skipX, skip_X, disableX, useX, verboseX
// - Default value is bool (true/false)
func IsToggleParam(p Param) bool {
	if p.Type == "bool" {
		return true
	}

	name := strings.ToLower(p.Name)
	togglePrefixes := []string{"enable", "skip", "disable", "use", "verbose"}
	for _, prefix := range togglePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
		if strings.HasPrefix(name, prefix+"_") {
			return true
		}
	}

	if p.Default != nil {
		switch p.Default.(type) {
		case bool:
			return true
		}
		defaultStr := strings.ToLower(p.DefaultString())
		if defaultStr == "true" || defaultStr == "false" {
			return true
		}
	}

	return false
}

// IsSpeedControlParam detects performance/speed parameters by:
// - Name contains: threads, timeout, rate, concurrency, delay, limit, workers, parallel, batch, interval, retry
// - Name ends with: depth, parallel
// - Default value matches time pattern: \d+[hms]
func IsSpeedControlParam(p Param) bool {
	name := strings.ToLower(p.Name)
	speedPatterns := []string{
		"threads", "timeout", "rate", "concurrency", "delay",
		"limit", "workers", "parallel", "batch", "interval", "retry",
	}

	for _, pattern := range speedPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	speedSuffixes := []string{"depth", "parallel"}
	for _, suffix := range speedSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	if p.Default != nil {
		defaultStr := p.DefaultString()
		timePatternRegex := regexp.MustCompile(`^\d+[hms]?$`)
		if timePatternRegex.MatchString(defaultStr) {
			if len(defaultStr) > 0 {
				lastChar := defaultStr[len(defaultStr)-1]
				if lastChar == 'h' || lastChar == 'm' || lastChar == 's' {
					return true
				}
				if p.Type == "int" || p.Type == "" {
					numericRegex := regexp.MustCompile(`^\d+$`)
					if numericRegex.MatchString(defaultStr) {
						val := 0
						_, _ = fmt.Sscanf(defaultStr, "%d", &val)
						if val > 100 {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// IsConfigParam detects configuration parameters by:
// - Name ends with: Config, config, Cfg, cfg
func IsConfigParam(p Param) bool {
	name := strings.ToLower(p.Name)
	configSuffixes := []string{"config", "cfg"}
	for _, suffix := range configSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// CategorizeParams groups params into Toggle, Speed, Config, and General categories
func CategorizeParams(params []Param) (toggle, speed, config, general []Param) {
	for _, p := range params {
		switch {
		case IsToggleParam(p):
			toggle = append(toggle, p)
		case IsSpeedControlParam(p):
			speed = append(speed, p)
		case IsConfigParam(p):
			config = append(config, p)
		default:
			general = append(general, p)
		}
	}
	return
}
