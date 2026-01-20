package linter

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrettyFormatter_Format(t *testing.T) {
	formatter := NewPrettyFormatter(true)
	formatter.NoColor = true // Disable colors for testing

	source := []byte(`name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "{{undefined}}"
`)

	result := &LintResult{
		FilePath: "test.yaml",
		Issues: []LintIssue{
			{
				Rule:       "undefined-variable",
				Severity:   SeverityError,
				Message:    "Variable 'undefined' is not defined",
				Suggestion: "Check variable name",
				Line:       6,
				Column:     19,
				Field:      "steps[0].command",
			},
		},
		Errors:   1,
		Warnings: 0,
	}

	output := formatter.Format(result, source)

	assert.Contains(t, output, "test.yaml")
	assert.Contains(t, output, "undefined-variable")
	assert.Contains(t, output, "Variable 'undefined' is not defined")
	assert.Contains(t, output, "Suggestion:")
}

func TestPrettyFormatter_FormatNoContext(t *testing.T) {
	formatter := NewPrettyFormatter(false)
	formatter.NoColor = true

	result := &LintResult{
		FilePath: "test.yaml",
		Issues: []LintIssue{
			{
				Rule:     "test-rule",
				Severity: SeverityWarning,
				Message:  "Test message",
				Line:     5,
				Column:   10,
			},
		},
	}

	output := formatter.Format(result, []byte("test content"))

	assert.Contains(t, output, "test.yaml:5:10")
	assert.Contains(t, output, "test-rule")
}

func TestPrettyFormatter_FormatEmpty(t *testing.T) {
	formatter := NewPrettyFormatter(true)

	result := &LintResult{
		FilePath: "test.yaml",
		Issues:   []LintIssue{},
	}

	output := formatter.Format(result, []byte{})
	assert.Empty(t, output)
}

func TestPrettyFormatter_FormatSummary(t *testing.T) {
	formatter := NewPrettyFormatter(true)
	formatter.NoColor = true

	t.Run("no issues", func(t *testing.T) {
		results := []*LintResult{{}}
		summary := formatter.FormatSummary(results)
		assert.Contains(t, summary, "No issues found")
	})

	t.Run("with issues", func(t *testing.T) {
		results := []*LintResult{
			{Errors: 2, Warnings: 1, Issues: []LintIssue{{}, {}, {}}},
		}
		summary := formatter.FormatSummary(results)
		assert.Contains(t, summary, "2 error")
		assert.Contains(t, summary, "1 warning")
	})
}

func TestJSONFormatter_Format(t *testing.T) {
	formatter := NewJSONFormatter()

	result := &LintResult{
		FilePath: "test.yaml",
		Issues: []LintIssue{
			{
				Rule:       "test-rule",
				Severity:   SeverityError,
				Message:    "Test error",
				Suggestion: "Fix it",
				Line:       10,
				Column:     5,
				Field:      "steps[0]",
			},
		},
		Errors:   1,
		Warnings: 0,
	}

	output := formatter.Format(result, []byte{})

	// Verify it's valid JSON
	var jsonOutput JSONOutput
	err := json.Unmarshal([]byte(output), &jsonOutput)
	require.NoError(t, err)

	assert.Equal(t, "test.yaml", jsonOutput.File)
	assert.Len(t, jsonOutput.Issues, 1)
	assert.Equal(t, "test-rule", jsonOutput.Issues[0].Rule)
	assert.Equal(t, "error", jsonOutput.Issues[0].Severity)
	assert.Equal(t, 1, jsonOutput.Summary.Errors)
}

func TestJSONFormatter_FormatSummary(t *testing.T) {
	formatter := NewJSONFormatter()

	results := []*LintResult{
		{
			FilePath: "file1.yaml",
			Issues:   []LintIssue{{Rule: "rule1"}},
			Errors:   1,
		},
		{
			FilePath: "file2.yaml",
			Issues:   []LintIssue{{Rule: "rule2"}},
			Warnings: 1,
		},
	}

	output := formatter.FormatSummary(results)

	// Verify it's valid JSON
	var summary map[string]interface{}
	err := json.Unmarshal([]byte(output), &summary)
	require.NoError(t, err)

	assert.Equal(t, float64(1), summary["total_errors"])
	assert.Equal(t, float64(1), summary["total_warnings"])
	assert.Equal(t, float64(2), summary["total_files"])
}

func TestGitHubFormatter_Format(t *testing.T) {
	formatter := NewGitHubFormatter()

	result := &LintResult{
		FilePath: "test.yaml",
		Issues: []LintIssue{
			{
				Rule:       "test-rule",
				Severity:   SeverityError,
				Message:    "Test error message",
				Suggestion: "Fix suggestion",
				Line:       10,
				Column:     5,
			},
			{
				Rule:     "warn-rule",
				Severity: SeverityWarning,
				Message:  "Warning message",
				Line:     15,
				Column:   1,
			},
		},
	}

	output := formatter.Format(result, []byte{})

	// GitHub annotation format
	assert.Contains(t, output, "::error file=test.yaml,line=10,col=5::")
	assert.Contains(t, output, "::warning file=test.yaml,line=15,col=1::")
	assert.Contains(t, output, "[test-rule]")
	assert.Contains(t, output, "[warn-rule]")
}

func TestGitHubFormatter_FormatEmpty(t *testing.T) {
	formatter := NewGitHubFormatter()

	result := &LintResult{
		FilePath: "test.yaml",
		Issues:   []LintIssue{},
	}

	output := formatter.Format(result, []byte{})
	assert.Empty(t, output)
}

func TestGitHubFormatter_FormatSummary(t *testing.T) {
	formatter := NewGitHubFormatter()

	t.Run("no issues", func(t *testing.T) {
		results := []*LintResult{{}}
		summary := formatter.FormatSummary(results)
		assert.Contains(t, summary, "::notice::")
		assert.Contains(t, summary, "passed")
	})

	t.Run("with issues", func(t *testing.T) {
		results := []*LintResult{
			{Errors: 2, Warnings: 3},
		}
		summary := formatter.FormatSummary(results)
		assert.Contains(t, summary, "::notice::")
		assert.Contains(t, summary, "2 error")
		assert.Contains(t, summary, "3 warning")
	})
}

func TestGetFormatter(t *testing.T) {
	t.Run("pretty format", func(t *testing.T) {
		f := GetFormatter(FormatPretty, true)
		_, ok := f.(*PrettyFormatter)
		assert.True(t, ok)
	})

	t.Run("json format", func(t *testing.T) {
		f := GetFormatter(FormatJSON, true)
		_, ok := f.(*JSONFormatter)
		assert.True(t, ok)
	})

	t.Run("github format", func(t *testing.T) {
		f := GetFormatter(FormatGitHub, true)
		_, ok := f.(*GitHubFormatter)
		assert.True(t, ok)
	})

	t.Run("unknown defaults to pretty", func(t *testing.T) {
		f := GetFormatter(OutputFormat("unknown"), true)
		_, ok := f.(*PrettyFormatter)
		assert.True(t, ok)
	})
}

func TestSeverity_String(t *testing.T) {
	assert.Equal(t, "info", SeverityInfo.String())
	assert.Equal(t, "warning", SeverityWarning.String())
	assert.Equal(t, "error", SeverityError.String())
}

func TestParseSeverity(t *testing.T) {
	assert.Equal(t, SeverityInfo, ParseSeverity("info"))
	assert.Equal(t, SeverityWarning, ParseSeverity("warning"))
	assert.Equal(t, SeverityError, ParseSeverity("error"))
	assert.Equal(t, SeverityWarning, ParseSeverity("unknown")) // default
}

func TestParseOutputFormat(t *testing.T) {
	assert.Equal(t, FormatJSON, ParseOutputFormat("json"))
	assert.Equal(t, FormatGitHub, ParseOutputFormat("github"))
	assert.Equal(t, FormatPretty, ParseOutputFormat("pretty"))
	assert.Equal(t, FormatPretty, ParseOutputFormat("unknown")) // default
}

func TestPrettyFormatter_ColorOutput(t *testing.T) {
	formatter := NewPrettyFormatter(true)
	formatter.NoColor = false // Enable colors

	result := &LintResult{
		FilePath: "test.yaml",
		Issues: []LintIssue{
			{
				Rule:     "error-rule",
				Severity: SeverityError,
				Message:  "Error",
				Line:     1,
				Column:   1,
			},
			{
				Rule:     "warn-rule",
				Severity: SeverityWarning,
				Message:  "Warning",
				Line:     2,
				Column:   1,
			},
		},
	}

	output := formatter.Format(result, []byte("line1\nline2"))

	// Should contain ANSI color codes
	assert.True(t, strings.Contains(output, "\033["))
}
