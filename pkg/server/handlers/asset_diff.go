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

// ListAssetDiffSnapshots handles listing stored asset diff snapshots
// @Summary List asset diff snapshots
// @Description Get a paginated list of stored asset diff snapshots
// @Tags Assets
// @Produce json
// @Param workspace query string false "Filter by workspace name"
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Success 200 {object} map[string]interface{} "List of asset diff snapshots with pagination"
// @Failure 500 {object} map[string]interface{} "Failed to fetch asset diff snapshots"
// @Security BearerAuth
// @Router /osm/api/assets/diffs [get]
func ListAssetDiffSnapshots(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspace := c.Query("workspace")
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		limit, _ := strconv.Atoi(c.Query("limit", "20"))

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
		result, err := database.ListAssetDiffSnapshots(ctx, workspace, offset, limit)
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

// ListVulnDiffSnapshots handles listing stored vulnerability diff snapshots
// @Summary List vulnerability diff snapshots
// @Description Get a paginated list of stored vulnerability diff snapshots
// @Tags Vulnerabilities
// @Produce json
// @Param workspace query string false "Filter by workspace name"
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Success 200 {object} map[string]interface{} "List of vulnerability diff snapshots with pagination"
// @Failure 500 {object} map[string]interface{} "Failed to fetch vulnerability diff snapshots"
// @Security BearerAuth
// @Router /osm/api/vulnerabilities/diffs [get]
func ListVulnDiffSnapshots(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspace := c.Query("workspace")
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		limit, _ := strconv.Atoi(c.Query("limit", "20"))

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
		result, err := database.ListVulnDiffSnapshots(ctx, workspace, offset, limit)
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
