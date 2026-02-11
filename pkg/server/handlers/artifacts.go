package handlers

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// DownloadWorkspaceArtifact handles downloading a file artifact from a workspace
// @Summary Download workspace artifact
// @Description Download a single file under the given workspace by relative artifact path
// @Tags Artifacts
// @Produce application/octet-stream
// @Param workspace_name path string true "Workspace name"
// @Param artifact_path query string true "Relative path to artifact under workspace"
// @Success 200 {file} binary "Artifact file"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Artifact not found"
// @Failure 500 {object} map[string]interface{} "Failed to download artifact"
// @Security BearerAuth
// @Router /osm/api/artifacts/{workspace_name} [get]
func DownloadWorkspaceArtifact(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspaceName := c.Params("workspace_name")
		artifactPath := c.Query("artifact_path")

		if !isValidWorkspaceName(workspaceName) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid workspace name",
			})
		}

		if artifactPath == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "artifact_path query parameter is required",
			})
		}

		ctx := context.Background()

		workspace, err := database.GetWorkspaceByName(ctx, workspaceName)
		var workspaceDir string
		if err == nil && workspace != nil && workspace.LocalPath != "" {
			workspaceDir = workspace.LocalPath
		} else {
			workspaceDir = filepath.Join(cfg.GetWorkspacesDir(), workspaceName)
		}

		cleanRel := filepath.Clean(artifactPath)
		if cleanRel == "." || filepath.IsAbs(cleanRel) || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) || cleanRel == ".." {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid artifact_path",
			})
		}

		fullPath := filepath.Join(workspaceDir, cleanRel)
		if !isPathUnderWorkspace(fullPath, workspaceDir) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Path traversal attempt detected",
			})
		}

		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   true,
					"message": "Artifact not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to read artifact: " + err.Error(),
			})
		}
		if info.IsDir() {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "artifact_path must point to a file",
			})
		}

		return c.SendFile(fullPath)
	}
}

// ListArtifacts handles listing artifacts with pagination and filtering
// @Summary List artifacts
// @Description Get a paginated list of artifacts with optional filtering and existence checks
// @Tags Artifacts
// @Produce json
// @Param workspace query string false "Filter by workspace name"
// @Param search query string false "Search in artifact name/path"
// @Param status_code query int false "Filter by HTTP status code (also accepts statusCode)"
// @Param verify_exist query bool false "Annotate results with path_exists and path_is_dir" default(false)
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Success 200 {object} map[string]interface{} "List of artifacts with pagination"
// @Failure 500 {object} map[string]interface{} "Failed to fetch artifacts"
// @Security BearerAuth
// @Router /osm/api/artifacts [get]
func ListArtifacts(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspace := c.Query("workspace")
		search := c.Query("search")
		statusCode, _ := strconv.Atoi(c.Query("statusCode", c.Query("status_code", "0")))
		verifyExist := c.Query("verify_exist", "false") == "true"
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

		result, err := database.ListArtifacts(ctx, database.ArtifactQuery{
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

		data := any(result.Data)
		if verifyExist {
			annotated := make([]fiber.Map, 0, len(result.Data))
			for _, a := range result.Data {
				exists := false
				isDir := false
				if a.ArtifactPath != "" {
					info, statErr := os.Stat(a.ArtifactPath)
					if statErr == nil {
						exists = true
						isDir = info.IsDir()
					}
				}

				if !exists {
					continue
				}

				annotated = append(annotated, fiber.Map{
					"id":            a.ID,
					"run_id":        a.RunID,
					"workspace":     a.Workspace,
					"name":          a.Name,
					"artifact_path": a.ArtifactPath,
					"artifact_type": a.ArtifactType,
					"content_type":  a.ContentType,
					"size_bytes":    a.SizeBytes,
					"line_count":    a.LineCount,
					"optional":      a.Optional,
					"description":   a.Description,
					"created_at":    a.CreatedAt,
					"path_exists":   exists,
					"path_is_dir":   isDir,
				})
			}
			data = annotated
		}

		return c.JSON(fiber.Map{
			"data": data,
			"pagination": fiber.Map{
				"total":  result.TotalCount,
				"offset": result.Offset,
				"limit":  result.Limit,
			},
		})
	}
}
