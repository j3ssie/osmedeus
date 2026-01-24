package executor

import (
	"runtime"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/template"
)

func TestDetectDocker(t *testing.T) {
	result := DetectDocker()
	// On a non-Docker environment, this should be false
	t.Logf("DetectDocker() = %v", result)
	// This is primarily a smoke test - we just verify it doesn't panic
}

func TestDetectKubernetes(t *testing.T) {
	result := DetectKubernetes()
	// On a non-Kubernetes environment, this should be false
	t.Logf("DetectKubernetes() = %v", result)
	// This is primarily a smoke test - we just verify it doesn't panic
}

func TestDetectCloudProvider(t *testing.T) {
	result := DetectCloudProvider()
	t.Logf("DetectCloudProvider() = %s", result)
	// On a local machine, this should be "local"
	// This is primarily a smoke test - we just verify it doesn't panic
}

func TestPlatformVariables(t *testing.T) {
	// Verify that runtime.GOOS and runtime.GOARCH return expected values
	t.Logf("runtime.GOOS = %s", runtime.GOOS)
	t.Logf("runtime.GOARCH = %s", runtime.GOARCH)

	// Verify they are not empty
	if runtime.GOOS == "" {
		t.Error("runtime.GOOS is empty")
	}
	if runtime.GOARCH == "" {
		t.Error("runtime.GOARCH is empty")
	}
}

func TestPlatformVariablesInjection(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		BaseFolder:     "/tmp",
		BinariesPath:   "/tmp/bin",
		DataPath:       "/tmp/data",
		WorkspacesPath: "/tmp/workspaces",
	}

	// Create execution context
	execCtx := core.NewExecutionContext("test-workflow", core.KindModule, "test-run-uuid", "example.com")

	// Create executor and inject variables
	e := NewExecutor()
	params := map[string]string{
		"target": "example.com",
	}
	e.injectBuiltinVariables(cfg, params, execCtx)

	// Verify platform variables are set
	tests := []struct {
		name     string
		expected interface{}
	}{
		{"PlatformOS", runtime.GOOS},
		{"PlatformArch", runtime.GOARCH},
		{"PlatformInDocker", DetectDocker()},
		{"PlatformInKubernetes", DetectKubernetes()},
		{"PlatformCloudProvider", DetectCloudProvider()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := execCtx.GetVariable(tt.name)
			if !ok {
				t.Errorf("variable %s not found in execution context", tt.name)
				return
			}
			if val != tt.expected {
				t.Errorf("variable %s = %v, expected %v", tt.name, val, tt.expected)
			}
			t.Logf("%s = %v", tt.name, val)
		})
	}

	// Verify variables are in GetVariables() output
	allVars := execCtx.GetVariables()
	for _, tt := range tests {
		if _, ok := allVars[tt.name]; !ok {
			t.Errorf("variable %s not found in GetVariables() output", tt.name)
		}
	}
}

func TestPlatformVariablesTemplateRendering(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		BaseFolder:     "/tmp",
		BinariesPath:   "/tmp/bin",
		DataPath:       "/tmp/data",
		WorkspacesPath: "/tmp/workspaces",
	}

	// Create execution context
	execCtx := core.NewExecutionContext("test-workflow", core.KindModule, "test-run-uuid", "example.com")

	// Create executor and inject variables
	e := NewExecutor()
	params := map[string]string{
		"target": "example.com",
	}
	e.injectBuiltinVariables(cfg, params, execCtx)

	// Create template engine and test rendering
	engine := template.NewEngine()
	vars := execCtx.GetVariables()

	// Debug: print all variables
	t.Logf("Total variables: %d", len(vars))
	for k, v := range vars {
		if k == "PlatformOS" || k == "PlatformArch" || k == "PlatformInDocker" || k == "PlatformInKubernetes" || k == "PlatformCloudProvider" {
			t.Logf("Variable %s = %v (type: %T)", k, v, v)
		}
	}

	tests := []struct {
		template string
		expected string
	}{
		{"echo {{PlatformOS}}", "echo " + runtime.GOOS},
		{"echo {{PlatformArch}}", "echo " + runtime.GOARCH},
		{"echo {{PlatformInDocker}}", "echo false"},
		{"echo {{PlatformInKubernetes}}", "echo false"},
		{"echo {{PlatformCloudProvider}}", "echo local"},
	}

	for _, tt := range tests {
		t.Run(tt.template, func(t *testing.T) {
			result, err := engine.Render(tt.template, vars)
			if err != nil {
				t.Errorf("error rendering template %q: %v", tt.template, err)
				return
			}
			if result != tt.expected {
				t.Errorf("render(%q) = %q, expected %q", tt.template, result, tt.expected)
			}
			t.Logf("render(%q) = %q", tt.template, result)
		})
	}
}

func TestPlatformVariablesWithStepDispatcher(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		BaseFolder:     "/tmp",
		BinariesPath:   "/tmp/bin",
		DataPath:       "/tmp/data",
		WorkspacesPath: "/tmp/workspaces",
	}

	// Create execution context
	execCtx := core.NewExecutionContext("test-workflow", core.KindModule, "test-run-uuid", "example.com")

	// Create executor and inject variables
	e := NewExecutor()
	params := map[string]string{
		"target": "example.com",
	}
	e.injectBuiltinVariables(cfg, params, execCtx)

	// Get the step dispatcher
	dispatcher := NewStepDispatcher()

	// Create a test step
	step := &core.Step{
		Name:    "test-platform-step",
		Type:    core.StepTypeBash,
		Command: "echo {{PlatformOS}} {{PlatformArch}}",
	}

	// Get the template engine from dispatcher and render
	vars := execCtx.GetVariables()
	engine := dispatcher.GetTemplateEngine()

	result, err := engine.Render(step.Command, vars)
	if err != nil {
		t.Fatalf("Failed to render command: %v", err)
	}

	expected := "echo " + runtime.GOOS + " " + runtime.GOARCH
	if result != expected {
		t.Errorf("Rendered command = %q, expected %q", result, expected)
	}
	t.Logf("Rendered command: %q", result)
}
