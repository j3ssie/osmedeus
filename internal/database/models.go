package database

import (
	"time"

	"github.com/uptrace/bun"
)

// Run represents a workflow execution
type Run struct {
	bun.BaseModel `bun:"table:runs,alias:r"`

	ID            int64                  `bun:"id,pk,autoincrement" json:"id"`
	RunUUID       string                 `bun:"run_uuid,unique,notnull" json:"run_uuid"`
	WorkflowName  string                 `bun:"workflow_name,notnull" json:"workflow_name"`
	WorkflowKind  string                 `bun:"workflow_kind,notnull" json:"workflow_kind"`
	Target        string                 `bun:"target,notnull" json:"target"`
	Params        map[string]interface{} `bun:"params,type:json" json:"params"`
	Status        string                 `bun:"status,notnull" json:"status"`
	Workspace string                 `bun:"workspace" json:"workspace"`
	StartedAt     *time.Time             `bun:"started_at" json:"started_at"`
	CompletedAt   *time.Time             `bun:"completed_at" json:"completed_at"`
	ErrorMessage  string                 `bun:"error_message" json:"error_message,omitempty"`
	CreatedAt     time.Time              `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time              `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// Scheduling context
	ScheduleID  string `bun:"schedule_id" json:"schedule_id,omitempty"`
	TriggerType string `bun:"trigger_type" json:"trigger_type,omitempty"` // manual, cron, event
	TriggerName string `bun:"trigger_name" json:"trigger_name,omitempty"`

	// Run grouping - multiple targets from same request share a RunGroupID
	RunGroupID string `bun:"run_group_id" json:"run_group_id,omitempty"`

	// Progress tracking
	TotalSteps     int `bun:"total_steps" json:"total_steps"`
	CompletedSteps int `bun:"completed_steps" json:"completed_steps"`

	// Relations
	Steps     []*StepResult `bun:"rel:has-many,join:id=run_id" json:"steps,omitempty"`
	Artifacts []*Artifact   `bun:"rel:has-many,join:id=run_id" json:"artifacts,omitempty"`
	Events    []*EventLog   `bun:"rel:has-many,join:run_uuid=run_id" json:"events,omitempty"`
}

// StepResult represents a step execution result
type StepResult struct {
	bun.BaseModel `bun:"table:step_results,alias:sr"`

	ID           string                 `bun:"id,pk,type:text" json:"id"`
	RunID        int64                  `bun:"run_id,notnull" json:"run_id"`
	StepName     string                 `bun:"step_name,notnull" json:"step_name"`
	StepType     string                 `bun:"step_type,notnull" json:"step_type"`
	Status       string                 `bun:"status,notnull" json:"status"`
	Command      string                 `bun:"command" json:"command,omitempty"`
	Output       string                 `bun:"output" json:"output,omitempty"`
	ErrorMessage string                 `bun:"error_message" json:"error_message,omitempty"`
	Exports      map[string]interface{} `bun:"exports,type:json" json:"exports,omitempty"`
	DurationMs   int64                  `bun:"duration_ms" json:"duration_ms"`
	LogFile      string                 `bun:"log_file" json:"log_file,omitempty"`
	StartedAt    *time.Time             `bun:"started_at" json:"started_at"`
	CompletedAt  *time.Time             `bun:"completed_at" json:"completed_at"`
	CreatedAt    time.Time              `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`

	// Relations
	Run *Run `bun:"rel:belongs-to,join:run_id=id" json:"run,omitempty"`
}

// Artifact content type constants
const (
	ContentTypeJSON     = "json"
	ContentTypeJSONL    = "jsonl"
	ContentTypeYAML     = "yaml"
	ContentTypeHTML     = "html"
	ContentTypeMarkdown = "md"
	ContentTypeLog      = "log"
	ContentTypePDF      = "pdf"
	ContentTypePNG      = "png"
	ContentTypeText     = "txt"
	ContentTypeZip      = "zip"
	ContentTypeFolder   = "folder"
	ContentTypeUnknown  = "unknown"
)

// Artifact type constants for categorization
const (
	ArtifactTypeReport     = "report"     // Workflow reports from reports: section
	ArtifactTypeStateFile  = "state_file" // State files like run-state.json
	ArtifactTypeOutput     = "output"     // General output files
	ArtifactTypeScreenshot = "screenshot" // Screenshots
)

// Default state file names
var DefaultStateFiles = []struct {
	Name         string
	FileName     string
	ContentType  string
	ArtifactType string
	Description  string
}{
	{"state-execution-log", "run-execution.log", ContentTypeLog, ArtifactTypeStateFile, "Execution log file"},
	{"state-console-log", "run-console.log", ContentTypeLog, ArtifactTypeStateFile, "Console output capture"},
	{"state-completed", "run-completed.json", ContentTypeJSON, ArtifactTypeStateFile, "Completed state marker"},
	{"state-file", "run-state.json", ContentTypeJSON, ArtifactTypeStateFile, "Run state tracking file"},
	{"state-workflow", "run-workflow.yaml", ContentTypeYAML, ArtifactTypeStateFile, "Workflow definition used for the run"},
}

// Artifact represents an output file from a run
type Artifact struct {
	bun.BaseModel `bun:"table:artifacts,alias:a"`

	ID           string    `bun:"id,pk,type:text" json:"id"`
	RunID        int64     `bun:"run_id,notnull" json:"run_id"`
	Workspace    string    `bun:"workspace,notnull" json:"workspace"`
	Name         string    `bun:"name,notnull" json:"name"`
	ArtifactPath string    `bun:"artifact_path,notnull" json:"artifact_path"`
	ArtifactType string    `bun:"artifact_type" json:"artifact_type,omitempty"` // report, state_file, output, screenshot
	ContentType  string    `bun:"content_type" json:"content_type,omitempty"`   // json, jsonl, yaml, html, md, log, pdf, png, txt, zip, folder, unknown
	SizeBytes    int64     `bun:"size_bytes" json:"size_bytes"`
	LineCount    int       `bun:"line_count" json:"line_count"`
	Description  string    `bun:"description" json:"description,omitempty"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`

	// Relations
	Run *Run `bun:"rel:belongs-to,join:run_id=id" json:"run,omitempty"`
}

// EventLog represents a system event for auditing and trigger history
type EventLog struct {
	bun.BaseModel `bun:"table:event_logs,alias:el"`

	ID       int64  `bun:"id,pk,autoincrement" json:"id"`
	Topic    string `bun:"topic,notnull" json:"topic"` // e.g., "webhook.received"
	EventID  string `bun:"event_id" json:"event_id"`   // UUID
	Name     string `bun:"name" json:"name"`           // e.g., "scan.started"
	Source   string `bun:"source" json:"source"`       // e.g., "scheduler", "api"
	DataType string `bun:"data_type" json:"data_type"` // e.g., "scan", "asset"
	Data     string `bun:"data" json:"data"`           // JSON payload

	// Context
	Workspace    string `bun:"workspace" json:"workspace,omitempty"`
	RunID        string `bun:"run_id" json:"run_id,omitempty"`
	WorkflowName string `bun:"workflow_name" json:"workflow_name,omitempty"`

	// Result
	Processed   bool       `bun:"processed,default:false" json:"processed"`
	ProcessedAt *time.Time `bun:"processed_at" json:"processed_at,omitempty"`
	Error       string     `bun:"error" json:"error,omitempty"`

	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// Schedule represents a workflow schedule
type Schedule struct {
	bun.BaseModel `bun:"table:schedules,alias:sch"`

	ID           string                 `bun:"id,pk,type:text" json:"id"`
	Name         string                 `bun:"name,notnull" json:"name"`
	WorkflowName string                 `bun:"workflow_name,notnull" json:"workflow_name"`
	WorkflowPath string                 `bun:"workflow_path,notnull" json:"workflow_path"`
	TriggerName  string                 `bun:"trigger_name,notnull" json:"trigger_name"`
	TriggerType  string                 `bun:"trigger_type,notnull" json:"trigger_type"`
	Schedule     string                 `bun:"schedule" json:"schedule,omitempty"`
	EventTopic   string                 `bun:"event_topic" json:"event_topic,omitempty"`
	WatchPath    string                 `bun:"watch_path" json:"watch_path,omitempty"`
	InputConfig  map[string]interface{} `bun:"input_config,type:json" json:"input_config,omitempty"`
	IsEnabled    bool                   `bun:"is_enabled,default:true" json:"is_enabled"`
	LastRun      *time.Time             `bun:"last_run" json:"last_run,omitempty"`
	NextRun      *time.Time             `bun:"next_run" json:"next_run,omitempty"`
	RunCount     int                    `bun:"run_count,default:0" json:"run_count"`
	CreatedAt    time.Time              `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time              `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// Event topic constants
const (
	TopicRunStarted        = "run.started"
	TopicRunCompleted      = "run.completed"
	TopicRunFailed         = "run.failed"
	TopicAssetDiscovered   = "asset.discovered"
	TopicAssetUpdated      = "asset.updated"
	TopicWebhookReceived   = "webhook.received"
	TopicScheduleTriggered = "schedule.triggered"
	TopicStepCompleted     = "step.completed"
	TopicStepFailed        = "step.failed"
)

// Asset represents an HTTP endpoint/asset discovered during scanning
type Asset struct {
	bun.BaseModel `bun:"table:assets,alias:as"`

	ID         int64  `bun:"id,pk,autoincrement" json:"id"`
	Workspace  string `bun:"workspace,notnull" json:"workspace"`
	AssetValue string `bun:"asset_value,notnull" json:"asset_value"`

	// HTTP data
	URL    string `bun:"url" json:"url,omitempty"`
	Input  string `bun:"input" json:"input,omitempty"`
	Scheme string `bun:"scheme" json:"scheme,omitempty"`
	Method string `bun:"method" json:"method,omitempty"`
	Path   string `bun:"path" json:"path,omitempty"`

	// Response data
	StatusCode    int    `bun:"status_code" json:"status_code,omitempty"`
	ContentType   string `bun:"content_type" json:"content_type,omitempty"`
	ContentLength int64  `bun:"content_length" json:"content_length,omitempty"`
	Title         string `bun:"title" json:"title,omitempty"`
	Words         int    `bun:"words" json:"words,omitempty"`
	Lines         int    `bun:"lines" json:"lines,omitempty"`

	// Network data
	HostIP     string   `bun:"host_ip" json:"host_ip,omitempty"`
	DnsRecords []string `bun:"dns_records,type:json" json:"a,omitempty"`
	TLS        string   `bun:"tls" json:"tls,omitempty"`

	// Metadata
	AssetType            string   `bun:"asset_type" json:"asset_type,omitempty"`
	Technologies         []string `bun:"technologies,type:json" json:"tech,omitempty"`
	ResponseTime         string   `bun:"response_time" json:"time,omitempty"`
	Labels               string   `bun:"labels" json:"remarks,omitempty"`
	Source               string   `bun:"source" json:"source,omitempty"`               // e.g., "httpx", "nuclei"
	RawJsonData          string   `bun:"raw_json_data" json:"raw_json_data,omitempty"` // Original JSON
	RawResponse          string   `bun:"raw_response" json:"raw_response,omitempty"`
	ScreenshotBase64Data string   `bun:"screenshot_base64_data" json:"screenshot_base64_data,omitempty"`

	// Timestamps
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	LastSeenAt time.Time `bun:"last_seen_at" json:"last_seen_at,omitempty"`
}

// Workspace represents a scan workspace with aggregated statistics
type Workspace struct {
	bun.BaseModel `bun:"table:workspaces,alias:ws"`

	ID         int64  `bun:"id,pk,autoincrement" json:"id"`
	Name       string `bun:"name,unique,notnull" json:"name"`
	LocalPath  string `bun:"local_path" json:"local_path"`
	DataSource string `bun:"data_source,default:'local'" json:"data_source"` // local, cloud, imported

	// Asset statistics
	TotalAssets     int `bun:"total_assets,default:0" json:"total_assets"`
	TotalSubdomains int `bun:"total_subdomains,default:0" json:"total_subdomains"`
	TotalURLs       int `bun:"total_urls,default:0" json:"total_urls"`
	TotalVulns      int `bun:"total_vulns,default:0" json:"total_vulns"`
	TotalIPs        int `bun:"total_ips,default:0" json:"total_ips"`
	TotalLinks      int `bun:"total_links,default:0" json:"total_links"`
	TotalContent    int `bun:"total_content,default:0" json:"total_content"`
	TotalArchive    int `bun:"total_archive,default:0" json:"total_archive"`

	// Vulnerability severity breakdown
	VulnCritical  int `bun:"vuln_critical,default:0" json:"vuln_critical"`
	VulnHigh      int `bun:"vuln_high,default:0" json:"vuln_high"`
	VulnMedium    int `bun:"vuln_medium,default:0" json:"vuln_medium"`
	VulnLow       int `bun:"vuln_low,default:0" json:"vuln_low"`
	VulnPotential int `bun:"vuln_potential,default:0" json:"vuln_potential"`

	// Risk and metadata
	RiskScore float64  `bun:"risk_score,default:0" json:"risk_score"`
	Tags      []string `bun:"tags,type:json" json:"tags"`

	// Run info
	LastRun     *time.Time `bun:"last_run" json:"last_run"`
	RunWorkflow string     `bun:"run_workflow" json:"run_workflow"`

	// State file paths
	StateExecutionLog   string `bun:"state_execution_log" json:"state_execution_log,omitempty"`
	StateCompletedFile  string `bun:"state_completed_file" json:"state_completed_file,omitempty"`
	StateWorkflowFile   string `bun:"state_workflow_file" json:"state_workflow_file,omitempty"`
	StateWorkflowFolder string `bun:"state_workflow_folder" json:"state_workflow_folder,omitempty"`

	// Timestamps
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// WorkflowMeta stores workflow metadata in database for faster querying
type WorkflowMeta struct {
	bun.BaseModel `bun:"table:workflow_meta,alias:wm"`

	ID          int64    `bun:"id,pk,autoincrement" json:"id"`
	Name        string   `bun:"name,unique,notnull" json:"name"`
	Kind        string   `bun:"kind,notnull" json:"kind"` // "module" or "flow"
	Description string   `bun:"description" json:"description"`
	FilePath    string   `bun:"file_path,notnull" json:"file_path"`
	Checksum    string   `bun:"checksum" json:"checksum"` // SHA256 for change detection
	Tags        []string `bun:"tags,type:json" json:"tags"`
	Hidden      bool     `bun:"hidden,default:false" json:"hidden"`

	// Metadata
	StepCount   int    `bun:"step_count" json:"step_count"`
	ModuleCount int    `bun:"module_count" json:"module_count"`
	ParamsJSON  string `bun:"params_json" json:"params_json"` // Serialized params

	// Timestamps
	IndexedAt time.Time `bun:"indexed_at,notnull" json:"indexed_at"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// Vulnerability represents a security vulnerability discovered during scanning
type Vulnerability struct {
	bun.BaseModel `bun:"table:vulnerabilities,alias:vl"`

	ID                 int64    `bun:"id,pk,autoincrement" json:"id"`
	Workspace          string   `bun:"workspace,notnull" json:"workspace"`
	VulnInfo           string   `bun:"vuln_info" json:"vuln_info"`
	VulnTitle          string   `bun:"vuln_title" json:"vuln_title"`
	VulnDesc           string   `bun:"vuln_desc" json:"vuln_desc"`
	VulnPOC            string   `bun:"vuln_poc" json:"vuln_poc"`
	Severity           string   `bun:"severity" json:"severity"`
	Confidence         string   `bun:"confidence" json:"confidence"` // Certain, Firm, Tentative, Manual Review Required
	AssetType          string   `bun:"asset_type" json:"asset_type"`
	AssetValue         string   `bun:"asset_value" json:"asset_value"`
	Tags               []string `bun:"tags,type:json" json:"tags,omitempty"`
	DetailHTTPRequest  string   `bun:"detail_http_request" json:"detail_http_request"`
	DetailHTTPResponse string   `bun:"detail_http_response" json:"detail_http_response"`
	RawVulnJSON        string   `bun:"raw_vuln_json" json:"raw_vuln_json"`

	// Timestamps
	CreatedAt  time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt  time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	LastSeenAt time.Time `bun:"last_seen_at" json:"last_seen_at,omitempty"`
}

// AssetDiffSnapshot stores a point-in-time diff calculation for assets
type AssetDiffSnapshot struct {
	bun.BaseModel `bun:"table:asset_diffs,alias:ad"`

	ID            int64     `bun:"id,pk,autoincrement" json:"id"`
	WorkspaceName string    `bun:"workspace_name,notnull" json:"workspace_name"`
	FromTime      time.Time `bun:"from_time,notnull" json:"from_time"`
	ToTime        time.Time `bun:"to_time,notnull" json:"to_time"`
	TotalAdded    int       `bun:"total_added" json:"total_added"`
	TotalRemoved  int       `bun:"total_removed" json:"total_removed"`
	TotalChanged  int       `bun:"total_changed" json:"total_changed"`
	DiffData      string    `bun:"diff_data,type:text" json:"diff_data"` // JSON serialized AssetDiff
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// VulnDiffSnapshot stores a point-in-time vulnerability diff calculation
type VulnDiffSnapshot struct {
	bun.BaseModel `bun:"table:vuln_diffs,alias:vd"`

	ID            int64     `bun:"id,pk,autoincrement" json:"id"`
	WorkspaceName string    `bun:"workspace_name,notnull" json:"workspace_name"`
	FromTime      time.Time `bun:"from_time,notnull" json:"from_time"`
	ToTime        time.Time `bun:"to_time,notnull" json:"to_time"`
	TotalAdded    int       `bun:"total_added" json:"total_added"`
	TotalRemoved  int       `bun:"total_removed" json:"total_removed"`
	TotalChanged  int       `bun:"total_changed" json:"total_changed"`
	DiffData      string    `bun:"diff_data,type:text" json:"diff_data"` // JSON serialized VulnerabilityDiff
	CreatedAt     time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
