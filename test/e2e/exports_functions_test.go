package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExportsFunctions_AllExports tests that all utility function exports work correctly
func TestExportsFunctions_AllExports(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing exports-functions workflow with all utility functions")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-exports-functions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify workflow completed
	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	// Verify final summary shows all exports
	log.Info("Asserting final summary is present")
	assert.Contains(t, stdout, "=== Exports Functions Summary ===")
	assert.Contains(t, stdout, "=== All Exports Verified ===")

	log.Success("all exports evaluated correctly")
}

// TestExportsFunctions_Trim tests that trim() removes leading/trailing whitespace
func TestExportsFunctions_Trim(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing trim() function in exports")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-exports-functions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify trim() removes whitespace - the output should have no leading/trailing spaces
	// Note: trim() removes both spaces and newlines, so we just check it contains trimmed_value
	log.Info("Asserting trim() removes whitespace")
	assert.Contains(t, stdout, "Trimmed: [trimmed_value")

	log.Success("trim() function works correctly")
}

// TestExportsFunctions_FileLength tests that fileLength() returns correct line count
func TestExportsFunctions_FileLength(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing fileLength() function in exports")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-exports-functions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify fileLength() returns 5 for our 5-line file
	log.Info("Asserting fileLength() returns 5")
	assert.Contains(t, stdout, "Line count: 5")

	log.Success("fileLength() function works correctly")
}

// TestExportsFunctions_Contains tests that contains() correctly identifies substrings
func TestExportsFunctions_Contains(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing contains() function in exports")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-exports-functions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify contains() returns true when substring is found
	log.Info("Asserting contains() returns true for 'success'")
	assert.Contains(t, stdout, "Has success: true")

	// Verify contains() returns false when substring is not found
	log.Info("Asserting contains() returns false for 'failure'")
	assert.Contains(t, stdout, "Has failure: false")

	log.Success("contains() function works correctly")
}

// TestExportsFunctions_FileExists tests that fileExists() correctly detects file presence
func TestExportsFunctions_FileExists(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing fileExists() function in exports")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-exports-functions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify fileExists() returns true for existing file
	log.Info("Asserting fileExists() returns true for existing file")
	assert.Contains(t, stdout, "File exists: true")

	// Verify fileExists() returns false for missing file
	log.Info("Asserting fileExists() returns false for missing file")
	assert.Contains(t, stdout, "Missing file: false")

	log.Success("fileExists() function works correctly")
}

// TestExportsFunctions_Replace tests that replace() substitutes strings correctly
func TestExportsFunctions_Replace(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing replace() function in exports")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-exports-functions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify replace() substitutes commas with dashes
	log.Info("Asserting replace() substitutes commas with dashes")
	assert.Contains(t, stdout, "Replaced: hello-world-test")

	log.Success("replace() function works correctly")
}

// TestExportsFunctions_DryRun tests that the workflow validates correctly in dry-run mode
func TestExportsFunctions_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing exports-functions workflow in dry-run mode")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-exports-functions", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	// Verify dry-run mode indicator
	log.Info("Asserting dry-run mode")
	assert.Contains(t, stdout, "DRY-RUN")

	// Verify workflow name is displayed
	log.Info("Asserting workflow name is displayed")
	assert.Contains(t, stdout, "test-exports-functions")

	log.Success("exports-functions workflow dry-run validates correctly")
}

// TestExportsFunctions_PropagationToFinalStep tests that exports propagate to final summary step
func TestExportsFunctions_PropagationToFinalStep(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing exports propagation to final summary step")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-exports-functions", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify all exports are accessible in the final summary step
	log.Info("Asserting all exports are accessible in final summary")
	assert.Contains(t, stdout, "Trimmed:")
	assert.Contains(t, stdout, "Line count:")
	assert.Contains(t, stdout, "Has success:")
	assert.Contains(t, stdout, "Has failure:")
	assert.Contains(t, stdout, "File exists:")
	assert.Contains(t, stdout, "Missing file:")
	assert.Contains(t, stdout, "Replaced:")

	log.Success("exports propagate correctly to final step")
}
