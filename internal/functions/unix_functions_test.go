package functions

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func TestGitCloneSubfolder_EmptyGitURL(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`git_clone_subfolder("", "subfolder", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestGitCloneSubfolder_EmptySubfolder(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`git_clone_subfolder("https://github.com/user/repo", "", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestGitCloneSubfolder_EmptyDest(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`git_clone_subfolder("https://github.com/user/repo", "subfolder", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestGitCloneSubfolder_UndefinedArgs(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`git_clone_subfolder()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestIsGitHubURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://github.com/user/repo", true},
		{"http://github.com/user/repo", true},
		{"git@github.com:user/repo.git", true},
		{"https://gitlab.com/user/repo", false},
		{"https://bitbucket.org/user/repo", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			assert.Equal(t, tt.expected, isGitHubURL(tt.url))
		})
	}
}

func TestConvertToGitHubZipURL(t *testing.T) {
	tests := []struct {
		gitURL   string
		expected string
	}{
		{
			"https://github.com/projectdiscovery/nuclei-templates",
			"https://github.com/projectdiscovery/nuclei-templates/archive/refs/heads/main.zip",
		},
		{
			"https://github.com/projectdiscovery/nuclei-templates.git",
			"https://github.com/projectdiscovery/nuclei-templates/archive/refs/heads/main.zip",
		},
		{
			"git@github.com:projectdiscovery/nuclei-templates.git",
			"https://github.com/projectdiscovery/nuclei-templates/archive/refs/heads/main.zip",
		},
		{
			"http://github.com/user/repo",
			"https://github.com/user/repo/archive/refs/heads/main.zip",
		},
		{
			"https://gitlab.com/user/repo",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.gitURL, func(t *testing.T) {
			assert.Equal(t, tt.expected, convertToGitHubZipURL(tt.gitURL))
		})
	}
}

func TestFindSubfolder(t *testing.T) {
	tmpDir := t.TempDir()

	// Create direct subfolder
	directSub := filepath.Join(tmpDir, "direct-sub")
	require.NoError(t, os.MkdirAll(directSub, 0755))

	// Create nested subfolder (simulating ZIP extract)
	nestedRoot := filepath.Join(tmpDir, "repo-main")
	nestedSub := filepath.Join(nestedRoot, "nested-sub")
	require.NoError(t, os.MkdirAll(nestedSub, 0755))

	// Test direct subfolder
	result := findSubfolder(tmpDir, "direct-sub")
	assert.Equal(t, directSub, result)

	// Test nested subfolder
	result = findSubfolder(tmpDir, "nested-sub")
	assert.Equal(t, nestedSub, result)

	// Test non-existent subfolder
	result = findSubfolder(tmpDir, "nonexistent")
	assert.Equal(t, "", result)
}

func TestCopyDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "copy-dest")

	// Create source structure
	subDir := filepath.Join(srcDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	// Create files
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "file2.txt"), []byte("content2"), 0644))

	// Copy
	err := copyDir(srcDir, dstDir)
	require.NoError(t, err)

	// Verify
	content1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content1", string(content1))

	content2, err := os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt"))
	require.NoError(t, err)
	assert.Equal(t, "content2", string(content2))
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")

	// Create source file
	content := "test content"
	require.NoError(t, os.WriteFile(srcFile, []byte(content), 0644))

	// Copy
	err := copyFile(srcFile, dstFile)
	require.NoError(t, err)

	// Verify
	result, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(result))
}

// --- wget (pure Go) tests ---

func TestWget_EmptyArgument(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`wget("")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestWget_EmptyOutputPath(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`wget("https://example.com/file.txt", "")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestWget_RemovesExistingFile(t *testing.T) {
	// Serve a small file
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte("new-content"))
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "existing.txt")

	// Write an existing file with old content
	require.NoError(t, os.WriteFile(outputPath, []byte("old-content"), 0644))

	registry := NewRegistry()
	result, err := registry.Execute(
		fmt.Sprintf(`wget('%s', '%s')`, ts.URL+"/file.txt", outputPath),
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify new content
	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, "new-content", string(data))
}

func TestWget_SmallFile(t *testing.T) {
	payload := "hello-world-test-data"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte(payload))
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "small.txt")

	registry := NewRegistry()
	result, err := registry.Execute(
		fmt.Sprintf(`wget('%s', '%s')`, ts.URL+"/small.txt", outputPath),
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, payload, string(data))
}

func TestWget_SegmentedDownload(t *testing.T) {
	// Create a payload >1MB to trigger segmented download
	payload := bytes.Repeat([]byte("A"), 2*1024*1024) // 2MB

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			w.Header().Set("Accept-Ranges", "bytes")
			w.WriteHeader(http.StatusOK)
			return
		}
		// Handle Range requests
		rangeHeader := r.Header.Get("Range")
		if rangeHeader != "" && strings.HasPrefix(rangeHeader, "bytes=") {
			rangeParts := strings.TrimPrefix(rangeHeader, "bytes=")
			parts := strings.Split(rangeParts, "-")
			if len(parts) == 2 {
				start, _ := strconv.ParseInt(parts[0], 10, 64)
				end, _ := strconv.ParseInt(parts[1], 10, 64)
				if end >= int64(len(payload)) {
					end = int64(len(payload)) - 1
				}
				w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(payload)))
				w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
				w.WriteHeader(http.StatusPartialContent)
				_, _ = w.Write(payload[start : end+1])
				return
			}
		}
		// Full response
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "large.bin")

	registry := NewRegistry()
	result, err := registry.Execute(
		fmt.Sprintf(`wget('%s', '%s')`, ts.URL+"/large.bin", outputPath),
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, len(payload), len(data))
	assert.True(t, bytes.Equal(payload, data))
}

func TestWget_FallbackNoRange(t *testing.T) {
	payload := bytes.Repeat([]byte("B"), 2*1024*1024) // 2MB, no Range support

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// HEAD response without Accept-Ranges
		if r.Method == "HEAD" {
			w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "norange.bin")

	registry := NewRegistry()
	result, err := registry.Execute(
		fmt.Sprintf(`wget('%s', '%s')`, ts.URL+"/norange.bin", outputPath),
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, len(payload), len(data))
	assert.True(t, bytes.Equal(payload, data))
}

func TestWget_CreatesParentDirs(t *testing.T) {
	payload := "test-parent-dir"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(payload))
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "a", "b", "c", "file.txt")

	registry := NewRegistry()
	result, err := registry.Execute(
		fmt.Sprintf(`wget('%s', '%s')`, ts.URL+"/file.txt", outputPath),
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	data, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Equal(t, payload, string(data))
}
