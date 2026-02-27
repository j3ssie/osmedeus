package handlers

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
)

// AgentChatRequest represents an OpenAI-compatible chat completion request for ACP agents.
type AgentChatRequest struct {
	Messages []AgentChatMessage `json:"messages"`
	Model    string             `json:"model,omitempty"` // maps to agent name (default: "claude-code")
}

// AgentChatMessage represents a single message in a chat conversation.
type AgentChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AgentChatResponse represents an OpenAI-compatible chat completion response.
type AgentChatResponse struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Model   string        `json:"model"`
	Choices []AgentChoice `json:"choices"`
}

// AgentChoice represents a single choice in the chat completion response.
type AgentChoice struct {
	Index        int              `json:"index"`
	Message      AgentChatMessage `json:"message"`
	FinishReason string           `json:"finish_reason"`
}

// Concurrency guard: only one ACP agent subprocess at a time.
var (
	agentMu      sync.Mutex
	agentRunning bool
)

// AgentChat handles ACP agent chat completion requests.
// @Summary Agent Chat Completion
// @Description Spawn a local ACP agent and return its output in OpenAI-compatible format
// @Tags Agent
// @Accept json
// @Produce json
// @Param request body AgentChatRequest true "Chat request"
// @Success 200 {object} AgentChatResponse "Chat response"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 409 {object} map[string]interface{} "Agent already running"
// @Failure 500 {object} map[string]interface{} "Agent error"
// @Security BearerAuth
// @Router /osm/api/agent/chat/completions [post]
func AgentChat(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req AgentChatRequest
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
				"message": "Messages field is required and must not be empty",
			})
		}

		// Build prompt by joining all message contents (prefixed with role)
		prompt := buildAgentPrompt(req.Messages)
		if prompt == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Messages must contain at least one non-empty content",
			})
		}

		// Resolve agent name from model field
		agentName := resolveAgentName(req.Model)

		// Acquire concurrency lock
		agentMu.Lock()
		if agentRunning {
			agentMu.Unlock()
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   true,
				"message": "An agent is already running. Only one agent subprocess is allowed at a time.",
			})
		}
		agentRunning = true
		agentMu.Unlock()

		defer func() {
			agentMu.Lock()
			agentRunning = false
			agentMu.Unlock()
		}()

		// Create context with 10-minute timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		// Run ACP agent
		output, _, err := executor.RunAgentACP(ctx, prompt, agentName, nil)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Agent execution failed: " + err.Error(),
			})
		}

		// Build OpenAI-compatible response
		resp := AgentChatResponse{
			ID:      fmt.Sprintf("agent-%d", time.Now().UnixNano()),
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   agentName,
			Choices: []AgentChoice{
				{
					Index: 0,
					Message: AgentChatMessage{
						Role:    "assistant",
						Content: output,
					},
					FinishReason: "stop",
				},
			},
		}

		return c.JSON(resp)
	}
}

// buildAgentPrompt joins chat messages into a single prompt string.
func buildAgentPrompt(messages []AgentChatMessage) string {
	var parts []string
	for _, msg := range messages {
		content := strings.TrimSpace(msg.Content)
		if content == "" {
			continue
		}
		if msg.Role != "" {
			parts = append(parts, msg.Role+": "+content)
		} else {
			parts = append(parts, content)
		}
	}
	return strings.Join(parts, "\n")
}

// resolveAgentName maps the model field to a known agent name.
// Falls back to "claude-code" for unknown values.
func resolveAgentName(model string) string {
	if model == "" {
		return "claude-code"
	}
	for _, name := range executor.ListAgentNames() {
		if name == model {
			return model
		}
	}
	return "claude-code"
}
