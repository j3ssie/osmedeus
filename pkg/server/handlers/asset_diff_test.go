package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAssetDiffTestDB(t *testing.T) (*config.Config, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_diff.sqlite")
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

	cleanup := func() {
		_ = database.Close()
		database.SetDB(nil)
	}

	return cfg, cleanup
}

func TestGetAssetDiff_Success(t *testing.T) {
	cfg, cleanup := setupAssetDiffTestDB(t)
	defer cleanup()

	ctx := context.Background()
	workspace := "test-workspace"
	now := time.Now()

	// Seed database with test assets
	newAsset := &database.Asset{
		Workspace:  workspace,
		AssetValue: "new.example.com",
		URL:        "https://new.example.com",
		StatusCode: 200,
		CreatedAt:  now.Add(-30 * time.Minute),
		UpdatedAt:  now.Add(-30 * time.Minute),
		LastSeenAt: now,
	}
	_, err := database.GetDB().NewInsert().Model(newAsset).Exec(ctx)
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/assets/diff", GetAssetDiff(cfg))

	// Make request - use URL encoding for the time parameter
	fromTime := now.Add(-1 * time.Hour).Format(time.RFC3339)
	req := httptest.NewRequest("GET", fmt.Sprintf("/assets/diff?workspace=%s&from=%s", workspace, url.QueryEscape(fromTime)), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, workspace, data["workspace_name"])

	summary, ok := data["summary"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), summary["total_added"])
}

func TestGetAssetDiff_MissingWorkspace(t *testing.T) {
	cfg, cleanup := setupAssetDiffTestDB(t)
	defer cleanup()

	app := fiber.New()
	app.Get("/assets/diff", GetAssetDiff(cfg))

	req := httptest.NewRequest("GET", "/assets/diff?from=2024-01-01T00:00:00Z", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, true, result["error"])
	assert.Contains(t, result["message"], "workspace")
}

func TestGetAssetDiff_MissingFromTime(t *testing.T) {
	cfg, cleanup := setupAssetDiffTestDB(t)
	defer cleanup()

	app := fiber.New()
	app.Get("/assets/diff", GetAssetDiff(cfg))

	req := httptest.NewRequest("GET", "/assets/diff?workspace=test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, true, result["error"])
	assert.Contains(t, result["message"], "from")
}

func TestGetAssetDiff_InvalidFromTime(t *testing.T) {
	cfg, cleanup := setupAssetDiffTestDB(t)
	defer cleanup()

	app := fiber.New()
	app.Get("/assets/diff", GetAssetDiff(cfg))

	req := httptest.NewRequest("GET", "/assets/diff?workspace=test&from=invalid-time", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, true, result["error"])
	assert.Contains(t, result["message"], "invalid")
}

func TestGetAssetDiff_UnixTimestamp(t *testing.T) {
	cfg, cleanup := setupAssetDiffTestDB(t)
	defer cleanup()

	app := fiber.New()
	app.Get("/assets/diff", GetAssetDiff(cfg))

	// Use Unix timestamp
	unixTime := time.Now().Add(-1 * time.Hour).Unix()
	req := httptest.NewRequest("GET", fmt.Sprintf("/assets/diff?workspace=test&from=%d", unixTime), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetVulnerabilityDiff_Success(t *testing.T) {
	cfg, cleanup := setupAssetDiffTestDB(t)
	defer cleanup()

	ctx := context.Background()
	workspace := "test-workspace"
	now := time.Now()

	// Seed database with test vulnerability
	newVuln := &database.Vulnerability{
		Workspace:  workspace,
		VulnInfo:   "CVE-2024-1234",
		VulnTitle:  "Test Vuln",
		Severity:   "high",
		AssetValue: "vulnerable.example.com",
		CreatedAt:  now.Add(-30 * time.Minute),
		UpdatedAt:  now.Add(-30 * time.Minute),
		LastSeenAt: now,
	}
	_, err := database.GetDB().NewInsert().Model(newVuln).Exec(ctx)
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/vulnerabilities/diff", GetVulnerabilityDiff(cfg))

	// Make request - use URL encoding for the time parameter
	fromTime := now.Add(-1 * time.Hour).Format(time.RFC3339)
	req := httptest.NewRequest("GET", fmt.Sprintf("/vulnerabilities/diff?workspace=%s&from=%s", workspace, url.QueryEscape(fromTime)), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	data, ok := result["data"].(map[string]interface{})
	require.True(t, ok)

	summary, ok := data["summary"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(1), summary["total_added"])
}

func TestParseTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "RFC3339",
			input:   "2024-01-15T10:30:00Z",
			wantErr: false,
		},
		{
			name:    "Unix timestamp",
			input:   "1705315800",
			wantErr: false,
		},
		{
			name:    "Date only",
			input:   "2024-01-15",
			wantErr: false,
		},
		{
			name:    "Date and time without zone",
			input:   "2024-01-15T10:30:00",
			wantErr: false,
		},
		{
			name:    "Date and time with space",
			input:   "2024-01-15 10:30:00",
			wantErr: false,
		},
		{
			name:    "Invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTime(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}
