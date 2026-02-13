package functions

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB initializes a test SQLite database
func setupTestDB(t *testing.T) (cleanup func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")
	cfg := &config.Config{
		BaseFolder: tmpDir,
		Database: config.DatabaseConfig{
			DBEngine: "sqlite",
			DBPath:   dbPath,
		},
	}

	_, err := database.Connect(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx))

	return func() {
		_ = database.Close()
		database.SetDB(nil)
	}
}

func TestDbImportAssetFromFile(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Use the sample httpx data file
	testFile := "../../test/testdata/sample-jsonl-output/http-data.jsonl"

	// Check if file exists
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample http-data.jsonl file must exist")

	result, err := registry.Execute(
		`db_import_asset_from_file("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	// http-data.jsonl has 3 lines - result is now a map with stats
	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, 3, stats["new"])
	assert.Equal(t, 3, stats["total"])

	// Verify assets were imported
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var assets []database.Asset
	err = db.NewSelect().Model(&assets).
		Where("workspace = ?", "test-workspace").
		Scan(ctx)
	require.NoError(t, err)
	assert.Len(t, assets, 3)

	// Check specific fields were mapped correctly
	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("asset_value = ?", "api.hackerone.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "http://api.hackerone.com", asset.URL)
	assert.Equal(t, "HackerOne API", asset.Title)
	assert.Equal(t, 200, asset.StatusCode)
	assert.Equal(t, "text/html", asset.ContentType)
	assert.Contains(t, asset.Technologies, "Cloudflare")
	assert.NotEmpty(t, asset.RawJsonData)

	// Check webserver and cdn_name are in remarks
	assert.Contains(t, asset.Remarks, "cloudflare") // webserver
	assert.Contains(t, asset.Remarks, "cloudflare") // cdn_name (same value here)

	// Check gslink asset has both nginx and cloudfront in remarks
	var gsAsset database.Asset
	err = db.NewSelect().Model(&gsAsset).
		Where("asset_value = ?", "gslink.hackerone.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Contains(t, gsAsset.Remarks, "nginx")
	assert.Contains(t, gsAsset.Remarks, "cloudfront")
}

func TestDbImportAssetFromFile_EmptyWorkspace(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_import_asset_from_file("", "/tmp/test.jsonl")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbImportAssetFromFile_NonExistentFile(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_import_asset_from_file("test-workspace", "/nonexistent/file.jsonl")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbImportVuln(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Single nuclei-style JSON
	jsonData := `{"template-id":"dns-saas-service-detection","info":{"name":"DNS SaaS Service Detection","severity":"info","tags":["dns","service","discovery"],"description":"A CNAME DNS record was discovered"},"type":"dns","host":"support.hackerone.com","matched-at":"support.hackerone.com"}`

	result, err := registry.Execute(
		"db_import_vuln(\"test-workspace\", '"+jsonData+"')",
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify vulnerability was imported
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var vuln database.Vulnerability
	err = db.NewSelect().Model(&vuln).
		Where("workspace = ?", "test-workspace").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "dns-saas-service-detection", vuln.VulnInfo)
	assert.Equal(t, "DNS SaaS Service Detection", vuln.VulnTitle)
	assert.Equal(t, "info", vuln.Severity)
	assert.Equal(t, "support.hackerone.com", vuln.AssetValue)
	assert.Equal(t, "dns", vuln.AssetType)
	assert.Contains(t, vuln.Tags, "dns")
	assert.Contains(t, vuln.Tags, "service")
	assert.Contains(t, vuln.Tags, "discovery")
}

func TestDbImportVuln_InvalidJSON(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_import_vuln("test-workspace", "not valid json")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbImportVulnFromFile(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Use the sample nuclei data file
	testFile := "../../test/testdata/sample-jsonl-output/vuln-data.jsonl"

	// Check if file exists
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample vuln-data.jsonl file must exist")

	result, err := registry.Execute(
		`db_import_vuln_from_file("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	// vuln-data.jsonl has 13 lines - result is now a map with stats
	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, 13, stats["new"])
	assert.Equal(t, 13, stats["total"])

	// Verify vulnerabilities were imported
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var vulns []database.Vulnerability
	err = db.NewSelect().Model(&vulns).
		Where("workspace = ?", "test-workspace").
		Scan(ctx)
	require.NoError(t, err)
	assert.Len(t, vulns, 13)

	// Check that different vulnerability types were imported
	var severities []string
	for _, v := range vulns {
		severities = append(severities, v.Severity)
	}
	assert.Contains(t, severities, "info")
}

func TestDbImportVulnFromFile_EmptyWorkspace(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_import_vuln_from_file("", "/tmp/test.jsonl")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbImportVulnFromFile_NonExistentFile(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_import_vuln_from_file("test-workspace", "/nonexistent/file.jsonl")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbImportAssetFromFile_Upsert(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Create a temp file with duplicate entries
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "assets.jsonl")
	content := `{"host":"example.com","url":"http://example.com","title":"First","status_code":200}
{"host":"example.com","url":"http://example.com","title":"Updated","status_code":301}
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	result, err := registry.Execute(
		`db_import_asset_from_file("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	// Result is now a map with stats - 1 new, 1 updated (same asset with different data)
	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, 2, stats["total"]) // Both lines processed

	// Verify only one asset exists (upsert)
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var assets []database.Asset
	err = db.NewSelect().Model(&assets).
		Where("workspace = ?", "test-workspace").
		Scan(ctx)
	require.NoError(t, err)
	assert.Len(t, assets, 1)

	// Should have the updated values
	assert.Equal(t, "Updated", assets[0].Title)
	assert.Equal(t, 301, assets[0].StatusCode)
}

func TestMapJSONToAsset(t *testing.T) {
	data := map[string]interface{}{
		"host":           "example.com",
		"url":            "http://example.com",
		"input":          "https://example.com",
		"scheme":         "http",
		"method":         "GET",
		"path":           "/",
		"status_code":    float64(200),
		"content_type":   "text/html",
		"content_length": float64(1234),
		"title":          "Example",
		"words":          float64(100),
		"lines":          float64(50),
		"host_ip":        "1.2.3.4",
		"a":              []interface{}{"1.2.3.4", "5.6.7.8"},
		"tech":           []interface{}{"Nginx", "PHP"},
		"time":           "123ms",
		"webserver":      "nginx",
	}

	asset := mapJSONToAsset(data, "test-workspace", `{"host":"example.com"}`)

	assert.Equal(t, "test-workspace", asset.Workspace)
	assert.Equal(t, "example.com", asset.AssetValue)
	assert.Equal(t, "http://example.com", asset.URL)
	assert.Equal(t, "https://example.com", asset.Input)
	assert.Equal(t, "http", asset.Scheme)
	assert.Equal(t, "GET", asset.Method)
	assert.Equal(t, "/", asset.Path)
	assert.Equal(t, 200, asset.StatusCode)
	assert.Equal(t, "text/html", asset.ContentType)
	assert.Equal(t, int64(1234), asset.ContentLength)
	assert.Equal(t, "Example", asset.Title)
	assert.Equal(t, 100, asset.Words)
	assert.Equal(t, 50, asset.Lines)
	assert.Equal(t, "1.2.3.4", asset.HostIP)
	assert.Equal(t, []string{"1.2.3.4", "5.6.7.8"}, asset.DnsRecords)
	assert.Equal(t, []string{"Nginx", "PHP"}, asset.Technologies)
	assert.Equal(t, "123ms", asset.ResponseTime)
	assert.Equal(t, "", asset.Source) // webserver no longer mapped to source
	assert.Contains(t, asset.Remarks, "nginx") // webserver now in remarks
	assert.NotEmpty(t, asset.RawJsonData)
}

func TestMapJSONToVuln(t *testing.T) {
	data := map[string]interface{}{
		"template-id": "test-vuln",
		"info": map[string]interface{}{
			"name":        "Test Vulnerability",
			"description": "A test vulnerability",
			"severity":    "high",
			"tags":        []interface{}{"tag1", "tag2"},
		},
		"host":       "example.com",
		"matched-at": "http://example.com/path",
		"type":       "http",
		"request":    "GET / HTTP/1.1",
		"response":   "HTTP/1.1 200 OK",
	}

	vuln := mapJSONToVuln(data, "test-workspace", `{"template-id":"test-vuln"}`)

	assert.Equal(t, "test-workspace", vuln.Workspace)
	assert.Equal(t, "test-vuln", vuln.VulnInfo)
	assert.Equal(t, "Test Vulnerability", vuln.VulnTitle)
	assert.Equal(t, "A test vulnerability", vuln.VulnDesc)
	assert.Equal(t, "high", vuln.Severity)
	assert.Equal(t, "example.com", vuln.AssetValue) // host takes precedence
	assert.Equal(t, "http", vuln.AssetType)
	assert.Equal(t, []string{"tag1", "tag2"}, vuln.Tags)
	assert.Equal(t, "GET / HTTP/1.1", vuln.DetailHTTPRequest)
	assert.Equal(t, "HTTP/1.1 200 OK", vuln.DetailHTTPResponse)
	assert.NotEmpty(t, vuln.RawVulnJSON)
}

func TestMapJSONToVuln_MatchedAtFallback(t *testing.T) {
	data := map[string]interface{}{
		"template-id": "test-vuln",
		"info": map[string]interface{}{
			"name":     "Test",
			"severity": "low",
		},
		"matched-at": "http://example.com/path",
		"type":       "http",
	}

	vuln := mapJSONToVuln(data, "test-workspace", `{}`)

	// When host is not present, matched-at is used as fallback
	assert.Equal(t, "http://example.com/path", vuln.AssetValue)
}

func TestDbPartialImportAsset(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_partial_import_asset("test-workspace", "domain", "sub.example.com")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify asset was imported with correct fields
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ? AND asset_value = ?", "test-workspace", "sub.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "domain", asset.AssetType)
	assert.Equal(t, "test-workspace", asset.Workspace)
	assert.Equal(t, "sub.example.com", asset.AssetValue)
}

func TestDbPartialImportAsset_Upsert(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Import first time
	_, err := registry.Execute(
		`db_partial_import_asset("test-workspace", "domain", "sub.example.com")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	// Import same asset again with different type
	result, err := registry.Execute(
		`db_partial_import_asset("test-workspace", "subdomain", "sub.example.com")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify only 1 row exists
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	count, err := db.NewSelect().Model((*database.Asset)(nil)).
		Where("workspace = ? AND asset_value = ?", "test-workspace", "sub.example.com").
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify type was updated
	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ? AND asset_value = ?", "test-workspace", "sub.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "subdomain", asset.AssetType)
}

func TestDbPartialImportAsset_MissingArgs(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_partial_import_asset("test-workspace", "domain")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbPartialImportAssetFile(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Create a temp file with 3 lines
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "domains.txt")
	content := "sub1.example.com\nsub2.example.com\nsub3.example.com\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	result, err := registry.Execute(
		`db_partial_import_asset_file("test-workspace", "domain", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result)

	// Verify all 3 assets are in the DB
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var assets []database.Asset
	err = db.NewSelect().Model(&assets).
		Where("workspace = ?", "test-workspace").
		Scan(ctx)
	require.NoError(t, err)
	assert.Len(t, assets, 3)

	// All should have the specified asset_type
	for _, a := range assets {
		assert.Equal(t, "domain", a.AssetType)
	}
}

func TestDbPartialImportAssetFile_SkipsEmptyLines(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Create a temp file with blank lines
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "domains.txt")
	content := "sub1.example.com\n\n  \nsub2.example.com\n\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	result, err := registry.Execute(
		`db_partial_import_asset_file("test-workspace", "domain", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result)
}

func TestDbPartialImportAssetFile_NonExistentFile(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_partial_import_asset_file("test-workspace", "domain", "/nonexistent/file.txt")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbQuickImportAsset(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Import a domain without specifying asset_type (should auto-classify)
	result, err := registry.Execute(
		`db_quick_import_asset("test-workspace", "sub.example.com")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify asset was imported
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ? AND asset_value = ?", "test-workspace", "sub.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "domain", asset.AssetType)
	assert.Equal(t, "test-workspace", asset.Workspace)

	// Verify event log was created for new asset
	var eventLog database.EventLog
	err = db.NewSelect().Model(&eventLog).
		Where("topic = ? AND workspace = ?", "db.new.asset", "test-workspace").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "db.new.asset", eventLog.Topic)
	assert.Equal(t, "db_quick_import_asset", eventLog.Source)
	assert.Equal(t, "function", eventLog.SourceType)
	assert.Contains(t, eventLog.Data, "sub.example.com")
}

func TestDbQuickImportAsset_WithType(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Import an IP with explicit asset_type
	result, err := registry.Execute(
		`db_quick_import_asset("test-workspace", "192.168.1.1", "ip")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify asset was imported with specified type
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ? AND asset_value = ?", "test-workspace", "192.168.1.1").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ip", asset.AssetType)
}

func TestDbQuickImportAsset_NoEventOnUpdate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Import asset first time
	_, err := registry.Execute(
		`db_quick_import_asset("test-workspace", "existing.example.com")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	// Count event logs
	count, err := db.NewSelect().Model((*database.EventLog)(nil)).
		Where("topic = ? AND workspace = ?", "db.new.asset", "test-workspace").
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Import same asset again (update)
	result, err := registry.Execute(
		`db_quick_import_asset("test-workspace", "existing.example.com")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Verify no new event log was created (still 1)
	count, err = db.NewSelect().Model((*database.EventLog)(nil)).
		Where("topic = ? AND workspace = ?", "db.new.asset", "test-workspace").
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "should not create event log on update")
}

func TestDbQuickImportAsset_MissingArgs(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_quick_import_asset("test-workspace")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbResetEventLogs(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	// Create some test event logs
	now := time.Now()
	events := []database.EventLog{
		{Topic: "db.new.asset", Workspace: "workspace1", Processed: true, ProcessedAt: &now},
		{Topic: "db.updated.asset", Workspace: "workspace1", Processed: true, ProcessedAt: &now},
		{Topic: "run.started", Workspace: "workspace1", Processed: true, ProcessedAt: &now},
		{Topic: "db.new.asset", Workspace: "workspace2", Processed: true, ProcessedAt: &now},
		{Topic: "run.completed", Workspace: "workspace2", Processed: false}, // Not processed
	}
	for _, e := range events {
		_, err := db.NewInsert().Model(&e).Exec(ctx)
		require.NoError(t, err)
	}

	registry := NewRegistry()

	// Test 1: Reset all event logs (no filters)
	result, err := registry.Execute(
		`db_reset_event_logs()`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats := result.(map[string]interface{})
	assert.Equal(t, int64(4), stats["reset"]) // 4 were processed=true
	assert.Equal(t, 4, stats["total"])

	// Verify all are now unprocessed
	count, err := db.NewSelect().Model((*database.EventLog)(nil)).
		Where("processed = ?", false).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, count)

	// Reset for next test - mark some as processed again
	_, err = db.NewUpdate().Model((*database.EventLog)(nil)).
		Set("processed = ?", true).
		Set("processed_at = ?", now).
		Where("topic LIKE ?", "db.%").
		Exec(ctx)
	require.NoError(t, err)

	// Test 2: Reset with workspace filter
	result, err = registry.Execute(
		`db_reset_event_logs("workspace1")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats = result.(map[string]interface{})
	assert.Equal(t, int64(2), stats["reset"]) // workspace1 has 2 db.* events that are processed

	// Test 3: Reset with topic pattern filter
	// First, mark events processed again for workspace2
	_, err = db.NewUpdate().Model((*database.EventLog)(nil)).
		Set("processed = ?", true).
		Set("processed_at = ?", now).
		Where("workspace = ?", "workspace2").
		Where("topic LIKE ?", "db.%").
		Exec(ctx)
	require.NoError(t, err)

	result, err = registry.Execute(
		`db_reset_event_logs("", "db.*")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats = result.(map[string]interface{})
	assert.Equal(t, int64(1), stats["reset"]) // Only workspace2's db.new.asset is processed

	// Test 4: Reset with both filters
	// Mark all as processed first
	_, err = db.NewUpdate().Model((*database.EventLog)(nil)).
		Set("processed = ?", true).
		Set("processed_at = ?", now).
		Where("1 = 1").
		Exec(ctx)
	require.NoError(t, err)

	result, err = registry.Execute(
		`db_reset_event_logs("workspace1", "run.*")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats = result.(map[string]interface{})
	assert.Equal(t, int64(1), stats["reset"]) // workspace1 has 1 run.* event
	assert.Equal(t, 1, stats["total"])
}

func TestGlobToSQLLike(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
	}{
		{"db.*", "db.%"},
		{"run.*", "run.%"},
		{"*", "%"},
		{"db.?.asset", "db._.asset"},
		{"test%pattern", "test\\%pattern"},
		{"test_pattern", "test\\_pattern"},
		{"*.asset*", "%.asset%"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := globToSQLLike(tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- DNS Asset Import Tests ---

func TestDbImportDNSAsset(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/dns-records.txt"
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample dns-records.txt file must exist")

	result, err := registry.Execute(
		`db_import_dns_asset("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	// 3 unique domains: example.com, sub.example.com, other.example.com
	assert.Equal(t, int64(3), result)

	// Verify assets were imported
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var assets []database.Asset
	err = db.NewSelect().Model(&assets).
		Where("workspace = ?", "test-workspace").
		OrderExpr("asset_value ASC").
		Scan(ctx)
	require.NoError(t, err)
	assert.Len(t, assets, 3)

	// Check example.com has 2 A records + 1 CNAME
	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "domain", asset.AssetType)
	assert.Equal(t, "dns", asset.Source)
	assert.Equal(t, "93.184.216.34", asset.HostIP)
	assert.Contains(t, asset.DnsRecords, "93.184.216.34")
	assert.Contains(t, asset.DnsRecords, "93.184.216.35")
	assert.Contains(t, asset.DnsRecords, "CNAME:cdn.example.net")

	// Check sub.example.com has A + MX
	err = db.NewSelect().Model(&asset).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "sub.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "10.0.0.1", asset.HostIP)
	assert.Contains(t, asset.DnsRecords, "10.0.0.1")
	assert.Contains(t, asset.DnsRecords, "MX:mail.example.com")
}

func TestDbImportDNSAsset_EmptyArgs(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_import_dns_asset("", "/tmp/test.txt")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")

	result, err = registry.Execute(
		`db_import_dns_asset("ws", "")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbImportDNSAsset_CNAME(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create a file with only CNAME records
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "cname-only.txt")
	err := os.WriteFile(testFile, []byte("www.example.com. CNAME example.com.\n"), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`db_import_dns_asset("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result)

	ctx := context.Background()
	db := database.GetDB()
	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "www.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "", asset.HostIP) // No A record
	assert.Contains(t, asset.DnsRecords, "CNAME:example.com")
}

func TestDbImportDNSAsset_Upsert(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/dns-records.txt"

	// Import twice
	result1, err := registry.Execute(
		`db_import_dns_asset("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result1)

	result2, err := registry.Execute(
		`db_import_dns_asset("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(3), result2)

	// Should still only have 3 assets (no duplicates)
	ctx := context.Background()
	db := database.GetDB()
	var count int
	count, err = db.NewSelect().Model((*database.Asset)(nil)).
		Where("workspace = ?", "test-workspace").
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, count)
}

// --- Custom Asset Import Tests ---

func TestDbImportCustomAsset(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/custom-assets.jsonl"
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample custom-assets.jsonl file must exist")

	result, err := registry.Execute(
		`db_import_custom_asset("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, 3, stats["new"])
	assert.Equal(t, 3, stats["total"])
	assert.Equal(t, 0, stats["errors"])

	// Verify specific asset fields
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "api.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com", asset.URL)
	assert.Equal(t, "url", asset.AssetType)
	assert.Equal(t, "custom-scan", asset.Source)
	assert.Contains(t, asset.Remarks, "API endpoint")
	assert.Equal(t, "https://jira.example.com/PROJ-123", asset.ExternalURL)
	assert.Contains(t, asset.Remarks, "api")
	assert.Contains(t, asset.Remarks, "production")

	// Check technologies mapped
	err = db.NewSelect().Model(&asset).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "cdn.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Contains(t, asset.Technologies, "Cloudflare")
	assert.Contains(t, asset.Technologies, "nginx")
}

func TestDbImportCustomAsset_EmptyArgs(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	result, err := registry.Execute(
		`db_import_custom_asset("", "/tmp/test.jsonl")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")

	result, err = registry.Execute(
		`db_import_custom_asset("ws", "")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Contains(t, result.(string), "error:")
}

func TestDbImportCustomAsset_UpdateExisting(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/custom-assets.jsonl"

	// First import
	result, err := registry.Execute(
		`db_import_custom_asset("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats := result.(map[string]interface{})
	assert.Equal(t, 3, stats["new"])
	assert.Equal(t, 0, stats["updated"])

	// Second import - should update
	result, err = registry.Execute(
		`db_import_custom_asset("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats = result.(map[string]interface{})
	assert.Equal(t, 0, stats["new"])
	assert.Equal(t, 3, stats["updated"])
	assert.Equal(t, 3, stats["total"])
}

func TestDbImportCustomAsset_TagsAndExternalURL(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create a JSONL with tags and external_url
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "tagged-assets.jsonl")
	content := `{"asset_value":"tagged.example.com","asset_type":"domain","tags":["critical","external"],"external_url":"https://bugbounty.example.com/report/42"}` + "\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`db_import_custom_asset("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	stats := result.(map[string]interface{})
	assert.Equal(t, 1, stats["new"])

	// Verify tags and external_url persisted
	ctx := context.Background()
	db := database.GetDB()
	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "tagged.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "https://bugbounty.example.com/report/42", asset.ExternalURL)
	assert.Contains(t, asset.Remarks, "critical")
	assert.Contains(t, asset.Remarks, "external")
}

// --- Model Field Tests ---

func TestAssetExternalURLField(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	now := time.Now()
	asset := database.Asset{
		Workspace:   "test-workspace",
		AssetValue:  "ext-url-test.example.com",
		AssetType:   "domain",
		ExternalURL: "https://tracker.example.com/issue/99",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := db.NewInsert().Model(&asset).Exec(ctx)
	require.NoError(t, err)

	var fetched database.Asset
	err = db.NewSelect().Model(&fetched).
		Where("asset_value = ?", "ext-url-test.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "https://tracker.example.com/issue/99", fetched.ExternalURL)
}

func TestAssetRemarksField(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	now := time.Now()
	asset := database.Asset{
		Workspace:  "test-workspace",
		AssetValue: "remarks-test.example.com",
		AssetType:  "domain",
		Remarks:    []string{"web", "api", "production"},
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_, err := db.NewInsert().Model(&asset).Exec(ctx)
	require.NoError(t, err)

	var fetched database.Asset
	err = db.NewSelect().Model(&fetched).
		Where("asset_value = ?", "remarks-test.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"web", "api", "production"}, fetched.Remarks)
}

func TestAssetSizeAndLOCFields(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	now := time.Now()
	asset := database.Asset{
		Workspace:   "test-workspace",
		AssetValue:  "example-corp/web-app",
		AssetType:   "repository",
		Language:    "Go",
		Size:        512000,
		LOC:         12000,
		ExternalURL: "https://github.com/example-corp/web-app",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := db.NewInsert().Model(&asset).Exec(ctx)
	require.NoError(t, err)

	var fetched database.Asset
	err = db.NewSelect().Model(&fetched).
		Where("asset_value = ?", "example-corp/web-app").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "repository", fetched.AssetType)
	assert.Equal(t, "Go", fetched.Language)
	assert.Equal(t, int64(512000), fetched.Size)
	assert.Equal(t, int64(12000), fetched.LOC)
	assert.Equal(t, "https://github.com/example-corp/web-app", fetched.ExternalURL)
}

// --- DNS Records Key Tests ---

func TestMapJSONToAsset_DnsRecordsKey(t *testing.T) {
	// "dns_records" key should be accepted as the primary key
	data := map[string]interface{}{
		"host":        "example.com",
		"url":         "http://example.com",
		"dns_records": []interface{}{"10.0.0.1", "10.0.0.2"},
	}
	asset := mapJSONToAsset(data, "test-workspace", `{"host":"example.com"}`)
	assert.Equal(t, []string{"10.0.0.1", "10.0.0.2"}, asset.DnsRecords)

	// "a" key should still work as a fallback
	dataLegacy := map[string]interface{}{
		"host": "legacy.com",
		"url":  "http://legacy.com",
		"a":    []interface{}{"1.2.3.4"},
	}
	assetLegacy := mapJSONToAsset(dataLegacy, "test-workspace", `{"host":"legacy.com"}`)
	assert.Equal(t, []string{"1.2.3.4"}, assetLegacy.DnsRecords)

	// When both keys exist, "dns_records" takes precedence
	dataBoth := map[string]interface{}{
		"host":        "both.com",
		"url":         "http://both.com",
		"dns_records": []interface{}{"10.0.0.1"},
		"a":           []interface{}{"1.2.3.4"},
	}
	assetBoth := mapJSONToAsset(dataBoth, "test-workspace", `{"host":"both.com"}`)
	assert.Equal(t, []string{"10.0.0.1"}, assetBoth.DnsRecords)
}

// --- Optional Defaults Tests ---

func TestDbImportCustomAsset_OptionalDefaults(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create JSONL: first line has no asset_type/source, second has them
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "defaults.jsonl")
	content := `{"asset_value":"no-type.example.com","url":"http://no-type.example.com"}` + "\n" +
		`{"asset_value":"has-type.example.com","url":"http://has-type.example.com","asset_type":"http","source":"httpx"}` + "\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	result, err := registry.Execute(
		`db_import_custom_asset("test-workspace", "`+testFile+`", "subdomain", "recon")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats := result.(map[string]interface{})
	assert.Equal(t, 2, stats["new"])

	ctx := context.Background()
	db := database.GetDB()

	// First asset should have defaults applied
	var asset1 database.Asset
	err = db.NewSelect().Model(&asset1).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "no-type.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "subdomain", asset1.AssetType)
	assert.Equal(t, "recon", asset1.Source)

	// Second asset should keep its own values
	var asset2 database.Asset
	err = db.NewSelect().Model(&asset2).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "has-type.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "http", asset2.AssetType)
	assert.Equal(t, "httpx", asset2.Source)
}

func TestDbImportCustomAsset_PartialDefaults(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "partial.jsonl")
	content := `{"asset_value":"partial.example.com","url":"http://partial.example.com"}` + "\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	registry := NewRegistry()
	// Only 3rd arg (asset_type) provided, no 4th arg (source)
	result, err := registry.Execute(
		`db_import_custom_asset("test-workspace", "`+testFile+`", "dns")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats := result.(map[string]interface{})
	assert.Equal(t, 1, stats["new"])

	ctx := context.Background()
	db := database.GetDB()

	var asset database.Asset
	err = db.NewSelect().Model(&asset).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "partial.example.com").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "dns", asset.AssetType)
	assert.Equal(t, "", asset.Source)
}

func TestDbImportCustomAsset_ContentDiscovery(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/cusom-content-discovery.jsonl"
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample cusom-content-discovery.jsonl file must exist")

	result, err := registry.Execute(
		`db_import_custom_asset("test-workspace", "`+testFile+`", "url", "content-discovery")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, 9, stats["new"])
	assert.Equal(t, 9, stats["total"])
	assert.Equal(t, 0, stats["errors"])

	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	// Line 1: url=http://cdn.doitac.example.com/common/, no asset_value, webserver=UploadServer
	var asset1 database.Asset
	err = db.NewSelect().Model(&asset1).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "http://cdn.doitac.example.com/common/").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "http://cdn.doitac.example.com/common/", asset1.URL)
	assert.Contains(t, asset1.Remarks, "UploadServer")

	// Line 6: url=http://e-procurement.example.com/static, remarks=["Modern-App"], webserver=SGW
	var asset6 database.Asset
	err = db.NewSelect().Model(&asset6).
		Where("workspace = ?", "test-workspace").
		Where("asset_value = ?", "http://e-procurement.example.com/static").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "http://e-procurement.example.com/static", asset6.URL)
	assert.Contains(t, asset6.Remarks, "Modern-App")
	assert.Contains(t, asset6.Remarks, "SGW")
}
