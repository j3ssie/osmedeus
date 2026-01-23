package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteCoordinator_Basic(t *testing.T) {
	cfg := &WriteCoordinatorConfig{
		FlushThreshold: 5,
		FlushInterval:  1 * time.Second,
	}

	wc := NewWriteCoordinator(1, "test-uuid", cfg)
	require.NotNil(t, wc)

	// Initially empty
	assert.Equal(t, 0, wc.Len())
	assert.Equal(t, 0, wc.PendingProgress())

	// Add step results
	now := time.Now()
	wc.AddStepResult("step1", "bash", "success", "echo hello", "hello", "", nil, 100, &now, &now)
	assert.Equal(t, 1, wc.Len())

	// Add progress
	wc.IncrementProgress(1)
	assert.Equal(t, 1, wc.PendingProgress())

	// Stop the coordinator
	wc.Stop()
}

func TestWriteCoordinator_AutoFlush(t *testing.T) {
	cfg := &WriteCoordinatorConfig{
		FlushThreshold: 3, // Small threshold for testing
		FlushInterval:  10 * time.Second,
	}

	wc := NewWriteCoordinator(1, "test-uuid", cfg)
	defer wc.Stop()

	now := time.Now()

	// Add 2 step results - should not auto-flush
	wc.AddStepResult("step1", "bash", "success", "cmd1", "out1", "", nil, 100, &now, &now)
	wc.AddStepResult("step2", "bash", "success", "cmd2", "out2", "", nil, 100, &now, &now)
	assert.Equal(t, 2, wc.Len())

	// Add 3rd step result - should trigger auto-flush (threshold reached)
	// Note: Without a DB, the flush clears the buffer but doesn't persist
	wc.AddStepResult("step3", "bash", "success", "cmd3", "out3", "", nil, 100, &now, &now)

	// After auto-flush, buffer should be cleared
	assert.Equal(t, 0, wc.Len())
}

func TestWriteCoordinator_ManualFlush(t *testing.T) {
	cfg := &WriteCoordinatorConfig{
		FlushThreshold: 100, // High threshold to prevent auto-flush
		FlushInterval:  10 * time.Second,
	}

	wc := NewWriteCoordinator(1, "test-uuid", cfg)
	defer wc.Stop()

	now := time.Now()

	// Add step results
	wc.AddStepResult("step1", "bash", "success", "cmd1", "out1", "", nil, 100, &now, &now)
	wc.IncrementProgress(1)

	assert.Equal(t, 1, wc.Len())
	assert.Equal(t, 1, wc.PendingProgress())

	// Manual flush
	err := wc.Flush(context.Background())
	require.NoError(t, err)

	// After flush, buffer should be cleared
	assert.Equal(t, 0, wc.Len())
	assert.Equal(t, 0, wc.PendingProgress())
}

func TestWriteCoordinator_FlushAll(t *testing.T) {
	cfg := &WriteCoordinatorConfig{
		FlushThreshold: 100,
		FlushInterval:  10 * time.Second,
	}

	wc := NewWriteCoordinator(1, "test-uuid", cfg)

	now := time.Now()
	wc.AddStepResult("step1", "bash", "success", "cmd1", "out1", "", nil, 100, &now, &now)
	wc.IncrementProgress(1)

	// FlushAll should stop the coordinator and flush
	err := wc.FlushAll(context.Background())
	require.NoError(t, err)

	// After FlushAll, buffer should be cleared
	assert.Equal(t, 0, wc.Len())
	assert.Equal(t, 0, wc.PendingProgress())

	// Subsequent FlushAll should be a no-op
	err = wc.FlushAll(context.Background())
	require.NoError(t, err)
}

func TestWriteCoordinator_EmptyFlush(t *testing.T) {
	wc := NewWriteCoordinator(1, "test-uuid", nil)
	defer wc.Stop()

	// Flushing empty coordinator should be a no-op
	err := wc.Flush(context.Background())
	require.NoError(t, err)
}

func TestWriteCoordinator_DefaultConfig(t *testing.T) {
	cfg := DefaultWriteCoordinatorConfig()
	require.NotNil(t, cfg)
	assert.Equal(t, 10, cfg.FlushThreshold)
	assert.Equal(t, 5*time.Second, cfg.FlushInterval)
}
