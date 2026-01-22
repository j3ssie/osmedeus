package linter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// Built-in variable names that are always available
// These match the variables injected by injectBuiltinVariables() in executor.go
var builtInVariables = map[string]bool{
	// Path Variables
	"BaseFolder":           true,
	"Binaries":             true,
	"binaries":             true,
	"Data":                 true,
	"data":                 true,
	"ExternalData":         true,
	"ExternalConfigs":      true,
	"ExternalAgentConfigs": true,
	"ExternalAgents":       true,
	"ExternalScripts":      true,
	"Workflows":            true,
	"MarkdownTemplates":    true,
	"ExternalMarkdowns":    true,
	"SnapshotsFolder":      true,
	"Workspaces":           true,

	// Target Variables
	"Target":      true,
	"target":      true,
	"TargetFile":  true,
	"TargetSpace": true,

	// Output/Workspace Variables
	"Output":    true,
	"output":    true,
	"Workspace": true,
	"workspace": true,

	// Thread Variables
	"threads":     true,
	"Threads":     true,
	"baseThreads": true,

	// Metadata Variables
	"Version":      true,
	"TaskDate":     true,
	"RunUUID":      true,
	"TimeStamp":    true,
	"CurrentTime":  true,
	"Today":        true,
	"RandomString": true,

	// State File Variables
	"StateExecutionLog":   true,
	"StateConsoleLog":     true,
	"StateCompletedFile":  true,
	"StateFile":           true,
	"StateWorkflowFile":   true,
	"StateWorkflowFolder": true,

	// Heuristic Variables (Target Type Detection)
	"TargetType":          true,
	"TargetRootDomain":    true,
	"TargetTLD":           true,
	"TargetSLD":           true,
	"Org":                 true,
	"TargetBaseURL":       true,
	"TargetRootURL":       true,
	"TargetHostname":      true,
	"TargetHost":          true,
	"TargetPort":          true,
	"TargetPath":          true,
	"TargetFileExt":       true,
	"TargetScheme":        true,
	"TargetIsWildcard":    true,
	"TargetResolvedIP":    true,
	"TargetStatusCode":    true,
	"TargetContentLength": true,
	"HeuristicsCheck":     true,

	// Chunk Mode Variables
	"ChunkIndex":  true,
	"ChunkSize":   true,
	"TotalChunks": true,
	"ChunkStart":  true,
	"ChunkEnd":    true,

	// Legacy/Aliases (for backward compatibility)
	"Base":     true,
	"base":     true,
	"Home":     true,
	"home":     true,
	"Storages": true,
	"storages": true,
	"Scripts":  true,
	"scripts":  true,
	"Cloud":    true,
	"cloud":    true,
	"RunID":    true,
	"run_id":   true,
}

// Regex patterns for variable extraction
var (
	// Standard template variables: {{variable}}
	templateVarPattern = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	// Foreach loop variables: [[variable]]
	foreachVarPattern = regexp.MustCompile(`\[\[([a-zA-Z_][a-zA-Z0-9_]*)\]\]`)
)

// UnusedVariableRule checks for variables exported but never used
type UnusedVariableRule struct{}

func (r *UnusedVariableRule) Name() string        { return "unused-variable" }
func (r *UnusedVariableRule) Description() string { return "Detects variables exported but never used" }
func (r *UnusedVariableRule) Severity() Severity  { return SeverityInfo }

func (r *UnusedVariableRule) Check(wast *WorkflowAST) []LintIssue {
	var issues []LintIssue
	w := wast.Workflow

	// Collect all exported variables with their positions
	exports := make(map[string]struct {
		stepIndex int
		stepName  string
	})
	for i, step := range w.Steps {
		for exportName := range step.Exports {
			exports[exportName] = struct {
				stepIndex int
				stepName  string
			}{i, step.Name}
		}
	}

	// Collect all referenced variables
	referenced := make(map[string]bool)
	for i, step := range w.Steps {
		// Check all string fields for variable references
		collectReferencedVars(&step, i, referenced, w.Steps)
	}

	// Also check params for references
	for _, param := range w.Params {
		if defaultStr := param.DefaultString(); defaultStr != "" {
			for _, v := range extractVariables(defaultStr) {
				referenced[v] = true
			}
		}
	}

	// Find unused exports
	for exportName, info := range exports {
		if !referenced[exportName] && !builtInVariables[exportName] {
			line, col := wast.FindExportPosition(info.stepIndex, exportName)
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   r.Severity(),
				Message:    fmt.Sprintf("Export '%s' is never referenced in subsequent steps", exportName),
				Suggestion: "Remove unused export or use it in a later step",
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d].exports.%s", info.stepIndex, exportName),
			})
		}
	}

	return issues
}

// UndefinedVariableRule checks for variables referenced but not defined
type UndefinedVariableRule struct{}

func (r *UndefinedVariableRule) Name() string        { return "undefined-variable" }
func (r *UndefinedVariableRule) Description() string { return "Detects variables referenced but not defined" }
func (r *UndefinedVariableRule) Severity() Severity  { return SeverityWarning }

func (r *UndefinedVariableRule) Check(wast *WorkflowAST) []LintIssue {
	var issues []LintIssue
	w := wast.Workflow

	// Collect all defined variables (params, exports from previous steps)
	defined := make(map[string]bool)
	for k := range builtInVariables {
		defined[k] = true
	}

	// Add params
	for _, param := range w.Params {
		defined[param.Name] = true
	}

	// Check each step for undefined variables
	for i, step := range w.Steps {
		// Check command
		checkStringForUndefinedVars(step.Command, fmt.Sprintf("steps[%d].command", i), defined, wast, r, &issues)

		// Check commands array
		for j, cmd := range step.Commands {
			checkStringForUndefinedVars(cmd, fmt.Sprintf("steps[%d].commands[%d]", i, j), defined, wast, r, &issues)
		}

		// Check parallel commands
		for j, cmd := range step.ParallelCommands {
			checkStringForUndefinedVars(cmd, fmt.Sprintf("steps[%d].parallel_commands[%d]", i, j), defined, wast, r, &issues)
		}

		// Check function fields
		checkStringForUndefinedVars(step.Function, fmt.Sprintf("steps[%d].function", i), defined, wast, r, &issues)
		for j, fn := range step.Functions {
			checkStringForUndefinedVars(fn, fmt.Sprintf("steps[%d].functions[%d]", i, j), defined, wast, r, &issues)
		}

		// Check pre_condition
		checkStringForUndefinedVars(step.PreCondition, fmt.Sprintf("steps[%d].pre_condition", i), defined, wast, r, &issues)

		// Check input for foreach
		checkStringForUndefinedVars(step.Input, fmt.Sprintf("steps[%d].input", i), defined, wast, r, &issues)

		// Check URL for HTTP steps
		checkStringForUndefinedVars(step.URL, fmt.Sprintf("steps[%d].url", i), defined, wast, r, &issues)

		// Check export values
		for exportName, exportValue := range step.Exports {
			checkStringForUndefinedVars(exportValue, fmt.Sprintf("steps[%d].exports.%s", i, exportName), defined, wast, r, &issues)
		}

		// Check decision switch
		if step.Decision != nil {
			checkStringForUndefinedVars(step.Decision.Switch, fmt.Sprintf("steps[%d].decision.switch", i), defined, wast, r, &issues)
		}

		// After processing this step, add its exports to defined
		for exportName := range step.Exports {
			defined[exportName] = true
		}

		// Add foreach variable to defined for nested step
		if step.Variable != "" {
			defined[step.Variable] = true
		}
	}

	return issues
}

// CircularDependencyRule checks for circular step dependencies
type CircularDependencyRule struct{}

func (r *CircularDependencyRule) Name() string        { return "circular-dependency" }
func (r *CircularDependencyRule) Description() string { return "Detects circular references in step dependencies" }
func (r *CircularDependencyRule) Severity() Severity  { return SeverityWarning }

func (r *CircularDependencyRule) Check(wast *WorkflowAST) []LintIssue {
	var issues []LintIssue
	w := wast.Workflow

	// Build dependency graph
	deps := make(map[string][]string)
	stepIndex := make(map[string]int)
	for i, step := range w.Steps {
		deps[step.Name] = step.DependsOn
		stepIndex[step.Name] = i
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var path []string

	var hasCycle func(step string) bool
	hasCycle = func(step string) bool {
		visited[step] = true
		recStack[step] = true
		path = append(path, step)

		for _, dep := range deps[step] {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				// Found cycle
				cycleStart := -1
				for i, p := range path {
					if p == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := append(path[cycleStart:], dep)
					idx := stepIndex[step]
					line, col := wast.FindStepFieldPosition(idx, "depends_on")
					issues = append(issues, LintIssue{
						Rule:       r.Name(),
						Severity:   r.Severity(),
						Message:    fmt.Sprintf("Circular dependency detected: %s", strings.Join(cycle, " -> ")),
						Suggestion: "Remove one of the dependencies to break the cycle",
						Line:       line,
						Column:     col,
						Field:      fmt.Sprintf("steps[%d].depends_on", idx),
					})
				}
				return true
			}
		}

		path = path[:len(path)-1]
		recStack[step] = false
		return false
	}

	for _, step := range w.Steps {
		if !visited[step.Name] {
			hasCycle(step.Name)
		}
	}

	return issues
}

// EmptyStepRule checks for steps with no executable content
type EmptyStepRule struct{}

func (r *EmptyStepRule) Name() string        { return "empty-step" }
func (r *EmptyStepRule) Description() string { return "Detects steps with no executable content" }
func (r *EmptyStepRule) Severity() Severity  { return SeverityWarning }

func (r *EmptyStepRule) Check(wast *WorkflowAST) []LintIssue {
	var issues []LintIssue
	w := wast.Workflow

	for i, step := range w.Steps {
		empty := false
		switch step.Type {
		case core.StepTypeBash, core.StepTypeRemoteBash:
			if step.Command == "" && len(step.Commands) == 0 && len(step.ParallelCommands) == 0 {
				empty = true
			}
		case core.StepTypeFunction:
			if step.Function == "" && len(step.Functions) == 0 && len(step.ParallelFunctions) == 0 {
				empty = true
			}
		case core.StepTypeParallel:
			if len(step.ParallelSteps) == 0 {
				empty = true
			}
		case core.StepTypeForeach:
			if step.Step == nil {
				empty = true
			}
		case core.StepTypeHTTP:
			if step.URL == "" {
				empty = true
			}
		case core.StepTypeLLM:
			if len(step.Messages) == 0 && len(step.EmbeddingInput) == 0 {
				empty = true
			}
		}

		if empty {
			line, col := wast.FindStepPosition(step.Name)
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   r.Severity(),
				Message:    fmt.Sprintf("Step '%s' has no executable content", step.Name),
				Suggestion: "Add a command, function, or other executable content to the step",
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d]", i),
			})
		}
	}

	return issues
}

// InvalidGotoRule checks for decision goto references to non-existent steps
type InvalidGotoRule struct{}

func (r *InvalidGotoRule) Name() string        { return "invalid-goto" }
func (r *InvalidGotoRule) Description() string { return "Detects decision goto references to non-existent steps" }
func (r *InvalidGotoRule) Severity() Severity  { return SeverityWarning }

func (r *InvalidGotoRule) Check(wast *WorkflowAST) []LintIssue {
	var issues []LintIssue
	w := wast.Workflow

	// Build set of valid step names
	validSteps := make(map[string]bool)
	validSteps["_end"] = true // Special value to end workflow
	for _, step := range w.Steps {
		validSteps[step.Name] = true
	}

	// Check each decision
	for i, step := range w.Steps {
		if step.Decision == nil {
			continue
		}

		// Check cases
		for caseValue, caseConfig := range step.Decision.Cases {
			if caseConfig.Goto != "" && !validSteps[caseConfig.Goto] {
				line, col := wast.FindStepFieldPosition(i, "decision")
				suggestion := "Use one of: " + strings.Join(getStepNames(w.Steps), ", ") + ", or _end"
				issues = append(issues, LintIssue{
					Rule:       r.Name(),
					Severity:   r.Severity(),
					Message:    fmt.Sprintf("Decision case '%s' references non-existent step '%s'", caseValue, caseConfig.Goto),
					Suggestion: suggestion,
					Line:       line,
					Column:     col,
					Field:      fmt.Sprintf("steps[%d].decision.cases.%s.goto", i, caseValue),
				})
			}
		}

		// Check default
		if step.Decision.Default != nil && step.Decision.Default.Goto != "" && !validSteps[step.Decision.Default.Goto] {
			line, col := wast.FindStepFieldPosition(i, "decision")
			suggestion := "Use one of: " + strings.Join(getStepNames(w.Steps), ", ") + ", or _end"
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   r.Severity(),
				Message:    fmt.Sprintf("Decision default references non-existent step '%s'", step.Decision.Default.Goto),
				Suggestion: suggestion,
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d].decision.default.goto", i),
			})
		}
	}

	return issues
}

// DuplicateStepNameRule checks for multiple steps with the same name
type DuplicateStepNameRule struct{}

func (r *DuplicateStepNameRule) Name() string        { return "duplicate-step-name" }
func (r *DuplicateStepNameRule) Description() string { return "Detects multiple steps with the same name" }
func (r *DuplicateStepNameRule) Severity() Severity  { return SeverityWarning }

func (r *DuplicateStepNameRule) Check(wast *WorkflowAST) []LintIssue {
	var issues []LintIssue
	w := wast.Workflow

	seen := make(map[string]int) // name -> first occurrence index
	for i, step := range w.Steps {
		if firstIdx, exists := seen[step.Name]; exists {
			line, col := wast.FindStepPosition(step.Name)
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   r.Severity(),
				Message:    fmt.Sprintf("Duplicate step name '%s' (first defined at step %d)", step.Name, firstIdx+1),
				Suggestion: "Use unique names for each step",
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d].name", i),
			})
		} else {
			seen[step.Name] = i
		}
	}

	return issues
}

// MissingRequiredFieldRule checks for required fields that are missing
type MissingRequiredFieldRule struct{}

func (r *MissingRequiredFieldRule) Name() string        { return "missing-required-field" }
func (r *MissingRequiredFieldRule) Description() string { return "Detects required fields that are missing" }
func (r *MissingRequiredFieldRule) Severity() Severity  { return SeverityWarning }

func (r *MissingRequiredFieldRule) Check(wast *WorkflowAST) []LintIssue {
	var issues []LintIssue
	w := wast.Workflow

	// Check workflow-level required fields
	if w.Name == "" {
		line, col := wast.GetNodePosition("name")
		if line == 0 {
			line = 1
		}
		issues = append(issues, LintIssue{
			Rule:       r.Name(),
			Severity:   r.Severity(),
			Message:    "Workflow name is required",
			Suggestion: "Add 'name: your-workflow-name' to the workflow",
			Line:       line,
			Column:     col,
			Field:      "name",
		})
	}

	if w.Kind == "" {
		line, col := wast.GetNodePosition("kind")
		if line == 0 {
			line = 1
		}
		issues = append(issues, LintIssue{
			Rule:       r.Name(),
			Severity:   r.Severity(),
			Message:    "Workflow kind is required",
			Suggestion: "Add 'kind: module' or 'kind: flow' to the workflow",
			Line:       line,
			Column:     col,
			Field:      "kind",
		})
	}

	// Check step-level required fields
	for i, step := range w.Steps {
		if step.Name == "" {
			line, col := wast.FindStepFieldPosition(i, "name")
			if line == 0 {
				line, col = wast.GetNodePosition(fmt.Sprintf("steps[%d]", i))
			}
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   r.Severity(),
				Message:    fmt.Sprintf("Step at index %d is missing required 'name' field", i),
				Suggestion: "Add 'name: step-name' to the step",
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d].name", i),
			})
		}

		if step.Type == "" {
			line, col := wast.FindStepFieldPosition(i, "type")
			if line == 0 {
				line, col = wast.GetNodePosition(fmt.Sprintf("steps[%d]", i))
			}
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   r.Severity(),
				Message:    fmt.Sprintf("Step '%s' is missing required 'type' field", step.Name),
				Suggestion: "Add 'type: bash', 'type: function', or another valid step type",
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d].type", i),
			})
		}
	}

	return issues
}

// InvalidDependsOnRule checks for depends_on referencing non-existent steps
type InvalidDependsOnRule struct{}

func (r *InvalidDependsOnRule) Name() string        { return "invalid-depends-on" }
func (r *InvalidDependsOnRule) Description() string { return "Detects depends_on references to non-existent steps" }
func (r *InvalidDependsOnRule) Severity() Severity  { return SeverityWarning }

func (r *InvalidDependsOnRule) Check(wast *WorkflowAST) []LintIssue {
	var issues []LintIssue
	w := wast.Workflow

	// Build set of valid step names
	validSteps := make(map[string]bool)
	for _, step := range w.Steps {
		validSteps[step.Name] = true
	}

	// Check each step's depends_on
	for i, step := range w.Steps {
		for _, dep := range step.DependsOn {
			if !validSteps[dep] {
				line, col := wast.FindStepFieldPosition(i, "depends_on")
				suggestion := findSimilarStep(dep, w.Steps)
				issues = append(issues, LintIssue{
					Rule:       r.Name(),
					Severity:   r.Severity(),
					Message:    fmt.Sprintf("Step '%s' depends on non-existent step '%s'", step.Name, dep),
					Suggestion: suggestion,
					Line:       line,
					Column:     col,
					Field:      fmt.Sprintf("steps[%d].depends_on", i),
				})
			}
		}
	}

	return issues
}

// Helper functions

func extractVariables(s string) []string {
	var vars []string
	// Extract standard template variables
	matches := templateVarPattern.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		if len(m) > 1 {
			vars = append(vars, m[1])
		}
	}
	// Extract foreach variables
	matches = foreachVarPattern.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		if len(m) > 1 {
			vars = append(vars, m[1])
		}
	}
	return vars
}

func checkStringForUndefinedVars(s, field string, defined map[string]bool, wast *WorkflowAST, rule *UndefinedVariableRule, issues *[]LintIssue) {
	if s == "" {
		return
	}

	for _, v := range extractVariables(s) {
		if !defined[v] && !builtInVariables[v] {
			line, col := wast.GetNodePosition(field)
			suggestion := findSimilarVariable(v, defined)
			*issues = append(*issues, LintIssue{
				Rule:       rule.Name(),
				Severity:   rule.Severity(),
				Message:    fmt.Sprintf("Variable '%s' is not defined", v),
				Suggestion: suggestion,
				Line:       line,
				Column:     col,
				Field:      field,
			})
		}
	}
}

func collectReferencedVars(step *core.Step, _ int, referenced map[string]bool, _ []core.Step) {
	// Collect from all string fields
	for _, v := range extractVariables(step.Command) {
		referenced[v] = true
	}
	for _, cmd := range step.Commands {
		for _, v := range extractVariables(cmd) {
			referenced[v] = true
		}
	}
	for _, cmd := range step.ParallelCommands {
		for _, v := range extractVariables(cmd) {
			referenced[v] = true
		}
	}
	for _, v := range extractVariables(step.Function) {
		referenced[v] = true
	}
	for _, fn := range step.Functions {
		for _, v := range extractVariables(fn) {
			referenced[v] = true
		}
	}
	for _, v := range extractVariables(step.PreCondition) {
		referenced[v] = true
	}
	for _, v := range extractVariables(step.Input) {
		referenced[v] = true
	}
	for _, v := range extractVariables(step.URL) {
		referenced[v] = true
	}
	for _, exportValue := range step.Exports {
		for _, v := range extractVariables(exportValue) {
			referenced[v] = true
		}
	}
	if step.Decision != nil {
		for _, v := range extractVariables(step.Decision.Switch) {
			referenced[v] = true
		}
	}
}

func findSimilarVariable(v string, defined map[string]bool) string {
	vLower := strings.ToLower(v)
	for d := range defined {
		if strings.ToLower(d) == vLower {
			return fmt.Sprintf("Did you mean '%s'?", d)
		}
	}
	for b := range builtInVariables {
		if strings.ToLower(b) == vLower {
			return fmt.Sprintf("Did you mean '%s'?", b)
		}
	}
	return "Check that the variable is defined in params or a previous step's exports"
}

func findSimilarStep(name string, steps []core.Step) string {
	nameLower := strings.ToLower(name)
	for _, s := range steps {
		if strings.ToLower(s.Name) == nameLower {
			return fmt.Sprintf("Did you mean '%s'?", s.Name)
		}
	}
	return "Valid steps: " + strings.Join(getStepNames(steps), ", ")
}

func getStepNames(steps []core.Step) []string {
	names := make([]string, len(steps))
	for i, s := range steps {
		names[i] = s.Name
	}
	return names
}

// GetDefaultRules returns all built-in linting rules
func GetDefaultRules() []LinterRule {
	return []LinterRule{
		&MissingRequiredFieldRule{},
		&DuplicateStepNameRule{},
		&EmptyStepRule{},
		&UnusedVariableRule{},
		&InvalidGotoRule{},
		&InvalidDependsOnRule{},
		&CircularDependencyRule{},
	}
}
