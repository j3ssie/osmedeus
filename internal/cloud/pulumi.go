package cloud

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// PulumiManager manages Pulumi stacks for infrastructure provisioning
type PulumiManager struct {
	projectName string
	stackName   string
	workspace   auto.Workspace
	stack       auto.Stack
}

// NewPulumiManager creates a new Pulumi manager with local backend
func NewPulumiManager(projectName, stackName, statePath string) (*PulumiManager, error) {
	// Ensure Pulumi is installed
	if err := ensurePulumiInstalled(); err != nil {
		return nil, err
	}

	// Ensure state directory exists
	if err := os.MkdirAll(statePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	ctx := context.Background()

	// Create workspace with local backend
	ws, err := auto.NewLocalWorkspace(ctx,
		auto.WorkDir(statePath),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulumi workspace: %w", err)
	}

	// Initialize stack with inline program
	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, func(ctx *pulumi.Context) error {
		// Placeholder program - will be replaced by provider-specific logic
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create stack: %w", err)
	}

	// Set workspace for stack
	stack.Workspace().SetProgram(func(ctx *pulumi.Context) error {
		return nil
	})

	return &PulumiManager{
		projectName: projectName,
		stackName:   stackName,
		workspace:   ws,
		stack:       stack,
	}, nil
}

// Up provisions infrastructure using the provided Pulumi program
func (pm *PulumiManager) Up(ctx context.Context, program pulumi.RunFunc) error {
	// Set the program for the stack
	pm.stack.Workspace().SetProgram(program)

	// Set stack configuration if needed
	// This can be extended to set provider-specific config

	// Run pulumi up with progress streaming
	_, err := pm.stack.Up(ctx, optup.ProgressStreams(os.Stdout))
	if err != nil {
		return fmt.Errorf("failed to provision infrastructure: %w", err)
	}

	return nil
}

// Destroy tears down the infrastructure
func (pm *PulumiManager) Destroy(ctx context.Context) error {
	_, err := pm.stack.Destroy(ctx)
	if err != nil {
		return fmt.Errorf("failed to destroy infrastructure: %w", err)
	}

	// Remove stack after successful destroy
	if err := pm.stack.Workspace().RemoveStack(ctx, pm.stackName); err != nil {
		return fmt.Errorf("failed to remove stack: %w", err)
	}

	return nil
}

// GetOutputs retrieves the stack outputs (IPs, IDs, etc.)
func (pm *PulumiManager) GetOutputs(ctx context.Context) (map[string]auto.OutputValue, error) {
	outputs, err := pm.stack.Outputs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get stack outputs: %w", err)
	}
	return outputs, nil
}

// GetStackName returns the stack name
func (pm *PulumiManager) GetStackName() string {
	return pm.stackName
}

// ensurePulumiInstalled checks if Pulumi CLI is installed
func ensurePulumiInstalled() error {
	// Check if pulumi is in PATH
	if _, err := exec.LookPath("pulumi"); err == nil {
		return nil
	}

	// Check in common installation paths
	commonPaths := []string{
		filepath.Join(os.Getenv("HOME"), ".pulumi", "bin", "pulumi"),
		"/usr/local/bin/pulumi",
		"/opt/homebrew/bin/pulumi",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			// Add to PATH for current session
			currentPath := os.Getenv("PATH")
			if err := os.Setenv("PATH", filepath.Dir(path)+":"+currentPath); err != nil {
				return fmt.Errorf("failed to set PATH: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("pulumi CLI not found - install via: curl -fsSL https://get.pulumi.com | sh")
}
