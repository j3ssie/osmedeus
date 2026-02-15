package executor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- stripTimeoutPrefix tests ---

func TestStripTimeoutPrefix_BasicTimeout(t *testing.T) {
	r := stripTimeoutPrefix("timeout 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
	assert.Equal(t, 30*time.Second, r.duration)
}

func TestStripTimeoutPrefix_WithKillAfterFlag(t *testing.T) {
	r := stripTimeoutPrefix("timeout -k 10 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
	assert.Equal(t, 30*time.Second, r.duration)
}

func TestStripTimeoutPrefix_WithKillAfterEquals(t *testing.T) {
	r := stripTimeoutPrefix("timeout --kill-after=10s 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_WithSignalFlag(t *testing.T) {
	r := stripTimeoutPrefix("timeout -s SIGKILL 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_WithSignalEquals(t *testing.T) {
	r := stripTimeoutPrefix("timeout --signal=TERM 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_WithForeground(t *testing.T) {
	r := stripTimeoutPrefix("timeout --foreground 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_WithPreserveStatus(t *testing.T) {
	r := stripTimeoutPrefix("timeout --preserve-status 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_WithVerbose(t *testing.T) {
	r := stripTimeoutPrefix("timeout -v 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_MultipleFlags(t *testing.T) {
	r := stripTimeoutPrefix("timeout -k 10 -s SIGKILL --foreground 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_ShortKillAttached(t *testing.T) {
	r := stripTimeoutPrefix("timeout -k5s 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_NoCommandAfterDuration(t *testing.T) {
	r := stripTimeoutPrefix("timeout 30")
	assert.True(t, r.stripped)
	assert.Equal(t, "", r.command)
	assert.Equal(t, 30*time.Second, r.duration)
}

func TestStripTimeoutPrefix_NotTimeout(t *testing.T) {
	r := stripTimeoutPrefix("nuclei -t templates")
	assert.False(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
}

func TestStripTimeoutPrefix_EmptyString(t *testing.T) {
	r := stripTimeoutPrefix("")
	assert.False(t, r.stripped)
	assert.Equal(t, "", r.command)
}

func TestStripTimeoutPrefix_FullPathTimeout(t *testing.T) {
	r := stripTimeoutPrefix("/usr/bin/timeout 30 nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
	assert.Equal(t, 30*time.Second, r.duration)
}

func TestStripTimeoutPrefix_DurationWithSuffix(t *testing.T) {
	r := stripTimeoutPrefix("timeout 2h nuclei -t templates")
	assert.True(t, r.stripped)
	assert.Equal(t, "nuclei -t templates", r.command)
	assert.Equal(t, 2*time.Hour, r.duration)
}

// --- parseTimeoutDuration tests ---

func TestParseTimeoutDuration_Seconds(t *testing.T) {
	assert.Equal(t, 30*time.Second, parseTimeoutDuration("30"))
}

func TestParseTimeoutDuration_SecondsWithSuffix(t *testing.T) {
	assert.Equal(t, 30*time.Second, parseTimeoutDuration("30s"))
}

func TestParseTimeoutDuration_Minutes(t *testing.T) {
	assert.Equal(t, 5*time.Minute, parseTimeoutDuration("5m"))
}

func TestParseTimeoutDuration_Hours(t *testing.T) {
	assert.Equal(t, 2*time.Hour, parseTimeoutDuration("2h"))
}

func TestParseTimeoutDuration_Days(t *testing.T) {
	assert.Equal(t, 24*time.Hour, parseTimeoutDuration("1d"))
}

func TestParseTimeoutDuration_Float(t *testing.T) {
	assert.Equal(t, time.Duration(1.5*float64(time.Second)), parseTimeoutDuration("1.5"))
}

func TestParseTimeoutDuration_FloatWithSuffix(t *testing.T) {
	assert.Equal(t, time.Duration(2.5*float64(time.Minute)), parseTimeoutDuration("2.5m"))
}

func TestParseTimeoutDuration_Empty(t *testing.T) {
	assert.Equal(t, time.Duration(0), parseTimeoutDuration(""))
}

func TestParseTimeoutDuration_Invalid(t *testing.T) {
	assert.Equal(t, time.Duration(0), parseTimeoutDuration("abc"))
}

// --- prependBinariesPath tests ---

func TestPrependBinariesPath_Normal(t *testing.T) {
	cmd, ok := prependBinariesPath("nuclei -t templates", "/opt/tools")
	assert.True(t, ok)
	assert.Equal(t, "/opt/tools/nuclei -t templates", cmd)
}

func TestPrependBinariesPath_AlreadyHasPath(t *testing.T) {
	cmd, ok := prependBinariesPath("/usr/bin/nuclei -t templates", "/opt/tools")
	assert.False(t, ok)
	assert.Equal(t, "/usr/bin/nuclei -t templates", cmd)
}

func TestPrependBinariesPath_RelativePath(t *testing.T) {
	cmd, ok := prependBinariesPath("./nuclei -t templates", "/opt/tools")
	assert.False(t, ok)
	assert.Equal(t, "./nuclei -t templates", cmd)
}

func TestPrependBinariesPath_EmptyBinariesPath(t *testing.T) {
	cmd, ok := prependBinariesPath("nuclei -t templates", "")
	assert.False(t, ok)
	assert.Equal(t, "nuclei -t templates", cmd)
}

func TestPrependBinariesPath_EmptyCommand(t *testing.T) {
	cmd, ok := prependBinariesPath("", "/opt/tools")
	assert.False(t, ok)
	assert.Equal(t, "", cmd)
}

func TestPrependBinariesPath_TrailingSlash(t *testing.T) {
	cmd, ok := prependBinariesPath("nuclei -t templates", "/opt/tools/")
	assert.True(t, ok)
	assert.Equal(t, "/opt/tools/nuclei -t templates", cmd)
}

func TestPrependBinariesPath_SingleWord(t *testing.T) {
	cmd, ok := prependBinariesPath("nuclei", "/opt/tools")
	assert.True(t, ok)
	assert.Equal(t, "/opt/tools/nuclei", cmd)
}

// --- exitCodeError tests ---

func TestExitCodeError_ErrorMessage(t *testing.T) {
	err := newExitCodeErrorf(127, "command not found: %s", "nuclei")
	assert.Equal(t, "command not found: nuclei", err.Error())
	assert.Equal(t, 127, err.code)
}

func TestExitCodeError_ErrorsAs(t *testing.T) {
	err := newExitCodeErrorf(127, "command not found")
	var ecErr *exitCodeError
	require.True(t, errors.As(err, &ecErr))
	assert.Equal(t, 127, ecErr.code)
}

func TestExitCodeError_NotExitCodeError(t *testing.T) {
	err := errors.New("some other error")
	var ecErr *exitCodeError
	assert.False(t, errors.As(err, &ecErr))
}

func TestNewExitCodeError(t *testing.T) {
	err := newExitCodeError(1, "failed")
	assert.Equal(t, 1, err.code)
	assert.Equal(t, "failed", err.Error())
}

// --- mockRunner for integration tests ---

type mockFallbackRunner struct {
	// responses maps command strings to their mock results
	responses map[string]*mockResponse
}

type mockResponse struct {
	output   string
	exitCode int
}

func newMockFallbackRunner() *mockFallbackRunner {
	return &mockFallbackRunner{
		responses: make(map[string]*mockResponse),
	}
}

func (m *mockFallbackRunner) addResponse(command string, output string, exitCode int) {
	m.responses[command] = &mockResponse{output: output, exitCode: exitCode}
}

func (m *mockFallbackRunner) Execute(_ context.Context, command string) (*runner.CommandResult, error) {
	if resp, ok := m.responses[command]; ok {
		return &runner.CommandResult{
			Output:   resp.output,
			ExitCode: resp.exitCode,
		}, nil
	}
	// Default: command not found
	return &runner.CommandResult{
		Output:   "sh: command not found",
		ExitCode: 127,
	}, nil
}

func (m *mockFallbackRunner) Setup(_ context.Context) error           { return nil }
func (m *mockFallbackRunner) Cleanup(_ context.Context) error         { return nil }
func (m *mockFallbackRunner) Type() core.RunnerType                   { return core.RunnerTypeHost }
func (m *mockFallbackRunner) IsRemote() bool                          { return false }
func (m *mockFallbackRunner) SetPIDCallbacks(_, _ runner.PIDCallback) {}
func (m *mockFallbackRunner) CopyFromRemote(_ context.Context, _, _ string) error {
	return nil
}

// --- Integration tests: BashExecutor fallback chain ---

func TestBashExecutor_Fallback_StripTimeout(t *testing.T) {
	mock := newMockFallbackRunner()
	// Original command with timeout fails (127)
	// Stripped command succeeds
	mock.addResponse("nuclei -t templates", "found 5 results", 0)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	output, err := exec.executeCommandWithFallback(
		context.Background(),
		"timeout 30 nuclei -t templates",
		0,
		"",
	)

	assert.NoError(t, err)
	assert.Equal(t, "found 5 results", output)
}

func TestBashExecutor_Fallback_PrependBinariesPath(t *testing.T) {
	mock := newMockFallbackRunner()
	// Original command fails (127)
	// Prepended command succeeds
	mock.addResponse("/opt/tools/nuclei -t templates", "found 3 results", 0)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	output, err := exec.executeCommandWithFallback(
		context.Background(),
		"nuclei -t templates",
		0,
		"/opt/tools",
	)

	assert.NoError(t, err)
	assert.Equal(t, "found 3 results", output)
}

func TestBashExecutor_Fallback_StripTimeoutThenPrependBinaries(t *testing.T) {
	mock := newMockFallbackRunner()
	// Original "timeout 30 nuclei -t templates" → 127
	// Stripped "nuclei -t templates" → still 127
	// Prepended "/opt/tools/nuclei -t templates" → success
	mock.addResponse("/opt/tools/nuclei -t templates", "found 7 results", 0)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	output, err := exec.executeCommandWithFallback(
		context.Background(),
		"timeout 30 nuclei -t templates",
		0,
		"/opt/tools",
	)

	assert.NoError(t, err)
	assert.Equal(t, "found 7 results", output)
}

func TestBashExecutor_Fallback_NoFallbackOnNon127(t *testing.T) {
	mock := newMockFallbackRunner()
	// Command fails with exit code 1 (not 127)
	mock.addResponse("nuclei -t templates", "error output", 1)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	_, err := exec.executeCommandWithFallback(
		context.Background(),
		"nuclei -t templates",
		0,
		"/opt/tools",
	)

	require.Error(t, err)
	var ecErr *exitCodeError
	require.True(t, errors.As(err, &ecErr))
	assert.Equal(t, 1, ecErr.code)
}

func TestBashExecutor_Fallback_SuccessNoFallback(t *testing.T) {
	mock := newMockFallbackRunner()
	// Command succeeds on first try
	mock.addResponse("nuclei -t templates", "all good", 0)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	output, err := exec.executeCommandWithFallback(
		context.Background(),
		"nuclei -t templates",
		0,
		"/opt/tools",
	)

	assert.NoError(t, err)
	assert.Equal(t, "all good", output)
}

func TestBashExecutor_Fallback_AllFallbacksFail(t *testing.T) {
	mock := newMockFallbackRunner()
	// All commands return 127 (everything is command not found)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	_, err := exec.executeCommandWithFallback(
		context.Background(),
		"timeout 30 nuclei -t templates",
		0,
		"/opt/tools",
	)

	require.Error(t, err)
	var ecErr *exitCodeError
	require.True(t, errors.As(err, &ecErr))
	assert.Equal(t, 127, ecErr.code)
}

func TestBashExecutor_Fallback_CancelledContextNoRetry(t *testing.T) {
	mock := newMockFallbackRunner()
	// All responses return 127
	// But we expect no fallback because context is cancelled

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := exec.executeCommandWithFallback(
		ctx,
		"timeout 30 nuclei -t templates",
		0,
		"/opt/tools",
	)

	// Should fail due to context cancellation, not 127 fallback
	require.Error(t, err)
}

func TestBashExecutor_Fallback_TimeoutFlagsWithEquals(t *testing.T) {
	mock := newMockFallbackRunner()
	// timeout with --kill-after= and --signal= flags
	mock.addResponse("nuclei -t templates", "found results", 0)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	output, err := exec.executeCommandWithFallback(
		context.Background(),
		"timeout --kill-after=10s --signal=TERM 30 nuclei -t templates",
		0,
		"",
	)

	assert.NoError(t, err)
	assert.Equal(t, "found results", output)
}

func TestBashExecutor_Fallback_BinaryWithPathNoFallback2(t *testing.T) {
	mock := newMockFallbackRunner()
	// Binary already has a path prefix — fallback 2 (prepend) should not apply
	// Fallback 1 (strip timeout) should work
	mock.addResponse("/usr/local/bin/nuclei -t templates", "found", 0)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	output, err := exec.executeCommandWithFallback(
		context.Background(),
		"timeout 30 /usr/local/bin/nuclei -t templates",
		0,
		"/opt/tools",
	)

	assert.NoError(t, err)
	assert.Equal(t, "found", output)
}

func TestBashExecutor_Fallback_SequentialUsesWithFallback(t *testing.T) {
	mock := newMockFallbackRunner()
	// First command needs fallback (strip timeout)
	mock.addResponse("nuclei -t templates", "result1", 0)
	// Second command succeeds directly
	mock.addResponse("echo done", "done", 0)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	output, err := exec.executeSequential(
		context.Background(),
		[]string{"timeout 30 nuclei -t templates", "echo done"},
		0,
		"",
	)

	assert.NoError(t, err)
	assert.Contains(t, output, "result1")
	assert.Contains(t, output, "done")
}

func TestBashExecutor_Fallback_ParallelUsesWithFallback(t *testing.T) {
	mock := newMockFallbackRunner()
	mock.addResponse("nuclei -t templates", "result1", 0)
	mock.addResponse("echo done", "done", 0)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	output, err := exec.executeParallel(
		context.Background(),
		[]string{"timeout 30 nuclei -t templates", "echo done"},
		0,
		"",
	)

	assert.NoError(t, err)
	assert.Contains(t, output, "result1")
	assert.Contains(t, output, "done")
}

func TestBashExecutor_Fallback_StripTimeoutNonZeroExitAfterStrip(t *testing.T) {
	mock := newMockFallbackRunner()
	// After stripping timeout, command fails with non-127 exit code
	mock.addResponse("nuclei -t templates", "some error", 2)

	exec := NewBashExecutor(nil)
	exec.SetRunner(mock)

	_, err := exec.executeCommandWithFallback(
		context.Background(),
		"timeout 30 nuclei -t templates",
		time.Duration(0),
		"/opt/tools",
	)

	// Should return the exit code 2 error, not try fallback 2
	require.Error(t, err)
	var ecErr *exitCodeError
	require.True(t, errors.As(err, &ecErr))
	assert.Equal(t, 2, ecErr.code)
}
