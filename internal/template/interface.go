package template

// TemplateEngine defines the interface for template rendering engines.
// This allows for different implementations (standard, sharded) to be used
// interchangeably throughout the codebase.
type TemplateEngine interface {
	// Render renders a template string with the given context.
	// Returns the rendered string or an error if rendering fails.
	Render(template string, ctx map[string]any) (string, error)

	// RenderMap renders all template values in a map.
	// Returns a new map with rendered values.
	RenderMap(m map[string]string, ctx map[string]any) (map[string]string, error)

	// RenderSlice renders all template values in a slice.
	// Returns a new slice with rendered values.
	RenderSlice(s []string, ctx map[string]any) ([]string, error)

	// RenderSecondary renders templates using [[ ]] delimiters.
	// Used for variables that only exist at runtime (e.g., foreach loop variables).
	RenderSecondary(template string, ctx map[string]any) (string, error)

	// HasSecondaryVariable checks if template contains [[ ]] delimiters.
	HasSecondaryVariable(template string) bool

	// ExecuteGenerator executes a generator function expression.
	ExecuteGenerator(expr string) (string, error)

	// RegisterGenerator registers a custom generator function.
	RegisterGenerator(name string, fn GeneratorFunc)

	// ExtractVariablesSet extracts all variable names referenced in a template string.
	// Returns a set of variable names (map for O(1) lookup).
	ExtractVariablesSet(template string) map[string]struct{}

	// RenderLazy renders a template with lazy context loading.
	// Only variables actually referenced in the template are looked up from the full context.
	RenderLazy(template string, fullCtx map[string]any) (string, error)
}

// BatchRenderer extends TemplateEngine with batch rendering capability
// for improved performance under high concurrency.
type BatchRenderer interface {
	TemplateEngine

	// RenderBatch renders multiple templates in a single operation.
	// This reduces lock contention by grouping templates and acquiring
	// locks fewer times.
	RenderBatch(requests []RenderRequest, ctx map[string]any) (map[string]string, error)
}

// RenderRequest represents a single template to render in a batch operation.
type RenderRequest struct {
	// Key is the identifier for this template (e.g., field name)
	Key string

	// Template is the template string to render
	Template string
}

// Precompiler provides access to pre-compiled templates for workflows.
type Precompiler interface {
	// PrecompileWorkflow scans a workflow and pre-compiles all template strings.
	PrecompileWorkflow(workflowName string, templates map[string]string) error

	// GetPrecompiled retrieves a pre-compiled template if available.
	// Returns nil if not found.
	GetPrecompiled(workflowName, key string) any

	// ClearPrecompiled removes pre-compiled templates for a workflow.
	ClearPrecompiled(workflowName string)
}
