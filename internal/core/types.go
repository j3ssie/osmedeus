package core

import (
	"encoding/json"
	"time"
)

// WorkflowKind represents the type of workflow
type WorkflowKind string

const (
	KindModule WorkflowKind = "module"
	KindFlow   WorkflowKind = "flow"
)

// StepType represents the type of step
type StepType string

const (
	StepTypeBash       StepType = "bash"
	StepTypeFunction   StepType = "function"
	StepTypeParallel   StepType = "parallel-steps"
	StepTypeForeach    StepType = "foreach"
	StepTypeRemoteBash StepType = "remote-bash"
	StepTypeHTTP       StepType = "http"
	StepTypeLLM        StepType = "llm"
)

// TriggerType represents trigger types
type TriggerType string

const (
	TriggerCron   TriggerType = "cron"
	TriggerEvent  TriggerType = "event"
	TriggerWatch  TriggerType = "watch"
	TriggerManual TriggerType = "manual"
)

// VariableType for dependency validation
type VariableType string

const (
	VarTypeDomain    VariableType = "domain"
	VarTypePath      VariableType = "path"
	VarTypeNumber    VariableType = "number"
	VarTypeFile      VariableType = "file"
	VarTypeFolder    VariableType = "folder"
	VarTypeString    VariableType = "string"
	VarTypeSubdomain VariableType = "subdomain"
	VarTypeURL       VariableType = "url"
	VarTypeCIDR      VariableType = "cidr"
	VarTypeRepo      VariableType = "repo"
)

type TargetType string

const (
	TargetTypeDomain    TargetType = "domain"
	TargetTypeSubdomain TargetType = "subdomain"
	TargetTypeURL       TargetType = "url"
	TargetTypeCIDR      TargetType = "cidr"
	TargetTypeRepo      TargetType = "repo"
	TargetTypePath      TargetType = "path"
	TargetTypeFile      TargetType = "file"
	TargetTypeFolder    TargetType = "folder"
	TargetTypeNumber    TargetType = "number"
	TargetTypeString    TargetType = "string"
)

// ActionType for on_success/on_error handlers
type ActionType string

const (
	ActionLog      ActionType = "log"
	ActionAbort    ActionType = "abort"
	ActionContinue ActionType = "continue"
	ActionExport   ActionType = "export"
	ActionRun      ActionType = "run"
	ActionNotify   ActionType = "notify"
)

// StepStatus represents the status of a step execution
type StepStatus string

const (
	StepStatusPending StepStatus = "pending"
	StepStatusRunning StepStatus = "running"
	StepStatusSuccess StepStatus = "success"
	StepStatusFailed  StepStatus = "failed"
	StepStatusSkipped StepStatus = "skipped"
)

// RunnerType represents the execution environment for workflows
type RunnerType string

const (
	RunnerTypeHost   RunnerType = "host"   // Execute on local machine (default)
	RunnerTypeDocker RunnerType = "docker" // Execute in Docker container
	RunnerTypeSSH    RunnerType = "ssh"    // Execute on remote machine via SSH
)

// RunStatus represents the status of a run
type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
	RunStatusSkipped   RunStatus = "skipped"
)

// StepResult holds step execution result
type StepResult struct {
	StepName  string
	Status    StepStatus
	Output    string
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Exports   map[string]interface{}
	NextStep  string // from decision routing
	LogFile   string
}

// WorkflowResult holds workflow execution result
type WorkflowResult struct {
	WorkflowName string
	WorkflowKind WorkflowKind
	RunID        string
	Target       string
	Status       RunStatus
	StartTime    time.Time
	EndTime      time.Time
	Steps        []*StepResult
	Artifacts    []string
	Exports      map[string]interface{}
	Error        error
	Message      string // Optional message (e.g., for skipped status)
}

// Event represents a system event for triggers
// Topics follow the format: <component>.<event_type>
// Examples: webhook.received, assets.new, db.change, watch.files
type Event struct {
	Topic      string                 `json:"topic" yaml:"topic"`         // e.g., "webhook.received", "assets.new"
	ID         string                 `json:"id" yaml:"id"`               // UUID of the event
	Name       string                 `json:"name" yaml:"name"`           // e.g., "vulnerability.discovered"
	Source     string                 `json:"source" yaml:"source"`       // e.g., "nuclei", "httpx"
	Data       string                 `json:"data" yaml:"data"`           // JSON string payload
	DataType   string                 `json:"data_type" yaml:"data_type"` // e.g., "endpoint", "vulnerability"
	Timestamp  time.Time              `json:"timestamp" yaml:"timestamp"` // When the event occurred
	ParsedData map[string]interface{} `json:"-" yaml:"-"`                 // Parsed JSON for filter evaluation
}

// ParseData parses the JSON data string into ParsedData map
func (e *Event) ParseData() error {
	if e.Data == "" {
		e.ParsedData = make(map[string]interface{})
		return nil
	}
	return json.Unmarshal([]byte(e.Data), &e.ParsedData)
}

// GetDataField retrieves a field from the parsed data
func (e *Event) GetDataField(field string) interface{} {
	if e.ParsedData == nil {
		if err := e.ParseData(); err != nil {
			return nil
		}
	}
	return e.ParsedData[field]
}
