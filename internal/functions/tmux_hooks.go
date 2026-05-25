package functions

import "sync"

// TmuxHooks provides callbacks for tracking tmux sessions created via
// tmux_run/tmux_kill against an active run. This indirection mirrors
// ExecuteHooks and avoids an import cycle between functions and executor.
type TmuxHooks struct {
	// OnSessionCreated fires after tmux_run successfully creates a session.
	// runUUID will be empty when invoked outside a tracked run (e.g. ad-hoc
	// `osmedeus func e`); implementations must tolerate that.
	OnSessionCreated func(runUUID, sessionName string)

	// OnSessionDestroyed fires after tmux_kill successfully destroys a session,
	// so the tracker can drop it before run cancellation tries to kill it again.
	OnSessionDestroyed func(runUUID, sessionName string)
}

var (
	tmuxHooks  *TmuxHooks
	tmuxHookMu sync.RWMutex
)

// RegisterTmuxHooks installs callbacks for tmux session lifecycle events.
// Called by the executor package at init.
func RegisterTmuxHooks(hooks *TmuxHooks) {
	tmuxHookMu.Lock()
	defer tmuxHookMu.Unlock()
	tmuxHooks = hooks
}

// UnregisterTmuxHooks clears the registered tmux hooks (used in tests).
func UnregisterTmuxHooks() {
	tmuxHookMu.Lock()
	defer tmuxHookMu.Unlock()
	tmuxHooks = nil
}

func notifyTmuxSessionCreated(runUUID, sessionName string) {
	tmuxHookMu.RLock()
	h := tmuxHooks
	tmuxHookMu.RUnlock()
	if h != nil && h.OnSessionCreated != nil {
		h.OnSessionCreated(runUUID, sessionName)
	}
}

func notifyTmuxSessionDestroyed(runUUID, sessionName string) {
	tmuxHookMu.RLock()
	h := tmuxHooks
	tmuxHookMu.RUnlock()
	if h != nil && h.OnSessionDestroyed != nil {
		h.OnSessionDestroyed(runUUID, sessionName)
	}
}
