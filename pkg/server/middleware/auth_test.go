package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name       string
		headerKey  string
		configKey  string
		wantStatus int
	}{
		{"valid key", "secret-key-123", "secret-key-123", fiber.StatusOK},
		{"missing header", "", "secret-key-123", fiber.StatusUnauthorized},
		{"whitespace only", "   ", "secret-key-123", fiber.StatusUnauthorized},
		{"null string", "null", "secret-key-123", fiber.StatusUnauthorized},
		{"NULL uppercase", "NULL", "secret-key-123", fiber.StatusUnauthorized},
		{"undefined string", "undefined", "secret-key-123", fiber.StatusUnauthorized},
		{"nil string", "nil", "secret-key-123", fiber.StatusUnauthorized},
		{"wrong key", "wrong-key", "secret-key-123", fiber.StatusUnauthorized},
		{"case mismatch", "Secret-Key-123", "secret-key-123", fiber.StatusUnauthorized},
		// Note: HTTP headers with leading/trailing whitespace are trimmed by the HTTP library
		{"leading whitespace trimmed by http", " secret-key-123", "secret-key-123", fiber.StatusOK},
		{"trailing whitespace trimmed by http", "secret-key-123 ", "secret-key-123", fiber.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Server: config.ServerConfig{
					EnabledAuthAPI: true,
					AuthAPIKey:     tt.configKey,
				},
			}

			app := fiber.New()
			app.Use(APIKeyAuth(cfg))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerKey != "" {
				req.Header.Set("x-osm-api-key", tt.headerKey)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestIsValidAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		provided string
		expected string
		want     bool
	}{
		{"exact match", "my-key", "my-key", true},
		{"empty provided", "", "my-key", false},
		{"whitespace only", "  ", "my-key", false},
		{"null lowercase", "null", "my-key", false},
		{"null uppercase", "NULL", "my-key", false},
		{"null mixed case", "Null", "my-key", false},
		{"undefined lowercase", "undefined", "my-key", false},
		{"undefined uppercase", "UNDEFINED", "my-key", false},
		{"nil lowercase", "nil", "my-key", false},
		{"nil uppercase", "NIL", "my-key", false},
		{"wrong key", "other", "my-key", false},
		{"case sensitive mismatch", "My-Key", "my-key", false},
		{"leading whitespace", " my-key", "my-key", false},
		{"trailing whitespace", "my-key ", "my-key", false},
		{"both have whitespace identical", " my-key ", " my-key ", true}, // exact match even with whitespace
		{"special characters", "my-key!@#$%", "my-key!@#$%", true},
		{"long key", "this-is-a-very-long-api-key-with-many-characters-1234567890", "this-is-a-very-long-api-key-with-many-characters-1234567890", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidAPIKey(tt.provided, tt.expected)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAPIKeyAuth_ResponseBody(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			EnabledAuthAPI: true,
			AuthAPIKey:     "test-key",
		},
	}

	app := fiber.New()
	app.Use(APIKeyAuth(cfg))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test that invalid key returns proper error response
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-osm-api-key", "wrong-key")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
}
