package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDockerFileOutputs tests std_file, step_remote_file, and host_output_file
// with the Docker step runner (remote-bash step type)
func TestDockerFileOutputs(t *testing.T) {
	// Skip if running in short mode (no Docker)
	if testing.Short() {
		t.Skip("Skipping Docker test in short mode")
	}

	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	// Setup temp directory for outputs
	outputDir := t.TempDir()

	workflow, err := loader.LoadWorkflow("test-docker-file-outputs")
	require.NoError(t, err, "Failed to load test-docker-file-outputs workflow")

	ctx := context.Background()
	cfg := testConfig(t)

	exec := executor.NewExecutor()
	exec.SetDryRun(false)
	exec.SetSpinner(false)

	result, err := exec.ExecuteModule(ctx, workflow, map[string]string{
		"target":     "docker-file-test",
		"output_dir": outputDir,
	}, cfg)

	require.NoError(t, err, "Workflow execution failed")
	assert.Equal(t, core.RunStatusCompleted, result.Status, "Expected workflow to complete successfully")

	// Test 1: Verify std_file output
	t.Run("std_file_output", func(t *testing.T) {
		stdFilePath := filepath.Join(outputDir, "std_file_output.txt")
		stdFileContent, err := os.ReadFile(stdFilePath)
		require.NoError(t, err, "Failed to read std_file output: %s", stdFilePath)
		assert.Contains(t, string(stdFileContent), "stdout from docker: docker-file-test",
			"std_file should contain expected stdout")
		assert.Contains(t, string(stdFileContent), "line 2",
			"std_file should contain multi-line output")
	})

	// Test 2: Verify remote file copy (step_remote_file -> host_output_file)
	t.Run("remote_file_copy", func(t *testing.T) {
		remoteCopyPath := filepath.Join(outputDir, "copied_from_container.txt")
		remoteCopyContent, err := os.ReadFile(remoteCopyPath)
		require.NoError(t, err, "Failed to read remote file copy: %s", remoteCopyPath)
		assert.Contains(t, string(remoteCopyContent), "created in container: docker-file-test",
			"Copied file should contain content created in container")
	})

	// Test 3: Verify combined outputs (std_file + remote file copy)
	t.Run("combined_outputs", func(t *testing.T) {
		// Check std_file from combined step
		combinedStdoutPath := filepath.Join(outputDir, "combined_stdout.txt")
		combinedStdout, err := os.ReadFile(combinedStdoutPath)
		require.NoError(t, err, "Failed to read combined stdout: %s", combinedStdoutPath)
		assert.Contains(t, string(combinedStdout), "Processing target: docker-file-test",
			"Combined std_file should contain expected output")

		// Check remote file copy from combined step
		combinedResultPath := filepath.Join(outputDir, "combined_result.txt")
		combinedResult, err := os.ReadFile(combinedResultPath)
		require.NoError(t, err, "Failed to read combined result: %s", combinedResultPath)
		assert.Contains(t, string(combinedResult), "result-data-docker-file-test",
			"Combined remote file should contain expected content")
	})
}

// TestDockerStdFileOnly tests std_file capture without remote file copy
func TestDockerStdFileOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker test in short mode")
	}

	workflowsPath := getWorkflowsPath()
	loader := parser.NewLoader(workflowsPath)

	workflow, err := loader.LoadWorkflow("test-docker-file-outputs")
	require.NoError(t, err)

	// Verify the workflow has the expected steps
	assert.GreaterOrEqual(t, len(workflow.Steps), 3, "Workflow should have at least 3 steps")

	// Verify step configurations
	stdFileStep := workflow.Steps[0]
	assert.Equal(t, "test-std-file", stdFileStep.Name)
	assert.Equal(t, core.StepTypeRemoteBash, stdFileStep.Type)
	assert.NotEmpty(t, stdFileStep.StdFile, "std_file should be set")
	assert.Empty(t, stdFileStep.StepRemoteFile, "step_remote_file should not be set for std_file only test")

	remoteFileStep := workflow.Steps[1]
	assert.Equal(t, "test-remote-file-copy", remoteFileStep.Name)
	assert.NotEmpty(t, remoteFileStep.StepRemoteFile, "step_remote_file should be set")
	assert.NotEmpty(t, remoteFileStep.HostOutputFile, "host_output_file should be set")
}
