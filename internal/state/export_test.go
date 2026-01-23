package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestExport_WithoutDatabase(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "state-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	stateFile := filepath.Join(tmpDir, "run-state.json")

	// Create context without database
	now := time.Now()
	ctx := &ExportContext{
		RunUUID:        "test-run-123",
		WorkflowName:   "test-workflow",
		WorkflowKind:   "module",
		Target:         "example.com",
		WorkspacePath:  "/tmp/workspace",
		WorkspaceName:  "example.com",
		Status:         "running",
		StartedAt:      &now,
		TotalSteps:     5,
		CompletedSteps: 2,
		Artifacts:      []string{"/tmp/workspace/output1.txt", "/tmp/workspace/output2.txt"},
	}

	// Export state
	err = Export(stateFile, ctx)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read and verify the exported file
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	var export StateExport
	if err := json.Unmarshal(data, &export); err != nil {
		t.Fatalf("Failed to unmarshal state file: %v", err)
	}

	// Verify run info
	if export.Run == nil {
		t.Fatal("Expected Run to be populated")
	}
	if export.Run.RunUUID != "test-run-123" {
		t.Errorf("Expected RunID 'test-run-123', got '%s'", export.Run.RunUUID)
	}
	if export.Run.WorkflowName != "test-workflow" {
		t.Errorf("Expected WorkflowName 'test-workflow', got '%s'", export.Run.WorkflowName)
	}
	if export.Run.WorkflowKind != "module" {
		t.Errorf("Expected WorkflowKind 'module', got '%s'", export.Run.WorkflowKind)
	}
	if export.Run.Target != "example.com" {
		t.Errorf("Expected Target 'example.com', got '%s'", export.Run.Target)
	}
	if export.Run.Status != "running" {
		t.Errorf("Expected Status 'running', got '%s'", export.Run.Status)
	}
	if export.Run.TotalSteps != 5 {
		t.Errorf("Expected TotalSteps 5, got %d", export.Run.TotalSteps)
	}
	if export.Run.CompletedSteps != 2 {
		t.Errorf("Expected CompletedSteps 2, got %d", export.Run.CompletedSteps)
	}

	// Verify workspace info
	if export.Workspace == nil {
		t.Fatal("Expected Workspace to be populated")
	}
	if export.Workspace.Name != "example.com" {
		t.Errorf("Expected Workspace Name 'example.com', got '%s'", export.Workspace.Name)
	}
	if export.Workspace.LocalPath != "/tmp/workspace" {
		t.Errorf("Expected Workspace LocalPath '/tmp/workspace', got '%s'", export.Workspace.LocalPath)
	}

	// Verify artifacts
	if len(export.Artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(export.Artifacts))
	}

	// Verify updated_at is set
	if export.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestExport_MinimalContext(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "state-test-minimal-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	stateFile := filepath.Join(tmpDir, "run-state.json")

	// Create minimal context
	ctx := &ExportContext{
		RunUUID:       "minimal-run",
		WorkspaceName: "minimal-workspace",
	}

	// Export state
	err = Export(stateFile, ctx)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Read and verify the exported file
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("Failed to read state file: %v", err)
	}

	var export StateExport
	if err := json.Unmarshal(data, &export); err != nil {
		t.Fatalf("Failed to unmarshal state file: %v", err)
	}

	// Verify run info is present even with minimal data
	if export.Run == nil {
		t.Fatal("Expected Run to be populated")
	}
	if export.Run.RunUUID != "minimal-run" {
		t.Errorf("Expected RunID 'minimal-run', got '%s'", export.Run.RunUUID)
	}

	// Verify workspace info is present
	if export.Workspace == nil {
		t.Fatal("Expected Workspace to be populated")
	}
	if export.Workspace.Name != "minimal-workspace" {
		t.Errorf("Expected Workspace Name 'minimal-workspace', got '%s'", export.Workspace.Name)
	}
}

func TestExport_EmptyStateFilePath(t *testing.T) {
	ctx := &ExportContext{
		RunUUID: "test",
	}

	err := Export("", ctx)
	if err == nil {
		t.Error("Expected error for empty state file path")
	}
}
