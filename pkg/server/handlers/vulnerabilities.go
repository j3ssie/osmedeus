package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// ListVulnerabilities handles listing vulnerabilities with pagination and filtering
// @Summary List vulnerabilities
// @Description Get a paginated list of vulnerabilities with optional workspace, severity, and confidence filtering
// @Tags Vulnerabilities
// @Produce json
// @Param workspace query string false "Filter by workspace name"
// @Param severity query string false "Filter by severity (critical, high, medium, low, info)"
// @Param confidence query string false "Filter by confidence (certain, firm, tentative, manual review required)"
// @Param asset_value query string false "Filter by asset value (partial match)"
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Success 200 {object} map[string]interface{} "List of vulnerabilities with pagination"
// @Failure 500 {object} map[string]interface{} "Failed to fetch vulnerabilities"
// @Security BearerAuth
// @Router /osm/api/vulnerabilities [get]
func ListVulnerabilities(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse query parameters
		workspace := c.Query("workspace")
		severity := c.Query("severity")
		confidence := c.Query("confidence")
		assetValue := c.Query("asset_value")
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

		// Get vulnerabilities from database
		result, err := database.ListVulnerabilities(ctx, database.VulnerabilityQuery{
			Workspace:  workspace,
			Severity:   severity,
			Confidence: confidence,
			AssetValue: assetValue,
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

// GetVulnerability handles getting a single vulnerability by ID
// @Summary Get vulnerability by ID
// @Description Get a single vulnerability by its ID
// @Tags Vulnerabilities
// @Produce json
// @Param id path int true "Vulnerability ID"
// @Success 200 {object} map[string]interface{} "Vulnerability details"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "Vulnerability not found"
// @Failure 500 {object} map[string]interface{} "Failed to fetch vulnerability"
// @Security BearerAuth
// @Router /osm/api/vulnerabilities/{id} [get]
func GetVulnerability(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse ID
		idStr := c.Params("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid vulnerability ID",
			})
		}

		ctx := context.Background()

		// Get vulnerability
		vuln, err := database.GetVulnerabilityByID(ctx, id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Vulnerability not found",
			})
		}

		return c.JSON(fiber.Map{
			"data": vuln,
		})
	}
}

// CreateVulnerabilityInput represents the input for creating a vulnerability
type CreateVulnerabilityInput struct {
	Workspace          string   `json:"workspace"`
	VulnInfo           string   `json:"vuln_info"`
	VulnTitle          string   `json:"vuln_title"`
	VulnDesc           string   `json:"vuln_desc"`
	VulnPOC            string   `json:"vuln_poc"`
	Severity           string   `json:"severity"`
	AssetType          string   `json:"asset_type"`
	AssetValue         string   `json:"asset_value"`
	Tags               []string `json:"tags"`
	DetailHTTPRequest  string   `json:"detail_http_request"`
	DetailHTTPResponse string   `json:"detail_http_response"`
	RawVulnJSON        string   `json:"raw_vuln_json"`
}

// CreateVulnerability handles creating a new vulnerability
// @Summary Create vulnerability
// @Description Create a new vulnerability record
// @Tags Vulnerabilities
// @Accept json
// @Produce json
// @Param vulnerability body CreateVulnerabilityInput true "Vulnerability data"
// @Success 201 {object} map[string]interface{} "Created vulnerability"
// @Failure 400 {object} map[string]interface{} "Invalid input"
// @Failure 500 {object} map[string]interface{} "Failed to create vulnerability"
// @Security BearerAuth
// @Router /osm/api/vulnerabilities [post]
func CreateVulnerability(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse input
		var input CreateVulnerabilityInput
		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		// Validate required fields
		if input.Workspace == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Workspace is required",
			})
		}

		ctx := context.Background()

		// Create vulnerability
		vuln := &database.Vulnerability{
			Workspace:          input.Workspace,
			VulnInfo:           input.VulnInfo,
			VulnTitle:          input.VulnTitle,
			VulnDesc:           input.VulnDesc,
			VulnPOC:            input.VulnPOC,
			Severity:           input.Severity,
			AssetType:          input.AssetType,
			AssetValue:         input.AssetValue,
			Tags:               input.Tags,
			DetailHTTPRequest:  input.DetailHTTPRequest,
			DetailHTTPResponse: input.DetailHTTPResponse,
			RawVulnJSON:        input.RawVulnJSON,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := database.CreateVulnerabilityRecord(ctx, vuln); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"data":    vuln,
			"message": "Vulnerability created successfully",
		})
	}
}

// DeleteVulnerability handles deleting a vulnerability
// @Summary Delete vulnerability
// @Description Delete a vulnerability by ID
// @Tags Vulnerabilities
// @Produce json
// @Param id path int true "Vulnerability ID"
// @Success 200 {object} map[string]interface{} "Vulnerability deleted"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "Vulnerability not found"
// @Failure 500 {object} map[string]interface{} "Failed to delete vulnerability"
// @Security BearerAuth
// @Router /osm/api/vulnerabilities/{id} [delete]
func DeleteVulnerability(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse ID
		idStr := c.Params("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid vulnerability ID",
			})
		}

		ctx := context.Background()

		// Delete vulnerability
		if err := database.DeleteVulnerabilityByID(ctx, id); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Vulnerability deleted successfully",
		})
	}
}

// GetVulnerabilitySummary returns severity summary for a workspace
// @Summary Get vulnerability summary
// @Description Get a summary of vulnerabilities grouped by severity
// @Tags Vulnerabilities
// @Produce json
// @Param workspace query string false "Filter by workspace name"
// @Success 200 {object} map[string]interface{} "Vulnerability summary by severity"
// @Failure 500 {object} map[string]interface{} "Failed to get summary"
// @Security BearerAuth
// @Router /osm/api/vulnerabilities/summary [get]
func GetVulnerabilitySummary(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspace := c.Query("workspace")

		ctx := context.Background()

		// Get summary
		summary, err := database.GetVulnerabilitySummary(ctx, workspace)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		// Calculate total
		total := 0
		for _, count := range summary {
			total += count
		}

		return c.JSON(fiber.Map{
			"data": fiber.Map{
				"by_severity": summary,
				"total":       total,
				"workspace":   workspace,
			},
		})
	}
}
