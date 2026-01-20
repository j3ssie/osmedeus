package linter

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLinter(t *testing.T) {
	t.Run("creates linter with default options", func(t *testing.T) {
		l := NewDefaultLinter()
		assert.NotNil(t, l)
		assert.NotEmpty(t, l.GetRules())
	})

	t.Run("creates linter with custom options", func(t *testing.T) {
		opts := LinterOptions{
			DisabledRules: []string{"unused-variable"},
			MinSeverity:   SeverityWarning,
		}
		l := NewLinter(opts)
		assert.NotNil(t, l)
	})
}

func TestLinter_Lint(t *testing.T) {
	testDataDir := filepath.Join("..", "..", "test", "testdata", "workflows", "linter")

	t.Run("valid workflow has no errors", func(t *testing.T) {
		l := NewDefaultLinter()
		result, err := l.Lint(filepath.Join(testDataDir, "valid-workflow.yaml"))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.HasErrors(), "valid workflow should not have errors")
	})

	t.Run("all-errors workflow has issues", func(t *testing.T) {
		l := NewDefaultLinter()
		result, err := l.Lint(filepath.Join(testDataDir, "all-errors.yaml"))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.HasIssues(), "all-errors workflow should have issues")
		assert.Greater(t, result.Warnings, 0, "all-errors workflow should have warnings")
	})

	t.Run("warnings-only workflow has infos but no errors", func(t *testing.T) {
		l := NewDefaultLinter()
		result, err := l.Lint(filepath.Join(testDataDir, "warnings-only.yaml"))
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.HasErrors(), "warnings-only should not have errors")
		assert.Greater(t, result.Infos, 0, "warnings-only should have infos")
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		l := NewDefaultLinter()
		_, err := l.Lint("nonexistent.yaml")
		require.Error(t, err)
	})
}

func TestLinter_LintContent(t *testing.T) {
	l := NewDefaultLinter()

	t.Run("lints valid content", func(t *testing.T) {
		content := []byte(`
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    timeout: 5m
`)
		result, err := l.LintContent(content, "test.yaml")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("lints content with issues", func(t *testing.T) {
		// Content with empty step (no command) triggers empty-step rule
		content := []byte(`
name: test
kind: module
steps:
  - name: step1
    type: bash
`)
		result, err := l.LintContent(content, "test.yaml")
		require.NoError(t, err)
		assert.True(t, result.HasIssues())
	})
}

func TestLinter_DisabledRules(t *testing.T) {
	// Content with empty step to test disabling empty-step rule
	content := []byte(`
name: test
kind: module
steps:
  - name: step1
    type: bash
`)

	t.Run("rule is active by default", func(t *testing.T) {
		l := NewDefaultLinter()
		result, err := l.LintContent(content, "test.yaml")
		require.NoError(t, err)
		hasEmptyStepIssue := false
		for _, issue := range result.Issues {
			if issue.Rule == "empty-step" {
				hasEmptyStepIssue = true
				break
			}
		}
		assert.True(t, hasEmptyStepIssue)
	})

	t.Run("disabled rule is not run", func(t *testing.T) {
		opts := LinterOptions{
			DisabledRules: []string{"empty-step"},
		}
		l := NewLinter(opts)
		result, err := l.LintContent(content, "test.yaml")
		require.NoError(t, err)
		hasEmptyStepIssue := false
		for _, issue := range result.Issues {
			if issue.Rule == "empty-step" {
				hasEmptyStepIssue = true
				break
			}
		}
		assert.False(t, hasEmptyStepIssue)
	})
}

func TestLinter_MinSeverity(t *testing.T) {
	// Content that produces infos (unused export)
	content := []byte(`
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      unused_export: "output"
  - name: step2
    type: bash
    command: echo "done"
`)

	t.Run("default shows infos", func(t *testing.T) {
		l := NewDefaultLinter()
		result, err := l.LintContent(content, "test.yaml")
		require.NoError(t, err)
		hasInfo := false
		for _, issue := range result.Issues {
			if issue.Severity == SeverityInfo {
				hasInfo = true
				break
			}
		}
		assert.True(t, hasInfo)
	})

	t.Run("error severity filters lower severities", func(t *testing.T) {
		opts := LinterOptions{
			MinSeverity: SeverityError,
		}
		l := NewLinter(opts)
		result, err := l.LintContent(content, "test.yaml")
		require.NoError(t, err)
		// All rules now return warnings/infos, so filtering by error should return no issues
		assert.Empty(t, result.Issues, "should have no issues when filtering by error severity")
	})
}

func TestLinter_RegisterRule(t *testing.T) {
	l := NewDefaultLinter()
	initialCount := len(l.GetRules())

	// Create a custom rule
	customRule := &UnusedVariableRule{} // Using existing rule as example
	l.RegisterRule(customRule)

	assert.Equal(t, initialCount+1, len(l.GetRules()))
}

func TestLinter_SetRules(t *testing.T) {
	l := NewDefaultLinter()

	// Set only one rule
	l.SetRules([]LinterRule{&UnusedVariableRule{}})

	assert.Len(t, l.GetRules(), 1)
}

func TestLintResult_HasErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		r := &LintResult{Errors: 0, Warnings: 2}
		assert.False(t, r.HasErrors())
	})

	t.Run("has errors", func(t *testing.T) {
		r := &LintResult{Errors: 1, Warnings: 0}
		assert.True(t, r.HasErrors())
	})
}

func TestLintResult_HasIssues(t *testing.T) {
	t.Run("no issues", func(t *testing.T) {
		r := &LintResult{Issues: []LintIssue{}}
		assert.False(t, r.HasIssues())
	})

	t.Run("has issues", func(t *testing.T) {
		r := &LintResult{Issues: []LintIssue{{Rule: "test"}}}
		assert.True(t, r.HasIssues())
	})
}

func TestTotalErrors(t *testing.T) {
	results := []*LintResult{
		{Errors: 2},
		{Errors: 3},
		{Errors: 0},
	}
	assert.Equal(t, 5, TotalErrors(results))
}

func TestTotalWarnings(t *testing.T) {
	results := []*LintResult{
		{Warnings: 1},
		{Warnings: 4},
		{Warnings: 0},
	}
	assert.Equal(t, 5, TotalWarnings(results))
}

func TestTotalIssues(t *testing.T) {
	results := []*LintResult{
		{Issues: []LintIssue{{}, {}}},
		{Issues: []LintIssue{{}}},
		{Issues: []LintIssue{}},
	}
	assert.Equal(t, 3, TotalIssues(results))
}
