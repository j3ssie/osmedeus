package executor

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRunRegistryRegisterAndGet(t *testing.T) {
	registry := &RunRegistry{
		runs: make(map[string]*ActiveRun),
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	runUUID := "test-run-123"
	activeRun := registry.Register(runUUID, cancel)

	if activeRun == nil {
		t.Fatal("Register returned nil")
	}

	if activeRun.RunUUID != runUUID {
		t.Errorf("Expected RunUUID %s, got %s", runUUID, activeRun.RunUUID)
	}

	// Test Get
	retrieved := registry.Get(runUUID)
	if retrieved != activeRun {
		t.Error("Get returned different instance")
	}

	// Test Get for non-existent run
	nonExistent := registry.Get("non-existent")
	if nonExistent != nil {
		t.Error("Get should return nil for non-existent run")
	}
}

func TestRunRegistryUnregister(t *testing.T) {
	registry := &RunRegistry{
		runs: make(map[string]*ActiveRun),
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	runUUID := "test-run-456"
	registry.Register(runUUID, cancel)

	// Verify registered
	if registry.Get(runUUID) == nil {
		t.Fatal("Run should be registered")
	}

	// Unregister
	registry.Unregister(runUUID)

	// Verify unregistered
	if registry.Get(runUUID) != nil {
		t.Error("Run should be unregistered")
	}
}

func TestRunRegistryCancel(t *testing.T) {
	registry := &RunRegistry{
		runs: make(map[string]*ActiveRun),
	}

	ctx, cancel := context.WithCancel(context.Background())
	runUUID := "test-run-789"
	registry.Register(runUUID, cancel)

	// Cancel should call the cancel function
	_, err := registry.Cancel(runUUID)
	if err != nil {
		t.Errorf("Cancel returned error: %v", err)
	}

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected
	default:
		t.Error("Context should be cancelled")
	}

	// Cancel non-existent run
	_, err = registry.Cancel("non-existent")
	if err == nil {
		t.Error("Cancel should return error for non-existent run")
	}
}

func TestRunRegistryPIDTracking(t *testing.T) {
	registry := &RunRegistry{
		runs: make(map[string]*ActiveRun),
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	runUUID := "test-run-pid"
	registry.Register(runUUID, cancel)

	// Add PIDs
	registry.AddPID(runUUID, 1234)
	registry.AddPID(runUUID, 5678)

	// Verify PIDs are tracked
	activeRun := registry.Get(runUUID)
	count := 0
	activeRun.PIDs.Range(func(_, _ any) bool {
		count++
		return true
	})
	if count != 2 {
		t.Errorf("Expected 2 PIDs, got %d", count)
	}

	// Remove a PID
	registry.RemovePID(runUUID, 1234)

	// Verify PID removed
	count = 0
	activeRun.PIDs.Range(func(_, _ any) bool {
		count++
		return true
	})
	if count != 1 {
		t.Errorf("Expected 1 PID after removal, got %d", count)
	}
}

func TestRunRegistryKillAllPIDs(t *testing.T) {
	// Create a fresh ActiveRun (not using the registry to avoid syscall issues)
	activeRun := &ActiveRun{
		RunUUID:   "test-kill",
		PIDs:      &sync.Map{},
		StartedAt: time.Now(),
	}

	// Add fake PIDs (these won't exist as real processes)
	activeRun.AddPID(99999)
	activeRun.AddPID(99998)

	// KillAllPIDs will try to kill these (and fail silently)
	killed := activeRun.KillAllPIDs()

	// Should report the PIDs it attempted to kill
	if len(killed) != 2 {
		t.Errorf("Expected 2 killed PIDs, got %d", len(killed))
	}

	// PIDs should be cleared
	count := 0
	activeRun.PIDs.Range(func(_, _ any) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Expected 0 PIDs after kill, got %d", count)
	}
}

func TestRunRegistryConcurrency(t *testing.T) {
	registry := &RunRegistry{
		runs: make(map[string]*ActiveRun),
	}

	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, cancel := context.WithCancel(context.Background())
			defer cancel()

			runUUID := "concurrent-run-" + string(rune('a'+i%26))
			registry.Register(runUUID, cancel)
			registry.AddPID(runUUID, i+1000)
			registry.RemovePID(runUUID, i+1000)
			registry.Get(runUUID)
			registry.Unregister(runUUID)
		}(i)
	}

	wg.Wait()

	// Should have no leftover runs
	if registry.Count() != 0 {
		t.Errorf("Expected 0 runs, got %d", registry.Count())
	}
}

func TestRunRegistryListActive(t *testing.T) {
	registry := &RunRegistry{
		runs: make(map[string]*ActiveRun),
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	registry.Register("run-1", cancel)
	registry.Register("run-2", cancel)
	registry.Register("run-3", cancel)

	active := registry.ListActive()
	if len(active) != 3 {
		t.Errorf("Expected 3 active runs, got %d", len(active))
	}

	if registry.Count() != 3 {
		t.Errorf("Expected count 3, got %d", registry.Count())
	}
}

func TestGetRunRegistrySingleton(t *testing.T) {
	// Get the singleton twice
	reg1 := GetRunRegistry()
	reg2 := GetRunRegistry()

	// Should be the same instance
	if reg1 != reg2 {
		t.Error("GetRunRegistry should return the same singleton instance")
	}
}
