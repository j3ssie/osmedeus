package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForeachPathFriendly_DryRun tests that the workflow validates in dry-run mode
func TestForeachPathFriendly_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing foreach-path-friendly workflow in dry-run mode")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-path-friendly", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-foreach-path-friendly")

	log.Success("foreach-path-friendly workflow dry-run validates correctly")
}

// TestForeachPathFriendly_BasicSanitization tests that _url_ replaces unsafe chars
func TestForeachPathFriendly_BasicSanitization(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing basic _url_ path-friendly variable sanitization")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-path-friendly", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting URL with slashes/colons is sanitized")
	assert.Contains(t, stdout, "VERIFY:basic_url_ok")

	log.Info("Asserting URL with port is sanitized")
	assert.Contains(t, stdout, "VERIFY:port_url_ok")

	log.Info("Asserting simple value without unsafe chars is unchanged")
	assert.Contains(t, stdout, "VERIFY:simple_ok")

	log.Success("basic path-friendly sanitization works correctly")
}

// TestForeachPathFriendly_DirectoryCreation tests using _url_ as a directory name
func TestForeachPathFriendly_DirectoryCreation(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing _url_ used as directory name for per-iteration output")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-path-friendly", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting directories were created with sanitized names")
	assert.Contains(t, stdout, "VERIFY:dir1_ok")
	assert.Contains(t, stdout, "VERIFY:dir2_ok")
	assert.Contains(t, stdout, "VERIFY:dir3_ok")

	log.Info("Asserting result files exist inside sanitized directories")
	assert.Contains(t, stdout, "VERIFY:file1_ok")

	log.Success("path-friendly variable works as directory name")
}

// TestForeachPathFriendly_LongValueTruncation tests deterministic truncation for long values
func TestForeachPathFriendly_LongValueTruncation(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing deterministic truncation for long URL values")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-path-friendly", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting same input produces same truncated output (deterministic)")
	assert.Contains(t, stdout, "VERIFY:deterministic_ok")

	log.Info("Asserting truncated result respects length limit")
	assert.Contains(t, stdout, "VERIFY:truncated_ok")

	log.Success("long value truncation is deterministic and within bounds")
}

// TestForeachPathFriendly_IDCoexistence tests that _id_ still works alongside _url_
func TestForeachPathFriendly_IDCoexistence(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing _id_ and _url_ coexist in foreach iterations")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-path-friendly", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting both _id_ and _url_ are available in the same iteration")
	assert.Contains(t, stdout, "VERIFY:id_and_safe_ok")

	log.Success("_id_ and _url_ path-friendly variables coexist correctly")
}

// TestForeachPathFriendly_FullWorkflow tests the complete workflow end-to-end
func TestForeachPathFriendly_FullWorkflow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing complete foreach-path-friendly workflow execution")

	workflowPath := getTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-foreach-path-friendly", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow completed successfully")
	assert.Contains(t, stdout, "completed")

	log.Info("Asserting final summary was reached")
	assert.Contains(t, stdout, "=== Foreach Path-Friendly Test Summary ===")
	assert.Contains(t, stdout, "=== Test Complete ===")

	log.Info("Asserting all verification checks passed")
	// Basic sanitization
	assert.Contains(t, stdout, "VERIFY:basic_url_ok")
	assert.Contains(t, stdout, "VERIFY:port_url_ok")
	assert.Contains(t, stdout, "VERIFY:simple_ok")
	// Directory creation
	assert.Contains(t, stdout, "VERIFY:dir1_ok")
	assert.Contains(t, stdout, "VERIFY:dir2_ok")
	assert.Contains(t, stdout, "VERIFY:dir3_ok")
	assert.Contains(t, stdout, "VERIFY:file1_ok")
	// Long value truncation
	assert.Contains(t, stdout, "VERIFY:deterministic_ok")
	assert.Contains(t, stdout, "VERIFY:truncated_ok")
	// ID coexistence
	assert.Contains(t, stdout, "VERIFY:id_and_safe_ok")

	log.Success("complete foreach-path-friendly workflow executed successfully")
}
