package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Help(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing run help command")

	stdout, _, err := runCLIWithLog(t, log, "run", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains required flags")
	assert.Contains(t, stdout, "--flow")
	assert.Contains(t, stdout, "--module")
	assert.Contains(t, stdout, "--target")

	log.Success("run help displays all required flags")
}

func TestRun_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing run dry-run mode")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-bash", "-t", "test.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains DRY-RUN indicator")
	assert.Contains(t, stdout, "DRY-RUN")

	log.Success("dry-run mode works correctly")
}

func TestRun_NoTarget(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing run without target (should fail)")

	workflowPath := getTestdataPath(t)
	_, stderr, err := runCLIWithLog(t, log, "run", "-m", "test-bash", "-F", workflowPath)

	log.Info("Asserting command returns error")
	assert.Error(t, err)

	log.Info("Asserting stderr mentions missing target")
	assert.Contains(t, stderr, "target")

	log.Success("missing target correctly reports error")
}

func TestRun_Module(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing run with module workflow")

	workflowPath := getTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-bash", "-t", "test.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains DRY-RUN and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-bash")

	log.Success("module workflow executed correctly")
}

func TestRun_WithParams(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing run with custom parameters")

	workflowPath := getTestdataPath(t)
	log.Info("Custom param: custom_key=custom_value")

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-bash", "-t", "test.com", "-p", "custom_key=custom_value", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains DRY-RUN indicator")
	assert.Contains(t, stdout, "DRY-RUN")

	log.Success("custom parameters accepted")
}

func TestRun_MultipleTargets(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing run with multiple targets")

	workflowPath := getTestdataPath(t)
	log.Info("Targets: target1.com, target2.com")

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-bash", "-t", "target1.com", "-t", "target2.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout indicates 2 targets")
	assert.Contains(t, stdout, "2 targets")

	log.Success("multiple targets processed correctly")
}
