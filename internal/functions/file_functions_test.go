package functions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveBlankLines(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("removes blank lines from file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		content := "line1\n\nline2\n   \nline3\n\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`remove_blank_lines("`+testFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Read back the file and verify
		data, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, "line1\nline2\nline3\n", string(data))
	})

	t.Run("handles file with only blank lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "blanks.txt")
		content := "\n\n   \n\t\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`remove_blank_lines("`+testFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// File should be empty (no trailing newline since no lines)
		data, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, "", string(data))
	})

	t.Run("handles file with no blank lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "noblanks.txt")
		content := "line1\nline2\nline3\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`remove_blank_lines("`+testFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// File should be unchanged
		data, err := os.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, "line1\nline2\nline3\n", string(data))
	})

	t.Run("empty path returns false", func(t *testing.T) {
		result, err := runtime.Execute(`remove_blank_lines("")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("non-existent file returns false", func(t *testing.T) {
		result, err := runtime.Execute(`remove_blank_lines("/nonexistent/file.txt")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("directory path returns false", func(t *testing.T) {
		tmpDir := t.TempDir()

		result, err := runtime.Execute(`remove_blank_lines("`+tmpDir+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("preserves whitespace-only lines at start/middle", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "whitespace.txt")
		// Lines with content should be preserved, blank/whitespace-only removed
		content := "  line with leading spaces\n\n\tmiddle with tab\n   \nlast line\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err)

		result, err := runtime.Execute(`remove_blank_lines("`+testFile+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		data, err := os.ReadFile(testFile)
		require.NoError(t, err)
		// Lines with actual content (even with leading whitespace) should be kept
		assert.Equal(t, "  line with leading spaces\n\tmiddle with tab\nlast line\n", string(data))
	})
}
