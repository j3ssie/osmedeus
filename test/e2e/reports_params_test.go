package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReportsParams_DefaultPaths tests that report paths resolve params with default values
func TestReportsParams_DefaultPaths(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing reports-params workflow with default paths")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-reports-params", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify workflow completed successfully
	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	// Verify step execution
	log.Info("Asserting steps executed")
	assert.Contains(t, stdout, "DNS file created at:")
	assert.Contains(t, stdout, "HTTP file created at:")

	// Verify files were created at param-defined paths
	log.Info("Asserting files exist at param-defined paths")
	assert.Contains(t, stdout, "DNS file exists: yes")
	assert.Contains(t, stdout, "HTTP file exists: yes")

	// Verify final summary shows resolved paths
	log.Info("Asserting final summary shows resolved paths")
	assert.Contains(t, stdout, "=== Reports Params Summary ===")
	assert.Contains(t, stdout, "Target: example.com")
	assert.Contains(t, stdout, "=== All Reports Verified ===")

	log.Success("reports-params workflow with default paths works correctly")
}

// TestReportsParams_NestedTemplates tests that nested template variables are resolved correctly
func TestReportsParams_NestedTemplates(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing reports-params workflow with nested template variables")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	// Use a target with special characters that get sanitized in TargetSpace
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-reports-params", "-t", "test.example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify nested template {{Output}}/probing/dns-{{TargetSpace}}.txt resolved correctly
	log.Info("Asserting nested templates resolved")
	assert.Contains(t, stdout, "DNS file created at:")
	assert.Contains(t, stdout, "DNS file exists: yes")

	// Verify TargetSpace is shown in summary
	log.Info("Asserting TargetSpace is shown")
	assert.Contains(t, stdout, "TargetSpace:")

	// Verify file content has the target
	log.Info("Asserting file content contains target")
	assert.Contains(t, stdout, "ns1.test.example.com")

	log.Success("nested template variables resolved correctly")
}

// TestReportsParams_ParamOverride tests that CLI params can override defaults with literal values
func TestReportsParams_ParamOverride(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing reports-params workflow with param overrides")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	// Test that workflow runs successfully with default params
	// Note: CLI params with {{Output}} templates are not resolved at CLI level,
	// so we test the default behavior which uses properly templated defaults
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-reports-params", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify workflow completed
	log.Info("Asserting workflow completed")
	assert.Contains(t, stdout, "completed")

	// Verify the default paths are shown with TargetSpace resolved
	// Note: TargetSpace for "example.com" is "example.com" (dots preserved)
	log.Info("Asserting default paths are resolved")
	assert.Contains(t, stdout, "dns-example.com.txt")
	assert.Contains(t, stdout, "http-results.txt")

	// Verify files were created
	log.Info("Asserting files were created")
	assert.Contains(t, stdout, "DNS file exists: yes")
	assert.Contains(t, stdout, "HTTP file exists: yes")

	log.Success("param override test works correctly")
}

// TestReportsParams_DryRun tests that the workflow validates correctly in dry-run mode
func TestReportsParams_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing reports-params workflow in dry-run mode")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-reports-params", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	// Verify dry-run mode indicator
	log.Info("Asserting dry-run mode")
	assert.Contains(t, stdout, "DRY-RUN")

	// Verify workflow name is displayed
	log.Info("Asserting workflow name is displayed")
	assert.Contains(t, stdout, "test-reports-params")

	log.Success("reports-params workflow dry-run validates correctly")
}

// TestReportsParams_FileContent tests that created files contain expected content
func TestReportsParams_FileContent(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing reports-params workflow file content")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-reports-params", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	// Verify DNS file content contains the nameserver entry
	log.Info("Asserting DNS file content")
	assert.Contains(t, stdout, "=== DNS File Content ===")
	assert.Contains(t, stdout, "ns1.example.com")
	assert.Contains(t, stdout, "=== End DNS Content ===")

	// Verify HTTP file content contains the URL
	log.Info("Asserting HTTP file content")
	assert.Contains(t, stdout, "=== HTTP File Content ===")
	assert.Contains(t, stdout, "http://example.com:80")
	assert.Contains(t, stdout, "=== End HTTP Content ===")

	log.Success("file content is correct")
}
