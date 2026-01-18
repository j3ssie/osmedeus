package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
)

// CreateSchedule handles creating a new schedule
// @Summary Create a new schedule
// @Description Create a scheduled workflow execution with cron expression
// @Tags Schedules
// @Accept json
// @Produce json
// @Param schedule body CreateScheduleRequest true "Schedule configuration"
// @Success 201 {object} map[string]interface{} "Schedule created"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Security BearerAuth
// @Router /osm/api/schedules [post]
func CreateSchedule(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateScheduleRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		// Validate required fields
		if req.Name == "" || req.WorkflowName == "" || req.Schedule == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "name, workflow_name, and schedule are required",
			})
		}

		ctx := context.Background()

		// Create schedule record
		schedule, err := database.CreateSchedule(ctx, database.CreateScheduleInput{
			Name:         req.Name,
			WorkflowName: req.WorkflowName,
			WorkflowKind: req.WorkflowKind,
			Target:       req.Target,
			Schedule:     req.Schedule,
			Enabled:      req.Enabled,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Schedule created",
			"data":    schedule,
		})
	}
}

// ListSchedules handles listing all schedules
// @Summary List all schedules
// @Description Get a paginated list of all scheduled workflows
// @Tags Schedules
// @Produce json
// @Param offset query int false "Number of records to skip" default(0)
// @Param limit query int false "Maximum number of records to return" default(20)
// @Success 200 {object} map[string]interface{} "List of schedules"
// @Security BearerAuth
// @Router /osm/api/schedules [get]
func ListSchedules(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		limit, _ := strconv.Atoi(c.Query("limit", "20"))

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

		result, err := database.ListSchedules(ctx, offset, limit)
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

// GetSchedule handles getting a single schedule
// @Summary Get schedule details
// @Description Get details of a specific schedule by ID
// @Tags Schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} map[string]interface{} "Schedule details"
// @Failure 404 {object} map[string]interface{} "Schedule not found"
// @Security BearerAuth
// @Router /osm/api/schedules/{id} [get]
func GetSchedule(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		ctx := context.Background()
		schedule, err := database.GetScheduleByID(ctx, id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Schedule not found",
			})
		}

		return c.JSON(fiber.Map{"data": schedule})
	}
}

// UpdateSchedule handles updating a schedule
// @Summary Update a schedule
// @Description Update an existing schedule
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "Schedule ID"
// @Param schedule body UpdateScheduleRequest true "Schedule update data"
// @Success 200 {object} map[string]interface{} "Schedule updated"
// @Failure 404 {object} map[string]interface{} "Schedule not found"
// @Security BearerAuth
// @Router /osm/api/schedules/{id} [put]
func UpdateSchedule(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var req UpdateScheduleRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		ctx := context.Background()
		schedule, err := database.UpdateSchedule(ctx, id, database.UpdateScheduleInput{
			Name:     req.Name,
			Target:   req.Target,
			Schedule: req.Schedule,
			Enabled:  req.Enabled,
		})
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Schedule updated",
			"data":    schedule,
		})
	}
}

// DeleteSchedule handles deleting a schedule
// @Summary Delete a schedule
// @Description Delete a schedule by ID
// @Tags Schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} map[string]interface{} "Schedule deleted"
// @Failure 404 {object} map[string]interface{} "Schedule not found"
// @Security BearerAuth
// @Router /osm/api/schedules/{id} [delete]
func DeleteSchedule(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		ctx := context.Background()
		if err := database.DeleteSchedule(ctx, id); err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{"message": "Schedule deleted"})
	}
}

// EnableSchedule handles enabling a schedule
// @Summary Enable a schedule
// @Description Enable a disabled schedule
// @Tags Schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} map[string]interface{} "Schedule enabled"
// @Security BearerAuth
// @Router /osm/api/schedules/{id}/enable [post]
func EnableSchedule(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		ctx := context.Background()
		enabled := true
		_, err := database.UpdateSchedule(ctx, id, database.UpdateScheduleInput{Enabled: &enabled})
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{"message": "Schedule enabled"})
	}
}

// DisableSchedule handles disabling a schedule
// @Summary Disable a schedule
// @Description Disable an enabled schedule
// @Tags Schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 200 {object} map[string]interface{} "Schedule disabled"
// @Security BearerAuth
// @Router /osm/api/schedules/{id}/disable [post]
func DisableSchedule(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		ctx := context.Background()
		enabled := false
		_, err := database.UpdateSchedule(ctx, id, database.UpdateScheduleInput{Enabled: &enabled})
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{"message": "Schedule disabled"})
	}
}

// TriggerSchedule handles manually triggering a scheduled workflow
// @Summary Trigger a schedule
// @Description Manually trigger a scheduled workflow execution
// @Tags Schedules
// @Produce json
// @Param id path string true "Schedule ID"
// @Success 202 {object} map[string]interface{} "Schedule triggered"
// @Security BearerAuth
// @Router /osm/api/schedules/{id}/trigger [post]
func TriggerSchedule(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		ctx := context.Background()
		schedule, err := database.GetScheduleByID(ctx, id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Schedule not found",
			})
		}

		// Load and execute the workflow
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflow, err := loader.LoadWorkflow(schedule.WorkflowName)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to load workflow",
			})
		}

		// Execute in background
		exec := executor.NewExecutor()
		cfgCopy := config.Get()
		go func() {
			bgCtx := context.Background()
			params := make(map[string]string)
			if schedule.InputConfig != nil {
				for k, v := range schedule.InputConfig {
					if s, ok := v.(string); ok {
						params[k] = s
					}
				}
			}
			if workflow.IsFlow() {
				_, _ = exec.ExecuteFlow(bgCtx, workflow, params, cfgCopy)
			} else {
				_, _ = exec.ExecuteModule(bgCtx, workflow, params, cfgCopy)
			}
		}()

		// Update last run time
		_ = database.UpdateScheduleLastRun(ctx, id)

		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"message":  "Schedule triggered",
			"schedule": schedule.Name,
			"workflow": schedule.WorkflowName,
		})
	}
}
