package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
)

// JobStatus represents the aggregated status of a job (run group)
type JobStatus struct {
	RunGroupID string          `json:"run_group_id"`
	Status     string          `json:"status"` // pending, running, completed, failed, partial
	Runs       []*database.Run `json:"runs"`
	Progress   JobProgress     `json:"progress"`
}

// JobProgress represents progress statistics for a job
type JobProgress struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Running   int `json:"running"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// GetJobStatus handles getting the status of a job (group of runs)
// @Summary Get job status
// @Description Get the aggregated status of a job and its runs
// @Tags Jobs
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]interface{} "Job status"
// @Failure 404 {object} map[string]interface{} "Job not found"
// @Security BearerAuth
// @Router /osm/api/jobs/{id} [get]
func GetJobStatus(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jobID := c.Params("id")
		if jobID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Job ID is required",
			})
		}

		ctx := context.Background()
		runs, err := database.GetRunsByRunGroupID(ctx, jobID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		if len(runs) == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Job not found",
			})
		}

		// Calculate progress
		progress := JobProgress{Total: len(runs)}
		for _, run := range runs {
			switch run.Status {
			case "pending":
				progress.Pending++
			case "running":
				progress.Running++
			case "completed":
				progress.Completed++
			case "failed":
				progress.Failed++
			}
		}

		// Determine aggregate status
		status := aggregateStatus(progress)

		return c.JSON(fiber.Map{
			"run_group_id": jobID,
			"status":       status,
			"runs":         runs,
			"progress":     progress,
		})
	}
}

// aggregateStatus determines the overall job status based on run statuses
func aggregateStatus(progress JobProgress) string {
	if progress.Total == 0 {
		return "pending"
	}
	if progress.Running > 0 {
		return "running"
	}
	if progress.Pending > 0 && progress.Completed == 0 && progress.Failed == 0 {
		return "pending"
	}
	if progress.Completed == progress.Total {
		return "completed"
	}
	if progress.Failed == progress.Total {
		return "failed"
	}
	if progress.Failed > 0 || progress.Completed > 0 {
		return "partial" // some completed, some failed
	}
	return "pending"
}
