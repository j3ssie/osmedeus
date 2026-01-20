package database

import (
	"context"
	"sync"
)

// DistributedHooks provides callback functions for routing database writes
// to Redis queues when running in distributed worker mode.
// This avoids import cycles between database and distributed packages.
type DistributedHooks struct {
	// SendRun is called when a Run should be sent to the master via Redis
	SendRun func(ctx context.Context, run *Run) error

	// SendStepResult is called when a StepResult should be sent to the master via Redis
	SendStepResult func(ctx context.Context, step *StepResult) error

	// SendEventLog is called when an EventLog should be sent to the master via Redis
	SendEventLog func(ctx context.Context, event *EventLog) error

	// SendArtifact is called when an Artifact should be sent to the master via Redis
	SendArtifact func(ctx context.Context, artifact *Artifact) error

	// ShouldUseRedis returns true if writes should go to Redis instead of local DB
	ShouldUseRedis func() bool
}

var (
	distributedHooks *DistributedHooks
	hooksMu          sync.RWMutex
)

// RegisterDistributedHooks registers callbacks for distributed mode.
// Called by the distributed package at worker startup.
func RegisterDistributedHooks(hooks *DistributedHooks) {
	hooksMu.Lock()
	defer hooksMu.Unlock()
	distributedHooks = hooks
}

// UnregisterDistributedHooks removes the distributed hooks.
func UnregisterDistributedHooks() {
	hooksMu.Lock()
	defer hooksMu.Unlock()
	distributedHooks = nil
}

// GetDistributedHooks returns the registered hooks (nil if not set).
func GetDistributedHooks() *DistributedHooks {
	hooksMu.RLock()
	defer hooksMu.RUnlock()
	return distributedHooks
}

// shouldUseDistributedHooks checks if we should route to Redis.
func shouldUseDistributedHooks() bool {
	hooksMu.RLock()
	hooks := distributedHooks
	hooksMu.RUnlock()

	if hooks == nil || hooks.ShouldUseRedis == nil {
		return false
	}
	return hooks.ShouldUseRedis()
}

// trySendRunToRedis attempts to send a run to Redis if in distributed worker mode.
// Returns true if sent to Redis, false if should use local DB.
func trySendRunToRedis(ctx context.Context, run *Run) bool {
	if !shouldUseDistributedHooks() {
		return false
	}

	hooksMu.RLock()
	hooks := distributedHooks
	hooksMu.RUnlock()

	if hooks == nil || hooks.SendRun == nil {
		return false
	}

	if err := hooks.SendRun(ctx, run); err != nil {
		// Log error but don't fail - fall back to local DB
		return false
	}
	return true
}

// trySendStepResultToRedis attempts to send a step result to Redis if in distributed worker mode.
// Returns true if sent to Redis, false if should use local DB.
func trySendStepResultToRedis(ctx context.Context, step *StepResult) bool {
	if !shouldUseDistributedHooks() {
		return false
	}

	hooksMu.RLock()
	hooks := distributedHooks
	hooksMu.RUnlock()

	if hooks == nil || hooks.SendStepResult == nil {
		return false
	}

	if err := hooks.SendStepResult(ctx, step); err != nil {
		return false
	}
	return true
}

// trySendEventLogToRedis attempts to send an event log to Redis if in distributed worker mode.
// Returns true if sent to Redis, false if should use local DB.
func trySendEventLogToRedis(ctx context.Context, event *EventLog) bool {
	if !shouldUseDistributedHooks() {
		return false
	}

	hooksMu.RLock()
	hooks := distributedHooks
	hooksMu.RUnlock()

	if hooks == nil || hooks.SendEventLog == nil {
		return false
	}

	if err := hooks.SendEventLog(ctx, event); err != nil {
		return false
	}
	return true
}

// trySendArtifactToRedis attempts to send an artifact to Redis if in distributed worker mode.
// Returns true if sent to Redis, false if should use local DB.
func trySendArtifactToRedis(ctx context.Context, artifact *Artifact) bool {
	if !shouldUseDistributedHooks() {
		return false
	}

	hooksMu.RLock()
	hooks := distributedHooks
	hooksMu.RUnlock()

	if hooks == nil || hooks.SendArtifact == nil {
		return false
	}

	if err := hooks.SendArtifact(ctx, artifact); err != nil {
		return false
	}
	return true
}
