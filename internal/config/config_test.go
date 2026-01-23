package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			name: "0.0.0.0 NOT converted (uses raw host)",
			config: ServerConfig{
				Host: "0.0.0.0",
				Port: 8002,
			},
			want: "http://0.0.0.0:8002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.GetEventReceiverURL()
			assert.Equal(t, tt.want, got)
		})
	}
}
