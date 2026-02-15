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
	assert.Contains(t, stdout, "set")

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

func TestWorker_EvalHelp(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing worker eval help command")

	stdout, _, err := runCLIWithLog(t, log, "worker", "eval", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains expected flags")
	assert.Contains(t, stdout, "--redis-url")
	assert.Contains(t, stdout, "--eval")
	assert.Contains(t, stdout, "--target")
	assert.Contains(t, stdout, "--stdin")

	log.Success("worker eval help displays all expected flags")
}

func TestWorker_LsAlias(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing worker ls alias for status")

	stdout, _, err := runCLIWithLog(t, log, "worker", "ls", "--help")
	require.NoError(t, err)

	log.Info("Asserting ls alias works like status")
	assert.Contains(t, stdout, "--redis-url")

	log.Success("worker ls alias works")
}

func TestWorker_SetHelp(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing worker set help command")

	stdout, _, err := runCLIWithLog(t, log, "worker", "set", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains expected flags and field info")
	assert.Contains(t, stdout, "--redis-url")
	assert.Contains(t, stdout, "alias")
	assert.Contains(t, stdout, "public-ip")

	log.Success("worker set help displays expected info")
}

func TestWorker_StatusJSON(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing worker status --json flag exists")

	stdout, _, err := runCLIWithLog(t, log, "worker", "status", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains --json flag")
	assert.Contains(t, stdout, "--json")

	log.Success("worker status help shows --json flag")
}

func TestWorker_JoinGetPublicIP(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing worker join --get-public-ip flag exists")

	stdout, _, err := runCLIWithLog(t, log, "worker", "join", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains --get-public-ip flag")
	assert.Contains(t, stdout, "--get-public-ip")

	log.Success("worker join help shows --get-public-ip flag")
}
