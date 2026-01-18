package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateRendering_PreCondition(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing pre_condition template rendering")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "template-rendering-module", "-t", "precond.test",
		"-F", workflowPath, "--debug")

	require.NoError(t, err, "module execution failed: %s", stderr)

	// Verify pre_condition was rendered - output contains target
	assert.Contains(t, stdout, "precond.test")
	assert.Contains(t, stdout, "Status: completed")

	// Verify template vars rendered in output (not raw {{variables}})
	assert.NotContains(t, stdout, "={{Target}}")
	assert.NotContains(t, stdout, "={{customPath}}")

	log.Success("pre_condition templates rendered correctly")
}

func TestTemplateRendering_Exports(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing exports template rendering")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "template-rendering-module", "-t", "export.test",
		"-F", workflowPath)

	require.NoError(t, err, "module execution failed: %s", stderr)

	// Verify exports contain rendered values (not raw templates)
	assert.NotContains(t, stdout, "={{Target}}")
	assert.NotContains(t, stdout, "={{customPath}}")
	assert.Contains(t, stdout, "bash_target=export.test")
	assert.Contains(t, stdout, "function_result=processed-export.test")
	assert.Contains(t, stdout, "Status: completed")

	log.Success("export templates rendered correctly")
}

func TestTemplateRendering_ForeachInput(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing foreach input path template rendering")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "template-foreach-module", "-t", "foreach.test",
		"-F", workflowPath)

	require.NoError(t, err, "module execution failed: %s", stderr)

	// Verify foreach completed successfully
	assert.Contains(t, stdout, "Status: completed")
	// Verify the template was rendered in the foreach input path
	assert.Contains(t, stdout, "items-foreach.test.txt")
	assert.Contains(t, stdout, "processed-foreach.test")

	// Verify template vars rendered in output
	assert.Contains(t, stdout, "foreach.test")
	assert.NotContains(t, stdout, "={{inputFile}}")
	assert.NotContains(t, stdout, "={{outputDir}}")

	log.Success("foreach input templates rendered correctly")
}

func TestTemplateRendering_ParallelSteps(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing parallel-steps template rendering")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "template-parallel-module", "-t", "parallel.test",
		"-F", workflowPath)

	require.NoError(t, err, "module execution failed: %s", stderr)

	// Verify parallel steps completed successfully
	assert.Contains(t, stdout, "Status: completed")

	// Verify individual parallel step exports were rendered
	assert.Contains(t, stdout, "p1_done=true")
	assert.Contains(t, stdout, "p2_done=true")
	assert.Contains(t, stdout, "p3_done=true")

	// Verify template vars rendered
	assert.NotContains(t, stdout, "={{Target}}")
	assert.NotContains(t, stdout, "={{parallelOutput}}")

	log.Success("parallel-steps templates rendered correctly")
}

func TestTemplateRendering_NestedFlow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing complete nested flow with all template rendering")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-f", "template-rendering-flow", "-t", "flow.example.com",
		"-F", workflowPath)

	require.NoError(t, err, "flow execution failed: %s", stderr)

	// Verify all modules completed
	assert.Contains(t, stdout, "Status: completed")

	// Verify no unrendered templates in output
	assert.NotContains(t, stdout, "={{Target}}")
	assert.NotContains(t, stdout, "={{Output}}")
	assert.NotContains(t, stdout, "={{TargetSpace}}")

	// Verify target appears in output (template was rendered)
	assert.Contains(t, stdout, "flow.example.com")

	log.Success("nested flow templates rendered correctly across all modules")
}

func TestTemplateRendering_ParamWithTemplates(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing param defaults with template variables")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "template-rendering-module", "-t", "param.test",
		"-F", workflowPath, "--debug")

	require.NoError(t, err, "module execution failed: %s", stderr)

	// Verify param defaults were rendered with target
	assert.Contains(t, stdout, "param.test")
	assert.Contains(t, stdout, "Status: completed")
	// Verify target is rendered in custom path
	assert.Contains(t, stdout, "custom-param.test.txt")
	assert.NotContains(t, stdout, "={{TargetSpace}}")

	log.Success("param template defaults rendered correctly")
}

func TestTemplateRendering_ExportsChaining(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing exports chaining with template rendering")

	workflowPath := getTestdataPath(t)
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log,
		"run", "-m", "template-rendering-module", "-t", "chain.test",
		"-F", workflowPath)

	require.NoError(t, err, "module execution failed: %s", stderr)

	// Verify exports are properly chained (function step uses bash_output_path export)
	assert.Contains(t, stdout, "function_result=processed-chain.test")
	assert.Contains(t, stdout, "parallel_done=true")
	assert.Contains(t, stdout, "Status: completed")

	// Verify no raw template syntax in exports
	assert.NotContains(t, stdout, "={{customPath}}")
	assert.NotContains(t, stdout, "bash_output_path={{")

	log.Success("exports chaining works correctly with template rendering")
}
