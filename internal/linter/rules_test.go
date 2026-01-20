package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseTestWorkflow(t *testing.T, content string) *WorkflowAST {
	ast, err := ParseWorkflowASTFromContent([]byte(content), "test.yaml")
	require.NoError(t, err)
	return ast
}

func TestUndefinedVariableRule(t *testing.T) {
	rule := &UndefinedVariableRule{}

	t.Run("detects undefined variable", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: step1
    type: bash
    command: echo "{{undefined_var}}"
`)
		issues := rule.Check(ast)
		assert.NotEmpty(t, issues)
		assert.Equal(t, "undefined-variable", issues[0].Rule)
		assert.Contains(t, issues[0].Message, "undefined_var")
	})

	t.Run("allows defined params", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: step1
    type: bash
    command: echo "{{target}}"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})

	t.Run("allows built-in variables", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "{{Target}} {{Output}} {{Workspace}}"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})

	t.Run("allows exports from previous steps", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      my_output: "output"
  - name: step2
    type: bash
    command: echo "{{my_output}}"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestUnusedVariableRule(t *testing.T) {
	rule := &UnusedVariableRule{}

	t.Run("detects unused export", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      unused_var: "output"
  - name: step2
    type: bash
    command: echo "done"
`)
		issues := rule.Check(ast)
		assert.NotEmpty(t, issues)
		assert.Equal(t, "unused-variable", issues[0].Rule)
		assert.Contains(t, issues[0].Message, "unused_var")
	})

	t.Run("no issue when export is used", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      my_output: "output"
  - name: step2
    type: bash
    command: echo "{{my_output}}"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestDuplicateStepNameRule(t *testing.T) {
	rule := &DuplicateStepNameRule{}

	t.Run("detects duplicate step names", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: same-name
    type: bash
    command: echo "first"
  - name: same-name
    type: bash
    command: echo "second"
`)
		issues := rule.Check(ast)
		assert.NotEmpty(t, issues)
		assert.Equal(t, "duplicate-step-name", issues[0].Rule)
	})

	t.Run("no issue with unique names", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step-one
    type: bash
    command: echo "first"
  - name: step-two
    type: bash
    command: echo "second"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestEmptyStepRule(t *testing.T) {
	rule := &EmptyStepRule{}

	t.Run("detects empty bash step", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: empty-step
    type: bash
`)
		issues := rule.Check(ast)
		assert.NotEmpty(t, issues)
		assert.Equal(t, "empty-step", issues[0].Rule)
	})

	t.Run("no issue with command", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestInvalidGotoRule(t *testing.T) {
	rule := &InvalidGotoRule{}

	t.Run("detects invalid goto reference", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    decision:
      switch: "{{target}}"
      cases:
        "skip":
          goto: nonexistent-step
`)
		issues := rule.Check(ast)
		assert.NotEmpty(t, issues)
		assert.Equal(t, "invalid-goto", issues[0].Rule)
	})

	t.Run("allows _end as goto target", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    decision:
      switch: "{{target}}"
      cases:
        "skip":
          goto: _end
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})

	t.Run("allows valid step reference", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    decision:
      switch: "{{target}}"
      cases:
        "jump":
          goto: step2
  - name: step2
    type: bash
    command: echo "world"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestInvalidDependsOnRule(t *testing.T) {
	rule := &InvalidDependsOnRule{}

	t.Run("detects invalid depends_on reference", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    depends_on:
      - nonexistent-step
`)
		issues := rule.Check(ast)
		assert.NotEmpty(t, issues)
		assert.Equal(t, "invalid-depends-on", issues[0].Rule)
	})

	t.Run("allows valid dependency", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
  - name: step2
    type: bash
    command: echo "world"
    depends_on:
      - step1
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestCircularDependencyRule(t *testing.T) {
	rule := &CircularDependencyRule{}

	t.Run("detects circular dependency", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step-a
    type: bash
    command: echo "a"
    depends_on:
      - step-b
  - name: step-b
    type: bash
    command: echo "b"
    depends_on:
      - step-a
`)
		issues := rule.Check(ast)
		assert.NotEmpty(t, issues)
		assert.Equal(t, "circular-dependency", issues[0].Rule)
	})

	t.Run("no issue with linear dependencies", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "first"
  - name: step2
    type: bash
    command: echo "second"
    depends_on:
      - step1
  - name: step3
    type: bash
    command: echo "third"
    depends_on:
      - step2
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestMissingRequiredFieldRule(t *testing.T) {
	rule := &MissingRequiredFieldRule{}

	t.Run("detects missing step name", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - type: bash
    command: echo "hello"
`)
		issues := rule.Check(ast)
		hasNameIssue := false
		for _, issue := range issues {
			if issue.Rule == "missing-required-field" && issue.Field == "steps[0].name" {
				hasNameIssue = true
				break
			}
		}
		assert.True(t, hasNameIssue, "should detect missing step name")
	})

	t.Run("detects missing step type", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    command: echo "hello"
`)
		issues := rule.Check(ast)
		hasTypeIssue := false
		for _, issue := range issues {
			if issue.Rule == "missing-required-field" && issue.Field == "steps[0].type" {
				hasTypeIssue = true
				break
			}
		}
		assert.True(t, hasTypeIssue, "should detect missing step type")
	})
}

func TestGetDefaultRules(t *testing.T) {
	rules := GetDefaultRules()
	assert.NotEmpty(t, rules)
	assert.Len(t, rules, 7, "should have 7 default rules")

	// Verify all rules have required methods
	for _, rule := range rules {
		assert.NotEmpty(t, rule.Name())
		assert.NotEmpty(t, rule.Description())
		// Severity is valid (0-2)
		assert.GreaterOrEqual(t, int(rule.Severity()), 0)
		assert.LessOrEqual(t, int(rule.Severity()), 2)
	}
}
