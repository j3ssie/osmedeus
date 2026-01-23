package executor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// ExportWorkflowState writes the workflow YAML to the state file
func ExportWorkflowState(stateFile string, workflow *core.Workflow) error {
	if stateFile == "" {
		return fmt.Errorf("state file path is empty")
	}

	// Ensure directory exists
	dir := filepath.Dir(stateFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal workflow to YAML
	data, err := yaml.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	// Write to file
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow state file: %w", err)
	}

	return nil
}

// ExportModuleWorkflowState writes a module workflow YAML to the modules folder
func ExportModuleWorkflowState(folder string, moduleName string, workflow *core.Workflow) error {
	if folder == "" || moduleName == "" {
		return fmt.Errorf("folder or module name is empty")
	}

	// Ensure directory exists
	if err := os.MkdirAll(folder, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Build filename: run-{module-name}.yaml
	filename := fmt.Sprintf("run-%s.yaml", moduleName)
	filePath := filepath.Join(folder, filename)

	// Marshal workflow to YAML
	data, err := yaml.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write module workflow state file: %w", err)
	}

	return nil
}
