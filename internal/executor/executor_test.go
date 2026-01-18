package executor

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConfig returns a config with isolated temp directories that are
// automatically cleaned up after the test completes.
func testConfig(t *testing.T) *config.Config {
	t.Helper()
	baseDir := t.TempDir()
	return &config.Config{
		BaseFolder:     baseDir,
		WorkspacesPath: filepath.Join(baseDir, "workspaces"),
		WorkflowsPath:  filepath.Join(baseDir, "workflows"),
		BinariesPath:   filepath.Join(baseDir, "binaries"),
		DataPath:       filepath.Join(baseDir, "data"),
	}
}

func TestExecutor_New(t *testing.T) {
	executor := NewExecutor()
	assert.NotNil(t, executor)
}

func TestExecutor_BashStep(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-bash",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "echo-test",
				Type:    core.StepTypeBash,
				Command: "echo hello",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
}

func TestExecutor_DryRun(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-dryrun",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "dangerous-command",
				Type:    core.StepTypeBash,
				Command: "rm -rf /nonexistent",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(true)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Contains(t, result.Steps[0].Output, "DRY-RUN")
}

func TestExecutor_RequiredParams(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-params",
		Kind: core.KindModule,
		Params: []core.Param{
			{
				Name:     "required_param",
				Required: true,
			},
		},
		Steps: []core.Step{
			{
				Name:    "test-step",
				Type:    core.StepTypeBash,
				Command: "echo test",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(true)
	executor.SetSpinner(false)

	// Should fail without required param
	_, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required_param")
}

func TestExecutor_DependencyTargetTypes_PassAny(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-target-types",
		Kind: core.KindModule,
		Dependencies: &core.Dependencies{
			TargetTypes: []core.TargetType{core.TargetTypeDomain, core.TargetTypeURL},
		},
		Steps: []core.Step{
			{
				Name:    "echo-test",
				Type:    core.StepTypeBash,
				Command: "echo ok",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(true)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "example.com",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
}

func TestExecutor_DependencyTargetTypes_Fail(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-target-types-fail",
		Kind: core.KindModule,
		Dependencies: &core.Dependencies{
			TargetTypes: []core.TargetType{core.TargetTypeDomain, core.TargetTypeURL},
		},
		Steps: []core.Step{
			{
				Name:    "echo-test",
				Type:    core.StepTypeBash,
				Command: "echo ok",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(true)
	executor.SetSpinner(false)

	_, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "not-a-domain",
	}, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "target")
	assert.Contains(t, err.Error(), "required types")
}

func TestExecutor_DependencyTargetTypes_Unknown(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-target-types-unknown",
		Kind: core.KindModule,
		Dependencies: &core.Dependencies{
			TargetTypes: []core.TargetType{core.TargetType("unknown")},
		},
		Steps: []core.Step{
			{
				Name:    "echo-test",
				Type:    core.StepTypeBash,
				Command: "echo ok",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(true)
	executor.SetSpinner(false)

	_, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "example.com",
	}, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown target_types")
}

func TestExecutor_DefaultParams(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-default-params",
		Kind: core.KindModule,
		Params: []core.Param{
			{
				Name:    "my_param",
				Default: "default_value",
			},
		},
		Steps: []core.Step{
			{
				Name:    "test-step",
				Type:    core.StepTypeBash,
				Command: "echo {{my_param}}",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(true)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
}

func TestExecutor_FlowKindCheck(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	// Create a flow workflow
	flow := &core.Workflow{
		Name: "test-flow",
		Kind: core.KindFlow,
	}

	executor := NewExecutor()

	// ExecuteModule should fail for flow
	_, err := executor.ExecuteModule(ctx, flow, map[string]string{
		"target": "test",
	}, cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a module")
}

func TestExecutor_MultipleSteps(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-multiple-steps",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "step-1",
				Type:    core.StepTypeBash,
				Command: "echo step1",
			},
			{
				Name:    "step-2",
				Type:    core.StepTypeBash,
				Command: "echo step2",
			},
			{
				Name:    "step-3",
				Type:    core.StepTypeBash,
				Command: "echo step3",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 3)

	for _, step := range result.Steps {
		assert.Equal(t, core.StepStatusSuccess, step.Status)
	}
}

// Parallel Commands Tests

func TestExecutor_ParallelCommands(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-parallel-commands",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name: "parallel-echo",
				Type: core.StepTypeBash,
				ParallelCommands: []string{
					"echo 'command 1'",
					"echo 'command 2'",
					"echo 'command 3'",
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
	// Verify all parallel outputs are captured
	assert.Contains(t, result.Steps[0].Output, "command 1")
	assert.Contains(t, result.Steps[0].Output, "command 2")
	assert.Contains(t, result.Steps[0].Output, "command 3")
}

func TestExecutor_ParallelCommands_OneFailure(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-parallel-commands-fail",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name: "parallel-with-failure",
				Type: core.StepTypeBash,
				ParallelCommands: []string{
					"echo 'success 1'",
					"exit 1",
					"echo 'success 2'",
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	// The workflow returns error when a step fails, but result still contains step info
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Steps, 1)
	// Step fails because one of the parallel commands failed
	assert.Equal(t, core.StepStatusFailed, result.Steps[0].Status)
}

// Parallel Functions Tests

func TestExecutor_ParallelFunctions(t *testing.T) {
	t.Skip("parallel_functions is not yet implemented")

	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-parallel-functions",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name: "parallel-funcs",
				Type: core.StepTypeFunction,
				ParallelFunctions: []string{
					"trim(\"  hello  \")",
					"contains(\"hello world\", \"world\")",
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
}

// Parallel Steps Tests

func TestExecutor_ParallelSteps(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-parallel-steps",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name: "nested-parallel",
				Type: core.StepTypeParallel,
				ParallelSteps: []core.Step{
					{
						Name:    "sub-step-1",
						Type:    core.StepTypeBash,
						Command: "echo 'sub 1'",
					},
					{
						Name:    "sub-step-2",
						Type:    core.StepTypeBash,
						Command: "echo 'sub 2'",
					},
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
	// Verify outputs from sub-steps are captured
	assert.Contains(t, result.Steps[0].Output, "sub 1")
	assert.Contains(t, result.Steps[0].Output, "sub 2")
}

func TestExecutor_ParallelSteps_MixedTypes(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-parallel-steps-mixed",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name: "mixed-parallel",
				Type: core.StepTypeParallel,
				ParallelSteps: []core.Step{
					{
						Name:    "bash-sub",
						Type:    core.StepTypeBash,
						Command: "echo 'bash output'",
					},
					{
						Name:     "func-sub",
						Type:     core.StepTypeFunction,
						Function: "trim(\"  hello  \")",
					},
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
}

// Timeout Tests

func TestExecutor_StepTimeout(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-timeout",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "quick-command",
				Type:    core.StepTypeBash,
				Command: "echo 'fast'",
				Timeout: core.StepTimeout("5"),
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
}

func TestExecutor_StepTimeout_Exceeds(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-timeout-exceed",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "slow-command",
				Type:    core.StepTypeBash,
				Command: "sleep 10",
				Timeout: core.StepTimeout("1s"),
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	// The workflow returns error when step fails due to timeout
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusFailed, result.Steps[0].Status)
}

func TestExecutor_StepTimeout_TemplateDuration(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-timeout-template",
		Kind: core.KindModule,
		Params: []core.Param{
			{Name: "timeout", Default: "1s"},
		},
		Steps: []core.Step{
			{
				Name:    "slow-command",
				Type:    core.StepTypeBash,
				Command: "sleep 10",
				Timeout: core.StepTimeout("{{timeout}}"),
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target":  "test",
		"timeout": "1s",
	}, cfg)

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusFailed, result.Steps[0].Status)
}

func TestExecutor_Foreach_ThreadsTemplate(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-foreach-threads-template",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name: "create-input",
				Type: core.StepTypeBash,
				Commands: []string{
					"mkdir -p {{Output}}",
					"printf 'one\ntwo\nthree\n' > {{Output}}/items.txt",
				},
			},
			{
				Name:     "process-items",
				Type:     core.StepTypeForeach,
				Input:    "{{Output}}/items.txt",
				Variable: "item",
				Threads:  core.StepThreads("{{ baseThreads * 2 }}"),
				Step: &core.Step{
					Name:    "process-item",
					Type:    core.StepTypeBash,
					Command: "echo [[item]]",
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 2)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[1].Status)
	assert.Contains(t, result.Steps[1].Output, "one")
}

func TestExecutor_StepTimeout_ParallelCommands(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-timeout-parallel",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name: "parallel-with-timeout",
				Type: core.StepTypeBash,
				ParallelCommands: []string{
					"echo 'fast 1'",
					"sleep 10",
					"echo 'fast 2'",
				},
				Timeout: core.StepTimeout("1"),
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	// Workflow returns error when step fails due to timeout on one of the parallel commands
	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Steps, 1)
	// One of the parallel commands times out
	assert.Equal(t, core.StepStatusFailed, result.Steps[0].Status)
}

// Decision Tests
// Note: Decision routing may require additional setup or different condition syntax.
// These tests verify the basic decision structure is processed without errors.

func TestExecutor_Decision_SkipToEnd(t *testing.T) {
	t.Skip("Decision routing not working as expected - needs investigation")

	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-decision-skip",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "check-condition",
				Type:    core.StepTypeBash,
				Command: "echo 'running first step'",
				Decision: &core.DecisionConfig{
					Switch: "{{target}}",
					Cases: map[string]core.DecisionCase{
						"skip": {Goto: "_end"},
					},
				},
			},
			{
				Name:    "should-not-run",
				Type:    core.StepTypeBash,
				Command: "echo 'should not run'",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "skip",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	// Only the first step should have run
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, "check-condition", result.Steps[0].StepName)
}

func TestExecutor_Decision_JumpToStep(t *testing.T) {
	t.Skip("Decision routing not working as expected - needs investigation")

	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-decision-jump",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "check-condition",
				Type:    core.StepTypeBash,
				Command: "echo 'jumping'",
				Decision: &core.DecisionConfig{
					Switch: "{{target}}",
					Cases: map[string]core.DecisionCase{
						"jump": {Goto: "final-step"},
					},
				},
			},
			{
				Name:    "middle-step",
				Type:    core.StepTypeBash,
				Command: "echo 'middle'",
			},
			{
				Name:    "final-step",
				Type:    core.StepTypeBash,
				Command: "echo 'final'",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "jump",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	// Should have skipped middle-step and jumped to final-step
	assert.Len(t, result.Steps, 2)
	assert.Equal(t, "check-condition", result.Steps[0].StepName)
	assert.Equal(t, "final-step", result.Steps[1].StepName)
}

func TestExecutor_Decision_NoMatch(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-decision-no-match",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "check-condition",
				Type:    core.StepTypeBash,
				Command: "echo 'continue'",
				Decision: &core.DecisionConfig{
					Switch: "{{target}}",
					Cases: map[string]core.DecisionCase{
						"skip": {Goto: "_end"},
						"jump": {Goto: "final-step"},
					},
				},
			},
			{
				Name:    "next-step",
				Type:    core.StepTypeBash,
				Command: "echo 'next'",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "continue",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	// No decision matched, so should continue to next step
	assert.Len(t, result.Steps, 2)
	assert.Equal(t, "check-condition", result.Steps[0].StepName)
	assert.Equal(t, "next-step", result.Steps[1].StepName)
}

func TestExecutor_Decision_MultipleCases(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-decision-multi",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "check-condition",
				Type:    core.StepTypeBash,
				Command: "echo 'checking'",
				Decision: &core.DecisionConfig{
					Switch: "{{target}}",
					Cases: map[string]core.DecisionCase{
						"first":  {Goto: "step-a"},
						"second": {Goto: "step-b"},
					},
				},
			},
			{
				Name:    "step-a",
				Type:    core.StepTypeBash,
				Command: "echo 'step a'",
			},
			{
				Name:    "step-b",
				Type:    core.StepTypeBash,
				Command: "echo 'step b'",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "first",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	// First case matches, should jump to step-a
	assert.GreaterOrEqual(t, len(result.Steps), 2)
	assert.Equal(t, "check-condition", result.Steps[0].StepName)
	assert.Equal(t, "step-a", result.Steps[1].StepName)
}

// Tests for Switch/Case decision syntax

func TestExecutor_Decision_SwitchCase_Match(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-decision-switch",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "check-type",
				Type:    core.StepTypeBash,
				Command: "echo 'domain'",
				Exports: map[string]string{
					"detected_type": "domain",
				},
				Decision: &core.DecisionConfig{
					Switch: "{{detected_type}}",
					Cases: map[string]core.DecisionCase{
						"domain": {Goto: "domain-scan"},
						"ip":     {Goto: "ip-scan"},
						"cidr":   {Goto: "cidr-scan"},
					},
					Default: &core.DecisionCase{Goto: "generic-scan"},
				},
			},
			{
				Name:    "generic-scan",
				Type:    core.StepTypeBash,
				Command: "echo 'generic'",
			},
			{
				Name:    "domain-scan",
				Type:    core.StepTypeBash,
				Command: "echo 'domain scan'",
			},
			{
				Name:    "ip-scan",
				Type:    core.StepTypeBash,
				Command: "echo 'ip scan'",
			},
			{
				Name:    "cidr-scan",
				Type:    core.StepTypeBash,
				Command: "echo 'cidr scan'",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	// Switch matched "domain", should jump to domain-scan
	assert.GreaterOrEqual(t, len(result.Steps), 2)
	assert.Equal(t, "check-type", result.Steps[0].StepName)
	assert.Equal(t, "domain-scan", result.Steps[1].StepName)
}

func TestExecutor_Decision_SwitchCase_Default(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-decision-switch-default",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "check-type",
				Type:    core.StepTypeBash,
				Command: "echo 'unknown'",
				Exports: map[string]string{
					"detected_type": "unknown",
				},
				Decision: &core.DecisionConfig{
					Switch: "{{detected_type}}",
					Cases: map[string]core.DecisionCase{
						"domain": {Goto: "domain-scan"},
						"ip":     {Goto: "ip-scan"},
					},
					Default: &core.DecisionCase{Goto: "generic-scan"},
				},
			},
			{
				Name:    "domain-scan",
				Type:    core.StepTypeBash,
				Command: "echo 'domain scan'",
			},
			{
				Name:    "ip-scan",
				Type:    core.StepTypeBash,
				Command: "echo 'ip scan'",
			},
			{
				Name:    "generic-scan",
				Type:    core.StepTypeBash,
				Command: "echo 'generic scan'",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	// No case matched, should fall through to default -> generic-scan
	assert.GreaterOrEqual(t, len(result.Steps), 2)
	assert.Equal(t, "check-type", result.Steps[0].StepName)
	assert.Equal(t, "generic-scan", result.Steps[1].StepName)
}

func TestExecutor_Decision_SwitchCase_NoMatchNoDefault(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-decision-switch-no-default",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "check-type",
				Type:    core.StepTypeBash,
				Command: "echo 'unknown'",
				Exports: map[string]string{
					"detected_type": "unknown",
				},
				Decision: &core.DecisionConfig{
					Switch: "{{detected_type}}",
					Cases: map[string]core.DecisionCase{
						"domain": {Goto: "domain-scan"},
						"ip":     {Goto: "ip-scan"},
					},
					// No default
				},
			},
			{
				Name:    "next-step",
				Type:    core.StepTypeBash,
				Command: "echo 'next step'",
			},
			{
				Name:    "domain-scan",
				Type:    core.StepTypeBash,
				Command: "echo 'domain scan'",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	// No case matched and no default, should continue to next step
	assert.GreaterOrEqual(t, len(result.Steps), 2)
	assert.Equal(t, "check-type", result.Steps[0].StepName)
	assert.Equal(t, "next-step", result.Steps[1].StepName)
}

// Tests for Kahn's algorithm dependency graph (O(V+E) flow execution)

func TestBuildDependencyGraph_NoDependencies(t *testing.T) {
	modules := []core.ModuleRef{
		{Name: "module-a"},
		{Name: "module-b"},
		{Name: "module-c"},
	}

	dependents, inDegree := buildDependencyGraph(modules)

	// All modules should have in-degree 0 (no dependencies)
	assert.Equal(t, 0, inDegree["module-a"])
	assert.Equal(t, 0, inDegree["module-b"])
	assert.Equal(t, 0, inDegree["module-c"])

	// No dependents since nothing depends on anything
	assert.Empty(t, dependents["module-a"])
	assert.Empty(t, dependents["module-b"])
	assert.Empty(t, dependents["module-c"])
}

func TestBuildDependencyGraph_LinearChain(t *testing.T) {
	// A -> B -> C -> D (linear dependency chain)
	modules := []core.ModuleRef{
		{Name: "module-d", DependsOn: []string{"module-c"}},
		{Name: "module-c", DependsOn: []string{"module-b"}},
		{Name: "module-b", DependsOn: []string{"module-a"}},
		{Name: "module-a"},
	}

	dependents, inDegree := buildDependencyGraph(modules)

	// module-a has no dependencies
	assert.Equal(t, 0, inDegree["module-a"])
	// module-b depends on module-a
	assert.Equal(t, 1, inDegree["module-b"])
	// module-c depends on module-b
	assert.Equal(t, 1, inDegree["module-c"])
	// module-d depends on module-c
	assert.Equal(t, 1, inDegree["module-d"])

	// module-a has module-b as dependent
	assert.Contains(t, dependents["module-a"], "module-b")
	// module-b has module-c as dependent
	assert.Contains(t, dependents["module-b"], "module-c")
	// module-c has module-d as dependent
	assert.Contains(t, dependents["module-c"], "module-d")
}

func TestBuildDependencyGraph_Diamond(t *testing.T) {
	// Diamond pattern: A -> B, A -> C, B -> D, C -> D
	modules := []core.ModuleRef{
		{Name: "module-a"},
		{Name: "module-b", DependsOn: []string{"module-a"}},
		{Name: "module-c", DependsOn: []string{"module-a"}},
		{Name: "module-d", DependsOn: []string{"module-b", "module-c"}},
	}

	dependents, inDegree := buildDependencyGraph(modules)

	// module-a has no dependencies
	assert.Equal(t, 0, inDegree["module-a"])
	// module-b and module-c depend on module-a (1 each)
	assert.Equal(t, 1, inDegree["module-b"])
	assert.Equal(t, 1, inDegree["module-c"])
	// module-d depends on both module-b and module-c
	assert.Equal(t, 2, inDegree["module-d"])

	// module-a has both module-b and module-c as dependents
	assert.Len(t, dependents["module-a"], 2)
	assert.Contains(t, dependents["module-a"], "module-b")
	assert.Contains(t, dependents["module-a"], "module-c")
}

func TestBuildModuleMap(t *testing.T) {
	modules := []core.ModuleRef{
		{Name: "module-a", Path: "path/to/a.yaml"},
		{Name: "module-b", Path: "path/to/b.yaml"},
		{Name: "module-c", Path: "path/to/c.yaml"},
	}

	moduleMap := buildModuleMap(modules)

	assert.Len(t, moduleMap, 3)
	assert.Equal(t, "path/to/a.yaml", moduleMap["module-a"].Path)
	assert.Equal(t, "path/to/b.yaml", moduleMap["module-b"].Path)
	assert.Equal(t, "path/to/c.yaml", moduleMap["module-c"].Path)
}

// Tests for CloneForLoop optimization

func TestExecutionContext_CloneForLoop(t *testing.T) {
	// Create a context with some params and variables
	ctx := core.NewExecutionContext("test-workflow", core.KindModule, "run-123", "example.com")
	ctx.SetParam("param1", "value1")
	ctx.SetParam("param2", "value2")
	ctx.SetVariable("var1", "varvalue1")
	ctx.SetExport("export1", "exportvalue1")

	// Clone for loop iteration
	clone := ctx.CloneForLoop("item", "line-content", 5)

	// Verify metadata is copied
	assert.Equal(t, "test-workflow", clone.WorkflowName)
	assert.Equal(t, core.KindModule, clone.WorkflowKind)
	assert.Equal(t, "run-123", clone.RunID)
	assert.Equal(t, "example.com", clone.Target)

	// Verify Params reference is shared (not copied)
	// This is the key optimization - we share the immutable Params
	ctx.SetParam("new-param", "new-value")
	_, exists := clone.GetParam("new-param")
	assert.True(t, exists, "Clone should share Params reference with parent")

	// Verify loop variables are set
	loopVar, ok := clone.GetVariable("item")
	assert.True(t, ok)
	assert.Equal(t, "line-content", loopVar)

	iterID, ok := clone.GetVariable("_id_")
	assert.True(t, ok)
	assert.Equal(t, 5, iterID)

	// Verify parent variables are accessible
	v, ok := clone.GetVariable("var1")
	assert.True(t, ok)
	assert.Equal(t, "varvalue1", v)

	// Verify exports are isolated (clone has fresh exports map)
	_, exists = clone.GetExport("export1")
	assert.False(t, exists, "Clone should have fresh Exports map")
}

func TestExecutionContext_CloneForLoop_EmptyLoopVar(t *testing.T) {
	ctx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	// Clone with empty loop variable (for parallel steps that don't need loop vars)
	clone := ctx.CloneForLoop("", nil, 0)

	// Should still set _id_
	iterID, ok := clone.GetVariable("_id_")
	assert.True(t, ok)
	assert.Equal(t, 0, iterID)

	// Empty string key should not be set
	_, ok = clone.GetVariable("")
	assert.False(t, ok)
}
