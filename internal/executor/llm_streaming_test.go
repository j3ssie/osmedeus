package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeSSEChunk writes a single SSE data line
func writeSSEChunk(w http.ResponseWriter, data string) {
	fmt.Fprintf(w, "data: %s\n\n", data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// makeStreamChunk creates a JSON string for a streaming chunk with content
func makeStreamChunk(id, model, content string) string {
	chunk := ChatCompletionStreamChunk{
		ID:    id,
		Model: model,
		Choices: []StreamChunkChoice{
			{
				Index: 0,
				Delta: ChatDelta{
					Content: content,
				},
			},
		},
	}
	b, _ := json.Marshal(chunk)
	return string(b)
}

// makeStreamChunkWithRole creates a streaming chunk with role
func makeStreamChunkWithRole(id, model, role string) string {
	chunk := ChatCompletionStreamChunk{
		ID:    id,
		Model: model,
		Choices: []StreamChunkChoice{
			{
				Index: 0,
				Delta: ChatDelta{
					Role: role,
				},
			},
		},
	}
	b, _ := json.Marshal(chunk)
	return string(b)
}

// makeStreamChunkDone creates a final chunk with finish_reason and optional usage
func makeStreamChunkDone(id, model string, usage *ChatUsage) string {
	chunk := ChatCompletionStreamChunk{
		ID:    id,
		Model: model,
		Choices: []StreamChunkChoice{
			{
				Index:        0,
				Delta:        ChatDelta{},
				FinishReason: "stop",
			},
		},
		Usage: usage,
	}
	b, _ := json.Marshal(chunk)
	return string(b)
}

// makeStreamToolCallChunk creates a streaming chunk with a tool call delta
func makeStreamToolCallChunk(id, model string, tcIndex int, tcID, tcType, funcName, funcArgs string) string {
	tc := StreamToolCall{
		Index: tcIndex,
	}
	if tcID != "" {
		tc.ID = tcID
	}
	if tcType != "" {
		tc.Type = tcType
	}
	if funcName != "" {
		tc.Function.Name = funcName
	}
	if funcArgs != "" {
		tc.Function.Arguments = funcArgs
	}

	chunk := ChatCompletionStreamChunk{
		ID:    id,
		Model: model,
		Choices: []StreamChunkChoice{
			{
				Index: 0,
				Delta: ChatDelta{
					ToolCalls: []StreamToolCall{tc},
				},
			},
		},
	}
	b, _ := json.Marshal(chunk)
	return string(b)
}

func newStreamingMockServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

// ============================================================================
// Streaming Tests
// ============================================================================

func TestStreamingSSEParsing(t *testing.T) {
	// Mock server returns SSE events with content tokens
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		writeSSEChunk(w, makeStreamChunkWithRole("chat-1", "test-model", "assistant"))
		writeSSEChunk(w, makeStreamChunk("chat-1", "test-model", "Hello"))
		writeSSEChunk(w, makeStreamChunk("chat-1", "test-model", " world"))
		writeSSEChunk(w, makeStreamChunk("chat-1", "test-model", "!"))
		writeSSEChunk(w, makeStreamChunkDone("chat-1", "test-model", &ChatUsage{
			PromptTokens:     10,
			CompletionTokens: 5,
			TotalTokens:      15,
		}))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{
		config: &config.Config{},
		silent: true,
	}

	ctx := context.Background()
	provider := &config.LLMProvider{
		BaseURL:   server.URL,
		AuthToken: "test-token",
		Model:     "test-model",
	}
	request := &ChatCompletionRequest{
		Model:    "test-model",
		Messages: []ChatMessage{{Role: "user", Content: "Hi"}},
		Stream:   true,
	}
	llmConfig := &MergedLLMConfig{Timeout: "30s"}

	var tokens []string
	response, err := executor.sendChatRequestStreaming(ctx, provider, request, llmConfig, func(token string) {
		tokens = append(tokens, token)
	})

	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify accumulated content
	require.Len(t, response.Choices, 1)
	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	assert.Equal(t, "Hello world!", content)

	// Verify metadata
	assert.Equal(t, "chat-1", response.ID)
	assert.Equal(t, "test-model", response.Model)
	assert.Equal(t, "assistant", response.Choices[0].Message.Role)
	assert.Equal(t, "stop", response.Choices[0].FinishReason)

	// Verify usage
	assert.Equal(t, 10, response.Usage.PromptTokens)
	assert.Equal(t, 5, response.Usage.CompletionTokens)
	assert.Equal(t, 15, response.Usage.TotalTokens)
}

func TestStreamingTokenCallback(t *testing.T) {
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		writeSSEChunk(w, makeStreamChunk("c1", "m1", "token1"))
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "token2"))
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "token3"))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	var tokens []string
	response, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		func(token string) { tokens = append(tokens, token) },
	)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify callback was called for each content delta + trailing newline
	assert.Equal(t, []string{"token1", "token2", "token3", "\n"}, tokens)
}

func TestStreamingToolCallAccumulation(t *testing.T) {
	// Simulate streaming tool calls: name arrives in one chunk, arguments split across chunks
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// First tool call: ID and name
		writeSSEChunk(w, makeStreamToolCallChunk("c1", "m1", 0, "call_1", "function", "bash", ""))
		// Arguments arrive in parts
		writeSSEChunk(w, makeStreamToolCallChunk("c1", "m1", 0, "", "", "", `{"comma`))
		writeSSEChunk(w, makeStreamToolCallChunk("c1", "m1", 0, "", "", "", `nd": "echo hi"}`))

		// Second tool call in same response
		writeSSEChunk(w, makeStreamToolCallChunk("c1", "m1", 1, "call_2", "function", "read_file", ""))
		writeSSEChunk(w, makeStreamToolCallChunk("c1", "m1", 1, "", "", "", `{"path": "/tmp/test"}`))

		// Finish
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	response, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.NoError(t, err)
	require.NotNil(t, response)
	require.Len(t, response.Choices, 1)

	toolCalls := response.Choices[0].Message.ToolCalls
	require.Len(t, toolCalls, 2)

	// First tool call
	assert.Equal(t, "call_1", toolCalls[0].ID)
	assert.Equal(t, "function", toolCalls[0].Type)
	assert.Equal(t, "bash", toolCalls[0].Function.Name)
	assert.Equal(t, `{"command": "echo hi"}`, toolCalls[0].Function.Arguments)

	// Second tool call
	assert.Equal(t, "call_2", toolCalls[1].ID)
	assert.Equal(t, "read_file", toolCalls[1].Function.Name)
	assert.Equal(t, `{"path": "/tmp/test"}`, toolCalls[1].Function.Arguments)
}

func TestStreamingDoneSignal(t *testing.T) {
	// Verify data: [DONE] terminates the stream cleanly
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		writeSSEChunk(w, makeStreamChunk("c1", "m1", "before done"))
		writeSSEChunk(w, "[DONE]")
		// Anything after [DONE] should be ignored
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "SHOULD NOT APPEAR"))
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	response, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.NoError(t, err)
	require.NotNil(t, response)

	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	assert.Equal(t, "before done", content)
	assert.NotContains(t, content, "SHOULD NOT APPEAR")
}

func TestStreamingFallbackOnError(t *testing.T) {
	// Server returns HTTP 500 — should error gracefully
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "server overloaded",
				"type":    "server_error",
				"code":    "500",
			},
		})
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	_, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "server overloaded")
}

func TestStreamingErrorInChunk(t *testing.T) {
	// Server streams normally then sends an error chunk
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		writeSSEChunk(w, makeStreamChunk("c1", "m1", "partial "))

		// Error chunk
		errorChunk := ChatCompletionStreamChunk{
			Error: &ChatError{
				Message: "context length exceeded",
				Type:    "invalid_request_error",
				Code:    "context_length_exceeded",
			},
		}
		b, _ := json.Marshal(errorChunk)
		writeSSEChunk(w, string(b))
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	_, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "context length exceeded")
}

func TestStreamingMalformedSSE(t *testing.T) {
	// Server sends some malformed data — should skip and continue
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		writeSSEChunk(w, makeStreamChunk("c1", "m1", "good"))
		writeSSEChunk(w, "{invalid json")
		writeSSEChunk(w, makeStreamChunk("c1", "m1", " data"))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	response, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.NoError(t, err)
	require.NotNil(t, response)

	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	assert.Equal(t, "good data", content)
}

func TestAgentExecutor_StreamFlagInRequest(t *testing.T) {
	// Verify that Stream is passed through to the LLM request when llmConfig.Stream is true
	var receivedStream bool

	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)
		receivedStream = req.Stream

		// Respond with non-streaming format since the mock doesn't do SSE
		// (the test captures what was sent, not how it responds)
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSEChunk(w, makeStreamChunkWithRole("c1", "m1", "assistant"))
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "streamed response"))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", &ChatUsage{
			PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8,
		}))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	cfg.LLM.Stream = true

	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "stream-agent",
		Type:          core.StepTypeAgent,
		Query:         "Test streaming",
		MaxIterations: 3,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.True(t, receivedStream, "Stream should be true in the LLM request")
	assert.Equal(t, "streamed response", result.Output)
}

func TestStepLevelStreamOverride(t *testing.T) {
	// Verify that step.Stream overrides global config
	t.Run("step stream true overrides global false", func(t *testing.T) {
		cfg := &config.Config{
			LLM: config.LLMConfig{
				Stream:     false,
				MaxTokens:  100,
				MaxRetries: 1,
				Timeout:    "30s",
			},
		}
		executor := &LLMExecutor{config: cfg}
		streamTrue := true
		step := &core.Step{
			Name:   "test",
			Stream: &streamTrue,
		}

		merged := executor.getMergedConfig(step)
		assert.True(t, merged.Stream, "step.Stream=true should override global false")
	})

	t.Run("step stream false overrides global true", func(t *testing.T) {
		cfg := &config.Config{
			LLM: config.LLMConfig{
				Stream:     true,
				MaxTokens:  100,
				MaxRetries: 1,
				Timeout:    "30s",
			},
		}
		executor := &LLMExecutor{config: cfg}
		streamFalse := false
		step := &core.Step{
			Name:   "test",
			Stream: &streamFalse,
		}

		merged := executor.getMergedConfig(step)
		assert.False(t, merged.Stream, "step.Stream=false should override global true")
	})

	t.Run("step stream nil inherits global", func(t *testing.T) {
		cfg := &config.Config{
			LLM: config.LLMConfig{
				Stream:     true,
				MaxTokens:  100,
				MaxRetries: 1,
				Timeout:    "30s",
			},
		}
		executor := &LLMExecutor{config: cfg}
		step := &core.Step{
			Name: "test",
		}

		merged := executor.getMergedConfig(step)
		assert.True(t, merged.Stream, "nil step.Stream should inherit global true")
	})

	t.Run("step stream overrides llm_config stream", func(t *testing.T) {
		cfg := &config.Config{
			LLM: config.LLMConfig{
				Stream:     false,
				MaxTokens:  100,
				MaxRetries: 1,
				Timeout:    "30s",
			},
		}
		executor := &LLMExecutor{config: cfg}
		llmConfigStream := false
		stepStream := true
		step := &core.Step{
			Name: "test",
			LLMConfig: &core.LLMStepConfig{
				Stream: &llmConfigStream,
			},
			Stream: &stepStream,
		}

		merged := executor.getMergedConfig(step)
		assert.True(t, merged.Stream, "step.Stream should override llm_config.Stream")
	})
}

func TestStreamingDispatchFromSendChatRequest(t *testing.T) {
	// Verify that sendChatRequest dispatches to streaming when request.Stream is true
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		writeSSEChunk(w, makeStreamChunkWithRole("c1", "m1", "assistant"))
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "streamed via dispatch"))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", &ChatUsage{
			PromptTokens: 5, CompletionTokens: 3, TotalTokens: 8,
		}))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{
		config: &config.Config{},
		silent: true,
	}

	ctx := context.Background()
	provider := &config.LLMProvider{BaseURL: server.URL, Model: "m1"}
	request := &ChatCompletionRequest{
		Model:    "m1",
		Messages: []ChatMessage{{Role: "user", Content: "test"}},
		Stream:   true,
	}
	llmConfig := &MergedLLMConfig{Timeout: "30s", Stream: true}

	response, err := executor.sendChatRequest(ctx, provider, request, llmConfig)
	require.NoError(t, err)
	require.NotNil(t, response)

	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	assert.Equal(t, "streamed via dispatch", content)
	assert.Equal(t, 8, response.Usage.TotalTokens)
}

func TestStreamingEmptyContent(t *testing.T) {
	// Stream that has no content tokens (e.g., only tool calls or empty response)
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		writeSSEChunk(w, makeStreamChunkWithRole("c1", "m1", "assistant"))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	var tokens []string
	response, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		func(token string) { tokens = append(tokens, token) },
	)

	require.NoError(t, err)
	require.NotNil(t, response)

	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	assert.Equal(t, "", content)

	// No content tokens, so no callback should have been called (no trailing newline either)
	assert.Empty(t, tokens)
}

func TestStreamingSSEWithComments(t *testing.T) {
	// SSE spec allows comment lines starting with ":" — these should be ignored
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// SSE comment (keep-alive)
		fmt.Fprintf(w, ": this is a comment\n\n")
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "with comments"))
		fmt.Fprintf(w, ": another comment\n\n")
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	response, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.NoError(t, err)
	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	assert.Equal(t, "with comments", content)
}

func TestStreamingLLMStepIntegration(t *testing.T) {
	// Test the full LLM step path with streaming via sendChatRequest dispatch
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)

		if req.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			writeSSEChunk(w, makeStreamChunkWithRole("c1", "m1", "assistant"))
			writeSSEChunk(w, makeStreamChunk("c1", "m1", "Streaming "))
			writeSSEChunk(w, makeStreamChunk("c1", "m1", "response"))
			writeSSEChunk(w, makeStreamChunkDone("c1", "m1", &ChatUsage{
				PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15,
			}))
			writeSSEChunk(w, "[DONE]")
		} else {
			// Fallback non-streaming response
			w.Header().Set("Content-Type", "application/json")
			resp := mockLLMResponse("Non-streaming response")
			json.NewEncoder(w).Encode(resp)
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	cfg.LLM.Stream = true

	dispatcher := NewStepDispatcher()
	llmExec := NewLLMExecutor(dispatcher.GetTemplateEngine())
	llmExec.SetConfig(cfg)
	llmExec.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name: "streaming-llm",
		Type: core.StepTypeLLM,
		Messages: []core.LLMMessage{
			{Role: core.LLMRoleUser, Content: "Hello"},
		},
	}

	result, err := llmExec.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Streaming response", result.Output)
}

func TestStreamingAuthHeader(t *testing.T) {
	// Verify auth headers are sent in streaming requests
	var receivedAuth string

	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "ok"))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	_, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, AuthToken: "secret-key", Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "Bearer secret-key", receivedAuth)
}

func TestStreamingCustomHeaders(t *testing.T) {
	// Verify custom headers are sent in streaming requests
	var receivedHeaders http.Header

	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "ok"))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	_, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{
			Timeout: "30s",
			CustomHeaders: map[string]string{
				"X-Custom": "value",
			},
		},
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "value", receivedHeaders.Get("X-Custom"))
	assert.Equal(t, "text/event-stream", receivedHeaders.Get("Accept"))
}

func TestStreamingNonStreamingFallback(t *testing.T) {
	// When Stream is false, sendChatRequest should NOT use streaming
	server := newMockLLMServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)

		assert.False(t, req.Stream, "stream should be false in request")

		w.Header().Set("Content-Type", "application/json")
		resp := mockLLMResponse("non-streaming")
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	response, err := executor.sendChatRequest(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: false},
		&MergedLLMConfig{Timeout: "30s"},
	)

	require.NoError(t, err)
	require.NotNil(t, response)

	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	assert.Equal(t, "non-streaming", content)
}

func TestStreamingContentWithSpecialChars(t *testing.T) {
	// Test streaming with content containing special characters, newlines, etc.
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		writeSSEChunk(w, makeStreamChunk("c1", "m1", "Hello\n"))
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "- bullet point\n"))
		writeSSEChunk(w, makeStreamChunk("c1", "m1", `{"json": "value"}`))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	response, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.NoError(t, err)
	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	expected := "Hello\n- bullet point\n" + `{"json": "value"}`
	assert.Equal(t, expected, content)
}

func TestStreamingAcceptHeader(t *testing.T) {
	// Verify that streaming requests include Accept: text/event-stream
	var receivedAccept string

	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "text/event-stream")
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "ok"))
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	_, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.NoError(t, err)
	assert.Equal(t, "text/event-stream", receivedAccept)
}

func TestStreamingEmptyDataLines(t *testing.T) {
	// SSE with blank lines and event: fields (should be ignored, only data: parsed)
	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		// Various non-data lines
		fmt.Fprintf(w, "event: message\n")
		fmt.Fprintf(w, "id: 1\n")
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "content"))
		fmt.Fprintf(w, "\n") // blank line
		writeSSEChunk(w, makeStreamChunkDone("c1", "m1", nil))
		writeSSEChunk(w, "[DONE]")
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	response, err := executor.sendChatRequestStreaming(
		context.Background(),
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	require.NoError(t, err)
	content, ok := response.Choices[0].Message.Content.(string)
	require.True(t, ok)
	assert.Equal(t, "content", content)
}

func TestStreamingLLMExecProcess_SkipsPrint(t *testing.T) {
	// Verify that processChatResponse with streamed=true does not call printLLMOutput
	// (We can't easily assert fmt.Print wasn't called, but we verify the Output is set correctly)

	executor := &LLMExecutor{config: &config.Config{}, silent: false}
	result := &core.StepResult{
		Exports: make(map[string]interface{}),
	}

	response := &ChatCompletionResponse{
		ID:    "test",
		Model: "m1",
		Choices: []ChatChoice{
			{
				Message: ChatMessage{
					Role:    "assistant",
					Content: "test content",
				},
				FinishReason: "stop",
			},
		},
	}

	// With streamed=true, should not print (we just verify it doesn't panic)
	executor.processChatResponse(result, "test-step", response, true)
	assert.Equal(t, "test content", result.Output)

	// Exports should still be populated
	_, ok := result.Exports["test_step_llm_resp"]
	assert.True(t, ok, "exports should be populated even when streamed")
	_, ok = result.Exports["test_step_content"]
	assert.True(t, ok, "content export should be populated")
}

func TestStreamingAgentWithToolCalls(t *testing.T) {
	// Test agent executor with streaming enabled, doing tool call + final response
	var callCount int32

	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req ChatCompletionRequest
		json.Unmarshal(body, &req)

		callCount++
		w.Header().Set("Content-Type", "text/event-stream")

		if callCount == 1 {
			// First call: stream tool call
			writeSSEChunk(w, makeStreamChunkWithRole("c1", "m1", "assistant"))
			writeSSEChunk(w, makeStreamToolCallChunk("c1", "m1", 0, "call_1", "function", "bash", `{"command": "echo test"}`))
			finishChunk := ChatCompletionStreamChunk{
				ID:    "c1",
				Model: "m1",
				Choices: []StreamChunkChoice{
					{Index: 0, Delta: ChatDelta{}, FinishReason: "tool_calls"},
				},
				Usage: &ChatUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			}
			b, _ := json.Marshal(finishChunk)
			writeSSEChunk(w, string(b))
			writeSSEChunk(w, "[DONE]")
		} else {
			// Second call: stream final response
			writeSSEChunk(w, makeStreamChunkWithRole("c1", "m1", "assistant"))
			writeSSEChunk(w, makeStreamChunk("c1", "m1", "Tool executed "))
			writeSSEChunk(w, makeStreamChunk("c1", "m1", "successfully"))
			writeSSEChunk(w, makeStreamChunkDone("c1", "m1", &ChatUsage{
				PromptTokens: 15, CompletionTokens: 5, TotalTokens: 20,
			}))
			writeSSEChunk(w, "[DONE]")
		}
	})
	defer server.Close()

	cfg := newMockConfig(t, server.URL)
	cfg.LLM.Stream = true

	dispatcher := NewStepDispatcher()
	executor := NewAgentExecutor(dispatcher.GetTemplateEngine(), dispatcher.GetFunctionRegistry())
	executor.SetConfig(cfg)
	executor.SetSilent(true)

	ctx := context.Background()
	execCtx := core.NewExecutionContext("test", core.KindModule, "run-1", "test.com")

	step := &core.Step{
		Name:          "streaming-tool-agent",
		Type:          core.StepTypeAgent,
		Query:         "Run echo test",
		MaxIterations: 5,
		AgentTools:    []core.AgentToolDef{{Preset: "bash"}},
	}

	result, err := executor.Execute(ctx, step, execCtx)
	require.NoError(t, err)
	assert.Equal(t, core.StepStatusSuccess, result.Status)
	assert.Equal(t, "Tool executed successfully", result.Output)
	assert.Equal(t, 2, result.Exports["agent_iterations"])

	// Verify tokens accumulated from both calls
	totalTokens, ok := result.Exports["agent_total_tokens"].(int)
	require.True(t, ok)
	assert.Equal(t, 35, totalTokens) // 15 + 20
}

func TestStreamingRespectsContext(t *testing.T) {
	// Verify that context cancellation is respected during streaming
	var _ = strings.NewReader // ensure strings import is used

	server := newStreamingMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		// Send one chunk then block (simulating slow stream)
		writeSSEChunk(w, makeStreamChunk("c1", "m1", "partial"))
		// The server will naturally end when the connection is closed
	})
	defer server.Close()

	executor := &LLMExecutor{config: &config.Config{}, silent: true}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := executor.sendChatRequestStreaming(
		ctx,
		&config.LLMProvider{BaseURL: server.URL, Model: "m1"},
		&ChatCompletionRequest{Model: "m1", Messages: []ChatMessage{{Role: "user", Content: "test"}}, Stream: true},
		&MergedLLMConfig{Timeout: "30s"},
		nil,
	)

	// Should error due to cancelled context
	assert.Error(t, err)
}
