package functions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortUnix_InPlace(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "unsorted.txt")

	// Create unsorted file with duplicates
	content := "zebra\napple\nbanana\napple\ncherry\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`sort_unix("`+testFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify file is sorted and unique
	sorted, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, "apple\nbanana\ncherry\nzebra\n", string(sorted))
}

func TestSortUnix_ToOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "input.txt")
	outputFile := filepath.Join(tmpDir, "output.txt")

	// Create unsorted file
	content := "3\n1\n2\n1\n"
	err := os.WriteFile(inputFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`sort_unix("`+inputFile+`", "`+outputFile+`")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify output file is sorted and unique
	sorted, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, "1\n2\n3\n", string(sorted))

	// Verify input file is unchanged
	original, err := os.ReadFile(inputFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(original))
}

func TestSortUnix_EmptyArgument(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sort_unix("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSortUnix_NonExistentFile(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sort_unix("/nonexistent/file.txt")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestGitClone_EmptyArgument(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`git_clone("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestWgetUnix_EmptyArgument(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`wget_unix("")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}
