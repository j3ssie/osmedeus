package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_JSONLUtilsWorkflow(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing JSONL utility functions via workflow execution")

	workflowPath := getTestdataPath(t)
	workspacesDir := t.TempDir()
	_, stdout, stderr, err := runCLIWithLogAndBase(t, log, "run", "-m", "test-jsonl-utils", "-t", "example.com", "-W", workspacesDir, "-F", workflowPath)
	require.NoError(t, err, "run failed: %s", stderr)
	assert.Contains(t, stdout, "Status: completed")

	outputDir := filepath.Join(workspacesDir, "example.com")

	filteredPath := filepath.Join(outputDir, "filtered.jsonl")
	uniquePath := filepath.Join(outputDir, "unique.jsonl")
	csvPath := filepath.Join(outputDir, "out.csv")
	backPath := filepath.Join(outputDir, "back.jsonl")

	assert.FileExists(t, filteredPath)
	assert.FileExists(t, uniquePath)
	assert.FileExists(t, csvPath)
	assert.FileExists(t, backPath)

	filteredBytes, err := os.ReadFile(filteredPath)
	require.NoError(t, err)
	filteredLines := strings.Split(strings.TrimSpace(string(filteredBytes)), "\n")
	require.Len(t, filteredLines, 3)
	assert.Contains(t, filteredLines[0], "\"name\"")
	assert.Contains(t, filteredLines[0], "hash.body_sha256")
	assert.NotContains(t, filteredLines[1], "hash.body_sha256")

	uniqueBytes, err := os.ReadFile(uniquePath)
	require.NoError(t, err)
	uniqueLines := strings.Split(strings.TrimSpace(string(uniqueBytes)), "\n")
	require.Len(t, uniqueLines, 2)

	csvBytes, err := os.ReadFile(csvPath)
	require.NoError(t, err)
	csvLines := strings.Split(strings.TrimSpace(string(csvBytes)), "\n")
	require.GreaterOrEqual(t, len(csvLines), 2)
	// Columns should be in the order they appear in the first JSON object
	assert.Equal(t, "name,age,hash", csvLines[0])

	backBytes, err := os.ReadFile(backPath)
	require.NoError(t, err)
	backLines := strings.Split(strings.TrimSpace(string(backBytes)), "\n")
	require.Len(t, backLines, 3)
}
