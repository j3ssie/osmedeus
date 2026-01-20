package template

import (
	"testing"
)

func TestPrecompiledRegistry_PrecompileWorkflow(t *testing.T) {
	registry := NewPrecompiledRegistry()

	templates := map[string]string{
		"step1:command": "nmap -sV {{target}}",
		"step1:output":  "{{output}}/nmap.txt",
		"step2:command": "nuclei -t {{templates}} -u {{target}}",
		"static":        "No variables here",
	}

	err := registry.PrecompileWorkflow("test-workflow", templates)
	if err != nil {
		t.Fatalf("PrecompileWorkflow() error = %v", err)
	}

	// Verify templates were compiled
	if registry.GetWorkflowCount() != 1 {
		t.Errorf("GetWorkflowCount() = %d, want 1", registry.GetWorkflowCount())
	}

	// Should have 3 compiled templates (static one is skipped)
	if registry.GetTemplateCount() != 3 {
		t.Errorf("GetTemplateCount() = %d, want 3", registry.GetTemplateCount())
	}
}

func TestPrecompiledRegistry_GetPrecompiled(t *testing.T) {
	registry := NewPrecompiledRegistry()

	templates := map[string]string{
		"step1:command": "nmap -sV {{target}}",
	}

	_ = registry.PrecompileWorkflow("test-workflow", templates)

	// Test existing template
	tpl := registry.GetPrecompiled("test-workflow", "step1:command")
	if tpl == nil {
		t.Error("GetPrecompiled() returned nil for existing template")
	}

	// Test non-existing template
	tpl = registry.GetPrecompiled("test-workflow", "nonexistent")
	if tpl != nil {
		t.Error("GetPrecompiled() returned non-nil for nonexistent template")
	}

	// Test non-existing workflow
	tpl = registry.GetPrecompiled("nonexistent-workflow", "step1:command")
	if tpl != nil {
		t.Error("GetPrecompiled() returned non-nil for nonexistent workflow")
	}
}

func TestPrecompiledRegistry_ClearPrecompiled(t *testing.T) {
	registry := NewPrecompiledRegistry()

	templates := map[string]string{
		"step1:command": "nmap -sV {{target}}",
	}

	_ = registry.PrecompileWorkflow("test-workflow", templates)

	if registry.GetWorkflowCount() != 1 {
		t.Fatalf("Setup failed: GetWorkflowCount() = %d", registry.GetWorkflowCount())
	}

	registry.ClearPrecompiled("test-workflow")

	if registry.GetWorkflowCount() != 0 {
		t.Errorf("After ClearPrecompiled: GetWorkflowCount() = %d, want 0", registry.GetWorkflowCount())
	}
}

func TestPrecompiledRegistry_ClearAll(t *testing.T) {
	registry := NewPrecompiledRegistry()

	_ = registry.PrecompileWorkflow("workflow1", map[string]string{"key": "{{value}}"})
	_ = registry.PrecompileWorkflow("workflow2", map[string]string{"key": "{{value}}"})

	if registry.GetWorkflowCount() != 2 {
		t.Fatalf("Setup failed: GetWorkflowCount() = %d", registry.GetWorkflowCount())
	}

	registry.ClearAll()

	if registry.GetWorkflowCount() != 0 {
		t.Errorf("After ClearAll: GetWorkflowCount() = %d, want 0", registry.GetWorkflowCount())
	}
}

func TestPrecompiledRegistry_InvalidTemplate(t *testing.T) {
	registry := NewPrecompiledRegistry()

	// Template with invalid syntax - should be skipped, not error
	templates := map[string]string{
		"valid":   "{{target}}",
		"invalid": "{{invalid syntax",
	}

	err := registry.PrecompileWorkflow("test-workflow", templates)
	if err != nil {
		t.Errorf("PrecompileWorkflow() should not error on invalid templates: %v", err)
	}

	// Only valid template should be compiled
	if registry.GetTemplateCount() != 1 {
		t.Errorf("GetTemplateCount() = %d, want 1 (invalid should be skipped)", registry.GetTemplateCount())
	}
}

func TestPrecompiledRegistry_EmptyTemplates(t *testing.T) {
	registry := NewPrecompiledRegistry()

	err := registry.PrecompileWorkflow("empty-workflow", map[string]string{})
	if err != nil {
		t.Errorf("PrecompileWorkflow() with empty templates error = %v", err)
	}

	if registry.GetWorkflowCount() != 1 {
		t.Errorf("GetWorkflowCount() = %d, want 1", registry.GetWorkflowCount())
	}
	if registry.GetTemplateCount() != 0 {
		t.Errorf("GetTemplateCount() = %d, want 0", registry.GetTemplateCount())
	}
}
