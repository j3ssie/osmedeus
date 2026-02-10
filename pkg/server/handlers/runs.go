package handlers

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"go.uber.org/zap"
)

// killProcessAndChildren kills a process and all its children using SIGKILL
// Returns true if the kill signal was sent successfully
func killProcessAndChildren(pid int) bool {
	if pid <= 0 {
		return false
	}

	// First, try to kill the process group (negative PID kills all processes in the group)
	// This ensures child processes are also terminated
	err := syscall.Kill(-pid, syscall.SIGKILL)
	if err != nil {
		// Process group kill failed, try killing just the process
		err = syscall.Kill(pid, syscall.SIGKILL)
		if err != nil {
			return false
		}
	}
	return true
}

// generateEmptyTarget creates a placeholder target name for empty_target mode
func generateEmptyTarget() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	random := make([]byte, 6)
	for i := range random {
		random[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("empty-%s-%d", string(random), time.Now().Unix())
}

// calculateTotalSteps returns the appropriate step count based on workflow kind.
// For module workflows, it returns len(Steps). For flow workflows, it sums steps from all modules.
func calculateTotalSteps(workflow *core.Workflow, loader *parser.Loader) int {
	if workflow.Kind != core.KindFlow {
		return len(workflow.Steps)
	}

	// Flow workflow: sum steps from all modules
	if loader == nil {
		return len(workflow.Modules)
	}

	log := logger.Get()
	totalSteps := 0

	for _, modRef := range workflow.Modules {
		if modRef.Path == "" {
			totalSteps++
			continue
		}

		module, err := loader.LoadWorkflowByPath(modRef.Path)
		if err != nil {
			log.Warn("Failed to load module for step counting",
				zap.String("module", modRef.Name),
				zap.String("path", modRef.Path),
				zap.Error(err),
			)
			totalSteps++
			continue
		}

		totalSteps += len(module.Steps)
	}

	return totalSteps
}

// computeWorkspace computes the workspace name from target and params
// This mirrors the executor's logic for computing TargetSpace
func computeWorkspace(target string, params map[string]string) string {
	// If space_name param provided (via -S flag or API), use it directly
	if spaceName := params["space_name"]; spaceName != "" {
		return spaceName
	}
	// Otherwise, sanitize the target for filesystem safety
	return sanitizeTargetForWorkspace(target)
}

// sanitizeTargetForWorkspace creates a filesystem-safe workspace name from target
// This mirrors the executor's sanitizeTargetSpace function
func sanitizeTargetForWorkspace(target string) string {
	sanitized := make([]rune, 0, len(target))
	for _, r := range target {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			sanitized = append(sanitized, '_')
		} else {
			sanitized = append(sanitized, r)
		}
	}
	result := string(sanitized)
	// Limit length to avoid filesystem issues
	if len(result) > 200 {
		result = result[:200]
	}
	return result
}

// createRunRecord creates a database record for a run
func createRunRecord(ctx context.Context, _ *config.Config, workflow *core.Workflow, loader *parser.Loader, target string, params map[string]string, triggerType, jobID, priority, runMode string) (*database.Run, error) {
	now := time.Now()
	runID := uuid.New().String()

	paramsInterface := make(map[string]interface{})
	for k, v := range params {
		paramsInterface[k] = v
	}

	// Compute workspace from target and params
	workspace := computeWorkspace(target, params)

	run := &database.Run{
		RunUUID:      runID,
		WorkflowName: workflow.Name,
		WorkflowKind: string(workflow.Kind),
		Target:       target,
		Params:       paramsInterface,
		Status:       "running",
		TriggerType:  triggerType,
		RunGroupID:   jobID,
		StartedAt:    &now,
		TotalSteps:   calculateTotalSteps(workflow, loader),
		Workspace:    workspace,
		RunPriority:  priority,
		RunMode:      runMode,
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
	priority string,
	runMode string,
) {
	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	loader := parser.NewLoader(cfg.WorkflowsPath)

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
			run, err := createRunRecord(ctx, cfg, workflow, loader, t, targetParams, "api", jobID, priority, runMode)
			var runUUID string
			var runID int64
			if err == nil && run != nil {
				runUUID = run.RunUUID
				runID = run.ID
			}

			// Execute workflow
			exec := executor.NewExecutor()
			exec.SetServerMode(true) // Enable file logging for server mode
			exec.SetLoader(loader)

			// Set up database progress tracking
			if runUUID != "" {
				exec.SetDBRunUUID(runUUID)
				exec.SetDBRunID(runID)
				exec.SetOnStepCompleted(func(stepCtx context.Context, dbRunUUID string) {
					_ = database.IncrementRunCompletedSteps(stepCtx, dbRunUUID)
				})
			}

			var execErr error
			if isFlow && workflow.IsFlow() {
				_, execErr = exec.ExecuteFlow(ctx, workflow, targetParams, cfg)
			} else {
				_, execErr = exec.ExecuteModule(ctx, workflow, targetParams, cfg)
			}

			// Update run status in database
			if runUUID != "" {
				if execErr != nil {
					_ = database.UpdateRunStatus(ctx, runUUID, "failed", execErr.Error())
				} else {
					_ = database.UpdateRunStatus(ctx, runUUID, "completed", "")
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
			priority = "normal"
		}
		// Validate priority
		validPriorities := map[string]bool{"low": true, "normal": true, "high": true, "critical": true}
		if !validPriorities[priority] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid priority. Must be one of: low, normal, high, critical",
			})
		}

		// Set default run_mode if not specified
		runMode := req.RunMode
		if runMode == "" {
			runMode = "local"
		}
		// Validate run_mode
		validModes := map[string]bool{"local": true, "distributed": true, "cloud": true}
		if !validModes[runMode] {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid run_mode. Must be one of: local, distributed, cloud",
			})
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
			run, _ := createRunRecord(ctx, cfgCopy, workflow, loader, targets[0], params, "api", jobID, priority, runMode)
			if run != nil {
				runIDs = append(runIDs, run.RunUUID)
			}

			exec := executor.NewExecutor()
			exec.SetServerMode(true) // Enable file logging for server mode
			exec.SetLoader(loader)
			if req.EmptyTarget {
				exec.SetSkipWorkspace(true)
			}

			// Set up database progress tracking
			if run != nil {
				exec.SetDBRunUUID(run.RunUUID)
				exec.SetDBRunID(run.ID)
				exec.SetOnStepCompleted(func(stepCtx context.Context, dbRunUUID string) {
					_ = database.IncrementRunCompletedSteps(stepCtx, dbRunUUID)
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
					return run.RunUUID
				}
				return ""
			}())
		} else {
			// Multiple targets - concurrent execution
			go executeRunsConcurrently(workflow, targets, params, cfgCopy, concurrency, isFlow, jobID, priority, runMode)
		}

		// Build response
		response := fiber.Map{
			"message":      "Run started",
			"workflow":     workflow.Name,
			"kind":         workflow.Kind,
			"target_count": len(targets),
			"priority":     priority,
			"run_mode":     runMode,
			"job_id":       jobID,
			"status":       "queued",
			"poll_url":     fmt.Sprintf("/osm/api/jobs/%s", jobID),
		}

		// For single target, include target field and run_uuid for backward compatibility
		if len(targets) == 1 {
			response["target"] = targets[0]
			if len(runIDs) > 0 {
				response["run_uuid"] = runIDs[0]
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
// @Param workspace query string false "Filter by workspace name (exact match)"
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
		workspace := c.Query("workspace")

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
		result, err := database.ListRuns(ctx, offset, limit, status, workflow, target, workspace)
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
// @Description Cancel a running workflow execution. This will terminate all running processes associated with the run.
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

		var killedPIDs []int
		var killMethod string

		// Try to cancel via control plane first (kills running processes tracked in memory)
		controlPlane := executor.GetRunControlPlane()
		controlPlanePIDs, controlPlaneErr := controlPlane.Cancel(run.RunUUID)

		if controlPlaneErr == nil && len(controlPlanePIDs) > 0 {
			// Control plane had the run and killed processes
			killedPIDs = controlPlanePIDs
			killMethod = "control_plane"
		} else if run.CurrentPID > 0 {
			// Run not in control plane, but we have a PID from database - kill it directly
			killed := killProcessAndChildren(run.CurrentPID)
			if killed {
				killedPIDs = []int{run.CurrentPID}
				killMethod = "database_pid"
			}
		}

		// Update database status
		err = database.UpdateRunStatus(ctx, run.RunUUID, "cancelled", "Cancelled by user")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to cancel run",
			})
		}

		response := fiber.Map{
			"message":  "Run cancelled successfully",
			"id":       run.ID,
			"run_uuid": run.RunUUID,
		}

		// Add PID information to response
		if len(killedPIDs) > 0 {
			response["killed_pids"] = killedPIDs
			response["processes_terminated"] = len(killedPIDs)
			response["kill_method"] = killMethod
		} else {
			response["note"] = "No active processes found to terminate; database status updated"
		}

		return c.JSON(response)
	}
}

// GetRunSteps handles getting run steps
func GetRunSteps(c *fiber.Ctx) error {
	runUUID := c.Params("id")
	if runUUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "run ID is required",
		})
	}

	ctx := context.Background()
	steps, err := database.GetRunSteps(ctx, runUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": steps,
	})
}

// GetRunArtifacts handles getting run artifacts
func GetRunArtifacts(c *fiber.Ctx) error {
	runUUID := c.Params("id")
	if runUUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "run ID is required",
		})
	}

	ctx := context.Background()
	artifacts, err := database.GetRunArtifacts(ctx, runUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": artifacts,
	})
}

// DownloadArtifact handles artifact download
func DownloadArtifact(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
		"error":   true,
		"message": "Artifact not found",
	})
}

// convertParamsToStringMap converts map[string]interface{} to map[string]string
func convertParamsToStringMap(params map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		if s, ok := v.(string); ok {
			result[k] = s
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// DuplicateRun duplicates an existing run with pending status
// @Summary Duplicate a run
// @Description Create a copy of an existing run with pending status. The duplicate will have the same workflow, target, and parameters but a new UUID.
// @Tags Runs
// @Produce json
// @Param id path string true "Run ID or RunUUID"
// @Success 201 {object} map[string]interface{} "Run duplicated"
// @Failure 404 {object} map[string]interface{} "Run not found"
// @Security BearerAuth
// @Router /osm/api/runs/{id}/duplicate [post]
func DuplicateRun(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		ctx := context.Background()

		// Get original run
		original, err := database.GetRunByID(ctx, id, false, false)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Run not found",
			})
		}

		// Create duplicate with new UUID and pending status
		newRun := &database.Run{
			RunUUID:      uuid.New().String(),
			WorkflowName: original.WorkflowName,
			WorkflowKind: original.WorkflowKind,
			Target:       original.Target,
			Params:       original.Params,
			Workspace:    original.Workspace,
			Status:       "pending",
			TriggerType:  "api",
			TotalSteps:   original.TotalSteps,
			// Reset timing/progress fields (StartedAt, CompletedAt, CompletedSteps are zero values)
		}

		if err := database.CreateRun(ctx, newRun); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to create duplicate run",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":           "Run duplicated",
			"original_run_uuid": original.RunUUID,
			"run_uuid":          newRun.RunUUID,
			"workflow":          newRun.WorkflowName,
			"target":            newRun.Target,
			"status":            newRun.Status,
		})
	}
}

// StartRun starts a pending run (triggers workflow execution)
// @Summary Start a pending run
// @Description Start a run that is in pending status. This triggers the workflow execution.
// @Tags Runs
// @Produce json
// @Param id path string true "Run ID or RunUUID"
// @Success 202 {object} map[string]interface{} "Run started"
// @Failure 400 {object} map[string]interface{} "Run cannot be started (not in pending status)"
// @Failure 404 {object} map[string]interface{} "Run or workflow not found"
// @Security BearerAuth
// @Router /osm/api/runs/{id}/start [post]
func StartRun(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		ctx := context.Background()

		// Get run record
		run, err := database.GetRunByID(ctx, id, false, false)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Run not found",
			})
		}

		// Verify status is pending
		if run.Status != "pending" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("Cannot start run with status '%s'. Only runs with 'pending' status can be started.", run.Status),
			})
		}

		// Load workflow
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflow, err := loader.LoadWorkflow(run.WorkflowName)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("Workflow '%s' not found", run.WorkflowName),
			})
		}

		// Update status to running and set start time
		now := time.Now()
		if err := database.UpdateRunStatus(ctx, run.RunUUID, "running", ""); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to update run status",
			})
		}

		// Convert params and execute in goroutine
		params := convertParamsToStringMap(run.Params)
		params["target"] = run.Target

		cfgCopy := config.Get()
		isFlow := run.WorkflowKind == string(core.KindFlow)

		go func(runUUID string, runID int64, startTime time.Time) {
			execCtx := context.Background()

			// Create executor
			exec := executor.NewExecutor()
			exec.SetServerMode(true)
			exec.SetLoader(loader)
			exec.SetDBRunUUID(runUUID)
			exec.SetDBRunID(runID)
			exec.SetOnStepCompleted(func(stepCtx context.Context, dbRunUUID string) {
				_ = database.IncrementRunCompletedSteps(stepCtx, dbRunUUID)
			})

			var execErr error
			if isFlow && workflow.IsFlow() {
				_, execErr = exec.ExecuteFlow(execCtx, workflow, params, cfgCopy)
			} else {
				_, execErr = exec.ExecuteModule(execCtx, workflow, params, cfgCopy)
			}

			// Update final status
			if execErr != nil {
				_ = database.UpdateRunStatus(execCtx, runUUID, "failed", execErr.Error())
			} else {
				_ = database.UpdateRunStatus(execCtx, runUUID, "completed", "")
			}
		}(run.RunUUID, run.ID, now)

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":  "Run started",
			"run_uuid": run.RunUUID,
			"workflow": run.WorkflowName,
			"kind":     run.WorkflowKind,
			"target":   run.Target,
			"status":   "running",
		})
	}
}
