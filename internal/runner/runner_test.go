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
