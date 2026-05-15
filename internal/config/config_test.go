package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfig_GetServerURL(t *testing.T) {
	tests := []struct {
		name   string
		config ServerConfig
		want   string
	}{
		{
			name: "EventReceiverURL takes precedence",
			config: ServerConfig{
				EventReceiverURL: "http://custom.example.com:9000",
				Host:             "localhost",
				Port:             8002,
			},
			want: "http://custom.example.com:9000",
		},
		{
			name: "EventReceiverURL trailing slash removed",
			config: ServerConfig{
				EventReceiverURL: "http://custom.example.com:9000/",
			},
			want: "http://custom.example.com:9000",
		},
		{
			name: "Computed from Host and Port",
			config: ServerConfig{
				Host: "localhost",
				Port: 8002,
			},
			want: "http://localhost:8002",
		},
		{
			name: "0.0.0.0 converted to 127.0.0.1",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: 8002,
			},
			want: "http://127.0.0.1:8002",
		},
		{
			name: "Empty when no config",
			config: ServerConfig{
				Host: "",
				Port: 0,
			},
			want: "",
		},
		{
			name: "Empty when only host set",
			config: ServerConfig{
				Host: "localhost",
				Port: 0,
			},
			want: "",
		},
		{
			name: "Empty when only port set",
			config: ServerConfig{
				Host: "",
				Port: 8002,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetServerURL()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestServerConfig_GetEventReceiverURL(t *testing.T) {
	tests := []struct {
		name   string
		config ServerConfig
		want   string
	}{
		{
			name: "EventReceiverURL set",
			config: ServerConfig{
				EventReceiverURL: "http://custom.example.com:9000",
			},
			want: "http://custom.example.com:9000",
		},
		{
			name: "Computed from Host and Port",
			config: ServerConfig{
				Host: "localhost",
				Port: 8002,
			},
			want: "http://localhost:8002",
		},
		{
			name: "0.0.0.0 converted to 127.0.0.1",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: 8002,
			},
			want: "http://127.0.0.1:8002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetEventReceiverURL()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLLMConfigGetProviderByName(t *testing.T) {
	llmCfg := LLMConfig{
		LLMProviders: []LLMProvider{
			{Provider: "openai", BaseURL: "https://api.openai.com/v1/chat/completions"},
			{Provider: "atlas", BaseURL: "https://api.atlascloud.ai/v1/chat/completions"},
		},
	}

	provider := llmCfg.GetProviderByName("ATLAS")
	require.NotNil(t, provider)
	assert.Equal(t, "atlas", provider.Provider)
	assert.Equal(t, "https://api.atlascloud.ai/v1/chat/completions", provider.BaseURL)
	assert.Nil(t, llmCfg.GetProviderByName("missing"))
}
