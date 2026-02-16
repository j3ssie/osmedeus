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

// GetAssetStats returns unique lists of technologies, sources, remarks, and asset types
// @Summary Get asset statistics
// @Description Get unique values for technologies, sources, remarks, and asset types across all assets or filtered by workspace
// @Tags Stats
// @Produce json
// @Param workspace query string false "Filter by workspace name"
// @Success 200 {object} database.AssetStatsData "Asset statistics"
// @Failure 500 {object} map[string]interface{} "Failed to get stats"
// @Security BearerAuth
// @Router /osm/api/asset-stats [get]
func GetAssetStats(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()
		workspace := c.Query("workspace")

		// Get asset stats
		stats, err := database.GetAssetStats(ctx, workspace)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data": stats,
		})
	}
}
