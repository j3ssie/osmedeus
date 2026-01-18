package lib

import (
	"context"
	"testing"
	"time"
)

func TestEval_SimpleExpression(t *testing.T) {
	result, err := Eval(`1 + 1`, nil)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	// Goja returns int64 for integer expressions
	if val, ok := result.(int64); !ok || val != 2 {
		t.Errorf("Expected 2, got %v", result)
	}
}

func TestEval_EmptyExpression(t *testing.T) {
	_, err := Eval("", nil)
	if err != ErrEmptyExpression {
		t.Errorf("Expected ErrEmptyExpression, got %v", err)
	}
}

func TestEval_StringFunction(t *testing.T) {
	result, err := Eval(`trim("  hello  ")`, nil)
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	if result != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}
}

func TestEval_WithContext(t *testing.T) {
	result, err := Eval(`trim(input)`, &EvalOptions{
		Context: map[string]interface{}{"input": "  test  "},
	})
	if err != nil {
		t.Fatalf("Eval failed: %v", err)
	}

	if result != "test" {
		t.Errorf("Expected 'test', got %v", result)
	}
}

func TestEvalCondition_True(t *testing.T) {
	result, err := EvalCondition(`1 < 2`, nil)
	if err != nil {
		t.Fatalf("EvalCondition failed: %v", err)
	}

	if !result {
		t.Error("Expected true, got false")
	}
}

func TestEvalCondition_False(t *testing.T) {
	result, err := EvalCondition(`1 > 2`, nil)
	if err != nil {
		t.Fatalf("EvalCondition failed: %v", err)
	}

	if result {
		t.Error("Expected false, got true")
	}
}

func TestEvalCondition_Empty(t *testing.T) {
	_, err := EvalCondition("", nil)
	if err != ErrEmptyExpression {
		t.Errorf("Expected ErrEmptyExpression, got %v", err)
	}
}

func TestEvalCondition_WithContext(t *testing.T) {
	result, err := EvalCondition(`value > 5`, &EvalOptions{
		Context: map[string]interface{}{"value": 10},
	})
	if err != nil {
		t.Fatalf("EvalCondition failed: %v", err)
	}

	if !result {
		t.Error("Expected true, got false")
	}
}

func TestEvalFunction(t *testing.T) {
	result, err := EvalFunction(`1 + 2`)
	if err != nil {
		t.Fatalf("EvalFunction failed: %v", err)
	}

	// Goja returns int64 for integer expressions
	if val, ok := result.(int64); !ok || val != 3 {
		t.Errorf("Expected 3, got %v", result)
	}
}

func TestEvalFunctionWithContext(t *testing.T) {
	result, err := EvalFunctionWithContext(`split(text, ",")`, map[string]interface{}{
		"text": "a,b,c",
	})
	if err != nil {
		t.Fatalf("EvalFunctionWithContext failed: %v", err)
	}

	// Otto's split returns []string
	arr, ok := result.([]string)
	if !ok {
		t.Fatalf("Expected []string, got %T", result)
	}

	if len(arr) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arr))
	}
}

func TestParseWorkflow_Valid(t *testing.T) {
	yaml := `
name: test-workflow
kind: module
steps:
  - name: echo
    type: bash
    command: echo hello
`
	workflow, err := ParseWorkflow(yaml)
	if err != nil {
		t.Fatalf("ParseWorkflow failed: %v", err)
	}

	if workflow.Name != "test-workflow" {
		t.Errorf("Expected name 'test-workflow', got %q", workflow.Name)
	}

	if !workflow.IsModule() {
		t.Error("Expected workflow to be a module")
	}

	if len(workflow.Steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(workflow.Steps))
	}
}

func TestParseWorkflow_Empty(t *testing.T) {
	_, err := ParseWorkflow("")
	if err != ErrEmptyWorkflow {
		t.Errorf("Expected ErrEmptyWorkflow, got %v", err)
	}
}

func TestParseWorkflow_Invalid(t *testing.T) {
	_, err := ParseWorkflow("invalid: yaml: content: :")
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}

	if !IsParseError(err) {
		t.Errorf("Expected ParseError, got %T", err)
	}
}

func TestValidateWorkflow_Valid(t *testing.T) {
	yaml := `
name: test-workflow
kind: module
steps:
  - name: echo
    type: bash
    command: echo hello
`
	err := ValidateWorkflow(yaml)
	if err != nil {
		t.Fatalf("ValidateWorkflow failed: %v", err)
	}
}

func TestValidateWorkflow_MissingName(t *testing.T) {
	yaml := `
kind: module
steps:
  - name: echo
    type: bash
    command: echo hello
`
	err := ValidateWorkflow(yaml)
	if err == nil {
		t.Error("Expected error for missing name")
	}

	if !IsValidationError(err) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestValidateWorkflow_InvalidKind(t *testing.T) {
	yaml := `
name: test
kind: invalid
steps:
  - name: echo
    type: bash
    command: echo hello
`
	err := ValidateWorkflow(yaml)
	if err == nil {
		t.Error("Expected error for invalid kind")
	}

	if !IsValidationError(err) {
		t.Errorf("Expected ValidationError, got %T", err)
	}
}

func TestRun_EmptyTarget(t *testing.T) {
	_, err := Run("", "workflow: yaml", nil)
	if err != ErrEmptyTarget {
		t.Errorf("Expected ErrEmptyTarget, got %v", err)
	}
}

func TestRun_EmptyWorkflow(t *testing.T) {
	_, err := Run("target", "", nil)
	if err != ErrEmptyWorkflow {
		t.Errorf("Expected ErrEmptyWorkflow, got %v", err)
	}
}

func TestRun_InvalidYAML(t *testing.T) {
	_, err := Run("target", "invalid: yaml: :", nil)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}

	if !IsParseError(err) {
		t.Errorf("Expected ParseError, got %T", err)
	}
}

func TestRun_NotModule(t *testing.T) {
	yaml := `
name: test-flow
kind: flow
modules:
  - name: m1
    path: ./modules/m1.yaml
`
	_, err := Run("target", yaml, nil)
	if err != ErrNotModule {
		t.Errorf("Expected ErrNotModule, got %v", err)
	}
}

func TestRun_SimpleWorkflow(t *testing.T) {
	yaml := `
name: simple-echo
kind: module
steps:
  - name: echo
    type: bash
    command: echo "hello world"
`
	result, err := Run("test.com", yaml, &RunOptions{
		Silent:               true,
		DisableWorkflowState: true,
		DisableDatabase:      true,
	})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if result.Status != "completed" {
		t.Errorf("Expected status 'completed', got %q", result.Status)
	}

	if result.WorkflowName != "simple-echo" {
		t.Errorf("Expected workflow name 'simple-echo', got %q", result.WorkflowName)
	}

	if result.Target != "test.com" {
		t.Errorf("Expected target 'test.com', got %q", result.Target)
	}

	if len(result.Steps) != 1 {
		t.Errorf("Expected 1 step result, got %d", len(result.Steps))
	}
}

func TestRunWithContext_Cancellation(t *testing.T) {
	yaml := `
name: slow-workflow
kind: module
steps:
  - name: slow
    type: bash
    command: sleep 10
`
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := RunWithContext(ctx, "test.com", yaml, &RunOptions{
		Silent:               true,
		DisableWorkflowState: true,
		DisableDatabase:      true,
	})

	// The workflow should be cancelled
	if err == nil && result != nil && result.Status != "cancelled" {
		t.Error("Expected cancellation or error")
	}
}

func TestRunOptions_WithMethods(t *testing.T) {
	opts := DefaultRunOptions()

	// Test WithParams
	params := map[string]string{"key": "value"}
	newOpts := opts.WithParams(params)
	if newOpts.Params["key"] != "value" {
		t.Error("WithParams failed")
	}

	// Test WithTactic
	newOpts = opts.WithTactic("aggressive")
	if newOpts.Tactic != "aggressive" {
		t.Error("WithTactic failed")
	}

	// Test WithDryRun
	newOpts = opts.WithDryRun(true)
	if !newOpts.DryRun {
		t.Error("WithDryRun failed")
	}

	// Test WithVerbose
	newOpts = opts.WithVerbose(true)
	if !newOpts.Verbose {
		t.Error("WithVerbose failed")
	}

	// Test WithSilent
	newOpts = opts.WithSilent(false)
	if newOpts.Silent {
		t.Error("WithSilent failed")
	}
}

func TestEvalOptions_WithMethods(t *testing.T) {
	opts := DefaultEvalOptions()

	// Test WithContext
	ctx := map[string]interface{}{"key": "value"}
	newOpts := opts.WithContext(ctx)
	if newOpts.Context["key"] != "value" {
		t.Error("WithContext failed")
	}

	// Test WithTarget
	newOpts = opts.WithTarget("test.com")
	if newOpts.Target != "test.com" {
		t.Error("WithTarget failed")
	}
}

func TestRunResult_Helpers(t *testing.T) {
	result := &RunResult{
		Status: "completed",
		Exports: map[string]interface{}{
			"stringVal": "hello",
			"boolVal":   true,
		},
		Steps: []*StepResult{
			{Status: "success"},
			{Status: "success"},
			{Status: "skipped"},
			{Status: "failed"},
		},
	}

	if !result.IsSuccess() {
		t.Error("Expected IsSuccess to be true")
	}

	if result.IsFailed() {
		t.Error("Expected IsFailed to be false")
	}

	// Test GetExport
	if v, ok := result.GetExport("stringVal"); !ok || v != "hello" {
		t.Error("GetExport failed for stringVal")
	}

	// Test GetExportString
	if v, ok := result.GetExportString("stringVal"); !ok || v != "hello" {
		t.Error("GetExportString failed")
	}

	// Test GetExportBool
	if v, ok := result.GetExportBool("boolVal"); !ok || !v {
		t.Error("GetExportBool failed")
	}

	// Test step counts
	if result.SuccessfulSteps() != 2 {
		t.Errorf("Expected 2 successful steps, got %d", result.SuccessfulSteps())
	}

	if result.FailedSteps() != 1 {
		t.Errorf("Expected 1 failed step, got %d", result.FailedSteps())
	}

	if result.SkippedSteps() != 1 {
		t.Errorf("Expected 1 skipped step, got %d", result.SkippedSteps())
	}
}

func TestStepResult_Helpers(t *testing.T) {
	step := &StepResult{
		Status: "success",
		Exports: map[string]interface{}{
			"output": "value",
		},
	}

	if !step.IsSuccess() {
		t.Error("Expected IsSuccess to be true")
	}

	if step.IsFailed() {
		t.Error("Expected IsFailed to be false")
	}

	if step.IsSkipped() {
		t.Error("Expected IsSkipped to be false")
	}

	if v, ok := step.GetExport("output"); !ok || v != "value" {
		t.Error("GetExport failed")
	}
}

func TestErrorTypes(t *testing.T) {
	// Test ParseError
	parseErr := NewParseError("test message", nil)
	if parseErr.Error() != "parse error: test message" {
		t.Errorf("Unexpected ParseError message: %s", parseErr.Error())
	}

	// Test ValidationError
	validErr := NewValidationError("field", "message")
	if validErr.Error() != "validation error in field 'field': message" {
		t.Errorf("Unexpected ValidationError message: %s", validErr.Error())
	}

	// Test ExecutionError
	execErr := NewExecutionError("step", "bash", "failed")
	if execErr.Error() != "execution error at step 'step' (bash): failed" {
		t.Errorf("Unexpected ExecutionError message: %s", execErr.Error())
	}

	// Test Is* functions
	if !IsParseError(parseErr) {
		t.Error("IsParseError should return true")
	}

	if !IsValidationError(validErr) {
		t.Error("IsValidationError should return true")
	}

	if !IsExecutionError(execErr) {
		t.Error("IsExecutionError should return true")
	}
}
