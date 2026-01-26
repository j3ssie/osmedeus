package client

import (
	"fmt"
	"time"
)

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

// APIError represents an API error with status code and message
type APIError struct {
	StatusCode int
	Message    string
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// Pagination represents pagination information in API responses
type Pagination struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// RunsResponse represents the response from listing runs
type RunsResponse struct {
	Data       []Run      `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Run represents a workflow run from the API
type Run struct {
	ID             int64                  `json:"id"`
	RunUUID        string                 `json:"run_uuid"`
	WorkflowName   string                 `json:"workflow_name"`
	WorkflowKind   string                 `json:"workflow_kind"`
	Target         string                 `json:"target"`
	Params         map[string]interface{} `json:"params,omitempty"`
	Status         string                 `json:"status"`
	TriggerType    string                 `json:"trigger_type,omitempty"`
	RunGroupID     string                 `json:"run_group_id,omitempty"`
	TotalSteps     int                    `json:"total_steps"`
	CompletedSteps int                    `json:"completed_steps"`
	Workspace      string                 `json:"workspace"`
	RunPriority    string                 `json:"run_priority,omitempty"`
	RunMode        string                 `json:"run_mode,omitempty"`
	StartedAt      *time.Time             `json:"started_at,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// CreateRunRequest represents a request to create a new run
type CreateRunRequest struct {
	Flow            string            `json:"flow,omitempty"`
	Module          string            `json:"module,omitempty"`
	Target          string            `json:"target,omitempty"`
	Targets         []string          `json:"targets,omitempty"`
	Params          map[string]string `json:"params,omitempty"`
	Concurrency     int               `json:"concurrency,omitempty"`
	Priority        string            `json:"priority,omitempty"`
	RunMode         string            `json:"run_mode,omitempty"`
	Timeout         int               `json:"timeout,omitempty"`
	RunnerType      string            `json:"runner_type,omitempty"`
	ThreadsHold     int               `json:"threads_hold,omitempty"`
	EmptyTarget     bool              `json:"empty_target,omitempty"`
	Repeat          bool              `json:"repeat,omitempty"`
	RepeatWaitTime  string            `json:"repeat_wait_time,omitempty"`
	HeuristicsCheck string            `json:"heuristics_check,omitempty"`
}

// CreateRunResponse represents the response from creating a run
type CreateRunResponse struct {
	Message     string   `json:"message"`
	Workflow    string   `json:"workflow"`
	Kind        string   `json:"kind"`
	TargetCount int      `json:"target_count"`
	Priority    string   `json:"priority"`
	JobID       string   `json:"job_id"`
	Status      string   `json:"status"`
	PollURL     string   `json:"poll_url"`
	Target      string   `json:"target,omitempty"`
	RunUUID     string   `json:"run_uuid,omitempty"`
	Targets     []string `json:"targets,omitempty"`
	Concurrency int      `json:"concurrency,omitempty"`
}

// CancelRunResponse represents the response from cancelling a run
type CancelRunResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
	RunUUID string `json:"run_uuid"`
}

// AssetsResponse represents the response from listing assets
type AssetsResponse struct {
	Data       []Asset    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Asset represents an asset from the API
type Asset struct {
	ID          int64     `json:"id"`
	Workspace   string    `json:"workspace"`
	AssetType   string    `json:"asset_type"`
	AssetValue  string    `json:"asset_value"`
	URL         string    `json:"url,omitempty"`
	Title       string    `json:"title,omitempty"`
	StatusCode  int       `json:"status_code,omitempty"`
	HostIP      string    `json:"host_ip,omitempty"`
	TechStack   []string  `json:"tech_stack,omitempty"`
	ContentType string    `json:"content_type,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkspacesResponse represents the response from listing workspaces
type WorkspacesResponse struct {
	Data       []Workspace `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// Workspace represents a workspace from the API
type Workspace struct {
	ID          int64     `json:"id,omitempty"`
	Name        string    `json:"name"`
	LocalPath   string    `json:"local_path,omitempty"`
	DataSource  string    `json:"data_source,omitempty"`
	TotalAssets int       `json:"total_assets,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// VulnerabilitiesResponse represents the response from listing vulnerabilities
type VulnerabilitiesResponse struct {
	Data       []Vulnerability `json:"data"`
	Pagination Pagination      `json:"pagination"`
}

// Vulnerability represents a vulnerability from the API
type Vulnerability struct {
	ID                 int64     `json:"id"`
	Workspace          string    `json:"workspace"`
	VulnInfo           string    `json:"vuln_info,omitempty"`
	VulnTitle          string    `json:"vuln_title,omitempty"`
	VulnDesc           string    `json:"vuln_desc,omitempty"`
	VulnPOC            string    `json:"vuln_poc,omitempty"`
	Severity           string    `json:"severity,omitempty"`
	Confidence         string    `json:"confidence,omitempty"`
	AssetType          string    `json:"asset_type,omitempty"`
	AssetValue         string    `json:"asset_value,omitempty"`
	Tags               []string  `json:"tags,omitempty"`
	DetailHTTPRequest  string    `json:"detail_http_request,omitempty"`
	DetailHTTPResponse string    `json:"detail_http_response,omitempty"`
	RawVulnJSON        string    `json:"raw_vuln_json,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// FunctionEvalRequest represents a request to evaluate a function
type FunctionEvalRequest struct {
	Script string            `json:"script"`
	Target string            `json:"target,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// FunctionEvalResponse represents the response from evaluating a function
type FunctionEvalResponse struct {
	Result         interface{} `json:"result"`
	RenderedScript string      `json:"rendered_script,omitempty"`
	Error          bool        `json:"error,omitempty"`
	Message        string      `json:"message,omitempty"`
}

// StepResultsResponse represents the response from listing step results
type StepResultsResponse struct {
	Data       []StepResult `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

// StepResult represents a step execution result from the API
type StepResult struct {
	ID           string                 `json:"id"`
	RunID        int64                  `json:"run_id"`
	StepName     string                 `json:"step_name"`
	StepType     string                 `json:"step_type"`
	Status       string                 `json:"status"`
	Command      string                 `json:"command,omitempty"`
	Output       string                 `json:"output,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Exports      map[string]interface{} `json:"exports,omitempty"`
	DurationMs   int64                  `json:"duration_ms"`
	LogFile      string                 `json:"log_file,omitempty"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ArtifactsResponse represents the response from listing artifacts
type ArtifactsResponse struct {
	Data       []Artifact `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Artifact represents an artifact from the API
type Artifact struct {
	ID           string    `json:"id"`
	RunID        int64     `json:"run_id"`
	Workspace    string    `json:"workspace"`
	Name         string    `json:"name"`
	ArtifactPath string    `json:"artifact_path"`
	ArtifactType string    `json:"artifact_type,omitempty"`
	ContentType  string    `json:"content_type,omitempty"`
	SizeBytes    int64     `json:"size_bytes"`
	LineCount    int       `json:"line_count"`
	Description  string    `json:"description,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// EventLogsResponse represents the response from listing event logs
type EventLogsResponse struct {
	Data       []EventLog `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// EventLog represents an event log from the API
type EventLog struct {
	ID           int64      `json:"id"`
	Topic        string     `json:"topic"`
	EventID      string     `json:"event_id,omitempty"`
	Name         string     `json:"name,omitempty"`
	SourceType   string     `json:"source_type,omitempty"` // "run", "eval", "api" - origin of the event
	Source       string     `json:"source,omitempty"`
	DataType     string     `json:"data_type,omitempty"`
	Data         string     `json:"data,omitempty"`
	Workspace    string     `json:"workspace,omitempty"`
	RunID        string     `json:"run_id,omitempty"`
	WorkflowName string     `json:"workflow_name,omitempty"`
	Processed    bool       `json:"processed"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	Error        string     `json:"error,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// SchedulesResponse represents the response from listing schedules
type SchedulesResponse struct {
	Data       []Schedule `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Schedule represents a schedule from the API
type Schedule struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	WorkflowName string                 `json:"workflow_name"`
	WorkflowKind string                 `json:"workflow_kind,omitempty"`
	Target       string                 `json:"target,omitempty"`
	Workspace    string                 `json:"workspace,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
	TriggerName  string                 `json:"trigger_name"`
	TriggerType  string                 `json:"trigger_type"`
	Schedule     string                 `json:"schedule,omitempty"`
	EventTopic   string                 `json:"event_topic,omitempty"`
	WatchPath    string                 `json:"watch_path,omitempty"`
	IsEnabled    bool                   `json:"is_enabled"`
	LastRun      *time.Time             `json:"last_run,omitempty"`
	NextRun      *time.Time             `json:"next_run,omitempty"`
	RunCount     int                    `json:"run_count"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// AssetDiffsResponse represents the response from listing asset diffs
type AssetDiffsResponse struct {
	Data       []AssetDiff `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// AssetDiff represents an asset diff snapshot from the API
type AssetDiff struct {
	ID            int64     `json:"id"`
	WorkspaceName string    `json:"workspace_name"`
	FromTime      time.Time `json:"from_time"`
	ToTime        time.Time `json:"to_time"`
	TotalAdded    int       `json:"total_added"`
	TotalRemoved  int       `json:"total_removed"`
	TotalChanged  int       `json:"total_changed"`
	DiffData      string    `json:"diff_data,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// VulnDiffsResponse represents the response from listing vulnerability diffs
type VulnDiffsResponse struct {
	Data       []VulnDiff `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// VulnDiff represents a vulnerability diff snapshot from the API
type VulnDiff struct {
	ID            int64     `json:"id"`
	WorkspaceName string    `json:"workspace_name"`
	FromTime      time.Time `json:"from_time"`
	ToTime        time.Time `json:"to_time"`
	TotalAdded    int       `json:"total_added"`
	TotalRemoved  int       `json:"total_removed"`
	TotalChanged  int       `json:"total_changed"`
	DiffData      string    `json:"diff_data,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
