package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHooks_Module_PreAndPost tests that both pre and post scan hooks execute around main steps
func TestHooks_Module_PreAndPost(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing module with pre and post scan hooks")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "test-hooks-module", "-t", "hooks.example.com", "-F", workflowPath)
	require.NoError(t, err, "hook module execution failed: %s", stderr)

	// Verify all hook steps ran
	log.Info("Asserting pre-scan hooks executed")
	assert.Contains(t, stdout, "[HOOK-PRE] Starting scan for hooks.example.com")
	assert.Contains(t, stdout, "[HOOK-PRE] Setup complete")

	log.Info("Asserting main steps executed")
	assert.Contains(t, stdout, "[MAIN] Scanning hooks.example.com")
	assert.Contains(t, stdout, "[MAIN] Analysis complete")

	log.Info("Asserting post-scan hooks executed")
	assert.Contains(t, stdout, "[HOOK-POST] Cleaning up for hooks.example.com")
	assert.Contains(t, stdout, "[HOOK-POST] Scan finished")

	log.Success("pre and post scan hooks both executed")
}

// TestHooks_Module_ExecutionOrder tests that pre hooks run before main steps and post hooks run after
func TestHooks_Module_ExecutionOrder(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing hook execution order: pre -> main -> post")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "test-hooks-module", "-t", "hooks.example.com", "-F", workflowPath)
	require.NoError(t, err, "execution failed: %s", stderr)

	// Verify order: pre hooks appear before main steps, main steps before post hooks
	preIdx := strings.Index(stdout, "[HOOK-PRE] Starting scan")
	mainIdx := strings.Index(stdout, "[MAIN] Scanning")
	postIdx := strings.Index(stdout, "[HOOK-POST] Cleaning up")

	log.Info("Asserting pre-hooks appear before main steps")
	assert.True(t, preIdx >= 0, "pre-hook output not found")
	assert.True(t, mainIdx >= 0, "main step output not found")
	assert.True(t, postIdx >= 0, "post-hook output not found")

	assert.True(t, preIdx < mainIdx, "pre-hook should execute before main steps")
	assert.True(t, mainIdx < postIdx, "main steps should execute before post-hooks")

	log.Success("execution order is correct: pre -> main -> post")
}

// TestHooks_Module_PreOnly tests a module with only pre-scan hooks
func TestHooks_Module_PreOnly(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing module with only pre-scan hooks")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "test-hooks-pre-only", "-t", "hooks.example.com", "-F", workflowPath)
	require.NoError(t, err, "execution failed: %s", stderr)

	log.Info("Asserting pre-scan hook executed")
	assert.Contains(t, stdout, "[HOOK-PRE] Initializing for hooks.example.com")

	log.Info("Asserting main step executed")
	assert.Contains(t, stdout, "[MAIN] Running for hooks.example.com")

	// No post hooks expected
	log.Info("Asserting no post-scan hooks")
	assert.NotContains(t, stdout, "[HOOK-POST]")

	log.Success("pre-only hooks work correctly")
}

// TestHooks_Module_PostOnly tests a module with only post-scan hooks
func TestHooks_Module_PostOnly(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing module with only post-scan hooks")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "test-hooks-post-only", "-t", "hooks.example.com", "-F", workflowPath)
	require.NoError(t, err, "execution failed: %s", stderr)

	// No pre hooks expected
	log.Info("Asserting no pre-scan hooks")
	assert.NotContains(t, stdout, "[HOOK-PRE]")

	log.Info("Asserting main step executed")
	assert.Contains(t, stdout, "[MAIN] Running for hooks.example.com")

	log.Info("Asserting post-scan hook executed")
	assert.Contains(t, stdout, "[HOOK-POST] Notify completion for hooks.example.com")

	log.Success("post-only hooks work correctly")
}

// TestHooks_Module_DefaultType tests that hook steps without a type default to bash
func TestHooks_Module_DefaultType(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing hook steps default to bash type when type is omitted")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "test-hooks-default-type", "-t", "hooks.example.com", "-F", workflowPath)
	require.NoError(t, err, "execution failed: %s", stderr)

	log.Info("Asserting pre-hook with no type still executed as bash")
	assert.Contains(t, stdout, "[HOOK-PRE] Default type for hooks.example.com")

	log.Info("Asserting post-hook with no type still executed as bash")
	assert.Contains(t, stdout, "[HOOK-POST] Default type for hooks.example.com")

	log.Info("Asserting main step executed")
	assert.Contains(t, stdout, "[MAIN] Running for hooks.example.com")

	log.Success("hook steps correctly default to bash type")
}

// TestHooks_Module_FailureNonFatal tests that failing hook steps don't abort the workflow
func TestHooks_Module_FailureNonFatal(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing that failing hook steps don't abort the workflow")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "test-hooks-failure", "-t", "hooks.example.com", "-F", workflowPath)
	require.NoError(t, err, "workflow should not fail due to hook failures: %s", stderr)

	// The main step should still execute despite hook failures
	log.Info("Asserting main step executed despite hook failures")
	assert.Contains(t, stdout, "[MAIN] Should still run despite hook failures")

	// Hook steps after failures should still execute
	log.Info("Asserting subsequent hook steps still run after a failure")
	assert.Contains(t, stdout, "[HOOK-PRE] Runs after failed hook")
	assert.Contains(t, stdout, "[HOOK-POST] Runs after failed hook")

	log.Success("hook failures are non-fatal")
}

// TestHooks_Module_DryRun tests that hook workflows validate in dry-run mode
func TestHooks_Module_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing hooks module in dry-run mode")

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log,
		"run", "-m", "test-hooks-module", "-t", "hooks.example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode indicator")
	assert.Contains(t, stdout, "DRY-RUN")

	log.Info("Asserting workflow name displayed")
	assert.Contains(t, stdout, "test-hooks-module")

	log.Success("hooks module dry-run validates correctly")
}

// TestHooks_Flow_PreAndPost tests that flow-level hooks execute around module execution
func TestHooks_Flow_PreAndPost(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing flow with pre and post scan hooks")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "test-hooks-flow", "-t", "hooks.example.com", "-F", workflowPath)
	require.NoError(t, err, "hook flow execution failed: %s", stderr)

	log.Info("Asserting flow pre-scan hook executed")
	assert.Contains(t, stdout, "[FLOW-HOOK-PRE] Initializing flow for hooks.example.com")

	log.Info("Asserting inline module executed")
	assert.Contains(t, stdout, "[FLOW-MODULE] Inline module executed for hooks.example.com")

	log.Info("Asserting flow post-scan hook executed")
	assert.Contains(t, stdout, "[FLOW-HOOK-POST] Flow completed for hooks.example.com")

	log.Success("flow-level pre and post hooks executed correctly")
}

// TestHooks_Flow_ExecutionOrder tests that flow hooks wrap module execution
func TestHooks_Flow_ExecutionOrder(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing flow hook execution order: pre -> modules -> post")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "test-hooks-flow", "-t", "hooks.example.com", "-F", workflowPath)
	require.NoError(t, err, "execution failed: %s", stderr)

	preIdx := strings.Index(stdout, "[FLOW-HOOK-PRE]")
	moduleIdx := strings.Index(stdout, "[FLOW-MODULE]")
	postIdx := strings.Index(stdout, "[FLOW-HOOK-POST]")

	log.Info("Asserting flow hooks wrap module execution")
	assert.True(t, preIdx >= 0, "flow pre-hook output not found")
	assert.True(t, moduleIdx >= 0, "flow module output not found")
	assert.True(t, postIdx >= 0, "flow post-hook output not found")

	assert.True(t, preIdx < moduleIdx, "flow pre-hook should execute before modules")
	assert.True(t, moduleIdx < postIdx, "modules should execute before flow post-hooks")

	log.Success("flow hook execution order is correct: pre -> modules -> post")
}
