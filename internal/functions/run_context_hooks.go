package functions

import (
	"context"
	"sync"
)

// RunContextHooks lets the executor expose the active run's cancellable
// context to helpers in this package without an import cycle. Helpers like
// ssh_exec / ssh_rsync / sync_* default to context.Background() otherwise,
// which means run cancellation never reaches them.
type RunContextHooks struct {
	// Lookup returns the cancellable context tied to the given runUUID,
	// or nil if no run is registered. Implementations must be safe for
	// concurrent use.
	Lookup func(runUUID string) context.Context
}

var (
	runContextHooks  *RunContextHooks
	runContextHookMu sync.RWMutex
)

// RegisterRunContextHooks installs the run-context lookup. Called from the
// executor package at init.
func RegisterRunContextHooks(hooks *RunContextHooks) {
	runContextHookMu.Lock()
	defer runContextHookMu.Unlock()
	runContextHooks = hooks
}

// UnregisterRunContextHooks clears the lookup (used in tests).
func UnregisterRunContextHooks() {
	runContextHookMu.Lock()
	defer runContextHookMu.Unlock()
	runContextHooks = nil
}

// vmScanID returns the RunUUID associated with a vmFunc's current VM, or "".
// Centralised here so SSH helpers don't each reach into VMContext internals.
func vmScanID(vf *vmFunc) string {
	if vf == nil {
		return ""
	}
	if c := vf.getContext(); c != nil {
		return c.scanID
	}
	return ""
}

// runContextFor returns the active run's context for runUUID, or
// context.Background() when no hook is registered or the run isn't tracked.
// Callers should derive their own deadlines from the result.
func runContextFor(runUUID string) context.Context {
	if runUUID == "" {
		return context.Background()
	}
	runContextHookMu.RLock()
	h := runContextHooks
	runContextHookMu.RUnlock()
	if h == nil || h.Lookup == nil {
		return context.Background()
	}
	if c := h.Lookup(runUUID); c != nil {
		return c
	}
	return context.Background()
}
