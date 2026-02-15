package distributed

import (
	"encoding/json"
	"time"
)

// TaskStatus represents the status of a distributed task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// Task represents a distributed scan task
type Task struct {
	ID           string                 `json:"id"`
	ScanID       string                 `json:"scan_id,omitempty"`
	WorkflowName string                 `json:"workflow_name"`
	WorkflowKind string                 `json:"workflow_kind"` // "module" or "flow"
	Target       string                 `json:"target"`
	Params       map[string]interface{} `json:"params,omitempty"`
	Status       TaskStatus             `json:"status"`
	WorkerID     string                 `json:"worker_id,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// TaskResult represents the result of a completed task
type TaskResult struct {
	TaskID      string                 `json:"task_id"`
	Status      TaskStatus             `json:"status"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Exports     map[string]interface{} `json:"exports,omitempty"`
	CompletedAt time.Time              `json:"completed_at"`
}

// WorkerInfo represents information about a worker node
type WorkerInfo struct {
	ID            string    `json:"id"`
	Hostname      string    `json:"hostname"`
	Status        string    `json:"status"` // "idle", "busy", "offline"
	CurrentTaskID string    `json:"current_task_id,omitempty"`
	JoinedAt      time.Time `json:"joined_at"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	TasksComplete int       `json:"tasks_complete"`
	TasksFailed   int       `json:"tasks_failed"`
	IPAddress     string    `json:"ip_address,omitempty"`
	PublicIP      string    `json:"public_ip,omitempty"`
	SSHEnabled    bool      `json:"ssh_enabled,omitempty"`
	SSHKeysPath   string    `json:"ssh_keys_path,omitempty"`
	Alias         string    `json:"alias,omitempty"`
}

// NewTask creates a new task with the given parameters
func NewTask(id, workflowName, workflowKind, target string, params map[string]interface{}) *Task {
	return &Task{
		ID:           id,
		WorkflowName: workflowName,
		WorkflowKind: workflowKind,
		Target:       target,
		Params:       params,
		Status:       TaskStatusPending,
		CreatedAt:    time.Now(),
	}
}

// MarshalJSON serializes a task to JSON
func (t *Task) MarshalJSON() ([]byte, error) {
	type Alias Task
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	})
}

// UnmarshalTask deserializes a task from JSON
func UnmarshalTask(data []byte) (*Task, error) {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// MarshalJSON serializes a task result to JSON
func (r *TaskResult) MarshalJSON() ([]byte, error) {
	type Alias TaskResult
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// UnmarshalTaskResult deserializes a task result from JSON
func UnmarshalTaskResult(data []byte) (*TaskResult, error) {
	var result TaskResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// MarshalJSON serializes worker info to JSON
func (w *WorkerInfo) MarshalJSON() ([]byte, error) {
	type Alias WorkerInfo
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(w),
	})
}

// UnmarshalWorkerInfo deserializes worker info from JSON
func UnmarshalWorkerInfo(data []byte) (*WorkerInfo, error) {
	var info WorkerInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// MarkRunning marks the task as running with the given worker
func (t *Task) MarkRunning(workerID string) {
	t.Status = TaskStatusRunning
	t.WorkerID = workerID
	now := time.Now()
	t.StartedAt = &now
}

// MarkCompleted marks the task as completed
func (t *Task) MarkCompleted() {
	t.Status = TaskStatusCompleted
	now := time.Now()
	t.CompletedAt = &now
}

// ExecuteRequest represents a request from a worker to execute something on the master or another worker.
// ExecuteType "func" executes a utility function expression.
// ExecuteType "run" submits a workflow task to the pending queue.
// TargetRole controls where the request is executed: "master" or "worker".
type ExecuteRequest struct {
	ExecuteType string `json:"execute_type"`           // "func", "run", or "bash"
	TargetRole  string `json:"target_role"`            // "master" or "worker"
	Data        string `json:"data,omitempty"`         // function expression (for "func"), workflow name (for "run"), or command (for "bash")
	Target      string `json:"target,omitempty"`       // target (for "run")
	Params      string `json:"params,omitempty"`       // comma-separated key=value (for "run")
	TargetScope string `json:"target_scope,omitempty"` // For worker-targeted: "all", alias, worker ID, or public IP
	Action      string `json:"action,omitempty"`       // deprecated: use ExecuteType instead
	Expr        string `json:"expr,omitempty"`         // deprecated: use Data instead
	Workflow    string `json:"workflow,omitempty"`     // deprecated: use Data instead (for "run")
}

// MarkFailed marks the task as failed with an error message
func (t *Task) MarkFailed(err string) {
	t.Status = TaskStatusFailed
	t.Error = err
	now := time.Now()
	t.CompletedAt = &now
}
