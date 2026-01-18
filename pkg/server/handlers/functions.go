package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/template"
)

// FunctionEval executes a utility function script
// @Summary Execute utility function
// @Description Execute a utility function script with template rendering and JavaScript execution
// @Tags Functions
// @Accept json
// @Produce json
// @Param request body FunctionEvalRequest true "Function evaluation request"
// @Success 200 {object} map[string]interface{} "Evaluation result"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Security BearerAuth
// @Router /osm/api/functions/eval [post]
func FunctionEval(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req FunctionEvalRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid request body",
			})
		}

		if req.Script == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "script is required",
			})
		}

		// Build context with target and params
		ctx := make(map[string]interface{})
		if req.Target != "" {
			ctx["target"] = req.Target
		}
		for k, v := range req.Params {
			ctx[k] = v
		}

		// Render template variables ({{target}}, etc.)
		templateEngine := template.NewEngine()
		renderedScript, err := templateEngine.Render(req.Script, ctx)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Template rendering failed: " + err.Error(),
			})
		}

		// Execute as JavaScript using Otto runtime
		registry := functions.NewRegistry()
		result, err := registry.Execute(renderedScript, ctx)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Execution failed: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"result":          result,
			"rendered_script": renderedScript,
		})
	}
}

// FunctionList returns all available utility functions
// @Summary List utility functions
// @Description Get a flat list of all available utility functions with metadata
// @Tags Functions
// @Produce json
// @Success 200 {object} map[string]interface{} "List of functions with total count"
// @Security BearerAuth
// @Router /osm/api/functions/list [get]
func FunctionList(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get function registry from constants (single source of truth)
		registry := functions.FunctionRegistry()

		// Convert to flat array format
		var result []FunctionListResponse
		for category, funcs := range registry {
			for _, fn := range funcs {
				result = append(result, FunctionListResponse{
					Name:        fn.Signature,
					Description: fn.Description,
					ReturnType:  fn.ReturnType,
					Example:     fn.Example,
					Tags:        []string{category},
				})
			}
		}

		return c.JSON(fiber.Map{
			"functions": result,
			"total":     len(result),
		})
	}
}
