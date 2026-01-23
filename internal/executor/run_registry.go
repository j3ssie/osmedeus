package executor

import (
	"context"
	"fmt"
	"sync"
	"syscall"
	"time"
)

// ActiveRun represents a currently executing workflow run
type ActiveRun struct {
	RunUUID   string
	Cancel    context.CancelFunc
	PIDs      *sync.Map // map[int]struct{} - currently running PIDs
	StartedAt time.Time
}

// AddPID adds a process ID to this run's tracked PIDs
func (a *ActiveRun) AddPID(pid int) {
	if a.PIDs != nil {
		a.PIDs.Store(pid, struct{}{})
	}
}

// RemovePID removes a process ID from this run's tracked PIDs
func (a *ActiveRun) RemovePID(pid int) {
	if a.PIDs != nil {
		a.PIDs.Delete(pid)
	}
}

// KillAllPIDs sends SIGKILL to all tracked PIDs and returns the list of killed PIDs
func (a *ActiveRun) KillAllPIDs() []int {
	var killed []int
	if a.PIDs == nil {
		return killed
	}

	a.PIDs.Range(func(key, _ any) bool {
		pid, ok := key.(int)
		if !ok {
			return true
		}

		// Kill the process group (negative PID kills all processes in the group)
		// This ensures child processes are also terminated
		if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
			// Try killing just the process if process group kill fails
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
		killed = append(killed, pid)
		a.PIDs.Delete(pid)
		return true
	})

	return killed
}

// RunRegistry tracks active workflow runs for cancellation support
type RunRegistry struct {
	mu   sync.RWMutex
	runs map[string]*ActiveRun
}

var globalRegistry *RunRegistry
var registryOnce sync.Once

// GetRunRegistry returns the singleton run registry
func GetRunRegistry() *RunRegistry {
	registryOnce.Do(func() {
		globalRegistry = &RunRegistry{
			runs: make(map[string]*ActiveRun),
		}
	})
	return globalRegistry
}

// Register adds a new run to the registry
func (r *RunRegistry) Register(runUUID string, cancel context.CancelFunc) *ActiveRun {
	r.mu.Lock()
	defer r.mu.Unlock()

	activeRun := &ActiveRun{
		RunUUID:   runUUID,
		Cancel:    cancel,
		PIDs:      &sync.Map{},
		StartedAt: time.Now(),
	}
	r.runs[runUUID] = activeRun
	return activeRun
}

// Unregister removes a run from the registry
func (r *RunRegistry) Unregister(runUUID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.runs, runUUID)
}

// Get retrieves an active run by its UUID
func (r *RunRegistry) Get(runUUID string) *ActiveRun {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.runs[runUUID]
}

// Cancel cancels a run by calling its cancel function and killing all tracked PIDs.
// Returns the list of killed PIDs and any error.
func (r *RunRegistry) Cancel(runUUID string) ([]int, error) {
	r.mu.Lock()
	activeRun, exists := r.runs[runUUID]
	r.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("run %s not found in registry", runUUID)
	}

	// First, cancel the context to stop any new operations
	if activeRun.Cancel != nil {
		activeRun.Cancel()
	}

	// Then kill all tracked processes
	killedPIDs := activeRun.KillAllPIDs()

	return killedPIDs, nil
}

// AddPID adds a PID to a run's tracked processes
func (r *RunRegistry) AddPID(runUUID string, pid int) {
	r.mu.RLock()
	activeRun := r.runs[runUUID]
	r.mu.RUnlock()

	if activeRun != nil {
		activeRun.AddPID(pid)
	}
}

// RemovePID removes a PID from a run's tracked processes
func (r *RunRegistry) RemovePID(runUUID string, pid int) {
	r.mu.RLock()
	activeRun := r.runs[runUUID]
	r.mu.RUnlock()

	if activeRun != nil {
		activeRun.RemovePID(pid)
	}
}

// ListActive returns a list of all active run UUIDs
func (r *RunRegistry) ListActive() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	uuids := make([]string, 0, len(r.runs))
	for uuid := range r.runs {
		uuids = append(uuids, uuid)
	}
	return uuids
}

// Count returns the number of active runs
func (r *RunRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.runs)
}
