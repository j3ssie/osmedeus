package database

import (
	"context"
	"fmt"
	"time"
)

func EnsureWorkspaceRuntime(ctx context.Context, name, localPath, runWorkflow, stateExecutionLog, stateCompletedFile, stateWorkflowFile, stateWorkflowFolder string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	if name == "" {
		return fmt.Errorf("workspace name cannot be empty")
	}

	now := time.Now()

	ws := &Workspace{
		Name:                name,
		LocalPath:           localPath,
		DataSource:          "local",
		LastRun:             &now,
		RunWorkflow:         runWorkflow,
		StateExecutionLog:   stateExecutionLog,
		StateCompletedFile:  stateCompletedFile,
		StateWorkflowFile:   stateWorkflowFile,
		StateWorkflowFolder: stateWorkflowFolder,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	_, err := db.NewInsert().Model(ws).
		On("CONFLICT (name) DO UPDATE").
		Set("local_path = EXCLUDED.local_path").
		Set("data_source = EXCLUDED.data_source").
		Set("last_run = EXCLUDED.last_run").
		Set("run_workflow = EXCLUDED.run_workflow").
		Set("state_execution_log = EXCLUDED.state_execution_log").
		Set("state_completed_file = EXCLUDED.state_completed_file").
		Set("state_workflow_file = EXCLUDED.state_workflow_file").
		Set("state_workflow_folder = EXCLUDED.state_workflow_folder").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	return err
}
