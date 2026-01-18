package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// ListEventLogs handles listing event logs with filtering and pagination
// @Summary List event logs
// @Description Get a paginated list of event logs with optional filtering
// @Tags EventLogs
// @Produce json
// @Param topic query string false "Filter by event topic (e.g., run.started, run.completed)"
// @Param name query string false "Filter by event name"
// @Param source query string false "Filter by source (scheduler, api, webhook)"
// @Param workspace query string false "Filter by workspace"
// @Param run_id query string false "Filter by run ID"
// @Param workflow_name query string false "Filter by workflow name"
// @Param processed query string false "Filter by processed status (true/false)"
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Success 200 {object} map[string]interface{} "List of event logs with pagination"
// @Failure 500 {object} map[string]interface{} "Failed to fetch event logs"
// @Security BearerAuth
// @Router /osm/api/event-logs [get]
func ListEventLogs(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
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

		// Build query
		query := database.EventLogQuery{
			Topic:        c.Query("topic"),
			Name:         c.Query("name"),
			Source:       c.Query("source"),
			Workspace:    c.Query("workspace"),
			RunID:        c.Query("run_id"),
			WorkflowName: c.Query("workflow_name"),
			Offset:       offset,
			Limit:        limit,
		}

		// Handle processed filter (optional bool)
		if processedStr := c.Query("processed"); processedStr != "" {
			processed := processedStr == "true"
			query.Processed = &processed
		}

		ctx := context.Background()

		// Fetch event logs
		result, err := database.ListEventLogs(ctx, query)
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
