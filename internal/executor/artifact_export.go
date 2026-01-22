package executor

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// RegisterArtifacts registers workflow reports and state files as artifacts in the database
// runID is the integer Run.ID used as a foreign key for artifacts
func RegisterArtifacts(workflow *core.Workflow, execCtx *core.ExecutionContext, runID int64, logger *zap.Logger) error {
	db := database.GetDB()
	if db == nil {
		return nil
	}

	ctx := context.Background()
	templateEngine := template.NewEngine()

	// Get output path
	outputPath, ok := execCtx.GetVariable("Output")
	if !ok {
		logger.Debug("Output variable not set, skipping artifact registration")
		return nil
	}
	outputStr, _ := outputPath.(string)

	// Register workflow reports
	for _, report := range workflow.Reports {
		// Render the path template
		renderedPath, err := templateEngine.Render(report.Path, execCtx.Variables)
		if err != nil {
			logger.Warn("Failed to render report path",
				zap.String("name", report.Name),
				zap.String("path", report.Path),
				zap.Error(err),
			)
			continue
		}

		// Determine content type from report type
		contentType := mapReportTypeToContentType(report.Type)

		artifact := database.Artifact{
			ID:           uuid.New().String(),
			RunID:        runID,
			Workspace:    execCtx.WorkspaceName,
			Name:         report.Name,
			ArtifactPath: renderedPath,
			ArtifactType: database.ArtifactTypeReport,
			ContentType:  contentType,
			Description:  report.Description,
			CreatedAt:    time.Now(),
		}

		// Get file stats if file exists
		if info, err := os.Stat(renderedPath); err == nil {
			artifact.SizeBytes = info.Size()
			if !info.IsDir() {
				artifact.LineCount = countLines(renderedPath)
			}
		}

		// Insert or update artifact
		_, err = db.NewInsert().Model(&artifact).
			On("CONFLICT (id) DO UPDATE").
			Set("artifact_path = EXCLUDED.artifact_path").
			Set("artifact_type = EXCLUDED.artifact_type").
			Set("content_type = EXCLUDED.content_type").
			Set("size_bytes = EXCLUDED.size_bytes").
			Set("line_count = EXCLUDED.line_count").
			Set("description = EXCLUDED.description").
			Exec(ctx)

		if err != nil {
			logger.Warn("Failed to register report artifact",
				zap.String("name", report.Name),
				zap.Error(err),
			)
		} else {
			logger.Debug("Registered report artifact",
				zap.String("name", report.Name),
				zap.String("path", renderedPath),
			)
		}
	}

	// Register state files
	for _, stateFile := range database.DefaultStateFiles {
		statePath := filepath.Join(outputStr, stateFile.FileName)

		// Check if artifact already exists for this workspace + name
		var existingArtifact database.Artifact
		err := db.NewSelect().
			Model(&existingArtifact).
			Where("workspace = ? AND name = ?", execCtx.WorkspaceName, stateFile.Name).
			Scan(ctx)

		if err == nil {
			// Artifact exists - update size_bytes and line_count only
			if info, statErr := os.Stat(statePath); statErr == nil {
				lineCount := 0
				if !info.IsDir() {
					lineCount = countLines(statePath)
				}
				_, updateErr := db.NewUpdate().
					Model(&existingArtifact).
					Set("size_bytes = ?", info.Size()).
					Set("line_count = ?", lineCount).
					Set("run_id = ?", runID). // Update run_id to latest run
					Where("id = ?", existingArtifact.ID).
					Exec(ctx)
				if updateErr != nil {
					logger.Warn("Failed to update artifact size",
						zap.String("name", stateFile.Name),
						zap.Error(updateErr),
					)
				} else {
					logger.Debug("Updated existing state file artifact",
						zap.String("name", stateFile.Name),
						zap.String("path", statePath),
					)
				}
			}
			continue // Skip insert
		}

		// Insert new artifact (only if doesn't exist)
		artifact := database.Artifact{
			ID:           uuid.New().String(),
			RunID:        runID,
			Workspace:    execCtx.WorkspaceName,
			Name:         stateFile.Name,
			ArtifactPath: statePath,
			ArtifactType: stateFile.ArtifactType,
			ContentType:  stateFile.ContentType,
			Description:  stateFile.Description,
			CreatedAt:    time.Now(),
		}

		// Get file stats if file exists
		if info, err := os.Stat(statePath); err == nil {
			artifact.SizeBytes = info.Size()
			if !info.IsDir() {
				artifact.LineCount = countLines(statePath)
			}
		}

		_, err = db.NewInsert().Model(&artifact).Exec(ctx)

		if err != nil {
			logger.Warn("Failed to register state file artifact",
				zap.String("name", stateFile.Name),
				zap.Error(err),
			)
		} else {
			logger.Debug("Registered state file artifact",
				zap.String("name", stateFile.Name),
				zap.String("path", statePath),
			)
		}
	}

	return nil
}

// mapReportTypeToContentType converts workflow report type to database content type
func mapReportTypeToContentType(reportType string) string {
	switch strings.ToLower(reportType) {
	case "json":
		return database.ContentTypeJSON
	case "jsonl":
		return database.ContentTypeJSONL
	case "yaml", "yml":
		return database.ContentTypeYAML
	case "html":
		return database.ContentTypeHTML
	case "markdown", "md":
		return database.ContentTypeMarkdown
	case "log":
		return database.ContentTypeLog
	case "pdf":
		return database.ContentTypePDF
	case "png", "image":
		return database.ContentTypePNG
	case "text", "txt":
		return database.ContentTypeText
	case "zip":
		return database.ContentTypeZip
	case "folder", "directory":
		return database.ContentTypeFolder
	default:
		return database.ContentTypeUnknown
	}
}

// countLines counts the number of lines in a file
func countLines(filePath string) int {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0
	}
	if len(data) == 0 {
		return 0
	}
	count := 1
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count
}
