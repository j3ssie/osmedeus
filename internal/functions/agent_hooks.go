package functions

import (
	"context"
	"sync"
)

// RunAgentFuncType is the signature for the ACP agent runner function.
// It spawns an ACP agent subprocess and returns (stdout, stderr, error).
// This is set by the executor package to avoid circular imports.
type RunAgentFuncType func(ctx context.Context, prompt, agentName string) (string, string, error)

var (
	runAgentFunc   RunAgentFuncType
	runAgentFuncMu sync.RWMutex
)

// RegisterRunAgentFunc registers the ACP agent runner function.
// Called by packages that have access to the executor (e.g., CLI init).
func RegisterRunAgentFunc(fn RunAgentFuncType) {
	runAgentFuncMu.Lock()
	defer runAgentFuncMu.Unlock()
	runAgentFunc = fn
}

// GetRunAgentFunc returns the registered ACP agent runner function, or nil.
func GetRunAgentFunc() RunAgentFuncType {
	runAgentFuncMu.RLock()
	defer runAgentFuncMu.RUnlock()
	return runAgentFunc
}
