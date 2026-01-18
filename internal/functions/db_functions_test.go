package functions

import (
	"context"
	"os"
	"path/filepath"
	"testing"

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

	// http-data.jsonl has 3 lines
	assert.Equal(t, int64(3), result)

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

	// vuln-data.jsonl has 13 lines
	assert.Equal(t, int64(13), result)

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
	assert.Equal(t, int64(2), result) // Both lines processed

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
	assert.Equal(t, "nginx", asset.Source)
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
