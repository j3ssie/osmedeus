package terminal

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// ANSI color codes
const (
	colorReset     = "\033[0m"
	colorBold      = "\033[1m"
	colorRed       = "\033[31m"
	colorGreen     = "\033[32m"
	colorYellow    = "\033[33m"
	colorBlue      = "\033[34m"
	colorMagenta   = "\033[35m"
	colorCyan      = "\033[36m"
	colorWhite     = "\033[37m"
	colorGray      = "\033[90m"
	colorHiGreen   = "\033[92m"
	colorHiBlue    = "\033[94m"
	colorHiMagenta = "\033[95m"
	colorHiCyan    = "\033[96m"
	colorHiWhite   = "\033[97m"
	colorTeal      = "\033[38;5;30m" // Teal (256-color mode)
)

var colorEnabled = true
var ciMode = false

func init() {
	// Disable colors if not a terminal or NO_COLOR is set
	if !term.IsTerminal(int(os.Stdout.Fd())) || os.Getenv("NO_COLOR") != "" {
		colorEnabled = false
	}
}

// IsColorEnabled returns whether color output is enabled
func IsColorEnabled() bool {
	return colorEnabled
}

// SetColorEnabled enables or disables color output
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

// SetCIMode enables or disables CI mode (suppresses decorative output, enables JSON)
func SetCIMode(enabled bool) {
	ciMode = enabled
}

// IsCIMode returns whether CI mode is enabled
func IsCIMode() bool {
	return ciMode
}

// colorize wraps text with ANSI color codes
func colorize(color, text string) string {
	if !colorEnabled {
		return text
	}
	return color + text + colorReset
}

// Red returns text in red color
func Red(s string) string {
	return colorize(colorRed, s)
}

// Green returns text in green color
func Green(s string) string {
	return colorize(colorGreen, s)
}

// Yellow returns text in yellow color
func Yellow(s string) string {
	return colorize(colorYellow, s)
}

// Blue returns text in blue color
func Blue(s string) string {
	return colorize(colorBlue, s)
}

// Magenta returns text in magenta color
func Magenta(s string) string {
	return colorize(colorMagenta, s)
}

// Cyan returns text in cyan color
func Cyan(s string) string {
	return colorize(colorCyan, s)
}

// White returns text in white color
func White(s string) string {
	return colorize(colorWhite, s)
}

// Gray returns text in gray color
func Gray(s string) string {
	return colorize(colorGray, s)
}

// Bold returns text in bold
func Bold(s string) string {
	return colorize(colorBold, s)
}

// BoldCyan returns text in bold cyan
func BoldCyan(s string) string {
	if !colorEnabled {
		return s
	}
	return colorBold + colorCyan + s + colorReset
}

// BoldGreen returns text in bold green
func BoldGreen(s string) string {
	if !colorEnabled {
		return s
	}
	return colorBold + colorGreen + s + colorReset
}

// BoldRed returns text in bold red
func BoldRed(s string) string {
	if !colorEnabled {
		return s
	}
	return colorBold + colorRed + s + colorReset
}

// BoldYellow returns text in bold yellow
func BoldYellow(s string) string {
	if !colorEnabled {
		return s
	}
	return colorBold + colorYellow + s + colorReset
}

// BoldMagenta returns text in bold magenta
func BoldMagenta(s string) string {
	if !colorEnabled {
		return s
	}
	return colorBold + colorMagenta + s + colorReset
}

// BoldBlue returns text in bold blue
func BoldBlue(s string) string {
	if !colorEnabled {
		return s
	}
	return colorBold + colorBlue + s + colorReset
}

// HiWhite returns text in bright white color
func HiWhite(s string) string {
	return colorize(colorHiWhite, s)
}

// HiGreen returns text in bright green color
func HiGreen(s string) string {
	return colorize(colorHiGreen, s)
}

// HiCyan returns text in bright cyan color
func HiCyan(s string) string {
	return colorize(colorHiCyan, s)
}

// HiMagenta returns text in bright magenta color
func HiMagenta(s string) string {
	return colorize(colorHiMagenta, s)
}

// HiBlue returns text in bright blue color
func HiBlue(s string) string {
	return colorize(colorHiBlue, s)
}

// Teal returns text in teal color
func Teal(s string) string {
	return colorize(colorTeal, s)
}

// ColorizeStatus applies ANSI color codes to status values for table display
func ColorizeStatus(status string) string {
	switch strings.ToLower(status) {
	case "running", "in_progress", "active":
		return Blue(status)
	case "failed", "error":
		return Red(status)
	case "completed", "success", "done":
		return Green(status)
	case "cancelled", "canceled":
		return Yellow(status)
	case "pending", "waiting", "queued":
		return Gray(status)
	default:
		return status
	}
}

// ColorizeTriggerType applies ANSI color codes to trigger_type values for table display
func ColorizeTriggerType(triggerType string) string {
	switch strings.ToLower(triggerType) {
	case "cli":
		return Cyan(triggerType)
	case "cron":
		return Yellow(triggerType)
	case "event":
		return Magenta(triggerType)
	case "webhook":
		return Blue(triggerType)
	case "manual":
		return Gray(triggerType)
	default:
		return triggerType
	}
}

// ColorizeEnabled applies ANSI color codes to boolean enabled/disabled values
func ColorizeEnabled(val string) string {
	switch strings.ToLower(val) {
	case "true", "yes", "1":
		return Green(val)
	case "false", "no", "0":
		return Red(val)
	default:
		return val
	}
}

// ColorizeWorkflowKind applies ANSI color codes to workflow kind values
func ColorizeWorkflowKind(kind string) string {
	switch strings.ToLower(kind) {
	case "flow":
		return Blue(kind)
	case "module":
		return Cyan(kind)
	default:
		return kind
	}
}

// ColorizeSchedule applies ANSI color codes to schedule/cron expressions
func ColorizeSchedule(schedule string) string {
	if schedule == "" {
		return schedule
	}
	return Gray(schedule)
}

// ColorizeStatusCode applies ANSI color codes to HTTP status codes
func ColorizeStatusCode(code string) string {
	if code == "" || code == "0" {
		return Gray(code)
	}
	if len(code) >= 1 {
		switch code[0] {
		case '2':
			return Green(code)
		case '3':
			return Cyan(code)
		case '4':
			return Yellow(code)
		case '5':
			return Red(code)
		}
	}
	return code
}

// ColorizeSource applies ANSI color codes to asset source values
func ColorizeSource(source string) string {
	if source == "" {
		return Gray(source)
	}
	switch strings.ToLower(source) {
	case "httpx":
		return Cyan(source)
	case "subfinder", "amass":
		return Blue(source)
	case "nmap", "rustscan":
		return Magenta(source)
	default:
		return Teal(source)
	}
}

// ColorizeAssetType applies ANSI color codes to asset type values
func ColorizeAssetType(assetType string) string {
	if assetType == "" {
		return Gray(assetType)
	}
	switch strings.ToLower(assetType) {
	case "web":
		return Green(assetType)
	case "subdomain":
		return Blue(assetType)
	case "ip":
		return Magenta(assetType)
	case "cidr":
		return Yellow(assetType)
	default:
		return Cyan(assetType)
	}
}
