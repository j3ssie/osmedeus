package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealth_Basic(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing health command basic")

	stdout, _, err := runCLIWithLog(t, log, "health")
	require.NoError(t, err)

	log.Info("Asserting stdout contains Folders section")
	assert.Contains(t, stdout, "Folders")

	log.Success("health command displays folder status")
}

func TestHealth_WithBaseFolder(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing health command with custom base folder")

	tmpDir := t.TempDir()
	log.Info("Using temp base folder: %s", tmpDir)

	stdout, _, err := runCLIWithLog(t, log, "health", "-b", tmpDir)
	require.NoError(t, err)

	log.Info("Asserting stdout contains Folders section")
	assert.Contains(t, stdout, "Folders")

	log.Success("health command works with custom base folder")
}
