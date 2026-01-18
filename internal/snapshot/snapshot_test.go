package snapshot

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsNumeric(t *testing.T) {
	t.Run("valid numeric string", func(t *testing.T) {
		assert.True(t, isNumeric("1234567890"))
	})

	t.Run("empty string returns false", func(t *testing.T) {
		assert.False(t, isNumeric(""))
	})

	t.Run("string with letters returns false", func(t *testing.T) {
		assert.False(t, isNumeric("123abc"))
	})

	t.Run("string with spaces returns false", func(t *testing.T) {
		assert.False(t, isNumeric("123 456"))
	})

	t.Run("timestamp-like string is numeric", func(t *testing.T) {
		assert.True(t, isNumeric("1704067200"))
	})
}

func TestCreateHighCompressionZip(t *testing.T) {
	t.Run("creates valid zip archive", func(t *testing.T) {
		// Create temp source directory
		sourceDir, err := os.MkdirTemp("", "snapshot-test-source-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(sourceDir) }()

		// Create test files
		err = os.WriteFile(filepath.Join(sourceDir, "test.txt"), []byte("test content"), 0644)
		require.NoError(t, err)

		subDir := filepath.Join(sourceDir, "subdir")
		err = os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("nested content"), 0644)
		require.NoError(t, err)

		// Create zip
		targetFile := filepath.Join(os.TempDir(), "test-snapshot.zip")
		defer func() { _ = os.Remove(targetFile) }()

		err = createHighCompressionZip(sourceDir, targetFile)
		require.NoError(t, err)

		// Verify zip exists and is valid
		reader, err := zip.OpenReader(targetFile)
		require.NoError(t, err)
		defer func() { _ = reader.Close() }()

		assert.GreaterOrEqual(t, len(reader.File), 2, "zip should contain at least 2 files")
	})

	t.Run("fails with non-existent source", func(t *testing.T) {
		err := createHighCompressionZip("/nonexistent/path", "/tmp/test.zip")
		assert.Error(t, err)
	})
}

func TestExtractZip(t *testing.T) {
	t.Run("extracts files correctly", func(t *testing.T) {
		// Create a test zip file
		zipPath := filepath.Join(os.TempDir(), "test-extract.zip")
		defer func() { _ = os.Remove(zipPath) }()

		// Create zip with test content
		zipFile, err := os.Create(zipPath)
		require.NoError(t, err)

		writer := zip.NewWriter(zipFile)
		f, err := writer.Create("test.txt")
		require.NoError(t, err)
		_, err = f.Write([]byte("test content"))
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)
		err = zipFile.Close()
		require.NoError(t, err)

		// Extract to temp directory
		destDir, err := os.MkdirTemp("", "snapshot-test-extract-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(destDir) }()

		filesCount, err := extractZip(zipPath, destDir)
		require.NoError(t, err)
		assert.Equal(t, 1, filesCount)

		// Verify file exists
		content, err := os.ReadFile(filepath.Join(destDir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))
	})

	t.Run("prevents zip slip attack", func(t *testing.T) {
		// Create a malicious zip with path traversal
		zipPath := filepath.Join(os.TempDir(), "test-zipslip.zip")
		defer func() { _ = os.Remove(zipPath) }()

		zipFile, err := os.Create(zipPath)
		require.NoError(t, err)

		writer := zip.NewWriter(zipFile)
		// Try to create a file with path traversal
		f, err := writer.Create("../../../etc/malicious.txt")
		require.NoError(t, err)
		_, err = f.Write([]byte("malicious content"))
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)
		err = zipFile.Close()
		require.NoError(t, err)

		// Attempt extraction should fail
		destDir, err := os.MkdirTemp("", "snapshot-test-zipslip-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(destDir) }()

		_, err = extractZip(zipPath, destDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "illegal file path")
	})

	t.Run("fails with non-existent zip", func(t *testing.T) {
		destDir, err := os.MkdirTemp("", "snapshot-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(destDir) }()

		_, err = extractZip("/nonexistent/file.zip", destDir)
		assert.Error(t, err)
	})
}

func TestExportWorkspace(t *testing.T) {
	t.Run("exports workspace successfully", func(t *testing.T) {
		// Create temp workspace directory
		workspaceDir, err := os.MkdirTemp("", "workspace-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workspaceDir) }()

		// Create test file
		err = os.WriteFile(filepath.Join(workspaceDir, "output.txt"), []byte("scan results"), 0644)
		require.NoError(t, err)

		// Export
		outputPath := filepath.Join(os.TempDir(), "workspace-export.zip")
		defer func() { _ = os.Remove(outputPath) }()

		result, err := ExportWorkspace(workspaceDir, outputPath)
		require.NoError(t, err)

		assert.Equal(t, filepath.Base(workspaceDir), result.WorkspaceName)
		assert.Equal(t, workspaceDir, result.SourcePath)
		assert.Equal(t, outputPath, result.OutputPath)
		assert.Greater(t, result.FileSize, int64(0))
	})

	t.Run("fails with non-existent workspace", func(t *testing.T) {
		_, err := ExportWorkspace("/nonexistent/workspace", "/tmp/test.zip")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "workspace not found")
	})

	t.Run("fails if workspace is a file not directory", func(t *testing.T) {
		// Create a file instead of directory
		tmpFile, err := os.CreateTemp("", "not-a-dir-*")
		require.NoError(t, err)
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		require.NoError(t, tmpFile.Close())

		_, err = ExportWorkspace(tmpFile.Name(), "/tmp/test.zip")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("generates output path if empty", func(t *testing.T) {
		// Create temp workspace directory
		workspaceDir, err := os.MkdirTemp("", "workspace-test-autopath-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workspaceDir) }()

		err = os.WriteFile(filepath.Join(workspaceDir, "file.txt"), []byte("content"), 0644)
		require.NoError(t, err)

		result, err := ExportWorkspace(workspaceDir, "")
		require.NoError(t, err)
		defer func() { _ = os.Remove(result.OutputPath) }()

		assert.Contains(t, result.OutputPath, filepath.Base(workspaceDir))
		assert.Contains(t, result.OutputPath, ".zip")
	})
}

func TestListSnapshots(t *testing.T) {
	t.Run("lists zip files in directory", func(t *testing.T) {
		// Create temp snapshot directory
		snapshotDir, err := os.MkdirTemp("", "snapshot-list-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(snapshotDir) }()

		// Create test zip files
		err = os.WriteFile(filepath.Join(snapshotDir, "workspace1_1234567890.zip"), []byte("zip1"), 0644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(snapshotDir, "workspace2_1234567891.zip"), []byte("zip2"), 0644)
		require.NoError(t, err)

		// Create a non-zip file (should be ignored)
		err = os.WriteFile(filepath.Join(snapshotDir, "readme.txt"), []byte("readme"), 0644)
		require.NoError(t, err)

		// Create a directory (should be ignored)
		err = os.MkdirAll(filepath.Join(snapshotDir, "subdir"), 0755)
		require.NoError(t, err)

		snapshots, err := ListSnapshots(snapshotDir)
		require.NoError(t, err)

		assert.Len(t, snapshots, 2)
	})

	t.Run("returns empty list for non-existent directory", func(t *testing.T) {
		snapshots, err := ListSnapshots("/nonexistent/path")
		require.NoError(t, err)
		assert.Empty(t, snapshots)
	})

	t.Run("returns empty list for empty directory", func(t *testing.T) {
		snapshotDir, err := os.MkdirTemp("", "snapshot-list-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(snapshotDir) }()

		snapshots, err := ListSnapshots(snapshotDir)
		require.NoError(t, err)
		assert.Empty(t, snapshots)
	})
}

func TestExportResult(t *testing.T) {
	t.Run("struct fields are populated", func(t *testing.T) {
		result := ExportResult{
			WorkspaceName: "example.com",
			SourcePath:    "/home/user/workspaces/example.com",
			OutputPath:    "/home/user/snapshot/example.com_123.zip",
			FileSize:      1024,
		}

		assert.Equal(t, "example.com", result.WorkspaceName)
		assert.Equal(t, "/home/user/workspaces/example.com", result.SourcePath)
		assert.Equal(t, "/home/user/snapshot/example.com_123.zip", result.OutputPath)
		assert.Equal(t, int64(1024), result.FileSize)
	})
}

func TestImportResult(t *testing.T) {
	t.Run("struct fields are populated", func(t *testing.T) {
		result := ImportResult{
			WorkspaceName: "example.com",
			LocalPath:     "/home/user/workspaces/example.com",
			DataSource:    "imported",
			FilesCount:    100,
		}

		assert.Equal(t, "example.com", result.WorkspaceName)
		assert.Equal(t, "/home/user/workspaces/example.com", result.LocalPath)
		assert.Equal(t, "imported", result.DataSource)
		assert.Equal(t, 100, result.FilesCount)
	})
}

func TestSnapshotInfo(t *testing.T) {
	t.Run("struct fields are accessible", func(t *testing.T) {
		info := SnapshotInfo{
			Name: "example.com_123.zip",
			Path: "/home/user/snapshot/example.com_123.zip",
			Size: 2048,
		}

		assert.Equal(t, "example.com_123.zip", info.Name)
		assert.Equal(t, "/home/user/snapshot/example.com_123.zip", info.Path)
		assert.Equal(t, int64(2048), info.Size)
	})
}
