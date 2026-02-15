package distributed

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/heuristics"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
)

// WorkerOptions holds optional configuration for creating a new Worker.
type WorkerOptions struct {
	GetPublicIP bool
	Alias       string
	SSHEnabled  bool
	SSHKeysPath string
}

// Worker represents a worker node that processes tasks
type Worker struct {
	ID       string
	Hostname string
	client   *Client
	config   *config.Config
	executor *executor.Executor
	loader   *parser.Loader
	printer  *terminal.Printer

	// Cleanup function for distributed hooks
	unregisterHooks func()

	// Metadata
	ipAddress   string
	publicIP    string
	sshEnabled  bool
	sshKeysPath string
	alias       string

	// Stats
	tasksComplete int
	tasksFailed   int
}

// NewWorker creates a new worker node
func NewWorker(cfg *config.Config, opts *WorkerOptions) (*Worker, error) {
	if opts == nil {
		opts = &WorkerOptions{}
	}

	client, err := NewClientFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("wosm-%s", uuid.NewString()[:8])

	exec := executor.NewExecutor()
	loader := parser.NewLoader(cfg.WorkflowsPath)
	exec.SetLoader(loader)

	p := terminal.NewPrinter()

	w := &Worker{
		ID:          workerID,
		Hostname:    hostname,
		client:      client,
		config:      cfg,
		executor:    exec,
		loader:      loader,
		printer:     p,
		ipAddress:   getOutboundIP(),
		sshEnabled:  opts.SSHEnabled,
		sshKeysPath: opts.SSHKeysPath,
		alias:       opts.Alias,
	}

	if opts.GetPublicIP {
		w.publicIP = fetchPublicIP()
		if w.publicIP != "" {
			p.Info("Detected public IP: %s", terminal.Cyan(w.publicIP))
		} else {
			p.Warning("Could not detect public IP")
		}
	}

	// Default alias: wosm-<public-ip> or wosm-<ip-address>
	if w.alias == "" {
		if w.publicIP != "" {
			w.alias = fmt.Sprintf("wosm-%s", w.publicIP)
		} else if w.ipAddress != "" {
			w.alias = fmt.Sprintf("wosm-%s", w.ipAddress)
		}
	}

	return w, nil
}

// getOutboundIP returns the preferred outbound IP address of the machine.
// It uses a UDP dial to 8.8.8.8:80 (no actual packet is sent) to determine the source address.
func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer func() { _ = conn.Close() }()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

// fetchPublicIP fetches the public IP from ipinfo.io.
func fetchPublicIP() string {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", "https://ipinfo.io/ip", nil)
	if err != nil {
		return ""
	}
	req.Header.Set("User-Agent", core.DefaultUA)
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}

// Run starts the worker loop
func (w *Worker) Run(ctx context.Context) error {
	// Test connection
	if err := w.client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Register worker
	if err := w.register(ctx); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	// Set worker mode in config
	config.SetWorkerMode(true, w.ID)

	// Register distributed hooks for database writes
	w.registerDistributedHooks()
	defer w.unregisterDistributedHooks()

	w.printer.Success("Worker %s joined successfully", terminal.Cyan(w.ID))
	w.printer.Info("Waiting for tasks...")

	// Start heartbeat goroutine
	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	defer cancelHeartbeat()
	go w.heartbeatLoop(heartbeatCtx)

	// Start execute listener goroutine for per-worker execute requests
	executeCtx, cancelExecute := context.WithCancel(ctx)
	defer cancelExecute()
	go w.executeListenerLoop(executeCtx)

	// Main task loop
	for {
		select {
		case <-ctx.Done():
			w.printer.Info("Worker %s shutting down...", terminal.Cyan(w.ID))
			w.cleanup(context.Background())
			return nil
		default:
			if err := w.processNextTask(ctx); err != nil {
				// Suppress context-canceled errors during shutdown
				if ctx.Err() != nil {
					continue
				}
				w.printer.Warning("Error processing task: %s", err)
				time.Sleep(time.Second) // Brief pause before retrying
			}
		}
	}
}

// register registers the worker with the master
func (w *Worker) register(ctx context.Context) error {
	info := &WorkerInfo{
		ID:            w.ID,
		Hostname:      w.Hostname,
		Status:        "idle",
		JoinedAt:      time.Now(),
		LastHeartbeat: time.Now(),
		IPAddress:     w.ipAddress,
		PublicIP:      w.publicIP,
		SSHEnabled:    w.sshEnabled,
		SSHKeysPath:   w.sshKeysPath,
		Alias:         w.alias,
	}

	if err := w.client.RegisterWorker(ctx, info); err != nil {
		return err
	}

	return w.client.UpdateWorkerHeartbeat(ctx, w.ID)
}

// heartbeatLoop sends periodic heartbeats
func (w *Worker) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.client.UpdateWorkerHeartbeat(ctx, w.ID); err != nil {
				w.printer.Warning("Failed to send heartbeat: %s", err)
			}
		}
	}
}

// processNextTask waits for and processes the next task
func (w *Worker) processNextTask(ctx context.Context) error {
	// Block waiting for a task
	task, err := w.client.PopTask(ctx, TaskPollTimeout)
	if err != nil {
		return err
	}
	if task == nil {
		return nil // Timeout, no task available
	}

	w.printer.Info("Received task %s: %s -> %s",
		terminal.Cyan(task.ID), terminal.Yellow(task.WorkflowName), terminal.Green(task.Target))

	// Mark task as running
	task.MarkRunning(w.ID)
	if err := w.client.SetTaskRunning(ctx, task); err != nil {
		w.printer.Warning("Failed to mark task running: %s", err)
	}

	// Update worker status
	w.updateStatus(ctx, "busy", task.ID)

	// Execute the task
	result := w.executeTask(ctx, task)

	// Report result
	if err := w.client.SetTaskResult(ctx, result); err != nil {
		w.printer.Warning("Failed to report task result: %s", err)
	}

	// Remove from running
	if err := w.client.RemoveTaskRunning(ctx, task.ID); err != nil {
		w.printer.Warning("Failed to remove task from running: %s", err)
	}

	// Update stats and status
	if result.Status == TaskStatusCompleted {
		w.tasksComplete++
		w.printer.Success("Task %s completed", terminal.Cyan(task.ID))
	} else {
		w.tasksFailed++
		w.printer.Error("Task %s failed: %s", terminal.Cyan(task.ID), result.Error)
	}
	w.updateStatus(ctx, "idle", "")

	return nil
}

// executeTask executes a workflow task
func (w *Worker) executeTask(ctx context.Context, task *Task) *TaskResult {
	result := &TaskResult{
		TaskID:      task.ID,
		CompletedAt: time.Now(),
	}

	// If task uses a file input, ensure the file exists locally
	if task.InputIsFile && task.InputFilePath != "" {
		if _, err := os.Stat(task.InputFilePath); os.IsNotExist(err) {
			w.printer.Warning("Input file %s not found locally, attempting rsync from master", task.InputFilePath)
			if syncErr := w.syncFileFromMaster(ctx, task.InputFilePath); syncErr != nil {
				result.Status = TaskStatusFailed
				result.Error = fmt.Sprintf("input file not available: %v", syncErr)
				return result
			}
		}
	}

	// Load workflow
	workflow, err := w.loader.LoadWorkflow(task.WorkflowName)
	if err != nil {
		result.Status = TaskStatusFailed
		result.Error = fmt.Sprintf("failed to load workflow: %v", err)
		return result
	}

	// Convert params to string map
	params := make(map[string]string)
	params["target"] = task.Target
	for k, v := range task.Params {
		if s, ok := v.(string); ok {
			params[k] = s
		}
	}

	// Create run record for distributed tracking
	now := time.Now()
	runUUID := uuid.New().String()
	paramsInterface := make(map[string]interface{})
	for k, v := range params {
		paramsInterface[k] = v
	}
	totalSteps := countWorkflowSteps(workflow, w.loader)

	run := &database.Run{
		RunUUID:      runUUID,
		WorkflowName: workflow.Name,
		WorkflowKind: string(workflow.Kind),
		Target:       task.Target,
		Params:       paramsInterface,
		Status:       "running",
		TriggerType:  "distributed",
		StartedAt:    &now,
		TotalSteps:   totalSteps,
		Workspace:    computeWorkspace(task.Target),
		RunPriority:  "high",
		RunMode:      "distributed",
		HooksEnabled: workflow.HookCount() > 0,
	}
	// Goes through distributed hooks → Redis → master DB
	_ = database.CreateRun(ctx, run)

	// Wire up executor for run tracking
	w.executor.SetDBRunUUID(runUUID)

	// Execute based on workflow kind
	var wfResult *core.WorkflowResult
	if workflow.IsFlow() {
		wfResult, err = w.executor.ExecuteFlow(ctx, workflow, params, w.config)
	} else {
		wfResult, err = w.executor.ExecuteModule(ctx, workflow, params, w.config)
	}

	// Determine final status and error message
	var finalStatus string
	var errorMsg string
	if err != nil {
		finalStatus = "failed"
		errorMsg = err.Error()
		result.Status = TaskStatusFailed
		result.Error = err.Error()
	} else if wfResult.Status == core.RunStatusFailed {
		finalStatus = "failed"
		result.Status = TaskStatusFailed
		if wfResult.Error != nil {
			errorMsg = wfResult.Error.Error()
			result.Error = errorMsg
		} else {
			errorMsg = "workflow execution failed"
			result.Error = errorMsg
		}
	} else {
		finalStatus = "completed"
		result.Status = TaskStatusCompleted
		result.Exports = wfResult.Exports
	}

	// Send final status update to master via Redis hooks
	completedAt := time.Now()
	run.Status = finalStatus
	run.ErrorMessage = errorMsg
	run.CompletedAt = &completedAt
	run.UpdatedAt = completedAt
	if finalStatus == "completed" {
		run.CompletedSteps = totalSteps
	}
	_ = database.CreateRun(ctx, run) // upsert — master matches by run_uuid

	result.CompletedAt = completedAt
	return result
}

// updateStatus updates the worker's status in Redis
func (w *Worker) updateStatus(ctx context.Context, status string, taskID string) {
	info := &WorkerInfo{
		ID:            w.ID,
		Hostname:      w.Hostname,
		Status:        status,
		CurrentTaskID: taskID,
		JoinedAt:      time.Now(), // This will be overwritten, but we need a value
		LastHeartbeat: time.Now(),
		TasksComplete: w.tasksComplete,
		TasksFailed:   w.tasksFailed,
		IPAddress:     w.ipAddress,
		PublicIP:      w.publicIP,
		SSHEnabled:    w.sshEnabled,
		SSHKeysPath:   w.sshKeysPath,
		Alias:         w.alias,
	}
	if err := w.client.RegisterWorker(ctx, info); err != nil {
		w.printer.Warning("Failed to update worker status: %s", err)
	}
}

// cleanup removes the worker from the registry
func (w *Worker) cleanup(ctx context.Context) {
	w.printer.Info("Cleaning up worker %s...", terminal.Cyan(w.ID))
	if err := w.client.RemoveWorker(ctx, w.ID); err != nil {
		w.printer.Warning("Failed to remove worker: %s", err)
	}
	w.client.Close()
}

// GetID returns the worker ID
func (w *Worker) GetID() string {
	return w.ID
}

// GetClient returns the Redis client
func (w *Worker) GetClient() *Client {
	return w.client
}

// syncFileFromMaster attempts to sync a file from the master node via the data queue.
// It sends a sync request and waits briefly, but file availability is best-effort.
func (w *Worker) syncFileFromMaster(ctx context.Context, filePath string) error {
	// Send a sync request to the master via the execute queue
	req := buildExecuteRequest("sync", filePath, "", filePath, "", "master", "")
	if err := w.client.PushData(ctx, KeyDataExecute, "execute", req, w.ID); err != nil {
		return fmt.Errorf("failed to send sync request: %w", err)
	}

	// Wait a short time for the sync to complete
	syncCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-syncCtx.Done():
			return fmt.Errorf("timeout waiting for file sync of %s", filePath)
		case <-ticker.C:
			if _, err := os.Stat(filePath); err == nil {
				w.printer.Success("File %s synced successfully", filePath)
				return nil
			}
		}
	}
}

// =============================================================================
// Execute Listener (per-worker execute queue)
// =============================================================================

// executeListenerLoop polls the per-worker execute queue for requests routed by the master.
func (w *Worker) executeListenerLoop(ctx context.Context) {
	key := KeyDataExecuteForWorker(w.ID)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			envelope, err := w.client.PopData(ctx, key, TaskPollTimeout)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				w.printer.Warning("Execute listener error: %s", err)
				time.Sleep(time.Second)
				continue
			}
			if envelope == nil {
				continue
			}
			w.processExecuteRequest(ctx, envelope)
		}
	}
}

// processExecuteRequest handles an execute request received on the worker's execute queue.
func (w *Worker) processExecuteRequest(ctx context.Context, envelope *DataEnvelope) {
	var req ExecuteRequest
	if err := json.Unmarshal(envelope.Data, &req); err != nil {
		w.printer.Warning("Failed to unmarshal execute request: %s", err)
		return
	}

	executeType := req.ExecuteType
	if executeType == "" {
		executeType = req.Action
	}

	w.printer.Info("Processing execute request: type=%s from=%s", terminal.Yellow(executeType), terminal.Cyan(envelope.WorkerID))

	switch executeType {
	case "func":
		expr := req.Data
		if expr == "" {
			expr = req.Expr
		}
		execCtx := executor.BuildBuiltinVariables(w.config, nil)
		registry := functions.NewRegistry()
		if _, err := registry.Execute(expr, execCtx); err != nil {
			w.printer.Warning("Execute func failed: %s (expr: %s)", err, expr)
		}

	case "run":
		workflow := req.Data
		if workflow == "" {
			workflow = req.Workflow
		}
		task := NewTask(uuid.NewString()[:8], workflow, "module", req.Target, nil)
		result := w.executeTask(ctx, task)
		if result.Status == TaskStatusFailed {
			w.printer.Warning("Execute run failed: %s", result.Error)
		}

	case "bash":
		command := req.Data
		if command == "" {
			command = req.Expr
		}
		// @NOTE: This is intentional - execute requests come from trusted workflow YAML files
		// via the distributed system. The master routes requests from run_on_worker() calls.
		cmd := exec.CommandContext(ctx, "sh", "-c", command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			w.printer.Warning("Execute bash failed: %s (output: %s)", err, string(output))
		}

	default:
		w.printer.Warning("Unknown execute type: %s", executeType)
	}
}

// =============================================================================
// Helpers
// =============================================================================

// countWorkflowSteps counts the total number of steps in a workflow.
// For flows, it sums the steps of all referenced modules.
func countWorkflowSteps(workflow *core.Workflow, loader *parser.Loader) int {
	if workflow.IsFlow() && loader != nil {
		total := 0
		for _, mod := range workflow.Modules {
			m, err := loader.LoadWorkflow(mod.Name)
			if err == nil {
				total += len(m.Steps)
			}
		}
		return total
	}
	return len(workflow.Steps)
}

// computeWorkspace derives a workspace name from the target using heuristic analysis.
func computeWorkspace(target string) string {
	info, err := heuristics.Analyze(target, "basic")
	if err == nil && info != nil && info.RootDomain != "" {
		return info.RootDomain
	}
	return target
}

// =============================================================================
// Data Queue Methods - Send data to master via Redis
// =============================================================================

// SendRunData sends run data to the master via Redis queue
func (w *Worker) SendRunData(ctx context.Context, run *database.Run) error {
	return w.client.PushData(ctx, KeyDataRuns, "run", run, w.ID)
}

// SendStepResult sends step result data to the master via Redis queue
func (w *Worker) SendStepResult(ctx context.Context, step *database.StepResult) error {
	return w.client.PushData(ctx, KeyDataSteps, "step", step, w.ID)
}

// SendEventLog sends event log data to the master via Redis queue
func (w *Worker) SendEventLog(ctx context.Context, eventLog *database.EventLog) error {
	return w.client.PushData(ctx, KeyDataEvents, "event", eventLog, w.ID)
}

// SendArtifact sends artifact data to the master via Redis queue
func (w *Worker) SendArtifact(ctx context.Context, artifact *database.Artifact) error {
	return w.client.PushData(ctx, KeyDataArtifacts, "artifact", artifact, w.ID)
}

// SendExecuteRequest sends an execute request to the master via Redis queue
func (w *Worker) SendExecuteRequest(ctx context.Context, action, expr, workflow, target, params, targetRole, targetScope string) error {
	req := buildExecuteRequest(action, expr, workflow, target, params, targetRole, targetScope)
	return w.client.PushData(ctx, KeyDataExecute, "execute", req, w.ID)
}

// buildExecuteRequest creates an ExecuteRequest with both new and legacy fields populated.
func buildExecuteRequest(action, expr, workflow, target, params, targetRole, targetScope string) *ExecuteRequest {
	if targetRole == "" {
		targetRole = "master"
	}
	data := expr
	if action == "run" {
		data = workflow
	}
	return &ExecuteRequest{
		ExecuteType: action,
		TargetRole:  targetRole,
		Data:        data,
		Target:      target,
		Params:      params,
		TargetScope: targetScope,
		// Legacy fields for backward compatibility
		Action:   action,
		Expr:     expr,
		Workflow: workflow,
	}
}

// =============================================================================
// Distributed Hooks Registration
// =============================================================================

// registerDistributedHooks registers callbacks for database writes to use Redis queues
func (w *Worker) registerDistributedHooks() {
	w.unregisterHooks = RegisterDistributedHooksFromClient(w.client, w.ID)
	w.printer.Info("Registered distributed hooks for database writes")
}

// unregisterDistributedHooks removes the distributed hooks
func (w *Worker) unregisterDistributedHooks() {
	if w.unregisterHooks != nil {
		w.unregisterHooks()
	}
	w.printer.Info("Unregistered distributed hooks")
}

// RegisterDistributedHooksFromClient registers distributed hooks using a bare
// Client and workerID, without requiring the full Worker struct. This is useful
// for one-shot operations (e.g., worker eval) that need run_on_master() routing
// without the full worker lifecycle (heartbeat, task loop, master registration).
// Returns a cleanup function that unregisters all hooks.
func RegisterDistributedHooksFromClient(client *Client, workerID string) func() {
	hooks := &database.DistributedHooks{
		SendRun: func(ctx context.Context, run *database.Run) error {
			return client.PushData(ctx, KeyDataRuns, "run", run, workerID)
		},
		SendStepResult: func(ctx context.Context, step *database.StepResult) error {
			return client.PushData(ctx, KeyDataSteps, "step", step, workerID)
		},
		SendEventLog: func(ctx context.Context, event *database.EventLog) error {
			return client.PushData(ctx, KeyDataEvents, "event", event, workerID)
		},
		SendArtifact: func(ctx context.Context, artifact *database.Artifact) error {
			return client.PushData(ctx, KeyDataArtifacts, "artifact", artifact, workerID)
		},
		ShouldUseRedis: func() bool {
			return config.ShouldUseRedisDataQueues()
		},
	}
	database.RegisterDistributedHooks(hooks)

	// Register execute hooks for run_on_master() and run_on_worker() functions
	execHooks := &functions.ExecuteHooks{
		SendExecuteRequest: func(ctx context.Context, action, expr, workflow, target, params, targetRole, targetScope string) error {
			req := buildExecuteRequest(action, expr, workflow, target, params, targetRole, targetScope)
			return client.PushData(ctx, KeyDataExecute, "execute", req, workerID)
		},
		ShouldUseRedis: func() bool {
			return config.ShouldUseRedisDataQueues()
		},
		ResolveWorkerSSH: func(ctx context.Context, identifier string) (*functions.WorkerSSHInfo, error) {
			// Try by ID first
			w, err := client.GetWorker(ctx, identifier)
			if err != nil {
				return nil, fmt.Errorf("failed to look up worker %q: %w", identifier, err)
			}
			// Try by alias if not found by ID
			if w == nil {
				w, err = client.GetWorkerByAlias(ctx, identifier)
				if err != nil {
					return nil, fmt.Errorf("failed to look up worker by alias %q: %w", identifier, err)
				}
			}
			// Try by PublicIP if still not found
			if w == nil {
				workers, err := client.GetAllWorkers(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to list workers: %w", err)
				}
				for _, cand := range workers {
					if cand.PublicIP == identifier || cand.IPAddress == identifier {
						w = cand
						break
					}
				}
			}
			if w == nil {
				return nil, fmt.Errorf("worker %q not found", identifier)
			}
			if !w.SSHEnabled {
				return nil, fmt.Errorf("worker %q does not have SSH enabled", identifier)
			}
			host := w.PublicIP
			if host == "" {
				host = w.IPAddress
			}
			return &functions.WorkerSSHInfo{
				ID:      w.ID,
				Host:    host,
				User:    "root",
				KeyPath: w.SSHKeysPath,
				Alias:   w.Alias,
				Port:    22,
			}, nil
		},
	}
	functions.RegisterExecuteHooks(execHooks)

	return func() {
		database.UnregisterDistributedHooks()
		functions.UnregisterExecuteHooks()
		config.SetWorkerMode(false, "")
	}
}
