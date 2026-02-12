package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- ssh_exec input validation tests ---

func TestSSHExec_EmptyHost(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_exec("", "whoami")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestSSHExec_UndefinedHost(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_exec(undefined, "whoami")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestSSHExec_EmptyCommand(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_exec("10.0.0.1", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestSSHExec_UndefinedCommand(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_exec("10.0.0.1")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestSSHExec_NoArgs(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_exec()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, "", result)
}

// --- ssh_rsync input validation tests ---

func TestSSHRsync_EmptyHost(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_rsync("", "/tmp/src", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSSHRsync_UndefinedHost(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_rsync(undefined, "/tmp/src", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSSHRsync_EmptySrc(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_rsync("10.0.0.1", "", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSSHRsync_EmptyDest(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_rsync("10.0.0.1", "/tmp/src", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSSHRsync_NoArgs(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_rsync()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

// --- parseSSHConfig tests ---

func TestSSHExec_DefaultUserAndPort(t *testing.T) {
	// When only host and command are provided, defaults should be used (user=root, port=22)
	// This will fail to connect but validates that defaults don't cause panics
	registry := NewRegistry()
	result, err := registry.Execute(
		`ssh_exec("192.0.2.1", "echo test")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	// Connection to unreachable host will fail, returns empty string
	assert.Equal(t, "", result)
}

// NOTE: Integration tests for ssh_exec and ssh_rsync against a real SSH server
// are in the e2e test suite (test-e2e-ssh). Unit tests here only validate
// input validation since connecting to a real host would be slow/flaky.
