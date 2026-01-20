package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// GetSettings returns basic settings
func GetSettings(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"base_folder": cfg.BaseFolder,
			"server": fiber.Map{
				"host": cfg.Server.Host,
				"port": cfg.Server.Port,
			},
			"version": core.VERSION,
		})
	}
}

// // UpdateSettings handles settings update
// func UpdateSettings(c *fiber.Ctx) error {
// 	return c.JSON(fiber.Map{"message": "Settings updated"})
// }

// GetSettingsYAML returns the entire YAML configuration with sensitive fields redacted
// @Summary Get YAML configuration
// @Description Returns the entire configuration file with sensitive fields redacted
// @Tags Settings
// @Produce text/yaml
// @Success 200 {string} string "YAML configuration content"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /osm/api/settings/yaml [get]
func GetSettingsYAML(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Read the config file
		settingsPath := filepath.Join(cfg.BaseFolder, "osm-settings.yaml")
		content, err := os.ReadFile(settingsPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("Failed to read config file: %v", err),
			})
		}

		// Redact sensitive fields
		redactedContent := redactSensitiveFields(string(content))

		// Return as YAML
		c.Set("Content-Type", "text/yaml")
		return c.SendString(redactedContent)
	}
}

// redactSensitiveFields redacts values of sensitive fields in YAML content
// Fields containing: _key, secret, password, username, _token (case-insensitive)
func redactSensitiveFields(content string) string {
	// Pattern matches YAML key-value pairs where key contains sensitive patterns
	// Handles both quoted and unquoted values, and preserves comments
	sensitivePatterns := []string{
		`_key`,
		`secret`,
		`password`,
		`username`,
		`_token`,
	}

	// Build regex pattern for sensitive field names
	patternStr := `(?i)^(\s*)([\w-]*(?:` + strings.Join(sensitivePatterns, "|") + `)[\w-]*):\s*(.+)$`
	re := regexp.MustCompile(patternStr)

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		// Skip comments and empty lines
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check if line matches sensitive pattern
		if matches := re.FindStringSubmatch(line); matches != nil {
			indent := matches[1]
			key := matches[2]
			value := matches[3]

			// Don't redact if value is already empty or a placeholder
			if value == `""` || value == "''" || value == "" {
				continue
			}

			// Redact the value
			lines[i] = fmt.Sprintf("%s%s: \"[REDACTED]\"", indent, key)
		}
	}

	return strings.Join(lines, "\n")
}

// // UpdateSettingsYAML replaces the entire YAML configuration
// // @Summary Update YAML configuration
// // @Description Replaces the entire configuration file with the provided YAML content
// // @Tags Settings
// // @Accept text/yaml
// // @Produce json
// // @Param config body string true "YAML configuration content"
// // @Success 200 {object} map[string]interface{} "Configuration updated successfully"
// // @Failure 400 {object} map[string]interface{} "Invalid YAML"
// // @Failure 500 {object} map[string]interface{} "Internal server error"
// // @Security BearerAuth
// // @Router /osm/api/settings/yaml [put]
// func UpdateSettingsYAML(cfg *config.Config) fiber.Handler {
// 	return func(c *fiber.Ctx) error {
// 		return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
// 			"error":   true,
// 			"message": "Updating osm-settings.yaml via API is disabled",
// 		})
// 	}
// }

// ReloadConfig forces a configuration reload (hot reload must be enabled)
// @Summary Force config reload
// @Description Forces an immediate reload of the configuration file. Hot reload must be enabled.
// @Tags Settings
// @Produce json
// @Success 200 {object} map[string]interface{} "Configuration reloaded successfully"
// @Failure 400 {object} map[string]interface{} "Hot reload not enabled"
// @Failure 500 {object} map[string]interface{} "Failed to reload configuration"
// @Security BearerAuth
// @Router /osm/api/settings/reload [post]
func ReloadConfig(hotConfig *config.HotReloadableConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if hotConfig == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Hot reload is not enabled. Start server with --hot-reload flag.",
			})
		}

		startTime := time.Now()
		if err := hotConfig.Reload(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("Failed to reload config: %v", err),
			})
		}

		return c.JSON(fiber.Map{
			"message":     "Configuration reloaded successfully",
			"version":     hotConfig.GetVersion(),
			"reload_time": time.Since(startTime).String(),
		})
	}
}

// GetConfigStatus returns the current configuration status including hot reload info
// @Summary Get config status
// @Description Returns the current configuration version and hot reload status
// @Tags Settings
// @Produce json
// @Success 200 {object} map[string]interface{} "Configuration status"
// @Security BearerAuth
// @Router /osm/api/settings/status [get]
func GetConfigStatus(hotConfig *config.HotReloadableConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if hotConfig == nil {
			return c.JSON(fiber.Map{
				"hot_reload_enabled": false,
				"version":            core.VERSION,
			})
		}

		cfg := hotConfig.Get()
		return c.JSON(fiber.Map{
			"hot_reload_enabled":  true,
			"config_version":      hotConfig.GetVersion(),
			"config_path":         hotConfig.GetConfigPath(),
			"watcher_running":     hotConfig.IsRunning(),
			"callback_count":      hotConfig.CallbackCount(),
			"server_version":      core.VERSION,
			"base_folder":         cfg.BaseFolder,
			"server_port":         cfg.Server.Port,
			"database_engine":     cfg.Database.DBEngine,
			"redis_configured":    cfg.IsRedisConfigured(),
			"storage_configured":  cfg.IsStorageConfigured(),
			"llm_configured":      cfg.IsLLMConfigured(),
			"telegram_configured": cfg.IsTelegramConfigured(),
		})
	}
}
