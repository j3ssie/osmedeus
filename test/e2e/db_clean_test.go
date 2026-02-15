package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBClean_RequiresForce(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing db clean without --force aborts")

	_, stderr, err := runCLIWithLog(t, log, "db", "clean")
	assert.Error(t, err)
	assert.Contains(t, stderr, "use --force to confirm")

	log.Success("db clean without --force correctly aborts")
}

func TestDBClean_Force(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing db clean --force succeeds")

	stdout, _, err := runCLIWithLog(t, log, "db", "clean", "--force")
	require.NoError(t, err)

	assert.Contains(t, stdout, "recreated with fresh schema")

	log.Success("db clean --force completes successfully")
}

func TestDBClean_WithoutCleanWS_PreservesWorkspaces(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing db clean --force without --clean-ws preserves workspaces")

	stdout, _, err := runCLIWithLog(t, log, "db", "clean", "--force")
	require.NoError(t, err)

	// Without --clean-ws, should NOT mention workspace cleanup
	log.Info("Asserting no workspace cleanup message")
	assert.NotContains(t, stdout, "Workspace data cleaned")
	assert.NotContains(t, stdout, "Removing workspace data")

	log.Success("workspaces preserved when --clean-ws not used")
}

func TestDBClean_CleanWS(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing db clean --force --clean-ws removes workspace directory contents")

	binary := getBinaryPath(t)

	// Set up a base dir with a settings file pointing workspaces to a temp dir
	baseDir := t.TempDir()
	wsDir := filepath.Join(baseDir, "workspaces-test")
	require.NoError(t, os.MkdirAll(wsDir, 0755))

	// Create dummy workspace content
	dummyWS := filepath.Join(wsDir, "example.com")
	require.NoError(t, os.MkdirAll(dummyWS, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dummyWS, "scan.log"), []byte("test data"), 0644))
	log.Info("Created dummy workspace at: %s", dummyWS)

	// Write a settings file with workspaces pointing to our temp dir
	settingsContent := "environments:\n  workspaces: " + wsDir + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "osm-settings.yaml"), []byte(settingsContent), 0644))

	// Verify file exists before clean
	_, err := os.Stat(filepath.Join(dummyWS, "scan.log"))
	require.NoError(t, err, "dummy file should exist before clean")

	// Run db clean --force --clean-ws using --base-folder (which calls ResolvePaths)
	args := []string{"--base-folder", baseDir, "db", "clean", "--force", "--clean-ws"}
	log.Command(args...)

	cmd := exec.Command(binary, args...)
	cmd.Env = append(os.Environ(), "OSM_SKIP_PATH_SETUP=1")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()
	log.Result(stdout, stderr)
	require.NoError(t, err, "db clean --force --clean-ws failed: %s", stderr)

	// Verify workspace content was removed
	log.Info("Asserting workspace content was cleaned")
	_, err = os.Stat(filepath.Join(dummyWS, "scan.log"))
	assert.True(t, os.IsNotExist(err), "workspace files should be removed after --clean-ws")

	// Verify the workspaces directory itself was recreated (empty)
	info, err := os.Stat(wsDir)
	assert.NoError(t, err, "workspaces directory should be recreated")
	assert.True(t, info.IsDir(), "workspaces path should be a directory")

	// Verify success message
	assert.Contains(t, stdout, "Workspace data cleaned")

	log.Success("--clean-ws removes workspace data and recreates empty directory")
}
