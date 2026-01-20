package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtends_ValidateBase(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing validate base workflow")

	workflowPath := getTestdataPath(t)
	log.Info("Validating base workflow: test-extends-base")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-extends-base", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passed")
	// Accept either old "is valid" or new "passed" message
	assert.True(t, strings.Contains(stdout, "is valid") || strings.Contains(stdout, "passed"),
		"Expected validation success message")

	log.Success("base workflow validated successfully")
}

func TestExtends_ValidateChild(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing validate child workflow with extends")

	workflowPath := getTestdataPath(t)
	log.Info("Validating child workflow: test-extends-fast")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-extends-fast", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passed")
	// Accept either old "is valid" or new "passed" message
	assert.True(t, strings.Contains(stdout, "is valid") || strings.Contains(stdout, "passed"),
		"Expected validation success message")

	log.Success("child workflow with extends validated successfully")
}

func TestExtends_ShowChild(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow show for child workflow")

	workflowPath := getTestdataPath(t)
	log.Info("Showing child workflow: test-extends-fast")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-extends-fast", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow details are shown with inherited steps")
	assert.Contains(t, stdout, "test-extends-fast")
	assert.Contains(t, stdout, "Steps:")

	log.Success("child workflow show displays inherited content")
}

func TestExtends_ShowChildYAML(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow show YAML for child workflow")

	workflowPath := getTestdataPath(t)
	log.Info("Showing child workflow as YAML: test-extends-fast")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-extends-fast", "--yaml", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting YAML shows the child workflow with extends reference")
	assert.Contains(t, stdout, "name: test-extends-fast")
	// The YAML output shows the source file, which includes the extends field
	assert.Contains(t, stdout, "extends: test-extends-base")
	// Should show the override section
	assert.Contains(t, stdout, "override:")
	assert.Contains(t, stdout, "threads:")

	log.Success("child workflow YAML shows extends reference and overrides")
}

func TestExtends_DryRunBase(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing dry-run with base workflow")

	workflowPath := getTestdataPath(t)
	log.Info("Running base workflow in dry-run mode")

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-extends-base", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode is indicated")
	assert.Contains(t, stdout, "DRY-RUN")
	// Verify the workflow name is shown
	assert.Contains(t, stdout, "test-extends-base")

	log.Success("base workflow dry-run works correctly")
}

func TestExtends_DryRunChildFast(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing dry-run with child workflow (fast variant)")

	workflowPath := getTestdataPath(t)
	log.Info("Running fast child workflow in dry-run mode")

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-extends-fast", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and inherited steps")
	assert.Contains(t, stdout, "DRY-RUN")
	// Verify inherited steps are shown
	assert.Contains(t, stdout, "show-config")
	assert.Contains(t, stdout, "scan-step")

	log.Success("fast child workflow dry-run shows inherited steps")
}

func TestExtends_DryRunChildAggressive(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing dry-run with child workflow (aggressive variant)")

	workflowPath := getTestdataPath(t)
	log.Info("Running aggressive child workflow in dry-run mode")

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-extends-aggressive", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and inherited steps")
	assert.Contains(t, stdout, "DRY-RUN")
	// Verify inherited steps are shown
	assert.Contains(t, stdout, "show-config")
	assert.Contains(t, stdout, "scan-step")

	log.Success("aggressive child workflow dry-run shows inherited steps")
}

func TestExtends_RunChildFast(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing actual run with child workflow (fast variant)")

	workflowPath := getTestdataPath(t)
	log.Info("Running fast child workflow")

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-extends-fast", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting run completed with overridden param values in output")
	// The echo commands should show the overridden values
	assert.Contains(t, stdout, "Threads=5")
	assert.Contains(t, stdout, "RateLimit=50")
	assert.Contains(t, stdout, "5 threads at rate 50")

	log.Success("fast child workflow executed with correct param overrides")
}

func TestExtends_RunChildAggressive(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing actual run with child workflow (aggressive variant)")

	workflowPath := getTestdataPath(t)
	log.Info("Running aggressive child workflow")

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-extends-aggressive", "-t", "example.com", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting run completed with overridden param values in output")
	// The echo commands should show the overridden values
	assert.Contains(t, stdout, "Threads=50")
	assert.Contains(t, stdout, "RateLimit=500")
	assert.Contains(t, stdout, "Verbose=true")
	assert.Contains(t, stdout, "50 threads at rate 500")

	log.Success("aggressive child workflow executed with correct param overrides")
}

func TestExtends_ParamOverrideFromCLI(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing CLI param override on top of extends override")

	workflowPath := getTestdataPath(t)
	log.Info("Running fast child workflow with CLI param override")

	// Override threads from CLI (should override both base and child defaults)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-extends-fast", "-t", "example.com",
		"-p", "threads=99", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting CLI param takes precedence over extends override")
	// CLI override should win: threads=99
	assert.Contains(t, stdout, "Threads=99")
	// rate_limit should still be from child override: 50
	assert.Contains(t, stdout, "RateLimit=50")

	log.Success("CLI param override takes precedence over extends")
}

func TestExtends_ListShowsChildWorkflows(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow list includes child workflows")

	workflowPath := getTestdataPath(t)
	log.Info("Listing all workflows")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "list", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting list contains both base and child workflows")
	assert.Contains(t, stdout, "test-extends-base")
	assert.Contains(t, stdout, "test-extends-fast")
	assert.Contains(t, stdout, "test-extends-aggressive")

	log.Success("workflow list shows all extends-related workflows")
}
