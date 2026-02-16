package functions

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// toStringSlice converts a goja export result to []string.
// Handles both []string and []interface{} from the VM.
func toStringSlice(v interface{}) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []interface{}:
		out := make([]string, len(s))
		for i, item := range s {
			out[i] = item.(string)
		}
		return out
	default:
		return nil
	}
}

func TestGenerateTmuxSessionName(t *testing.T) {
	name := generateTmuxSessionName()
	assert.True(t, strings.HasPrefix(name, "bosm-"), "should have bosm- prefix")
	// "bosm-" (5) + 8 random chars = 13
	assert.Equal(t, 13, len(name), "should be 13 characters total")
}

func TestGenerateTmuxSessionNameUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		name := generateTmuxSessionName()
		assert.False(t, seen[name], "should generate unique names, got duplicate: %s", name)
		seen[name] = true
	}
}

func TestFindTmuxBin(t *testing.T) {
	path := findTmuxBin()
	// Just verify it returns a non-empty string if tmux is installed,
	// or empty string if not. Either way is valid.
	if path != "" {
		assert.True(t, strings.Contains(path, "tmux"), "path should contain 'tmux'")
	}
}

func TestTmuxSessionExistsNonexistent(t *testing.T) {
	tmuxBin := findTmuxBin()
	if tmuxBin == "" {
		t.Skip("tmux not installed")
	}
	assert.False(t, tmuxSessionExists(tmuxBin, "nonexistent-session-bosm-test-999"))
}

// Integration tests - require tmux to be installed

func tmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func TestTmuxRunAndKill(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()

	// Run a session
	result, err := rt.Execute(`tmux_run("sleep 60")`, nil)
	assert.NoError(t, err)

	sessionName, ok := result.(string)
	assert.True(t, ok, "should return string")
	assert.True(t, strings.HasPrefix(sessionName, "bosm-"), "should have bosm- prefix")

	// Verify session exists
	tmuxBin := findTmuxBin()
	assert.True(t, tmuxSessionExists(tmuxBin, sessionName))

	// Kill the session
	killResult, err := rt.Execute(`tmux_kill("`+sessionName+`")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, true, killResult)

	// Verify session is gone
	assert.False(t, tmuxSessionExists(tmuxBin, sessionName))
}

func TestTmuxRunWithCustomName(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()
	customName := "test-tmux-custom-bosm"

	result, err := rt.Execute(`tmux_run("sleep 60", "`+customName+`")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, customName, result)

	// Cleanup
	_, _ = rt.Execute(`tmux_kill("`+customName+`")`, nil)
}

func TestTmuxCapture(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()

	// Create a long-lived bash session
	result, err := rt.Execute(`tmux_run("bash")`, nil)
	assert.NoError(t, err)

	sessionName := result.(string)

	// Send echo command to the session
	_, err = rt.Execute(`tmux_send("`+sessionName+`", "echo 'hello-tmux-test'")`, nil)
	assert.NoError(t, err)

	// Give the command a moment to execute
	_, _ = rt.Execute(`sleep(1)`, nil)

	// Capture output
	captureResult, err := rt.Execute(`tmux_capture("`+sessionName+`")`, nil)
	assert.NoError(t, err)

	captured, ok := captureResult.(string)
	assert.True(t, ok)
	assert.Contains(t, captured, "hello-tmux-test")

	// Cleanup
	_, _ = rt.Execute(`tmux_kill("`+sessionName+`")`, nil)
}

func TestTmuxSend(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()

	// Create a session with bash
	result, err := rt.Execute(`tmux_run("bash")`, nil)
	assert.NoError(t, err)

	sessionName := result.(string)

	// Send a command
	sendResult, err := rt.Execute(`tmux_send("`+sessionName+`", "echo 'sent-test'")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, true, sendResult)

	// Cleanup
	_, _ = rt.Execute(`tmux_kill("`+sessionName+`")`, nil)
}

func TestTmuxRunEmptyCommand(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()
	result, err := rt.Execute(`tmux_run("")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestTmuxCaptureNonexistent(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()
	result, err := rt.Execute(`tmux_capture("nonexistent-bosm-999")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestTmuxSendNonexistent(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()
	result, err := rt.Execute(`tmux_send("nonexistent-bosm-999", "echo hi")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestTmuxKillNonexistent(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()
	result, err := rt.Execute(`tmux_kill("nonexistent-bosm-999")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, false, result)
}

func TestTmuxList(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()

	// Create two sessions
	result1, err := rt.Execute(`tmux_run("sleep 60", "bosm-list-test-1")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "bosm-list-test-1", result1)

	result2, err := rt.Execute(`tmux_run("sleep 60", "bosm-list-test-2")`, nil)
	assert.NoError(t, err)
	assert.Equal(t, "bosm-list-test-2", result2)

	// List sessions
	listResult, err := rt.Execute(`tmux_list()`, nil)
	assert.NoError(t, err)

	names := toStringSlice(listResult)
	assert.Contains(t, names, "bosm-list-test-1")
	assert.Contains(t, names, "bosm-list-test-2")

	// Cleanup
	_, _ = rt.Execute(`tmux_kill("bosm-list-test-1")`, nil)
	_, _ = rt.Execute(`tmux_kill("bosm-list-test-2")`, nil)
}

func TestTmuxListEmpty(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()

	// tmux_list() should return an array (possibly with other sessions from the system)
	listResult, err := rt.Execute(`tmux_list()`, nil)
	assert.NoError(t, err)

	// Should be a slice type (even if empty)
	names := toStringSlice(listResult)
	assert.NotNil(t, names, "should return a slice")
}

func TestTmuxCaptureAll(t *testing.T) {
	if !tmuxAvailable() {
		t.Skip("tmux not installed")
	}

	rt := NewGojaRuntime()

	// Create two bash sessions with distinct output
	_, err := rt.Execute(`tmux_run("bash", "bosm-capall-1")`, nil)
	assert.NoError(t, err)

	_, err = rt.Execute(`tmux_run("bash", "bosm-capall-2")`, nil)
	assert.NoError(t, err)

	// Send distinct echo commands
	_, err = rt.Execute(`tmux_send("bosm-capall-1", "echo 'MARKER_ALPHA_123'")`, nil)
	assert.NoError(t, err)

	_, err = rt.Execute(`tmux_send("bosm-capall-2", "echo 'MARKER_BETA_456'")`, nil)
	assert.NoError(t, err)

	// Wait for commands to execute
	_, _ = rt.Execute(`sleep(1)`, nil)

	// Capture all
	captureResult, err := rt.Execute(`tmux_capture("all")`, nil)
	assert.NoError(t, err)

	captured, ok := captureResult.(string)
	assert.True(t, ok)

	// Verify both session headers and outputs are present
	assert.Contains(t, captured, "=== session: bosm-capall-1 ===")
	assert.Contains(t, captured, "=== session: bosm-capall-2 ===")
	assert.Contains(t, captured, "MARKER_ALPHA_123")
	assert.Contains(t, captured, "MARKER_BETA_456")

	// Cleanup
	_, _ = rt.Execute(`tmux_kill("bosm-capall-1")`, nil)
	_, _ = rt.Execute(`tmux_kill("bosm-capall-2")`, nil)
}
