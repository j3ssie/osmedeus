package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/database/repository"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"go.uber.org/zap"
)

// ListWebhookRuns returns all runs with a webhook UUID (authenticated)
func ListWebhookRuns(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		db := database.GetDB()
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Database not connected",
			})
		}

		runRepo := repository.NewRunRepository(db)
		runs, err := runRepo.ListWebhookRuns(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to list webhook runs",
			})
		}

		// Build response with webhook URLs
		type webhookRunResponse struct {
			*database.Run
			WebhookURL string `json:"webhook_url"`
		}

		var results []webhookRunResponse
		for _, run := range runs {
			webhookURL := fmt.Sprintf("/osm/api/webhook-runs/%s/trigger", run.WebhookUUID)
			results = append(results, webhookRunResponse{
				Run:        run,
				WebhookURL: webhookURL,
			})
		}

		return c.JSON(fiber.Map{
			"data":  results,
			"total": len(results),
		})
	}
}

// TriggerWebhookRun triggers a run via its webhook UUID (unauthenticated)
func TriggerWebhookRun(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		webhookUUID := c.Params("uuid")
		if webhookUUID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Webhook UUID is required",
			})
		}

		db := database.GetDB()
		if db == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Database not connected",
			})
		}

		// Look up the webhook template run
		runRepo := repository.NewRunRepository(db)
		templateRun, err := runRepo.GetByWebhookUUID(c.Context(), webhookUUID)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Webhook not found",
			})
		}

		// Validate auth key if one is set
		if templateRun.WebhookAuthKey != "" {
			providedKey := c.Query("key")
			if providedKey != templateRun.WebhookAuthKey {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   true,
					"message": "Invalid or missing authentication key",
				})
			}
		}

		// Parse optional overrides from POST body
		target := templateRun.Target
		workflowName := templateRun.WorkflowName
		workflowKind := templateRun.WorkflowKind

		if c.Method() == fiber.MethodPost {
			var overrides struct {
				Target string `json:"target"`
				Flow   string `json:"flow"`
				Module string `json:"module"`
			}
			// Ignore parse errors — overrides are optional
			_ = c.BodyParser(&overrides)

			if overrides.Target != "" {
				target = overrides.Target
			}
			if overrides.Flow != "" {
				workflowName = overrides.Flow
				workflowKind = "flow"
			}
			if overrides.Module != "" {
				workflowName = overrides.Module
				workflowKind = "module"
			}
		}

		// Load workflow
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflow, err := loader.LoadWorkflow(workflowName)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": fmt.Sprintf("Workflow '%s' not found", workflowName),
			})
		}

		// Build params
		params := make(map[string]string)
		params["target"] = target
		for k, v := range templateRun.Params {
			if s, ok := v.(string); ok {
				params[k] = s
			}
		}

		// Create a new run record (the template stays as-is)
		isFlow := workflowKind == "flow"
		newRunUUID := uuid.New().String()
		now := time.Now()

		paramsInterface := make(map[string]interface{})
		for k, v := range params {
			paramsInterface[k] = v
		}

		workspace := computeWorkspace(target, params)

		newRun := &database.Run{
			RunUUID:      newRunUUID,
			WorkflowName: workflow.Name,
			WorkflowKind: string(workflow.Kind),
			Target:       target,
			Params:       paramsInterface,
			Status:       "running",
			TriggerType:  "webhook",
			StartedAt:    &now,
			TotalSteps:   calculateTotalSteps(workflow, loader),
			Workspace:    workspace,
			RunPriority:  "high",
			RunMode:      "local",
			HooksEnabled: workflow.HookCount() > 0,
		}

		if err := database.CreateRun(c.Context(), newRun); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to create run record",
			})
		}

		// Execute in background goroutine
		go func() {
			log := logger.Get()
			ctx := context.Background()

			exec := executor.NewExecutor()
			exec.SetServerMode(true)
			exec.SetLoader(loader)
			exec.SetDBRunUUID(newRunUUID)
			exec.SetDBRunID(newRun.ID)
			exec.SetOnStepCompleted(func(stepCtx context.Context, dbRunUUID string) {
				_ = database.IncrementRunCompletedSteps(stepCtx, dbRunUUID)
			})

			var execErr error
			if isFlow && workflow.IsFlow() {
				_, execErr = exec.ExecuteFlow(ctx, workflow, params, cfg)
			} else {
				_, execErr = exec.ExecuteModule(ctx, workflow, params, cfg)
			}

			if execErr != nil {
				log.Error("Webhook-triggered run failed",
					zap.String("run_uuid", newRunUUID),
					zap.String("target", target),
					zap.Error(execErr),
				)
				_ = database.UpdateRunStatus(ctx, newRunUUID, "failed", execErr.Error())
			} else {
				_ = database.UpdateRunStatus(ctx, newRunUUID, "completed", "")
			}
		}()

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":  "Run triggered successfully",
			"run_uuid": newRunUUID,
			"run_id":   newRun.ID,
			"target":   target,
			"workflow": workflowName,
		})
	}
}
