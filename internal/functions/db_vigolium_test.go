package functions

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// A single finding record can embed a multi-MB response body, producing a JSONL
// line longer than bufio.Scanner's max token size. The reader-based import must
// stream these without "token too long".
func TestDbImportVigolium_OversizedLine(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// 12MB response body — comfortably past the old 10MB scanner buffer.
	bigBody := strings.Repeat("A", 12*1024*1024)
	line := `{"type":"finding","data":{` +
		`"finding_hash":"oversized-line-hash",` +
		`"module_id":"big-response",` +
		`"module_name":"Big Response",` +
		`"severity":"high",` +
		`"url":"https://example.com/big",` +
		`"response":"` + bigBody + `"}}`

	testFile := filepath.Join(t.TempDir(), "vigolium-oversized.jsonl")
	require.NoError(t, os.WriteFile(testFile, []byte(line+"\n"), 0o644))

	result, err := registry.Execute(
		`db_import_vigolium("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a nested stats map, got: %v", result)

	vulnStats := stats["vulns"].(map[string]interface{})
	assert.Equal(t, 1, vulnStats["new"], "oversized finding should import")
	assert.Equal(t, 0, vulnStats["errors"])
	assert.Equal(t, 1, stats["total"])

	ctx := context.Background()
	db := database.GetDB()
	var vuln database.Vulnerability
	err = db.NewSelect().Model(&vuln).
		Where("finding_hash = ?", "oversized-line-hash").Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "big-response", vuln.VulnInfo)
	assert.Len(t, vuln.DetailHTTPResponse, 12*1024*1024)
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

const vigoliumV057Fixture = "../../test/testdata/sample-jsonl-output/vigolium-v057-output.jsonl"

// importVigolium runs db_import_vigolium and returns its nested stats.
func importVigolium(t *testing.T, workspace, file string) (assets, vulns map[string]interface{}, stats map[string]interface{}) {
	t.Helper()
	result, err := NewRegistry().Execute(
		`db_import_vigolium("`+workspace+`", "`+file+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats = result.(map[string]interface{})
	return stats["assets"].(map[string]interface{}), stats["vulns"].(map[string]interface{}), stats
}

// TestDbImportVigolium_V057 covers the fields vigolium v0.5.7 exports that the
// importer used to drop: technology and response_time_ms on http_record, and
// cwe_id, cvss_score, remediation, tags and vuln_class on finding. Also pins
// the new "scope" record type and the "suspect" severity.
func TestDbImportVigolium_V057(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	assetStats, vulnStats, stats := importVigolium(t, "ws", vigoliumV057Fixture)

	// 2 http_record, 2 finding; scan + oast_interaction + scope are skipped.
	assert.Equal(t, 2, assetStats["new"])
	assert.Equal(t, 0, assetStats["errors"])
	assert.Equal(t, 2, vulnStats["new"])
	assert.Equal(t, 0, vulnStats["errors"])
	assert.Equal(t, 3, stats["skipped"])
	assert.Equal(t, 7, stats["total"])

	ctx := context.Background()
	db := database.GetDB()

	// technology -> Technologies, response_time_ms -> ResponseTime
	var withTech database.Asset
	require.NoError(t, db.NewSelect().Model(&withTech).
		Where("workspace = ?", "ws").Where("url = ?", "http://legacy.example.com/").Scan(ctx))
	assert.Equal(t, []string{"apache/2.2", "php/5.6"}, withTech.Technologies)
	assert.Equal(t, "350ms", withTech.ResponseTime)
	assert.Equal(t, "Legacy Portal", withTech.Title)

	var noTech database.Asset
	require.NoError(t, db.NewSelect().Model(&noTech).
		Where("workspace = ?", "ws").Where("url = ?", "https://api.shop.local/api/v1/admin/config").Scan(ctx))
	assert.Empty(t, noTech.Technologies, "null technology stays empty")
	assert.Equal(t, "4ms", noTech.ResponseTime)

	// Classification fields, vuln_class and vigolium's own tags
	var ssrf database.Vulnerability
	require.NoError(t, db.NewSelect().Model(&ssrf).
		Where("finding_hash = ?", "b469638a3a72a81568e826e04479e490").Scan(ctx))
	assert.Equal(t, "CWE-918", ssrf.CWEID)
	assert.InDelta(t, 9.1, ssrf.CVSSScore, 1e-9)
	assert.NotEmpty(t, ssrf.Remediation)
	assert.Equal(t, "critical", ssrf.Severity)
	assert.Contains(t, ssrf.Tags, "cloud-metadata", "vigolium's own tags are imported")
	assert.Contains(t, ssrf.Tags, "ssrf", "vuln_class is imported as a tag")
	seen := map[string]int{}
	for _, tag := range ssrf.Tags {
		seen[tag]++
	}
	assert.Equal(t, 1, seen["ssrf"], "vuln_class and an identical vigolium tag collapse to one")

	// suspect severity is stored as-is; a null cwe_id and zero cvss stay empty
	var suspect database.Vulnerability
	require.NoError(t, db.NewSelect().Model(&suspect).
		Where("finding_hash = ?", "2cf3e2b2fda30e3360476ecf40c1734c").Scan(ctx))
	assert.Equal(t, "suspect", suspect.Severity)
	assert.Equal(t, "tentative", suspect.Confidence)
	assert.Empty(t, suspect.CWEID)
	assert.Zero(t, suspect.CVSSScore)
	assert.Empty(t, suspect.Remediation)
}

// TestDbImportVigolium_V057Idempotent: a second import of the same file changes
// nothing — the new tags must not read as a change on every pass.
func TestDbImportVigolium_V057Idempotent(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	importVigolium(t, "ws", vigoliumV057Fixture)
	assetStats, vulnStats, _ := importVigolium(t, "ws", vigoliumV057Fixture)

	assert.Equal(t, 0, assetStats["new"])
	assert.Equal(t, 0, vulnStats["new"])
	assert.Equal(t, 0, vulnStats["updated"])
	assert.Equal(t, 2, vulnStats["unchanged"])
}

// writeStrippedFixture writes the v0.5.7 fixture with the classification
// fields removed from every finding, as an older vigolium export looks.
func writeStrippedFixture(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile(vigoliumV057Fixture)
	require.NoError(t, err)
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var rec map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(line), &rec))
		if rec["type"] == "finding" {
			data := rec["data"].(map[string]interface{})
			for _, k := range []string{"cwe_id", "cvss_score", "remediation", "vuln_class", "tags"} {
				delete(data, k)
			}
		}
		b, err := json.Marshal(rec)
		require.NoError(t, err)
		out = append(out, string(b))
	}
	path := filepath.Join(t.TempDir(), "vigolium-stripped.jsonl")
	require.NoError(t, os.WriteFile(path, []byte(strings.Join(out, "\n")+"\n"), 0o644))
	return path
}

// TestDbImportVigolium_ClassificationBackfill: a row imported before the
// classification columns existed gains them on the next full import, instead
// of being judged unchanged and left blank.
func TestDbImportVigolium_ClassificationBackfill(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	importVigolium(t, "ws", writeStrippedFixture(t))
	_, vulnStats, _ := importVigolium(t, "ws", vigoliumV057Fixture)
	assert.Equal(t, 0, vulnStats["new"])
	assert.GreaterOrEqual(t, vulnStats["updated"], 1, "the finding that gained CWE/CVSS must be updated")

	var ssrf database.Vulnerability
	require.NoError(t, database.GetDB().NewSelect().Model(&ssrf).
		Where("finding_hash = ?", "b469638a3a72a81568e826e04479e490").Scan(context.Background()))
	assert.Equal(t, "CWE-918", ssrf.CWEID)
	assert.InDelta(t, 9.1, ssrf.CVSSScore, 1e-9)
}

// TestDbImportVigolium_ClassificationSurvivesSparseReimport: a later import
// that lacks the classification fields must not blank the stored ones.
func TestDbImportVigolium_ClassificationSurvivesSparseReimport(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	importVigolium(t, "ws", vigoliumV057Fixture)
	importVigolium(t, "ws", writeStrippedFixture(t))

	var ssrf database.Vulnerability
	require.NoError(t, database.GetDB().NewSelect().Model(&ssrf).
		Where("finding_hash = ?", "b469638a3a72a81568e826e04479e490").Scan(context.Background()))
	assert.Equal(t, "CWE-918", ssrf.CWEID)
	assert.InDelta(t, 9.1, ssrf.CVSSScore, 1e-9)
	assert.NotEmpty(t, ssrf.Remediation)
}
