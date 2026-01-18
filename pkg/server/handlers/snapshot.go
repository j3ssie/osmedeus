package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/snapshot"
)

// SnapshotExportRequest represents the request body for snapshot export
type SnapshotExportRequest struct {
	Workspace string `json:"workspace"`
}

// SnapshotImportURLRequest represents the request body for import via URL
type SnapshotImportURLRequest struct {
	URL string `json:"url"`
}

// ListSnapshots handles listing available snapshots
// @Summary List snapshots
// @Description Get a list of available snapshot files in the snapshot directory
// @Tags Snapshots
// @Produce json
// @Success 200 {object} map[string]interface{} "List of snapshots"
// @Failure 500 {object} map[string]interface{} "Failed to list snapshots"
// @Security BearerAuth
// @Router /osm/api/snapshots [get]
func ListSnapshots(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		snapshots, err := snapshot.ListSnapshots(cfg.SnapshotPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to list snapshots: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data":  snapshots,
			"count": len(snapshots),
			"path":  cfg.SnapshotPath,
		})
	}
}

// SnapshotExport handles exporting a workspace to a snapshot
// @Summary Export workspace snapshot
// @Description Export a workspace to a compressed zip archive and download it
// @Tags Snapshots
// @Accept json
// @Produce application/zip
// @Param body body SnapshotExportRequest true "Workspace to export"
// @Success 200 {file} binary "Snapshot zip file"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Workspace not found"
// @Failure 500 {object} map[string]interface{} "Failed to create snapshot"
// @Security BearerAuth
// @Router /osm/api/snapshots/export [post]
func SnapshotExport(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SnapshotExportRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		if req.Workspace == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Workspace name is required",
			})
		}

		workspacePath := filepath.Join(cfg.WorkspacesPath, req.Workspace)
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Workspace not found: " + req.Workspace,
			})
		}

		// Ensure snapshot directory exists
		if err := os.MkdirAll(cfg.SnapshotPath, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to create snapshot directory",
			})
		}

		// Generate output path
		zipFilename := fmt.Sprintf("%s_%d.zip", req.Workspace, time.Now().Unix())
		outputPath := filepath.Join(cfg.SnapshotPath, zipFilename)

		// Export workspace
		result, err := snapshot.ExportWorkspace(workspacePath, outputPath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to create snapshot: " + err.Error(),
			})
		}

		// Set headers for file download
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))
		c.Set("Content-Type", "application/zip")
		c.Set("X-Snapshot-Size", fmt.Sprintf("%d", result.FileSize))

		return c.SendFile(result.OutputPath)
	}
}

// SnapshotImport handles importing a workspace from a snapshot
// @Summary Import workspace snapshot
// @Description Import a workspace from an uploaded zip file or URL
// @Tags Snapshots
// @Accept multipart/form-data
// @Produce json
// @Param file formData file false "Snapshot zip file to import"
// @Param url formData string false "URL of snapshot to download and import"
// @Param force formData bool false "Overwrite existing workspace if present"
// @Param skip_db formData bool false "Skip database import (files only)"
// @Success 200 {object} map[string]interface{} "Import result"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Failed to import snapshot"
// @Security BearerAuth
// @Router /osm/api/snapshots/import [post]
func SnapshotImport(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		force := c.FormValue("force") == "true"
		skipDB := c.FormValue("skip_db") == "true"
		url := c.FormValue("url")

		var source string

		// Check for file upload
		file, err := c.FormFile("file")
		if err == nil && file != nil {
			// Save uploaded file temporarily
			tempDir := os.TempDir()
			tempPath := filepath.Join(tempDir, file.Filename)

			if err := c.SaveFile(file, tempPath); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   true,
					"message": "Failed to save uploaded file: " + err.Error(),
				})
			}

			source = tempPath
			defer func() { _ = os.Remove(tempPath) }()
		} else if url != "" {
			source = url
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Either file or url is required",
			})
		}

		var result *snapshot.ImportResult

		if force {
			result, err = snapshot.ForceImportWorkspace(source, cfg.WorkspacesPath, skipDB, cfg)
		} else {
			result, err = snapshot.ImportWorkspace(source, cfg.WorkspacesPath, skipDB, cfg)
		}

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to import snapshot: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message":     "Workspace imported successfully",
			"workspace":   result.WorkspaceName,
			"local_path":  result.LocalPath,
			"data_source": result.DataSource,
			"files_count": result.FilesCount,
			"warning":     "Imported workspace database state may be unstable. Only import from trusted sources.",
		})
	}
}

// DeleteSnapshot handles deleting a snapshot file
// @Summary Delete snapshot
// @Description Delete a snapshot file by name
// @Tags Snapshots
// @Produce json
// @Param name path string true "Snapshot filename"
// @Success 200 {object} map[string]interface{} "Snapshot deleted"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Snapshot not found"
// @Failure 500 {object} map[string]interface{} "Failed to delete snapshot"
// @Security BearerAuth
// @Router /osm/api/snapshots/{name} [delete]
func DeleteSnapshot(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.Params("name")
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Snapshot name is required",
			})
		}

		// Sanitize path to prevent directory traversal
		if filepath.Base(name) != name {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid snapshot name",
			})
		}

		snapshotPath := filepath.Join(cfg.SnapshotPath, name)

		// Check if file exists
		if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Snapshot not found: " + name,
			})
		}

		// Delete the file
		if err := os.Remove(snapshotPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to delete snapshot: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Snapshot deleted successfully",
			"name":    name,
		})
	}
}
