package functions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilepathInstaller_Binary(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	destDir := filepath.Join(tmpDir, "dest")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("Failed to create dest dir: %v", err)
	}

	// Create a fake binary file
	srcBinary := filepath.Join(srcDir, "mytool")
	if err := os.WriteFile(srcBinary, []byte("#!/bin/sh\necho hello"), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Test the function
	runtime := NewGojaRuntime()
	ctx := map[string]interface{}{
		"Binaries": destDir,
	}

	// Test installing a binary file (no archive extension)
	result, err := runtime.Execute(
		"filepath_installer('"+srcBinary+"', 'mytool', '"+destDir+"')",
		ctx,
	)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != true {
		t.Errorf("Expected true, got %v", result)
	}

	// Verify the file was copied
	destBinary := filepath.Join(destDir, "mytool")
	info, err := os.Stat(destBinary)
	if err != nil {
		t.Fatalf("Destination binary not found: %v", err)
	}

	// Check executable permissions
	if info.Mode().Perm()&0100 == 0 {
		t.Error("Destination binary should be executable")
	}
}

func TestFilepathInstaller_MissingFile(t *testing.T) {
	runtime := NewGojaRuntime()
	ctx := map[string]interface{}{}

	// Test with non-existent file
	result, err := runtime.Execute(
		"filepath_installer('/nonexistent/path/to/file', 'mytool', '/tmp/test-dest')",
		ctx,
	)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != false {
		t.Errorf("Expected false for missing file, got %v", result)
	}
}

func TestFilepathInstaller_MissingArgs(t *testing.T) {
	runtime := NewGojaRuntime()
	ctx := map[string]interface{}{}

	// Test with missing arguments
	result, err := runtime.Execute(
		"filepath_installer('/some/path')",
		ctx,
	)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != false {
		t.Errorf("Expected false for missing arguments, got %v", result)
	}
}

func TestFindBinaryInDir(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create binary in subdir
	binaryPath := filepath.Join(subDir, "mytool")
	if err := os.WriteFile(binaryPath, []byte("content"), 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Test finding binary
	found := findBinaryInDir(tmpDir, "mytool")
	if found != binaryPath {
		t.Errorf("Expected %s, got %s", binaryPath, found)
	}

	// Test not finding binary
	notFound := findBinaryInDir(tmpDir, "nonexistent")
	if notFound != "" {
		t.Errorf("Expected empty string for nonexistent, got %s", notFound)
	}
}

func TestCopyBinaryFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tmpDir, "src")
	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	dstPath := filepath.Join(tmpDir, "dst")
	if err := copyBinaryFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyBinaryFile failed: %v", err)
	}

	// Verify content
	copied, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}
	if string(copied) != string(content) {
		t.Errorf("Content mismatch: expected %s, got %s", content, copied)
	}
}
