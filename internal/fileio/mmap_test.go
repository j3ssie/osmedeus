package fileio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenFile_SmallFile(t *testing.T) {
	// Create a small test file (< 1MB)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "small.txt")
	content := "line1\nline2\nline3\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := OpenFile(testFile)
	require.NoError(t, err)
	defer func() { _ = mf.Close() }()

	assert.Equal(t, int64(len(content)), mf.Size())
	assert.False(t, mf.IsMapped()) // Small file should not be mmap'd
	assert.Equal(t, content, mf.String())
}

func TestOpenFile_LargeFile(t *testing.T) {
	// Create a large test file (> 1MB)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")

	// Create file with 2MB of content
	content := strings.Repeat("x", MmapThreshold+1024)
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := OpenFile(testFile)
	require.NoError(t, err)
	defer func() { _ = mf.Close() }()

	assert.Equal(t, int64(len(content)), mf.Size())
	assert.True(t, mf.IsMapped()) // Large file should be mmap'd
	assert.Equal(t, content, mf.String())
}

func TestReadLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "lines.txt")
	content := "line1\n\nline2\n  \nline3\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := OpenFile(testFile)
	require.NoError(t, err)
	defer func() { _ = mf.Close() }()

	var lines []string
	iter := mf.ReadLines()
	for iter.Next() {
		lines = append(lines, iter.Line())
	}
	require.NoError(t, iter.Err())

	// Should skip empty lines
	assert.Equal(t, []string{"line1", "line2", "line3"}, lines)
}

func TestReadLines_WindowsLineEndings(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "windows.txt")
	content := "line1\r\nline2\r\nline3\r\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	mf, err := OpenFile(testFile)
	require.NoError(t, err)
	defer func() { _ = mf.Close() }()

	var lines []string
	iter := mf.ReadLines()
	for iter.Next() {
		lines = append(lines, iter.Line())
	}
	require.NoError(t, iter.Err())

	assert.Equal(t, []string{"line1", "line2", "line3"}, lines)
}

func TestReadLinesFiltered(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "filtered.txt")
	content := "# comment\nexample.com\n\n# another comment\ntest.com\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	lines, err := ReadLinesFiltered(testFile)
	require.NoError(t, err)

	// Should skip comments and empty lines
	assert.Equal(t, []string{"example.com", "test.com"}, lines)
}

func TestCountNonEmptyLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "count.txt")
	content := "line1\n\nline2\n  \nline3\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	count, err := CountNonEmptyLines(testFile)
	require.NoError(t, err)

	assert.Equal(t, 3, count)
}

func TestOpenFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")
	err := os.WriteFile(testFile, []byte{}, 0644)
	require.NoError(t, err)

	mf, err := OpenFile(testFile)
	require.NoError(t, err)
	defer func() { _ = mf.Close() }()

	assert.Equal(t, int64(0), mf.Size())
	assert.Equal(t, "", mf.String())

	// Iterating over empty file should yield no lines
	iter := mf.ReadLines()
	assert.False(t, iter.Next())
}

func TestOpenFile_FileNotFound(t *testing.T) {
	_, err := OpenFile("/nonexistent/file.txt")
	assert.Error(t, err)
}

func BenchmarkReadLines_Mmap(b *testing.B) {
	// Create a large test file
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.txt")

	// Create file with 5MB of content (100k lines)
	var builder strings.Builder
	for i := 0; i < 100000; i++ {
		builder.WriteString("line content here that is moderately long\n")
	}
	err := os.WriteFile(testFile, []byte(builder.String()), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mf, _ := OpenFile(testFile)
		count := 0
		iter := mf.ReadLines()
		for iter.Next() {
			count++
		}
		_ = mf.Close()
	}
}

func BenchmarkReadLines_Bufio(b *testing.B) {
	// Create a large test file
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "bench.txt")

	// Create file with 5MB of content (100k lines)
	var builder strings.Builder
	for i := 0; i < 100000; i++ {
		builder.WriteString("line content here that is moderately long\n")
	}
	err := os.WriteFile(testFile, []byte(builder.String()), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f, _ := os.Open(testFile)
		scanner := strings.NewReader(builder.String())
		count := 0
		buf := make([]byte, 64*1024)
		for {
			n, err := scanner.Read(buf)
			if n == 0 || err != nil {
				break
			}
			for _, b := range buf[:n] {
				if b == '\n' {
					count++
				}
			}
		}
		_ = f.Close()
	}
}
