package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/pkg/server/middleware"
)

// Login handles user authentication
// @Summary User login
// @Description Authenticate user and get JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Router /osm/api/login [post]
func Login(cfg *config.Config, noAuth bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		// If running in no-auth mode, accept any credentials
		if noAuth {
			// Use a default username if none provided
			username := req.Username
			if username == "" {
				username = "anonymous"
			}
			token, err := middleware.GenerateToken(username, cfg)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   true,
					"message": "Failed to generate token",
				})
			}

			// Set session cookie for browser clients
			c.Cookie(&fiber.Cookie{
				Name:     "osmedeus_session",
				Value:    token,
				Expires:  time.Now().Add(time.Duration(cfg.Server.JWT.ExpirationMinutes) * time.Minute),
				HTTPOnly: false, // Allow JS to read for UI state
				Secure:   false, // Set to true in production with HTTPS
				SameSite: "Lax",
				Path:     "/",
			})

			return c.JSON(fiber.Map{
				"token": token,
			})
		}

		// Validate credentials against user map
		expectedPassword, userExists := cfg.Server.SimpleUserMapKey[req.Username]
		if !userExists || expectedPassword != req.Password {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid credentials",
			})
		}

		// Generate token
		token, err := middleware.GenerateToken(req.Username, cfg)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to generate token",
			})
		}

		// Set session cookie for browser clients
		c.Cookie(&fiber.Cookie{
			Name:     "osmedeus_session",
			Value:    token,
			Expires:  time.Now().Add(time.Duration(cfg.Server.JWT.ExpirationMinutes) * time.Minute),
			HTTPOnly: false, // Allow JS to read for UI state
			Secure:   false, // Set to true in production with HTTPS
			SameSite: "Lax",
			Path:     "/",
		})

		return c.JSON(fiber.Map{
			"token": token,
		})
	}
}

// Logout handles user logout by clearing the session cookie
// @Summary User logout
// @Description Clear the session cookie
// @Tags Auth
// @Produce json
// @Success 200 {object} map[string]string "Logout message"
// @Router /osm/api/logout [post]
func Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Clear the session cookie by setting it to expire in the past
		c.Cookie(&fiber.Cookie{
			Name:     "osmedeus_session",
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: false,
			Path:     "/",
		})
		return c.JSON(fiber.Map{
			"message": "Logged out",
		})
	}
}

// RefreshToken handles token refresh
func RefreshToken(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims := middleware.GetUser(c)
		if claims == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid token",
			})
		}

		token, err := middleware.GenerateToken(claims.Username, cfg)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to generate token",
			})
		}

		return c.JSON(fiber.Map{
			"token": token,
		})
	}
}
