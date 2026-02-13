package runner

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHostRunner_Execute(t *testing.T) {
	ctx := context.Background()
	runner := NewHostRunner("")

	err := runner.Setup(ctx)
	require.NoError(t, err)
	defer func() { _ = runner.Cleanup(ctx) }()

	result, err := runner.Execute(ctx, "echo hello")

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Output, "hello")
}

func TestHostRunner_Type(t *testing.T) {
	runner := NewHostRunner("")
	assert.Equal(t, core.RunnerTypeHost, runner.Type())
	assert.False(t, runner.IsRemote())
}

func TestHostRunner_ExitCode(t *testing.T) {
	ctx := context.Background()
	runner := NewHostRunner("")

	result, err := runner.Execute(ctx, "exit 1")

	require.NoError(t, err)
	assert.Equal(t, 1, result.ExitCode)
}

func TestHostRunner_WithBinariesPath(t *testing.T) {
	ctx := context.Background()
	runner := NewHostRunner("/tmp/test-binaries")

	result, err := runner.Execute(ctx, "echo $PATH")

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Output, "/tmp/test-binaries")
}

// Integration test - requires Docker
func TestDockerRunner_Execute_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	config := &core.RunnerConfig{
		Image:      "alpine:latest",
		Persistent: false,
	}

	runner, err := NewDockerRunner(config, "")
	require.NoError(t, err)

	err = runner.Setup(ctx)
	require.NoError(t, err)
	defer func() { _ = runner.Cleanup(ctx) }()

	result, err := runner.Execute(ctx, "echo hello from docker")

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Output, "hello from docker")
}

func TestDockerRunner_Type(t *testing.T) {
	config := &core.RunnerConfig{
		Image: "alpine:latest",
	}
	runner, err := NewDockerRunner(config, "")
	require.NoError(t, err)

	assert.Equal(t, core.RunnerTypeDocker, runner.Type())
	assert.True(t, runner.IsRemote())
}

func TestDockerRunner_RequiresImage(t *testing.T) {
	config := &core.RunnerConfig{}
	_, err := NewDockerRunner(config, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image")
}

// Integration test - requires SSH server (linuxserver/openssh-server)
func TestSSHRunner_Execute_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	conn, err := net.DialTimeout("tcp", "localhost:2222", 500*time.Millisecond)
	if err != nil {
		t.Skip("skipping integration test: SSH server not available on localhost:2222")
	}
	_ = conn.Close()

	ctx := context.Background()
	config := &core.RunnerConfig{
		Host:     "localhost",
		Port:     2222,
		User:     "testuser",
		Password: "testpass",
	}

	runner, err := NewSSHRunner(config, "")
	require.NoError(t, err)

	err = runner.Setup(ctx)
	require.NoError(t, err)
	defer func() { _ = runner.Cleanup(ctx) }()

	result, err := runner.Execute(ctx, "echo hello from ssh")

	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Output, "hello from ssh")
}

func TestSSHRunner_Type(t *testing.T) {
	config := &core.RunnerConfig{
		Host: "localhost",
		User: "test",
	}
	runner, err := NewSSHRunner(config, "")
	require.NoError(t, err)

	assert.Equal(t, core.RunnerTypeSSH, runner.Type())
	assert.True(t, runner.IsRemote())
}

func TestSSHRunner_RequiresHost(t *testing.T) {
	config := &core.RunnerConfig{
		User: "test",
	}
	_, err := NewSSHRunner(config, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

func TestSSHRunner_RequiresUser(t *testing.T) {
	config := &core.RunnerConfig{
		Host: "localhost",
	}
	_, err := NewSSHRunner(config, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user")
}

func TestNewRunner_Host(t *testing.T) {
	workflow := &core.Workflow{
		Name:   "test",
		Kind:   core.KindModule,
		Runner: core.RunnerTypeHost,
	}

	runner, err := NewRunner(workflow, "")
	require.NoError(t, err)
	assert.Equal(t, core.RunnerTypeHost, runner.Type())
}

func TestNewRunner_DefaultsToHost(t *testing.T) {
	workflow := &core.Workflow{
		Name: "test",
		Kind: core.KindModule,
	}

	runner, err := NewRunner(workflow, "")
	require.NoError(t, err)
	assert.Equal(t, core.RunnerTypeHost, runner.Type())
}

// ============================================================================
// LimitedBuffer tests
// ============================================================================

func TestLimitedBuffer_WritesUnderLimit(t *testing.T) {
	buf := NewLimitedBuffer(100)

	n, err := buf.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, 5, buf.Len())
	assert.False(t, buf.Overflow())
	assert.Equal(t, "hello", string(buf.Bytes()))
}

func TestLimitedBuffer_TruncatesAtLimit(t *testing.T) {
	buf := NewLimitedBuffer(10)

	n, err := buf.Write([]byte("hello world!")) // 12 bytes > 10 limit
	require.NoError(t, err)
	assert.Equal(t, 12, n) // reports full length written
	assert.Equal(t, 10, buf.Len())
	assert.True(t, buf.Overflow())
	assert.Equal(t, "hello worl", string(buf.Bytes()))
}

func TestLimitedBuffer_DiscardsAfterFull(t *testing.T) {
	buf := NewLimitedBuffer(5)

	_, _ = buf.Write([]byte("hello"))
	assert.Equal(t, 5, buf.Len())
	assert.False(t, buf.Overflow())

	// Further writes are silently discarded
	n, err := buf.Write([]byte(" world"))
	require.NoError(t, err)
	assert.Equal(t, 6, n) // reports full length
	assert.Equal(t, 5, buf.Len())
	assert.True(t, buf.Overflow())
	assert.Equal(t, "hello", string(buf.Bytes()))
}

func TestLimitedBuffer_MultipleWrites(t *testing.T) {
	buf := NewLimitedBuffer(10)

	_, _ = buf.Write([]byte("aaa"))  // 3 bytes, total 3
	_, _ = buf.Write([]byte("bbb"))  // 3 bytes, total 6
	_, _ = buf.Write([]byte("ccc"))  // 3 bytes, total 9
	_, _ = buf.Write([]byte("dddd")) // 4 bytes, only 1 fits -> total 10

	assert.Equal(t, 10, buf.Len())
	assert.True(t, buf.Overflow())
	assert.Equal(t, "aaabbbcccd", string(buf.Bytes()))
}

func TestLimitedBuffer_ZeroSize(t *testing.T) {
	buf := NewLimitedBuffer(0)

	n, err := buf.Write([]byte("anything"))
	require.NoError(t, err)
	assert.Equal(t, 8, n)
	assert.Equal(t, 0, buf.Len())
	assert.True(t, buf.Overflow())
}

func TestCombineOutput_Normal(t *testing.T) {
	stdout := NewLimitedBuffer(100)
	stderr := NewLimitedBuffer(100)

	_, _ = stdout.Write([]byte("out"))
	_, _ = stderr.Write([]byte("err"))

	result := combineOutput(stdout, stderr)
	assert.Equal(t, "outerr", result)
}

func TestCombineOutput_Empty(t *testing.T) {
	stdout := NewLimitedBuffer(100)
	stderr := NewLimitedBuffer(100)

	result := combineOutput(stdout, stderr)
	assert.Equal(t, "", result)
}

func TestCombineOutput_Truncated(t *testing.T) {
	stdout := NewLimitedBuffer(5)
	stderr := NewLimitedBuffer(5)

	_, _ = stdout.Write([]byte("long output that overflows"))
	_, _ = stderr.Write([]byte("err"))

	result := combineOutput(stdout, stderr)
	assert.True(t, stdout.Overflow())
	assert.Contains(t, result, "[output truncated]")
}

func TestHostRunner_LargeOutput_Bounded(t *testing.T) {
	ctx := context.Background()
	runner := NewHostRunner("")

	// Generate output larger than MaxOutputSize limit
	// Use printf to generate ~20KB of output (well under limit, but proves buffer works)
	result, err := runner.Execute(ctx, "printf '%0.s-' {1..20000}")
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.True(t, len(result.Output) > 0)
	assert.True(t, len(result.Output) <= MaxOutputSize+20) // +20 for "[output truncated]\n"

	// Verify that very large output gets truncated
	// Generate output of ~11MB (> 10MB limit)
	bigResult, err := runner.Execute(ctx, "head -c 11000000 /dev/zero | tr '\\0' 'A'")
	require.NoError(t, err)
	assert.True(t, len(bigResult.Output) <= MaxOutputSize+20)
	if len(bigResult.Output) > MaxOutputSize {
		assert.Contains(t, bigResult.Output, "[output truncated]")
	}
}
