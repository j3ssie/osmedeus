package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLLMResponse creates a ChatCompletionResponse with text content and no tool calls
func mockLLMResponse(content string) ChatCompletionResponse {
	return ChatCompletionResponse{
		ID:    "test-id",
		Model: "test-model",
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Usage: ChatUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}
}

// mockLLMToolCallResponse creates a response with tool calls
func mockLLMToolCallResponse(toolCalls []core.LLMToolCall) ChatCompletionResponse {
	return ChatCompletionResponse{
		ID:    "test-id",
		Model: "test-model",
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:      "assistant",
					Content:   "",
					ToolCalls: toolCalls,
				},
				FinishReason: "tool_calls",
			},
		},
		Usage: ChatUsage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}
}

// newMockLLMServer creates a mock OpenAI-compatible API server.
// handler is called for each request and should write the response.
func newMockLLMServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// newMockConfig creates a config with a mock LLM provider pointing to the given URL
func newMockConfig(t *testing.T, serverURL string) *config.Config {
	t.Helper()
	baseDir := t.TempDir()
	return &config.Config{
		BaseFolder:     baseDir,
		WorkspacesPath: filepath.Join(baseDir, "workspaces"),
		LLM: config.LLMConfig{
			LLMProviders: []config.LLMProvider{
				{
					Provider:  "mock",
					BaseURL:   serverURL,
					AuthToken: "test-token",
					Model:     "test-model",
				},
			},
			MaxTokens:  1000,
			MaxRetries: 1,
			Timeout:    "30s",
		},
	}
}

// ============================================================================
// Agent Executor Unit Tests
// ============================================================================

func TestAgentExecutor_Name(t *testing.T) {
	executor := NewAgentExecutor(nil, nil)
	assert.Equal(t, "agent", executor.Name())
}

func TestAgentExecutor_StepTypes(t *testing.T) {
	executor := NewAgentExecutor(nil, nil)
	types := executor.StepTypes()
	assert.Len(t, types, 1)
	assert.Equal(t, core.StepTypeAgent, types[0])
}

func TestAgentExecutor_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	t.Run("missing config", func(t *testing.T) {
		executor := NewAgentExecutor(nil, nil)
		step := &core.Step{Name: "test", Type: core.StepTypeAgent}

		result, err := executor.Execute(ctx, step, execCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config not set")
		assert.Equal(t, core.StepStatusFailed, result.Status)
	})

	t.Run("missing query", func(t *testing.T) {
		server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {})
		defer server.Close()
		cfg := newMockConfig(t, server.URL)

		executor := NewAgentExecutor(nil, nil)
		executor.SetConfig(cfg)

		step := &core.Step{
			Name:          "test",
			Type:          core.StepTypeAgent,
			MaxIterations: 5,
			AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		}

		result, err := executor.Execute(ctx, step, execCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires 'query'")
		assert.Equal(t, core.StepStatusFailed, result.Status)
	})

	t.Run("missing max_iterations", func(t *testing.T) {
		server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {})
		defer server.Close()
		cfg := newMockConfig(t, server.URL)

		executor := NewAgentExecutor(nil, nil)
		executor.SetConfig(cfg)

		step := &core.Step{
			Name:       "test",
			Type:       core.StepTypeAgent,
			Query:      "test query",
			AgentTools: []core.AgentToolDef{{Preset: "bash"}},
		}

		result, err := executor.Execute(ctx, step, execCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires 'max_iterations'")
		assert.Equal(t, core.StepStatusFailed, result.Status)
	})

	t.Run("missing agent_tools", func(t *testing.T) {
		server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {})
		defer server.Close()
		cfg := newMockConfig(t, server.URL)

		executor := NewAgentExecutor(nil, nil)
		executor.SetConfig(cfg)

		step := &core.Step{
			Name:          "test",
			Type:          core.StepTypeAgent,
			Query:         "test query",
			MaxIterations: 5,
		}

		result, err := executor.Execute(ctx, step, execCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "requires 'agent_tools'")
		assert.Equal(t, core.StepStatusFailed, result.Status)
	})

	t.Run("unknown preset tool", func(t *testing.T) {
		server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {})
		defer server.Close()
		cfg := newMockConfig(t, server.URL)

		executor := NewAgentExecutor(nil, nil)
		executor.SetConfig(cfg)

		step := &core.Step{
			Name:          "test",
			Type:          core.StepTypeAgent,
			Query:         "test query",
			MaxIterations: 5,
			AgentTools:    []core.AgentToolDef{{Preset: "nonexistent_tool"}},
		}

		result, err := executor.Execute(ctx, step, execCtx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown preset tool")
		assert.Equal(t, core.StepStatusFailed, result.Status)
	})
}

func TestAgentExecutor_SimpleCompletion(t *testing.T) {
	// Mock server that responds without tool calls (agent should complete in 1 iteration)
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		resp := mockLLMResponse("The analysis is complete.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "simple-agent",
		Type:          core.StepTypeAgent,
		Query:         "Analyze the target",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "The analysis is complete.", result.Output)
	assert.Equal(t, "The analysis is complete.", result.Exports["agent_content"])
	assert.Equal(t, 1, result.Exports["agent_iterations"])
	assert.Equal(t, 30, result.Exports["agent_total_tokens"])
	assert.NotEmpty(t, result.Exports["agent_history"])
}

func TestAgentExecutor_ToolCallLoop(t *testing.T) {
	// Mock server that:
	// 1st call: returns a tool call for exec_cmd
	// 2nd call: returns final text response
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if count == 1 {
			// First call: return tool call
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "bash",
						Arguments: `{"command": "echo hello"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else {
			// Second call: return final response
			resp := mockLLMResponse("Command output was: hello")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "tool-call-agent",
		Type:          core.StepTypeAgent,
		Query:         "Run echo hello",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Command output was: hello", result.Output)
	assert.Equal(t, 2, result.Exports["agent_iterations"])
	// 2 calls * 30 tokens each
	assert.Equal(t, 60, result.Exports["agent_total_tokens"])
}

func TestAgentExecutor_MaxIterationsLimit(t *testing.T) {
	// Mock server that always returns tool calls (never finishes)
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMToolCallResponse([]core.LLMToolCall{
			{
				ID:   "call_1",
				Type: "function",
				Function: core.LLMToolCallFunction{
					Name:      "bash",
					Arguments: `{"command": "echo loop"}`,
				},
			},
		})
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "loop-agent",
		Type:          core.StepTypeAgent,
		Query:         "Keep running",
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, 3, result.Exports["agent_iterations"])
}

func TestAgentExecutor_StopCondition(t *testing.T) {
	// Mock server: first call returns "checking...", second returns "DONE"
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if count == 1 {
			// Return tool call first
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "bash",
						Arguments: `{"command": "echo check"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else {
			// Return text with DONE keyword
			resp := mockLLMResponse("Analysis DONE successfully")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "stop-agent",
		Type:          core.StepTypeAgent,
		Query:         "Analyze and say DONE when finished",
		MaxIterations: 10,
		StopCondition: `contains(agent_content, "DONE")`,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Contains(t, result.Output, "DONE")
	// Should have stopped after 2 iterations (not 10)
	assert.Equal(t, 2, result.Exports["agent_iterations"])
}

func TestAgentExecutor_SystemPrompt(t *testing.T) {
	// Verify that system prompt is included in the request
	var receivedMessages []ChatMessage

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)
		receivedMessages = req.Messages

		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMResponse("Response")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "system-prompt-agent",
		Type:          core.StepTypeAgent,
		SystemPrompt:  "You are a security analyst.",
		Query:         "Analyze target",
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	_, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	// Verify messages contain system prompt and user query
	require.Len(t, receivedMessages, 2)
	assert.Equal(t, "system", receivedMessages[0].Role)
	assert.Equal(t, "You are a security analyst.", receivedMessages[0].Content)
	assert.Equal(t, "user", receivedMessages[1].Role)
	assert.Equal(t, "Analyze target", receivedMessages[1].Content)
}

func TestAgentExecutor_ToolsInRequest(t *testing.T) {
	// Verify that tools are included in the LLM request
	var receivedTools []core.LLMTool

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)
		receivedTools = req.Tools

		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMResponse("Done")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "tools-check",
		Type:          core.StepTypeAgent,
		Query:         "Test",
		MaxIterations: 3,
		AgentTools: []core.AgentToolDef{
			{Preset: "bash"},
			{Preset: "read_file"},
			{
				Name:        "custom",
				Description: "Custom tool",
				Parameters:  map[string]interface{}{"type": "object"},
			},
		},
	}

	_, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	// Verify 3 tools were sent to the LLM
	assert.Len(t, receivedTools, 3)
	assert.Equal(t, "bash", receivedTools[0].Function.Name)
	assert.Equal(t, "read_file", receivedTools[1].Function.Name)
	assert.Equal(t, "custom", receivedTools[2].Function.Name)
}

func TestAgentExecutor_MemoryPersist(t *testing.T) {
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMResponse("Persisted response")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	persistPath := filepath.Join(t.TempDir(), "agent", "conversation.json")

	step := &core.Step{
		Name:          "persist-agent",
		Type:          core.StepTypeAgent,
		Query:         "Test persist",
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		Memory: &core.AgentMemoryConfig{
			PersistPath: persistPath,
		},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)

	// Verify conversation was persisted
	data, err := os.ReadFile(persistPath)
	require.NoError(t, err)

	var messages []ChatMessage
	err = json.Unmarshal(data, &messages)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(messages), 2) // at least user + assistant
}

func TestAgentExecutor_MemoryResume(t *testing.T) {
	// Create a conversation file to resume from
	tmpDir := t.TempDir()
	resumePath := filepath.Join(tmpDir, "resume.json")

	priorMessages := []ChatMessage{
		{Role: "system", Content: "You are a helper."},
		{Role: "user", Content: "Previous question"},
		{Role: "assistant", Content: "Previous answer"},
	}
	data, _ := json.MarshalIndent(priorMessages, "", "  ")
	os.WriteFile(resumePath, data, 0o644)

	var receivedMessages []ChatMessage

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)
		receivedMessages = req.Messages

		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMResponse("Resumed response")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "resume-agent",
		Type:          core.StepTypeAgent,
		Query:         "Follow-up question",
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		Memory: &core.AgentMemoryConfig{
			ResumePath: resumePath,
		},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)

	// Verify resumed messages + new query
	// Should be: system + previous user + previous assistant + new user query
	assert.Len(t, receivedMessages, 4)
	assert.Equal(t, "system", receivedMessages[0].Role)
	assert.Equal(t, "user", receivedMessages[1].Role)
	assert.Equal(t, "Previous question", receivedMessages[1].Content)
	assert.Equal(t, "assistant", receivedMessages[2].Role)
	assert.Equal(t, "user", receivedMessages[3].Role)
	assert.Equal(t, "Follow-up question", receivedMessages[3].Content)
}

func TestAgentExecutor_CustomToolHandler(t *testing.T) {
	// First call returns a tool call for "custom_tool", second returns final answer
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if count == 1 {
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "greet",
						Arguments: `{"name": "World"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := mockLLMResponse("Greeting sent")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "custom-handler-agent",
		Type:          core.StepTypeAgent,
		Query:         "Greet World",
		MaxIterations: 5,
		AgentTools: []core.AgentToolDef{
			{
				Name:        "greet",
				Description: "Greet someone",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{"type": "string"},
					},
				},
				Handler: `"Hello, " + args.name + "!"`,
			},
		},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Greeting sent", result.Output)
}

func TestAgentExecutor_ContextCancellation(t *testing.T) {
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		// This should not be reached because context is already cancelled
		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMResponse("Should not reach")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "cancel-agent",
		Type:          core.StepTypeAgent,
		Query:         "This should be cancelled",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	assert.Error(t, err)
	assert.Equal(t, core.StepStatusFailed, result.Status)
}

func TestAgentExecutor_MultipleToolCalls(t *testing.T) {
	// Mock server: first call returns 2 tool calls, second returns final answer
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if count == 1 {
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "bash",
						Arguments: `{"command": "echo first"}`,
					},
				},
				{
					ID:   "call_2",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "bash",
						Arguments: `{"command": "echo second"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := mockLLMResponse("Both commands executed")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "multi-tool-agent",
		Type:          core.StepTypeAgent,
		Query:         "Run two commands",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Both commands executed", result.Output)
	assert.Equal(t, 2, result.Exports["agent_iterations"])
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestBuildPresetCallExpr(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     map[string]interface{}
		expected string
	}{
		{
			name:     "bash",
			funcName: "bash",
			args:     map[string]interface{}{"command": "echo hello"},
			expected: `bash("echo hello")`,
		},
		{
			name:     "read_file",
			funcName: "read_file",
			args:     map[string]interface{}{"path": "/tmp/test.txt"},
			expected: `read_file("/tmp/test.txt")`,
		},
		{
			name:     "save_content",
			funcName: "save_content",
			args:     map[string]interface{}{"content": "hello", "path": "/tmp/out.txt"},
			expected: `save_content("hello", "/tmp/out.txt")`,
		},
		{
			name:     "grep_string",
			funcName: "grep_string",
			args:     map[string]interface{}{"source": "/tmp/file.txt", "str": "pattern"},
			expected: `grep_string("/tmp/file.txt", "pattern")`,
		},
		{
			name:     "jq",
			funcName: "jq",
			args:     map[string]interface{}{"json_data": `{"key":"value"}`, "expression": ".key"},
			expected: `jq("{\"key\":\"value\"}", ".key")`,
		},
		{
			name:     "exec_python",
			funcName: "exec_python",
			args:     map[string]interface{}{"code": "print('hello')"},
			expected: `exec_python("print('hello')")`,
		},
		{
			name:     "exec_python_file",
			funcName: "exec_python_file",
			args:     map[string]interface{}{"path": "/tmp/script.py"},
			expected: `exec_python_file("/tmp/script.py")`,
		},
		{
			name:     "run_module",
			funcName: "run_module",
			args:     map[string]interface{}{"module": "subdomain", "target": "example.com", "params": "threads=10"},
			expected: `run_module("subdomain", "example.com", "threads=10")`,
		},
		{
			name:     "run_flow",
			funcName: "run_flow",
			args:     map[string]interface{}{"flow": "general", "target": "example.com", "params": ""},
			expected: `run_flow("general", "example.com", "")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPresetCallExpr(tt.funcName, tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJsQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`hello`, `"hello"`},
		{`say "hi"`, `"say \"hi\""`},
		{"line1\nline2", `"line1\nline2"`},
		{`path\to\file`, `"path\\to\\file"`},
		{"tab\there", `"tab\there"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.expected, jsQuote(tt.input))
		})
	}
}

func TestFormatToolResult(t *testing.T) {
	assert.Equal(t, "hello", formatToolResult("hello"))
	assert.Equal(t, "true", formatToolResult(true))
	assert.Equal(t, "false", formatToolResult(false))
	assert.Equal(t, "42", formatToolResult(42))
	assert.Equal(t, "3.14", formatToolResult(3.14))
	assert.Equal(t, "", formatToolResult(nil))

	// Complex type gets JSON serialized
	result := formatToolResult(map[string]string{"key": "value"})
	assert.Contains(t, result, "key")
	assert.Contains(t, result, "value")
}

func TestGetStringArg(t *testing.T) {
	args := map[string]interface{}{
		"str_arg": "hello",
		"int_arg": 42,
		"nil_arg": nil,
	}

	assert.Equal(t, "hello", getStringArg(args, "str_arg"))
	assert.Equal(t, "42", getStringArg(args, "int_arg"))
	assert.Equal(t, "", getStringArg(args, "missing"))
}

// ============================================================================
// Integration Test: Full Agent Loop via Executor
// ============================================================================

func TestExecutor_AgentStep_FullIntegration(t *testing.T) {
	// Mock LLM server that simulates a tool call loop
	var callCount int32
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if count == 1 {
			// First: ask to run a command
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "bash",
						Arguments: `{"command": "echo integration_test"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := mockLLMResponse("Integration test passed")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := testConfig(t)
	cfg.LLM = config.LLMConfig{
		LLMProviders: []config.LLMProvider{
			{
				Provider: "mock",
				BaseURL:  server.URL,
				Model:    "test-model",
			},
		},
		MaxTokens:  1000,
		MaxRetries: 1,
		Timeout:    "30s",
	}

	module := &core.Workflow{
		Name: "test-agent-integration",
		Kind: core.KindModule,
		Steps: []core.Step{
			{
				Name:          "agent-step",
				Type:          core.StepTypeAgent,
				Query:         "Run echo integration_test",
				MaxIterations: 5,
				AgentTools: []core.AgentToolDef{
					{Preset: "bash"},
				},
				Exports: map[string]string{
					"result": "{{agent_content}}",
				},
			},
			{
				Name:    "verify-step",
				Type:    core.StepTypeBash,
				Command: fmt.Sprintf("echo 'agent result: {{result}}'"),
			},
		},
	}

	executor := NewExecutor()
	executor.SetDryRun(false)
	executor.SetSpinner(false)

	ctx := context.Background()
	result, err := executor.ExecuteModule(ctx, module, map[string]string{
		"target": "test",
	}, cfg)

	require.NoError(t, err)
	assert.Equal(t, core.RunStatusCompleted, result.Status)
	assert.Len(t, result.Steps, 2)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[0].Status)
	assert.Equal(t, "Integration test passed", result.Steps[0].Output)
	assert.Equal(t, core.StepStatusSuccess, result.Steps[1].Status)
}

// ============================================================================
// Group 2: Planning Stage Tests
// ============================================================================

func TestAgentExecutor_PlanningStage(t *testing.T) {
	// Mock server: 1st call = planning response (no tools), 2nd call = final response
	var callCount int32
	var receivedMessages [][]ChatMessage

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)
		receivedMessages = append(receivedMessages, req.Messages)

		if count == 1 {
			// Planning phase: return plan text
			resp := mockLLMResponse("Step 1: Scan ports\nStep 2: Check services\nStep 3: Report findings")
			json.NewEncoder(w).Encode(resp)
		} else {
			// Execution phase: return final response
			resp := mockLLMResponse("Plan executed successfully")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "plan-agent",
		Type:          core.StepTypeAgent,
		PlanPrompt:    "Create a plan for scanning test.com",
		Query:         "Execute the plan",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Plan executed successfully", result.Output)

	// Verify agent_plan export is populated
	planContent, ok := result.Exports["agent_plan"]
	assert.True(t, ok)
	assert.Contains(t, planContent, "Step 1")

	// Verify planning request had the plan prompt
	require.Len(t, receivedMessages, 2)
	planMsgs := receivedMessages[0]
	assert.Equal(t, "user", planMsgs[len(planMsgs)-1].Role)
	assert.Equal(t, "Create a plan for scanning test.com", planMsgs[len(planMsgs)-1].Content)

	// Verify execution request includes the plan as assistant message
	execMsgs := receivedMessages[1]
	foundPlan := false
	for _, msg := range execMsgs {
		if msg.Role == "assistant" {
			if content, ok := msg.Content.(string); ok && strings.Contains(content, "Step 1") {
				foundPlan = true
			}
		}
	}
	assert.True(t, foundPlan, "execution messages should include the plan as assistant message")

	// Verify tokens are tracked from both planning and execution
	assert.Equal(t, 60, result.Exports["agent_total_tokens"]) // 30 from planning + 30 from execution
}

func TestAgentExecutor_PlanningStage_WithMaxTokens(t *testing.T) {
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)

		// Return plan or final based on whether tools are present
		if len(req.Tools) == 0 {
			// Planning request — verify max_tokens
			assert.Equal(t, 500, req.MaxTokens)
			resp := mockLLMResponse("Short plan")
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := mockLLMResponse("Done")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	planMaxTokens := 500
	step := &core.Step{
		Name:          "plan-tokens-agent",
		Type:          core.StepTypeAgent,
		PlanPrompt:    "Create a plan",
		PlanMaxTokens: &planMaxTokens,
		Query:         "Execute",
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
}

// ============================================================================
// Group 3: Multi-Goal Execution Tests
// ============================================================================

func TestAgentExecutor_MultiGoal(t *testing.T) {
	// Mock server: responds to each query with a unique response
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		resp := mockLLMResponse(fmt.Sprintf("Goal %d complete", count))
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "multi-goal-agent",
		Type:          core.StepTypeAgent,
		Queries:       []string{"List files", "Summarize findings", "Report"},
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	assert.Equal(t, core.StepStatusSuccess, result.Status)
	// Final content should be from the last goal
	assert.Equal(t, "Goal 3 complete", result.Output)

	// Verify agent_goal_results export
	goalResultsJSON, ok := result.Exports["agent_goal_results"]
	assert.True(t, ok, "agent_goal_results should be exported for multi-goal")
	assert.Contains(t, goalResultsJSON, "List files")
	assert.Contains(t, goalResultsJSON, "Summarize findings")
	assert.Contains(t, goalResultsJSON, "Report")

	// Tokens should be accumulated from all goals
	assert.Equal(t, 90, result.Exports["agent_total_tokens"]) // 30 * 3 goals
}

func TestAgentExecutor_QueryAndQueriesMutualExclusion(t *testing.T) {
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()
	cfg := newMockConfig(t, server.URL)

	executor := NewAgentExecutor(nil, nil)
	executor.SetConfig(cfg)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "both-query",
		Type:          core.StepTypeAgent,
		Query:         "single query",
		Queries:       []string{"query 1", "query 2"},
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot have both")
	assert.Equal(t, core.StepStatusFailed, result.Status)
}

func TestAgentExecutor_MultiGoal_SharedConversation(t *testing.T) {
	// Verify subsequent goals see messages from previous goals
	var receivedMessages [][]ChatMessage

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)
		receivedMessages = append(receivedMessages, req.Messages)

		resp := mockLLMResponse("response")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "shared-conv",
		Type:          core.StepTypeAgent,
		Queries:       []string{"First query", "Second query"},
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)

	// Second goal should have more messages (accumulated from first goal)
	require.Len(t, receivedMessages, 2)
	assert.True(t, len(receivedMessages[1]) > len(receivedMessages[0]),
		"second goal should see accumulated messages from first goal")
}

// ============================================================================
// Group 5: Model Fallback Tests
// ============================================================================

func TestAgentExecutor_ModelFallback(t *testing.T) {
	var receivedModels []string

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)
		receivedModels = append(receivedModels, req.Model)

		// Fail for first model, succeed for second
		if req.Model == "bad-model" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "model not found",
					"type":    "invalid_request_error",
				},
			})
			return
		}

		resp := mockLLMResponse("Fallback succeeded")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "fallback-agent",
		Type:          core.StepTypeAgent,
		Query:         "Test fallback",
		Models:        []string{"bad-model", "test-model"},
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Fallback succeeded", result.Output)

	// Verify both models were attempted
	assert.Contains(t, receivedModels, "bad-model")
	assert.Contains(t, receivedModels, "test-model")
}

// ============================================================================
// Group 6: Structured Output Tests
// ============================================================================

func TestAgentExecutor_StructuredOutput(t *testing.T) {
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)

		if count == 1 {
			// First call: agent completes with unstructured text
			resp := mockLLMResponse("Found some results")
			json.NewEncoder(w).Encode(resp)
		} else {
			// Second call: structured output request — verify response_format is set
			assert.NotNil(t, req.ResponseFormat, "structured output call should have response_format")
			resp := mockLLMResponse(`{"subdomains":[{"name":"sub.test.com","status":"active"}]}`)
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "structured-agent",
		Type:          core.StepTypeAgent,
		Query:         "Find subdomains",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		OutputSchema:  `{"type":"object","properties":{"subdomains":{"type":"array"}}}`,
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	// Final content should be the structured JSON
	assert.Contains(t, result.Output, "subdomains")
	assert.Contains(t, result.Output, "sub.test.com")
}

func TestAgentExecutor_StructuredOutput_AlreadyJSON(t *testing.T) {
	// If the agent already returns valid JSON, no extra call should be made
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMResponse(`{"results":["a","b"]}`)
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "already-json-agent",
		Type:          core.StepTypeAgent,
		Query:         "Return JSON",
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		OutputSchema:  `{"type":"object"}`,
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	// Should only have 1 LLM call (no extra structured output request)
	assert.Equal(t, 30, result.Exports["agent_total_tokens"])
}

// ============================================================================
// Group 7: Tool Tracing Hooks Tests
// ============================================================================

func TestAgentExecutor_TracingHooks(t *testing.T) {
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if count == 1 {
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "bash",
						Arguments: `{"command": "echo traced"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := mockLLMResponse("Traced complete")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	// Use hooks that call log_info (which is a registered function)
	step := &core.Step{
		Name:          "traced-agent",
		Type:          core.StepTypeAgent,
		Query:         "Run traced command",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		OnToolStart:   `log_info("HOOK_START: " + tool_name)`,
		OnToolEnd:     `log_info("HOOK_END: " + tool_name + " dur=" + duration)`,
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Traced complete", result.Output)
}

func TestAgentExecutor_TracingHooks_EmptyHooks(t *testing.T) {
	// Verify hooks are no-ops when empty strings
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMResponse("No hooks")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "no-hooks-agent",
		Type:          core.StepTypeAgent,
		Query:         "Test without hooks",
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		OnToolStart:   "",
		OnToolEnd:     "",
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
}

// ============================================================================
// Sub-Agent Orchestration Tests
// ============================================================================

func TestAgentExecutor_SubAgentSpawn(t *testing.T) {
	// Mock server:
	// 1st call (parent): returns spawn_agent tool call
	// 2nd call (child): returns child final response
	// 3rd call (parent): returns parent final response using child result
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		switch count {
		case 1:
			// Parent: spawn recon_agent
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_spawn",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "spawn_agent",
						Arguments: `{"agent":"recon_agent","query":"scan ports on target"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		case 2:
			// Child: return final result
			resp := mockLLMResponse("Found open ports: 80, 443, 8080")
			json.NewEncoder(w).Encode(resp)
		case 3:
			// Parent: final response using child result
			resp := mockLLMResponse("Recon complete. Open ports: 80, 443, 8080")
			json.NewEncoder(w).Encode(resp)
		default:
			resp := mockLLMResponse("unexpected call")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "orchestrator",
		Type:          core.StepTypeAgent,
		Query:         "Analyze target by coordinating specialists",
		SystemPrompt:  "You are an orchestrator. Delegate tasks to sub-agents.",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		SubAgents: []core.SubAgentDef{
			{
				Name:          "recon_agent",
				Description:   "Specialized agent for reconnaissance",
				SystemPrompt:  "You are a recon specialist",
				MaxIterations: 5,
				AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
			},
		},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Contains(t, result.Output, "Recon complete")
	// 3 LLM calls × 30 tokens each = 90
	assert.Equal(t, 90, result.Exports["agent_total_tokens"])
}

func TestAgentExecutor_SubAgentDepthLimit(t *testing.T) {
	// Mock server: parent calls spawn_agent which tries to exceed depth
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if count == 1 {
			// Parent: spawn sub-agent
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_spawn",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "spawn_agent",
						Arguments: `{"agent":"child","query":"do work"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := mockLLMResponse("Done with depth error")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "depth-limit-parent",
		Type:          core.StepTypeAgent,
		Query:         "Delegate work",
		MaxIterations: 5,
		MaxAgentDepth: 1, // Only 1 level allowed, but parent is at 0 so child at 1 is OK
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		SubAgents: []core.SubAgentDef{
			{
				Name:          "child",
				Description:   "Child agent",
				MaxIterations: 3,
				AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
				SubAgents: []core.SubAgentDef{
					{
						Name:          "grandchild",
						Description:   "Grandchild agent",
						MaxIterations: 3,
						AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
					},
				},
			},
		},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	// Parent should succeed even though the child depth limit was hit
	// The error is returned as a tool result to the parent LLM
	assert.Equal(t, core.StepStatusSuccess, result.Status)
}

func TestAgentExecutor_SubAgentFailure(t *testing.T) {
	// Mock server: child agent fails (e.g., config not available or bad response)
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		if count == 1 {
			// Parent: spawn sub-agent
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_spawn",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "spawn_agent",
						Arguments: `{"agent":"failing_agent","query":"do work"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else if count == 2 {
			// Child: return server error
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "internal server error",
					"type":    "server_error",
				},
			})
		} else {
			// Parent: handle error gracefully
			resp := mockLLMResponse("Sub-agent failed but I handled it")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "failure-parent",
		Type:          core.StepTypeAgent,
		Query:         "Delegate work",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		SubAgents: []core.SubAgentDef{
			{
				Name:          "failing_agent",
				Description:   "Agent that will fail",
				MaxIterations: 3,
				AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
			},
		},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	// Parent should succeed — child failure is returned as tool result string
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Contains(t, result.Output, "handled")
}

func TestAgentExecutor_SubAgentRecursive(t *testing.T) {
	// Test 3-level nesting: parent → child → grandchild
	// Mock server handles 5 calls:
	// 1. Parent spawns child
	// 2. Child spawns grandchild
	// 3. Grandchild returns result
	// 4. Child returns using grandchild result
	// 5. Parent returns using child result
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		switch count {
		case 1:
			// Parent: spawn child
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_1",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "spawn_agent",
						Arguments: `{"agent":"child","query":"intermediate task"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		case 2:
			// Child: spawn grandchild
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   "call_2",
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "spawn_agent",
						Arguments: `{"agent":"grandchild","query":"leaf task"}`,
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		case 3:
			// Grandchild: return result
			resp := mockLLMResponse("Grandchild result: data collected")
			json.NewEncoder(w).Encode(resp)
		case 4:
			// Child: return using grandchild data
			resp := mockLLMResponse("Child processed grandchild data")
			json.NewEncoder(w).Encode(resp)
		case 5:
			// Parent: final response
			resp := mockLLMResponse("Parent summarized: all levels complete")
			json.NewEncoder(w).Encode(resp)
		default:
			resp := mockLLMResponse("unexpected")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "recursive-parent",
		Type:          core.StepTypeAgent,
		Query:         "Coordinate 3-level task",
		MaxIterations: 5,
		MaxAgentDepth: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		SubAgents: []core.SubAgentDef{
			{
				Name:          "child",
				Description:   "Intermediate agent",
				MaxIterations: 5,
				AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
				SubAgents: []core.SubAgentDef{
					{
						Name:          "grandchild",
						Description:   "Leaf agent",
						MaxIterations: 5,
						AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
					},
				},
			},
		},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)

	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Contains(t, result.Output, "all levels complete")
	// 5 LLM calls × 30 tokens = 150
	assert.Equal(t, 150, result.Exports["agent_total_tokens"])
}

func TestAgentState_MergeTokens(t *testing.T) {
	state := &agentState{
		totalTokens:      100,
		promptTokens:     60,
		completionTokens: 40,
	}

	state.MergeTokens(30, 20, 10)
	assert.Equal(t, 130, state.totalTokens)
	assert.Equal(t, 80, state.promptTokens)
	assert.Equal(t, 50, state.completionTokens)
}

// ============================================================================
// Group 4: Conversation Compression Tests
// ============================================================================

func TestAgentExecutor_MessageWindowWithSummary(t *testing.T) {
	// Mock server: returns tool calls for several iterations, then final answer.
	// Also handles summarization request.
	var callCount int32

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&callCount, 1)
		w.Header().Set("Content-Type", "application/json")

		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)

		// Check if this is a summarization request (no tools)
		if len(req.Tools) == 0 && len(req.Messages) > 0 {
			lastMsg := req.Messages[len(req.Messages)-1]
			if content, ok := lastMsg.Content.(string); ok && strings.Contains(content, "Summarize") {
				resp := mockLLMResponse("Summary of earlier context")
				json.NewEncoder(w).Encode(resp)
				return
			}
		}

		if count <= 3 {
			// Return tool calls to build up conversation
			resp := mockLLMToolCallResponse([]core.LLMToolCall{
				{
					ID:   fmt.Sprintf("call_%d", count),
					Type: "function",
					Function: core.LLMToolCallFunction{
						Name:      "bash",
						Arguments: fmt.Sprintf(`{"command": "echo iter_%d"}`, count),
					},
				},
			})
			json.NewEncoder(w).Encode(resp)
		} else {
			resp := mockLLMResponse("Compression test complete")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "compress-agent",
		Type:          core.StepTypeAgent,
		Query:         "Run multiple commands",
		MaxIterations: 10,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
		Memory: &core.AgentMemoryConfig{
			MaxMessages:         5,
			SummarizeOnTruncate: true,
		},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Compression test complete", result.Output)
}
