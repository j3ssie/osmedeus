package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNestedFlow_TargetSharing(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested flow target sharing")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "testing-nested-flow", "-t", "example.com", "-F", workflowPath)

	require.NoError(t, err, "flow execution failed: %s", stderr)
	assert.Contains(t, stdout, "Status: completed")

	// Verify both modules received the target
	assert.Contains(t, stdout, "Module 1: Target=example.com")
	assert.Contains(t, stdout, "Module 2: Target=example.com")

	log.Success("target shared correctly across nested modules")
}

func TestNestedFlow_ParamFromFlow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested flow param inheritance")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "testing-nested-flow", "-t", "example.com", "-F", workflowPath)

	require.NoError(t, err, "flow execution failed: %s", stderr)

	// Verify both modules received the flow param
	assert.Contains(t, stdout, "Module 1: paramFromFlowFile=flow-value")
	assert.Contains(t, stdout, "Module 2: paramFromFlowFile=flow-value")

	log.Success("flow params inherited by nested modules")
}

func TestNestedFlow_ParamOverride(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested flow param override via CLI")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "testing-nested-flow", "-t", "example.com",
		"-p", "paramFromFlowFile=cli-override", "-F", workflowPath)

	require.NoError(t, err, "flow execution failed: %s", stderr)

	// Verify CLI param overrides flow default
	assert.Contains(t, stdout, "paramFromFlowFile=cli-override")

	log.Success("CLI param overrides flow default")
}

func TestNestedFlow_ExportPropagation(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested flow export propagation between modules")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "testing-nested-flow", "-t", "example.com", "-F", workflowPath)

	require.NoError(t, err, "flow execution failed: %s", stderr)

	// Module 2 should see Module 1's export
	assert.Contains(t, stdout, "Module 2: module1_completed=true")

	log.Success("exports propagated between modules")
}

func TestNestedFlow_DependencyOrder(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested flow dependency ordering")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "testing-nested-flow", "-t", "example.com", "-F", workflowPath)

	require.NoError(t, err, "flow execution failed: %s", stderr)

	// Module 1 should appear before Module 2 in output
	idx1 := strings.Index(stdout, "Module 1:")
	idx2 := strings.Index(stdout, "Module 2:")
	assert.True(t, idx1 >= 0, "Module 1 output not found")
	assert.True(t, idx2 >= 0, "Module 2 output not found")
	assert.True(t, idx1 < idx2, "Module 1 should execute before Module 2")

	log.Success("dependency order respected")
}

func TestNestedFlow_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested flow dry-run mode")

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log,
		"run", "-f", "testing-nested-flow", "-t", "example.com",
		"--dry-run", "-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "DRY-RUN")

	log.Success("dry-run mode works for nested flows")
}
