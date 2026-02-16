package e2e

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSudo_TipMessage verifies that running a workflow containing sudo commands
// as a non-root user without --sudo-aware prints the tip message.
func TestSudo_TipMessage(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("test must run as non-root user")
	}

	log := NewTestLogger(t)
	log.Step("Testing sudo tip message (non-root, no --sudo-aware)")

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-sudo", "-t", "localhost", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains sudo-aware tip")
	assert.Contains(t, stdout, "sudo")
	assert.Contains(t, stdout, "--sudo-aware")

	log.Success("sudo tip message displayed correctly")
}

// TestSudo_NoTipWhenNoSudo verifies that workflows without sudo commands
// do not print the sudo tip.
func TestSudo_NoTipWhenNoSudo(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("test must run as non-root user")
	}

	log := NewTestLogger(t)
	log.Step("Testing no sudo tip for workflow without sudo")

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-bash", "-t", "localhost", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout does NOT contain sudo-aware tip")
	assert.NotContains(t, stdout, "--sudo-aware")

	log.Success("no false sudo tip for non-sudo workflow")
}

// TestSudo_FlagAccepted verifies that --sudo-aware flag is accepted by the CLI.
// NOTE: This test requires an interactive sudo prompt and is intended for
// `make test-sudo`. It authenticates sudo once and keeps credentials alive.
func TestSudo_FlagAccepted(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("test must run as non-root user")
	}
	if os.Getenv("OSM_TEST_SUDO") == "" {
		t.Skip("set OSM_TEST_SUDO=1 to run interactive sudo tests")
	}

	log := NewTestLogger(t)
	log.Step("Testing --sudo-aware flag with sudo workflow")

	workflowPath := getTestdataPath(t)
	stdout, _, err := runCLIWithLog(t, log, "run", "-m", "test-sudo", "-t", "localhost", "--sudo-aware", "--dry-run", "-F", workflowPath)
	require.NoError(t, err)

	log.Info("Asserting stdout contains authentication message")
	assert.Contains(t, stdout, "Sudo commands detected")

	log.Success("--sudo-aware flag triggers authentication")
}
