package snapshot

import (
	"archive/zip"
	"compress/flate"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// ExportResult contains the result of a workspace export
type ExportResult struct {
	WorkspaceName string
	SourcePath    string
	OutputPath    string
	FileSize      int64
}

// ImportResult contains the result of a workspace import
type ImportResult struct {
	WorkspaceName string
	LocalPath     string
	DataSource    string
	FilesCount    int
}

// ExportWorkspace creates a compressed zip of the workspace folder
func ExportWorkspace(workspacePath, outputPath string) (*ExportResult, error) {
	// Validate source exists
	info, err := os.Stat(workspacePath)
	if err != nil {
		return nil, fmt.Errorf("workspace not found: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("workspace path is not a directory: %s", workspacePath)
	}

	// Extract workspace name from path
	workspaceName := filepath.Base(workspacePath)

	// Generate output path if not specified
	if outputPath == "" {
		outputPath = fmt.Sprintf("%s_%d.zip", workspaceName, time.Now().Unix())
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the zip archive with highest compression
	if err := createHighCompressionZip(workspacePath, outputPath); err != nil {
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}

	// Get file size
	zipInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get archive info: %w", err)
	}

	return &ExportResult{
		WorkspaceName: workspaceName,
		SourcePath:    workspacePath,
		OutputPath:    outputPath,
		FileSize:      zipInfo.Size(),
	}, nil
}

// ImportWorkspace extracts a zip file and optionally imports to database
func ImportWorkspace(source, workspacesPath string, skipDB bool, cfg *config.Config) (*ImportResult, error) {
	var zipPath string
	var cleanup func()

	// Check if source is URL or local file
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Download the file
		tempPath, err := downloadFile(source)
		if err != nil {
			return nil, fmt.Errorf("failed to download: %w", err)
		}
		zipPath = tempPath
		cleanup = func() { _ = os.Remove(tempPath) }
	} else {
		// Use local file
		if _, err := os.Stat(source); err != nil {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		zipPath = source
		cleanup = func() {} // No cleanup needed for local files
	}
	defer cleanup()

	// Extract workspace name from zip filename
	zipName := filepath.Base(zipPath)
	workspaceName := strings.TrimSuffix(zipName, ".zip")
	// Remove timestamp suffix if present (e.g., "example.com_1234567890" -> "example.com")
	if idx := strings.LastIndex(workspaceName, "_"); idx > 0 {
		potentialTimestamp := workspaceName[idx+1:]
		// Check if it looks like a Unix timestamp (all digits, ~10 chars)
		if len(potentialTimestamp) >= 10 && isNumeric(potentialTimestamp) {
			workspaceName = workspaceName[:idx]
		}
	}

	// Destination path
	destPath := filepath.Join(workspacesPath, workspaceName)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return nil, fmt.Errorf("workspace already exists: %s (use --force to overwrite)", destPath)
	}

	// Extract the zip
	filesCount, err := extractZip(zipPath, destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	result := &ImportResult{
		WorkspaceName: workspaceName,
		LocalPath:     destPath,
		DataSource:    "imported",
		FilesCount:    filesCount,
	}

	// Import to database if not skipped
	if !skipDB && cfg != nil {
		if err := ImportWorkspaceToDB(context.Background(), workspaceName, destPath, cfg); err != nil {
			// Log warning but don't fail the import
			fmt.Printf("Warning: Failed to import to database: %v\n", err)
		}
	}

	return result, nil
}

// ImportWorkspaceToDB creates or updates a workspace record with data_source="imported"
func ImportWorkspaceToDB(ctx context.Context, workspaceName, localPath string, cfg *config.Config) error {
	// Connect to database
	_, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ensure tables exist
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Create or update workspace record
	workspace := &database.Workspace{
		Name:       workspaceName,
		LocalPath:  localPath,
		DataSource: "imported",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return database.UpsertWorkspace(ctx, workspace)
}

// ForceImportWorkspace imports a workspace, overwriting if it exists
func ForceImportWorkspace(source, workspacesPath string, skipDB bool, cfg *config.Config) (*ImportResult, error) {
	var zipPath string
	var cleanup func()

	// Check if source is URL or local file
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		tempPath, err := downloadFile(source)
		if err != nil {
			return nil, fmt.Errorf("failed to download: %w", err)
		}
		zipPath = tempPath
		cleanup = func() { _ = os.Remove(tempPath) }
	} else {
		if _, err := os.Stat(source); err != nil {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		zipPath = source
		cleanup = func() {}
	}
	defer cleanup()

	// Extract workspace name
	zipName := filepath.Base(zipPath)
	workspaceName := strings.TrimSuffix(zipName, ".zip")
	if idx := strings.LastIndex(workspaceName, "_"); idx > 0 {
		potentialTimestamp := workspaceName[idx+1:]
		if len(potentialTimestamp) >= 10 && isNumeric(potentialTimestamp) {
			workspaceName = workspaceName[:idx]
		}
	}

	destPath := filepath.Join(workspacesPath, workspaceName)

	// Remove existing if present
	if _, err := os.Stat(destPath); err == nil {
		if err := os.RemoveAll(destPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing workspace: %w", err)
		}
	}

	// Extract the zip
	filesCount, err := extractZip(zipPath, destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	result := &ImportResult{
		WorkspaceName: workspaceName,
		LocalPath:     destPath,
		DataSource:    "imported",
		FilesCount:    filesCount,
	}

	// Import to database if not skipped
	if !skipDB && cfg != nil {
		if err := ImportWorkspaceToDB(context.Background(), workspaceName, destPath, cfg); err != nil {
			fmt.Printf("Warning: Failed to import to database: %v\n", err)
		}
	}

	return result, nil
}

// createHighCompressionZip creates a zip archive with highest compression level
func createHighCompressionZip(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() { _ = zipFile.Close() }()

	archive := zip.NewWriter(zipFile)
	defer func() { _ = archive.Close() }()

	// Register highest compression level
	archive.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	baseDir := filepath.Base(source)

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Include base directory name in archive path
		archivePath := filepath.Join(baseDir, relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = archivePath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(writer, file)
		return err
	})
}

// extractZip extracts a zip archive to destination, returns count of files extracted
func extractZip(src, dest string) (int, error) {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return 0, err
	}
	defer func() { _ = reader.Close() }()

	filesCount := 0

	for _, file := range reader.File {
		// Prevent zip slip attack
		destPath := filepath.Join(dest, file.Name)
		if !strings.HasPrefix(destPath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filesCount, fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return filesCount, err
			}
			continue
		}

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return filesCount, err
		}

		// Extract file
		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return filesCount, err
		}

		rc, err := file.Open()
		if err != nil {
			_ = outFile.Close()
			return filesCount, err
		}

		_, err = io.Copy(outFile, rc)
		_ = rc.Close()
		_ = outFile.Close()

		if err != nil {
			return filesCount, err
		}

		filesCount++
	}

	return filesCount, nil
}

// downloadFile downloads a file from URL to a temp file
func downloadFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "snapshot-*.zip")
	if err != nil {
		return "", err
	}
	defer func() { _ = tempFile.Close() }()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		_ = os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// ListSnapshots returns a list of snapshot files in the snapshot directory
func ListSnapshots(snapshotPath string) ([]SnapshotInfo, error) {
	if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
		return []SnapshotInfo{}, nil
	}

	entries, err := os.ReadDir(snapshotPath)
	if err != nil {
		return nil, err
	}

	var snapshots []SnapshotInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zip") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		snapshots = append(snapshots, SnapshotInfo{
			Name:      entry.Name(),
			Path:      filepath.Join(snapshotPath, entry.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	return snapshots, nil
}

// SnapshotInfo contains information about a snapshot file
type SnapshotInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}
