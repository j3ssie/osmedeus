package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// GetAssetDiff returns the diff between two time points for assets
// @Summary Get asset diff
// @Description Compare assets between two time points to find added, removed, and changed assets
// @Tags Assets
// @Produce json
// @Param workspace query string true "Workspace name"
// @Param from query string true "Start time (RFC3339 format or Unix timestamp)"
// @Param to query string false "End time (default: now)"
// @Success 200 {object} map[string]interface{} "Asset diff result"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Failed to get asset diff"
// @Security BearerAuth
// @Router /osm/api/assets/diff [get]
func GetAssetDiff(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspace := c.Query("workspace")
		if workspace == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "workspace parameter is required",
			})
		}

		fromStr := c.Query("from")
		if fromStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "from parameter is required",
			})
		}

		fromTime, err := parseTime(fromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("invalid from time format: %v", err),
			})
		}

		toTime := time.Now()
		if toStr := c.Query("to"); toStr != "" {
			toTime, err = parseTime(toStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": fmt.Sprintf("invalid to time format: %v", err),
				})
			}
		}

		ctx := context.Background()
		diff, err := database.GetAssetDiff(ctx, workspace, fromTime, toTime)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data": diff,
		})
	}
}

// GetVulnerabilityDiff returns the diff between two time points for vulnerabilities
// @Summary Get vulnerability diff
// @Description Compare vulnerabilities between two time points to find added, removed, and changed vulnerabilities
// @Tags Vulnerabilities
// @Produce json
// @Param workspace query string true "Workspace name"
// @Param from query string true "Start time (RFC3339 format or Unix timestamp)"
// @Param to query string false "End time (default: now)"
// @Success 200 {object} map[string]interface{} "Vulnerability diff result"
// @Failure 400 {object} map[string]interface{} "Invalid parameters"
// @Failure 500 {object} map[string]interface{} "Failed to get vulnerability diff"
// @Security BearerAuth
// @Router /osm/api/vulnerabilities/diff [get]
func GetVulnerabilityDiff(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspace := c.Query("workspace")
		if workspace == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "workspace parameter is required",
			})
		}

		fromStr := c.Query("from")
		if fromStr == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "from parameter is required",
			})
		}

		fromTime, err := parseTime(fromStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("invalid from time format: %v", err),
			})
		}

		toTime := time.Now()
		if toStr := c.Query("to"); toStr != "" {
			toTime, err = parseTime(toStr)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": fmt.Sprintf("invalid to time format: %v", err),
				})
			}
		}

		ctx := context.Background()
		diff, err := database.GetVulnerabilityDiff(ctx, workspace, fromTime, toTime)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data": diff,
		})
	}
}

// parseTime parses a time string in RFC3339 format or Unix timestamp
func parseTime(s string) (time.Time, error) {
	// Try RFC3339 first
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}

	// Try Unix timestamp (seconds)
	if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(ts, 0), nil
	}

	// Try common date formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format, use RFC3339 or Unix timestamp")
}
