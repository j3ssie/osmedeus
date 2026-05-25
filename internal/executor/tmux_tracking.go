package executor

import (
	"context"

	"github.com/j3ssie/osmedeus/v5/internal/functions"
)

// init wires the functions package's lifecycle hooks into the run control
// plane:
//   - tmux: sessions spawned by tmux_run get killed when a run is cancelled.
//     Without this, detached `bosm-<rand8>` sessions survive parent termination
//     and keep scanning long after the user clicked "stop" in the UI.
//   - run context: ssh_exec / ssh_rsync / sync_* can look up the active run's
//     cancellable context and derive their own deadlines from it, so run
//     cancellation propagates to ad-hoc SSH/rsync calls.
func init() {
	functions.RegisterTmuxHooks(&functions.TmuxHooks{
		OnSessionCreated: func(runUUID, sessionName string) {
			GetRunControlPlane().AddTmuxSession(runUUID, sessionName)
		},
		OnSessionDestroyed: func(runUUID, sessionName string) {
			GetRunControlPlane().RemoveTmuxSession(runUUID, sessionName)
		},
	})

	functions.RegisterRunContextHooks(&functions.RunContextHooks{
		Lookup: func(runUUID string) context.Context {
			ar := GetRunControlPlane().Get(runUUID)
			if ar == nil {
				return nil
			}
			return ar.Ctx
		},
	})
}
