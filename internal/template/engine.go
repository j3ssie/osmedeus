package template

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/flosch/pongo2/v6"
	lru "github.com/hashicorp/golang-lru/v2"
)

// DefaultCacheSize is the default number of parsed templates to cache
const DefaultCacheSize = 1024

// Pre-compiled regex patterns (compile once at package init)
var (
	generatorExprPattern = regexp.MustCompile(`^(\w+)\((.*)\)$`)
	variablePattern      = regexp.MustCompile(`\{\{\s*(\w+)\s*\}\}`)
)

// Engine wraps pongo2 template engine with custom functionality
type Engine struct {
	mu            sync.Mutex
	generators    map[string]GeneratorFunc
	templateCache *lru.Cache[string, *pongo2.Template]
}

// NewEngine creates a new template engine with default cache size
func NewEngine() *Engine {
	return NewEngineWithCacheSize(DefaultCacheSize)
}

// NewEngineWithCacheSize creates a new template engine with specified cache size
func NewEngineWithCacheSize(cacheSize int) *Engine {
	cache, _ := lru.New[string, *pongo2.Template](cacheSize)
	e := &Engine{
		generators:    make(map[string]GeneratorFunc),
		templateCache: cache,
	}
	e.registerBuiltinGenerators()
	return e
}

// Render renders a template string with the given context
// Uses LRU cache to avoid re-parsing the same template strings
func (e *Engine) Render(template string, ctx map[string]any) (string, error) {
	// Quick path: no template variables present
	if !strings.Contains(template, "{{") {
		return template, nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	// Check cache for pre-parsed template
	tpl, ok := e.templateCache.Get(template)
	if !ok {
		// Parse and cache the template
		var err error
		tpl, err = pongo2.FromString(template)
		if err != nil {
			return "", fmt.Errorf("template parse error: %w", err)
		}
		e.templateCache.Add(template, tpl)
	}

	// Preprocess context: convert bool values to lowercase strings
	// This ensures {{bool_var}} renders as "true"/"false" not "True"/"False"
	processedCtx := normalizeBoolsForTemplate(ctx)

	result, err := tpl.Execute(pongo2.Context(processedCtx))
	if err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	// Fix incomplete expressions caused by undefined variables
	result = fixIncompleteExpressions(result)

	return result, nil
}

// normalizeBoolsForTemplate converts bool values to lowercase strings for template rendering
// This ensures consistent "true"/"false" output instead of pongo2's "True"/"False"
func normalizeBoolsForTemplate(ctx map[string]any) map[string]any {
	result := make(map[string]any, len(ctx))
	for k, v := range ctx {
		if b, ok := v.(bool); ok {
			if b {
				result[k] = "true"
			} else {
				result[k] = "false"
			}
		} else {
			result[k] = v
		}
	}
	return result
}

// Pre-compiled regex patterns for incomplete expression fixing
var incompleteExprPatterns = []struct {
	pattern *regexp.Regexp
	replace string
}{
	// Comparison operators followed by ) - numeric context
	{regexp.MustCompile(`>\s*\)`), "> 0)"},
	{regexp.MustCompile(`<\s*\)`), "< 0)"},
	{regexp.MustCompile(`>=\s*\)`), ">= 0)"},
	{regexp.MustCompile(`<=\s*\)`), "<= 0)"},
	{regexp.MustCompile(`==\s*\)`), "== 0)"},
	{regexp.MustCompile(`!=\s*\)`), "!= 0)"},
	// Comparison operators at end of string - numeric context
	{regexp.MustCompile(`>\s*$`), "> 0"},
	{regexp.MustCompile(`<\s*$`), "< 0"},
	{regexp.MustCompile(`>=\s*$`), ">= 0"},
	{regexp.MustCompile(`<=\s*$`), "<= 0"},
	// Equality at end of string - use empty string for safer default
	{regexp.MustCompile(`==\s*$`), "== ''"},
	{regexp.MustCompile(`!=\s*$`), "!= ''"},
}

// fixIncompleteExpressions handles cases where undefined template variables
// leave incomplete expressions (e.g., "> " becomes "> 0")
// This prevents JavaScript syntax errors when evaluating pre_conditions
func fixIncompleteExpressions(s string) string {
	for _, p := range incompleteExprPatterns {
		s = p.pattern.ReplaceAllString(s, p.replace)
	}
	return s
}

// RenderMap renders all template values in a map
func (e *Engine) RenderMap(m map[string]string, ctx map[string]any) (map[string]string, error) {
	result := make(map[string]string, len(m))
	for k, v := range m {
		rendered, err := e.Render(v, ctx)
		if err != nil {
			return nil, fmt.Errorf("error rendering %s: %w", k, err)
		}
		result[k] = rendered
	}
	return result, nil
}

// RenderSlice renders all template values in a slice
func (e *Engine) RenderSlice(s []string, ctx map[string]any) ([]string, error) {
	result := make([]string, len(s))
	for i, v := range s {
		rendered, err := e.Render(v, ctx)
		if err != nil {
			return nil, fmt.Errorf("error rendering index %d: %w", i, err)
		}
		result[i] = rendered
	}
	return result, nil
}

// RenderSecondary renders templates using [[ ]] delimiters
// Used for variables that only exist at runtime (e.g., foreach loop variables)
func (e *Engine) RenderSecondary(template string, ctx map[string]any) (string, error) {
	// Quick path: no secondary delimiters
	if !strings.Contains(template, "[[") {
		return template, nil
	}

	// Convert [[ ]] to {{ }} for pongo2 processing
	converted := strings.ReplaceAll(template, "[[", "{{")
	converted = strings.ReplaceAll(converted, "]]", "}}")

	return e.Render(converted, ctx)
}

// HasSecondaryVariable checks if template contains [[ ]] delimiters
func (e *Engine) HasSecondaryVariable(template string) bool {
	return strings.Contains(template, "[[") && strings.Contains(template, "]]")
}

// ExecuteGenerator executes a generator function expression
func (e *Engine) ExecuteGenerator(expr string) (string, error) {
	// Parse generator expression: funcName(args...)
	// Examples: uuid(), currentDate("2006-01-02"), getEnvVar("KEY", "default")
	name, args, err := e.parseGeneratorExpr(expr)
	if err != nil {
		return "", err
	}

	gen, ok := e.generators[name]
	if !ok {
		return "", fmt.Errorf("unknown generator function: %s", name)
	}

	return gen(args...)
}

// parseGeneratorExpr parses a generator expression into function name and arguments
func (e *Engine) parseGeneratorExpr(expr string) (string, []string, error) {
	// Match: funcName(arg1, arg2, ...)
	matches := generatorExprPattern.FindStringSubmatch(strings.TrimSpace(expr))
	if len(matches) != 3 {
		return "", nil, fmt.Errorf("invalid generator expression: %s", expr)
	}

	funcName := matches[1]
	argsStr := strings.TrimSpace(matches[2])

	if argsStr == "" {
		return funcName, nil, nil
	}

	// Parse arguments (handle quoted strings)
	args := e.parseArgs(argsStr)
	return funcName, args, nil
}

// parseArgs parses comma-separated arguments, handling quoted strings
func (e *Engine) parseArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range argsStr {
		switch {
		case (ch == '"' || ch == '\'') && !inQuote:
			inQuote = true
			quoteChar = ch
		case ch == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
		case ch == ',' && !inQuote:
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}

	return args
}

// RegisterGenerator registers a custom generator function
func (e *Engine) RegisterGenerator(name string, fn GeneratorFunc) {
	e.generators[name] = fn
}

// HasVariable checks if a string contains template variables
func HasVariable(s string) bool {
	return strings.Contains(s, "{{") && strings.Contains(s, "}}")
}

// ExtractVariables extracts variable names from a template string
func ExtractVariables(s string) []string {
	matches := variablePattern.FindAllStringSubmatch(s, -1)

	var vars []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			vars = append(vars, match[1])
			seen[match[1]] = true
		}
	}
	return vars
}

// RenderBatch renders multiple templates in a single operation.
// For the standard Engine, this simply iterates through requests.
// Use ShardedEngine for optimized batch rendering.
func (e *Engine) RenderBatch(requests []RenderRequest, ctx map[string]any) (map[string]string, error) {
	if len(requests) == 0 {
		return make(map[string]string), nil
	}

	results := make(map[string]string, len(requests))
	for _, req := range requests {
		rendered, err := e.Render(req.Template, ctx)
		if err != nil {
			return nil, fmt.Errorf("error rendering %s: %w", req.Key, err)
		}
		results[req.Key] = rendered
	}
	return results, nil
}

// Verify interface compliance at compile time
var _ TemplateEngine = (*Engine)(nil)
var _ BatchRenderer = (*Engine)(nil)
