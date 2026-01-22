package state

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// Export exports state to a JSON file
// Uses database data if available, otherwise uses data from ExportContext
func Export(stateFile string, ctx *ExportContext) error {
	if stateFile == "" {
		return fmt.Errorf("state file path is empty")
	}

	// Ensure directory exists
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	dbCtx := context.Background()
	db := database.GetDB()

	export := StateExport{
		UpdatedAt: time.Now(),
	}

	// Always use context data first for run info (has correct in-memory status)
	// The DB read happens before status is updated, so it returns stale data.
	// Context data comes from the in-memory result which has the correct status.
	if ctx != nil {
		export.Run = runInfoFromContext(ctx)
	}

	// Optionally enrich with DB data for fields not available in context
	// (currently all fields are available in context, so this is just for future-proofing)
	if export.Run == nil && ctx != nil && ctx.RunUUID != "" && db != nil {
		var run database.Run
		err := db.NewSelect().Model(&run).
			Where("run_uuid = ?", ctx.RunUUID).Scan(dbCtx)
		if err == nil {
			export.Run = runInfoFromDB(&run)
		}
	}

	// Try to load workspace from database first
	workspaceLoaded := false
	if ctx != nil && ctx.WorkspaceName != "" && db != nil {
		var ws database.Workspace
		err := db.NewSelect().Model(&ws).
			Where("name = ?", ctx.WorkspaceName).Scan(dbCtx)
		if err == nil {
			export.Workspace = workspaceInfoFromDB(&ws)
			workspaceLoaded = true
		}
	}

	// Fallback: create workspace info from context
	if !workspaceLoaded && ctx != nil {
		export.Workspace = workspaceInfoFromContext(ctx)
	}

	// Add artifacts
	if ctx != nil && len(ctx.Artifacts) > 0 {
		export.Artifacts = ctx.Artifacts
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state export: %w", err)
	}

	// Write to file
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

func runInfoFromDB(run *database.Run) *RunInfo {
	return &RunInfo{
		RunUUID:        run.RunUUID,
		WorkflowName:   run.WorkflowName,
		WorkflowKind:   run.WorkflowKind,
		Target:         run.Target,
		Params:         run.Params,
		Status:         run.Status,
		Workspace:      run.Workspace,
		StartedAt:      run.StartedAt,
		CompletedAt:    run.CompletedAt,
		ErrorMessage:   run.ErrorMessage,
		TotalSteps:     run.TotalSteps,
		CompletedSteps: run.CompletedSteps,
	}
}

func runInfoFromContext(ctx *ExportContext) *RunInfo {
	if ctx.RunUUID == "" && ctx.WorkflowName == "" {
		return nil
	}
	return &RunInfo{
		RunUUID:        ctx.RunUUID,
		WorkflowName:   ctx.WorkflowName,
		WorkflowKind:   ctx.WorkflowKind,
		Target:         ctx.Target,
		Params:         ctx.Params,
		Status:         ctx.Status,
		Workspace:      ctx.WorkspaceName,
		StartedAt:      ctx.StartedAt,
		CompletedAt:    ctx.CompletedAt,
		ErrorMessage:   ctx.ErrorMessage,
		TotalSteps:     ctx.TotalSteps,
		CompletedSteps: ctx.CompletedSteps,
	}
}

func workspaceInfoFromDB(ws *database.Workspace) *WorkspaceInfo {
	return &WorkspaceInfo{
		Name:            ws.Name,
		LocalPath:       ws.LocalPath,
		TotalAssets:     ws.TotalAssets,
		TotalSubdomains: ws.TotalSubdomains,
		TotalURLs:       ws.TotalURLs,
		TotalVulns:      ws.TotalVulns,
		VulnCritical:    ws.VulnCritical,
		VulnHigh:        ws.VulnHigh,
		VulnMedium:      ws.VulnMedium,
		VulnLow:         ws.VulnLow,
		VulnPotential:   ws.VulnPotential,
		RiskScore:       ws.RiskScore,
		Tags:            ws.Tags,
		LastRun:         ws.LastRun,
		RunWorkflow:     ws.RunWorkflow,
	}
}

func workspaceInfoFromContext(ctx *ExportContext) *WorkspaceInfo {
	if ctx.WorkspaceName == "" {
		return nil
	}
	now := time.Now()
	return &WorkspaceInfo{
		Name:        ctx.WorkspaceName,
		LocalPath:   ctx.WorkspacePath,
		LastRun:     &now,
		RunWorkflow: ctx.WorkflowName,
	}
}
