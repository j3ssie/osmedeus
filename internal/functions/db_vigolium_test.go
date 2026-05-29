package functions

import (
	"context"
	"os"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDbImportVigolium(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/vigolium-juice-shop-output.jsonl"
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample vigolium jsonl file must exist")

	result, err := registry.Execute(
		`db_import_vigolium("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a nested stats map")

	assetStats, ok := stats["assets"].(map[string]interface{})
	require.True(t, ok, "assets stats should be a map")
	vulnStats, ok := stats["vulns"].(map[string]interface{})
	require.True(t, ok, "vulns stats should be a map")

	// 179 http_record -> assets (all distinct urls), 146 finding -> vulns,
	// 45 oast_interaction + 4 scan = 49 skipped, 374 total.
	assert.Equal(t, 179, assetStats["new"])
	assert.Equal(t, 0, assetStats["updated"])
	assert.Equal(t, 0, assetStats["errors"])
	assert.Equal(t, 146, vulnStats["new"])
	assert.Equal(t, 0, vulnStats["errors"])
	assert.Equal(t, 49, stats["skipped"])
	assert.Equal(t, 374, stats["total"])

	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	// Assets landed in the assets table
	assetCount, err := db.NewSelect().Model((*database.Asset)(nil)).
		Where("workspace = ?", "test-workspace").Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 179, assetCount)

	// Findings landed in the vulnerabilities table
	vulnCount, err := db.NewSelect().Model((*database.Vulnerability)(nil)).
		Where("workspace = ?", "test-workspace").Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 146, vulnCount)

	// Spot-check a finding -> vulnerability field mapping
	var vuln database.Vulnerability
	err = db.NewSelect().Model(&vuln).
		Where("finding_hash = ?", "907c7c2ec0eff89214a3698185032ce6c9305403").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "input-reflection-detect", vuln.VulnInfo)
	assert.Equal(t, "Input Reflection Detect", vuln.VulnTitle)
	assert.Equal(t, "info", vuln.Severity)
	assert.Equal(t, "tentative", vuln.Confidence)
	assert.Equal(t, "https://ginandjuice.shop/about/", vuln.AssetValue)
	assert.Equal(t, "http", vuln.AssetType)
	assert.Contains(t, vuln.Tags, "passive")
	assert.NotEmpty(t, vuln.RawVulnJSON)
}

// Re-importing the same file must not create duplicates: findings dedupe on
// finding_hash and assets dedupe on (workspace, asset_value, url).
func TestDbImportVigolium_Idempotent(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()
	testFile := "../../test/testdata/sample-jsonl-output/vigolium-juice-shop-output.jsonl"
	_, err := os.Stat(testFile)
	require.NoError(t, err)

	call := `db_import_vigolium("test-workspace", "` + testFile + `")`

	_, err = registry.Execute(call, map[string]interface{}{})
	require.NoError(t, err)

	result, err := registry.Execute(call, map[string]interface{}{})
	require.NoError(t, err)

	stats := result.(map[string]interface{})
	assetStats := stats["assets"].(map[string]interface{})
	vulnStats := stats["vulns"].(map[string]interface{})

	// Second pass: everything matches existing rows, nothing new.
	assert.Equal(t, 0, assetStats["new"])
	assert.Equal(t, 179, assetStats["updated"])
	assert.Equal(t, 0, vulnStats["new"])
	assert.Equal(t, 146, vulnStats["unchanged"])

	ctx := context.Background()
	db := database.GetDB()

	assetCount, err := db.NewSelect().Model((*database.Asset)(nil)).
		Where("workspace = ?", "test-workspace").Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 179, assetCount, "no duplicate assets after re-import")

	vulnCount, err := db.NewSelect().Model((*database.Vulnerability)(nil)).
		Where("workspace = ?", "test-workspace").Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 146, vulnCount, "no duplicate vulns after re-import")
}

func TestDbImportVigolium_EmptyArgs(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(`db_import_vigolium("", "/tmp/test.jsonl")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")

	result, err = registry.Execute(`db_import_vigolium("ws", "")`, map[string]interface{}{})
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}
