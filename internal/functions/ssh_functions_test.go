package functions

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

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

// --- sync_from_master input validation tests ---

func TestSyncFromMaster_EmptySrc(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sync_from_master("", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSyncFromMaster_EmptyDest(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sync_from_master("/tmp/src", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSyncFromMaster_NoArgs(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sync_from_master()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

// --- sync_from_worker input validation tests ---

func TestSyncFromWorker_EmptyIdentifier(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sync_from_worker("", "10.0.0.2", "/tmp/src", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSyncFromWorker_EmptySrc(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sync_from_worker("worker-1", "10.0.0.2", "", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSyncFromWorker_EmptyDest(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sync_from_worker("worker-1", "10.0.0.2", "/tmp/src", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSyncFromWorker_NoArgs(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`sync_from_worker()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestSyncFromWorker_NoHost(t *testing.T) {
	// No SSH hooks registered and no explicit IP -> no host resolved
	registry := NewRegistry()
	result, err := registry.Execute(
		`sync_from_worker("worker-1", "", "/tmp/src", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

// --- rsync_to_worker input validation tests ---

func TestRsyncToWorker_EmptyIdentifier(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`rsync_to_worker("", "10.0.0.2", "/tmp/src", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestRsyncToWorker_EmptySrc(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`rsync_to_worker("worker-1", "10.0.0.2", "", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestRsyncToWorker_EmptyDest(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`rsync_to_worker("worker-1", "10.0.0.2", "/tmp/src", "")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestRsyncToWorker_NoArgs(t *testing.T) {
	registry := NewRegistry()
	result, err := registry.Execute(
		`rsync_to_worker()`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestRsyncToWorker_NoHost(t *testing.T) {
	// No SSH hooks registered and no explicit IP -> no host resolved
	registry := NewRegistry()
	result, err := registry.Execute(
		`rsync_to_worker("worker-1", "", "/tmp/src", "/tmp/dest")`,
		map[string]interface{}{},
	)

	require.NoError(t, err)
	assert.Equal(t, false, result)
}

// NOTE: Integration tests for ssh_exec, ssh_rsync, and sync functions against
// a real SSH server are in the e2e test suite (test-e2e-ssh). Unit tests here
// only validate input validation since connecting to a real host would be slow/flaky.

// --- Run-context hook tests ---

// TestRunContextHook_LookupAndDefault verifies that runContextFor falls back
// to context.Background when no hook is registered, and returns the hook's
// context when one is.
func TestRunContextHook_LookupAndDefault(t *testing.T) {
	UnregisterRunContextHooks()
	defer UnregisterRunContextHooks()

	// No hook: should always return non-nil background-equivalent.
	got := runContextFor("any-run")
	assert.NotNil(t, got)
	assert.NoError(t, got.Err(), "background ctx should not be already-cancelled")

	// With hook: should return the hook's context for matching uuid.
	cancellable, cancel := context.WithCancel(context.Background())
	RegisterRunContextHooks(&RunContextHooks{
		Lookup: func(runUUID string) context.Context {
			if runUUID == "run-abc" {
				return cancellable
			}
			return nil
		},
	})

	got = runContextFor("run-abc")
	require.NotNil(t, got)
	cancel()
	select {
	case <-got.Done():
	case <-time.After(time.Second):
		t.Fatal("hook-supplied context should fire when parent cancelled")
	}

	// Unknown uuid: hook returns nil, helper falls back to Background.
	got = runContextFor("unknown")
	require.NotNil(t, got)
	assert.NoError(t, got.Err())
}

// --- SSH exec cancellation integration test ---

// TestSSHExec_RunCancelKillsRemoteProcess verifies that cancelling the active
// run's context (via the run-context hook) propagates into ssh_exec, killing
// the remote process group. Without the fix, the goja function would block
// for its full 5-minute timeout regardless of run cancellation.
func TestSSHExec_RunCancelKillsRemoteProcess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	conn, err := net.DialTimeout("tcp", "localhost:2222", 500*time.Millisecond)
	if err != nil {
		t.Skip("skipping integration test: SSH server not available on localhost:2222")
	}
	_ = conn.Close()

	// Install a run-context hook that returns a cancellable context for our
	// test runUUID. Unregister on exit so we don't pollute global state.
	UnregisterRunContextHooks()
	defer UnregisterRunContextHooks()
	runCtx, cancelRun := context.WithCancel(context.Background())
	RegisterRunContextHooks(&RunContextHooks{
		Lookup: func(runUUID string) context.Context {
			if runUUID == "ssh-exec-cancel-test" {
				return runCtx
			}
			return nil
		},
	})

	markerID := fmt.Sprintf("osm-sshexec-cancel-%d", time.Now().UnixNano())
	markerFile := fmt.Sprintf("/tmp/%s.heartbeat", markerID)

	registry := NewRegistry()
	jsExpr := fmt.Sprintf(
		`ssh_exec("localhost", `+
			"%q"+
			`, "testuser", "", "testpass", 2222)`,
		fmt.Sprintf(`while :; do date +%%s > %s; sleep 1; done # %s`, markerFile, markerID),
	)

	// Run ssh_exec in a goroutine so we can cancel mid-flight.
	resultCh := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = registry.Execute(jsExpr, map[string]interface{}{
			"RunUUID": "ssh-exec-cancel-test",
		})
		close(resultCh)
	}()

	// Wait until the heartbeat file appears, then cancel the run.
	deadline := time.Now().Add(8 * time.Second)
	heartbeatCheck := fmt.Sprintf(`ssh_exec("localhost", "test -s %s && echo ok", "testuser", "", "testpass", 2222)`, markerFile)
	for time.Now().Before(deadline) {
		probe, _ := registry.Execute(heartbeatCheck, map[string]interface{}{})
		if s, ok := probe.(string); ok && strings.Contains(s, "ok") {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	cancelRun()

	// ssh_exec must return promptly after run cancellation, not after the
	// 5-minute internal timeout. Allow some slack for the kill round-trip.
	select {
	case <-resultCh:
	case <-time.After(15 * time.Second):
		t.Fatal("ssh_exec did not return within 15s of run cancellation")
	}

	// Verify the remote process is gone.
	time.Sleep(2 * time.Second)
	psCheck := fmt.Sprintf(`ssh_exec("localhost", "ps -ef 2>/dev/null | grep %s | grep -v grep | wc -l", "testuser", "", "testpass", 2222)`, markerID)
	psResult, _ := registry.Execute(psCheck, map[string]interface{}{})
	psStr, _ := psResult.(string)
	assert.Equal(t, "0", strings.TrimSpace(psStr),
		"expected no remote processes matching %s after cancel; got: %s", markerID, psStr)

	// Cleanup
	_, _ = registry.Execute(
		fmt.Sprintf(`ssh_exec("localhost", "rm -f %s", "testuser", "", "testpass", 2222)`, markerFile),
		map[string]interface{}{},
	)
	wg.Wait()
}
