package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
)

// ServerInfoData holds cached server info set once at startup
type ServerInfoData struct {
	License string
	Version string
	Binary  string
	Repo    string
	Author  string
	Docs    string
}

// cachedServerInfo holds server info set once at startup
var cachedServerInfo *ServerInfoData

// SetServerInfo sets the cached server info (called once at server startup)
func SetServerInfo(info *ServerInfoData) {
	cachedServerInfo = info
}

// HealthCheck handles health check requests
// @Summary Health check
// @Description Check if the server is running
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "status: ok"
// @Router /health [get]
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
	})
}

// ReadinessCheck handles readiness check requests
// @Summary Readiness check
// @Description Check if the server is ready to accept requests
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "status: ready"
// @Router /health/ready [get]
func ReadinessCheck(c *fiber.Ctx) error {
	// Check database connection, etc.
	return c.JSON(fiber.Map{
		"status": "ready",
	})
}

// Root handles the root endpoint showing version info
// @Summary Server info
// @Description Get server version and info
// @Tags Info
// @Produce json
// @Success 200 {object} map[string]string "Server information"
// @Router / [get]
func Root(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Oh dear me, how delightful to notice you're taking a look at this! I'm ever so pleased to let you know that %s is ticking along quite nicely, thank you.", cachedServerInfo.Binary),
			"version": cachedServerInfo.Version,
			"repo":    cachedServerInfo.Repo,
			"author":  cachedServerInfo.Author,
			"docs":    cachedServerInfo.Docs,
			"license": cachedServerInfo.License,
		})
	}
}

// ServerInfo handles the /server-info endpoint
// @Summary Server info JSON
// @Description Get server version and info in JSON
// @Tags Info
// @Produce json
// @Success 200 {object} map[string]string "Server information"
// @Router /server-info [get]
func ServerInfo(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": fmt.Sprintf("Oh dear me, how delightful to notice you're taking a look at this! I'm ever so pleased to let you know that %s is ticking along quite nicely, thank you.", cachedServerInfo.Binary),
			"version": cachedServerInfo.Version,
			"repo":    cachedServerInfo.Repo,
			"author":  cachedServerInfo.Author,
			"docs":    cachedServerInfo.Docs,
			"license": cachedServerInfo.License,
		})
	}
}
