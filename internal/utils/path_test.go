package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookPathWithBinaries_SystemPath(t *testing.T) {
	// echo should be found in system PATH
	path, err := LookPathWithBinaries("echo", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestLookPathWithBinaries_NotFound(t *testing.T) {
	// nonexistent command should not be found
	_, err := LookPathWithBinaries("nonexistent-command-xyz-12345", "")
	assert.Error(t, err)
}

func TestLookPathWithBinaries_InBinariesPath(t *testing.T) {
	// Create a temp directory with a fake binary
	tmpDir, err := os.MkdirTemp("", "test-binaries")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a fake executable
	fakeBinary := filepath.Join(tmpDir, "fake-tool")
	err = os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho hello"), 0755)
	require.NoError(t, err)

	// Should find the binary in the specified binaries path
	path, err := LookPathWithBinaries("fake-tool", tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, fakeBinary, path)
}

func TestLookPathWithBinaries_BinariesPathPriority(t *testing.T) {
	// Create a temp directory with a binary that shadows a system command
	tmpDir, err := os.MkdirTemp("", "test-binaries")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a fake "echo" in our binaries path
	fakeBinary := filepath.Join(tmpDir, "echo")
	err = os.WriteFile(fakeBinary, []byte("#!/bin/sh\necho custom"), 0755)
	require.NoError(t, err)

	// Should find our custom echo, not the system one
	path, err := LookPathWithBinaries("echo", tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, fakeBinary, path)
}

func TestLookPathWithBinaries_NonExecutable(t *testing.T) {
	// Create a temp directory with a non-executable file
	tmpDir, err := os.MkdirTemp("", "test-binaries")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a non-executable file
	nonExecFile := filepath.Join(tmpDir, "not-executable")
	err = os.WriteFile(nonExecFile, []byte("data"), 0644)
	require.NoError(t, err)

	// Should not find the non-executable file, should fall back to system PATH
	_, err = LookPathWithBinaries("not-executable", tmpDir)
	// This should error because the file is not executable and not in system PATH
	assert.Error(t, err)
}

func TestLookPathWithBinaries_Directory(t *testing.T) {
	// Create a temp directory structure
	tmpDir, err := os.MkdirTemp("", "test-binaries")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a subdirectory with the same name as a potential binary
	subDir := filepath.Join(tmpDir, "some-tool")
	err = os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	// Should not treat directories as binaries
	_, err = LookPathWithBinaries("some-tool", tmpDir)
	// This should error because it's a directory, not a file
	assert.Error(t, err)
}
