package executor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/metrics"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// LLMExecutor executes LLM steps
type LLMExecutor struct {
	templateEngine template.TemplateEngine
	client         *http.Client
	config         *config.Config
	silent         bool
}

// NewLLMExecutor creates a new LLM executor
func NewLLMExecutor(engine template.TemplateEngine) *LLMExecutor {
	return &LLMExecutor{
		templateEngine: engine,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// Name returns the executor name for logging/debugging
func (e *LLMExecutor) Name() string {
	return "llm"
}

// StepTypes returns the step types this executor handles
func (e *LLMExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeLLM}
}

// SetConfig sets the application config for LLM settings
func (e *LLMExecutor) SetConfig(cfg *config.Config) {
	e.config = cfg
}

// SetSilent enables or disables silent mode (suppresses output)
func (e *LLMExecutor) SetSilent(s bool) {
	e.silent = s
}

// CanHandle returns true if this executor can handle the given step type
func (e *LLMExecutor) CanHandle(stepType core.StepType) bool {
	return stepType == core.StepTypeLLM
}

// MergedLLMConfig holds the final merged configuration
type MergedLLMConfig struct {
	Model          string
	MaxTokens      int
	Temperature    float64
	TopK           int
	TopP           float64
	N              int
	Timeout        string
	MaxRetries     int
	Stream         bool
	ResponseFormat *core.LLMResponseFormat
	CustomHeaders  map[string]string
	SystemPrompt   string
}

// ChatCompletionRequest is the OpenAI-compatible request format
type ChatCompletionRequest struct {
	Model          string                  `json:"model"`
	Messages       []ChatMessage           `json:"messages"`
	MaxTokens      int                     `json:"max_tokens,omitempty"`
	Temperature    float64                 `json:"temperature,omitempty"`
	TopP           float64                 `json:"top_p,omitempty"`
	TopK           int                     `json:"top_k,omitempty"`
	N              int                     `json:"n,omitempty"`
	Stream         bool                    `json:"stream,omitempty"`
	Tools          []core.LLMTool          `json:"tools,omitempty"`
	ToolChoice     interface{}             `json:"tool_choice,omitempty"`
	ResponseFormat *core.LLMResponseFormat `json:"response_format,omitempty"`
}

// ChatMessage is the wire format for messages
type ChatMessage struct {
	Role       string             `json:"role"`
	Content    interface{}        `json:"content"` // string or []ContentPart
	Name       string             `json:"name,omitempty"`
	ToolCallID string             `json:"tool_call_id,omitempty"`
	ToolCalls  []core.LLMToolCall `json:"tool_calls,omitempty"`
}

// ChatCompletionResponse is the OpenAI-compatible response format
type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   ChatUsage    `json:"usage"`
	Error   *ChatError   `json:"error,omitempty"`
}

// ChatChoice represents a single choice in the response
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage represents token usage in the response
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatError represents an error in the API response
type ChatError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// ChatCompletionStreamChunk is a single SSE event in a streaming response
type ChatCompletionStreamChunk struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []StreamChunkChoice `json:"choices"`
	Usage   *ChatUsage          `json:"usage,omitempty"`
	Error   *ChatError          `json:"error,omitempty"`
}

// StreamChunkChoice is a single choice in a streaming chunk
type StreamChunkChoice struct {
	Index        int       `json:"index"`
	Delta        ChatDelta `json:"delta"`
	FinishReason string    `json:"finish_reason,omitempty"`
}

// ChatDelta contains incremental content from streaming
type ChatDelta struct {
	Role      string           `json:"role,omitempty"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []StreamToolCall `json:"tool_calls,omitempty"`
}

// StreamToolCall is a partial tool call from streaming (index-based accumulation)
type StreamToolCall struct {
	Index    int                    `json:"index"`
	ID       string                 `json:"id,omitempty"`
	Type     string                 `json:"type,omitempty"`
	Function StreamToolCallFunction `json:"function,omitempty"`
}

// StreamToolCallFunction contains partial function data from streaming
type StreamToolCallFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

// EmbeddingRequest represents a request for embeddings
type EmbeddingRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

// EmbeddingResponse represents the response from embeddings API
type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *ChatError `json:"error,omitempty"`
}

// EmbeddingData represents a single embedding in the response
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// Execute executes an LLM step
func (e *LLMExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	log := logger.Get()
	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
		Exports:   make(map[string]interface{}),
	}

	// Validate config is set
	if e.config == nil {
		err := fmt.Errorf("LLM executor config not set")
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Get merged LLM configuration
	llmConfig := e.getMergedConfig(step)

	// Validate required fields
	if len(step.Messages) == 0 && len(step.EmbeddingInput) == 0 {
		err := fmt.Errorf("LLM step '%s' requires 'messages' or 'embedding_input' field", step.Name)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	log.Debug("Executing LLM step",
		zap.String("step", step.Name),
		zap.Bool("is_embedding", step.IsEmbedding),
		zap.Int("messages_count", len(step.Messages)),
	)

	// Handle embedding vs chat completion
	if step.IsEmbedding || len(step.EmbeddingInput) > 0 {
		return e.executeEmbedding(ctx, step, execCtx, result, llmConfig)
	}

	return e.executeChatCompletion(ctx, step, execCtx, result, llmConfig)
}

// executeChatCompletion executes a chat completion request with provider rotation
func (e *LLMExecutor) executeChatCompletion(
	ctx context.Context,
	step *core.Step,
	execCtx *core.ExecutionContext,
	result *core.StepResult,
	llmConfig *MergedLLMConfig,
) (*core.StepResult, error) {
	log := logger.Get()

	// Build request
	request, err := e.buildChatRequest(step, llmConfig)
	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Execute with retry and provider rotation
	var response *ChatCompletionResponse
	var lastErr error

	maxRetries := llmConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	providerCount := e.config.LLM.GetProviderCount()
	if providerCount == 0 {
		err := fmt.Errorf("no LLM providers configured")
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	totalAttempts := maxRetries * providerCount

	for attempt := 0; attempt < totalAttempts; attempt++ {
		provider := e.config.LLM.GetCurrentProvider()
		if provider == nil {
			lastErr = fmt.Errorf("no LLM providers available")
			break
		}

		// Update model from provider if not overridden
		if llmConfig.Model == "" {
			request.Model = provider.Model
		}

		log.Debug("Attempting LLM request",
			zap.String("provider", provider.Provider),
			zap.String("model", request.Model),
			zap.Int("attempt", attempt+1),
			zap.Int("max_attempts", totalAttempts),
		)

		response, lastErr = e.sendChatRequest(ctx, provider, request, llmConfig)

		if lastErr == nil && response.Error == nil {
			break // Success
		}

		// Check if we should rotate provider
		if isProviderError(lastErr) || isRateLimitError(response) {
			// Record rate limit hit for metrics
			if isRateLimitError(response) {
				metrics.RecordRateLimitHit(provider.Provider, "llm")
			}
			log.Warn("Provider error, rotating",
				zap.String("provider", provider.Provider),
				zap.Error(lastErr),
			)
			e.config.LLM.RotateProvider()
		}

		// Small backoff before retry
		if attempt < totalAttempts-1 {
			select {
			case <-ctx.Done():
				result.Status = core.StepStatusFailed
				result.Error = ctx.Err()
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
				return result, ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 500 * time.Millisecond):
			}
		}
	}

	if lastErr != nil {
		result.Status = core.StepStatusFailed
		result.Error = lastErr
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, lastErr
	}

	if response != nil && response.Error != nil {
		err := fmt.Errorf("LLM API error: %s (%s)", response.Error.Message, response.Error.Type)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Process response and exports (skip print if streamed — output was already printed token-by-token)
	e.processChatResponse(result, step.Name, response, llmConfig.Stream)

	result.Status = core.StepStatusSuccess
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// executeEmbedding executes an embedding request
func (e *LLMExecutor) executeEmbedding(
	ctx context.Context,
	step *core.Step,
	execCtx *core.ExecutionContext,
	result *core.StepResult,
	llmConfig *MergedLLMConfig,
) (*core.StepResult, error) {
	log := logger.Get()

	if len(step.EmbeddingInput) == 0 {
		err := fmt.Errorf("embedding step '%s' requires 'embedding_input' field", step.Name)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Build embedding request
	request := &EmbeddingRequest{
		Model: llmConfig.Model,
		Input: step.EmbeddingInput,
	}

	// Execute with retry and provider rotation
	var response *EmbeddingResponse
	var lastErr error

	maxRetries := llmConfig.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	providerCount := e.config.LLM.GetProviderCount()
	if providerCount == 0 {
		err := fmt.Errorf("no LLM providers configured")
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	totalAttempts := maxRetries * providerCount

	for attempt := 0; attempt < totalAttempts; attempt++ {
		provider := e.config.LLM.GetCurrentProvider()
		if provider == nil {
			lastErr = fmt.Errorf("no LLM providers available")
			break
		}

		// Update model from provider if not overridden
		if request.Model == "" {
			request.Model = provider.Model
		}

		log.Debug("Attempting embedding request",
			zap.String("provider", provider.Provider),
			zap.String("model", request.Model),
			zap.Int("attempt", attempt+1),
		)

		response, lastErr = e.sendEmbeddingRequest(ctx, provider, request, llmConfig)

		if lastErr == nil && response.Error == nil {
			break // Success
		}

		// Check if we should rotate provider
		if isProviderError(lastErr) || (response != nil && response.Error != nil) {
			log.Warn("Provider error, rotating",
				zap.String("provider", provider.Provider),
				zap.Error(lastErr),
			)
			e.config.LLM.RotateProvider()
		}
	}

	if lastErr != nil {
		result.Status = core.StepStatusFailed
		result.Error = lastErr
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, lastErr
	}

	if response != nil && response.Error != nil {
		err := fmt.Errorf("embedding API error: %s (%s)", response.Error.Message, response.Error.Type)
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Process embedding response
	e.processEmbeddingResponse(result, step.Name, response)

	result.Status = core.StepStatusSuccess
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// buildChatRequest builds an OpenAI-compatible chat request
func (e *LLMExecutor) buildChatRequest(step *core.Step, llmConfig *MergedLLMConfig) (*ChatCompletionRequest, error) {
	request := &ChatCompletionRequest{
		Model:       llmConfig.Model,
		MaxTokens:   llmConfig.MaxTokens,
		Temperature: llmConfig.Temperature,
		TopP:        llmConfig.TopP,
		TopK:        llmConfig.TopK,
		N:           llmConfig.N,
		Stream:      llmConfig.Stream,
	}

	// Convert messages
	messages := make([]ChatMessage, 0, len(step.Messages)+1)

	// Auto-prepend system prompt if global one exists and step doesn't have one
	if llmConfig.SystemPrompt != "" {
		hasSystemMessage := false
		for _, msg := range step.Messages {
			if msg.Role == core.LLMRoleSystem {
				hasSystemMessage = true
				break
			}
		}
		if !hasSystemMessage {
			messages = append(messages, ChatMessage{
				Role:    string(core.LLMRoleSystem),
				Content: llmConfig.SystemPrompt,
			})
		}
	}

	// Add step messages
	for _, msg := range step.Messages {
		chatMsg := ChatMessage{
			Role:       string(msg.Role),
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallID,
			ToolCalls:  msg.ToolCalls,
		}
		messages = append(messages, chatMsg)
	}

	request.Messages = messages

	// Add tools if specified
	if len(step.Tools) > 0 {
		request.Tools = step.Tools
	}

	// Add tool choice if specified
	if step.ToolChoice != nil {
		request.ToolChoice = step.ToolChoice
	}

	// Add response format if specified
	if llmConfig.ResponseFormat != nil {
		request.ResponseFormat = llmConfig.ResponseFormat
	}

	return request, nil
}

// sendChatRequest sends an HTTP request to the LLM provider.
// When request.Stream is true, it delegates to sendChatRequestStreaming for SSE handling.
func (e *LLMExecutor) sendChatRequest(
	ctx context.Context,
	provider *config.LLMProvider,
	request *ChatCompletionRequest,
	llmConfig *MergedLLMConfig,
) (*ChatCompletionResponse, error) {
	if request.Stream {
		return e.sendChatRequestStreaming(ctx, provider, request, llmConfig, func(token string) {
			if !e.silent {
				fmt.Print(token)
			}
		})
	}

	// Marshal request to JSON
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	if provider.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+provider.AuthToken)
	}

	// Add custom headers
	for key, value := range llmConfig.CustomHeaders {
		req.Header.Set(key, value)
	}

	// Set timeout
	timeout, err := time.ParseDuration(llmConfig.Timeout)
	if err != nil {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response ChatCompletionResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		if response.Error != nil {
			return &response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, response.Error.Message)
		}
		return &response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return &response, nil
}

// sendChatRequestStreaming handles SSE streaming responses from the LLM provider.
// It reads the stream line-by-line, accumulates content and tool calls, and calls
// onToken for each text chunk for real-time display.
func (e *LLMExecutor) sendChatRequestStreaming(
	ctx context.Context,
	provider *config.LLMProvider,
	request *ChatCompletionRequest,
	llmConfig *MergedLLMConfig,
	onToken func(token string),
) (*ChatCompletionResponse, error) {
	log := logger.Get()

	// Marshal request to JSON
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	if provider.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+provider.AuthToken)
	}

	// Add custom headers
	for key, value := range llmConfig.CustomHeaders {
		req.Header.Set(key, value)
	}

	// Set timeout
	timeout, err := time.ParseDuration(llmConfig.Timeout)
	if err != nil {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check for HTTP errors before reading the stream
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		var errResp ChatCompletionResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != nil {
			return &errResp, fmt.Errorf("HTTP %d: %s", resp.StatusCode, errResp.Error.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse SSE stream
	var contentBuilder strings.Builder
	var role string
	var finishReason string
	var responseID, responseModel string
	var usage ChatUsage
	// Accumulate tool calls by index
	toolCallMap := make(map[int]*core.LLMToolCall)

	scanner := bufio.NewScanner(resp.Body)
	// Increase buffer size for potentially large SSE lines
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// SSE format: lines starting with "data: "
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Stream terminator
		if data == "[DONE]" {
			break
		}

		var chunk ChatCompletionStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			log.Debug("Failed to parse streaming chunk, skipping",
				zap.String("data", data),
				zap.Error(err),
			)
			continue
		}

		// Check for error in chunk
		if chunk.Error != nil {
			return &ChatCompletionResponse{Error: chunk.Error},
				fmt.Errorf("streaming error: %s", chunk.Error.Message)
		}

		// Capture metadata from first chunk
		if responseID == "" && chunk.ID != "" {
			responseID = chunk.ID
		}
		if responseModel == "" && chunk.Model != "" {
			responseModel = chunk.Model
		}

		// Capture usage from final chunk
		if chunk.Usage != nil {
			usage = *chunk.Usage
		}

		// Process choices
		for _, choice := range chunk.Choices {
			// Capture role from first delta
			if choice.Delta.Role != "" {
				role = choice.Delta.Role
			}

			// Accumulate content
			if choice.Delta.Content != "" {
				contentBuilder.WriteString(choice.Delta.Content)
				if onToken != nil {
					onToken(choice.Delta.Content)
				}
			}

			// Accumulate tool calls (index-based)
			for _, tc := range choice.Delta.ToolCalls {
				existing, ok := toolCallMap[tc.Index]
				if !ok {
					existing = &core.LLMToolCall{
						Type: "function",
					}
					toolCallMap[tc.Index] = existing
				}
				if tc.ID != "" {
					existing.ID = tc.ID
				}
				if tc.Type != "" {
					existing.Type = tc.Type
				}
				if tc.Function.Name != "" {
					existing.Function.Name = tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					existing.Function.Arguments += tc.Function.Arguments
				}
			}

			// Capture finish reason
			if choice.FinishReason != "" {
				finishReason = choice.FinishReason
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %w", err)
	}

	// Print trailing newline after streaming output
	if onToken != nil && contentBuilder.Len() > 0 {
		onToken("\n")
	}

	// Build tool calls slice from map (ordered by index)
	var toolCalls []core.LLMToolCall
	if len(toolCallMap) > 0 {
		toolCalls = make([]core.LLMToolCall, len(toolCallMap))
		for idx, tc := range toolCallMap {
			if idx < len(toolCalls) {
				toolCalls[idx] = *tc
			}
		}
	}

	// Assemble final response
	if role == "" {
		role = "assistant"
	}

	var msgContent interface{} = contentBuilder.String()

	response := &ChatCompletionResponse{
		ID:    responseID,
		Model: responseModel,
		Choices: []ChatChoice{
			{
				Index: 0,
				Message: ChatMessage{
					Role:      role,
					Content:   msgContent,
					ToolCalls: toolCalls,
				},
				FinishReason: finishReason,
			},
		},
		Usage: usage,
	}

	return response, nil
}

// sendEmbeddingRequest sends an embedding request to the LLM provider
func (e *LLMExecutor) sendEmbeddingRequest(
	ctx context.Context,
	provider *config.LLMProvider,
	request *EmbeddingRequest,
	llmConfig *MergedLLMConfig,
) (*EmbeddingResponse, error) {
	// Marshal request to JSON
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Determine embedding endpoint - typically /v1/embeddings
	embeddingURL := provider.BaseURL
	if strings.HasSuffix(embeddingURL, "/chat/completions") {
		embeddingURL = strings.Replace(embeddingURL, "/chat/completions", "/embeddings", 1)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", embeddingURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	if provider.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+provider.AuthToken)
	}

	// Add custom headers
	for key, value := range llmConfig.CustomHeaders {
		req.Header.Set(key, value)
	}

	// Set timeout
	timeout, err := time.ParseDuration(llmConfig.Timeout)
	if err != nil {
		timeout = 120 * time.Second
	}
	client := &http.Client{Timeout: timeout}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response EmbeddingResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		if response.Error != nil {
			return &response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, response.Error.Message)
		}
		return &response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return &response, nil
}

// printLLMOutput prints LLM response with glamour markdown rendering
func printLLMOutput(content string) {
	// Render with glamour for markdown highlighting
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(120),
	)

	var rendered string
	if err == nil {
		if out, renderErr := renderer.Render(content); renderErr == nil {
			rendered = out
		} else {
			rendered = content + "\n"
		}
	} else {
		rendered = content + "\n"
	}

	fmt.Print(rendered)
}

// processChatResponse exports the LLM response to step result
func (e *LLMExecutor) processChatResponse(result *core.StepResult, stepName string, response *ChatCompletionResponse, streamed bool) {
	exportKey := sanitizeStepName(stepName) + "_llm_resp"

	// Build comprehensive export structure
	llmResp := map[string]interface{}{
		"id":      response.ID,
		"model":   response.Model,
		"created": response.Created,
		"usage": map[string]interface{}{
			"prompt_tokens":     response.Usage.PromptTokens,
			"completion_tokens": response.Usage.CompletionTokens,
			"total_tokens":      response.Usage.TotalTokens,
		},
	}

	// Export choices
	if len(response.Choices) > 0 {
		choice := response.Choices[0]
		llmResp["content"] = choice.Message.Content
		llmResp["finish_reason"] = choice.FinishReason
		llmResp["role"] = choice.Message.Role

		// Export tool calls if present
		if len(choice.Message.ToolCalls) > 0 {
			llmResp["tool_calls"] = choice.Message.ToolCalls
		}

		// Set output to content for display
		if content, ok := choice.Message.Content.(string); ok {
			result.Output = content
			// Print LLM output with markdown formatting (skip if silent or already streamed)
			if !e.silent && !streamed {
				printLLMOutput(content)
			}
		}
	}

	// All choices for n > 1
	if len(response.Choices) > 1 {
		allContents := make([]interface{}, len(response.Choices))
		for i, c := range response.Choices {
			allContents[i] = c.Message.Content
		}
		llmResp["all_contents"] = allContents
	}

	result.Exports[exportKey] = llmResp

	// Also export content directly for easy access
	if len(response.Choices) > 0 {
		contentKey := sanitizeStepName(stepName) + "_content"
		result.Exports[contentKey] = response.Choices[0].Message.Content
	}
}

// processEmbeddingResponse exports the embedding response to step result
func (e *LLMExecutor) processEmbeddingResponse(result *core.StepResult, stepName string, response *EmbeddingResponse) {
	exportKey := sanitizeStepName(stepName) + "_llm_resp"

	// Build export structure
	llmResp := map[string]interface{}{
		"model": response.Model,
		"usage": map[string]interface{}{
			"prompt_tokens": response.Usage.PromptTokens,
			"total_tokens":  response.Usage.TotalTokens,
		},
	}

	// Export embeddings
	if len(response.Data) > 0 {
		embeddings := make([][]float64, len(response.Data))
		for i, d := range response.Data {
			embeddings[i] = d.Embedding
		}
		llmResp["embeddings"] = embeddings

		// Set output to summary
		result.Output = fmt.Sprintf("Generated %d embeddings", len(embeddings))
	}

	result.Exports[exportKey] = llmResp
}

// getMergedConfig merges global llm_config with step-level overrides
func (e *LLMExecutor) getMergedConfig(step *core.Step) *MergedLLMConfig {
	globalLLM := &e.config.LLM

	merged := &MergedLLMConfig{
		MaxTokens:     globalLLM.MaxTokens,
		Temperature:   globalLLM.Temperature,
		TopK:          globalLLM.TopK,
		TopP:          globalLLM.TopP,
		N:             globalLLM.N,
		Timeout:       globalLLM.Timeout,
		MaxRetries:    globalLLM.MaxRetries,
		Stream:        globalLLM.Stream,
		SystemPrompt:  globalLLM.SystemPrompt,
		CustomHeaders: make(map[string]string),
	}

	// Set default response format if structured JSON is enabled globally
	if globalLLM.StructuredJSONFormat {
		merged.ResponseFormat = &core.LLMResponseFormat{
			Type: "json_object",
		}
	}

	// Parse global custom headers (format: "Key1: Value1, Key2: Value2")
	if globalLLM.CustomHeaders != "" {
		for _, h := range strings.Split(globalLLM.CustomHeaders, ",") {
			if parts := strings.SplitN(strings.TrimSpace(h), ":", 2); len(parts) == 2 {
				merged.CustomHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Apply step-level overrides
	if step.LLMConfig != nil {
		cfg := step.LLMConfig

		if cfg.Model != "" {
			merged.Model = cfg.Model
		}
		if cfg.MaxTokens != nil {
			merged.MaxTokens = *cfg.MaxTokens
		}
		if cfg.Temperature != nil {
			merged.Temperature = *cfg.Temperature
		}
		if cfg.TopK != nil {
			merged.TopK = *cfg.TopK
		}
		if cfg.TopP != nil {
			merged.TopP = *cfg.TopP
		}
		if cfg.N != nil {
			merged.N = *cfg.N
		}
		if cfg.Timeout != "" {
			merged.Timeout = cfg.Timeout
		}
		if cfg.MaxRetries != nil {
			merged.MaxRetries = *cfg.MaxRetries
		}
		if cfg.Stream != nil {
			merged.Stream = *cfg.Stream
		}
		if cfg.ResponseFormat != nil {
			merged.ResponseFormat = cfg.ResponseFormat
		}

		// Merge custom headers (step overrides global)
		for k, v := range cfg.CustomHeaders {
			merged.CustomHeaders[k] = v
		}
	}

	// Apply extra LLM parameters (these can override anything)
	if step.ExtraLLMParams != nil {
		if model, ok := step.ExtraLLMParams["model"].(string); ok {
			merged.Model = model
		}
		if maxTokens, ok := step.ExtraLLMParams["max_tokens"].(int); ok {
			merged.MaxTokens = maxTokens
		}
		if temp, ok := step.ExtraLLMParams["temperature"].(float64); ok {
			merged.Temperature = temp
		}
		if topK, ok := step.ExtraLLMParams["top_k"].(int); ok {
			merged.TopK = topK
		}
		if topP, ok := step.ExtraLLMParams["top_p"].(float64); ok {
			merged.TopP = topP
		}
	}

	// Top-level step.Stream overrides everything (highest precedence)
	if step.Stream != nil {
		merged.Stream = *step.Stream
	}

	return merged
}

// isProviderError checks if error indicates provider-level failure
func isProviderError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "i/o timeout")
}

// isRateLimitError checks if response indicates rate limiting
func isRateLimitError(resp *ChatCompletionResponse) bool {
	if resp == nil || resp.Error == nil {
		return false
	}
	return resp.Error.Type == "rate_limit_error" ||
		strings.Contains(resp.Error.Code, "rate_limit") ||
		strings.Contains(resp.Error.Message, "rate limit") ||
		strings.Contains(resp.Error.Message, "Rate limit")
}
