package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDecisionConditions_SingleMatch tests a single condition that evaluates to true
func TestDecisionConditions_SingleMatch(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing single condition match")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-conditions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	log.Info("Asserting single condition function executed")
	assert.Contains(t, stdout, "[COND] Target is present")

	log.Success("single condition match works correctly")
}

// TestDecisionConditions_MultipleMatch tests that all matching conditions execute (no short-circuit)
func TestDecisionConditions_MultipleMatch(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing multiple conditions - all matching ones execute")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-conditions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting both matching conditions executed")
	assert.Contains(t, stdout, "[COND-A] Flag A is true")
	assert.Contains(t, stdout, "[COND-B] Flag B is true")

	log.Info("Asserting non-matching condition did not execute")
	assert.NotContains(t, stdout, "[COND-C]")

	log.Success("multiple condition matching works correctly")
}

// TestDecisionConditions_InlineCommand tests condition with inline bash command
func TestDecisionConditions_InlineCommand(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing condition with inline command")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-conditions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting inline command executed")
	assert.Contains(t, stdout, "[COND-CMD] Extra command executed")

	log.Info("Asserting inline bash step appears in results")
	assert.Contains(t, stdout, "decision-condition-bash")

	log.Success("condition inline command works correctly")
}

// TestDecisionConditions_GotoSkipsStep tests condition with goto skips intermediate steps
func TestDecisionConditions_GotoSkipsStep(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing condition goto skips intermediate steps")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-conditions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting condition function ran before goto")
	assert.Contains(t, stdout, "[COND-GOTO] Jumping to final")

	log.Info("Asserting skipped step was not executed")
	assert.NotContains(t, stdout, "[SKIPPED]")

	log.Info("Asserting final step executed after goto")
	assert.Contains(t, stdout, "[FINAL] Scan complete for example.com")

	// Verify ordering: goto function before final
	idxGoto := strings.Index(stdout, "[COND-GOTO]")
	idxFinal := strings.Index(stdout, "[FINAL]")
	assert.True(t, idxGoto >= 0, "goto condition output not found")
	assert.True(t, idxFinal >= 0, "final output not found")
	assert.True(t, idxGoto < idxFinal, "goto condition should appear before final step")

	log.Success("condition goto correctly skips intermediate steps")
}

// TestDecisionConditions_NoMatchDisabled tests that false conditions do not execute
func TestDecisionConditions_NoMatchDisabled(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing condition with enable_extra=false does not run command")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-conditions", "-t", "example.com",
		"-p", "enable_extra=false", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	log.Info("Asserting inline command did NOT execute when condition is false")
	assert.NotContains(t, stdout, "[COND-CMD] Extra command executed")

	log.Success("false condition correctly prevents execution")
}

// TestDecisionConditions_DryRun tests that the workflow validates in dry-run mode
func TestDecisionConditions_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing decision-conditions workflow in dry-run mode")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-decision-conditions", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-decision-conditions")

	log.Success("decision-conditions workflow dry-run works correctly")
}

// TestDecisionConditions_WorkflowValidate tests that the workflow passes validation
func TestDecisionConditions_WorkflowValidate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing decision-conditions workflow validation")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-decision-conditions", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow is valid")
	assert.True(t, strings.Contains(stdout, "is valid") || strings.Contains(stdout, "passed all lint checks"),
		"expected validation success message")

	log.Success("decision-conditions workflow passes validation")
}
