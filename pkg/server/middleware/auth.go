package middleware

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/j3ssie/osmedeus/v5/internal/config"
)

// Claims represents JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTAuth creates JWT authentication middleware
func JWTAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string

		// Try Authorization header first (API clients)
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// Fall back to cookie (browser clients)
		if tokenString == "" {
			cookie := c.Cookies("osmedeus_session")
			if cookie != "" {
				// URL decode the cookie value
				decoded, err := url.QueryUnescape(cookie)
				if err == nil {
					// Try to parse as JSON to extract token
					var sessionData map[string]string
					if json.Unmarshal([]byte(decoded), &sessionData) == nil {
						if t, ok := sessionData["token"]; ok && t != "" {
							tokenString = t
						}
					}
				}
				// Fallback: if not JSON, treat as raw token (backward compat)
				if tokenString == "" {
					tokenString = cookie
				}
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Missing authorization header or session cookie",
			})
		}

		// Parse and validate token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Server.JWT.SecretSigningKey), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid or expired token",
			})
		}

		// Store claims in context
		c.Locals("user", claims)

		return c.Next()
	}
}

// GenerateToken generates a JWT token
func GenerateToken(username string, cfg *config.Config) (string, error) {
	expiration := time.Duration(cfg.Server.JWT.ExpirationMinutes) * time.Minute

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Server.JWT.SecretSigningKey))
}

// GetUser gets the current user from context
func GetUser(c *fiber.Ctx) *Claims {
	claims, ok := c.Locals("user").(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// APIKeyAuth creates API key authentication middleware
func APIKeyAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("x-osm-api-key")

		if !isValidAPIKey(apiKey, cfg.Server.AuthAPIKey) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid or missing API key",
			})
		}

		return c.Next()
	}
}

// isValidAPIKey validates the provided API key against expected value
func isValidAPIKey(provided, expected string) bool {
	// Reject empty or whitespace-only keys
	trimmed := strings.TrimSpace(provided)
	if trimmed == "" {
		return false
	}

	// Reject suspicious placeholder values
	lower := strings.ToLower(trimmed)
	if lower == "null" || lower == "undefined" || lower == "nil" {
		return false
	}

	// Compare with expected (case-sensitive, exact match)
	return provided == expected
}

// CombinedAuth creates a middleware that tries API key auth first, then falls back to JWT auth.
// This allows both authentication methods to coexist when EnabledAuthAPI is true.
func CombinedAuth(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Try API key first
		apiKey := c.Get("x-osm-api-key")
		if apiKey != "" && isValidAPIKey(apiKey, cfg.Server.AuthAPIKey) {
			return c.Next()
		}

		// Fall back to JWT auth
		return JWTAuth(cfg)(c)
	}
}
