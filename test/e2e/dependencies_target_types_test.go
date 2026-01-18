package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_DependencyTargetTypes_AllowsDomain(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing dependencies.target_types allows domain target")

	workflowPath := getTestdataPath(t)
	stdout, stderr, err := runCLIWithLog(t, log, "run", "-m", "test-target-types", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err, "run failed: %s", stderr)
	assert.Contains(t, stdout, "DRY-RUN")
}

func TestRun_DependencyTargetTypes_AllowsURL(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing dependencies.target_types allows url target")

	workflowPath := getTestdataPath(t)
	stdout, stderr, err := runCLIWithLog(t, log, "run", "-m", "test-target-types", "-t", "https://example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err, "run failed: %s", stderr)
	assert.Contains(t, stdout, "DRY-RUN")
}

func TestRun_DependencyTargetTypes_RejectsOther(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing dependencies.target_types rejects non-matching target")

	workflowPath := getTestdataPath(t)
	_, stderr, err := runCLIWithLog(t, log, "run", "-m", "test-target-types", "-t", "not-a-domain", "--dry-run", "-F", workflowPath)
	assert.Error(t, err)
	assert.Contains(t, stderr, "dependency")
	assert.Contains(t, stderr, "required types")
}
