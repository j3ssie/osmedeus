package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// LLMChatRequest represents a direct LLM chat completion request
type LLMChatRequest struct {
	Messages       []core.LLMMessage       `json:"messages"`
	Model          string                  `json:"model,omitempty"`
	MaxTokens      int                     `json:"max_tokens,omitempty"`
	Temperature    *float64                `json:"temperature,omitempty"`
	TopP           *float64                `json:"top_p,omitempty"`
	TopK           *int                    `json:"top_k,omitempty"`
	N              int                     `json:"n,omitempty"`
	Stream         bool                    `json:"stream,omitempty"`
	Tools          []core.LLMTool          `json:"tools,omitempty"`
	ToolChoice     interface{}             `json:"tool_choice,omitempty"`
	ResponseFormat *core.LLMResponseFormat `json:"response_format,omitempty"`
}

// LLMChatResponse represents the chat completion response
type LLMChatResponse struct {
	ID           string             `json:"id"`
	Model        string             `json:"model"`
	Content      interface{}        `json:"content"`
	FinishReason string             `json:"finish_reason"`
	ToolCalls    []core.LLMToolCall `json:"tool_calls,omitempty"`
	Usage        map[string]int     `json:"usage"`
}

// LLMEmbeddingRequest represents an embedding request
type LLMEmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model,omitempty"`
}

// LLMEmbeddingResponse represents the embedding response
type LLMEmbeddingResponse struct {
	Model      string         `json:"model"`
	Embeddings [][]float64    `json:"embeddings"`
	Usage      map[string]int `json:"usage"`
}

// Internal types for API communication
type llmChatAPIRequest struct {
	Model          string                  `json:"model"`
	Messages       []llmChatMessage        `json:"messages"`
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

type llmChatMessage struct {
	Role       string             `json:"role"`
	Content    interface{}        `json:"content"`
	Name       string             `json:"name,omitempty"`
	ToolCallID string             `json:"tool_call_id,omitempty"`
	ToolCalls  []core.LLMToolCall `json:"tool_calls,omitempty"`
}

type llmChatAPIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []llmAPIChoice `json:"choices"`
	Usage   llmAPIUsage    `json:"usage"`
	Error   *llmAPIError   `json:"error,omitempty"`
}

type llmAPIChoice struct {
	Index        int            `json:"index"`
	Message      llmChatMessage `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type llmAPIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type llmAPIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type llmEmbeddingAPIRequest struct {
	Model          string   `json:"model"`
	Input          []string `json:"input"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
}

type llmEmbeddingAPIResponse struct {
	Object string               `json:"object"`
	Data   []llmEmbeddingData   `json:"data"`
	Model  string               `json:"model"`
	Usage  llmEmbeddingAPIUsage `json:"usage"`
	Error  *llmAPIError         `json:"error,omitempty"`
}

type llmEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type llmEmbeddingAPIUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// LLMChat handles direct LLM chat completion requests
// @Summary LLM Chat Completion
// @Description Send a chat completion request to the configured LLM provider (OpenAI-compatible)
// @Tags LLM
// @Accept json
// @Produce json
// @Param request body LLMChatRequest true "Chat request"
// @Success 200 {object} LLMChatResponse "Chat response"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "LLM error"
// @Security BearerAuth
// @Router /osm/api/llm/v1/chat/completions [post]
func LLMChat(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LLMChatRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body: " + err.Error(),
			})
		}

		// Validate messages
		if len(req.Messages) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Messages field is required",
			})
		}

		// Validate LLM configuration
		if cfg.LLM.GetProviderCount() == 0 {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "No LLM providers configured",
			})
		}

		// Get provider
		provider := cfg.LLM.GetCurrentProvider()
		if provider == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "No LLM provider available",
			})
		}

		// Build API request
		apiReq := buildLLMChatAPIRequest(&req, cfg, provider)

		// Execute request with retry
		ctx, cancel := context.WithTimeout(c.Context(), 120*time.Second)
		defer cancel()

		response, err := executeLLMChatRequest(ctx, cfg, provider, apiReq)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "LLM request failed: " + err.Error(),
			})
		}

		if response.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("LLM API error: %s (%s)", response.Error.Message, response.Error.Type),
			})
		}

		// Build response
		result := &LLMChatResponse{
			ID:    response.ID,
			Model: response.Model,
			Usage: map[string]int{
				"prompt_tokens":     response.Usage.PromptTokens,
				"completion_tokens": response.Usage.CompletionTokens,
				"total_tokens":      response.Usage.TotalTokens,
			},
		}

		if len(response.Choices) > 0 {
			choice := response.Choices[0]
			result.Content = choice.Message.Content
			result.FinishReason = choice.FinishReason
			if len(choice.Message.ToolCalls) > 0 {
				result.ToolCalls = choice.Message.ToolCalls
			}
		}

		return c.JSON(result)
	}
}

// LLMEmbedding handles embedding generation requests
// @Summary Generate Embeddings
// @Description Generate embeddings for input text using the configured LLM provider
// @Tags LLM
// @Accept json
// @Produce json
// @Param request body LLMEmbeddingRequest true "Embedding request"
// @Success 200 {object} LLMEmbeddingResponse "Embedding response"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "LLM error"
// @Security BearerAuth
// @Router /osm/api/llm/v1/embeddings [post]
func LLMEmbedding(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LLMEmbeddingRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body: " + err.Error(),
			})
		}

		// Validate input
		if len(req.Input) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Input field is required",
			})
		}

		// Validate LLM configuration
		if cfg.LLM.GetProviderCount() == 0 {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "No LLM providers configured",
			})
		}

		// Get provider
		provider := cfg.LLM.GetCurrentProvider()
		if provider == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "No LLM provider available",
			})
		}

		// Build API request
		apiReq := &llmEmbeddingAPIRequest{
			Model: req.Model,
			Input: req.Input,
		}
		if apiReq.Model == "" {
			apiReq.Model = provider.Model
		}

		// Execute request
		ctx, cancel := context.WithTimeout(c.Context(), 120*time.Second)
		defer cancel()

		response, err := executeLLMEmbeddingRequest(ctx, cfg, provider, apiReq)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Embedding request failed: " + err.Error(),
			})
		}

		if response.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("Embedding API error: %s (%s)", response.Error.Message, response.Error.Type),
			})
		}

		// Build response
		embeddings := make([][]float64, len(response.Data))
		for i, d := range response.Data {
			embeddings[i] = d.Embedding
		}

		result := &LLMEmbeddingResponse{
			Model:      response.Model,
			Embeddings: embeddings,
			Usage: map[string]int{
				"prompt_tokens": response.Usage.PromptTokens,
				"total_tokens":  response.Usage.TotalTokens,
			},
		}

		return c.JSON(result)
	}
}

// buildLLMChatAPIRequest builds the API request from handler request
func buildLLMChatAPIRequest(req *LLMChatRequest, cfg *config.Config, provider *config.LLMProvider) *llmChatAPIRequest {
	apiReq := &llmChatAPIRequest{
		Model:          req.Model,
		MaxTokens:      req.MaxTokens,
		N:              req.N,
		Stream:         req.Stream,
		Tools:          req.Tools,
		ToolChoice:     req.ToolChoice,
		ResponseFormat: req.ResponseFormat,
	}

	// Use defaults from config if not specified
	if apiReq.Model == "" {
		apiReq.Model = provider.Model
	}
	if apiReq.MaxTokens == 0 {
		apiReq.MaxTokens = cfg.LLM.MaxTokens
	}

	// Temperature - use request value, config default, or 0.7
	if req.Temperature != nil {
		apiReq.Temperature = *req.Temperature
	} else if cfg.LLM.Temperature > 0 {
		apiReq.Temperature = cfg.LLM.Temperature
	} else {
		apiReq.Temperature = 0.7
	}

	// TopP
	if req.TopP != nil {
		apiReq.TopP = *req.TopP
	} else {
		apiReq.TopP = cfg.LLM.TopP
	}

	// TopK
	if req.TopK != nil {
		apiReq.TopK = *req.TopK
	} else {
		apiReq.TopK = cfg.LLM.TopK
	}

	// Convert messages
	messages := make([]llmChatMessage, 0, len(req.Messages)+1)

	// Auto-prepend system prompt if configured and no system message exists
	if cfg.LLM.SystemPrompt != "" {
		hasSystem := false
		for _, msg := range req.Messages {
			if msg.Role == core.LLMRoleSystem {
				hasSystem = true
				break
			}
		}
		if !hasSystem {
			messages = append(messages, llmChatMessage{
				Role:    string(core.LLMRoleSystem),
				Content: cfg.LLM.SystemPrompt,
			})
		}
	}

	for _, msg := range req.Messages {
		messages = append(messages, llmChatMessage{
			Role:       string(msg.Role),
			Content:    msg.Content,
			Name:       msg.Name,
			ToolCallID: msg.ToolCallID,
			ToolCalls:  msg.ToolCalls,
		})
	}

	apiReq.Messages = messages

	return apiReq
}

// executeLLMChatRequest executes an HTTP request to the LLM provider
func executeLLMChatRequest(ctx context.Context, cfg *config.Config, provider *config.LLMProvider, apiReq *llmChatAPIRequest) (*llmChatAPIResponse, error) {
	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if provider.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+provider.AuthToken)
	}

	// Parse custom headers
	if cfg.LLM.CustomHeaders != "" {
		for _, h := range strings.Split(cfg.LLM.CustomHeaders, ",") {
			if parts := strings.SplitN(strings.TrimSpace(h), ":", 2); len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response llmChatAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
	}

	if resp.StatusCode >= 400 {
		if response.Error != nil {
			return &response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, response.Error.Message)
		}
		return &response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return &response, nil
}

// executeLLMEmbeddingRequest executes an embedding request to the LLM provider
func executeLLMEmbeddingRequest(ctx context.Context, cfg *config.Config, provider *config.LLMProvider, apiReq *llmEmbeddingAPIRequest) (*llmEmbeddingAPIResponse, error) {
	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Determine embedding endpoint
	embeddingURL := provider.BaseURL
	if strings.HasSuffix(embeddingURL, "/chat/completions") {
		embeddingURL = strings.Replace(embeddingURL, "/chat/completions", "/embeddings", 1)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", embeddingURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if provider.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+provider.AuthToken)
	}

	// Parse custom headers
	if cfg.LLM.CustomHeaders != "" {
		for _, h := range strings.Split(cfg.LLM.CustomHeaders, ",") {
			if parts := strings.SplitN(strings.TrimSpace(h), ":", 2); len(parts) == 2 {
				req.Header.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response llmEmbeddingAPIResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w (body: %s)", err, string(respBody))
	}

	if resp.StatusCode >= 400 {
		if response.Error != nil {
			return &response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, response.Error.Message)
		}
		return &response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return &response, nil
}
