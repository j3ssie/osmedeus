package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorker_Help(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing worker help command")

	stdout, _, err := runCLIWithLog(t, log, "worker", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains worker subcommands")
	assert.Contains(t, stdout, "join")
	assert.Contains(t, stdout, "status")

	log.Success("worker help displays all subcommands")
}

func TestWorker_JoinHelp(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing worker join help command")

	stdout, _, err := runCLIWithLog(t, log, "worker", "join", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains redis-url option")
	assert.Contains(t, stdout, "--redis-url")

	log.Success("worker join help displays redis options")
}

func TestWorker_StatusHelp(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing worker status help command")

	stdout, _, err := runCLIWithLog(t, log, "worker", "status", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains redis-url option")
	assert.Contains(t, stdout, "--redis-url")

	log.Success("worker status help displays redis options")
}
