package executor

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_RemoteBashStep_MissingStepRunner(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-remote-bash-no-config",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:    "missing-config",
				Type:    core.StepTypeRemoteBash,
				Command: "echo hello",
				// No StepRunner - should fail
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	_, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "step_runner")
}

func TestExecutor_RemoteBashStep_HostRunnerNotSupported(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-remote-bash-host",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:       "host-runner",
				Type:       core.StepTypeRemoteBash,
				StepRunner: core.RunnerTypeHost, // Should fail - host not supported
				Command:    "echo hello",
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	_, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "docker")
}

func TestExecutor_RemoteBashStep_DockerMissingImage(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-remote-bash-docker-no-image",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:       "docker-no-image",
				Type:       core.StepTypeRemoteBash,
				StepRunner: core.RunnerTypeDocker,
				Command:    "echo hello",
				StepRunnerConfig: &core.StepRunnerConfig{
					RunnerConfig: &core.RunnerConfig{
						// No Image - should fail
					},
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	_, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image")
}

func TestExecutor_RemoteBashStep_SSHMissingHost(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-remote-bash-ssh-no-host",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:       "ssh-no-host",
				Type:       core.StepTypeRemoteBash,
				StepRunner: core.RunnerTypeSSH,
				Command:    "echo hello",
				StepRunnerConfig: &core.StepRunnerConfig{
					RunnerConfig: &core.RunnerConfig{
						User: "testuser",
						// No Host - should fail
					},
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	_, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host")
}

// Integration test - requires Docker
func TestExecutor_RemoteBashStep_Docker(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode")
	}

	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-remote-bash-docker",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:       "docker-echo",
				Type:       core.StepTypeRemoteBash,
				StepRunner: core.RunnerTypeDocker,
				Command:    "echo 'Hello from Docker'",
				StepRunnerConfig: &core.StepRunnerConfig{
					RunnerConfig: &core.RunnerConfig{
						Image: "alpine:latest",
					},
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	require.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
	assert.Contains(t, result.Steps[0].Output, "Hello from Docker")
}

// Integration test - requires Docker
func TestExecutor_RemoteBashStep_DockerMultipleCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode")
	}

	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-remote-bash-docker-multi",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:       "docker-multi",
				Type:       core.StepTypeRemoteBash,
				StepRunner: core.RunnerTypeDocker,
				Commands: []string{
					"echo 'First'",
					"echo 'Second'",
					"echo 'Third'",
				},
				StepRunnerConfig: &core.StepRunnerConfig{
					RunnerConfig: &core.RunnerConfig{
						Image: "alpine:latest",
					},
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	require.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
	assert.Contains(t, result.Steps[0].Output, "First")
	assert.Contains(t, result.Steps[0].Output, "Second")
	assert.Contains(t, result.Steps[0].Output, "Third")
}

// Integration test - requires Docker
func TestExecutor_RemoteBashStep_DockerParallelCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode")
	}

	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-remote-bash-docker-parallel",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:       "docker-parallel",
				Type:       core.StepTypeRemoteBash,
				StepRunner: core.RunnerTypeDocker,
				ParallelCommands: []string{
					"echo 'Parallel 1'",
					"echo 'Parallel 2'",
					"echo 'Parallel 3'",
				},
				StepRunnerConfig: &core.StepRunnerConfig{
					RunnerConfig: &core.RunnerConfig{
						Image: "alpine:latest",
					},
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	require.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
	assert.Contains(t, result.Steps[0].Output, "Parallel 1")
	assert.Contains(t, result.Steps[0].Output, "Parallel 2")
	assert.Contains(t, result.Steps[0].Output, "Parallel 3")
}

// Integration test - requires SSH server on localhost:2222
func TestExecutor_RemoteBashStep_SSH(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping SSH integration test in short mode")
	}
	conn, err := net.DialTimeout("tcp", "localhost:2222", 500*time.Millisecond)
	if err != nil {
		t.Skip("skipping SSH integration test: SSH server not available on localhost:2222")
	}
	_ = conn.Close()

	ctx := context.Background()
	cfg := testConfig(t)

	module := &core.Workflow{
		Name: "test-remote-bash-ssh",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:       "ssh-echo",
				Type:       core.StepTypeRemoteBash,
				StepRunner: core.RunnerTypeSSH,
				Command:    "echo 'Hello from SSH'",
				StepRunnerConfig: &core.StepRunnerConfig{
					RunnerConfig: &core.RunnerConfig{
						Host:     "localhost",
						Port:     2222,
						User:     "testuser",
						Password: "testpass",
					},
				},
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	require.Len(t, result.Steps, 1)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
	assert.Contains(t, result.Steps[0].Output, "Hello from SSH")
}

func TestRemoteBashExecutor_CanHandle(t *testing.T) {
	executor := NewRemoteBashExecutor(nil)

	assert.True(t, executor.CanHandle(core.StepTypeRemoteBash))
	assert.False(t, executor.CanHandle(core.StepTypeBash))
	assert.False(t, executor.CanHandle(core.StepTypeFunction))
	assert.False(t, executor.CanHandle(core.StepTypeParallel))
	assert.False(t, executor.CanHandle(core.StepTypeForeach))
}

func TestStepRunnerConfig_Clone(t *testing.T) {
	step := &core.Step{
		Name:       "test-step",
		Type:       core.StepTypeRemoteBash,
		StepRunner: core.RunnerTypeDocker,
		StepRunnerConfig: &core.StepRunnerConfig{
			RunnerConfig: &core.RunnerConfig{
				Image:   "alpine:latest",
				Volumes: []string{"/host:/container"},
				Env:     map[string]string{"KEY": "VALUE"},
			},
		},
	}

	cloned := step.Clone()

	// Verify deep copy
	assert.Equal(t, step.StepRunner, cloned.StepRunner)
	assert.Equal(t, step.StepRunnerConfig.Image, cloned.StepRunnerConfig.Image)
	assert.Equal(t, step.StepRunnerConfig.Volumes, cloned.StepRunnerConfig.Volumes)
	assert.Equal(t, step.StepRunnerConfig.Env, cloned.StepRunnerConfig.Env)

	// Modify cloned and verify original is unchanged
	cloned.StepRunner = core.RunnerTypeSSH
	cloned.StepRunnerConfig.Image = "ubuntu:latest"
	cloned.StepRunnerConfig.Volumes[0] = "/other:/path"
	cloned.StepRunnerConfig.Env["KEY"] = "CHANGED"

	assert.Equal(t, core.RunnerTypeDocker, step.StepRunner)
	assert.Equal(t, "alpine:latest", step.StepRunnerConfig.Image)
	assert.Equal(t, "/host:/container", step.StepRunnerConfig.Volumes[0])
	assert.Equal(t, "VALUE", step.StepRunnerConfig.Env["KEY"])
}
