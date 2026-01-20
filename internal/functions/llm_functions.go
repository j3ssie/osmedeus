package functions

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/retry"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// llmFuncConfig holds resolved LLM configuration for function calls
type llmFuncConfig struct {
	BaseURL   string
	AuthToken string
	Model     string
}

// getLLMConfig resolves LLM config with environment variable overrides (highest priority)
func getLLMConfig() (*llmFuncConfig, error) {
	cfg := config.Get()
	result := &llmFuncConfig{}

	// Start with config file values
	if cfg != nil && len(cfg.LLM.LLMProviders) > 0 {
		provider := cfg.LLM.GetCurrentProvider()
		if provider != nil {
			result.BaseURL = provider.BaseURL
			result.AuthToken = provider.AuthToken
			result.Model = provider.Model
		}
	}

	// Environment overrides (highest priority)
	if v := os.Getenv("OSM_LLM_BASE_URL"); v != "" {
		result.BaseURL = v
	}
	if v := os.Getenv("OSM_LLM_AUTH_TOKEN"); v != "" {
		result.AuthToken = v
	}
	if v := os.Getenv("OSM_LLM_MODEL"); v != "" {
		result.Model = v
	}

	// Validate we have required configuration
	if result.BaseURL == "" {
		return nil, fmt.Errorf("LLM base URL not configured (set OSM_LLM_BASE_URL or configure llm_config in settings)")
	}

	return result, nil
}

// llmChatRequest represents an OpenAI-compatible chat completion request
type llmChatRequest struct {
	Model    string       `json:"model,omitempty"`
	Messages []llmMessage `json:"messages"`
}

// llmMessage represents a single message in the chat
type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// llmChatResponse represents an OpenAI-compatible chat completion response
type llmChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// sendLLMRequest sends a request to the LLM API with retry logic
func sendLLMRequest(llmCfg *llmFuncConfig, bodyBytes []byte) (string, error) {
	// Create request
	req, err := http.NewRequest("POST", llmCfg.BaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", core.DefaultUA)
	if llmCfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+llmCfg.AuthToken)
	}

	// Create client with timeout and TLS skip verify
	client := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Execute request with retry logic (3 attempts, exponential backoff)
	var resp *http.Response
	ctx := context.Background()
	retryCfg := retry.Config{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}

	err = retry.Do(ctx, retryCfg, func() error {
		var reqErr error
		// Need to recreate body for retries since it may have been consumed
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		resp, reqErr = client.Do(req)
		if reqErr != nil {
			return retry.Retryable(reqErr)
		}
		// Retry on server errors (5xx) and rate limits (429)
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			_ = resp.Body.Close()
			return retry.Retryable(fmt.Errorf("server error: %d", resp.StatusCode))
		}
		return nil
	})

	if err != nil {
		return "", fmt.Errorf("request failed after retries: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var llmResp llmChatResponse
	if err := json.Unmarshal(respBody, &llmResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error in response
	if llmResp.Error != nil {
		return "", fmt.Errorf("API error: %s", llmResp.Error.Message)
	}

	// Extract content from first choice
	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return llmResp.Choices[0].Message.Content, nil
}

// llmInvoke makes a simple LLM call with a single message
// Usage: llm_invoke(message) -> string
func (vf *vmFunc) llmInvoke(call goja.FunctionCall) goja.Value {
	message := call.Argument(0).String()
	logger.Get().Debug("Calling "+terminal.HiGreen(FnLLMInvoke), zap.Int("messageLength", len(message)))

	if message == "undefined" || message == "" {
		logger.Get().Warn(FnLLMInvoke + ": message is required")
		return vf.vm.ToValue("")
	}

	// Get LLM config
	llmCfg, err := getLLMConfig()
	if err != nil {
		logger.Get().Warn(FnLLMInvoke+": config error", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Build request body
	reqBody := llmChatRequest{
		Model: llmCfg.Model,
		Messages: []llmMessage{
			{Role: "user", Content: message},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		logger.Get().Warn(FnLLMInvoke+": failed to marshal request", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Send request
	content, err := sendLLMRequest(llmCfg, bodyBytes)
	if err != nil {
		logger.Get().Warn(FnLLMInvoke+": request failed", zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen(FnLLMInvoke)+" result", zap.Int("responseLength", len(content)))
	return vf.vm.ToValue(content)
}

// llmInvokeCustom makes an LLM call with a custom request body
// Usage: llm_invoke_custom(message, body_json) -> string
// The body_json can contain {{message}} placeholder that will be replaced with the message
func (vf *vmFunc) llmInvokeCustom(call goja.FunctionCall) goja.Value {
	message := call.Argument(0).String()
	bodyJSON := call.Argument(1).String()
	logger.Get().Debug("Calling "+terminal.HiGreen(FnLLMInvokeCustom), zap.Int("messageLength", len(message)), zap.Int("bodyLength", len(bodyJSON)))

	if message == "undefined" || message == "" {
		logger.Get().Warn(FnLLMInvokeCustom + ": message is required")
		return vf.vm.ToValue("")
	}

	if bodyJSON == "undefined" || bodyJSON == "" {
		logger.Get().Warn(FnLLMInvokeCustom + ": body_json is required")
		return vf.vm.ToValue("")
	}

	// Get LLM config
	llmCfg, err := getLLMConfig()
	if err != nil {
		logger.Get().Warn(FnLLMInvokeCustom+": config error", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Replace {{message}} placeholder in body
	// Escape the message for JSON string embedding
	escapedMessage, err := json.Marshal(message)
	if err != nil {
		logger.Get().Warn(FnLLMInvokeCustom+": failed to escape message", zap.Error(err))
		return vf.vm.ToValue("")
	}
	// Remove surrounding quotes from JSON encoding
	escapedMessageStr := string(escapedMessage[1 : len(escapedMessage)-1])
	bodyStr := strings.ReplaceAll(bodyJSON, "{{message}}", escapedMessageStr)

	// Validate JSON
	var bodyMap map[string]interface{}
	if err := json.Unmarshal([]byte(bodyStr), &bodyMap); err != nil {
		logger.Get().Warn(FnLLMInvokeCustom+": invalid JSON body", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Send request
	content, err := sendLLMRequest(llmCfg, []byte(bodyStr))
	if err != nil {
		logger.Get().Warn(FnLLMInvokeCustom+": request failed", zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen(FnLLMInvokeCustom)+" result", zap.Int("responseLength", len(content)))
	return vf.vm.ToValue(content)
}

// llmConversations makes an LLM call with multiple messages
// Usage: llm_conversations(msg1, msg2, ...) -> string
// Each message should be in format "role:content" where role is system, user, or assistant
func (vf *vmFunc) llmConversations(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling "+terminal.HiGreen(FnLLMConversations), zap.Int("argCount", len(call.Arguments)))

	if len(call.Arguments) == 0 {
		logger.Get().Warn(FnLLMConversations + ": at least one message is required")
		return vf.vm.ToValue("")
	}

	// Get LLM config
	llmCfg, err := getLLMConfig()
	if err != nil {
		logger.Get().Warn(FnLLMConversations+": config error", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Parse messages
	var messages []llmMessage
	for i, arg := range call.Arguments {
		msgStr := arg.String()
		if msgStr == "undefined" || msgStr == "" {
			continue
		}

		// Parse "role:content" format
		colonIdx := strings.Index(msgStr, ":")
		if colonIdx == -1 {
			logger.Get().Warn(FnLLMConversations+": invalid message format (expected 'role:content')",
				zap.Int("argIndex", i), zap.String("message", msgStr))
			return vf.vm.ToValue("")
		}

		role := strings.ToLower(strings.TrimSpace(msgStr[:colonIdx]))
		content := strings.TrimSpace(msgStr[colonIdx+1:])

		// Validate role
		validRoles := map[string]bool{"system": true, "user": true, "assistant": true}
		if !validRoles[role] {
			logger.Get().Warn(FnLLMConversations+": invalid role (expected system, user, or assistant)",
				zap.Int("argIndex", i), zap.String("role", role))
			return vf.vm.ToValue("")
		}

		messages = append(messages, llmMessage{
			Role:    role,
			Content: content,
		})
	}

	if len(messages) == 0 {
		logger.Get().Warn(FnLLMConversations + ": no valid messages provided")
		return vf.vm.ToValue("")
	}

	// Build request body
	reqBody := llmChatRequest{
		Model:    llmCfg.Model,
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		logger.Get().Warn(FnLLMConversations+": failed to marshal request", zap.Error(err))
		return vf.vm.ToValue("")
	}

	// Send request
	content, err := sendLLMRequest(llmCfg, bodyBytes)
	if err != nil {
		logger.Get().Warn(FnLLMConversations+": request failed", zap.Error(err))
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen(FnLLMConversations)+" result", zap.Int("responseLength", len(content)))
	return vf.vm.ToValue(content)
}
