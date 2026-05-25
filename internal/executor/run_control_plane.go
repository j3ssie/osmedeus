package executor

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// ActiveRun represents a currently executing workflow run
type ActiveRun struct {
	RunUUID      string
	Ctx          context.Context // cancellable context bound to this run
	Cancel       context.CancelFunc
	PIDs         *sync.Map // map[int]struct{} - currently running PIDs
	TmuxSessions *sync.Map // map[string]struct{} - tmux sessions spawned by tmux_run
	StartedAt    time.Time
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

// AddTmuxSession records a tmux session name spawned during this run.
func (a *ActiveRun) AddTmuxSession(name string) {
	if a.TmuxSessions != nil && name != "" {
		a.TmuxSessions.Store(name, struct{}{})
	}
}

// RemoveTmuxSession drops a tmux session name from tracking (e.g. on tmux_kill).
func (a *ActiveRun) RemoveTmuxSession(name string) {
	if a.TmuxSessions != nil && name != "" {
		a.TmuxSessions.Delete(name)
	}
}

// KillAllTmuxSessions issues `tmux kill-session` for every tracked session and
// returns the list of session names it attempted to kill. Errors are swallowed
// because best-effort cleanup is preferred over aborting cancellation.
func (a *ActiveRun) KillAllTmuxSessions() []string {
	var killed []string
	if a.TmuxSessions == nil {
		return killed
	}

	tmuxBin, lookErr := exec.LookPath("tmux")
	a.TmuxSessions.Range(func(key, _ any) bool {
		name, ok := key.(string)
		if !ok || name == "" {
			return true
		}
		if lookErr == nil {
			_ = exec.Command(tmuxBin, "kill-session", "-t", name).Run()
		}
		killed = append(killed, name)
		a.TmuxSessions.Delete(name)
		return true
	})
	return killed
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

// RunControlPlane tracks active workflow runs for cancellation support
type RunControlPlane struct {
	mu   sync.RWMutex
	runs map[string]*ActiveRun
}

var globalControlPlane *RunControlPlane
var controlPlaneOnce sync.Once

// GetRunControlPlane returns the singleton run control plane
func GetRunControlPlane() *RunControlPlane {
	controlPlaneOnce.Do(func() {
		globalControlPlane = &RunControlPlane{
			runs: make(map[string]*ActiveRun),
		}
	})
	return globalControlPlane
}

// Register adds a new run to the control plane.
// The provided context must be the cancellable context whose CancelFunc is
// `cancel`; storing it lets background helpers (e.g. ssh_exec) derive their
// own deadlines from the run's lifetime so cancellation propagates.
func (r *RunControlPlane) Register(runUUID string, cancel context.CancelFunc) *ActiveRun {
	return r.RegisterWithContext(runUUID, nil, cancel)
}

// RegisterWithContext is like Register but also records the parent context.
// New callers should prefer this over Register so the stored Ctx is non-nil.
func (r *RunControlPlane) RegisterWithContext(runUUID string, ctx context.Context, cancel context.CancelFunc) *ActiveRun {
	r.mu.Lock()
	defer r.mu.Unlock()

	activeRun := &ActiveRun{
		RunUUID:      runUUID,
		Ctx:          ctx,
		Cancel:       cancel,
		PIDs:         &sync.Map{},
		TmuxSessions: &sync.Map{},
		StartedAt:    time.Now(),
	}
	r.runs[runUUID] = activeRun
	return activeRun
}

// Unregister removes a run from the control plane
func (r *RunControlPlane) Unregister(runUUID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.runs, runUUID)
}

// Get retrieves an active run by its UUID
func (r *RunControlPlane) Get(runUUID string) *ActiveRun {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.runs[runUUID]
}

// Cancel cancels a run by calling its cancel function and killing all tracked
// PIDs and tmux sessions. Returns the list of killed PIDs, killed tmux sessions,
// and any error.
func (r *RunControlPlane) Cancel(runUUID string) ([]int, []string, error) {
	r.mu.Lock()
	activeRun, exists := r.runs[runUUID]
	r.mu.Unlock()

	if !exists {
		return nil, nil, fmt.Errorf("run %s not found in control plane", runUUID)
	}

	// First, cancel the context to stop any new operations
	if activeRun.Cancel != nil {
		activeRun.Cancel()
	}

	// Then tear down tracked side effects: tmux sessions outlive the parent
	// process by design, so kill them before any tracked PIDs.
	killedSessions := activeRun.KillAllTmuxSessions()
	killedPIDs := activeRun.KillAllPIDs()

	return killedPIDs, killedSessions, nil
}

// AddPID adds a PID to a run's tracked processes
func (r *RunControlPlane) AddPID(runUUID string, pid int) {
	r.mu.RLock()
	activeRun := r.runs[runUUID]
	r.mu.RUnlock()

	if activeRun != nil {
		activeRun.AddPID(pid)
	}
}

// RemovePID removes a PID from a run's tracked processes
func (r *RunControlPlane) RemovePID(runUUID string, pid int) {
	r.mu.RLock()
	activeRun := r.runs[runUUID]
	r.mu.RUnlock()

	if activeRun != nil {
		activeRun.RemovePID(pid)
	}
}

// AddTmuxSession registers a tmux session name against a run so the session
// can be torn down when the run is cancelled.
func (r *RunControlPlane) AddTmuxSession(runUUID, name string) {
	if runUUID == "" || name == "" {
		return
	}
	r.mu.RLock()
	activeRun := r.runs[runUUID]
	r.mu.RUnlock()

	if activeRun != nil {
		activeRun.AddTmuxSession(name)
	}
}

// RemoveTmuxSession unregisters a tmux session name (e.g. after tmux_kill).
func (r *RunControlPlane) RemoveTmuxSession(runUUID, name string) {
	if runUUID == "" || name == "" {
		return
	}
	r.mu.RLock()
	activeRun := r.runs[runUUID]
	r.mu.RUnlock()

	if activeRun != nil {
		activeRun.RemoveTmuxSession(name)
	}
}

// ListActive returns a list of all active run UUIDs
func (r *RunControlPlane) ListActive() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	uuids := make([]string, 0, len(r.runs))
	for uuid := range r.runs {
		uuids = append(uuids, uuid)
	}
	return uuids
}

// Count returns the number of active runs
func (r *RunControlPlane) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.runs)
}
