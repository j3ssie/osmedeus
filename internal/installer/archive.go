package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip extracts a zip file to the destination directory
func ExtractZip(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		if err := extractZipFile(f, dest); err != nil {
			return err
		}
	}

	return nil
}

func extractZipFile(f *zip.File, dest string) error {
	// Sanitize the path to prevent zip slip vulnerability
	path := filepath.Join(dest, f.Name)
	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(path, f.Mode())
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Create the file
	outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = outFile.Close() }()

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer func() { _ = rc.Close() }()

	_, err = io.Copy(outFile, rc)
	return err
}

// ExtractTarGz extracts a tar.gz file to the destination directory
func ExtractTarGz(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer func() { _ = file.Close() }()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar entry: %w", err)
		}

		if err := extractTarEntry(header, tr, dest); err != nil {
			return err
		}
	}

	return nil
}

func extractTarEntry(header *tar.Header, tr *tar.Reader, dest string) error {
	// Sanitize the path to prevent path traversal
	path := filepath.Join(dest, header.Name)
	if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", header.Name)
	}

	switch header.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
			return err
		}

	case tar.TypeReg:
		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}

		if _, err := io.Copy(outFile, tr); err != nil {
			_ = outFile.Close()
			return err
		}
		_ = outFile.Close()

	case tar.TypeSymlink:
		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.Symlink(header.Linkname, path); err != nil {
			return err
		}
	}

	return nil
}

// ExtractGz extracts a single gzip file (not tar.gz) to the destination
// The dest should be the full file path, not a directory
func ExtractGz(src, dest string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open gz file: %w", err)
	}
	defer func() { _ = file.Close() }()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gzr.Close() }()

	outFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	if _, err := io.Copy(outFile, gzr); err != nil {
		return fmt.Errorf("failed to extract gz file: %w", err)
	}

	// Make executable
	if err := os.Chmod(dest, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// DetectArchiveType returns the archive type based on file extension
func DetectArchiveType(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(lower, ".gz"):
		return "gz"
	case strings.HasSuffix(lower, ".zip"):
		return "zip"
	default:
		return "unknown"
	}
}
