package database

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDiffTestDB(t *testing.T) func() {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_diff.sqlite")
	cfg := &config.Config{
		BaseFolder: tmpDir,
		Database: config.DatabaseConfig{
			DBEngine: "sqlite",
			DBPath:   dbPath,
		},
	}

	_, err := Connect(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, Migrate(ctx))

	return func() {
		_ = Close()
		SetDB(nil)
	}
}

func TestGetAssetDiff_Added(t *testing.T) {
	cleanup := setupDiffTestDB(t)
	defer cleanup()

	ctx := context.Background()
	workspace := "test-workspace"
	now := time.Now()
	fromTime := now.Add(-1 * time.Hour)
	toTime := now

	// Create a baseline asset (before fromTime)
	oldAsset := &Asset{
		Workspace:  workspace,
		AssetValue: "old.example.com",
		URL:        "https://old.example.com",
		StatusCode: 200,
		CreatedAt:  fromTime.Add(-2 * time.Hour),
		UpdatedAt:  fromTime.Add(-2 * time.Hour),
		LastSeenAt: fromTime.Add(-1 * time.Hour),
	}
	_, err := GetDB().NewInsert().Model(oldAsset).Exec(ctx)
	require.NoError(t, err)

	// Create a new asset (after fromTime)
	newAsset := &Asset{
		Workspace:  workspace,
		AssetValue: "new.example.com",
		URL:        "https://new.example.com",
		StatusCode: 200,
		CreatedAt:  now.Add(-30 * time.Minute),
		UpdatedAt:  now.Add(-30 * time.Minute),
		LastSeenAt: now,
	}
	_, err = GetDB().NewInsert().Model(newAsset).Exec(ctx)
	require.NoError(t, err)

	// Get diff
	diff, err := GetAssetDiff(ctx, workspace, fromTime, toTime)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, workspace, diff.WorkspaceName)
	assert.Equal(t, 1, diff.Summary.TotalAdded)
	assert.Len(t, diff.Added, 1)
	assert.Equal(t, "new.example.com", diff.Added[0].AssetValue)
}

func TestGetAssetDiff_Removed(t *testing.T) {
	cleanup := setupDiffTestDB(t)
	defer cleanup()

	ctx := context.Background()
	workspace := "test-workspace"
	now := time.Now()
	fromTime := now.Add(-1 * time.Hour)
	toTime := now

	// Create an old asset that hasn't been seen since before fromTime
	oldAsset := &Asset{
		Workspace:  workspace,
		AssetValue: "stale.example.com",
		URL:        "https://stale.example.com",
		StatusCode: 200,
		CreatedAt:  fromTime.Add(-48 * time.Hour),
		UpdatedAt:  fromTime.Add(-48 * time.Hour),
		LastSeenAt: fromTime.Add(-24 * time.Hour), // Last seen 24 hours before fromTime
	}
	_, err := GetDB().NewInsert().Model(oldAsset).Exec(ctx)
	require.NoError(t, err)

	// Get diff
	diff, err := GetAssetDiff(ctx, workspace, fromTime, toTime)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, 1, diff.Summary.TotalRemoved)
	assert.Len(t, diff.Removed, 1)
	assert.Equal(t, "stale.example.com", diff.Removed[0].AssetValue)
}

func TestGetAssetDiff_Changed(t *testing.T) {
	cleanup := setupDiffTestDB(t)
	defer cleanup()

	ctx := context.Background()
	workspace := "test-workspace"
	now := time.Now()
	fromTime := now.Add(-1 * time.Hour)
	toTime := now

	// Create an asset that was created before fromTime but updated within the range
	changedAsset := &Asset{
		Workspace:  workspace,
		AssetValue: "changed.example.com",
		URL:        "https://changed.example.com",
		StatusCode: 200,
		CreatedAt:  fromTime.Add(-2 * time.Hour), // Created before fromTime
		UpdatedAt:  now.Add(-30 * time.Minute),   // Updated within range
		LastSeenAt: now,
	}
	_, err := GetDB().NewInsert().Model(changedAsset).Exec(ctx)
	require.NoError(t, err)

	// Get diff
	diff, err := GetAssetDiff(ctx, workspace, fromTime, toTime)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, 1, diff.Summary.TotalChanged)
	assert.Len(t, diff.Changed, 1)
	assert.Equal(t, "changed.example.com", diff.Changed[0].AssetValue)
}

func TestGetAssetDiff_Empty(t *testing.T) {
	cleanup := setupDiffTestDB(t)
	defer cleanup()

	ctx := context.Background()
	workspace := "empty-workspace"
	now := time.Now()
	fromTime := now.Add(-1 * time.Hour)
	toTime := now

	// Get diff for empty workspace
	diff, err := GetAssetDiff(ctx, workspace, fromTime, toTime)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, workspace, diff.WorkspaceName)
	assert.Equal(t, 0, diff.Summary.TotalAdded)
	assert.Equal(t, 0, diff.Summary.TotalRemoved)
	assert.Equal(t, 0, diff.Summary.TotalChanged)
	assert.Empty(t, diff.Added)
	assert.Empty(t, diff.Removed)
	assert.Empty(t, diff.Changed)
}

func TestGetAssetDiff_NoDatabase(t *testing.T) {
	// Ensure db is nil
	originalDB := db
	db = nil
	defer func() { db = originalDB }()

	ctx := context.Background()
	diff, err := GetAssetDiff(ctx, "test", time.Now().Add(-1*time.Hour), time.Now())

	assert.Error(t, err)
	assert.Nil(t, diff)
	assert.Contains(t, err.Error(), "database not connected")
}

func TestGetVulnerabilityDiff_Added(t *testing.T) {
	cleanup := setupDiffTestDB(t)
	defer cleanup()

	ctx := context.Background()
	workspace := "test-workspace"
	now := time.Now()
	fromTime := now.Add(-1 * time.Hour)
	toTime := now

	// Create a new vulnerability (after fromTime)
	newVuln := &Vulnerability{
		Workspace:  workspace,
		VulnInfo:   "CVE-2024-1234",
		VulnTitle:  "Test Vulnerability",
		Severity:   "high",
		AssetValue: "vulnerable.example.com",
		CreatedAt:  now.Add(-30 * time.Minute),
		UpdatedAt:  now.Add(-30 * time.Minute),
		LastSeenAt: now,
	}
	_, err := GetDB().NewInsert().Model(newVuln).Exec(ctx)
	require.NoError(t, err)

	// Get diff
	diff, err := GetVulnerabilityDiff(ctx, workspace, fromTime, toTime)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, workspace, diff.WorkspaceName)
	assert.Equal(t, 1, diff.Summary.TotalAdded)
	assert.Len(t, diff.Added, 1)
	assert.Equal(t, "CVE-2024-1234", diff.Added[0].VulnInfo)
}

func TestImportStats(t *testing.T) {
	stats := ImportStats{
		New:       5,
		Updated:   3,
		Unchanged: 10,
		Errors:    2,
	}

	assert.Equal(t, 5, stats.New)
	assert.Equal(t, 3, stats.Updated)
	assert.Equal(t, 10, stats.Unchanged)
	assert.Equal(t, 2, stats.Errors)
}
