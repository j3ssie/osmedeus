package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestWorkflows(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create modules directory
	modulesDir := filepath.Join(tmpDir, "modules")
	err := os.MkdirAll(modulesDir, 0755)
	require.NoError(t, err)

	// Create a test flow
	flowContent := `kind: flow
name: test-flow
description: Test flow for unit testing

params:
  - name: target
    required: true

modules:
  - name: test-module
    path: modules/test-module.yaml
`
	err = os.WriteFile(filepath.Join(tmpDir, "test-flow.yaml"), []byte(flowContent), 0644)
	require.NoError(t, err)

	// Create a test module
	moduleContent := `kind: module
name: test-module
description: Test module for unit testing

params:
  - name: target
    required: true
  - name: threads
    default: "10"

steps:
  - name: echo-test
    type: bash
    command: echo "Hello {{target}}"
`
	err = os.WriteFile(filepath.Join(modulesDir, "test-module.yaml"), []byte(moduleContent), 0644)
	require.NoError(t, err)

	// Create another module with triggers
	moduleWithTriggersContent := `kind: module
name: triggered-module
description: Module with triggers

triggers:
  - name: manual
    on: manual
    enabled: true
  - name: cron-daily
    on: cron
    schedule: "0 0 * * *"
    enabled: true

params:
  - name: target
    required: true

steps:
  - name: run-task
    type: bash
    command: echo "Running on {{target}}"
`
	err = os.WriteFile(filepath.Join(modulesDir, "triggered-module.yaml"), []byte(moduleWithTriggersContent), 0644)
	require.NoError(t, err)

	return tmpDir
}

func TestLoader_ListFlows(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	flows, err := loader.ListFlows()
	require.NoError(t, err)
	assert.Contains(t, flows, "test-flow")
}

func TestLoader_ListModules(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	modules, err := loader.ListModules()
	require.NoError(t, err)
	assert.Len(t, modules, 2)
	assert.Contains(t, modules, "test-module")
	assert.Contains(t, modules, "triggered-module")
}

func TestLoader_LoadWorkflow_Flow(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	workflow, err := loader.LoadWorkflow("test-flow")
	require.NoError(t, err)

	assert.Equal(t, "test-flow", workflow.Name)
	assert.Equal(t, core.KindFlow, workflow.Kind)
	assert.Equal(t, "Test flow for unit testing", workflow.Description)
	assert.True(t, workflow.IsFlow())
	assert.False(t, workflow.IsModule())
	assert.Len(t, workflow.Modules, 1)
}

func TestLoader_LoadWorkflow_Module(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	workflow, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)

	assert.Equal(t, "test-module", workflow.Name)
	assert.Equal(t, core.KindModule, workflow.Kind)
	assert.True(t, workflow.IsModule())
	assert.False(t, workflow.IsFlow())
	assert.Len(t, workflow.Steps, 1)
	assert.Equal(t, "echo-test", workflow.Steps[0].Name)
	assert.Equal(t, core.StepTypeBash, workflow.Steps[0].Type)
}

func TestLoader_LoadWorkflow_WithTriggers(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	workflow, err := loader.LoadWorkflow("triggered-module")
	require.NoError(t, err)

	assert.Equal(t, "triggered-module", workflow.Name)
	assert.True(t, workflow.HasTriggers())
	assert.Len(t, workflow.Triggers, 2)

	// Check manual trigger
	var manualTrigger *core.Trigger
	for i := range workflow.Triggers {
		if workflow.Triggers[i].Name == "manual" {
			manualTrigger = &workflow.Triggers[i]
			break
		}
	}
	require.NotNil(t, manualTrigger)
	assert.Equal(t, core.TriggerManual, manualTrigger.On)
	assert.True(t, manualTrigger.Enabled)
}

func TestLoader_LoadWorkflow_NotFound(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	_, err := loader.LoadWorkflow("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoader_LoadAllWorkflows(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	workflows, err := loader.LoadAllWorkflows()
	require.NoError(t, err)

	assert.Len(t, workflows, 3) // 1 flow + 2 modules

	// Check we have both flows and modules
	var flowCount, moduleCount int
	for _, w := range workflows {
		if w.IsFlow() {
			flowCount++
		} else if w.IsModule() {
			moduleCount++
		}
	}
	assert.Equal(t, 1, flowCount)
	assert.Equal(t, 2, moduleCount)
}

func TestLoader_GetRequiredParams(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	workflow, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)

	required := workflow.GetRequiredParams()
	assert.Len(t, required, 1)
	assert.Equal(t, "target", required[0].Name)
}

func TestLoader_CacheWorkflow(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	// First load
	workflow1, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)

	// Second load should come from cache
	workflow2, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)

	assert.Same(t, workflow1, workflow2)
}

func TestLoader_ClearCache(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	// Load workflow
	workflow1, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)

	// Clear cache
	loader.ClearCache()

	// Load again - should be different instance
	workflow2, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)

	assert.NotSame(t, workflow1, workflow2)
	assert.Equal(t, workflow1.Name, workflow2.Name)
}

func TestLoader_GetAllCached(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	// Initially empty
	cached := loader.GetAllCached()
	assert.Len(t, cached, 0)

	// Load some workflows
	_, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)
	_, err = loader.LoadWorkflow("test-flow")
	require.NoError(t, err)

	// Should now have 2 cached
	cached = loader.GetAllCached()
	assert.Len(t, cached, 2)
}

func TestLoader_IsManualExecutionAllowed(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	// Workflow with manual trigger enabled
	workflow, err := loader.LoadWorkflow("triggered-module")
	require.NoError(t, err)
	assert.True(t, workflow.IsManualExecutionAllowed())

	// Workflow without triggers - manual allowed by default
	workflow, err = loader.LoadWorkflow("test-module")
	require.NoError(t, err)
	assert.True(t, workflow.IsManualExecutionAllowed())
}

func TestLoader_LoadWorkflowByPath_ExplicitLocalPath(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	// Create a workflow file in a separate location (not in workflows dir)
	localDir := t.TempDir()
	localWorkflow := filepath.Join(localDir, "local-module.yaml")
	content := `kind: module
name: local-module
description: Local module
steps:
  - name: test
    type: bash
    command: echo "test"
`
	err := os.WriteFile(localWorkflow, []byte(content), 0644)
	require.NoError(t, err)

	// Change to localDir to test relative path
	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	require.NoError(t, os.Chdir(localDir))

	// Test: ./local-module.yaml should load from CWD
	workflow, err := loader.LoadWorkflowByPath("./local-module.yaml")
	require.NoError(t, err)
	assert.Equal(t, "local-module", workflow.Name)

	// Test: ./nonexistent.yaml should error, not fall back
	_, err = loader.LoadWorkflowByPath("./nonexistent.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.NotContains(t, err.Error(), "workflows") // Should not mention fallback

	// Test: ../nonexistent.yaml should also error, not fall back
	_, err = loader.LoadWorkflowByPath("../nonexistent.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoader_CacheMtimeInvalidation(t *testing.T) {
	tmpDir := setupTestWorkflows(t)
	loader := NewLoader(tmpDir)

	modulePath := filepath.Join(tmpDir, "modules", "test-module.yaml")

	// First load
	workflow1, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)
	assert.Equal(t, "test-module", workflow1.Name)
	assert.Len(t, workflow1.Steps, 1)

	// Second load should come from cache (same instance)
	workflow2, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)
	assert.Same(t, workflow1, workflow2)

	// Modify the file - add a new step
	modifiedContent := `kind: module
name: test-module
description: Test module for unit testing (MODIFIED)

params:
  - name: target
    required: true
  - name: threads
    default: "10"

steps:
  - name: echo-test
    type: bash
    command: echo "Hello {{target}}"
  - name: new-step
    type: bash
    command: echo "New step"
`
	err = os.WriteFile(modulePath, []byte(modifiedContent), 0644)
	require.NoError(t, err)

	// Third load should detect the change and re-parse (different instance)
	workflow3, err := loader.LoadWorkflow("test-module")
	require.NoError(t, err)
	assert.NotSame(t, workflow1, workflow3)
	assert.Equal(t, "Test module for unit testing (MODIFIED)", workflow3.Description)
	assert.Len(t, workflow3.Steps, 2)
	assert.Equal(t, "new-step", workflow3.Steps[1].Name)
}
