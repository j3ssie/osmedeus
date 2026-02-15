package functions

import (
	"context"
	"sync"

	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// WorkerSSHInfo holds SSH connection details for a worker node.
// Lives in the functions package to avoid circular imports between
// functions and distributed packages.
type WorkerSSHInfo struct {
	ID      string
	Host    string // PublicIP or IPAddress
	User    string // default "root"
	KeyPath string // SSHKeysPath
	Alias   string
	Port    int // default 22
}

// ExecuteHooks provides callback functions for sending execute requests
// to the master node when running in distributed worker mode.
// This follows the same pattern as database.DistributedHooks to avoid
// import cycles between functions and distributed packages.
type ExecuteHooks struct {
	// SendExecuteRequest pushes an execute request via Redis.
	// targetRole is "master" or "worker" to control where the request is executed.
	// targetScope specifies which worker(s) to target: "all", alias, worker ID, or public IP.
	SendExecuteRequest func(ctx context.Context, action, expr, workflow, target, params, targetRole, targetScope string) error

	// ShouldUseRedis returns true if we are in distributed worker mode
	ShouldUseRedis func() bool

	// ResolveWorkerSSH looks up SSH connection info for a worker by ID, alias, or IP.
	ResolveWorkerSSH func(ctx context.Context, identifier string) (*WorkerSSHInfo, error)
}

var (
	executeHooks  *ExecuteHooks
	executeHookMu sync.RWMutex
)

// RegisterExecuteHooks registers callbacks for distributed execute requests.
// Called by the distributed package at worker startup.
func RegisterExecuteHooks(hooks *ExecuteHooks) {
	executeHookMu.Lock()
	defer executeHookMu.Unlock()
	executeHooks = hooks
}

// UnregisterExecuteHooks removes the execute hooks.
func UnregisterExecuteHooks() {
	executeHookMu.Lock()
	defer executeHookMu.Unlock()
	executeHooks = nil
}

// trySendExecuteRequest attempts to send an execute request via Redis if in distributed mode.
// targetRole is "master" or "worker" to control where the request is executed.
// targetScope specifies which worker(s) to target: "all", alias, worker ID, or public IP (only used when targetRole="worker").
// Returns true if sent to Redis, false if caller should fall back to local execution.
func trySendExecuteRequest(action, expr, workflow, target, params, targetRole, targetScope string) bool {
	executeHookMu.RLock()
	hooks := executeHooks
	executeHookMu.RUnlock()

	if hooks == nil || hooks.ShouldUseRedis == nil || !hooks.ShouldUseRedis() {
		return false
	}

	if hooks.SendExecuteRequest == nil {
		return false
	}

	if err := hooks.SendExecuteRequest(context.Background(), action, expr, workflow, target, params, targetRole, targetScope); err != nil {
		logger.Get().Warn("trySendExecuteRequest: failed to send to Redis, falling back to local",
			zap.String("action", action),
			zap.String("targetRole", targetRole),
			zap.Error(err))
		return false
	}
	return true
}

// tryResolveWorkerSSH attempts to resolve SSH connection info for a worker
// by ID, alias, or IP address. Returns nil if hooks are not registered or resolution fails.
func tryResolveWorkerSSH(identifier string) *WorkerSSHInfo {
	executeHookMu.RLock()
	hooks := executeHooks
	executeHookMu.RUnlock()

	if hooks == nil || hooks.ResolveWorkerSSH == nil {
		return nil
	}

	info, err := hooks.ResolveWorkerSSH(context.Background(), identifier)
	if err != nil {
		logger.Get().Warn("tryResolveWorkerSSH failed",
			zap.String("identifier", identifier),
			zap.Error(err))
		return nil
	}
	return info
}
