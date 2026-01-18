package handlers

import (
	"archive/zip"
	"compress/flate"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
)

// UploadFile handles uploading a file containing a list of inputs
// @Summary Upload input file
// @Description Upload a file containing a list of inputs (targets, URLs, etc.) for later use in runs
// @Tags Files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Input file to upload"
// @Success 200 {object} map[string]interface{} "File uploaded with path"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Security BearerAuth
// @Router /osm/api/upload-file [post]
func UploadFile(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the file from the form
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "No file provided",
			})
		}

		// Create uploads directory if it doesn't exist
		uploadsDir := cfg.DataPath + "/uploads"
		if err := os.MkdirAll(uploadsDir, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to create uploads directory",
			})
		}

		// Generate a unique filename to avoid conflicts
		uniqueFilename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		destPath := uploadsDir + "/" + uniqueFilename

		// Save the file
		if err := c.SaveFile(file, destPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to save file",
			})
		}

		// Count lines in the file
		lineCount := 0
		content, err := os.ReadFile(destPath)
		if err == nil {
			for _, b := range content {
				if b == '\n' {
					lineCount++
				}
			}
			// Add 1 if file doesn't end with newline but has content
			if len(content) > 0 && content[len(content)-1] != '\n' {
				lineCount++
			}
		}

		return c.JSON(fiber.Map{
			"message":  "File uploaded",
			"filename": uniqueFilename,
			"path":     destPath,
			"size":     file.Size,
			"lines":    lineCount,
		})
	}
}

// UploadWorkflow handles uploading a workflow YAML file
// @Summary Upload workflow file
// @Description Upload a raw YAML workflow file and save it to the workflows directory
// @Tags Workflows
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Workflow YAML file"
// @Success 201 {object} map[string]interface{} "Workflow uploaded"
// @Failure 400 {object} map[string]interface{} "Invalid request or YAML"
// @Security BearerAuth
// @Router /osm/api/workflow-upload [post]
func UploadWorkflow(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the file from the form
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "No file provided",
			})
		}

		// Validate file extension
		filename := file.Filename
		if !strings.HasSuffix(filename, ".yaml") && !strings.HasSuffix(filename, ".yml") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Only YAML files (.yaml or .yml) are allowed",
			})
		}

		// Open the file to read its content
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to read file",
			})
		}
		defer func() { _ = src.Close() }()

		// Read the content
		content, err := io.ReadAll(src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to read file content",
			})
		}

		// Parse the workflow to validate and get its properties
		workflow, err := parser.ParseContent(content)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid workflow YAML: " + err.Error(),
			})
		}

		// Validate workflow using the parser
		p := parser.NewParser()
		if err := p.Validate(workflow); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Workflow validation failed: " + err.Error(),
			})
		}

		// Determine target directory based on workflow kind
		var targetDir string
		if workflow.IsFlow() {
			targetDir = cfg.WorkflowsPath + "/flows"
		} else {
			targetDir = cfg.WorkflowsPath + "/modules"
		}

		// Create target directory if it doesn't exist
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to create workflows directory",
			})
		}

		// Save the workflow file
		destPath := targetDir + "/" + filename
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to save workflow file",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":     "Workflow uploaded",
			"name":        workflow.Name,
			"kind":        workflow.Kind,
			"description": workflow.Description,
			"path":        destPath,
		})
	}
}

// SnapshotDownload compresses a workspace and serves it for download
// @Summary Download workspace snapshot
// @Description Compress a workspace folder into a zip file and download it
// @Tags Snapshots
// @Produce application/zip
// @Param workspace_name path string true "Workspace name"
// @Success 200 {file} file "Zip file download"
// @Failure 404 {object} map[string]interface{} "Workspace not found"
// @Failure 500 {object} map[string]interface{} "Failed to create snapshot"
// @Security BearerAuth
// @Router /osm/api/snapshot-download/{workspace_name} [get]
func SnapshotDownload(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workspaceName := c.Params("workspace_name")
		if workspaceName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Workspace name is required",
			})
		}

		// Validate workspace exists
		workspacePath := cfg.WorkspacesPath + "/" + workspaceName
		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Workspace not found: " + workspaceName,
			})
		}

		// Create snapshot directory if it doesn't exist
		if err := os.MkdirAll(cfg.SnapshotPath, 0755); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to create snapshot directory",
			})
		}

		// Generate zip filename with timestamp
		zipFilename := fmt.Sprintf("%s_%d.zip", workspaceName, time.Now().Unix())
		zipPath := cfg.SnapshotPath + "/" + zipFilename

		// Create the zip file
		if err := createZipArchive(workspacePath, zipPath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to create snapshot: " + err.Error(),
			})
		}

		// Set headers for file download
		c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", zipFilename))
		c.Set("Content-Type", "application/zip")

		// Send the file
		return c.SendFile(zipPath)
	}
}

// createZipArchive creates a zip archive of the source directory
func createZipArchive(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer func() { _ = zipFile.Close() }()

	archive := zip.NewWriter(zipFile)
	defer func() { _ = archive.Close() }()

	// Register highest compression level for best compression ratio
	archive.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})

	// Walk through the source directory
	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Create header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = relPath

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		// Create writer for this file
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		// If it's a directory, we're done
		if info.IsDir() {
			return nil
		}

		// Copy file contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = file.Close() }()

		_, err = io.Copy(writer, file)
		return err
	})
}
