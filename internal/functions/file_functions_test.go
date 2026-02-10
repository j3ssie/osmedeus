package functions

import (
	"os"
	"path/filepath"
	"strings"
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

func TestChunkFile(t *testing.T) {
	runtime := NewOttoRuntime()

	t.Run("splits file into equal chunks", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "urls.txt")
		output := filepath.Join(tmpDir, "chunks.txt")

		// Write 6 lines
		content := "line1\nline2\nline3\nline4\nline5\nline6\n"
		require.NoError(t, os.WriteFile(input, []byte(content), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 2, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		// Verify manifest
		manifest, err := os.ReadFile(output)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
		assert.Len(t, lines, 3, "should have 3 chunks for 6 lines with chunk_size=2")

		// Verify chunk 0
		chunk0, err := os.ReadFile(lines[0])
		require.NoError(t, err)
		assert.Equal(t, "line1\nline2\n", string(chunk0))

		// Verify chunk 1
		chunk1, err := os.ReadFile(lines[1])
		require.NoError(t, err)
		assert.Equal(t, "line3\nline4\n", string(chunk1))

		// Verify chunk 2
		chunk2, err := os.ReadFile(lines[2])
		require.NoError(t, err)
		assert.Equal(t, "line5\nline6\n", string(chunk2))
	})

	t.Run("handles uneven split (last chunk smaller)", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "data.txt")
		output := filepath.Join(tmpDir, "chunks.txt")

		// Write 5 lines
		content := "a\nb\nc\nd\ne\n"
		require.NoError(t, os.WriteFile(input, []byte(content), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 3, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		manifest, err := os.ReadFile(output)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
		assert.Len(t, lines, 2, "should have 2 chunks for 5 lines with chunk_size=3")

		// First chunk has 3 lines
		chunk0, err := os.ReadFile(lines[0])
		require.NoError(t, err)
		assert.Equal(t, "a\nb\nc\n", string(chunk0))

		// Second chunk has 2 lines
		chunk1, err := os.ReadFile(lines[1])
		require.NoError(t, err)
		assert.Equal(t, "d\ne\n", string(chunk1))
	})

	t.Run("skips blank lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "mixed.txt")
		output := filepath.Join(tmpDir, "chunks.txt")

		content := "line1\n\nline2\n   \nline3\n\n"
		require.NoError(t, os.WriteFile(input, []byte(content), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 2, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		manifest, err := os.ReadFile(output)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
		assert.Len(t, lines, 2, "should have 2 chunks for 3 non-blank lines with chunk_size=2")

		chunk0, err := os.ReadFile(lines[0])
		require.NoError(t, err)
		assert.Equal(t, "line1\nline2\n", string(chunk0))

		chunk1, err := os.ReadFile(lines[1])
		require.NoError(t, err)
		assert.Equal(t, "line3\n", string(chunk1))
	})

	t.Run("preserves extension in chunk names", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "urls.txt")
		output := filepath.Join(tmpDir, "chunks.txt")

		content := "url1\nurl2\nurl3\n"
		require.NoError(t, os.WriteFile(input, []byte(content), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 2, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		manifest, err := os.ReadFile(output)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")

		assert.Contains(t, lines[0], "urls_part_0.txt")
		assert.Contains(t, lines[1], "urls_part_1.txt")
	})

	t.Run("empty file produces empty manifest", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "empty.txt")
		output := filepath.Join(tmpDir, "chunks.txt")

		require.NoError(t, os.WriteFile(input, []byte(""), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 10, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		manifest, err := os.ReadFile(output)
		require.NoError(t, err)
		assert.Equal(t, "", string(manifest))
	})

	t.Run("file with only blank lines produces empty manifest", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "blanks.txt")
		output := filepath.Join(tmpDir, "chunks.txt")

		require.NoError(t, os.WriteFile(input, []byte("\n\n   \n\n"), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 5, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		manifest, err := os.ReadFile(output)
		require.NoError(t, err)
		assert.Equal(t, "", string(manifest))
	})

	t.Run("chunk size larger than file creates single chunk", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "small.txt")
		output := filepath.Join(tmpDir, "chunks.txt")

		content := "a\nb\nc\n"
		require.NoError(t, os.WriteFile(input, []byte(content), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 100, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, true, result)

		manifest, err := os.ReadFile(output)
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSpace(string(manifest)), "\n")
		assert.Len(t, lines, 1, "should have 1 chunk when chunk_size > total lines")

		chunk, err := os.ReadFile(lines[0])
		require.NoError(t, err)
		assert.Equal(t, "a\nb\nc\n", string(chunk))
	})

	t.Run("empty input path returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		output := filepath.Join(tmpDir, "chunks.txt")

		result, err := runtime.Execute(`chunk_file("", 10, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("empty output path returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "input.txt")
		require.NoError(t, os.WriteFile(input, []byte("line\n"), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 10, "")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("chunk size zero returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "input.txt")
		output := filepath.Join(tmpDir, "chunks.txt")
		require.NoError(t, os.WriteFile(input, []byte("line\n"), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", 0, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("negative chunk size returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		input := filepath.Join(tmpDir, "input.txt")
		output := filepath.Join(tmpDir, "chunks.txt")
		require.NoError(t, os.WriteFile(input, []byte("line\n"), 0644))

		result, err := runtime.Execute(`chunk_file("`+input+`", -5, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})

	t.Run("non-existent input file returns false", func(t *testing.T) {
		tmpDir := t.TempDir()
		output := filepath.Join(tmpDir, "chunks.txt")

		result, err := runtime.Execute(`chunk_file("/nonexistent/input.txt", 10, "`+output+`")`, nil)
		require.NoError(t, err)
		assert.Equal(t, false, result)
	})
}
