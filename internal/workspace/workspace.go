package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// Workspace manages the execution workspace
type Workspace struct {
	BaseDir     string
	WorkflowDir string
	LogDir      string
	OutputDir   string
	ReportDir   string
	RunUUID     string
	Target      string
}

// NewWorkspace creates a new workspace for a workflow execution
func NewWorkspace(baseDir, workflowName, runID, target string) (*Workspace, error) {
	// Sanitize target for use in path
	sanitizedTarget := sanitizePathComponent(target)

	// Create workspace path: baseDir/target/workflow-runID-timestamp
	timestamp := time.Now().Format("20060102-150405")
	workflowDir := filepath.Join(baseDir, sanitizedTarget, fmt.Sprintf("%s-%s-%s", workflowName, runID, timestamp))

	w := &Workspace{
		BaseDir:     baseDir,
		WorkflowDir: workflowDir,
		LogDir:      filepath.Join(workflowDir, "logs"),
		OutputDir:   filepath.Join(workflowDir, "output"),
		ReportDir:   filepath.Join(workflowDir, "reports"),
		RunUUID:     runID,
		Target:      target,
	}

	// Create directories
	if err := w.createDirectories(); err != nil {
		return nil, err
	}

	return w, nil
}

// createDirectories creates all workspace directories
func (w *Workspace) createDirectories() error {
	dirs := []string{
		w.WorkflowDir,
		w.LogDir,
		w.OutputDir,
		w.ReportDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetPath returns the full path for a relative path within the workspace
func (w *Workspace) GetPath(relativePath string) string {
	return filepath.Join(w.WorkflowDir, relativePath)
}

// GetOutputPath returns the full path for an output file
func (w *Workspace) GetOutputPath(filename string) string {
	return filepath.Join(w.OutputDir, filename)
}

// GetLogPath returns the full path for a log file
func (w *Workspace) GetLogPath(filename string) string {
	return filepath.Join(w.LogDir, filename)
}

// GetReportPath returns the full path for a report file
func (w *Workspace) GetReportPath(filename string) string {
	return filepath.Join(w.ReportDir, filename)
}

// WriteOutput writes content to an output file
func (w *Workspace) WriteOutput(filename, content string) error {
	path := w.GetOutputPath(filename)
	return os.WriteFile(path, []byte(content), 0644)
}

// WriteLog writes content to a log file
func (w *Workspace) WriteLog(filename, content string) error {
	path := w.GetLogPath(filename)
	return os.WriteFile(path, []byte(content), 0644)
}

// AppendLog appends content to a log file
func (w *Workspace) AppendLog(filename, content string) error {
	path := w.GetLogPath(filename)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(content)
	return err
}

// ListOutputFiles lists all files in the output directory
func (w *Workspace) ListOutputFiles() ([]string, error) {
	return listFiles(w.OutputDir)
}

// ListReportFiles lists all files in the reports directory
func (w *Workspace) ListReportFiles() ([]string, error) {
	return listFiles(w.ReportDir)
}

// Cleanup removes the workspace directory
func (w *Workspace) Cleanup() error {
	return os.RemoveAll(w.WorkflowDir)
}

// GetVariables returns workspace-related variables for template rendering
func (w *Workspace) GetVariables() map[string]interface{} {
	return map[string]interface{}{
		"workspace":  w.WorkflowDir,
		"output":     w.OutputDir,
		"output_dir": w.OutputDir,
		"log_dir":    w.LogDir,
		"report_dir": w.ReportDir,
		"run_uuid":   w.RunUUID,
		"target":     w.Target,
	}
}

// ApplyToContext applies workspace variables to an execution context
func (w *Workspace) ApplyToContext(ctx *core.ExecutionContext) {
	ctx.WorkspacePath = w.WorkflowDir
	for k, v := range w.GetVariables() {
		ctx.SetVariable(k, v)
	}
}

// sanitizePathComponent makes a string safe for use in a file path
func sanitizePathComponent(s string) string {
	// Replace common problematic characters
	replacer := map[rune]rune{
		'/':  '_',
		'\\': '_',
		':':  '_',
		'*':  '_',
		'?':  '_',
		'"':  '_',
		'<':  '_',
		'>':  '_',
		'|':  '_',
	}

	result := make([]rune, 0, len(s))
	for _, r := range s {
		if replacement, ok := replacer[r]; ok {
			result = append(result, replacement)
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

// listFiles lists all files in a directory
func listFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// Manager manages multiple workspaces
type Manager struct {
	baseDir    string
	workspaces map[string]*Workspace
}

// NewManager creates a new workspace manager
func NewManager(baseDir string) *Manager {
	return &Manager{
		baseDir:    baseDir,
		workspaces: make(map[string]*Workspace),
	}
}

// CreateWorkspace creates a new workspace
func (m *Manager) CreateWorkspace(workflowName, runID, target string) (*Workspace, error) {
	w, err := NewWorkspace(m.baseDir, workflowName, runID, target)
	if err != nil {
		return nil, err
	}
	m.workspaces[runID] = w
	return w, nil
}

// GetWorkspace returns a workspace by run ID
func (m *Manager) GetWorkspace(runID string) (*Workspace, bool) {
	w, ok := m.workspaces[runID]
	return w, ok
}

// RemoveWorkspace removes a workspace from management (does not delete files)
func (m *Manager) RemoveWorkspace(runID string) {
	delete(m.workspaces, runID)
}

// ListWorkspaces returns all managed workspaces
func (m *Manager) ListWorkspaces() []*Workspace {
	workspaces := make([]*Workspace, 0, len(m.workspaces))
	for _, w := range m.workspaces {
		workspaces = append(workspaces, w)
	}
	return workspaces
}
