package e2e

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion_Flag(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing version flag")

	stdout, _, err := runCLIWithLog(t, log, "--version")
	require.NoError(t, err)

	log.Info("Asserting stdout contains version info")
	assert.Contains(t, stdout, "Version:")
	assert.Contains(t, stdout, "Author:")

	log.Success("version flag displays version and author")
}

func TestVersion_Subcommand(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing version subcommand")

	stdout, _, err := runCLIWithLog(t, log, "version")
	require.NoError(t, err)

	log.Info("Asserting stdout contains version info")
	assert.True(t, strings.Contains(stdout, "Version:") || strings.Contains(stdout, "v5"))

	log.Success("version subcommand displays version info")
}
