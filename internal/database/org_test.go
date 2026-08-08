package database

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOrgTestDB(t *testing.T) func() {
	t.Helper()
	tmpDir := t.TempDir()
	cfg := &config.Config{
		BaseFolder: tmpDir,
		Database: config.DatabaseConfig{
			DBEngine: "sqlite",
			DBPath:   filepath.Join(tmpDir, "test_org.sqlite"),
		},
	}

	_, err := Connect(cfg)
	require.NoError(t, err)
	require.NoError(t, Migrate(context.Background()))

	return func() {
		_ = Close()
		SetDB(nil)
	}
}

// TestDefaultOrgUUIDMatchesStructTags guards the one piece of duplication the org
// layer needs: struct tags cannot reference a constant, so the default UUID is
// written out longhand in every org_uuid tag. If DefaultOrgUUID is ever changed
// without updating the tags, fresh databases would use a different default than
// migrated ones.
func TestDefaultOrgUUIDMatchesStructTags(t *testing.T) {
	models := []interface{}{Run{}, Asset{}, Workspace{}, Vulnerability{}}

	for _, model := range models {
		typ := reflect.TypeOf(model)
		field, ok := typ.FieldByName("OrgUUID")
		require.Truef(t, ok, "%s has no OrgUUID field", typ.Name())

		tag := field.Tag.Get("bun")
		want := fmt.Sprintf("default:'%s'", DefaultOrgUUID)
		assert.Containsf(t, tag, want, "%s.OrgUUID bun tag %q must carry %q", typ.Name(), tag, want)
		assert.Containsf(t, tag, "notnull", "%s.OrgUUID must be notnull", typ.Name())
	}
}

func TestMigrateSeedsDefaultOrg(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()

	org, err := GetOrgByUUID(ctx, DefaultOrgUUID)
	require.NoError(t, err)
	assert.Equal(t, DefaultOrgName, org.Name)
	assert.True(t, org.IsDefault())

	// Migrate is idempotent - re-running must not duplicate the default org.
	require.NoError(t, Migrate(ctx))

	orgs, err := ListOrgs(ctx)
	require.NoError(t, err)
	require.Len(t, orgs, 1)
}

func TestNormalizeOrgUUID(t *testing.T) {
	assert.Equal(t, DefaultOrgUUID, NormalizeOrgUUID(""))
	assert.Equal(t, DefaultOrgUUID, NormalizeOrgUUID("   "))
	assert.Equal(t, "custom-uuid", NormalizeOrgUUID("custom-uuid"))
}

func TestCreateOrgAndResolveRef(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()

	created, err := CreateOrg(ctx, "acme", "ACME Corp", "", []string{"corp"})
	require.NoError(t, err)
	assert.NotEmpty(t, created.UUID)

	// --org accepts either form.
	byName, err := ResolveOrgRef(ctx, "acme")
	require.NoError(t, err)
	assert.Equal(t, created.UUID, byName.UUID)

	byUUID, err := ResolveOrgRef(ctx, created.UUID)
	require.NoError(t, err)
	assert.Equal(t, "acme", byUUID.Name)

	// Name matching is case-insensitive.
	byMixedCase, err := ResolveOrgRef(ctx, "AcMe")
	require.NoError(t, err)
	assert.Equal(t, created.UUID, byMixedCase.UUID)

	_, err = ResolveOrgRef(ctx, "nope")
	assert.ErrorIs(t, err, ErrOrgNotFound)
}

func TestResolveOrgUUIDEmptyMeansNoFilter(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	got, err := ResolveOrgUUID(context.Background(), "")
	require.NoError(t, err)
	assert.Empty(t, got, "empty ref must resolve to empty so read paths skip the filter")
}

func TestCreateOrgRejectsDuplicateName(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	_, err := CreateOrg(ctx, "acme", "", "", nil)
	require.NoError(t, err)

	_, err = CreateOrg(ctx, "acme", "", "", nil)
	assert.Error(t, err)
}

// seedOrgFixtures creates one workspace with an asset, a vuln and a run.
func seedOrgFixtures(t *testing.T, ctx context.Context, workspace string) {
	t.Helper()

	_, err := db.NewInsert().Model(&Workspace{Name: workspace, OrgUUID: DefaultOrgUUID}).Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewInsert().Model(&Asset{
		Workspace: workspace, AssetValue: workspace + "-asset", OrgUUID: DefaultOrgUUID,
	}).Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewInsert().Model(&Vulnerability{
		Workspace: workspace, VulnTitle: "finding", Severity: "high", OrgUUID: DefaultOrgUUID,
	}).Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewInsert().Model(&Run{
		RunUUID: workspace + "-run", WorkflowName: "w", WorkflowKind: "module",
		Target: workspace, Status: "done", Workspace: workspace, OrgUUID: DefaultOrgUUID,
	}).Exec(ctx)
	require.NoError(t, err)
}

func TestAssignWorkspacesToOrgCascades(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	seedOrgFixtures(t, ctx, "acme.com")
	seedOrgFixtures(t, ctx, "acme.io")
	seedOrgFixtures(t, ctx, "unrelated.com")

	org, err := CreateOrg(ctx, "acme", "", "", nil)
	require.NoError(t, err)

	counts, err := AssignWorkspacesToOrg(ctx, org.UUID, []string{"acme.com", "acme.io"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), counts["workspaces"])
	assert.Equal(t, int64(2), counts["assets"])
	assert.Equal(t, int64(2), counts["vulnerabilities"])
	assert.Equal(t, int64(2), counts["runs"])

	stats, err := GetOrgStats(ctx, org.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.TotalWorkspaces)
	assert.Equal(t, 2, stats.TotalAssets)
	assert.Equal(t, 2, stats.TotalVulns)
	assert.Equal(t, 2, stats.TotalRuns)
	assert.Equal(t, []string{"acme.com", "acme.io"}, stats.Workspaces)

	// The untouched workspace stays in the default org.
	defaultStats, err := GetOrgStats(ctx, DefaultOrgUUID)
	require.NoError(t, err)
	assert.Equal(t, 1, defaultStats.TotalWorkspaces)
	assert.Equal(t, []string{"unrelated.com"}, defaultStats.Workspaces)
}

func TestDeleteOrgReassignsDataByDefault(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	seedOrgFixtures(t, ctx, "acme.com")

	org, err := CreateOrg(ctx, "acme", "", "", nil)
	require.NoError(t, err)
	_, err = AssignWorkspacesToOrg(ctx, org.UUID, []string{"acme.com"})
	require.NoError(t, err)

	require.NoError(t, DeleteOrg(ctx, org.UUID, false))

	_, err = GetOrgByUUID(ctx, org.UUID)
	assert.ErrorIs(t, err, ErrOrgNotFound)

	// Data survives, reattributed to the default org.
	stats, err := GetOrgStats(ctx, DefaultOrgUUID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalWorkspaces)
	assert.Equal(t, 1, stats.TotalAssets)
	assert.Equal(t, 1, stats.TotalVulns)
	assert.Equal(t, 1, stats.TotalRuns)
}

func TestDeleteOrgPurgeRemovesData(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	seedOrgFixtures(t, ctx, "acme.com")
	seedOrgFixtures(t, ctx, "keep.com")

	org, err := CreateOrg(ctx, "acme", "", "", nil)
	require.NoError(t, err)
	_, err = AssignWorkspacesToOrg(ctx, org.UUID, []string{"acme.com"})
	require.NoError(t, err)

	require.NoError(t, DeleteOrg(ctx, org.UUID, true))

	stats, err := GetOrgStats(ctx, DefaultOrgUUID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalWorkspaces)
	assert.Equal(t, []string{"keep.com"}, stats.Workspaces)
	assert.Equal(t, 1, stats.TotalAssets)
	assert.Equal(t, 1, stats.TotalVulns)
	assert.Equal(t, 1, stats.TotalRuns)
}

func TestDefaultOrgCannotBeDeletedOrRenamed(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()

	assert.ErrorIs(t, DeleteOrg(ctx, DefaultOrgUUID, false), ErrDefaultOrgImmutable)
	assert.ErrorIs(t, DeleteOrg(ctx, DefaultOrgUUID, true), ErrDefaultOrgImmutable)

	_, err := UpdateOrg(ctx, DefaultOrgUUID, "renamed", "", nil)
	assert.ErrorIs(t, err, ErrDefaultOrgImmutable)

	// Updating only the description is still allowed.
	updated, err := UpdateOrg(ctx, DefaultOrgUUID, "", "new description", nil)
	require.NoError(t, err)
	assert.Equal(t, "new description", updated.Description)
	assert.Equal(t, DefaultOrgName, updated.Name)
}

// TestBackfillReclaimsEmptyOrgUUID covers the Bun zero-value trap: a write path
// that forgets to set OrgUUID stores an empty string rather than falling back to
// the column
// DEFAULT, stranding the row outside every org.
func TestBackfillReclaimsEmptyOrgUUID(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := db.ExecContext(ctx,
		"INSERT INTO assets (workspace, asset_value, org_uuid) VALUES (?, ?, '')",
		"orphan.com", "orphan-asset")
	require.NoError(t, err)

	var stranded int
	stranded, err = db.NewSelect().Model((*Asset)(nil)).Where("org_uuid = ''").Count(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, stranded)

	require.NoError(t, Migrate(ctx))

	stranded, err = db.NewSelect().Model((*Asset)(nil)).Where("org_uuid = ''").Count(ctx)
	require.NoError(t, err)
	assert.Zero(t, stranded)

	stats, err := GetOrgStats(ctx, DefaultOrgUUID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalAssets)
}

// TestExistingRowsLandInDefaultOrg is the backward-compatibility contract: data
// written before the org column existed must be readable and attributed to the
// default org, not stranded.
func TestExistingRowsLandInDefaultOrg(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Simulate a pre-org database by dropping the column, inserting, then
	// re-running the migration that adds it back. SQLite refuses to drop a column
	// an index still references, so the org indexes go first.
	for _, idx := range []string{
		"idx_workspaces_org_uuid", "idx_assets_org_uuid",
		"idx_vulnerabilities_org_uuid", "idx_runs_org_uuid", "idx_assets_org_workspace",
	} {
		_, err := db.ExecContext(ctx, "DROP INDEX IF EXISTS "+idx)
		require.NoError(t, err, "drop index %s", idx)
	}
	for _, table := range orgScopedTables {
		_, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s DROP COLUMN org_uuid", table))
		require.NoError(t, err, "drop org_uuid from %s", table)
	}

	_, err := db.ExecContext(ctx,
		"INSERT INTO assets (workspace, asset_value) VALUES (?, ?)", "legacy.com", "legacy-asset")
	require.NoError(t, err)

	require.NoError(t, Migrate(ctx))

	stats, err := GetOrgStats(ctx, DefaultOrgUUID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalAssets, "legacy rows must be attributed to the default org")
}

// TestInsertInheritsOrgFromWorkspace covers the BeforeAppendModel hook: an import
// that does not know about orgs still lands its rows in the workspace's org.
func TestInsertInheritsOrgFromWorkspace(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	seedOrgFixtures(t, ctx, "acme.com")

	org, err := CreateOrg(ctx, "acme", "", "", nil)
	require.NoError(t, err)
	_, err = AssignWorkspacesToOrg(ctx, org.UUID, []string{"acme.com"})
	require.NoError(t, err)

	// Simulate an importer that never sets OrgUUID.
	_, err = db.NewInsert().Model(&Asset{
		Workspace: "acme.com", AssetValue: "new.acme.com",
	}).Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewInsert().Model(&Vulnerability{
		Workspace: "acme.com", VulnTitle: "new finding", Severity: "low",
	}).Exec(ctx)
	require.NoError(t, err)

	_, err = db.NewInsert().Model(&Run{
		RunUUID: "new-run", WorkflowName: "w", WorkflowKind: "module",
		Target: "acme.com", Status: "done", Workspace: "acme.com",
	}).Exec(ctx)
	require.NoError(t, err)

	stats, err := GetOrgStats(ctx, org.UUID)
	require.NoError(t, err)
	assert.Equal(t, 2, stats.TotalAssets, "new asset must inherit the workspace's org")
	assert.Equal(t, 2, stats.TotalVulns)
	assert.Equal(t, 2, stats.TotalRuns)
}

// TestInsertWithoutWorkspaceOrgFallsBackToDefault covers a workspace that has no
// row yet - the first import of a scan can land before EnsureWorkspaceRuntime.
func TestInsertWithoutWorkspaceOrgFallsBackToDefault(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()

	_, err := db.NewInsert().Model(&Asset{
		Workspace: "never-registered.com", AssetValue: "a.never-registered.com",
	}).Exec(ctx)
	require.NoError(t, err)

	var got string
	err = db.NewSelect().Model((*Asset)(nil)).Column("org_uuid").
		Where("asset_value = ?", "a.never-registered.com").Scan(ctx, &got)
	require.NoError(t, err)
	assert.Equal(t, DefaultOrgUUID, got)
}

// TestUpdateDoesNotResetOrg is the re-scan case. Import paths build a fresh Asset
// from scan output and UPDATE it WherePK; that struct has an empty OrgUUID meaning
// "not loaded", and blindly writing it would evict the asset from its org.
func TestUpdateDoesNotResetOrg(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	seedOrgFixtures(t, ctx, "acme.com")

	org, err := CreateOrg(ctx, "acme", "", "", nil)
	require.NoError(t, err)
	_, err = AssignWorkspacesToOrg(ctx, org.UUID, []string{"acme.com"})
	require.NoError(t, err)

	var existing Asset
	require.NoError(t, db.NewSelect().Model(&existing).
		Where("asset_value = ?", "acme.com-asset").Scan(ctx))
	require.Equal(t, org.UUID, existing.OrgUUID)

	// A fresh struct carrying only scan output, as the importers build it.
	rescanned := Asset{
		ID:         existing.ID,
		Workspace:  "acme.com",
		AssetValue: "acme.com-asset",
		StatusCode: 200,
		Title:      "updated title",
	}
	_, err = db.NewUpdate().Model(&rescanned).WherePK().Exec(ctx)
	require.NoError(t, err)

	var after Asset
	require.NoError(t, db.NewSelect().Model(&after).Where("id = ?", existing.ID).Scan(ctx))
	assert.Equal(t, org.UUID, after.OrgUUID, "re-scan must not evict the asset from its org")
	assert.Equal(t, "updated title", after.Title)

	stats, err := GetOrgStats(ctx, org.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalAssets)
}

// TestExplicitOrgUUIDIsNeverOverwritten guards the hook's early exit.
func TestExplicitOrgUUIDIsNeverOverwritten(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	seedOrgFixtures(t, ctx, "acme.com")

	org, err := CreateOrg(ctx, "acme", "", "", nil)
	require.NoError(t, err)

	// Workspace is still in the default org, but the caller names another one.
	_, err = db.NewInsert().Model(&Asset{
		Workspace: "acme.com", AssetValue: "explicit.acme.com", OrgUUID: org.UUID,
	}).Exec(ctx)
	require.NoError(t, err)

	var got string
	err = db.NewSelect().Model((*Asset)(nil)).Column("org_uuid").
		Where("asset_value = ?", "explicit.acme.com").Scan(ctx, &got)
	require.NoError(t, err)
	assert.Equal(t, org.UUID, got)
}

// TestEnsureWorkspaceRuntimePreservesOrg is the counterpart for scans: re-running
// a scan without --org must not drag an assigned workspace back to the default org.
func TestEnsureWorkspaceRuntimePreservesOrg(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, EnsureWorkspaceRuntime(ctx, "acme.com", "/tmp/acme", "recon", "", "", "", "", ""))

	org, err := CreateOrg(ctx, "acme", "", "", nil)
	require.NoError(t, err)
	_, err = AssignWorkspacesToOrg(ctx, org.UUID, []string{"acme.com"})
	require.NoError(t, err)

	// A later scan with no --org.
	require.NoError(t, EnsureWorkspaceRuntime(ctx, "acme.com", "/tmp/acme", "recon2", "", "", "", "", ""))

	stats, err := GetOrgStats(ctx, org.UUID)
	require.NoError(t, err)
	assert.Equal(t, 1, stats.TotalWorkspaces, "re-scan without --org must not clear the org")

	// A scan that names a different org does move it.
	require.NoError(t, EnsureWorkspaceRuntime(
		ctx, "acme.com", "/tmp/acme", "recon3", "", "", "", "", DefaultOrgUUID))

	stats, err = GetOrgStats(ctx, org.UUID)
	require.NoError(t, err)
	assert.Zero(t, stats.TotalWorkspaces, "an explicit --org must move the workspace")
}

func TestOrgScopedTablesAllHaveColumn(t *testing.T) {
	cleanup := setupOrgTestDB(t)
	defer cleanup()

	ctx := context.Background()
	for _, table := range orgScopedTables {
		var n int
		err := db.NewRaw(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE org_uuid = ?", table), DefaultOrgUUID).
			Scan(ctx, &n)
		require.NoErrorf(t, err, "table %s is missing org_uuid", table)
	}
	assert.Equal(t, []string{"workspaces", "assets", "vulnerabilities", "runs"}, orgScopedTables)
	assert.True(t, strings.HasPrefix(DefaultOrgUUID, "00000000-"))
}
