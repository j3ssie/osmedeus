package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlowName_InlineModules(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing FlowName variable with inline modules")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "test-flowname-inline", "-t", "flowname.test",
		"-F", workflowPath)

	require.NoError(t, err, "flow execution failed: %s", stderr)

	// Verify flow completed successfully
	assert.Contains(t, stdout, "Status: completed")

	// Verify FlowName variable is rendered correctly in inline modules
	// The flow name should be "test-flowname-inline"
	assert.Contains(t, stdout, "test-flowname-inline")
	assert.Contains(t, stdout, "Hello from test-flowname-inline")
	assert.Contains(t, stdout, "FlowName=test-flowname-inline")
	assert.Contains(t, stdout, "Flow test-flowname-inline completed")

	// Verify no unrendered FlowName template
	assert.NotContains(t, stdout, "={{FlowName}}")
	assert.NotContains(t, stdout, "{{FlowName}}")

	// Verify target is also rendered
	assert.Contains(t, stdout, "flowname.test")

	log.Success("FlowName variable rendered correctly in inline modules")
}

func TestFlowName_EmptyWhenDirectModule(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing FlowName is empty when running module directly")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "test-bash", "-t", "direct.test",
		"-F", workflowPath)

	require.NoError(t, err, "module execution failed: %s", stderr)

	// Verify module completed successfully
	assert.Contains(t, stdout, "Status: completed")

	// When running a module directly (not through a flow),
	// FlowName should be empty
	assert.NotContains(t, stdout, "={{FlowName}}")

	log.Success("FlowName is empty when running module directly")
}
