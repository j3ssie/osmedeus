package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "single color code",
			input:    "\x1b[90mhello\x1b[0m",
			expected: "hello",
		},
		{
			name:     "multiple color codes",
			input:    "\x1b[1m\x1b[31mred bold\x1b[0m",
			expected: "red bold",
		},
		{
			name:     "color in middle of text",
			input:    "start \x1b[90mmiddle\x1b[0m end",
			expected: "start middle end",
		},
		{
			name:     "timestamp with ANSI",
			input:    "\x1b[90m2026-01-23T17:59:20+08:00\x1b[0m",
			expected: "2026-01-23T17:59:20+08:00",
		},
		{
			name:     "JSON block with embedded ANSI timestamp",
			input:    "{\"timestamp\": \"\x1b[90m2026-01-23T17:59:20+08:00\x1b[0m\"}",
			expected: "{\"timestamp\": \"2026-01-23T17:59:20+08:00\"}",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorJSONFields(t *testing.T) {
	const gray = "\x1b[90m"
	const reset = "\x1b[0m"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no JSON",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "simple JSON block",
			input:    `{"key": "value"}`,
			expected: gray + `{"key": "value"}` + reset,
		},
		{
			name:     "text with JSON",
			input:    `INFO message {"data": "test"}`,
			expected: `INFO message ` + gray + `{"data": "test"}` + reset,
		},
		{
			name:     "JSON with embedded ANSI (stripped)",
			input:    "{\"timestamp\": \"\x1b[90m2026-01-23\x1b[0m\"}",
			expected: gray + `{"timestamp": "2026-01-23"}` + reset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colorJSONFields(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
