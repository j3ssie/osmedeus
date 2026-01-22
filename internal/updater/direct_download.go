package updater

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/installer"
)

// DirectDownloader provides fallback download functionality when go-selfupdate fails
type DirectDownloader struct {
	owner   string
	repo    string
	verbose bool
}

// NewDirectDownloader creates a new direct downloader
func NewDirectDownloader(owner, repo string, verbose bool) *DirectDownloader {
	return &DirectDownloader{
		owner:   owner,
		repo:    repo,
		verbose: verbose,
	}
}

// BuildAssetURL constructs the GitHub release asset URL based on goreleaser naming convention
// Format: osmedeus_{version}_{os}_{arch}.tar.gz
func (d *DirectDownloader) BuildAssetURL(version string) string {
	// Strip 'v' prefix for the filename
	ver := strings.TrimPrefix(version, "v")

	// Map GOOS to goreleaser output names
	osName := runtime.GOOS
	switch osName {
	case "darwin":
		osName = "darwin"
	case "linux":
		osName = "linux"
	case "windows":
		osName = "windows"
	}

	// Map GOARCH to goreleaser output names
	arch := runtime.GOARCH
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "386"
	}

	// Determine extension based on OS
	ext := "tar.gz"
	if osName == "windows" {
		ext = "zip"
	}

	// Build the asset filename
	assetName := fmt.Sprintf("osmedeus_%s_%s_%s.%s", ver, osName, arch, ext)

	// Build the full URL
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s",
		d.owner, d.repo, ver, assetName)
}

// UpdateBinary downloads and replaces the current binary using direct download
func (d *DirectDownloader) UpdateBinary(ctx context.Context, release *Release) error {
	// Get the executable path
	exe, err := GetExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create a temp directory for the download
	tempDir, err := os.MkdirTemp("", "osmedeus-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Build the asset URL
	assetURL := d.BuildAssetURL(release.Version)
	if d.verbose {
		fmt.Printf("[debug] Direct download URL: %s\n", assetURL)
	}

	// Determine archive extension
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	// Download the archive
	archivePath := filepath.Join(tempDir, fmt.Sprintf("osmedeus.%s", ext))
	if d.verbose {
		fmt.Printf("[debug] Downloading to: %s\n", archivePath)
	}

	if err := installer.DownloadFile(assetURL, archivePath, nil); err != nil {
		return fmt.Errorf("failed to download release: %w", err)
	}

	// Extract the archive
	extractDir := filepath.Join(tempDir, "extracted")
	if ext == "zip" {
		if err := installer.ExtractZip(archivePath, extractDir); err != nil {
			return fmt.Errorf("failed to extract zip: %w", err)
		}
	} else {
		if err := installer.ExtractTarGz(archivePath, extractDir); err != nil {
			return fmt.Errorf("failed to extract tar.gz: %w", err)
		}
	}

	// Find the binary in the extracted directory
	binaryName := "osmedeus"
	if runtime.GOOS == "windows" {
		binaryName = "osmedeus.exe"
	}

	newBinary := filepath.Join(extractDir, binaryName)
	if _, err := os.Stat(newBinary); os.IsNotExist(err) {
		return fmt.Errorf("binary not found in archive: %s", binaryName)
	}

	if d.verbose {
		fmt.Printf("[debug] Found binary: %s\n", newBinary)
		fmt.Printf("[debug] Replacing: %s\n", exe)
	}

	// Perform atomic replacement
	if err := atomicReplace(newBinary, exe); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// atomicReplace replaces the target file with the source file atomically
func atomicReplace(src, dst string) error {
	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source binary: %w", err)
	}

	// Get the permissions of the original file
	info, err := os.Stat(dst)
	if err != nil {
		return fmt.Errorf("failed to stat destination: %w", err)
	}
	mode := info.Mode()

	// Create a temp file in the same directory as destination (for atomic rename)
	dstDir := filepath.Dir(dst)
	tmpFile, err := os.CreateTemp(dstDir, ".osmedeus-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file if something goes wrong
	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	// Write the new binary
	if _, err = tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set the same permissions as the original
	if err = os.Chmod(tmpPath, mode); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename (works on Unix, on Windows we need a different approach)
	if runtime.GOOS == "windows" {
		// On Windows, we can't rename over an existing file, so we need to:
		// 1. Rename the old file
		// 2. Rename the new file to the target
		// 3. Delete the old file
		oldPath := dst + ".old"
		os.Remove(oldPath) // Remove any leftover from previous update
		if err = os.Rename(dst, oldPath); err != nil {
			return fmt.Errorf("failed to rename old binary: %w", err)
		}
		if err = os.Rename(tmpPath, dst); err != nil {
			// Try to restore the old binary
			os.Rename(oldPath, dst)
			return fmt.Errorf("failed to rename new binary: %w", err)
		}
		os.Remove(oldPath) // Clean up old binary
	} else {
		if err = os.Rename(tmpPath, dst); err != nil {
			return fmt.Errorf("failed to rename: %w", err)
		}
	}

	return nil
}
