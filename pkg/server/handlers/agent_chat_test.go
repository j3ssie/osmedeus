package handlers

import (
	"testing"
)

func TestBuildAgentPrompt(t *testing.T) {
	tests := []struct {
		name     string
		messages []AgentChatMessage
		want     string
	}{
		{
			name:     "empty messages",
			messages: []AgentChatMessage{},
			want:     "",
		},
		{
			name: "single user message",
			messages: []AgentChatMessage{
				{Role: "user", Content: "Hello"},
			},
			want: "user: Hello",
		},
		{
			name: "multiple messages",
			messages: []AgentChatMessage{
				{Role: "system", Content: "You are helpful."},
				{Role: "user", Content: "Hello"},
			},
			want: "system: You are helpful.\nuser: Hello",
		},
		{
			name: "skips empty content",
			messages: []AgentChatMessage{
				{Role: "user", Content: "Hello"},
				{Role: "assistant", Content: ""},
				{Role: "user", Content: "World"},
			},
			want: "user: Hello\nuser: World",
		},
		{
			name: "empty role",
			messages: []AgentChatMessage{
				{Role: "", Content: "No role here"},
			},
			want: "No role here",
		},
		{
			name: "whitespace-only content",
			messages: []AgentChatMessage{
				{Role: "user", Content: "   "},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildAgentPrompt(tt.messages)
			if got != tt.want {
				t.Errorf("buildAgentPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveAgentName(t *testing.T) {
	tests := []struct {
		name  string
		model string
		want  string
	}{
		{
			name:  "empty defaults to claude-code",
			model: "",
			want:  "claude-code",
		},
		{
			name:  "claude-code stays claude-code",
			model: "claude-code",
			want:  "claude-code",
		},
		{
			name:  "codex resolves correctly",
			model: "codex",
			want:  "codex",
		},
		{
			name:  "unknown model defaults to claude-code",
			model: "gpt-4o",
			want:  "claude-code",
		},
		{
			name:  "opencode resolves correctly",
			model: "opencode",
			want:  "opencode",
		},
		{
			name:  "gemini resolves correctly",
			model: "gemini",
			want:  "gemini",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveAgentName(tt.model)
			if got != tt.want {
				t.Errorf("resolveAgentName(%q) = %q, want %q", tt.model, got, tt.want)
			}
		})
	}
}
