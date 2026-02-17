package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDecisionInline_QuickMode tests inline execution in decision cases with default "quick" mode
func TestDecisionInline_QuickMode(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing decision inline execution with quick mode (default)")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-inline", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify workflow completed
	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	// Verify inline function executed for "quick" case (log_info output is visible)
	log.Info("Asserting inline function ran for quick mode")
	assert.Contains(t, stdout, "[INLINE] Quick scan selected")

	// Verify inline functions (plural) executed
	log.Info("Asserting inline functions (plural) ran")
	assert.Contains(t, stdout, "[MULTI-FUNC-1] Configuring quick scan")
	assert.Contains(t, stdout, "[MULTI-FUNC-2] Quick scan configured")

	// Verify inline bash steps appear in step table
	log.Info("Asserting inline bash steps appear in results")
	assert.Contains(t, stdout, "decision-inline-bash")
	assert.Contains(t, stdout, "decision-inline-function")

	// Verify inline function+goto worked (inline runs, then jumps)
	log.Info("Asserting inline function with goto ran")
	assert.Contains(t, stdout, "[INLINE-GOTO] Quick inline before jump")

	// Verify goto skipped the skipped-step
	log.Info("Asserting skipped step was not executed")
	assert.NotContains(t, stdout, "[SKIPPED]")

	// Verify final step ran
	log.Info("Asserting final step executed")
	assert.Contains(t, stdout, "[FINAL] Scan complete for example.com in quick mode")

	log.Success("decision inline execution works correctly for quick mode")
}

// TestDecisionInline_DeepMode tests inline execution when a different case is matched
func TestDecisionInline_DeepMode(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing decision inline execution with deep mode")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-inline", "-t", "target.io",
		"-p", "scan_mode=deep", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	// Verify deep mode inline function executed
	log.Info("Asserting deep mode inline function ran")
	assert.Contains(t, stdout, "[INLINE] Deep scan selected")

	// Verify deep mode multi-functions executed
	log.Info("Asserting deep mode multi-functions ran")
	assert.Contains(t, stdout, "[MULTI-FUNC-1] Configuring deep scan")
	assert.Contains(t, stdout, "[MULTI-FUNC-2] Deep scan configured")

	// Verify quick mode was not executed
	log.Info("Asserting quick mode functions did not run")
	assert.NotContains(t, stdout, "Quick scan selected")
	assert.NotContains(t, stdout, "Configuring quick scan")

	// Verify final step
	log.Info("Asserting final step executed")
	assert.Contains(t, stdout, "[FINAL] Scan complete for target.io in deep mode")

	log.Success("decision inline execution works correctly for deep mode")
}

// TestDecisionInline_DefaultCase tests that the default case inline runs when no case matches
func TestDecisionInline_DefaultCase(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing decision inline with unmatched value (default case)")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-inline", "-t", "example.com",
		"-p", "scan_mode=unknown", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	// Verify default case inline function executed
	log.Info("Asserting default inline function ran")
	assert.Contains(t, stdout, "[INLINE] Default mode selected")

	// Verify no quick or deep functions ran
	log.Info("Asserting non-matching cases did not run")
	assert.NotContains(t, stdout, "Quick scan selected")
	assert.NotContains(t, stdout, "Deep scan selected")

	// Verify default goto worked in branch-and-jump step
	log.Info("Asserting default goto jumped correctly")
	assert.Contains(t, stdout, "[INLINE-GOTO] Default inline before jump")
	assert.NotContains(t, stdout, "[SKIPPED]")

	log.Success("default case inline execution works correctly")
}

// TestDecisionInline_GotoSkipsStep tests that goto after inline execution correctly skips intermediate steps
func TestDecisionInline_GotoSkipsStep(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing that inline+goto correctly skips intermediate steps")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-inline", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// The inline function should run before goto
	log.Info("Asserting inline function ran before goto")
	assert.Contains(t, stdout, "[INLINE-GOTO] Quick inline before jump")

	// The skipped-step should not appear
	log.Info("Asserting skipped-step was skipped by goto")
	assert.NotContains(t, stdout, "[SKIPPED] This should not appear")

	// The final-step should appear after the inline goto
	log.Info("Asserting final-step ran after goto")
	assert.Contains(t, stdout, "[FINAL] Scan complete")

	// Verify ordering: inline-goto appears before final
	idxInline := strings.Index(stdout, "[INLINE-GOTO]")
	idxFinal := strings.Index(stdout, "[FINAL]")
	assert.True(t, idxInline >= 0, "inline-goto output not found")
	assert.True(t, idxFinal >= 0, "final output not found")
	assert.True(t, idxInline < idxFinal, "inline-goto should appear before final step")

	log.Success("inline+goto correctly skips intermediate steps")
}

// TestDecisionInline_StepTable tests that inline steps appear correctly in the step results table
func TestDecisionInline_StepTable(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing inline step entries in step results table")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-inline", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify step table contains inline step entries
	log.Info("Asserting step table contains inline bash steps")
	assert.Contains(t, stdout, "decision-inline-bash")

	log.Info("Asserting step table contains inline function steps")
	assert.Contains(t, stdout, "decision-inline-function")

	// Verify the main workflow steps are present
	log.Info("Asserting main steps are present")
	assert.Contains(t, stdout, "detect-mode")
	assert.Contains(t, stdout, "setup-scan")
	assert.Contains(t, stdout, "branch-and-jump")
	assert.Contains(t, stdout, "final-step")

	// Verify total completed steps > workflow step count (inline steps add extra)
	log.Info("Asserting completed steps include inline steps")
	assert.Contains(t, stdout, "completed_steps")

	log.Success("inline steps appear correctly in step results table")
}

// TestDecisionInline_DryRun tests that the workflow validates in dry-run mode
func TestDecisionInline_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing decision-inline workflow in dry-run mode")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-inline", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-decision-inline")

	log.Success("decision-inline workflow dry-run works correctly")
}

// TestDecisionInline_WorkflowValidate tests that the workflow passes validation
func TestDecisionInline_WorkflowValidate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing decision-inline workflow validation")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-decision-inline", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow is valid")
	assert.True(t, strings.Contains(stdout, "is valid") || strings.Contains(stdout, "passed all lint checks"),
		"expected validation success message")

	log.Success("decision-inline workflow passes validation")
}
