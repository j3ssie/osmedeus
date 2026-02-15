package executor

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// exitCodeError wraps an error with its numeric exit code so callers
// can check for specific codes (e.g., 127 = command not found).
type exitCodeError struct {
	code int
	msg  string
}

func (e *exitCodeError) Error() string {
	return e.msg
}

// newExitCodeError creates an exitCodeError with the given code and message.
func newExitCodeError(code int, msg string) *exitCodeError {
	return &exitCodeError{code: code, msg: msg}
}

// newExitCodeErrorf creates an exitCodeError with formatted message.
func newExitCodeErrorf(code int, format string, args ...any) *exitCodeError {
	return &exitCodeError{code: code, msg: fmt.Sprintf(format, args...)}
}

// stripTimeoutResult holds the result of parsing a timeout prefix.
type stripTimeoutResult struct {
	command  string        // remaining command after stripping prefix
	duration time.Duration // parsed duration from the timeout prefix (0 if unparseable)
	stripped bool          // true if a timeout prefix was found and stripped
}

// stripTimeoutPrefix removes a "timeout" command prefix from a command string.
// It handles various flag forms: -k VAL, --kill-after=VAL, -s SIG, --signal=SIG,
// --foreground, --preserve-status, -v, --verbose.
// Returns a stripTimeoutResult with the remaining command, parsed duration, and whether stripping occurred.
func stripTimeoutPrefix(command string) stripTimeoutResult {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return stripTimeoutResult{command: command}
	}

	// First token must be "timeout" (or a path ending in /timeout)
	base := fields[0]
	if idx := strings.LastIndex(base, "/"); idx >= 0 {
		base = base[idx+1:]
	}
	if base != "timeout" {
		return stripTimeoutResult{command: command}
	}

	i := 1 // skip "timeout"

	// Parse optional flags before the DURATION argument
	for i < len(fields) {
		arg := fields[i]

		// Flags that take a separate value: -k VAL, -s SIG
		if arg == "-k" || arg == "--kill-after" || arg == "-s" || arg == "--signal" {
			i += 2 // skip flag + value
			continue
		}

		// Flags with = form: --kill-after=VAL, --signal=SIG
		if strings.HasPrefix(arg, "--kill-after=") || strings.HasPrefix(arg, "--signal=") {
			i++
			continue
		}

		// Short form -k5s (value attached)
		if len(arg) > 2 && arg[0] == '-' && arg[1] == 'k' {
			i++
			continue
		}

		// Boolean flags
		if arg == "--foreground" || arg == "--preserve-status" || arg == "-v" || arg == "--verbose" {
			i++
			continue
		}

		// Not a recognized flag — this should be the DURATION
		break
	}

	// Skip the DURATION argument
	if i >= len(fields) {
		// No duration found — malformed, don't strip
		return stripTimeoutResult{command: command}
	}
	durationStr := fields[i]
	parsedDuration := parseTimeoutDuration(durationStr)
	i++ // skip duration

	// Everything after duration is the actual command
	if i >= len(fields) {
		// Nothing after duration — no command to run
		return stripTimeoutResult{command: "", duration: parsedDuration, stripped: true}
	}

	return stripTimeoutResult{
		command:  strings.Join(fields[i:], " "),
		duration: parsedDuration,
		stripped: true,
	}
}

// parseTimeoutDuration parses a GNU coreutils timeout duration string.
// Supports formats: plain number (seconds), or number with suffix s/m/h/d.
// Returns 0 if the string cannot be parsed.
func parseTimeoutDuration(s string) time.Duration {
	if s == "" {
		return 0
	}

	// Check for suffix
	last := s[len(s)-1]
	switch last {
	case 's':
		return parseDurationNumber(s[:len(s)-1], time.Second)
	case 'm':
		return parseDurationNumber(s[:len(s)-1], time.Minute)
	case 'h':
		return parseDurationNumber(s[:len(s)-1], time.Hour)
	case 'd':
		return parseDurationNumber(s[:len(s)-1], 24*time.Hour)
	default:
		// No suffix — default is seconds
		return parseDurationNumber(s, time.Second)
	}
}

// parseDurationNumber parses a numeric string and multiplies by the given unit.
// Supports both integer and floating-point values. Returns 0 on parse error.
func parseDurationNumber(s string, unit time.Duration) time.Duration {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return time.Duration(f * float64(unit))
}

// prependBinariesPath prepends binariesPath to the first token (binary name)
// of a command if it doesn't already contain a path separator.
// Returns the modified command and true if the path was prepended.
func prependBinariesPath(command, binariesPath string) (string, bool) {
	if binariesPath == "" || command == "" {
		return command, false
	}

	fields := strings.Fields(command)
	if len(fields) == 0 {
		return command, false
	}

	binary := fields[0]

	// Don't prepend if binary already has a path
	if strings.Contains(binary, "/") {
		return command, false
	}

	// Ensure binariesPath doesn't have trailing slash
	binariesPath = strings.TrimRight(binariesPath, "/")

	fields[0] = binariesPath + "/" + binary
	return strings.Join(fields, " "), true
}
