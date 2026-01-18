package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParamsExports_DefaultParams tests the workflow with default parameter values
func TestParamsExports_DefaultParams(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing params-exports workflow with default params")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-params-exports", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify workflow executed successfully
	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	// Verify default params were used correctly
	log.Info("Asserting default param values")
	assert.Contains(t, stdout, "enable_feature=true")
	assert.Contains(t, stdout, "skip_validation=false")

	// Verify correct decision branch was taken (enable_feature=true -> feature-enabled-step)
	log.Info("Asserting feature-enabled branch was taken")
	assert.Contains(t, stdout, "Feature is ENABLED")
	assert.Contains(t, stdout, "feature-enabled-step")

	// Verify correct decision branch was taken (skip_validation=false -> run-validation)
	log.Info("Asserting run-validation branch was taken")
	assert.Contains(t, stdout, "Validation PASSED")
	assert.Contains(t, stdout, "run-validation")

	// Verify exports are correctly propagated to final summary
	log.Info("Asserting exports are propagated")
	assert.Contains(t, stdout, "Feature Status: ENABLED")
	assert.Contains(t, stdout, "Validation Result: PASSED")
	assert.Contains(t, stdout, "Exported Custom: default_value")

	log.Success("params-exports workflow with default params works correctly")
}

// TestParamsExports_CustomParams tests the workflow with custom parameter values
func TestParamsExports_CustomParams(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing params-exports workflow with custom params")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	// Run with custom params that trigger different branches
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-params-exports", "-t", "example.com",
		"-p", "enable_feature=false",
		"-p", "skip_validation=true",
		"-p", "custom_value=my_custom_value",
		"-F", workflowPath)
	require.NoError(t, err)

	// Verify workflow executed successfully
	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	// Verify custom params were used correctly
	log.Info("Asserting custom param values")
	assert.Contains(t, stdout, "enable_feature=false")
	assert.Contains(t, stdout, "skip_validation=true")

	// Verify correct decision branch was taken (enable_feature=false -> feature-disabled-step)
	log.Info("Asserting feature-disabled branch was taken")
	assert.Contains(t, stdout, "Feature is DISABLED")
	assert.Contains(t, stdout, "feature-disabled-step")

	// Verify correct decision branch was taken (skip_validation=true -> validation-skipped)
	log.Info("Asserting validation-skipped branch was taken")
	assert.Contains(t, stdout, "Validation SKIPPED")
	assert.Contains(t, stdout, "validation-skipped")

	// Verify exports are correctly propagated to final summary
	log.Info("Asserting exports are propagated")
	assert.Contains(t, stdout, "Feature Status: DISABLED")
	assert.Contains(t, stdout, "Validation Result: SKIPPED")
	assert.Contains(t, stdout, "Exported Custom: my_custom_value")

	log.Success("params-exports workflow with custom params works correctly")
}

// TestParamsExports_ExportVerification tests that exports from previous steps are accessible
func TestParamsExports_ExportVerification(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing exports verification in params-exports workflow")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-params-exports", "-t", "test.example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify the verify-exports step ran successfully
	log.Info("Asserting verify-exports step executed")
	assert.Contains(t, stdout, "verify-exports")
	assert.Contains(t, stdout, "=== Export Verification ===")

	// Verify all exports are accessible in the verification step
	log.Info("Asserting all exports are accessible")
	assert.Contains(t, stdout, "feature_enabled: true")
	assert.Contains(t, stdout, "feature_status: ENABLED")
	assert.Contains(t, stdout, "validation_result: PASSED")
	assert.Contains(t, stdout, "exported_custom: default_value")
	assert.Contains(t, stdout, "target: test.example.com")

	// Verify final summary contains all expected values
	log.Info("Asserting final summary contains all values")
	assert.Contains(t, stdout, "=== Workflow Summary ===")
	assert.Contains(t, stdout, "Verification Complete: true")

	log.Success("exports verification works correctly")
}

// TestParamsExports_DecisionRouting tests that decision routing works correctly
func TestParamsExports_DecisionRouting(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing decision routing in params-exports workflow")

	workflowPath := getTestdataPath(t)

	// Test case 1: enable_feature=true should skip feature-disabled-step
	log.Info("Test case 1: enable_feature=true")
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-params-exports", "-t", "example.com",
		"-p", "enable_feature=true", "-F", workflowPath)
	require.NoError(t, err)
	assert.Contains(t, stdout, "feature-enabled-step")
	assert.NotContains(t, stdout, "feature-disabled-step")

	// Test case 2: enable_feature=false should skip feature-enabled-step
	log.Info("Test case 2: enable_feature=false")
	stdout, _, err = runCLIWithLog(t, log, "run", "-m", "test-params-exports", "-t", "example.com",
		"-p", "enable_feature=false", "-F", workflowPath)
	require.NoError(t, err)
	assert.Contains(t, stdout, "feature-disabled-step")
	assert.NotContains(t, stdout, "feature-enabled-step")

	// Test case 3: skip_validation=false should run validation
	log.Info("Test case 3: skip_validation=false")
	stdout, _, err = runCLIWithLog(t, log, "run", "-m", "test-params-exports", "-t", "example.com",
		"-p", "skip_validation=false", "-F", workflowPath)
	require.NoError(t, err)
	assert.Contains(t, stdout, "run-validation")
	assert.NotContains(t, stdout, "validation-skipped")

	// Test case 4: skip_validation=true should skip validation
	log.Info("Test case 4: skip_validation=true")
	stdout, _, err = runCLIWithLog(t, log, "run", "-m", "test-params-exports", "-t", "example.com",
		"-p", "skip_validation=true", "-F", workflowPath)
	require.NoError(t, err)
	assert.Contains(t, stdout, "validation-skipped")
	assert.NotContains(t, stdout, "run-validation")

	log.Success("decision routing works correctly for all cases")
}

// TestParamsExports_DryRun tests that the workflow can be validated with dry-run
func TestParamsExports_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing params-exports workflow in dry-run mode")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-params-exports", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	// Verify dry-run mode indicator
	log.Info("Asserting dry-run mode")
	assert.Contains(t, stdout, "DRY-RUN")

	// Verify workflow name is displayed
	log.Info("Asserting workflow name is displayed")
	assert.Contains(t, stdout, "test-params-exports")

	log.Success("params-exports workflow dry-run works correctly")
}
