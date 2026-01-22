package database

import (
	"context"
	"sync"
	"testing"
)

func TestDistributedHooksRegistration(t *testing.T) {
	// Ensure clean state
	UnregisterDistributedHooks()

	// Verify no hooks initially
	if GetDistributedHooks() != nil {
		t.Error("Expected no hooks initially")
	}

	// Register hooks
	hooks := &DistributedHooks{
		ShouldUseRedis: func() bool { return true },
	}
	RegisterDistributedHooks(hooks)

	// Verify hooks are registered
	if GetDistributedHooks() == nil {
		t.Error("Expected hooks to be registered")
	}

	// Unregister
	UnregisterDistributedHooks()

	// Verify hooks are unregistered
	if GetDistributedHooks() != nil {
		t.Error("Expected hooks to be unregistered")
	}
}

func TestShouldUseDistributedHooks(t *testing.T) {
	// Ensure clean state
	UnregisterDistributedHooks()

	// Without hooks, should return false
	if shouldUseDistributedHooks() {
		t.Error("Expected false when no hooks registered")
	}

	// Register hooks that return false
	hooks := &DistributedHooks{
		ShouldUseRedis: func() bool { return false },
	}
	RegisterDistributedHooks(hooks)

	if shouldUseDistributedHooks() {
		t.Error("Expected false when ShouldUseRedis returns false")
	}

	// Register hooks that return true
	hooks.ShouldUseRedis = func() bool { return true }
	RegisterDistributedHooks(hooks)

	if !shouldUseDistributedHooks() {
		t.Error("Expected true when ShouldUseRedis returns true")
	}

	// Cleanup
	UnregisterDistributedHooks()
}

func TestTrySendRunToRedis(t *testing.T) {
	// Ensure clean state
	UnregisterDistributedHooks()

	ctx := context.Background()
	run := &Run{ID: 1, RunUUID: "run-123"}

	// Without hooks, should return false
	if trySendRunToRedis(ctx, run) {
		t.Error("Expected false when no hooks registered")
	}

	// Register hooks with working SendRun
	var sentRun *Run
	hooks := &DistributedHooks{
		ShouldUseRedis: func() bool { return true },
		SendRun: func(ctx context.Context, r *Run) error {
			sentRun = r
			return nil
		},
	}
	RegisterDistributedHooks(hooks)

	// Should now return true and call SendRun
	if !trySendRunToRedis(ctx, run) {
		t.Error("Expected true when hooks properly configured")
	}
	if sentRun == nil {
		t.Error("SendRun was not called")
	}
	if sentRun.ID != run.ID {
		t.Errorf("Wrong run sent: expected %d, got %d", run.ID, sentRun.ID)
	}

	// Cleanup
	UnregisterDistributedHooks()
}

func TestTrySendStepResultToRedis(t *testing.T) {
	// Ensure clean state
	UnregisterDistributedHooks()

	ctx := context.Background()
	step := &StepResult{ID: "test-step", StepName: "test-step-name"}

	// Without hooks, should return false
	if trySendStepResultToRedis(ctx, step) {
		t.Error("Expected false when no hooks registered")
	}

	// Register hooks with working SendStepResult
	var sentStep *StepResult
	hooks := &DistributedHooks{
		ShouldUseRedis: func() bool { return true },
		SendStepResult: func(ctx context.Context, s *StepResult) error {
			sentStep = s
			return nil
		},
	}
	RegisterDistributedHooks(hooks)

	// Should now return true and call SendStepResult
	if !trySendStepResultToRedis(ctx, step) {
		t.Error("Expected true when hooks properly configured")
	}
	if sentStep == nil {
		t.Error("SendStepResult was not called")
	}
	if sentStep.ID != step.ID {
		t.Errorf("Wrong step sent: expected %s, got %s", step.ID, sentStep.ID)
	}

	// Cleanup
	UnregisterDistributedHooks()
}

func TestTrySendEventLogToRedis(t *testing.T) {
	// Ensure clean state
	UnregisterDistributedHooks()

	ctx := context.Background()
	event := &EventLog{Topic: "test.topic", EventID: "event-123"}

	// Without hooks, should return false
	if trySendEventLogToRedis(ctx, event) {
		t.Error("Expected false when no hooks registered")
	}

	// Register hooks with working SendEventLog
	var sentEvent *EventLog
	hooks := &DistributedHooks{
		ShouldUseRedis: func() bool { return true },
		SendEventLog: func(ctx context.Context, e *EventLog) error {
			sentEvent = e
			return nil
		},
	}
	RegisterDistributedHooks(hooks)

	// Should now return true and call SendEventLog
	if !trySendEventLogToRedis(ctx, event) {
		t.Error("Expected true when hooks properly configured")
	}
	if sentEvent == nil {
		t.Error("SendEventLog was not called")
	}
	if sentEvent.EventID != event.EventID {
		t.Errorf("Wrong event sent: expected %s, got %s", event.EventID, sentEvent.EventID)
	}

	// Cleanup
	UnregisterDistributedHooks()
}

func TestTrySendArtifactToRedis(t *testing.T) {
	// Ensure clean state
	UnregisterDistributedHooks()

	ctx := context.Background()
	artifact := &Artifact{ID: "test-artifact", Name: "test-name"}

	// Without hooks, should return false
	if trySendArtifactToRedis(ctx, artifact) {
		t.Error("Expected false when no hooks registered")
	}

	// Register hooks with working SendArtifact
	var sentArtifact *Artifact
	hooks := &DistributedHooks{
		ShouldUseRedis: func() bool { return true },
		SendArtifact: func(ctx context.Context, a *Artifact) error {
			sentArtifact = a
			return nil
		},
	}
	RegisterDistributedHooks(hooks)

	// Should now return true and call SendArtifact
	if !trySendArtifactToRedis(ctx, artifact) {
		t.Error("Expected true when hooks properly configured")
	}
	if sentArtifact == nil {
		t.Error("SendArtifact was not called")
	}
	if sentArtifact.ID != artifact.ID {
		t.Errorf("Wrong artifact sent: expected %s, got %s", artifact.ID, sentArtifact.ID)
	}

	// Cleanup
	UnregisterDistributedHooks()
}

func TestConcurrentHooksAccess(t *testing.T) {
	// Ensure clean state
	UnregisterDistributedHooks()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Register initial hooks
	callCount := 0
	var mu sync.Mutex
	hooks := &DistributedHooks{
		ShouldUseRedis: func() bool { return true },
		SendRun: func(ctx context.Context, r *Run) error {
			mu.Lock()
			callCount++
			mu.Unlock()
			return nil
		},
	}
	RegisterDistributedHooks(hooks)

	// Spawn multiple goroutines accessing hooks concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			run := &Run{ID: 1, RunUUID: "concurrent-test"}
			trySendRunToRedis(ctx, run)
		}()
	}

	wg.Wait()

	// All calls should have succeeded
	mu.Lock()
	if callCount != 100 {
		t.Errorf("Expected 100 calls, got %d", callCount)
	}
	mu.Unlock()

	// Cleanup
	UnregisterDistributedHooks()
}
