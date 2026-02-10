package functions

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBImportSARIF(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/semgrep-results.sarif"
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample semgrep-results.sarif file must exist")

	result, err := registry.Execute(
		`db_import_sarif("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	// 27 results but 5 share same ruleId+URI (xss.ejs.explicit-unescape + products.ejs has 6 occurrences),
	// so upsert produces 22 new + 5 unchanged (same fields) = 27 total processed
	assert.Equal(t, 22, stats["new"])
	assert.Equal(t, 5, stats["unchanged"])
	assert.Equal(t, 27, stats["total"])
	assert.Equal(t, 0, stats["errors"])

	// Verify vulnerabilities were imported (22 unique by ruleId+assetValue)
	ctx := context.Background()
	db := database.GetDB()
	require.NotNil(t, db)

	var vulns []database.Vulnerability
	err = db.NewSelect().Model(&vulns).
		Where("workspace = ?", "test-workspace").
		Scan(ctx)
	require.NoError(t, err)
	assert.Len(t, vulns, 22)

	// Check specific fields on the first result (missing-user)
	var vuln database.Vulnerability
	err = db.NewSelect().Model(&vuln).
		Where("workspace = ?", "test-workspace").
		Where("vuln_info = ?", "dockerfile.security.missing-user.missing-user").
		Scan(ctx)
	require.NoError(t, err)
	assert.Equal(t, "repo", vuln.AssetType)
	assert.Equal(t, "Dockerfile", vuln.AssetValue)
	assert.Equal(t, "high", vuln.Severity) // error -> high
	assert.Equal(t, "dockerfile.security.missing-user.missing-user", vuln.VulnTitle)
	assert.Contains(t, vuln.VulnDesc, "not specifying a USER")
	assert.NotEmpty(t, vuln.RawVulnJSON)
	assert.NotEmpty(t, vuln.Tags)

	// Verify idempotency - re-import should produce no new inserts
	result2, err := registry.Execute(
		`db_import_sarif("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	stats2, ok := result2.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 0, stats2["new"])
	assert.Equal(t, 27, stats2["total"])
}

func TestDBImportSARIF_EmptyResults(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/trivy-fs-results.sarif"
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample trivy-fs-results.sarif file must exist")

	result, err := registry.Execute(
		`db_import_sarif("test-workspace", "`+testFile+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)

	stats, ok := result.(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, 0, stats["new"])
	assert.Equal(t, 0, stats["total"])
	assert.Equal(t, 0, stats["errors"])
}

func TestDBImportSARIF_InvalidArgs(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	registry := NewRegistry()

	// Empty workspace
	result, err := registry.Execute(
		`db_import_sarif("", "/some/file.sarif")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	str, ok := result.(string)
	require.True(t, ok)
	assert.Contains(t, str, "error")

	// Empty path
	result, err = registry.Execute(
		`db_import_sarif("test-workspace", "")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	str, ok = result.(string)
	require.True(t, ok)
	assert.Contains(t, str, "error")

	// Nonexistent file
	result, err = registry.Execute(
		`db_import_sarif("test-workspace", "/nonexistent/file.sarif")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	str, ok = result.(string)
	require.True(t, ok)
	assert.Contains(t, str, "error")
}

func TestConvertSARIFToMarkdown(t *testing.T) {
	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/semgrep-results.sarif"
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample semgrep-results.sarif file must exist")

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "sarif-report.md")

	result, err := registry.Execute(
		`convert_sarif_to_markdown("`+testFile+`", "`+outputPath+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Read output and verify
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	md := string(content)

	// Check header
	assert.Contains(t, md, "# SARIF Results")
	assert.Contains(t, md, "Semgrep")

	// Check summary line
	assert.Contains(t, md, "Total: 27 findings")
	assert.Contains(t, md, "High: 5")
	assert.Contains(t, md, "Medium: 21")
	assert.Contains(t, md, "Low: 1")

	// Check table structure
	assert.Contains(t, md, "| Severity | Location | Title | Description |")
	assert.Contains(t, md, "| --- | --- | --- | --- |")

	// Check that high severity comes before medium
	highIdx := strings.Index(md, "| high |")
	mediumIdx := strings.Index(md, "| medium |")
	assert.True(t, highIdx < mediumIdx, "high severity should come before medium")

	// Check that we have Dockerfile reference
	assert.Contains(t, md, "Dockerfile")
}

func TestConvertSARIFToMarkdown_EmptyResults(t *testing.T) {
	registry := NewRegistry()

	testFile := "../../test/testdata/sample-jsonl-output/trivy-fs-results.sarif"
	_, err := os.Stat(testFile)
	require.NoError(t, err, "sample trivy-fs-results.sarif file must exist")

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "trivy-report.md")

	result, err := registry.Execute(
		`convert_sarif_to_markdown("`+testFile+`", "`+outputPath+`")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, true, result)

	// Read output - should have header but no table rows
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	md := string(content)

	assert.Contains(t, md, "# SARIF Results")
	assert.Contains(t, md, "Trivy")
	assert.Contains(t, md, "Total: 0 findings")
	// Should NOT contain table since no findings
	assert.NotContains(t, md, "| --- | --- | --- | --- |")
}

func TestConvertSARIFToMarkdown_InvalidArgs(t *testing.T) {
	registry := NewRegistry()

	// Empty input path
	result, err := registry.Execute(
		`convert_sarif_to_markdown("", "/tmp/output.md")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, false, result)

	// Empty output path
	result, err = registry.Execute(
		`convert_sarif_to_markdown("/some/input.sarif", "")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, false, result)

	// Nonexistent file
	result, err = registry.Execute(
		`convert_sarif_to_markdown("/nonexistent/file.sarif", "/tmp/output.md")`,
		map[string]interface{}{},
	)
	require.NoError(t, err)
	assert.Equal(t, false, result)
}
