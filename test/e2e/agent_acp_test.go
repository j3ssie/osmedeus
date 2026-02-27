package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Basic agent-acp workflow tests ---

func TestRun_AgentACP_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent-acp module in dry-run mode")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-acp", "-t", "test.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains DRY-RUN indicator and agent-acp steps")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-acp")

	log.Success("agent-acp module dry-run works correctly")
}

func TestRun_AgentACP_WorkflowValidate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent-acp workflow validation")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Validating test-agent-acp workflow")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-acp", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-acp")

	log.Success("agent-acp workflow validates successfully")
}

func TestRun_AgentACP_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent-acp workflow show")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Showing test-agent-acp workflow details")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-acp", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow details contain agent-acp step information")
	assert.Contains(t, stdout, "test-agent-acp")
	assert.Contains(t, stdout, "acp-basic")
	assert.Contains(t, stdout, "agent-acp")

	log.Success("agent-acp workflow show works correctly")
}

// --- Minimal agent-acp workflow tests ---

func TestRun_AgentACPMinimal_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing minimal agent-acp module in dry-run mode")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-acp-minimal", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains DRY-RUN indicator")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-acp-minimal")

	log.Success("minimal agent-acp module dry-run works correctly")
}

func TestRun_AgentACPMinimal_WorkflowValidate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing minimal agent-acp workflow validation")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Validating test-agent-acp-minimal workflow")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-acp-minimal", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-acp-minimal")

	log.Success("minimal agent-acp workflow validates successfully")
}

func TestRun_AgentACPMinimal_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing minimal agent-acp workflow show")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Showing test-agent-acp-minimal workflow details")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-acp-minimal", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow details contain minimal agent-acp step")
	assert.Contains(t, stdout, "quick-acp-agent")
	assert.Contains(t, stdout, "agent-acp")

	log.Success("minimal agent-acp workflow show works correctly")
}

// --- Configured agent-acp workflow tests ---

func TestRun_AgentACPConfig_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing configured agent-acp workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-acp-config", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-acp-config")

	log.Success("configured agent-acp dry-run works correctly")
}

func TestRun_AgentACPConfig_WorkflowValidate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing configured agent-acp workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-acp-config", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-acp-config")

	log.Success("configured agent-acp workflow validates successfully")
}

func TestRun_AgentACPConfig_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing configured agent-acp workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-acp-config", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow show contains configured agent-acp steps")
	assert.Contains(t, stdout, "acp-configured")
	assert.Contains(t, stdout, "verify-acp-exports")
	assert.Contains(t, stdout, "agent-acp")
	assert.Contains(t, stdout, "bash")

	log.Success("configured agent-acp workflow show works correctly")
}

// --- Codex agent-acp workflow tests ---

func TestRun_AgentACPCodex_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing codex agent-acp workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-acp-codex", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-acp-codex")

	log.Success("codex agent-acp dry-run works correctly")
}

func TestRun_AgentACPCodex_WorkflowValidate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing codex agent-acp workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-acp-codex", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-acp-codex")

	log.Success("codex agent-acp workflow validates successfully")
}

func TestRun_AgentACPCodex_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing codex agent-acp workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-acp-codex", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow show contains codex agent-acp step")
	assert.Contains(t, stdout, "acp-codex")
	assert.Contains(t, stdout, "agent-acp")

	log.Success("codex agent-acp workflow show works correctly")
}
