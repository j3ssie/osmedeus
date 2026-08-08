package database

import (
	"context"
	"fmt"
	"time"
)

// EnsureWorkspaceRuntime upserts a workspace and attributes it to an org.
//
// orgUUID is the org the caller explicitly selected (--org / $OSMEDEUS_ORG / the
// active org). An empty orgUUID means "no org was named", which is treated very
// differently on insert and update:
//
//   - New workspace: it lands in the default org.
//   - Existing workspace: its org is left alone. Re-scanning a workspace that
//     `osmedeus org assign` put into an org must not silently drag it back to the
//     default org just because this run did not pass --org.
//
// Naming an org explicitly does move the workspace, since that is the whole point
// of `osmedeus run --org`.
func EnsureWorkspaceRuntime(ctx context.Context, name, localPath, runWorkflow, stateExecutionLog, stateCompletedFile, stateWorkflowFile, stateWorkflowFolder, orgUUID string) error {
	if db == nil {
		return fmt.Errorf("database not connected")
	}

	if name == "" {
		return fmt.Errorf("workspace name cannot be empty")
	}

	now := time.Now()

	ws := &Workspace{
		Name:                name,
		OrgUUID:             NormalizeOrgUUID(orgUUID),
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

	q := db.NewInsert().Model(ws).
		On("CONFLICT (name) DO UPDATE").
		Set("local_path = EXCLUDED.local_path").
		Set("data_source = EXCLUDED.data_source").
		Set("last_run = EXCLUDED.last_run").
		Set("run_workflow = EXCLUDED.run_workflow").
		Set("state_execution_log = EXCLUDED.state_execution_log").
		Set("state_completed_file = EXCLUDED.state_completed_file").
		Set("state_workflow_file = EXCLUDED.state_workflow_file").
		Set("state_workflow_folder = EXCLUDED.state_workflow_folder").
		Set("updated_at = EXCLUDED.updated_at")

	if orgUUID != "" {
		q = q.Set("org_uuid = EXCLUDED.org_uuid")
	}

	_, err := q.Exec(ctx)
	if err == nil && orgUUID != "" {
		invalidateWorkspaceOrgCache()
	}
	return err
}
