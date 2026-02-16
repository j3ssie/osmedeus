package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
)

// LifecycleManager orchestrates the complete lifecycle of cloud infrastructure
type LifecycleManager struct {
	cfg      *config.CloudConfigs
	provider Provider
	client   *distributed.Client
	tracker  *CostTracker
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(cfg *config.CloudConfigs, provider Provider, client *distributed.Client) *LifecycleManager {
	return &LifecycleManager{
		cfg:      cfg,
		provider: provider,
		client:   client,
	}
}

// CreateAndRun provisions infrastructure, waits for workers, runs workflow, and cleans up
func (lm *LifecycleManager) CreateAndRun(ctx context.Context, opts *CreateOptions) (*Infrastructure, error) {
	var infra *Infrastructure
	var err error

	// Setup cleanup on failure if configured
	defer func() {
		if err != nil && lm.cfg.Defaults.CleanupOnFailure && infra != nil {
			_ = lm.provider.DestroyInfrastructure(context.Background(), infra)
		}
	}()

	// Step 1: Validate cost limits
	estimate, err := lm.provider.EstimateCost(opts.Mode, opts.InstanceCount)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate cost: %w", err)
	}

	if err := lm.validateCostLimits(estimate); err != nil {
		return nil, fmt.Errorf("cost validation failed: %w", err)
	}

	// Initialize cost tracker
	lm.tracker = NewCostTracker(estimate.HourlyCost, lm.cfg.Limits.MaxTotalSpend)

	// Step 2: Create infrastructure
	infra, err = lm.provider.CreateInfrastructure(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create infrastructure: %w", err)
	}

	// Save state for recovery
	if err := SaveInfrastructureState(infra, lm.cfg.State.Path); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to save infrastructure state: %v\n", err)
	}

	// Step 3: Wait for workers to register
	workerIDs, err := WaitForWorkers(ctx, lm.client, opts.InstanceCount, 5*time.Minute)
	if err != nil {
		return infra, fmt.Errorf("failed to wait for workers: %w", err)
	}

	// Update infrastructure with worker IDs
	for i, workerID := range workerIDs {
		if i < len(infra.Resources) {
			infra.Resources[i].WorkerID = workerID
		}
	}

	return infra, nil
}

// Destroy tears down infrastructure
func (lm *LifecycleManager) Destroy(ctx context.Context, infra *Infrastructure) error {
	if err := lm.provider.DestroyInfrastructure(ctx, infra); err != nil {
		return fmt.Errorf("failed to destroy infrastructure: %w", err)
	}

	// Remove state file
	if err := RemoveInfrastructureState(infra.ID, lm.cfg.State.Path); err != nil {
		// Log warning but don't fail
		fmt.Printf("Warning: failed to remove infrastructure state: %v\n", err)
	}

	return nil
}

// GetStatus retrieves infrastructure status
func (lm *LifecycleManager) GetStatus(ctx context.Context, infra *Infrastructure) (*InfraStatus, error) {
	status, err := lm.provider.GetStatus(ctx, infra)
	if err != nil {
		return nil, fmt.Errorf("failed to get infrastructure status: %w", err)
	}
	return status, nil
}

// validateCostLimits checks if estimated cost is within configured limits
func (lm *LifecycleManager) validateCostLimits(estimate *CostEstimate) error {
	if lm.cfg.Limits.MaxHourlySpend > 0 && estimate.HourlyCost > lm.cfg.Limits.MaxHourlySpend {
		return fmt.Errorf("estimated hourly cost ($%.2f) exceeds limit ($%.2f)",
			estimate.HourlyCost, lm.cfg.Limits.MaxHourlySpend)
	}

	if lm.cfg.Limits.MaxTotalSpend > 0 && estimate.DailyCost > lm.cfg.Limits.MaxTotalSpend {
		return fmt.Errorf("estimated daily cost ($%.2f) exceeds total spend limit ($%.2f)",
			estimate.DailyCost, lm.cfg.Limits.MaxTotalSpend)
	}

	return nil
}

// MonitorCost monitors ongoing costs during execution
func (lm *LifecycleManager) MonitorCost(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := lm.tracker.CheckLimits(); err != nil {
				return err
			}
		}
	}
}
