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
