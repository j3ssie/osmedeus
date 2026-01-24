package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// ListStepResults handles listing all step results with pagination and filtering
// @Summary List step results
// @Description Get a paginated list of step results with optional filtering
// @Tags Steps
// @Produce json
// @Param workspace query string false "Filter by workspace name"
// @Param status query string false "Filter by status (pending, running, completed, failed)"
// @Param step_type query string false "Filter by step type (bash, function, etc.)"
// @Param run_uuid query string false "Filter by run UUID"
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Success 200 {object} map[string]interface{} "List of step results with pagination"
// @Failure 500 {object} map[string]interface{} "Failed to fetch step results"
// @Security BearerAuth
// @Router /osm/api/step-results [get]
func ListStepResults(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
		workspace := c.Query("workspace")
		status := c.Query("status")
		stepType := c.Query("step_type")
		runUUID := c.Query("run_uuid")
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

		// Get step results from database
		result, err := database.ListStepResults(ctx, database.StepResultQuery{
			Workspace: workspace,
			Status:    status,
			StepType:  stepType,
			RunUUID:   runUUID,
			Offset:    offset,
			Limit:     limit,
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
