package distributed

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// Worker represents a worker node that processes tasks
type Worker struct {
	ID       string
	Hostname string
	client   *Client
	config   *config.Config
	executor *executor.Executor
	loader   *parser.Loader
	logger   *zap.Logger
	printer  *terminal.Printer

	// Stats
	tasksComplete int
	tasksFailed   int
}

// NewWorker creates a new worker node
func NewWorker(cfg *config.Config) (*Worker, error) {
	client, err := NewClientFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("%s-%s", hostname, uuid.NewString()[:8])

	logger, _ := zap.NewProduction()
	exec := executor.NewExecutor()

	return &Worker{
		ID:       workerID,
		Hostname: hostname,
		client:   client,
		config:   cfg,
		executor: exec,
		loader:   parser.NewLoader(cfg.WorkflowsPath),
		logger:   logger,
		printer:  terminal.NewPrinter(),
	}, nil
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

	w.printer.Success("Worker %s joined successfully", w.ID)
	w.printer.Info("Waiting for tasks...")

	// Start heartbeat goroutine
	heartbeatCtx, cancelHeartbeat := context.WithCancel(ctx)
	defer cancelHeartbeat()
	go w.heartbeatLoop(heartbeatCtx)

	// Main task loop
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("worker shutting down", zap.String("worker_id", w.ID))
			w.cleanup(context.Background())
			return nil
		default:
			if err := w.processNextTask(ctx); err != nil {
				w.logger.Error("error processing task", zap.Error(err))
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
				w.logger.Warn("failed to send heartbeat", zap.Error(err))
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

	w.logger.Info("received task",
		zap.String("task_id", task.ID),
		zap.String("workflow", task.WorkflowName),
		zap.String("target", task.Target),
	)
	w.printer.Info("Received task %s: %s -> %s", task.ID, task.WorkflowName, task.Target)

	// Mark task as running
	task.MarkRunning(w.ID)
	if err := w.client.SetTaskRunning(ctx, task); err != nil {
		w.logger.Error("failed to mark task running", zap.Error(err))
	}

	// Update worker status
	w.updateStatus(ctx, "busy", task.ID)

	// Execute the task
	result := w.executeTask(ctx, task)

	// Report result
	if err := w.client.SetTaskResult(ctx, result); err != nil {
		w.logger.Error("failed to report task result", zap.Error(err))
	}

	// Remove from running
	if err := w.client.RemoveTaskRunning(ctx, task.ID); err != nil {
		w.logger.Error("failed to remove task from running", zap.Error(err))
	}

	// Update stats and status
	if result.Status == TaskStatusCompleted {
		w.tasksComplete++
		w.printer.Success("Task %s completed", task.ID)
	} else {
		w.tasksFailed++
		w.printer.Error("Task %s failed: %s", task.ID, result.Error)
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

	// Execute based on workflow kind
	var wfResult *core.WorkflowResult
	if workflow.IsFlow() {
		wfResult, err = w.executor.ExecuteFlow(ctx, workflow, params, w.config)
	} else {
		wfResult, err = w.executor.ExecuteModule(ctx, workflow, params, w.config)
	}

	if err != nil {
		result.Status = TaskStatusFailed
		result.Error = err.Error()
		return result
	}

	// Check result status
	if wfResult.Status == core.RunStatusFailed {
		result.Status = TaskStatusFailed
		if wfResult.Error != nil {
			result.Error = wfResult.Error.Error()
		} else {
			result.Error = "workflow execution failed"
		}
	} else {
		result.Status = TaskStatusCompleted
		result.Exports = wfResult.Exports
	}

	result.CompletedAt = time.Now()
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
	}
	if err := w.client.RegisterWorker(ctx, info); err != nil {
		w.logger.Warn("failed to update worker status", zap.Error(err))
	}
}

// cleanup removes the worker from the registry
func (w *Worker) cleanup(ctx context.Context) {
	w.printer.Info("Cleaning up worker %s...", w.ID)
	if err := w.client.RemoveWorker(ctx, w.ID); err != nil {
		w.logger.Warn("failed to remove worker", zap.Error(err))
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

// =============================================================================
// Distributed Hooks Registration
// =============================================================================

// registerDistributedHooks registers callbacks for database writes to use Redis queues
func (w *Worker) registerDistributedHooks() {
	hooks := &database.DistributedHooks{
		SendRun: func(ctx context.Context, run *database.Run) error {
			return w.SendRunData(ctx, run)
		},
		SendStepResult: func(ctx context.Context, step *database.StepResult) error {
			return w.SendStepResult(ctx, step)
		},
		SendEventLog: func(ctx context.Context, event *database.EventLog) error {
			return w.SendEventLog(ctx, event)
		},
		SendArtifact: func(ctx context.Context, artifact *database.Artifact) error {
			return w.SendArtifact(ctx, artifact)
		},
		ShouldUseRedis: func() bool {
			return config.ShouldUseRedisDataQueues()
		},
	}
	database.RegisterDistributedHooks(hooks)
	w.logger.Info("registered distributed hooks for database writes")
}

// unregisterDistributedHooks removes the distributed hooks
func (w *Worker) unregisterDistributedHooks() {
	database.UnregisterDistributedHooks()
	config.SetWorkerMode(false, "")
	w.logger.Info("unregistered distributed hooks")
}
