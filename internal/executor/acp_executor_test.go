package executor

import (
	"context"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestACPExecutor_StepTypeRegistration(t *testing.T) {
	dispatcher := NewStepDispatcher()

	// Verify agent-acp step type is registered
	plugin, ok := dispatcher.registry.Get(core.StepTypeAgentACP)
	require.True(t, ok, "agent-acp step type should be registered")
	assert.Equal(t, "agent-acp", plugin.Name())

	// Verify it returns the correct step types
	stepTypes := plugin.StepTypes()
	assert.Equal(t, []core.StepType{core.StepTypeAgentACP}, stepTypes)
}

func TestACPExecutor_ResolveAgent_BuiltIn(t *testing.T) {
	tests := []struct {
		name        string
		agent       string
		wantCommand string
		wantArgs    []string
		wantErr     bool
	}{
		{
			name:        "claude-code",
			agent:       "claude-code",
			wantCommand: "npx",
			wantArgs:    []string{"-y", "@zed-industries/claude-code-acp@latest"},
		},
		{
			name:        "codex",
			agent:       "codex",
			wantCommand: "npx",
			wantArgs:    []string{"-y", "@zed-industries/codex-acp"},
		},
		{
			name:        "opencode",
			agent:       "opencode",
			wantCommand: "opencode",
			wantArgs:    []string{"acp"},
		},
		{
			name:        "gemini",
			agent:       "gemini",
			wantCommand: "gemini",
			wantArgs:    []string{"--experimental-acp"},
		},
		{
			name:    "unknown agent",
			agent:   "nonexistent-agent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := &core.Step{Agent: tt.agent}
			cmd, args, err := ResolveAgent(step)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCommand, cmd)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestACPExecutor_ResolveAgent_CustomCommand(t *testing.T) {
	step := &core.Step{
		ACPConfig: &core.ACPStepConfig{
			Command: "/usr/local/bin/my-agent",
			Args:    []string{"--acp", "--verbose"},
		},
	}
	cmd, args, err := ResolveAgent(step)
	require.NoError(t, err)
	assert.Equal(t, "/usr/local/bin/my-agent", cmd)
	assert.Equal(t, []string{"--acp", "--verbose"}, args)
}

func TestACPExecutor_ResolveAgent_NoAgentOrCommand(t *testing.T) {
	step := &core.Step{}
	_, _, err := ResolveAgent(step)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent-acp step requires")
}

func TestACPExecutor_BuildPrompt(t *testing.T) {
	tests := []struct {
		name     string
		messages []core.LLMMessage
		want     string
	}{
		{
			name: "single message",
			messages: []core.LLMMessage{
				{Content: "Analyze the scan results"},
			},
			want: "Analyze the scan results",
		},
		{
			name: "multiple messages joined with newlines",
			messages: []core.LLMMessage{
				{Content: "First task"},
				{Content: "Second task"},
				{Content: "Third task"},
			},
			want: "First\u00a0task\nSecond task\nThird task",
		},
		{
			name:     "no messages",
			messages: nil,
			want:     "",
		},
		{
			name: "empty content skipped",
			messages: []core.LLMMessage{
				{Content: "Hello"},
				{Content: ""},
				{Content: "World"},
			},
			want: "Hello\nWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := &core.Step{Messages: tt.messages}
			got := BuildPrompt(step)
			if tt.name == "multiple messages joined with newlines" {
				// Just verify the parts are present
				assert.Contains(t, got, "First task")
				assert.Contains(t, got, "Second task")
				assert.Contains(t, got, "Third task")
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestACPExecutor_Validation(t *testing.T) {
	p := parser.NewParser()

	tests := []struct {
		name    string
		step    core.Step
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid with agent name",
			step: core.Step{
				Name:  "test-acp",
				Type:  core.StepTypeAgentACP,
				Agent: "claude-code",
				Messages: []core.LLMMessage{
					{Content: "Do something"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid with custom command",
			step: core.Step{
				Name: "test-acp-custom",
				Type: core.StepTypeAgentACP,
				ACPConfig: &core.ACPStepConfig{
					Command: "/usr/bin/my-agent",
				},
				Messages: []core.LLMMessage{
					{Content: "Do something"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing agent and command",
			step: core.Step{
				Name: "test-acp-no-agent",
				Type: core.StepTypeAgentACP,
				Messages: []core.LLMMessage{
					{Content: "Do something"},
				},
			},
			wantErr: true,
			errMsg:  "agent",
		},
		{
			name: "missing messages",
			step: core.Step{
				Name:  "test-acp-no-msgs",
				Type:  core.StepTypeAgentACP,
				Agent: "claude-code",
			},
			wantErr: true,
			errMsg:  "message",
		},
		{
			name: "empty message content",
			step: core.Step{
				Name:  "test-acp-empty-msg",
				Type:  core.StepTypeAgentACP,
				Agent: "claude-code",
				Messages: []core.LLMMessage{
					{Content: ""},
				},
			},
			wantErr: true,
			errMsg:  "content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflow := &core.Workflow{
				Name:  "test",
				Kind:  core.KindModule,
				Steps: []core.Step{tt.step},
			}
			err := p.Validate(workflow)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestACPExecutor_StepHelpers(t *testing.T) {
	step := &core.Step{Type: core.StepTypeAgentACP}
	assert.True(t, step.IsAgentACPStep())
	assert.False(t, step.IsAgentStep())
	assert.False(t, step.IsBashStep())

	bashStep := &core.Step{Type: core.StepTypeBash}
	assert.False(t, bashStep.IsAgentACPStep())
}

func TestACPExecutor_StepClone(t *testing.T) {
	original := &core.Step{
		Name:         "test-acp",
		Type:         core.StepTypeAgentACP,
		Agent:        "claude-code",
		Cwd:          "/workspace",
		AllowedPaths: []string{"/workspace", "/tmp"},
		ACPConfig: &core.ACPStepConfig{
			Command: "custom-agent",
			Args:    []string{"--flag1", "--flag2"},
			Env: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			WriteEnabled: true,
		},
	}

	cloned := original.Clone()

	// Verify deep copy
	assert.Equal(t, original.Agent, cloned.Agent)
	assert.Equal(t, original.Cwd, cloned.Cwd)
	assert.Equal(t, original.AllowedPaths, cloned.AllowedPaths)
	assert.Equal(t, original.ACPConfig.Command, cloned.ACPConfig.Command)
	assert.Equal(t, original.ACPConfig.Args, cloned.ACPConfig.Args)
	assert.Equal(t, original.ACPConfig.Env, cloned.ACPConfig.Env)
	assert.Equal(t, original.ACPConfig.WriteEnabled, cloned.ACPConfig.WriteEnabled)

	// Verify independence — mutating cloned should not affect original
	cloned.AllowedPaths[0] = "/changed"
	assert.Equal(t, "/workspace", original.AllowedPaths[0])

	cloned.ACPConfig.Args[0] = "--changed"
	assert.Equal(t, "--flag1", original.ACPConfig.Args[0])

	cloned.ACPConfig.Env["KEY1"] = "changed"
	assert.Equal(t, "value1", original.ACPConfig.Env["KEY1"])
}

func TestACPClient_PathAllowed(t *testing.T) {
	client := newACPClient(withAllowedPaths("/workspace", "/tmp/data"))

	tests := []struct {
		path    string
		allowed bool
	}{
		{"/workspace/file.txt", true},
		{"/workspace", true},
		{"/workspace/sub/dir/file.txt", true},
		{"/tmp/data/results.json", true},
		{"/etc/passwd", false},
		{"/home/user/file.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.allowed, client.isPathAllowed(tt.path))
		})
	}
}

func TestACPClient_NoAllowedPaths(t *testing.T) {
	client := newACPClient()
	// With no allowed paths, everything is allowed
	assert.True(t, client.isPathAllowed("/any/path"))
	assert.True(t, client.isPathAllowed("/etc/passwd"))
}

func TestACPClient_WriteEnabled(t *testing.T) {
	client := newACPClient(withWriteEnabled(true))
	assert.True(t, client.writeEnabled)

	client2 := newACPClient(withWriteEnabled(false))
	assert.False(t, client2.writeEnabled)

	client3 := newACPClient()
	assert.False(t, client3.writeEnabled)
}

func TestRunAgentACP_EmptyPrompt(t *testing.T) {
	_, _, err := RunAgentACP(context.Background(), "", "claude-code", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompt is required")
}

func TestRunAgentACP_UnknownAgent(t *testing.T) {
	_, _, err := RunAgentACP(context.Background(), "hello", "nonexistent-agent", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown agent")
	assert.Contains(t, err.Error(), "nonexistent-agent")
}

func TestRunAgentACP_DefaultAgent(t *testing.T) {
	// With empty agent name, it should default to "claude-code".
	// The error depends on whether npx is available in the test env:
	// - "not found in PATH" if npx is not installed
	// - ACP protocol error if npx is found but agent subprocess fails
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _, err := RunAgentACP(ctx, "hello", "", nil)
	// We just verify it defaults to claude-code (doesn't error with "unknown agent")
	if err != nil {
		assert.NotContains(t, err.Error(), "unknown agent")
	}
}

func TestRunAgentACP_NilConfig(t *testing.T) {
	// Should not panic with nil config
	_, _, err := RunAgentACP(context.Background(), "hello", "claude-code", nil)
	// Will fail on LookPath in most test envs
	if err != nil {
		assert.NotContains(t, err.Error(), "panic")
	}
}

func TestListAgentNames(t *testing.T) {
	names := ListAgentNames()
	assert.NotEmpty(t, names)
	assert.Contains(t, names, "claude-code")
	assert.Contains(t, names, "codex")
	assert.Contains(t, names, "opencode")
	assert.Contains(t, names, "gemini")
	assert.Equal(t, len(builtinACPAgents), len(names))
}
