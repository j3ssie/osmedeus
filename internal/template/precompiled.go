package template

import (
	"strings"
	"sync"

	"github.com/flosch/pongo2/v6"
)

// WorkflowTemplates holds pre-compiled templates for a single workflow
type WorkflowTemplates struct {
	// Templates maps field keys to pre-compiled templates
	// Key format: "stepName:fieldName" (e.g., "scan:command", "scan:exports:output")
	Templates map[string]*pongo2.Template
}

// PrecompiledRegistry stores pre-compiled templates for workflows.
// Templates are compiled once at workflow load time and reused during execution.
type PrecompiledRegistry struct {
	mu        sync.RWMutex
	workflows map[string]*WorkflowTemplates
}

// NewPrecompiledRegistry creates a new pre-compiled template registry
func NewPrecompiledRegistry() *PrecompiledRegistry {
	return &PrecompiledRegistry{
		workflows: make(map[string]*WorkflowTemplates),
	}
}

// PrecompileWorkflow pre-compiles all template strings for a workflow.
// The templates map should use keys like "stepName:fieldName".
func (r *PrecompiledRegistry) PrecompileWorkflow(workflowName string, templates map[string]string) error {
	compiled := &WorkflowTemplates{
		Templates: make(map[string]*pongo2.Template, len(templates)),
	}

	for key, tmplStr := range templates {
		// Skip if no template variables
		if !strings.Contains(tmplStr, "{{") {
			continue
		}

		tpl, err := pongo2.FromString(tmplStr)
		if err != nil {
			// Log warning but continue - invalid templates will be handled at runtime
			continue
		}
		compiled.Templates[key] = tpl
	}

	r.mu.Lock()
	r.workflows[workflowName] = compiled
	r.mu.Unlock()

	return nil
}

// GetPrecompiled retrieves a pre-compiled template for a workflow field.
// Returns nil if not found.
func (r *PrecompiledRegistry) GetPrecompiled(workflowName, key string) any {
	r.mu.RLock()
	wf, ok := r.workflows[workflowName]
	r.mu.RUnlock()

	if !ok {
		return nil
	}

	// Note: the WorkflowTemplates map doesn't need locking for reads
	// as it's never modified after creation
	tpl, exists := wf.Templates[key]
	if !exists {
		return nil
	}
	return tpl
}

// ClearPrecompiled removes pre-compiled templates for a workflow
func (r *PrecompiledRegistry) ClearPrecompiled(workflowName string) {
	r.mu.Lock()
	delete(r.workflows, workflowName)
	r.mu.Unlock()
}

// ClearAll removes all pre-compiled templates
func (r *PrecompiledRegistry) ClearAll() {
	r.mu.Lock()
	r.workflows = make(map[string]*WorkflowTemplates)
	r.mu.Unlock()
}

// GetWorkflowCount returns the number of workflows with pre-compiled templates
func (r *PrecompiledRegistry) GetWorkflowCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.workflows)
}

// GetTemplateCount returns the total number of pre-compiled templates
func (r *PrecompiledRegistry) GetTemplateCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, wf := range r.workflows {
		count += len(wf.Templates)
	}
	return count
}

// Verify interface compliance at compile time
var _ Precompiler = (*PrecompiledRegistry)(nil)
