package executor

import (
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/state"
)

// ExportState exports the current run state to a JSON file
// It uses database data if available, otherwise falls back to in-memory data from result and execCtx
func ExportState(stateFile string, result *core.WorkflowResult, execCtx *core.ExecutionContext) error {
	ctx := buildExportContext(result, execCtx)
	return state.Export(stateFile, ctx)
}

func buildExportContext(result *core.WorkflowResult, execCtx *core.ExecutionContext) *state.ExportContext {
	ctx := &state.ExportContext{}

	// Populate from execCtx
	if execCtx != nil {
		ctx.RunUUID = execCtx.RunUUID
		ctx.WorkflowName = execCtx.WorkflowName
		ctx.WorkflowKind = string(execCtx.WorkflowKind)
		ctx.Target = execCtx.Target
		ctx.WorkspacePath = execCtx.WorkspacePath
		ctx.WorkspaceName = execCtx.WorkspaceName
		ctx.Params = execCtx.Params

		// Read hook/run metadata from execution context variables
		if v, ok := execCtx.GetVariable("HooksEnabled"); ok {
			if b, ok := v.(bool); ok {
				ctx.HooksEnabled = b
			}
		}
		if v, ok := execCtx.GetVariable("RunMode"); ok {
			if s, ok := v.(string); ok {
				ctx.RunMode = s
			}
		}
		if v, ok := execCtx.GetVariable("RunPriority"); ok {
			if s, ok := v.(string); ok {
				ctx.RunPriority = s
			}
		}
	}

	// Populate/override from result
	if result != nil {
		if ctx.RunUUID == "" {
			ctx.RunUUID = result.RunUUID
		}
		if ctx.WorkflowName == "" {
			ctx.WorkflowName = result.WorkflowName
		}
		ctx.WorkflowKind = string(result.WorkflowKind)
		ctx.Target = result.Target
		ctx.Status = string(result.Status)
		startTime := result.StartTime
		endTime := result.EndTime
		ctx.StartedAt = &startTime
		ctx.CompletedAt = &endTime
		ctx.TotalSteps = len(result.Steps)

		completedSteps := 0
		for _, step := range result.Steps {
			if step.Status == core.StepStatusSuccess {
				completedSteps++
			}
		}
		ctx.CompletedSteps = completedSteps

		if result.Error != nil {
			ctx.ErrorMessage = result.Error.Error()
		}
		ctx.Artifacts = result.Artifacts
	}

	return ctx
}
