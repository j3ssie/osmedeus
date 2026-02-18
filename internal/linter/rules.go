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
	"ModuleName":   true,
	"FlowName":     true,

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

	// Platform Variables
	"PlatformOS":            true,
	"PlatformArch":          true,
	"PlatformInDocker":      true,
	"PlatformInKubernetes":  true,
	"PlatformCloudProvider": true,

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
	"run_uuid": true,
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

func (r *UndefinedVariableRule) Name() string { return "undefined-variable" }
func (r *UndefinedVariableRule) Description() string {
	return "Detects variables referenced but not defined"
}
func (r *UndefinedVariableRule) Severity() Severity { return SeverityWarning }

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

	// Collect variables provided by trigger inputs
	triggerVars := make(map[string]bool)
	hasEventTrigger := false
	for _, trigger := range w.Triggers {
		if trigger.On == core.TriggerEvent {
			hasEventTrigger = true
		}
		if trigger.Input.HasVars() {
			for varName := range trigger.Input.Vars {
				triggerVars[varName] = true
			}
		}
		// Legacy syntax
		if trigger.Input.Name != "" {
			triggerVars[trigger.Input.Name] = true
		}
	}

	// If workflow has event triggers, add event envelope variables as defined
	if hasEventTrigger {
		for _, v := range []string{"EventEnvelope", "EventTopic", "EventSource", "EventDataType", "EventTimestamp", "EventData"} {
			defined[v] = true
		}
	}

	// Check each step for undefined variables
	for i, step := range w.Steps {
		stepPrefix := fmt.Sprintf("steps[%d]", i)

		// Check all fields of this step
		checkStepFieldsForUndefinedVars(&step, stepPrefix, defined, triggerVars, hasEventTrigger, wast, r, &issues)

		// Recurse into foreach inner step with scoped defined copy
		if step.Step != nil {
			innerDefined := copyDefinedMap(defined)
			if step.Variable != "" {
				innerDefined[step.Variable] = true
			}
			// _id_ is always available in foreach inner steps
			innerDefined["_id_"] = true
			checkStepFieldsForUndefinedVars(step.Step, stepPrefix+".step", innerDefined, triggerVars, hasEventTrigger, wast, r, &issues)
		}

		// Recurse into parallel_steps
		for j := range step.ParallelSteps {
			checkStepFieldsForUndefinedVars(&step.ParallelSteps[j], fmt.Sprintf("%s.parallel_steps[%d]", stepPrefix, j), defined, triggerVars, hasEventTrigger, wast, r, &issues)
		}

		// After processing this step, add its exports to defined
		for exportName := range step.Exports {
			defined[exportName] = true
		}

		// Add foreach variable to defined for subsequent steps
		if step.Variable != "" {
			defined[step.Variable] = true
		}
	}

	return issues
}

// CircularDependencyRule checks for circular step dependencies
type CircularDependencyRule struct{}

func (r *CircularDependencyRule) Name() string { return "circular-dependency" }
func (r *CircularDependencyRule) Description() string {
	return "Detects circular references in step dependencies"
}
func (r *CircularDependencyRule) Severity() Severity { return SeverityWarning }

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
		case core.StepTypeAgent:
			if (step.Query == "" && len(step.Queries) == 0) || len(step.AgentTools) == 0 {
				empty = true
			}
			// Agent-specific warnings
			if step.Query != "" && len(step.Queries) > 0 {
				line, col := wast.FindStepFieldPosition(i, "queries")
				issues = append(issues, LintIssue{
					Rule:       r.Name(),
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("Step '%s' has both 'query' and 'queries' — only one should be used", step.Name),
					Suggestion: "Remove either 'query' or 'queries'",
					Line:       line,
					Column:     col,
					Field:      fmt.Sprintf("steps[%d].queries", i),
				})
			}
			if step.PlanPrompt != "" && step.Query == "" && len(step.Queries) == 0 {
				line, col := wast.FindStepFieldPosition(i, "plan_prompt")
				issues = append(issues, LintIssue{
					Rule:       r.Name(),
					Severity:   SeverityWarning,
					Message:    fmt.Sprintf("Step '%s' has plan_prompt but no query/queries", step.Name),
					Suggestion: "Add a 'query' or 'queries' field for the agent to execute after planning",
					Line:       line,
					Column:     col,
					Field:      fmt.Sprintf("steps[%d].plan_prompt", i),
				})
			}
			if step.MaxIterations > 50 {
				line, col := wast.FindStepFieldPosition(i, "max_iterations")
				issues = append(issues, LintIssue{
					Rule:       r.Name(),
					Severity:   SeverityInfo,
					Message:    fmt.Sprintf("Step '%s' has max_iterations=%d which is suspiciously high", step.Name, step.MaxIterations),
					Suggestion: "Consider lowering max_iterations to avoid excessive LLM calls",
					Line:       line,
					Column:     col,
					Field:      fmt.Sprintf("steps[%d].max_iterations", i),
				})
			}
			for j, tool := range step.AgentTools {
				if tool.IsPreset() {
					if _, ok := core.GetPresetTool(tool.Preset); !ok {
						line, col := wast.FindStepFieldPosition(i, "agent_tools")
						issues = append(issues, LintIssue{
							Rule:       r.Name(),
							Severity:   SeverityWarning,
							Message:    fmt.Sprintf("Step '%s' references unknown preset tool '%s'", step.Name, tool.Preset),
							Suggestion: "Check the preset tool name against the preset tool registry",
							Line:       line,
							Column:     col,
							Field:      fmt.Sprintf("steps[%d].agent_tools[%d].preset", i, j),
						})
					}
				} else if tool.Handler != "" && tool.Description == "" {
					line, col := wast.FindStepFieldPosition(i, "agent_tools")
					issues = append(issues, LintIssue{
						Rule:       r.Name(),
						Severity:   SeverityWarning,
						Message:    fmt.Sprintf("Step '%s' has custom tool '%s' with handler but no description", step.Name, tool.Name),
						Suggestion: "Add a description so the LLM knows when to use this tool",
						Line:       line,
						Column:     col,
						Field:      fmt.Sprintf("steps[%d].agent_tools[%d].description", i, j),
					})
				}
			}
			// Sub-agent validation
			issues = append(issues, validateSubAgents(step, i, wast, r)...)
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

func (r *InvalidGotoRule) Name() string { return "invalid-goto" }
func (r *InvalidGotoRule) Description() string {
	return "Detects decision goto references to non-existent steps"
}
func (r *InvalidGotoRule) Severity() Severity { return SeverityWarning }

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

func (r *DuplicateStepNameRule) Name() string { return "duplicate-step-name" }
func (r *DuplicateStepNameRule) Description() string {
	return "Detects multiple steps with the same name"
}
func (r *DuplicateStepNameRule) Severity() Severity { return SeverityWarning }

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

func (r *MissingRequiredFieldRule) Name() string { return "missing-required-field" }
func (r *MissingRequiredFieldRule) Description() string {
	return "Detects required fields that are missing"
}
func (r *MissingRequiredFieldRule) Severity() Severity { return SeverityWarning }

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

func (r *InvalidDependsOnRule) Name() string { return "invalid-depends-on" }
func (r *InvalidDependsOnRule) Description() string {
	return "Detects depends_on references to non-existent steps"
}
func (r *InvalidDependsOnRule) Severity() Severity { return SeverityWarning }

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

// copyDefinedMap creates a shallow copy of a map[string]bool
func copyDefinedMap(m map[string]bool) map[string]bool {
	cp := make(map[string]bool, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

// checkStepFieldsForUndefinedVars checks ALL template-renderable fields of a single step.
// stepPrefix is the field path prefix (e.g. "steps[0]" or "steps[0].step").
func checkStepFieldsForUndefinedVars(step *core.Step, stepPrefix string, defined map[string]bool, triggerVars map[string]bool, hasEventTrigger bool, wast *WorkflowAST, rule *UndefinedVariableRule, issues *[]LintIssue) {
	check := func(s, field string) {
		checkStringForUndefinedVars(s, field, defined, triggerVars, hasEventTrigger, wast, rule, issues)
	}
	checkSlice := func(slice []string, fieldBase string) {
		for j, s := range slice {
			check(s, fmt.Sprintf("%s[%d]", fieldBase, j))
		}
	}

	// Bash fields
	check(step.Command, stepPrefix+".command")
	checkSlice(step.Commands, stepPrefix+".commands")
	checkSlice(step.ParallelCommands, stepPrefix+".parallel_commands")

	// Structured args
	check(step.SpeedArgs, stepPrefix+".speed_args")
	check(step.ConfigArgs, stepPrefix+".config_args")
	check(step.InputArgs, stepPrefix+".input_args")
	check(step.OutputArgs, stepPrefix+".output_args")

	// Function fields
	check(step.Function, stepPrefix+".function")
	checkSlice(step.Functions, stepPrefix+".functions")
	checkSlice(step.ParallelFunctions, stepPrefix+".parallel_functions")

	// Common fields
	check(step.PreCondition, stepPrefix+".pre_condition")
	check(step.Log, stepPrefix+".log")
	check(step.Input, stepPrefix+".input")
	check(step.VariablePreProcess, stepPrefix+".variable_pre_process")

	// File paths
	check(step.StdFile, stepPrefix+".std_file")
	check(step.StepRemoteFile, stepPrefix+".step_remote_file")
	check(step.HostOutputFile, stepPrefix+".host_output_file")

	// HTTP fields
	check(step.URL, stepPrefix+".url")
	check(step.RequestBody, stepPrefix+".request_body")
	for hk, hv := range step.Headers {
		check(hv, fmt.Sprintf("%s.headers.%s", stepPrefix, hk))
	}

	// LLM fields
	for j, msg := range step.Messages {
		if s, ok := msg.Content.(string); ok {
			check(s, fmt.Sprintf("%s.messages[%d].content", stepPrefix, j))
		}
	}
	checkSlice(step.EmbeddingInput, stepPrefix+".embedding_input")

	// Agent fields
	check(step.Query, stepPrefix+".query")
	checkSlice(step.Queries, stepPrefix+".queries")
	check(step.SystemPrompt, stepPrefix+".system_prompt")
	check(step.StopCondition, stepPrefix+".stop_condition")
	check(step.PlanPrompt, stepPrefix+".plan_prompt")
	check(step.OnToolStart, stepPrefix+".on_tool_start")
	check(step.OnToolEnd, stepPrefix+".on_tool_end")

	// Agent memory paths
	if step.Memory != nil {
		check(step.Memory.PersistPath, stepPrefix+".memory.persist_path")
		check(step.Memory.ResumePath, stepPrefix+".memory.resume_path")
	}

	// Export values
	for exportName, exportValue := range step.Exports {
		check(exportValue, fmt.Sprintf("%s.exports.%s", stepPrefix, exportName))
	}

	// Decision fields
	if step.Decision != nil {
		check(step.Decision.Switch, stepPrefix+".decision.switch")

		for caseVal, dc := range step.Decision.Cases {
			casePrefix := fmt.Sprintf("%s.decision.cases.%s", stepPrefix, caseVal)
			check(dc.Command, casePrefix+".command")
			checkSlice(dc.Commands, casePrefix+".commands")
			check(dc.Function, casePrefix+".function")
			checkSlice(dc.Functions, casePrefix+".functions")
		}

		if step.Decision.Default != nil {
			defPrefix := stepPrefix + ".decision.default"
			check(step.Decision.Default.Command, defPrefix+".command")
			checkSlice(step.Decision.Default.Commands, defPrefix+".commands")
			check(step.Decision.Default.Function, defPrefix+".function")
			checkSlice(step.Decision.Default.Functions, defPrefix+".functions")
		}

		for j, cond := range step.Decision.Conditions {
			condPrefix := fmt.Sprintf("%s.decision.conditions[%d]", stepPrefix, j)
			check(cond.If, condPrefix+".if")
			check(cond.Command, condPrefix+".command")
			checkSlice(cond.Commands, condPrefix+".commands")
			check(cond.Function, condPrefix+".function")
			checkSlice(cond.Functions, condPrefix+".functions")
		}
	}
}

func checkStringForUndefinedVars(s, field string, defined map[string]bool, triggerVars map[string]bool, hasEventTrigger bool, wast *WorkflowAST, rule *UndefinedVariableRule, issues *[]LintIssue) {
	if s == "" {
		return
	}

	for _, v := range extractVariables(s) {
		if !defined[v] && !builtInVariables[v] {
			if triggerVars[v] {
				line, col := wast.GetNodePosition(field)
				*issues = append(*issues, LintIssue{
					Rule:     rule.Name(),
					Severity: SeverityInfo,
					Message:  fmt.Sprintf("Variable '%s' is provided by trigger input (not statically defined)", v),
					Line:     line,
					Column:   col,
					Field:    field,
				})
			} else if hasEventTrigger {
				line, col := wast.GetNodePosition(field)
				*issues = append(*issues, LintIssue{
					Rule:     rule.Name(),
					Severity: SeverityInfo,
					Message:  fmt.Sprintf("Variable '%s' may be provided by event data at runtime", v),
					Line:     line,
					Column:   col,
					Field:    field,
				})
			} else {
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
}

// collectReferencedVarsFromStep extracts all variable references from a step's fields.
func collectReferencedVarsFromStep(step *core.Step, referenced map[string]bool) {
	addVars := func(s string) {
		for _, v := range extractVariables(s) {
			referenced[v] = true
		}
	}
	addSliceVars := func(slice []string) {
		for _, s := range slice {
			addVars(s)
		}
	}

	// Bash fields
	addVars(step.Command)
	addSliceVars(step.Commands)
	addSliceVars(step.ParallelCommands)

	// Structured args
	addVars(step.SpeedArgs)
	addVars(step.ConfigArgs)
	addVars(step.InputArgs)
	addVars(step.OutputArgs)

	// Function fields
	addVars(step.Function)
	addSliceVars(step.Functions)
	addSliceVars(step.ParallelFunctions)

	// Common fields
	addVars(step.PreCondition)
	addVars(step.Log)
	addVars(step.Input)
	addVars(step.VariablePreProcess)

	// File paths
	addVars(step.StdFile)
	addVars(step.StepRemoteFile)
	addVars(step.HostOutputFile)

	// HTTP fields
	addVars(step.URL)
	addVars(step.RequestBody)
	for _, hv := range step.Headers {
		addVars(hv)
	}

	// LLM fields
	for _, msg := range step.Messages {
		if s, ok := msg.Content.(string); ok {
			addVars(s)
		}
	}
	addSliceVars(step.EmbeddingInput)

	// Agent fields
	addVars(step.Query)
	addSliceVars(step.Queries)
	addVars(step.SystemPrompt)
	addVars(step.StopCondition)
	addVars(step.PlanPrompt)
	addVars(step.OnToolStart)
	addVars(step.OnToolEnd)

	// Agent memory paths
	if step.Memory != nil {
		addVars(step.Memory.PersistPath)
		addVars(step.Memory.ResumePath)
	}

	// Export values
	for _, exportValue := range step.Exports {
		addVars(exportValue)
	}

	// Decision fields
	if step.Decision != nil {
		addVars(step.Decision.Switch)

		for _, dc := range step.Decision.Cases {
			addVars(dc.Command)
			addSliceVars(dc.Commands)
			addVars(dc.Function)
			addSliceVars(dc.Functions)
		}

		if step.Decision.Default != nil {
			addVars(step.Decision.Default.Command)
			addSliceVars(step.Decision.Default.Commands)
			addVars(step.Decision.Default.Function)
			addSliceVars(step.Decision.Default.Functions)
		}

		for _, cond := range step.Decision.Conditions {
			addVars(cond.If)
			addVars(cond.Command)
			addSliceVars(cond.Commands)
			addVars(cond.Function)
			addSliceVars(cond.Functions)
		}
	}

	// Recurse into foreach inner step
	if step.Step != nil {
		collectReferencedVarsFromStep(step.Step, referenced)
	}

	// Recurse into parallel_steps
	for i := range step.ParallelSteps {
		collectReferencedVarsFromStep(&step.ParallelSteps[i], referenced)
	}
}

func collectReferencedVars(step *core.Step, _ int, referenced map[string]bool, _ []core.Step) {
	collectReferencedVarsFromStep(step, referenced)
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

// validateSubAgents checks sub-agent definitions for common issues
func validateSubAgents(step core.Step, stepIdx int, wast *WorkflowAST, r *EmptyStepRule) []LintIssue {
	var issues []LintIssue
	if len(step.SubAgents) == 0 {
		return issues
	}

	seenNames := make(map[string]bool)
	for j, sa := range step.SubAgents {
		// Error: sub-agent missing name
		if sa.Name == "" {
			line, col := wast.FindStepFieldPosition(stepIdx, "sub_agents")
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("Step '%s' sub-agent at index %d is missing required 'name' field", step.Name, j),
				Suggestion: "Add a 'name' field to the sub-agent definition",
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d].sub_agents[%d].name", stepIdx, j),
			})
		}

		// Warning: sub-agent missing description
		if sa.Name != "" && sa.Description == "" {
			line, col := wast.FindStepFieldPosition(stepIdx, "sub_agents")
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   SeverityInfo,
				Message:    fmt.Sprintf("Step '%s' sub-agent '%s' has no description", step.Name, sa.Name),
				Suggestion: "Add a description so the LLM knows when to delegate to this agent",
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d].sub_agents[%d].description", stepIdx, j),
			})
		}

		// Error: duplicate sub-agent names
		if sa.Name != "" {
			if seenNames[sa.Name] {
				line, col := wast.FindStepFieldPosition(stepIdx, "sub_agents")
				issues = append(issues, LintIssue{
					Rule:       r.Name(),
					Severity:   SeverityError,
					Message:    fmt.Sprintf("Step '%s' has duplicate sub-agent name '%s'", step.Name, sa.Name),
					Suggestion: "Use unique names for each sub-agent",
					Line:       line,
					Column:     col,
					Field:      fmt.Sprintf("steps[%d].sub_agents[%d].name", stepIdx, j),
				})
			}
			seenNames[sa.Name] = true
		}

		// Warning: sub-agent with no agent_tools
		if len(sa.AgentTools) == 0 {
			line, col := wast.FindStepFieldPosition(stepIdx, "sub_agents")
			issues = append(issues, LintIssue{
				Rule:       r.Name(),
				Severity:   SeverityInfo,
				Message:    fmt.Sprintf("Step '%s' sub-agent '%s' has no agent_tools", step.Name, sa.Name),
				Suggestion: "Add agent_tools so the sub-agent can perform actions",
				Line:       line,
				Column:     col,
				Field:      fmt.Sprintf("steps[%d].sub_agents[%d].agent_tools", stepIdx, j),
			})
		}

		// Warning: unknown preset tool in sub-agent
		for k, tool := range sa.AgentTools {
			if tool.IsPreset() {
				if _, ok := core.GetPresetTool(tool.Preset); !ok {
					line, col := wast.FindStepFieldPosition(stepIdx, "sub_agents")
					issues = append(issues, LintIssue{
						Rule:       r.Name(),
						Severity:   SeverityWarning,
						Message:    fmt.Sprintf("Step '%s' sub-agent '%s' references unknown preset tool '%s'", step.Name, sa.Name, tool.Preset),
						Suggestion: "Check the preset tool name against the preset tool registry",
						Line:       line,
						Column:     col,
						Field:      fmt.Sprintf("steps[%d].sub_agents[%d].agent_tools[%d].preset", stepIdx, j, k),
					})
				}
			}
		}
	}

	// Info: deep nesting warning
	maxDepth := countSubAgentDepth(step.SubAgents)
	if maxDepth > core.DefaultMaxAgentDepth {
		line, col := wast.FindStepFieldPosition(stepIdx, "sub_agents")
		issues = append(issues, LintIssue{
			Rule:       r.Name(),
			Severity:   SeverityInfo,
			Message:    fmt.Sprintf("Step '%s' has sub-agent nesting depth of %d (default limit is %d)", step.Name, maxDepth, core.DefaultMaxAgentDepth),
			Suggestion: "Consider increasing max_agent_depth or reducing nesting",
			Line:       line,
			Column:     col,
			Field:      fmt.Sprintf("steps[%d].sub_agents", stepIdx),
		})
	}

	return issues
}

// countSubAgentDepth returns the maximum nesting depth of sub-agents
func countSubAgentDepth(subAgents []core.SubAgentDef) int {
	if len(subAgents) == 0 {
		return 0
	}
	maxChild := 0
	for _, sa := range subAgents {
		childDepth := countSubAgentDepth(sa.SubAgents)
		if childDepth > maxChild {
			maxChild = childDepth
		}
	}
	return 1 + maxChild
}

// GetDefaultRules returns all built-in linting rules
func GetDefaultRules() []LinterRule {
	return []LinterRule{
		&MissingRequiredFieldRule{},
		&DuplicateStepNameRule{},
		&EmptyStepRule{},
		&UndefinedVariableRule{},
		&UnusedVariableRule{},
		&InvalidGotoRule{},
		&InvalidDependsOnRule{},
		&CircularDependencyRule{},
	}
}
