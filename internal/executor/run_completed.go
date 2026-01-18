package executor

import (
	"os"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// ExportRunCompleted writes run completion state to a JSON file
// Uses the same format as run-state.json (StateExport) for consistency
func ExportRunCompleted(path string, result *core.WorkflowResult, execCtx *core.ExecutionContext) error {
	return ExportState(path, result, execCtx)
}

// RemoveRunCompleted removes the run-completed.json file if it exists
func RemoveRunCompleted(path string) {
	if path != "" {
		_ = os.Remove(path)
	}
}
