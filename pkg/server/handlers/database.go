package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// ClearTableRequest represents a request to clear a database table
type ClearTableRequest struct {
	Force bool `json:"force"` // Must be true to confirm the destructive operation
}

// ListDatabaseTables returns information about all database tables
// @Summary List database tables
// @Description Get list of all database tables with row counts
// @Tags Database
// @Produce json
// @Success 200 {object} map[string]interface{} "List of tables with counts"
// @Failure 500 {object} map[string]interface{} "Internal error"
// @Security BearerAuth
// @Router /osm/api/database/tables [get]
func ListDatabaseTables(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		tables, err := database.ListTables(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to list tables: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data":         tables,
			"valid_tables": database.ValidTableNames(),
		})
	}
}

// ClearDatabaseTable clears all records from a specific table
// @Summary Clear database table
// @Description Delete all records from a specific database table (destructive operation)
// @Tags Database
// @Accept json
// @Produce json
// @Param table path string true "Table name (runs, step_results, event_logs, artifacts, assets, schedules, workspaces, vulnerabilities, asset_diffs, vuln_diffs)"
// @Param body body ClearTableRequest true "Confirmation with force=true required"
// @Success 200 {object} map[string]interface{} "Table cleared successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or missing force confirmation"
// @Failure 500 {object} map[string]interface{} "Failed to clear table"
// @Security BearerAuth
// @Router /osm/api/database/tables/{table}/clear [post]
func ClearDatabaseTable(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tableName := c.Params("table")
		if tableName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Table name is required",
			})
		}

		var req ClearTableRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		// Require force confirmation for destructive operation
		if !req.Force {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        true,
				"message":      "This is a destructive operation. Set force=true to confirm deletion of all records.",
				"valid_tables": database.ValidTableNames(),
			})
		}

		ctx := context.Background()

		if err := database.ClearTable(ctx, tableName); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":        true,
				"message":      err.Error(),
				"valid_tables": database.ValidTableNames(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Table cleared successfully",
			"table":   tableName,
		})
	}
}
