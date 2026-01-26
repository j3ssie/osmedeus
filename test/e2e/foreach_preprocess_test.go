package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForeachPreprocess_DryRun tests that the workflow validates correctly in dry-run mode
func TestForeachPreprocess_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing foreach-preprocess workflow in dry-run mode")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-preprocess", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode")
	assert.Contains(t, stdout, "DRY-RUN")

	log.Info("Asserting workflow name is displayed")
	assert.Contains(t, stdout, "test-foreach-preprocess")

	log.Success("foreach-preprocess workflow dry-run validates correctly")
}

// TestForeachPreprocess_GetParentURL tests that get_parent_url pre-processing works
func TestForeachPreprocess_GetParentURL(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing get_parent_url() pre-processing in foreach loop")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-preprocess", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	log.Info("Asserting parent URLs were extracted correctly")
	assert.Contains(t, stdout, "VERIFY:path_parent_ok")
	assert.Contains(t, stdout, "VERIFY:api_parent_ok")
	assert.Contains(t, stdout, "VERIFY:nested_parent_ok")

	log.Success("get_parent_url() pre-processing works correctly")
}

// TestForeachPreprocess_ParseURLDomain tests that parse_url pre-processing extracts domains
func TestForeachPreprocess_ParseURLDomain(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing parse_url() pre-processing to extract domains")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-preprocess", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting domains were extracted")
	assert.Contains(t, stdout, "VERIFY:domain1_ok")
	assert.Contains(t, stdout, "VERIFY:domain2_ok")
	assert.Contains(t, stdout, "VERIFY:domain3_ok")

	log.Success("parse_url() pre-processing extracts domains correctly")
}

// TestForeachPreprocess_ChainedFunctions tests chained function calls in pre-processing
func TestForeachPreprocess_ChainedFunctions(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing chained functions in variable_pre_process")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-preprocess", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting chained functions work (to_lower_case + parse_url)")
	assert.Contains(t, stdout, "VERIFY:chained1_ok")
	assert.Contains(t, stdout, "VERIFY:chained2_ok")
	assert.Contains(t, stdout, "VERIFY:chained3_ok")

	log.Success("chained functions in pre-processing work correctly")
}

// TestForeachPreprocess_NoPreprocess tests that foreach without pre-processing still works
func TestForeachPreprocess_NoPreprocess(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing foreach without variable_pre_process (original behavior)")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-preprocess", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting original URLs are preserved when no pre-processing")
	assert.Contains(t, stdout, "VERIFY:original1_ok")
	assert.Contains(t, stdout, "VERIFY:original2_ok")

	log.Success("foreach without pre-processing preserves original values")
}

// TestForeachPreprocess_FullWorkflow tests the complete workflow execution
func TestForeachPreprocess_FullWorkflow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing complete foreach-preprocess workflow execution")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-preprocess", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow completed successfully")
	assert.Contains(t, stdout, "completed")

	log.Info("Asserting final summary was reached")
	assert.Contains(t, stdout, "=== Foreach Pre-process Test Summary ===")
	assert.Contains(t, stdout, "=== Test Complete ===")

	log.Info("Asserting all verification checks passed")
	// Parent URL checks
	assert.Contains(t, stdout, "VERIFY:path_parent_ok")
	assert.Contains(t, stdout, "VERIFY:api_parent_ok")
	assert.Contains(t, stdout, "VERIFY:nested_parent_ok")
	// Domain extraction checks
	assert.Contains(t, stdout, "VERIFY:domain1_ok")
	assert.Contains(t, stdout, "VERIFY:domain2_ok")
	assert.Contains(t, stdout, "VERIFY:domain3_ok")
	// Chained function checks
	assert.Contains(t, stdout, "VERIFY:chained1_ok")
	assert.Contains(t, stdout, "VERIFY:chained2_ok")
	assert.Contains(t, stdout, "VERIFY:chained3_ok")
	// Original value checks
	assert.Contains(t, stdout, "VERIFY:original1_ok")
	assert.Contains(t, stdout, "VERIFY:original2_ok")

	log.Success("complete foreach-preprocess workflow executed successfully")
}
