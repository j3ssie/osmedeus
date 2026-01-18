package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
)

// ListWorkers returns all registered workers with their current status
// @Summary List all workers
// @Description Get a list of all registered workers in the distributed pool
// @Tags Distributed
// @Produce json
// @Success 200 {object} map[string]interface{} "List of workers"
// @Failure 500 {object} map[string]interface{} "Failed to list workers"
// @Security BearerAuth
// @Router /osm/api/workers [get]
func ListWorkers(master *distributed.Master) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		workers, err := master.ListWorkers(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		data := make([]fiber.Map, 0, len(workers))
		for _, w := range workers {
			data = append(data, fiber.Map{
				"id":             w.ID,
				"hostname":       w.Hostname,
				"status":         w.Status,
				"current_task":   w.CurrentTaskID,
				"joined_at":      w.JoinedAt,
				"last_heartbeat": w.LastHeartbeat,
				"tasks_complete": w.TasksComplete,
				"tasks_failed":   w.TasksFailed,
			})
		}

		return c.JSON(fiber.Map{
			"data":  data,
			"count": len(data),
		})
	}
}

// GetWorker returns details for a specific worker
// @Summary Get worker details
// @Description Get details for a specific worker by ID
// @Tags Distributed
// @Produce json
// @Param id path string true "Worker ID"
// @Success 200 {object} map[string]interface{} "Worker details"
// @Failure 404 {object} map[string]interface{} "Worker not found"
// @Failure 500 {object} map[string]interface{} "Failed to get worker"
// @Security BearerAuth
// @Router /osm/api/workers/{id} [get]
func GetWorker(master *distributed.Master) fiber.Handler {
	return func(c *fiber.Ctx) error {
		workerID := c.Params("id")
		ctx := c.Context()

		workers, err := master.ListWorkers(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		for _, w := range workers {
			if w.ID == workerID {
				return c.JSON(fiber.Map{
					"id":             w.ID,
					"hostname":       w.Hostname,
					"status":         w.Status,
					"current_task":   w.CurrentTaskID,
					"joined_at":      w.JoinedAt,
					"last_heartbeat": w.LastHeartbeat,
					"tasks_complete": w.TasksComplete,
					"tasks_failed":   w.TasksFailed,
				})
			}
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Worker not found",
		})
	}
}

// ListTasks returns all tasks (running and completed)
// @Summary List all tasks
// @Description Get a list of all running and completed tasks
// @Tags Distributed
// @Produce json
// @Success 200 {object} map[string]interface{} "List of running and completed tasks"
// @Failure 500 {object} map[string]interface{} "Failed to list tasks"
// @Security BearerAuth
// @Router /osm/api/tasks [get]
func ListTasks(master *distributed.Master) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		running, completed, err := master.ListTasks(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		runningData := make([]fiber.Map, 0, len(running))
		for _, t := range running {
			runningData = append(runningData, taskToMap(t))
		}

		completedData := make([]fiber.Map, 0)
		for _, r := range completed {
			completedData = append(completedData, taskResultToMap(r))
		}

		return c.JSON(fiber.Map{
			"running":   runningData,
			"completed": completedData,
		})
	}
}

// GetTask returns details for a specific task
// @Summary Get task details
// @Description Get details for a specific task by ID
// @Tags Distributed
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} map[string]interface{} "Task details"
// @Failure 404 {object} map[string]interface{} "Task not found"
// @Security BearerAuth
// @Router /osm/api/tasks/{id} [get]
func GetTask(master *distributed.Master) fiber.Handler {
	return func(c *fiber.Ctx) error {
		taskID := c.Params("id")
		ctx := c.Context()

		task, result, err := master.GetTaskStatus(ctx, taskID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		if task != nil {
			return c.JSON(taskToMap(task))
		}

		if result != nil {
			return c.JSON(taskResultToMap(result))
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "Task not found",
		})
	}
}

// SubmitTaskRequest represents a task submission request
type SubmitTaskRequest struct {
	WorkflowName string                 `json:"workflow_name"`
	WorkflowKind string                 `json:"workflow_kind"`
	Target       string                 `json:"target"`
	Params       map[string]interface{} `json:"params"`
}

// SubmitTask submits a new task to the distributed queue
// @Summary Submit a new task
// @Description Submit a new task to the distributed worker queue
// @Tags Distributed
// @Accept json
// @Produce json
// @Param task body SubmitTaskRequest true "Task configuration"
// @Success 202 {object} map[string]interface{} "Task submitted"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 500 {object} map[string]interface{} "Failed to submit task"
// @Security BearerAuth
// @Router /osm/api/tasks [post]
func SubmitTask(master *distributed.Master) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req SubmitTaskRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		if req.WorkflowName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "workflow_name is required",
			})
		}

		if req.Target == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "target is required",
			})
		}

		task := &distributed.Task{
			WorkflowName: req.WorkflowName,
			WorkflowKind: req.WorkflowKind,
			Target:       req.Target,
			Params:       req.Params,
		}

		ctx := c.Context()
		if err := master.SubmitTask(ctx, task); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message": "Task submitted",
			"task_id": task.ID,
		})
	}
}

// taskToMap converts a Task to a fiber.Map
func taskToMap(t *distributed.Task) fiber.Map {
	m := fiber.Map{
		"id":            t.ID,
		"scan_id":       t.ScanID,
		"workflow_name": t.WorkflowName,
		"workflow_kind": t.WorkflowKind,
		"target":        t.Target,
		"status":        t.Status,
		"worker_id":     t.WorkerID,
		"created_at":    t.CreatedAt,
	}
	if t.StartedAt != nil {
		m["started_at"] = t.StartedAt
	}
	return m
}

// taskResultToMap converts a TaskResult to a fiber.Map
func taskResultToMap(r *distributed.TaskResult) fiber.Map {
	return fiber.Map{
		"task_id":      r.TaskID,
		"status":       r.Status,
		"output":       r.Output,
		"error":        r.Error,
		"exports":      r.Exports,
		"completed_at": r.CompletedAt,
	}
}
