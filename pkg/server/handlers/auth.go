package handlers

import (
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
func Login(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
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

		return c.JSON(fiber.Map{
			"token": token,
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
