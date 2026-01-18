package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// GetSystemStats returns aggregated system statistics
// @Summary Get system statistics
// @Description Get aggregated counts for workflows, runs, workspaces, assets, vulnerabilities, and schedules
// @Tags Stats
// @Produce json
// @Success 200 {object} database.SystemStats "System statistics"
// @Failure 500 {object} map[string]interface{} "Failed to get stats"
// @Security BearerAuth
// @Router /osm/api/stats [get]
func GetSystemStats(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		// Get system stats
		stats, err := database.GetSystemStats(ctx, cfg.WorkflowsPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(stats)
	}
}
