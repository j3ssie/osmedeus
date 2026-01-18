package state

import "time"

// StateExport represents the run state exported to JSON
type StateExport struct {
	Run       *RunInfo       `json:"run,omitempty"`
	Workspace *WorkspaceInfo `json:"workspace,omitempty"`
	Artifacts []string       `json:"artifacts,omitempty"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// RunInfo contains run information for export (mirrors database.Run fields)
type RunInfo struct {
	RunID          string                 `json:"run_id"`
	WorkflowName   string                 `json:"workflow_name"`
	WorkflowKind   string                 `json:"workflow_kind"`
	Target         string                 `json:"target"`
	Params         map[string]any `json:"params,omitempty"`
	Status         string                 `json:"status"`
	WorkspacePath  string                 `json:"workspace_path"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	TotalSteps     int                    `json:"total_steps"`
	CompletedSteps int                    `json:"completed_steps"`
}

// WorkspaceInfo contains workspace information for export
type WorkspaceInfo struct {
	Name            string     `json:"name"`
	LocalPath       string     `json:"local_path,omitempty"`
	TotalAssets     int        `json:"total_assets"`
	TotalSubdomains int        `json:"total_subdomains"`
	TotalURLs       int        `json:"total_urls"`
	TotalVulns      int        `json:"total_vulns"`
	VulnCritical    int        `json:"vuln_critical"`
	VulnHigh        int        `json:"vuln_high"`
	VulnMedium      int        `json:"vuln_medium"`
	VulnLow         int        `json:"vuln_low"`
	VulnPotential   int        `json:"vuln_potential"`
	RiskScore       float64    `json:"risk_score"`
	Tags            []string   `json:"tags,omitempty"`
	LastRun         *time.Time `json:"last_run,omitempty"`
	RunWorkflow     string     `json:"run_workflow,omitempty"`
}

// ExportContext provides the context needed for state export
// This allows callers to provide whatever information they have available
type ExportContext struct {
	RunID          string
	WorkflowName   string
	WorkflowKind   string
	Target         string
	WorkspacePath  string
	WorkspaceName  string
	Params         map[string]any
	Status         string
	StartedAt      *time.Time
	CompletedAt    *time.Time
	ErrorMessage   string
	TotalSteps     int
	CompletedSteps int
	Artifacts      []string
}
