package handlers

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
)

// ListWorkflows handles listing workflows
func ListWorkflows(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflows, err := loader.LoadAllWorkflows()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to load workflows",
			})
		}

		data := make([]fiber.Map, 0, len(workflows))
		for _, w := range workflows {
			data = append(data, fiber.Map{
				"name":        w.Name,
				"kind":        w.Kind,
				"description": w.Description,
				"file_path":   w.FilePath,
			})
		}

		return c.JSON(fiber.Map{
			"data": data,
		})
	}
}

// GetWorkflow handles getting a single workflow
func GetWorkflow(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.Params("name")
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflow, err := loader.LoadWorkflow(name)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Workflow not found",
			})
		}

		return c.JSON(workflowToMap(workflow))
	}
}

// ListWorkflowsVerbose handles listing workflows with verbose details
// @Summary List all workflows
// @Description Get a list of all available workflows with details
// @Tags Workflows
// @Produce json
// @Success 200 {object} map[string]interface{} "List of workflows"
// @Failure 500 {object} map[string]interface{} "Failed to load workflows"
// @Security BearerAuth
// @Router /osm/api/workflows [get]
func ListWorkflowsVerbose(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check source parameter - "filesystem" for direct loading, default is "db"
		source := c.Query("source", "db")

		// If source is filesystem, use the original approach
		if source == "filesystem" {
			return listWorkflowsFromFilesystem(c, cfg)
		}

		// Parse query parameters for DB-based listing
		offset, _ := strconv.Atoi(c.Query("offset", "0"))
		limit, _ := strconv.Atoi(c.Query("limit", "50"))
		kind := c.Query("kind", "")
		search := c.Query("search", "")
		tagsStr := c.Query("tags", "")

		var tags []string
		if tagsStr != "" {
			tags = strings.Split(tagsStr, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}

		ctx := context.Background()

		// Query database
		query := database.WorkflowQuery{
			Tags:   tags,
			Kind:   kind,
			Search: search,
			Offset: offset,
			Limit:  limit,
		}

		result, err := database.ListWorkflowsFromDB(ctx, query)
		if err != nil {
			// Fallback to filesystem if DB query fails
			return listWorkflowsFromFilesystem(c, cfg)
		}

		// If no workflows in DB, suggest indexing
		if result.TotalCount == 0 {
			// Check if DB is empty vs no matches
			totalCount, _ := database.GetWorkflowCount(ctx)
			if totalCount == 0 {
				return c.JSON(fiber.Map{
					"data":    []interface{}{},
					"count":   0,
					"message": "No workflows indexed. Run 'osmedeus db index workflow' or POST /osm/api/workflows/refresh to index workflows.",
				})
			}
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

// listWorkflowsFromFilesystem loads workflows directly from filesystem
func listWorkflowsFromFilesystem(c *fiber.Ctx, cfg *config.Config) error {
	loader := parser.NewLoader(cfg.WorkflowsPath)
	workflows, err := loader.LoadAllWorkflows()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "Failed to load workflows",
		})
	}

	data := make([]fiber.Map, 0, len(workflows))
	for _, w := range workflows {
		// Build params detail
		params := make([]fiber.Map, 0, len(w.Params))
		requiredParams := []string{}
		for _, p := range w.Params {
			params = append(params, fiber.Map{
				"name":      p.Name,
				"default":   p.Default,
				"required":  p.Required,
				"generator": p.Generator,
			})
			if p.Required {
				requiredParams = append(requiredParams, p.Name)
			}
		}

		data = append(data, fiber.Map{
			"name":            w.Name,
			"kind":            w.Kind,
			"description":     w.Description,
			"tags":            w.Tags,
			"file_path":       w.FilePath,
			"params":          params,
			"required_params": requiredParams,
			"step_count":      len(w.Steps),
			"module_count":    len(w.Modules),
		})
	}

	return c.JSON(fiber.Map{
		"data":  data,
		"count": len(data),
	})
}

// GetWorkflowVerbose handles getting a single workflow with verbose details
// @Summary Get workflow details
// @Description Get workflow content. Returns raw YAML by default. Use json=true to get JSON with parsed details.
// @Tags Workflows
// @Produce json,text/yaml
// @Param name path string true "Workflow name"
// @Param json query bool false "Return JSON with parsed details instead of raw YAML"
// @Success 200 {object} map[string]interface{} "Workflow details (JSON) or raw YAML content"
// @Failure 404 {object} map[string]interface{} "Workflow not found"
// @Security BearerAuth
// @Router /osm/api/workflows/{name} [get]
func GetWorkflowVerbose(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.Params("name")
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflow, err := loader.LoadWorkflow(name)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Workflow not found",
			})
		}

		// Check if JSON format is requested (default is YAML)
		if c.Query("json") == "true" {
			return returnWorkflowJSON(c, workflow)
		}

		// Return YAML by default
		content, err := os.ReadFile(workflow.FilePath)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to read workflow file",
			})
		}
		c.Set("Content-Type", "text/yaml; charset=utf-8")
		return c.SendString(string(content))
	}
}

// returnWorkflowJSON returns workflow details as JSON
func returnWorkflowJSON(c *fiber.Ctx, workflow *core.Workflow) error {
	// Build params detail
	params := make([]fiber.Map, 0, len(workflow.Params))
	for _, p := range workflow.Params {
		params = append(params, fiber.Map{
			"name":      p.Name,
			"default":   p.Default,
			"required":  p.Required,
			"generator": p.Generator,
		})
	}

	// Build steps detail with full information
	steps := make([]fiber.Map, 0, len(workflow.Steps))
	for i, s := range workflow.Steps {
		stepMap := fiber.Map{
			"index":         i,
			"name":          s.Name,
			"type":          s.Type,
			"command":       s.Command,
			"timeout":       s.Timeout,
			"pre_condition": s.PreCondition,
			"step_runner":   s.StepRunner,
		}
		// Add exports if present
		if len(s.Exports) > 0 {
			stepMap["exports"] = s.Exports
		}
		// Add step runner config if present
		if s.StepRunnerConfig != nil && s.StepRunnerConfig.RunnerConfig != nil {
			stepMap["step_runner_config"] = fiber.Map{
				"image":   s.StepRunnerConfig.Image,
				"host":    s.StepRunnerConfig.Host,
				"user":    s.StepRunnerConfig.User,
				"volumes": s.StepRunnerConfig.Volumes,
			}
		}
		// Add parallel steps if present
		if len(s.ParallelSteps) > 0 {
			substeps := make([]fiber.Map, 0, len(s.ParallelSteps))
			for j, ss := range s.ParallelSteps {
				substeps = append(substeps, fiber.Map{
					"index":   j,
					"name":    ss.Name,
					"type":    ss.Type,
					"command": ss.Command,
				})
			}
			stepMap["parallel_steps"] = substeps
		}
		steps = append(steps, stepMap)
	}

	// Build modules detail with full information
	modules := make([]fiber.Map, 0, len(workflow.Modules))
	for i, m := range workflow.Modules {
		modules = append(modules, fiber.Map{
			"index":      i,
			"name":       m.Name,
			"path":       m.Path,
			"depends_on": m.DependsOn,
			"condition":  m.Condition,
		})
	}

	// Build triggers detail
	triggers := make([]fiber.Map, 0, len(workflow.Triggers))
	for _, t := range workflow.Triggers {
		triggers = append(triggers, fiber.Map{
			"name":     t.Name,
			"on":       t.On,
			"schedule": t.Schedule,
			"enabled":  t.Enabled,
		})
	}

	// Build dependencies info
	var dependencies fiber.Map
	if workflow.Dependencies != nil {
		dependencies = fiber.Map{
			"commands": workflow.Dependencies.Commands,
			"files":    workflow.Dependencies.Files,
		}
	}

	return c.JSON(fiber.Map{
		"name":         workflow.Name,
		"kind":         workflow.Kind,
		"description":  workflow.Description,
		"file_path":    workflow.FilePath,
		"params":       params,
		"steps":        steps,
		"modules":      modules,
		"triggers":     triggers,
		"dependencies": dependencies,
	})
}

// ValidateWorkflow handles workflow validation
func ValidateWorkflow(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		name := c.Params("name")
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflow, err := loader.LoadWorkflow(name)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"message": "Workflow not found",
			})
		}

		if err := parser.Validate(workflow); err != nil {
			return c.JSON(fiber.Map{
				"valid":   false,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"valid":   true,
			"message": "Workflow is valid",
		})
	}
}

// ReloadWorkflows handles reloading workflows from disk
func ReloadWorkflows(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		loader := parser.NewLoader(cfg.WorkflowsPath)
		if err := loader.ReloadWorkflows(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": "Failed to reload workflows",
			})
		}

		return c.JSON(fiber.Map{
			"message": "Workflows reloaded",
		})
	}
}

// RefreshWorkflowIndex re-indexes workflows from filesystem to database
// @Summary Refresh workflow index
// @Description Re-index all workflows from filesystem to database
// @Tags Workflows
// @Accept json
// @Produce json
// @Param force query bool false "Force re-index all workflows regardless of checksum"
// @Success 200 {object} map[string]interface{} "Indexing result"
// @Failure 500 {object} map[string]interface{} "Indexing failed"
// @Security BearerAuth
// @Router /osm/api/workflows/refresh [post]
func RefreshWorkflowIndex(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		force := c.Query("force") == "true"

		ctx := context.Background()

		// Index workflows
		result, err := database.IndexWorkflowsFromFilesystem(ctx, cfg.WorkflowsPath, force)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"message": "Workflows indexed successfully",
			"added":   result.Added,
			"updated": result.Updated,
			"removed": result.Removed,
			"errors":  result.Errors,
		})
	}
}

// GetAllWorkflowTags returns all unique tags from indexed workflows
// @Summary Get all workflow tags
// @Description Get all unique tags from indexed workflows in database
// @Tags Workflows
// @Produce json
// @Success 200 {object} map[string]interface{} "List of tags"
// @Failure 500 {object} map[string]interface{} "Failed to get tags"
// @Security BearerAuth
// @Router /osm/api/workflows/tags [get]
func GetAllWorkflowTags(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		tags, err := database.GetAllTags(ctx)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"tags":  tags,
			"count": len(tags),
		})
	}
}

// workflowToMap converts a workflow to a map for JSON response
func workflowToMap(w *core.Workflow) fiber.Map {
	params := make([]fiber.Map, 0, len(w.Params))
	for _, p := range w.Params {
		params = append(params, fiber.Map{
			"name":      p.Name,
			"default":   p.Default,
			"required":  p.Required,
			"generator": p.Generator,
		})
	}

	triggers := make([]fiber.Map, 0, len(w.Triggers))
	for _, t := range w.Triggers {
		triggers = append(triggers, fiber.Map{
			"name":     t.Name,
			"on":       t.On,
			"schedule": t.Schedule,
			"enabled":  t.Enabled,
		})
	}

	return fiber.Map{
		"name":        w.Name,
		"kind":        w.Kind,
		"description": w.Description,
		"file_path":   w.FilePath,
		"params":      params,
		"triggers":    triggers,
	}
}
