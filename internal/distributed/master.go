package distributed

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

const (
	MasterLockTTL     = 60 * time.Second
	MasterLockRefresh = 30 * time.Second
	WorkerCheckPeriod = 30 * time.Second
)

// Master represents a master node that coordinates workers
type Master struct {
	ID      string
	client  *Client
	config  *config.Config
	logger  *zap.Logger
	printer *terminal.Printer

	// For tracking
	mu      sync.RWMutex
	running bool
}

// NewMaster creates a new master node
func NewMaster(cfg *config.Config) (*Master, error) {
	client, err := NewClientFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	hostname, _ := os.Hostname()
	masterID := fmt.Sprintf("master-%s-%s", hostname, uuid.NewString()[:8])

	logger, _ := zap.NewProduction()

	return &Master{
		ID:      masterID,
		client:  client,
		config:  cfg,
		logger:  logger,
		printer: terminal.NewPrinter(),
	}, nil
}

// Start starts the master node
func (m *Master) Start(ctx context.Context) error {
	// Test connection
	if err := m.client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Acquire master lock
	acquired, err := m.client.AcquireMasterLock(ctx, m.ID, MasterLockTTL)
	if err != nil {
		return fmt.Errorf("failed to acquire master lock: %w", err)
	}
	if !acquired {
		return fmt.Errorf("another master is already running")
	}

	m.mu.Lock()
	m.running = true
	m.mu.Unlock()

	m.printer.Success("Master %s started", m.ID)
	m.printer.Info("Waiting for workers and tasks...")

	// Start lock refresh goroutine
	lockCtx, cancelLock := context.WithCancel(ctx)
	defer cancelLock()
	go m.lockRefreshLoop(lockCtx)

	// Start worker monitor goroutine
	monitorCtx, cancelMonitor := context.WithCancel(ctx)
	defer cancelMonitor()
	go m.workerMonitorLoop(monitorCtx)

	// Wait for shutdown
	<-ctx.Done()

	m.logger.Info("master shutting down", zap.String("master_id", m.ID))
	m.cleanup(context.Background())

	return nil
}

// lockRefreshLoop periodically refreshes the master lock
func (m *Master) lockRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(MasterLockRefresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.client.RefreshMasterLock(ctx, m.ID, MasterLockTTL); err != nil {
				m.logger.Error("failed to refresh master lock", zap.Error(err))
				// If we lose the lock, we should stop
				m.mu.Lock()
				m.running = false
				m.mu.Unlock()
				return
			}
		}
	}
}

// workerMonitorLoop monitors worker heartbeats and handles failures
func (m *Master) workerMonitorLoop(ctx context.Context) {
	ticker := time.NewTicker(WorkerCheckPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkWorkerHealth(ctx)
		}
	}
}

// checkWorkerHealth checks for dead workers and reassigns their tasks
func (m *Master) checkWorkerHealth(ctx context.Context) {
	workers, err := m.client.GetAllWorkers(ctx)
	if err != nil {
		m.logger.Warn("failed to get workers", zap.Error(err))
		return
	}

	now := time.Now()
	for _, worker := range workers {
		heartbeat, err := m.client.GetWorkerHeartbeat(ctx, worker.ID)
		if err != nil {
			continue
		}

		// Check if worker is dead (missed heartbeats)
		if heartbeat.IsZero() || now.Sub(heartbeat) > HeartbeatTimeout {
			m.logger.Warn("worker appears dead",
				zap.String("worker_id", worker.ID),
				zap.Duration("since_heartbeat", now.Sub(heartbeat)),
			)
			m.printer.Warning("Worker %s appears dead, reassigning tasks...", worker.ID)

			// Reassign the worker's running tasks
			m.reassignWorkerTasks(ctx, worker.ID)

			// Remove the dead worker
			if err := m.client.RemoveWorker(ctx, worker.ID); err != nil {
				m.logger.Error("failed to remove dead worker", zap.Error(err))
			}
		}
	}
}

// reassignWorkerTasks moves a dead worker's tasks back to pending
func (m *Master) reassignWorkerTasks(ctx context.Context, workerID string) {
	tasks, err := m.client.GetAllRunningTasks(ctx)
	if err != nil {
		m.logger.Error("failed to get running tasks", zap.Error(err))
		return
	}

	for _, task := range tasks {
		if task.WorkerID == workerID {
			m.logger.Info("reassigning task",
				zap.String("task_id", task.ID),
				zap.String("worker_id", workerID),
			)

			// Reset task status
			task.Status = TaskStatusPending
			task.WorkerID = ""
			task.StartedAt = nil

			// Push back to pending queue
			if err := m.client.PushTask(ctx, task); err != nil {
				m.logger.Error("failed to reassign task", zap.Error(err))
				continue
			}

			// Remove from running
			if err := m.client.RemoveTaskRunning(ctx, task.ID); err != nil {
				m.logger.Error("failed to remove task from running", zap.Error(err))
			}
		}
	}
}

// cleanup releases the master lock
func (m *Master) cleanup(ctx context.Context) {
	m.printer.Info("Cleaning up master %s...", m.ID)

	m.mu.Lock()
	m.running = false
	m.mu.Unlock()

	if err := m.client.ReleaseMasterLock(ctx, m.ID); err != nil {
		m.logger.Warn("failed to release master lock", zap.Error(err))
	}

	m.client.Close()
}

// SubmitTask submits a new task to the pending queue
func (m *Master) SubmitTask(ctx context.Context, task *Task) error {
	if task.ID == "" {
		task.ID = uuid.NewString()[:8]
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	task.Status = TaskStatusPending

	m.logger.Info("submitting task",
		zap.String("task_id", task.ID),
		zap.String("workflow", task.WorkflowName),
		zap.String("target", task.Target),
	)

	return m.client.PushTask(ctx, task)
}

// GetTaskStatus retrieves the status of a task
func (m *Master) GetTaskStatus(ctx context.Context, taskID string) (*Task, *TaskResult, error) {
	// Check running tasks first
	task, err := m.client.GetRunningTask(ctx, taskID)
	if err != nil {
		return nil, nil, err
	}
	if task != nil {
		return task, nil, nil
	}

	// Check completed tasks
	result, err := m.client.GetTaskResult(ctx, taskID)
	if err != nil {
		return nil, nil, err
	}
	if result != nil {
		return nil, result, nil
	}

	return nil, nil, fmt.Errorf("task not found: %s", taskID)
}

// ListWorkers returns all registered workers with their current status
func (m *Master) ListWorkers(ctx context.Context) ([]*WorkerInfo, error) {
	workers, err := m.client.GetAllWorkers(ctx)
	if err != nil {
		return nil, err
	}

	// Enrich with heartbeat info
	now := time.Now()
	for _, worker := range workers {
		heartbeat, err := m.client.GetWorkerHeartbeat(ctx, worker.ID)
		if err == nil {
			worker.LastHeartbeat = heartbeat

			// Update status based on heartbeat
			if now.Sub(heartbeat) > HeartbeatTimeout {
				worker.Status = "offline"
			}
		}
	}

	return workers, nil
}

// ListTasks returns all tasks (running and completed)
func (m *Master) ListTasks(ctx context.Context) ([]*Task, []*TaskResult, error) {
	running, err := m.client.GetAllRunningTasks(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Get completed tasks (we'd need to iterate the hash)
	// For now, return running tasks
	return running, nil, nil
}

// GetClient returns the Redis client for external use
func (m *Master) GetClient() *Client {
	return m.client
}

// IsRunning returns whether the master is currently running
func (m *Master) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}
