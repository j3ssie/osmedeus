package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflow_DefaultToList(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow default to list command")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains Total")
	assert.Contains(t, stdout, "Total:")

	log.Success("workflow default lists workflows")
}

func TestWorkflow_List(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow list command")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "list", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains Total and test-bash")
	assert.Contains(t, stdout, "Total:")
	assert.Contains(t, stdout, "test-bash")

	log.Success("workflow list displays workflows correctly")
}

func TestWorkflow_Show(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow show command")

	workflowPath := getTestdataPath(t)
	log.Info("Showing workflow: test-bash")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-bash", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains Metadata and Steps (table format is default)")
	assert.Contains(t, stdout, "Metadata:")
	assert.Contains(t, stdout, "Steps:")

	log.Success("workflow show displays workflow details")
}

func TestWorkflow_Show_Verbose(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow show verbose command")

	workflowPath := getTestdataPath(t)
	log.Info("Showing workflow with verbose flag: test-bash")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-bash", "-v", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains Variable and Description (table format with verbose)")
	assert.Contains(t, stdout, "Variable")
	assert.Contains(t, stdout, "Description")

	log.Success("workflow show verbose displays extra details")
}

func TestWorkflow_Show_Yaml(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow show YAML output")

	workflowPath := getTestdataPath(t)
	log.Info("Showing workflow as YAML: test-bash")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-bash", "--yaml", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains YAML fields (requires --yaml flag)")
	assert.Contains(t, stdout, "name:")
	assert.Contains(t, stdout, "kind:")
	assert.Contains(t, stdout, "steps:")

	log.Success("workflow show YAML outputs valid YAML")
}

func TestWorkflow_Show_NotFound(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow show with nonexistent workflow")

	workflowPath := getTestdataPath(t)
	log.Info("Attempting to show nonexistent workflow")

	_, stderr, err := runCLIWithLog(t, log, "workflow", "show", "nonexistent-workflow", "-F", workflowPath)

	log.Info("Asserting command returns error")
	assert.Error(t, err)

	log.Info("Asserting stderr contains failure message")
	assert.Contains(t, stderr, "Failed")

	log.Success("workflow show correctly reports not found error")
}

func TestWorkflow_Validate_Success(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow validate with valid workflow")

	workflowPath := getTestdataPath(t)
	log.Info("Validating workflow: test-bash")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-bash", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains validation success")
	assert.Contains(t, stdout, "is valid")

	log.Success("workflow validate reports valid workflow")
}

func TestWorkflow_Validate_Fail(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing workflow validate with invalid workflow")

	tmpDir := t.TempDir()
	log.Info("Creating invalid workflow in temp dir: %s", tmpDir)

	// Create invalid workflow YAML
	invalidYAML := `name: invalid-workflow
kind: module
steps:
  - name: missing-type
    command: echo hello
`
	err := os.WriteFile(filepath.Join(tmpDir, "invalid.yaml"), []byte(invalidYAML), 0644)
	require.NoError(t, err)

	log.Info("Validating invalid workflow")
	_, stderr, err := runCLIWithLog(t, log, "workflow", "validate", "invalid", "-F", tmpDir)

	log.Info("Asserting command returns error")
	assert.Error(t, err)

	log.Info("Asserting stderr mentions type error")
	assert.Contains(t, stderr, "type")

	log.Success("workflow validate correctly reports invalid workflow")
}
