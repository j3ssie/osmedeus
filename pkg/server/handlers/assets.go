package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// ListAssets handles listing assets with pagination and filtering
// @Summary List assets
// @Description Get a paginated list of assets with optional filtering
// @Tags Assets
// @Produce json
// @Param workspace query string false "Filter by workspace name"
// @Param search query string false "Search in asset_value, url, title, host_ip"
// @Param status_code query int false "Filter by HTTP status code"
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Success 200 {object} map[string]interface{} "List of assets with pagination"
// @Failure 500 {object} map[string]interface{} "Failed to fetch assets"
// @Security BearerAuth
// @Router /osm/api/assets [get]
func ListAssets(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
		workspace := c.Query("workspace")
		search := c.Query("search")
		statusCode, _ := strconv.Atoi(c.Query("status_code", "0"))
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		limit, _ := strconv.Atoi(c.Query("limit", "20"))

		// Validate pagination
		if offset < 0 {
			offset = 0
		}
		if limit <= 0 {
			limit = 20
		}
		if limit > 10000 {
			limit = 10000
		}

		ctx := context.Background()

		// Get assets from database
		result, err := database.ListAssets(ctx, database.AssetQuery{
			Workspace:  workspace,
			Search:     search,
			StatusCode: statusCode,
			Offset:     offset,
			Limit:      limit,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data": result.Data,
			"pagination": fiber.Map{
				"total":  result.TotalCount,
				"offset": result.Offset,
				"limit":  result.Limit,
			},
		})
	}
}
