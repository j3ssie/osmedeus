package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
)

// generateEmptyTarget creates a placeholder target name for empty_target mode
func generateEmptyTarget() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	random := make([]byte, 6)
	for i := range random {
		random[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("empty-%s-%d", string(random), time.Now().Unix())
}

// createRunRecord creates a database record for a run
func createRunRecord(ctx context.Context, _ *config.Config, workflow *core.Workflow, target string, params map[string]string, triggerType, jobID string) (*database.Run, error) {
	now := time.Now()
	runID := uuid.New().String()

	paramsInterface := make(map[string]interface{})
	for k, v := range params {
		paramsInterface[k] = v
	}

	run := &database.Run{
		ID:           uuid.New().String(),
		RunID:        runID,
		WorkflowName: workflow.Name,
		WorkflowKind: string(workflow.Kind),
		Target:       target,
		Params:       paramsInterface,
		Status:       "running",
		TriggerType:  triggerType,
		JobID:        jobID,
		StartedAt:    &now,
		TotalSteps:   len(workflow.Steps),
	}

	if err := database.CreateRun(ctx, run); err != nil {
		return nil, err
	}

	return run, nil
}

// collectTargetsFromRequest collects targets from Target, Targets, and TargetFile fields
func collectTargetsFromRequest(req *CreateRunRequest) ([]string, error) {
	var allTargets []string

	// 1. Add single target if provided
	if req.Target != "" {
		allTargets = append(allTargets, req.Target)
	}

	// 2. Add targets array if provided
	allTargets = append(allTargets, req.Targets...)

	// 3. Read targets from file if provided
	if req.TargetFile != "" {
		fileTargets, err := readTargetsFromFile(req.TargetFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read target file: %w", err)
		}
		allTargets = append(allTargets, fileTargets...)
	}

	// Deduplicate and filter empty entries
	return deduplicateTargets(allTargets), nil
}

// executeRunsConcurrently runs workflows for multiple targets with concurrency control
func executeRunsConcurrently(
	workflow *core.Workflow,
	targets []string,
	baseParams map[string]string,
	cfg *config.Config,
	maxConcurrency int,
	isFlow bool,
	jobID string,
) {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	sem := make(chan struct{}, maxConcurrency) // Semaphore
	var wg sync.WaitGroup

	for _, target := range targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			// Clone params and set target
			targetParams := make(map[string]string)
			for k, v := range baseParams {
				targetParams[k] = v
			}
			targetParams["target"] = t

			ctx := context.Background()

			// Create run record in database
			run, err := createRunRecord(ctx, cfg, workflow, t, targetParams, "api", jobID)
			var runID string
			if err == nil && run != nil {
				runID = run.RunID
			}

			// Execute workflow
			exec := executor.NewExecutor()
			exec.SetServerMode(true) // Enable file logging for server mode

			// Set up database progress tracking
			if runID != "" {
				exec.SetDBRunID(runID)
				exec.SetOnStepCompleted(func(stepCtx context.Context, dbRunID string) {
					_ = database.IncrementRunCompletedSteps(stepCtx, dbRunID)
				})
			}

			var execErr error
			if isFlow && workflow.IsFlow() {
				_, execErr = exec.ExecuteFlow(ctx, workflow, targetParams, cfg)
			} else {
				_, execErr = exec.ExecuteModule(ctx, workflow, targetParams, cfg)
			}

			// Update run status in database
			if runID != "" {
				if execErr != nil {
					_ = database.UpdateRunStatus(ctx, runID, "failed", execErr.Error())
				} else {
					_ = database.UpdateRunStatus(ctx, runID, "completed", "")
				}
			}
		}(target)
	}

	// Wait for all runs to complete (optional - could also return immediately)
	wg.Wait()
}

// CreateRun handles run creation
// @Summary Create a new run
// @Description Execute a workflow against one or more targets. Supports multiple targets via array or file, concurrency control, priority levels, custom timeouts, runner configuration (host/docker/ssh), and scheduling via cron expressions.
// @Tags Runs
// @Accept json
// @Produce json
// @Param run body CreateRunRequest true "Run configuration with optional priority, timeout, runner config, and scheduling"
// @Success 202 {object} map[string]interface{} "Run started"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Workflow not found"
// @Security BearerAuth
// @Router /osm/api/runs [post]
func CreateRun(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateRunRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		// Handle empty target mode
		if req.EmptyTarget {
			req.Target = generateEmptyTarget()
		}

		// Collect all targets from Target, Targets, and TargetFile
		targets, err := collectTargetsFromRequest(&req)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		if len(targets) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "At least one target is required (target, targets, target_file, or empty_target)",
			})
		}

		// Validate heuristics_check if provided
		if req.HeuristicsCheck != "" {
			validHeuristics := map[string]bool{"none": true, "basic": true, "advanced": true}
			if !validHeuristics[req.HeuristicsCheck] {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   true,
					"message": "Invalid heuristics_check value. Must be: none, basic, or advanced",
				})
			}
		}

		// Determine workflow name and kind
		workflowName := req.Flow
		isFlow := true
		if workflowName == "" {
			workflowName = req.Module
			isFlow = false
		}

		if workflowName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Either 'flow' or 'module' is required",
			})
		}

		// Load workflow
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflow, err := loader.LoadWorkflow(workflowName)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Workflow not found",
			})
		}

		// Initialize params
		params := req.Params
		if params == nil {
			params = make(map[string]string)
		}

		// Set default priority if not specified
		priority := req.Priority
		if priority == "" {
			priority = "medium"
		}

		// Add runner configuration to params if specified
		if req.RunnerType != "" {
			params["runner_type"] = req.RunnerType
		}
		if req.DockerImage != "" {
			params["docker_image"] = req.DockerImage
		}
		if req.SSHHost != "" {
			params["ssh_host"] = req.SSHHost
		}
		if req.Timeout > 0 {
			params["timeout"] = fmt.Sprintf("%d", req.Timeout)
		}

		// Add execution options to params
		if req.ThreadsHold > 0 {
			params["threads_hold"] = fmt.Sprintf("%d", req.ThreadsHold)
		}
		if req.HeuristicsCheck != "" {
			params["heuristics_check"] = req.HeuristicsCheck
		}
		if req.Repeat {
			params["repeat"] = "true"
		}
		if req.RepeatWaitTime != "" {
			params["repeat_wait_time"] = req.RepeatWaitTime
		}

		// Set concurrency default
		concurrency := req.Concurrency
		if concurrency <= 0 {
			concurrency = 1
		}

		cfgCopy := config.Get()

		// Generate a job ID for grouping runs from this request
		jobID := uuid.New().String()[:8]

		// Create run record(s) and execute
		var runIDs []string
		if len(targets) == 1 {
			// Single target - existing behavior
			params["target"] = targets[0]

			// Create run record in database
			ctx := context.Background()
			run, _ := createRunRecord(ctx, cfgCopy, workflow, targets[0], params, "api", jobID)
			if run != nil {
				runIDs = append(runIDs, run.RunID)
			}

			exec := executor.NewExecutor()
			exec.SetServerMode(true) // Enable file logging for server mode

			// Set up database progress tracking
			if run != nil {
				exec.SetDBRunID(run.RunID)
				exec.SetOnStepCompleted(func(stepCtx context.Context, dbRunID string) {
					_ = database.IncrementRunCompletedSteps(stepCtx, dbRunID)
				})
			}

			go func(runID string) {
				ctx := context.Background()
				var execErr error
				if isFlow && workflow.IsFlow() {
					_, execErr = exec.ExecuteFlow(ctx, workflow, params, cfgCopy)
				} else {
					_, execErr = exec.ExecuteModule(ctx, workflow, params, cfgCopy)
				}
				// Update run status in database
				if runID != "" {
					if execErr != nil {
						_ = database.UpdateRunStatus(ctx, runID, "failed", execErr.Error())
					} else {
						_ = database.UpdateRunStatus(ctx, runID, "completed", "")
					}
				}
			}(func() string {
				if run != nil {
					return run.RunID
				}
				return ""
			}())
		} else {
			// Multiple targets - concurrent execution
			go executeRunsConcurrently(workflow, targets, params, cfgCopy, concurrency, isFlow, jobID)
		}

		// Build response
		response := fiber.Map{
			"message":      "Run started",
			"workflow":     workflow.Name,
			"kind":         workflow.Kind,
			"target_count": len(targets),
			"priority":     priority,
			"job_id":       jobID,
			"status":       "queued",
			"poll_url":     fmt.Sprintf("/osm/api/jobs/%s", jobID),
		}

		// For single target, include target field and run_id for backward compatibility
		if len(targets) == 1 {
			response["target"] = targets[0]
			if len(runIDs) > 0 {
				response["run_id"] = runIDs[0]
			}
		} else {
			response["targets"] = targets
			response["concurrency"] = concurrency
		}

		// Add optional fields to response
		if req.RunnerType != "" {
			response["runner_type"] = req.RunnerType
		}
		if req.Timeout > 0 {
			response["timeout"] = req.Timeout
		}
		if req.Schedule != "" {
			response["schedule"] = req.Schedule
			response["schedule_enabled"] = req.ScheduleEnabled
		}
		if req.ThreadsHold > 0 {
			response["threads_hold"] = req.ThreadsHold
		}
		if req.EmptyTarget {
			response["empty_target"] = true
		}
		if req.HeuristicsCheck != "" {
			response["heuristics_check"] = req.HeuristicsCheck
		}
		if req.Repeat {
			response["repeat"] = true
			if req.RepeatWaitTime != "" {
				response["repeat_wait_time"] = req.RepeatWaitTime
			}
		}

		return c.Status(fiber.StatusAccepted).JSON(response)
	}
}

// ListRuns handles listing runs
// @Summary List runs
// @Description Get a paginated list of workflow runs with optional filters
// @Tags Runs
// @Produce json
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Param status query string false "Filter by status (pending, running, completed, failed, cancelled)"
// @Param workflow query string false "Filter by workflow name"
// @Param target query string false "Filter by target (partial match)"
// @Success 200 {object} map[string]interface{} "List of runs"
// @Security BearerAuth
// @Router /osm/api/runs [get]
func ListRuns(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		limit, _ := strconv.Atoi(c.Query("limit", "20"))
		status := c.Query("status")
		workflow := c.Query("workflow")
		target := c.Query("target")

		if offset < 0 {
			offset = 0
		}
		if limit <= 0 {
			limit = 20
		}
		if limit > 10000 {
			limit = 10000
		}

		ctx := context.Background()
		result, err := database.ListRuns(ctx, offset, limit, status, workflow, target)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"data": result.Data,
			"pagination": fiber.Map{
				"total":  result.TotalCount,
				"offset": result.Offset,
				"limit":  result.Limit,
			},
		})
	}
}

// GetRun handles getting a single run
// @Summary Get run details
// @Description Get details of a specific run by ID, including steps and artifacts
// @Tags Runs
// @Produce json
// @Param id path string true "Run ID or RunID"
// @Param include_steps query bool false "Include step results" default(false)
// @Param include_artifacts query bool false "Include artifacts" default(false)
// @Success 200 {object} map[string]interface{} "Run details"
// @Failure 404 {object} map[string]interface{} "Run not found"
// @Security BearerAuth
// @Router /osm/api/runs/{id} [get]
func GetRun(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		includeSteps := c.Query("include_steps") == "true"
		includeArtifacts := c.Query("include_artifacts") == "true"

		ctx := context.Background()
		run, err := database.GetRunByID(ctx, id, includeSteps, includeArtifacts)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Run not found",
			})
		}

		return c.JSON(fiber.Map{"data": run})
	}
}

// CancelRun handles cancelling a run
// @Summary Cancel a run
// @Description Cancel a running workflow execution
// @Tags Runs
// @Produce json
// @Param id path string true "Run ID or RunID"
// @Success 200 {object} map[string]interface{} "Run cancelled"
// @Failure 404 {object} map[string]interface{} "Run not found"
// @Failure 400 {object} map[string]interface{} "Run cannot be cancelled"
// @Security BearerAuth
// @Router /osm/api/runs/{id} [delete]
func CancelRun(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		ctx := context.Background()

		run, err := database.GetRunByID(ctx, id, false, false)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Run not found",
			})
		}

		if run.Status != "pending" && run.Status != "running" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("Cannot cancel run with status '%s'", run.Status),
			})
		}

		err = database.UpdateRunStatus(ctx, id, "cancelled", "Cancelled by user")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to cancel run",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Run cancelled successfully",
			"id":      run.ID,
			"run_id":  run.RunID,
		})
	}
}

// GetRunSteps handles getting run steps
func GetRunSteps(c *fiber.Ctx) error {
	// TODO: Implement with database
	return c.JSON(fiber.Map{
		"data": []interface{}{},
	})
}

// GetRunArtifacts handles getting run artifacts
func GetRunArtifacts(c *fiber.Ctx) error {
	// TODO: Implement with database
	return c.JSON(fiber.Map{
		"data": []interface{}{},
	})
}

// DownloadArtifact handles artifact download
func DownloadArtifact(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error":   true,
		"message": "Artifact not found",
	})
}
