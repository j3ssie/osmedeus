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
	assert.Len(t, rules, 8, "should have 8 default rules")

	// Verify all rules have required methods
	for _, rule := range rules {
		assert.NotEmpty(t, rule.Name())
		assert.NotEmpty(t, rule.Description())
		// Severity is valid (0-2)
		assert.GreaterOrEqual(t, int(rule.Severity()), 0)
		assert.LessOrEqual(t, int(rule.Severity()), 2)
	}
}

// --- Comprehensive tests for enhanced UndefinedVariableRule ---

func TestUndefinedVariableRule_ParallelFunctions(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: step1
    type: function
    parallel_functions:
      - 'log_info("{{target}}")'
      - 'log_info("{{undefined_pf}}")'
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 1)
	assert.Contains(t, issues[0].Message, "undefined_pf")
}

func TestUndefinedVariableRule_ForeachInnerStep(t *testing.T) {
	rule := &UndefinedVariableRule{}

	t.Run("detects undefined in foreach inner step", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: scan
    type: foreach
    input: "{{target}}"
    variable: line
    step:
      name: inner
      type: bash
      command: echo "{{undefined_inner}}"
`)
		issues := rule.Check(ast)
		assert.NotEmpty(t, issues)
		hasUndefined := false
		for _, issue := range issues {
			if issue.Rule == "undefined-variable" {
				assert.Contains(t, issue.Message, "undefined_inner")
				hasUndefined = true
			}
		}
		assert.True(t, hasUndefined)
	})

	t.Run("allows loop variable in foreach inner step", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: scan
    type: foreach
    input: "{{target}}"
    variable: line
    step:
      name: inner
      type: bash
      command: echo "[[line]]"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})

	t.Run("allows _id_ in foreach inner step", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: scan
    type: foreach
    input: "{{target}}"
    variable: line
    step:
      name: inner
      type: bash
      command: echo "{{_id_}} [[line]]"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestUndefinedVariableRule_ParallelSteps(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: parallel
    type: parallel-steps
    parallel_steps:
      - name: sub-a
        type: bash
        command: echo "{{target}}"
      - name: sub-b
        type: bash
        command: echo "{{undefined_parallel}}"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 1)
	assert.Contains(t, issues[0].Message, "undefined_parallel")
}

func TestUndefinedVariableRule_DecisionCaseCommands(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
  - name: mode
steps:
  - name: step1
    type: bash
    command: echo "{{target}}"
    decision:
      switch: "{{mode}}"
      cases:
        "fast":
          command: echo "{{undefined_case}}"
        "slow":
          functions:
            - 'log_info("{{target}}")'
      default:
        command: echo "{{undefined_default}}"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 2)
	messages := []string{issues[0].Message, issues[1].Message}
	assert.True(t, containsStr(messages, "undefined_case"), "should detect undefined_case")
	assert.True(t, containsStr(messages, "undefined_default"), "should detect undefined_default")
}

func TestUndefinedVariableRule_DecisionConditions(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: step1
    type: bash
    command: echo "{{target}}"
    decision:
      conditions:
        - if: "{{target}} == 'foo'"
          command: echo "{{undefined_cond}}"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 1)
	assert.Contains(t, issues[0].Message, "undefined_cond")
}

func TestUndefinedVariableRule_SpeedArgs(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: step1
    type: bash
    command: echo "{{target}}"
    speed_args: "--threads {{undefined_speed}}"
    config_args: "--config {{undefined_config}}"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 2)
	messages := []string{issues[0].Message, issues[1].Message}
	assert.True(t, containsStr(messages, "undefined_speed"))
	assert.True(t, containsStr(messages, "undefined_config"))
}

func TestUndefinedVariableRule_AgentFields(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: agent-step
    type: agent
    query: "Scan {{target}} for {{undefined_agent_q}}"
    system_prompt: "You are {{undefined_agent_sp}}"
    max_iterations: 5
    agent_tools:
      - preset: bash
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 2)
	messages := []string{issues[0].Message, issues[1].Message}
	assert.True(t, containsStr(messages, "undefined_agent_q"))
	assert.True(t, containsStr(messages, "undefined_agent_sp"))
}

func TestUndefinedVariableRule_HTTPFields(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: http-step
    type: http
    url: "https://{{target}}/api"
    request_body: '{"host":"{{undefined_body}}"}'
    headers:
      Authorization: "Bearer {{undefined_header}}"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 2)
	messages := []string{issues[0].Message, issues[1].Message}
	assert.True(t, containsStr(messages, "undefined_body"))
	assert.True(t, containsStr(messages, "undefined_header"))
}

func TestUndefinedVariableRule_LLMMessages(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: llm-step
    type: llm
    messages:
      - role: user
        content: "Analyze {{undefined_llm}}"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 1)
	assert.Contains(t, issues[0].Message, "undefined_llm")
}

func TestUndefinedVariableRule_LogField(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: step1
    type: bash
    command: echo "{{target}}"
    log: "Running {{undefined_log}}"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 1)
	assert.Contains(t, issues[0].Message, "undefined_log")
}

func TestUndefinedVariableRule_StdFile(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: step1
    type: bash
    command: echo "{{target}}"
    std_file: "{{Output}}/{{undefined_std}}.txt"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 1)
	assert.Contains(t, issues[0].Message, "undefined_std")
}

func TestUndefinedVariableRule_MemoryPaths(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: agent-step
    type: agent
    query: "Scan {{target}}"
    max_iterations: 5
    agent_tools:
      - preset: bash
    memory:
      max_messages: 30
      persist_path: "{{Output}}/{{undefined_persist}}.json"
`)
	issues := rule.Check(ast)
	assert.Len(t, issues, 1)
	assert.Contains(t, issues[0].Message, "undefined_persist")
}

func TestUndefinedVariableRule_PlatformVariables(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "{{PlatformOS}} {{PlatformArch}} {{PlatformInDocker}} {{PlatformInKubernetes}} {{PlatformCloudProvider}}"
`)
	issues := rule.Check(ast)
	assert.Empty(t, issues)
}

// --- Comprehensive tests for enhanced UnusedVariableRule ---

func TestUnusedVariableRule_ReferencedInNewFields(t *testing.T) {
	rule := &UnusedVariableRule{}

	t.Run("export used in parallel_functions", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      my_var: "output"
  - name: step2
    type: function
    parallel_functions:
      - 'log_info("{{my_var}}")'
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})

	t.Run("export used in speed_args", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      thread_count: "10"
  - name: step2
    type: bash
    command: echo "run"
    speed_args: "--threads {{thread_count}}"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})

	t.Run("export used in agent query", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      scan_data: "output"
  - name: step2
    type: agent
    query: "Analyze {{scan_data}}"
    max_iterations: 5
    agent_tools:
      - preset: bash
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})

	t.Run("export used in decision case command", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: mode
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      result_path: "/tmp/out"
  - name: step2
    type: bash
    command: echo "check"
    decision:
      switch: "{{mode}}"
      cases:
        "process":
          command: echo "{{result_path}}"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})

	t.Run("export used in foreach inner step", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: test
kind: module
params:
  - name: target
steps:
  - name: step1
    type: bash
    command: echo "hello"
    exports:
      targets_file: "targets.txt"
  - name: step2
    type: foreach
    input: "{{targets_file}}"
    variable: line
    step:
      name: inner
      type: bash
      command: echo "[[line]] from {{targets_file}}"
`)
		issues := rule.Check(ast)
		assert.Empty(t, issues)
	})
}

func TestUndefinedVariableRule_FunctionsAndForeachCommands(t *testing.T) {
	rule := &UndefinedVariableRule{}
	ast := parseTestWorkflow(t, `
name: undefined-vars-comprehensive
kind: module
params:
  - name: target
    required: true
  - name: threads
    default: 10
steps:
  - name: func-with-undef
    type: function
    functions:
      - 'log_info("target={{target}}")'
      - 'log_info("bad={{funcUndef}}")'
  - name: foreach-with-undef
    type: foreach
    input: "{{target}}"
    variable: line
    step:
      name: inner-step
      type: bash
      commands:
        - echo "valid=[[line]]"
        - echo "bad={{foreachCmdUndef}}"
  - name: foreach-func-undef
    type: foreach
    input: "{{target}}"
    variable: item
    step:
      name: inner-func
      type: function
      function: 'log_info("bad={{foreachFuncUndef}}")'
  - name: valid-step
    type: bash
    command: echo "{{target}} {{Output}}"
`)

	// Should detect exactly 3 undefined variables
	undefinedIssues := []LintIssue{}
	for _, issue := range ast.Workflow.Steps {
		_ = issue // just iterating to ensure parse worked
	}
	issues := rule.Check(ast)
	for _, issue := range issues {
		if issue.Rule == "undefined-variable" {
			undefinedIssues = append(undefinedIssues, issue)
		}
	}
	assert.Len(t, undefinedIssues, 3, "expected exactly 3 undefined variable issues, got %d: %v", len(undefinedIssues), undefinedIssues)

	// Collect all messages
	messages := make([]string, len(undefinedIssues))
	for i, issue := range undefinedIssues {
		messages[i] = issue.Message
	}

	assert.True(t, containsStr(messages, "funcUndef"), "should detect funcUndef in functions array")
	assert.True(t, containsStr(messages, "foreachCmdUndef"), "should detect foreachCmdUndef in foreach inner commands")
	assert.True(t, containsStr(messages, "foreachFuncUndef"), "should detect foreachFuncUndef in foreach inner function")
}

// containsStr checks if any string in the slice contains the substring.
func containsStr(strs []string, substr string) bool {
	for _, s := range strs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestUndefinedVariableRule_TriggerInputVars(t *testing.T) {
	rule := &UndefinedVariableRule{}

	t.Run("trigger input vars downgraded to info", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: event-handler
kind: module
params:
  - name: target
triggers:
  - name: on-new-asset
    on: event
    enabled: true
    event:
      topic: assets.new
    input:
      source: event_data.source
      description: event_data.desc
steps:
  - name: process
    type: bash
    command: echo "{{target}} {{source}} {{description}}"
`)
		issues := rule.Check(ast)

		// Should have issues for source and description, but at info severity
		var infoIssues, warningIssues []LintIssue
		for _, issue := range issues {
			if issue.Rule == "undefined-variable" {
				switch issue.Severity {
				case SeverityInfo:
					infoIssues = append(infoIssues, issue)
				case SeverityWarning:
					warningIssues = append(warningIssues, issue)
				}
			}
		}

		assert.Len(t, infoIssues, 2, "should have 2 info-level issues for trigger vars")
		assert.Empty(t, warningIssues, "should have no warning-level undefined-variable issues")

		// Verify messages mention trigger input
		for _, issue := range infoIssues {
			assert.Contains(t, issue.Message, "provided by trigger input")
		}
	})

	t.Run("legacy trigger input syntax downgraded to info", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: event-handler-legacy
kind: module
triggers:
  - name: on-webhook
    on: event
    enabled: true
    event:
      topic: webhook.received
    input:
      type: event_data
      field: url
      name: webhook_target
steps:
  - name: process
    type: bash
    command: echo "{{webhook_target}}"
`)
		issues := rule.Check(ast)

		var infoIssues, warningIssues []LintIssue
		for _, issue := range issues {
			if issue.Rule == "undefined-variable" {
				switch issue.Severity {
				case SeverityInfo:
					infoIssues = append(infoIssues, issue)
				case SeverityWarning:
					warningIssues = append(warningIssues, issue)
				}
			}
		}

		assert.Len(t, infoIssues, 1, "should have 1 info-level issue for legacy trigger var")
		assert.Empty(t, warningIssues, "should have no warning-level issues")
		assert.Contains(t, infoIssues[0].Message, "webhook_target")
	})

	t.Run("event envelope variables recognized in event-triggered workflow", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: event-handler-envelope
kind: module
triggers:
  - name: on-asset
    on: event
    enabled: true
    event:
      topic: assets.new
    input:
      source: event_data.source
steps:
  - name: process
    type: bash
    command: echo "{{EventTopic}} {{EventSource}} {{EventTimestamp}} {{EventData}} {{EventEnvelope}} {{EventDataType}}"
`)
		issues := rule.Check(ast)

		// Event envelope variables should NOT produce any undefined-variable issues
		var undefinedWarnings []LintIssue
		for _, issue := range issues {
			if issue.Rule == "undefined-variable" && issue.Severity == SeverityWarning {
				undefinedWarnings = append(undefinedWarnings, issue)
			}
		}
		assert.Empty(t, undefinedWarnings, "event envelope variables should be recognized as defined")
	})

	t.Run("truly undefined vars still warned alongside trigger vars", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: mixed-vars
kind: module
triggers:
  - name: on-event
    on: event
    enabled: true
    event:
      topic: test.topic
    input:
      source: event_data.source
steps:
  - name: process
    type: bash
    command: echo "{{source}} {{truly_undefined}}"
`)
		issues := rule.Check(ast)

		var infoIssues, warningIssues []LintIssue
		for _, issue := range issues {
			if issue.Rule == "undefined-variable" {
				switch issue.Severity {
				case SeverityInfo:
					infoIssues = append(infoIssues, issue)
				case SeverityWarning:
					warningIssues = append(warningIssues, issue)
				}
			}
		}

		// truly_undefined should be info (downgraded because event trigger exists)
		// source is in triggerVars so also info
		assert.Len(t, infoIssues, 2, "both source and truly_undefined should be info due to event trigger")
		assert.Empty(t, warningIssues, "no warnings expected when event trigger is present")
	})

	t.Run("no trigger means no downgrade", func(t *testing.T) {
		ast := parseTestWorkflow(t, `
name: no-triggers
kind: module
steps:
  - name: step1
    type: bash
    command: echo "{{source}}"
`)
		issues := rule.Check(ast)

		// source should be a normal warning, not info
		var warningIssues []LintIssue
		for _, issue := range issues {
			if issue.Rule == "undefined-variable" && issue.Severity == SeverityWarning {
				warningIssues = append(warningIssues, issue)
			}
		}
		assert.Len(t, warningIssues, 1)
		assert.Contains(t, warningIssues[0].Message, "source")
	})

	t.Run("event trigger downgrades unlisted vars to info", func(t *testing.T) {
		// Scenario: input only maps "target", but steps also use source, description,
		// asset_type — these should be SeverityInfo, not SeverityWarning.
		ast := parseTestWorkflow(t, `
name: event-handler-partial-input
kind: module
triggers:
  - name: on-new-asset
    on: event
    enabled: true
    event:
      topic: assets.new
    input:
      target: event_data.url
steps:
  - name: process
    type: bash
    command: echo "{{target}} {{source}} {{description}} {{asset_type}}"
`)
		issues := rule.Check(ast)

		var infoIssues, warningIssues []LintIssue
		for _, issue := range issues {
			if issue.Rule == "undefined-variable" {
				switch issue.Severity {
				case SeverityInfo:
					infoIssues = append(infoIssues, issue)
				case SeverityWarning:
					warningIssues = append(warningIssues, issue)
				}
			}
		}

		// target is a built-in variable so no issue is emitted
		// source, description, asset_type are NOT in triggerVars but hasEventTrigger → info ("may be provided by event data")
		assert.Len(t, infoIssues, 3, "source, description, asset_type should be info-level")
		assert.Empty(t, warningIssues, "no warnings when event trigger is present")

		// Verify the messages indicate possible event data
		for _, issue := range infoIssues {
			assert.Contains(t, issue.Message, "may be provided by event data")
		}
	})

	t.Run("no event trigger means undefined vars are warnings", func(t *testing.T) {
		// Workflow with a cron trigger (not event) — undefined vars should stay as warnings
		ast := parseTestWorkflow(t, `
name: cron-handler
kind: module
triggers:
  - name: nightly
    on: cron
    enabled: true
    schedule: "0 0 * * *"
steps:
  - name: process
    type: bash
    command: echo "{{source}} {{description}}"
`)
		issues := rule.Check(ast)

		var infoIssues, warningIssues []LintIssue
		for _, issue := range issues {
			if issue.Rule == "undefined-variable" {
				switch issue.Severity {
				case SeverityInfo:
					infoIssues = append(infoIssues, issue)
				case SeverityWarning:
					warningIssues = append(warningIssues, issue)
				}
			}
		}

		assert.Empty(t, infoIssues, "no info downgrades without event trigger")
		assert.Len(t, warningIssues, 2, "source and description should be warnings")
	})
}
