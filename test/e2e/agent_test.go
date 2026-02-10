package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_AgentModule_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent module in dry-run mode")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent", "-t", "test.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains DRY-RUN indicator and agent steps")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent")

	log.Success("agent module dry-run works correctly")
}

func TestRun_AgentMinimalModule_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing minimal agent module in dry-run mode")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Using workflow path: %s", workflowPath)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-minimal", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains DRY-RUN indicator")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-minimal")

	log.Success("minimal agent module dry-run works correctly")
}

func TestRun_AgentModule_WorkflowValidate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent workflow validation")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Validating test-agent workflow")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent")

	log.Success("agent workflow validates successfully")
}

func TestRun_AgentMinimalModule_WorkflowValidate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing minimal agent workflow validation")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Validating test-agent-minimal workflow")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-minimal", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-minimal")

	log.Success("minimal agent workflow validates successfully")
}

func TestRun_AgentModule_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent workflow show")

	workflowPath := getAgentTestdataPath(t)
	log.Info("Showing test-agent workflow details")

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow details contain agent step information")
	assert.Contains(t, stdout, "test-agent")
	assert.Contains(t, stdout, "agent")

	log.Success("agent workflow show works correctly")
}

// --- Custom tools workflow tests ---

func TestRun_AgentCustomTools_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent custom tools workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-custom-tools", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-custom-tools")

	log.Success("agent custom tools dry-run works correctly")
}

func TestRun_AgentCustomTools_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent custom tools workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-custom-tools", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-custom-tools")

	log.Success("agent custom tools workflow validates successfully")
}

func TestRun_AgentCustomTools_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent custom tools workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-custom-tools", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow show contains agent steps and bash step")
	assert.Contains(t, stdout, "agent-custom-handler")
	assert.Contains(t, stdout, "verify-custom-tool")
	assert.Contains(t, stdout, "agent")
	assert.Contains(t, stdout, "bash")

	log.Success("agent custom tools workflow show works correctly")
}

// --- Exports workflow tests ---

func TestRun_AgentExports_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent exports workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-exports", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-exports")

	log.Success("agent exports dry-run works correctly")
}

func TestRun_AgentExports_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent exports workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-exports", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-exports")

	log.Success("agent exports workflow validates successfully")
}

func TestRun_AgentExports_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent exports workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-exports", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow show contains agent and verify steps")
	assert.Contains(t, stdout, "agent-with-exports")
	assert.Contains(t, stdout, "verify-exports")
	assert.Contains(t, stdout, "agent")

	log.Success("agent exports workflow show works correctly")
}

// --- Validation failure tests ---

func TestRun_AgentValidationFail_DuplicateTools(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent workflow validation fails for duplicate tool names")

	workflowPath := getAgentTestdataPath(t)

	_, stderr, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-validation-fail", "-F", workflowPath)

	log.Info("Asserting validation returns error")
	assert.Error(t, err)

	log.Info("Asserting error mentions duplicate tool")
	assert.Contains(t, stderr, "duplicate")

	log.Success("agent workflow correctly rejects duplicate tool names")
}

func TestRun_AgentValidationFail_UnknownPreset(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent workflow validation fails for unknown preset tool")

	workflowPath := getAgentTestdataPath(t)

	_, stderr, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-unknown-preset", "-F", workflowPath)

	log.Info("Asserting validation returns error")
	assert.Error(t, err)

	log.Info("Asserting error mentions unknown preset")
	assert.Contains(t, stderr, "unknown preset tool")

	log.Success("agent workflow correctly rejects unknown preset tools")
}

// --- Planning workflow tests ---

func TestRun_AgentPlanning_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent planning workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-planning", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-planning")

	log.Success("agent planning dry-run works correctly")
}

func TestRun_AgentPlanning_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent planning workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-planning", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-planning")

	log.Success("agent planning workflow validates successfully")
}

func TestRun_AgentPlanning_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent planning workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-planning", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow show contains planning agent steps")
	assert.Contains(t, stdout, "agent-with-plan")
	assert.Contains(t, stdout, "verify-plan")
	assert.Contains(t, stdout, "agent")

	log.Success("agent planning workflow show works correctly")
}

// --- Multi-goal workflow tests ---

func TestRun_AgentMultiGoal_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent multi-goal workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-multi-goal", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-multi-goal")

	log.Success("agent multi-goal dry-run works correctly")
}

func TestRun_AgentMultiGoal_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent multi-goal workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-multi-goal", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-multi-goal")

	log.Success("agent multi-goal workflow validates successfully")
}

func TestRun_AgentMultiGoal_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent multi-goal workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-multi-goal", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow show contains multi-goal steps")
	assert.Contains(t, stdout, "multi-goal-agent")
	assert.Contains(t, stdout, "verify-goals")
	assert.Contains(t, stdout, "agent")

	log.Success("agent multi-goal workflow show works correctly")
}

// --- Structured output workflow tests ---

func TestRun_AgentStructured_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent structured output workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-structured", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-structured")

	log.Success("agent structured output dry-run works correctly")
}

func TestRun_AgentStructured_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent structured output workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-structured", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-structured")

	log.Success("agent structured output workflow validates successfully")
}

// --- Tracing hooks workflow tests ---

func TestRun_AgentTracing_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent tracing hooks workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-tracing", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting dry-run mode and workflow name")
	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-tracing")

	log.Success("agent tracing dry-run works correctly")
}

func TestRun_AgentTracing_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent tracing hooks workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-tracing", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting validation passes")
	assert.Contains(t, stdout, "test-agent-tracing")

	log.Success("agent tracing workflow validates successfully")
}

func TestRun_AgentTracing_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent tracing hooks workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-tracing", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting workflow show contains tracing agent steps")
	assert.Contains(t, stdout, "traced-agent")
	assert.Contains(t, stdout, "agent")

	log.Success("agent tracing workflow show works correctly")
}

// --- Validation failure tests ---

// --- File tools workflow tests ---

func TestRun_AgentFileTools_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent file tools workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-file-tools", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-file-tools")

	log.Success("agent file tools dry-run works correctly")
}

func TestRun_AgentFileTools_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent file tools workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-file-tools", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "test-agent-file-tools")

	log.Success("agent file tools workflow validates successfully")
}

func TestRun_AgentFileTools_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent file tools workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-file-tools", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "agent-file-tools")
	assert.Contains(t, stdout, "verify-file-tools")
	assert.Contains(t, stdout, "agent")

	log.Success("agent file tools workflow show works correctly")
}

// --- Orchestration workflow tests ---

func TestRun_AgentOrchestration_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent orchestration workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-orchestration", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-orchestration")

	log.Success("agent orchestration dry-run works correctly")
}

func TestRun_AgentOrchestration_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent orchestration workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-orchestration", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "test-agent-orchestration")

	log.Success("agent orchestration workflow validates successfully")
}

func TestRun_AgentOrchestration_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent orchestration workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-orchestration", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "agent-orchestrator")
	assert.Contains(t, stdout, "verify-orchestration")
	assert.Contains(t, stdout, "agent")

	log.Success("agent orchestration workflow show works correctly")
}

// --- Python tools workflow tests ---

func TestRun_AgentPythonTools_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent Python tools workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-python-tools", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-python-tools")

	log.Success("agent Python tools dry-run works correctly")
}

func TestRun_AgentPythonTools_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent Python tools workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-python-tools", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "test-agent-python-tools")

	log.Success("agent Python tools workflow validates successfully")
}

func TestRun_AgentPythonTools_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent Python tools workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-python-tools", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "agent-python")
	assert.Contains(t, stdout, "verify-python-tools")
	assert.Contains(t, stdout, "agent")

	log.Success("agent Python tools workflow show works correctly")
}

// --- Sub-agents workflow tests ---

func TestRun_AgentSubAgents_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent sub-agents workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-sub-agents", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-sub-agents")

	log.Success("agent sub-agents dry-run works correctly")
}

func TestRun_AgentSubAgents_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent sub-agents workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-sub-agents", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "test-agent-sub-agents")

	log.Success("agent sub-agents workflow validates successfully")
}

func TestRun_AgentSubAgents_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent sub-agents workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-sub-agents", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "orchestrator")
	assert.Contains(t, stdout, "agent")

	log.Success("agent sub-agents workflow show works correctly")
}

// --- Nested sub-agents workflow tests ---

func TestRun_AgentSubAgentsNested_DryRun(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested sub-agents workflow in dry-run mode")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-agent-sub-agents-nested", "-t", "example.com", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "DRY-RUN")
	assert.Contains(t, stdout, "test-agent-sub-agents-nested")

	log.Success("nested sub-agents dry-run works correctly")
}

func TestRun_AgentSubAgentsNested_Validate(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested sub-agents workflow validation")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-sub-agents-nested", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "test-agent-sub-agents-nested")

	log.Success("nested sub-agents workflow validates successfully")
}

func TestRun_AgentSubAgentsNested_WorkflowShow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing nested sub-agents workflow show")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "show", "test-agent-sub-agents-nested", "-F", workflowPath)
	require.NoError(t, err)

	assert.Contains(t, stdout, "nested-orchestrator")
	assert.Contains(t, stdout, "agent")

	log.Success("nested sub-agents workflow show works correctly")
}

// --- Sub-agents validation failure test (duplicate names) ---

func TestRun_AgentSubAgentsValidationFail_DuplicateNames(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing sub-agent validation fails for duplicate sub-agent names")

	workflowPath := getAgentTestdataPath(t)

	stdout, _, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-sub-agents-validation-fail", "-F", workflowPath, "--check")

	log.Info("Asserting validation returns error")
	assert.Error(t, err)

	log.Info("Asserting output mentions duplicate sub-agent")
	assert.Contains(t, stdout, "duplicate")

	log.Success("agent workflow correctly rejects duplicate sub-agent names")
}

// --- Validation failure tests ---

func TestRun_AgentValidationFail_MissingDescription(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing agent validation fails for custom tool missing description")

	tmpDir := t.TempDir()

	invalidYAML := `kind: module
name: test-agent-no-desc
description: Agent with custom tool missing description
params:
  - name: target
    required: true
    default: example.com
steps:
  - name: bad-agent
    type: agent
    query: "test query"
    max_iterations: 3
    agent_tools:
      - name: my_tool
        parameters:
          type: object
          properties:
            input:
              type: string
        handler: 'args.input'
`
	err := os.WriteFile(filepath.Join(tmpDir, "test-agent-no-desc.yaml"), []byte(invalidYAML), 0644)
	require.NoError(t, err)

	_, stderr, err := runCLIWithLog(t, log, "workflow", "validate", "test-agent-no-desc", "-F", tmpDir)

	log.Info("Asserting validation returns error")
	assert.Error(t, err)

	log.Info("Asserting error mentions missing description")
	assert.Contains(t, stderr, "description")

	log.Success("agent workflow correctly rejects custom tools without description")
}
