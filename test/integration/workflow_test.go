package integration

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestdataPath returns the absolute path to the testdata directory
func getTestdataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "testdata")
}

// getWorkflowsPath returns the path to the workflows testdata directory
func getWorkflowsPath() string {
	return filepath.Join(getTestdataPath(), "workflows")
}

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

// TestLoadAllWorkflows tests that all workflow YAML files can be loaded and parsed
func TestLoadAllWorkflows(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	// Skip files that use experimental features or have validation issues
	skipFiles := map[string]string{
		"test-remote-bash.yaml":           "uses remote-bash step type (requires Docker)",
		"test-remote-bash-ssh.yaml":       "uses remote-bash step type (requires SSH)",
		"test-remote-bash-docker.yaml":    "uses remote-bash step type (requires Docker)",
		"test-docker-file-outputs.yaml":   "uses remote-bash step type (requires Docker)",
		"test-agent-validation-fail.yaml": "intentionally invalid (duplicate agent tools)",
		"test-agent-unknown-preset.yaml":  "intentionally invalid (unknown preset tool)",
	}

	// Get all workflow files (top-level + agent-and-llm subdirectory)
	files, err := filepath.Glob(filepath.Join(workflowsPath, "*.yaml"))
	require.NoError(t, err)
	subFiles, err := filepath.Glob(filepath.Join(workflowsPath, "agent-and-llm", "*.yaml"))
	require.NoError(t, err)
	files = append(files, subFiles...)
	require.Greater(t, len(files), 0, "No workflow files found")

	t.Logf("Found %d workflow files to load", len(files))

	for _, file := range files {
		name := filepath.Base(file)
		if reason, skip := skipFiles[name]; skip {
			t.Run(name, func(t *testing.T) {
				t.Skipf("Skipping: %s", reason)
			})
			continue
		}
		t.Run(name, func(t *testing.T) {
			workflow, err := loader.LoadWorkflowByPath(file)
			require.NoError(t, err, "Failed to load workflow: %s", file)
			assert.NotEmpty(t, workflow.Name, "Workflow name should not be empty")
			assert.NotEmpty(t, workflow.Kind, "Workflow kind should not be empty")
		})
	}
}

// TestValidateAllWorkflows tests that all workflow YAML files pass validation
func TestValidateAllWorkflows(t *testing.T) {
	workflowsPath := getWorkflowsPath()

	// Get all workflow files (top-level + agent-and-llm subdirectory)
	files, err := filepath.Glob(filepath.Join(workflowsPath, "*.yaml"))
	require.NoError(t, err)
	subFiles, err := filepath.Glob(filepath.Join(workflowsPath, "agent-and-llm", "*.yaml"))
	require.NoError(t, err)
	files = append(files, subFiles...)

	// Skip validation test files that are meant to fail or use experimental features
	skipFiles := map[string]string{
		"test-requirements-fail.yaml":     "expected to fail validation",
		"test-remote-bash.yaml":           "uses remote-bash step type",
		"test-remote-bash-ssh.yaml":       "uses remote-bash step type",
		"test-remote-bash-docker.yaml":    "uses remote-bash step type",
		"test-docker-file-outputs.yaml":   "uses remote-bash step type (requires Docker)",
		"test-agent-validation-fail.yaml": "intentionally invalid (duplicate agent tools)",
		"test-agent-unknown-preset.yaml":  "intentionally invalid (unknown preset tool)",
	}

	for _, file := range files {
		name := filepath.Base(file)
		if reason, skip := skipFiles[name]; skip {
			t.Run(name, func(t *testing.T) {
				t.Skipf("Skipping: %s", reason)
			})
			continue
		}

		t.Run(name, func(t *testing.T) {
			p := parser.NewParser()
			workflow, err := p.Parse(file)
			require.NoError(t, err, "Failed to parse workflow: %s", file)

			err = parser.Validate(workflow)
			require.NoError(t, err, "Validation failed for workflow: %s", file)
		})
	}
}

// TestExecuteBashWorkflow tests executing a basic bash workflow
func TestExecuteBashWorkflow(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-bash")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "integration-test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
}

// TestExecuteForeachWorkflow tests executing a foreach loop workflow
func TestExecuteForeachWorkflow(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-foreach")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "foreach-test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)

	// Should have 4 steps: create-input, process-items (foreach), verify-output, cleanup
	assert.Len(t, result.Steps, 4)

	// All steps should succeed
	for _, step := range result.Steps {
		assert.Equal(t, core.StepStatusSuccess, step.Status, "Step %s failed", step.StepName)
	}
}

// TestExecuteParallelCommandsWorkflow tests executing parallel commands
func TestExecuteParallelCommandsWorkflow(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-parallel-commands")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "parallel-test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 2)

	// All steps should succeed
	for _, step := range result.Steps {
		assert.Equal(t, core.StepStatusSuccess, step.Status, "Step %s failed", step.StepName)
	}
}

// TestExecuteParallelStepsWorkflow tests executing parallel steps
func TestExecuteParallelStepsWorkflow(t *testing.T) {
	// TODO: Skip this test until export evaluation is fixed for function steps
	// The workflow uses a function step with exports that requires 'output' variable
	t.Skip("Skipping: function step export evaluation needs fixing")

	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-parallel-steps")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "parallel-steps-test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
}

// TestExecuteFunctionsWorkflow tests executing utility functions
func TestExecuteFunctionsWorkflow(t *testing.T) {
	// TODO: Skip this test until export evaluation is fixed for function steps
	// The issue is that function steps don't properly set 'output' variable for export evaluation
	t.Skip("Skipping: function step export evaluation needs fixing")

	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-functions")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "functions-test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)

	// All 4 steps should complete successfully
	assert.Len(t, result.Steps, 4)
	for _, step := range result.Steps {
		assert.Equal(t, core.StepStatusSuccess, step.Status, "Step %s failed", step.StepName)
	}
}

// TestTimeoutWorkflowSuccess tests workflow with timeout that succeeds
func TestTimeoutWorkflowSuccess(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-timeout")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "timeout-test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)

	// Both steps should succeed within timeout
	assert.Len(t, result.Steps, 2)
	for _, step := range result.Steps {
		assert.Equal(t, core.StepStatusSuccess, step.Status, "Step %s failed", step.StepName)
	}
}

// TestTimeoutWorkflowExceeds tests workflow where step exceeds timeout
func TestTimeoutWorkflowExceeds(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-timeout-exceed")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "timeout-exceed-test",
	}, cfg)

	// Execution should fail due to timeout
	require.Error(t, err)
	assert.Equal(t, core.RunStatusFailed, result.Status)
}

// TestRequirementsWorkflowSuccess tests workflow with satisfied dependencies
func TestRequirementsWorkflowSuccess(t *testing.T) {
	workflowsPath := getWorkflowsPath()

	p := parser.NewParser()
	workflow, err := p.Parse(filepath.Join(workflowsPath, "test-requirements.yaml"))
	require.NoError(t, err)

	// Check dependencies
	depChecker := parser.NewDependencyChecker()
	if workflow.Dependencies != nil {
		err = depChecker.CheckCommands(workflow.Dependencies.Commands, "")
		require.NoError(t, err, "Dependency check should pass for common commands like echo, cat")
	}
}

// TestRequirementsWorkflowFail tests workflow with missing dependencies
func TestRequirementsWorkflowFail(t *testing.T) {
	workflowsPath := getWorkflowsPath()

	p := parser.NewParser()
	workflow, err := p.Parse(filepath.Join(workflowsPath, "test-requirements-fail.yaml"))
	require.NoError(t, err)

	// Check dependencies - should fail for nonexistent commands
	depChecker := parser.NewDependencyChecker()
	if workflow.Dependencies != nil {
		err = depChecker.CheckCommands(workflow.Dependencies.Commands, "")
		require.Error(t, err, "Dependency check should fail for nonexistent commands")
	}
}

// TestLoadComplexWorkflows tests loading flow-type workflows
func TestLoadComplexWorkflows(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-flow")
	require.NoError(t, err)

	assert.Equal(t, "test-flow", workflow.Name)
	assert.Equal(t, core.KindFlow, workflow.Kind)
	assert.True(t, workflow.IsFlow())
	assert.Greater(t, len(workflow.Modules), 0, "Flow should have at least one module")
}

// TestListWorkflowsByKind tests listing workflows categorized by kind
func TestListWorkflowsByKind(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	flows, modules, err := loader.ListAllWorkflows()
	require.NoError(t, err)

	t.Logf("Found %d flows and %d modules", len(flows), len(modules))

	assert.Greater(t, len(flows), 0, "Expected at least one flow in workflows directory")
	assert.Greater(t, len(modules), 0, "Expected at least one module")
}

// TestDryRunExecution tests dry-run mode for workflow execution
func TestDryRunExecution(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-bash")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(true)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "dry-run-test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	// In dry-run mode, output should indicate dry-run
	assert.Contains(t, result.Steps[0].Output, "DRY-RUN")
}

// TestMissingRequiredParam tests that execution fails with missing required params
func TestMissingRequiredParam(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-bash")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	// Execute without required 'target' param
	_, err = exec.ExecuteModule(ctx, workflow, map[string]string{}, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "target")
}

// TestWorkflowCaching tests that workflow caching works correctly
func TestWorkflowCaching(t *testing.T) {
	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	// First load
	workflow1, err := loader.LoadWorkflow("test-bash")
	require.NoError(t, err)

	// Second load should return cached version (same pointer)
	workflow2, err := loader.LoadWorkflow("test-bash")
	require.NoError(t, err)

	assert.Same(t, workflow1, workflow2, "Expected same cached instance")

	// Clear cache and reload
	loader.ClearCache()

	workflow3, err := loader.LoadWorkflow("test-bash")
	require.NoError(t, err)

	assert.NotSame(t, workflow1, workflow3, "Expected different instance after cache clear")
}

// TestDecisionWorkflow tests decision/conditional routing
func TestDecisionWorkflow(t *testing.T) {
	// TODO: Skip this test until export evaluation is fixed for decision steps
	// The issue is that decision steps don't properly set 'output' variable for export evaluation
	t.Skip("Skipping: decision step export evaluation needs fixing")

	workflowsPath := getWorkflowsPath()

	// Check if decision workflow exists
	decisionPath := filepath.Join(workflowsPath, "test-decision.yaml")
	if _, err := os.Stat(decisionPath); os.IsNotExist(err) {
		t.Skip("test-decision.yaml not found")
	}

	loader := parser.NewLoader(workflowsPath)
	workflow, err := loader.LoadWorkflow("test-decision")
	require.NoError(t, err)

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target": "continue",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
}
