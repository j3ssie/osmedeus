package executor

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRunControlPlaneRegisterAndGet(t *testing.T) {
	controlPlane := &RunControlPlane{
		runs: make(map[string]*ActiveRun),
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	runUUID := "test-run-123"
	activeRun := controlPlane.Register(runUUID, cancel)

	if activeRun == nil {
		t.Fatal("Register returned nil")
		return
	}

	if activeRun.RunUUID != runUUID {
		t.Errorf("Expected RunUUID %s, got %s", runUUID, activeRun.RunUUID)
	}

	// Test Get
	retrieved := controlPlane.Get(runUUID)
	if retrieved != activeRun {
		t.Error("Get returned different instance")
	}

	// Test Get for non-existent run
	nonExistent := controlPlane.Get("non-existent")
	if nonExistent != nil {
		t.Error("Get should return nil for non-existent run")
	}
}

func TestRunControlPlaneUnregister(t *testing.T) {
	controlPlane := &RunControlPlane{
		runs: make(map[string]*ActiveRun),
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	runUUID := "test-run-456"
	controlPlane.Register(runUUID, cancel)

	// Verify registered
	if controlPlane.Get(runUUID) == nil {
		t.Fatal("Run should be registered")
	}

	// Unregister
	controlPlane.Unregister(runUUID)

	// Verify unregistered
	if controlPlane.Get(runUUID) != nil {
		t.Error("Run should be unregistered")
	}
}

func TestRunControlPlaneCancel(t *testing.T) {
	controlPlane := &RunControlPlane{
		runs: make(map[string]*ActiveRun),
	}

	ctx, cancel := context.WithCancel(context.Background())
	runUUID := "test-run-789"
	controlPlane.Register(runUUID, cancel)

	// Cancel should call the cancel function
	_, err := controlPlane.Cancel(runUUID)
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
	_, err = controlPlane.Cancel("non-existent")
	if err == nil {
		t.Error("Cancel should return error for non-existent run")
	}
}

func TestRunControlPlanePIDTracking(t *testing.T) {
	controlPlane := &RunControlPlane{
		runs: make(map[string]*ActiveRun),
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	runUUID := "test-run-pid"
	controlPlane.Register(runUUID, cancel)

	// Add PIDs
	controlPlane.AddPID(runUUID, 1234)
	controlPlane.AddPID(runUUID, 5678)

	// Verify PIDs are tracked
	activeRun := controlPlane.Get(runUUID)
	count := 0
	activeRun.PIDs.Range(func(_, _ any) bool {
		count++
		return true
	})
	if count != 2 {
		t.Errorf("Expected 2 PIDs, got %d", count)
	}

	// Remove a PID
	controlPlane.RemovePID(runUUID, 1234)

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

func TestRunControlPlaneKillAllPIDs(t *testing.T) {
	// Create a fresh ActiveRun (not using the control plane to avoid syscall issues)
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

func TestRunControlPlaneConcurrency(t *testing.T) {
	controlPlane := &RunControlPlane{
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
			controlPlane.Register(runUUID, cancel)
			controlPlane.AddPID(runUUID, i+1000)
			controlPlane.RemovePID(runUUID, i+1000)
			controlPlane.Get(runUUID)
			controlPlane.Unregister(runUUID)
		}(i)
	}

	wg.Wait()

	// Should have no leftover runs
	if controlPlane.Count() != 0 {
		t.Errorf("Expected 0 runs, got %d", controlPlane.Count())
	}
}

func TestRunControlPlaneListActive(t *testing.T) {
	controlPlane := &RunControlPlane{
		runs: make(map[string]*ActiveRun),
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	controlPlane.Register("run-1", cancel)
	controlPlane.Register("run-2", cancel)
	controlPlane.Register("run-3", cancel)

	active := controlPlane.ListActive()
	if len(active) != 3 {
		t.Errorf("Expected 3 active runs, got %d", len(active))
	}

	if controlPlane.Count() != 3 {
		t.Errorf("Expected count 3, got %d", controlPlane.Count())
	}
}

func TestGetRunControlPlaneSingleton(t *testing.T) {
	// Get the singleton twice
	cp1 := GetRunControlPlane()
	cp2 := GetRunControlPlane()

	// Should be the same instance
	if cp1 != cp2 {
		t.Error("GetRunControlPlane should return the same singleton instance")
	}
}
