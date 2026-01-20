package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_ChunkMode_InfoDisplay(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing chunk info display mode")

	// Create targets file with 10 targets
	tmpDir := t.TempDir()
	targetsFile := filepath.Join(tmpDir, "targets.txt")
	content := "t1\nt2\nt3\nt4\nt5\nt6\nt7\nt8\nt9\nt10\n"
	require.NoError(t, os.WriteFile(targetsFile, []byte(content), 0644))

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-echo",
		"-T", targetsFile, "--chunk-size", "3", "-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "10 total targets")
	assert.Contains(t, stdout, "4 chunks")
	log.Success("Chunk info displayed correctly")
}

func TestE2E_ChunkMode_SpecificChunk(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing specific chunk execution")

	tmpDir := t.TempDir()
	targetsFile := filepath.Join(tmpDir, "targets.txt")
	content := "target1\ntarget2\ntarget3\ntarget4\ntarget5\ntarget6\n"
	require.NoError(t, os.WriteFile(targetsFile, []byte(content), 0644))

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-echo",
		"-T", targetsFile, "--chunk-size", "2", "--chunk-part", "1",
		"--dry-run", "-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "chunk 2/3")
	assert.Contains(t, stdout, "2 targets")
	log.Success("Specific chunk executed correctly")
}

func TestE2E_ChunkMode_InvalidChunkPart(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing invalid chunk-part error")

	tmpDir := t.TempDir()
	targetsFile := filepath.Join(tmpDir, "targets.txt")
	content := "t1\nt2\nt3\n"
	require.NoError(t, os.WriteFile(targetsFile, []byte(content), 0644))

	workflowPath := getTestdataPath(t)
	_, stderr, err := runCLIWithLog(t, log, "run", "-m", "test-echo",
		"-T", targetsFile, "--chunk-size", "3", "--chunk-part", "10",
		"-F", workflowPath)

	assert.Error(t, err)
	// Error message should be in stdout or stderr
	combined := stderr
	assert.Contains(t, combined, "exceeds total chunks")
	log.Success("Invalid chunk-part error handled correctly")
}

func TestE2E_ChunkMode_WithChunkThreads(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing chunk-threads override")

	tmpDir := t.TempDir()
	targetsFile := filepath.Join(tmpDir, "targets.txt")
	var targets []string
	for i := 0; i < 10; i++ {
		targets = append(targets, "target"+string(rune('0'+i)))
	}
	require.NoError(t, os.WriteFile(targetsFile,
		[]byte(strings.Join(targets, "\n")), 0644))

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-echo",
		"-T", targetsFile, "--chunk-size", "5", "--chunk-part", "0",
		"--chunk-threads", "3", "--dry-run", "-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "5 targets")
	log.Success("Chunk-threads override works correctly")
}

func TestE2E_ChunkMode_FirstChunk(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing first chunk execution (chunk-part 0)")

	tmpDir := t.TempDir()
	targetsFile := filepath.Join(tmpDir, "targets.txt")
	content := "alpha\nbeta\ngamma\ndelta\nepsilon\nzeta\n"
	require.NoError(t, os.WriteFile(targetsFile, []byte(content), 0644))

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-echo",
		"-T", targetsFile, "--chunk-size", "2", "--chunk-part", "0",
		"--dry-run", "-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "chunk 1/3")
	assert.Contains(t, stdout, "indices 0-1")
	log.Success("First chunk executed correctly")
}

func TestE2E_ChunkMode_LastChunk(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing last chunk execution")

	tmpDir := t.TempDir()
	targetsFile := filepath.Join(tmpDir, "targets.txt")
	content := "a\nb\nc\nd\ne\nf\ng\n" // 7 targets, chunk size 3 = 3 chunks (3,3,1)
	require.NoError(t, os.WriteFile(targetsFile, []byte(content), 0644))

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-echo",
		"-T", targetsFile, "--chunk-size", "3", "--chunk-part", "2",
		"--dry-run", "-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "chunk 3/3")
	assert.Contains(t, stdout, "1 targets") // Last chunk has 1 target
	log.Success("Last chunk executed correctly")
}

func TestE2E_ChunkMode_ChunkSizeLargerThanTargets(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing chunk size larger than target count")

	tmpDir := t.TempDir()
	targetsFile := filepath.Join(tmpDir, "targets.txt")
	content := "a\nb\nc\n" // 3 targets
	require.NoError(t, os.WriteFile(targetsFile, []byte(content), 0644))

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-echo",
		"-T", targetsFile, "--chunk-size", "100", "--chunk-part", "0",
		"--dry-run", "-F", workflowPath)

	require.NoError(t, err)
	assert.Contains(t, stdout, "chunk 1/1")
	assert.Contains(t, stdout, "3 targets")
	log.Success("Chunk size larger than targets handled correctly")
}
