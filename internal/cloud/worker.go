package cloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/distributed"
)

// WaitForWorkers waits for the expected number of workers to register
func WaitForWorkers(ctx context.Context, client *distributed.Client, expectedCount int, timeout time.Duration) ([]string, error) {
	if client == nil {
		return nil, fmt.Errorf("distributed client is nil")
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 5 * time.Second

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-ticker.C:
			// Check if deadline exceeded
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for workers (expected: %d, timeout: %v)", expectedCount, timeout)
			}

			// Get all registered workers
			workers, err := client.GetAllWorkers(ctx)
			if err != nil {
				// Log warning but continue polling
				fmt.Printf("Warning: failed to get workers: %v\n", err)
				continue
			}

			// Filter cloud workers (those with cloud- prefix or wosm- prefix)
			cloudWorkers := filterCloudWorkers(workers)

			// Check if we have enough workers
			if len(cloudWorkers) >= expectedCount {
				return cloudWorkers[:expectedCount], nil
			}

			// Log progress
			fmt.Printf("Waiting for workers... (%d/%d registered)\n", len(cloudWorkers), expectedCount)
		}
	}
}

// filterCloudWorkers filters workers that are cloud-provisioned and returns their IDs
func filterCloudWorkers(workers []*distributed.WorkerInfo) []string {
	var cloudWorkers []string

	for _, worker := range workers {
		// Cloud workers typically have wosm- prefix (from --get-public-ip flag)
		// or cloud- alias prefix
		if strings.HasPrefix(worker.ID, "wosm-") || strings.HasPrefix(worker.ID, "cloud-") ||
			strings.HasPrefix(worker.Alias, "wosm-") || strings.HasPrefix(worker.Alias, "cloud-") {
			cloudWorkers = append(cloudWorkers, worker.ID)
		}
	}

	return cloudWorkers
}

// WaitForWorkersWithStatus waits for workers and reports detailed status
func WaitForWorkersWithStatus(ctx context.Context, client *distributed.Client, expectedCount int, timeout time.Duration, statusCallback func(current, expected int)) ([]string, error) {
	if client == nil {
		return nil, fmt.Errorf("distributed client is nil")
	}

	deadline := time.Now().Add(timeout)
	pollInterval := 5 * time.Second

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-ticker.C:
			// Check if deadline exceeded
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("timeout waiting for workers (expected: %d, timeout: %v)", expectedCount, timeout)
			}

			// Get all registered workers
			workers, err := client.GetAllWorkers(ctx)
			if err != nil {
				// Log warning but continue polling
				fmt.Printf("Warning: failed to get workers: %v\n", err)
				continue
			}

			// Filter cloud workers
			cloudWorkers := filterCloudWorkers(workers)

			// Report status via callback
			if statusCallback != nil {
				statusCallback(len(cloudWorkers), expectedCount)
			}

			// Check if we have enough workers
			if len(cloudWorkers) >= expectedCount {
				return cloudWorkers[:expectedCount], nil
			}
		}
	}
}
