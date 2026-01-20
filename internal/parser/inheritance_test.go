package parser

import (
	"path/filepath"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testWorkflowsDir = "../../test/testdata/workflows/extends"

func TestInheritanceResolver_SimpleExtends(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "child-simple.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	// Check that child name and description are preserved
	assert.Equal(t, "child-simple", workflow.Name)
	assert.Equal(t, "Simple child that only overrides params", workflow.Description)

	// Check that inheritance was tracked
	assert.Equal(t, "base-module", workflow.ResolvedFrom)

	// Check that parent steps were inherited
	assert.Len(t, workflow.Steps, 3)
	assert.Equal(t, "step-one", workflow.Steps[0].Name)
	assert.Equal(t, "step-two", workflow.Steps[1].Name)
	assert.Equal(t, "step-three", workflow.Steps[2].Name)
}

func TestInheritanceResolver_ParamOverride(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "child-simple.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	// Find the overridden params
	paramMap := make(map[string]core.Param)
	for _, p := range workflow.Params {
		paramMap[p.Name] = p
	}

	// threads should be overridden to "5"
	threadsParam, ok := paramMap["threads"]
	require.True(t, ok, "threads param should exist")
	assert.Equal(t, "5", threadsParam.DefaultString())

	// timeout should be overridden to "1800"
	timeoutParam, ok := paramMap["timeout"]
	require.True(t, ok, "timeout param should exist")
	assert.Equal(t, "1800", timeoutParam.DefaultString())

	// target should be inherited from parent
	targetParam, ok := paramMap["target"]
	require.True(t, ok, "target param should exist")
	assert.True(t, targetParam.Required)
}

func TestInheritanceResolver_StepsAppend(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "child-override-steps.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	// Should have 4 steps (3 from parent + 1 appended)
	assert.Len(t, workflow.Steps, 4)

	// First 3 steps from parent
	assert.Equal(t, "step-one", workflow.Steps[0].Name)
	assert.Equal(t, "step-two", workflow.Steps[1].Name)
	assert.Equal(t, "step-three", workflow.Steps[2].Name)

	// Last step appended by child
	assert.Equal(t, "step-four", workflow.Steps[3].Name)
	assert.Contains(t, workflow.Steps[3].Command, "Step four added by child")
}

func TestInheritanceResolver_StepsPrepend(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "child-prepend-steps.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	// Should have 4 steps (1 prepended + 3 from parent)
	assert.Len(t, workflow.Steps, 4)

	// First step prepended by child
	assert.Equal(t, "step-zero", workflow.Steps[0].Name)
	assert.Contains(t, workflow.Steps[0].Command, "Step zero prepended")

	// Remaining steps from parent
	assert.Equal(t, "step-one", workflow.Steps[1].Name)
	assert.Equal(t, "step-two", workflow.Steps[2].Name)
	assert.Equal(t, "step-three", workflow.Steps[3].Name)
}

func TestInheritanceResolver_StepsReplace(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "child-replace-steps.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	// Should have only 2 steps (completely replaced)
	assert.Len(t, workflow.Steps, 2)

	// Both steps are new
	assert.Equal(t, "new-step-one", workflow.Steps[0].Name)
	assert.Equal(t, "new-step-two", workflow.Steps[1].Name)
}

func TestInheritanceResolver_StepsMerge(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "child-merge-steps.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	// Should have 3 steps: step-one (kept), step-two (replaced), step-new (added)
	// step-three was removed
	assert.Len(t, workflow.Steps, 3)

	// Build step map for easier checking
	stepMap := make(map[string]core.Step)
	for _, s := range workflow.Steps {
		stepMap[s.Name] = s
	}

	// step-one should be unchanged
	stepOne, ok := stepMap["step-one"]
	require.True(t, ok, "step-one should exist")
	assert.Contains(t, stepOne.Command, "Step one")

	// step-two should be replaced
	stepTwo, ok := stepMap["step-two"]
	require.True(t, ok, "step-two should exist")
	assert.Contains(t, stepTwo.Command, "replaced with new command")

	// step-three should be removed
	_, ok = stepMap["step-three"]
	assert.False(t, ok, "step-three should be removed")

	// step-new should be added
	stepNew, ok := stepMap["step-new"]
	require.True(t, ok, "step-new should exist")
	assert.Contains(t, stepNew.Command, "New step added")
}

func TestInheritanceResolver_CircularDetection(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	_, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "circular-a.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestInheritanceResolver_DeepChain(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "chain-c.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	assert.Equal(t, "chain-c", workflow.Name)

	// Should have steps from all three levels (a, b, c)
	assert.Len(t, workflow.Steps, 3)

	// Check step names
	stepNames := make([]string, len(workflow.Steps))
	for i, s := range workflow.Steps {
		stepNames[i] = s.Name
	}
	assert.Contains(t, stepNames, "step-from-a")
	assert.Contains(t, stepNames, "step-from-b")
	assert.Contains(t, stepNames, "step-from-c")

	// Check that param was overridden through the chain
	paramMap := make(map[string]core.Param)
	for _, p := range workflow.Params {
		paramMap[p.Name] = p
	}

	paramA, ok := paramMap["param-a"]
	require.True(t, ok, "param-a should exist")
	assert.Equal(t, "final-from-c", paramA.DefaultString())
}

func TestInheritanceResolver_KindMismatch(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	_, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "kind-mismatch.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kind mismatch")
}

func TestInheritanceResolver_ParentNotFound(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	// Create a temporary workflow that extends a non-existent parent
	parser := NewParser()
	content := []byte(`
kind: module
name: orphan
extends: non-existent-parent
steps:
  - name: test
    type: bash
    command: echo "test"
`)

	workflow, err := parser.ParseContent(content)
	require.NoError(t, err)

	resolver := NewInheritanceResolver(loader)
	_, err = resolver.Resolve(workflow)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load parent")
}

func TestInheritanceResolver_FlowModulesOverride(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "child-flow.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	assert.Equal(t, "child-flow", workflow.Name)
	assert.Equal(t, core.KindFlow, workflow.Kind)

	// Should have 3 modules (2 from parent + 1 appended)
	assert.Len(t, workflow.Modules, 3)

	// Check module names
	moduleNames := make([]string, len(workflow.Modules))
	for i, m := range workflow.Modules {
		moduleNames[i] = m.Name
	}
	assert.Equal(t, "module-a", moduleNames[0])
	assert.Equal(t, "module-b", moduleNames[1])
	assert.Equal(t, "module-c", moduleNames[2])

	// module-c should depend on module-b
	assert.Contains(t, workflow.Modules[2].DependsOn, "module-b")
}

func TestInheritanceResolver_NoInheritance(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "base-module.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	// ResolvedFrom should be empty since no inheritance
	assert.Empty(t, workflow.ResolvedFrom)

	// Should have original 3 steps
	assert.Len(t, workflow.Steps, 3)
}

func TestInheritanceResolver_DependenciesMerge(t *testing.T) {
	loader := NewLoader(testWorkflowsDir)

	// The base-module has dependencies on "echo" and "/bin/bash"
	workflow, err := loader.LoadWorkflowByPath(filepath.Join(testWorkflowsDir, "child-simple.yaml"))
	require.NoError(t, err)
	require.NotNil(t, workflow)

	// Dependencies should be inherited
	require.NotNil(t, workflow.Dependencies)
	assert.Contains(t, workflow.Dependencies.Commands, "echo")
	assert.Contains(t, workflow.Dependencies.Files, "/bin/bash")
}

func TestClone_Workflow(t *testing.T) {
	original := &core.Workflow{
		Kind:        core.KindModule,
		Name:        "test-workflow",
		Description: "Test workflow",
		Tags:        core.TagList{"tag1", "tag2"},
		Params: []core.Param{
			{Name: "param1", Default: "value1"},
			{Name: "param2", Default: "value2"},
		},
		Steps: []core.Step{
			{Name: "step1", Type: core.StepTypeBash, Command: "echo 1"},
			{Name: "step2", Type: core.StepTypeBash, Command: "echo 2"},
		},
		Dependencies: &core.Dependencies{
			Commands: []string{"cmd1", "cmd2"},
			Files:    []string{"file1", "file2"},
		},
		Preferences: &core.Preferences{
			Silent: boolPtr(true),
		},
	}

	cloned := original.Clone()

	// Verify values are copied
	assert.Equal(t, original.Name, cloned.Name)
	assert.Equal(t, original.Description, cloned.Description)
	assert.Equal(t, len(original.Tags), len(cloned.Tags))
	assert.Equal(t, len(original.Params), len(cloned.Params))
	assert.Equal(t, len(original.Steps), len(cloned.Steps))

	// Verify deep copy - modifying clone doesn't affect original
	cloned.Name = "modified"
	cloned.Tags[0] = "modified-tag"
	cloned.Params[0].Name = "modified-param"
	cloned.Steps[0].Name = "modified-step"
	cloned.Dependencies.Commands[0] = "modified-cmd"

	assert.Equal(t, "test-workflow", original.Name)
	assert.Equal(t, "tag1", original.Tags[0])
	assert.Equal(t, "param1", original.Params[0].Name)
	assert.Equal(t, "step1", original.Steps[0].Name)
	assert.Equal(t, "cmd1", original.Dependencies.Commands[0])
}

func TestClone_NilWorkflow(t *testing.T) {
	var w *core.Workflow
	cloned := w.Clone()
	assert.Nil(t, cloned)
}

func TestValidateOverride_InvalidStepsMode(t *testing.T) {
	parser := NewParser()
	content := []byte(`
kind: module
name: invalid-mode
extends: base-module
override:
  steps:
    mode: invalid-mode
    steps: []
steps:
  - name: test
    type: bash
    command: echo test
`)

	workflow, err := parser.ParseContent(content)
	require.NoError(t, err)

	err = parser.Validate(workflow)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid mode")
}

func TestValidateOverride_StepsOnFlow(t *testing.T) {
	parser := NewParser()
	content := []byte(`
kind: flow
name: flow-with-steps-override
extends: base-flow
override:
  steps:
    mode: append
    steps:
      - name: bad-step
        type: bash
        command: echo bad
modules:
  - name: m1
    path: some-path
`)

	workflow, err := parser.ParseContent(content)
	require.NoError(t, err)

	err = parser.Validate(workflow)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "steps override can only be used with module")
}

func TestValidateOverride_ModulesOnModule(t *testing.T) {
	parser := NewParser()
	content := []byte(`
kind: module
name: module-with-modules-override
extends: base-module
override:
  modules:
    mode: append
    modules:
      - name: bad-module
        path: some-path
steps:
  - name: test
    type: bash
    command: echo test
`)

	workflow, err := parser.ParseContent(content)
	require.NoError(t, err)

	err = parser.Validate(workflow)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "modules override can only be used with flow")
}

func boolPtr(b bool) *bool {
	return &b
}
