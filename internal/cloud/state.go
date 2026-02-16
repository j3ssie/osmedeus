package cloud

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/j3ssie/osmedeus/v5/internal/json"
)

// SaveInfrastructureState persists infrastructure state to disk
func SaveInfrastructureState(infra *Infrastructure, statePath string) error {
	// Ensure state directory exists
	stateDir := filepath.Join(statePath, "infrastructure")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Create state file path
	stateFile := filepath.Join(stateDir, fmt.Sprintf("%s.json", infra.ID))

	// Marshal infrastructure to JSON
	data, err := json.Marshal(infra)
	if err != nil {
		return fmt.Errorf("failed to marshal infrastructure state: %w", err)
	}

	// Write state file
	if err := os.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadInfrastructureState loads infrastructure state from disk
func LoadInfrastructureState(infraID, statePath string) (*Infrastructure, error) {
	stateFile := filepath.Join(statePath, "infrastructure", fmt.Sprintf("%s.json", infraID))

	// Read state file
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal infrastructure
	var infra Infrastructure
	if err := json.Unmarshal(data, &infra); err != nil {
		return nil, fmt.Errorf("failed to unmarshal infrastructure state: %w", err)
	}

	return &infra, nil
}

// ListInfrastructures lists all saved infrastructure states
func ListInfrastructures(statePath string) ([]*Infrastructure, error) {
	stateDir := filepath.Join(statePath, "infrastructure")

	// Check if directory exists
	if _, err := os.Stat(stateDir); os.IsNotExist(err) {
		return []*Infrastructure{}, nil
	}

	// Read directory entries
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	// Load each infrastructure state
	var infrastructures []*Infrastructure
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		infraID := entry.Name()[:len(entry.Name())-5] // Remove .json extension
		infra, err := LoadInfrastructureState(infraID, statePath)
		if err != nil {
			// Log warning but continue
			fmt.Printf("Warning: failed to load infrastructure %s: %v\n", infraID, err)
			continue
		}

		infrastructures = append(infrastructures, infra)
	}

	return infrastructures, nil
}

// RemoveInfrastructureState removes infrastructure state from disk
func RemoveInfrastructureState(infraID, statePath string) error {
	stateFile := filepath.Join(statePath, "infrastructure", fmt.Sprintf("%s.json", infraID))

	if err := os.Remove(stateFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove state file: %w", err)
	}

	return nil
}

// InfrastructureExists checks if an infrastructure state exists
func InfrastructureExists(infraID, statePath string) bool {
	stateFile := filepath.Join(statePath, "infrastructure", fmt.Sprintf("%s.json", infraID))
	_, err := os.Stat(stateFile)
	return err == nil
}
