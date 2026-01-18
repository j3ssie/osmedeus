package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestWorkflowDir(t *testing.T) (*config.Config, string) {
	tmpDir := t.TempDir()

	// Create modules directory
	modulesDir := filepath.Join(tmpDir, "modules")
	err := os.MkdirAll(modulesDir, 0755)
	require.NoError(t, err)

	// Create a test flow
	flowContent := `kind: flow
name: test-flow
description: Test flow for API testing

params:
  - name: target
    required: true

modules:
  - name: test-module
    path: modules/test-module.yaml
`
	err = os.WriteFile(filepath.Join(tmpDir, "test-flow.yaml"), []byte(flowContent), 0644)
	require.NoError(t, err)

	// Create a test module
	moduleContent := `kind: module
name: test-module
description: Test module for API testing

params:
  - name: target
    required: true
  - name: threads
    default: "10"

steps:
  - name: echo-test
    type: bash
    command: echo "Hello {{target}}"
`
	err = os.WriteFile(filepath.Join(modulesDir, "test-module.yaml"), []byte(moduleContent), 0644)
	require.NoError(t, err)

	// Create another module
	module2Content := `kind: module
name: scan-module
description: Scan module for testing

trigger:
  - name: manual
    on: manual
    enabled: true
  - name: cron
    on: cron
    schedule: "0 0 * * *"
    enabled: false

params:
  - name: target
    required: true

steps:
  - name: scan
    type: bash
    command: echo "Scanning {{target}}"
  - name: report
    type: bash
    command: echo "Reporting"
`
	err = os.WriteFile(filepath.Join(modulesDir, "scan-module.yaml"), []byte(module2Content), 0644)
	require.NoError(t, err)

	cfg := &config.Config{
		WorkflowsPath: tmpDir,
	}

	return cfg, tmpDir
}

func TestHealthCheck(t *testing.T) {
	app := fiber.New()
	app.Get("/health", HealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "ok", result["status"])
}

func TestReadinessCheck(t *testing.T) {
	app := fiber.New()
	app.Get("/ready", ReadinessCheck)

	req := httptest.NewRequest("GET", "/ready", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "ready", result["status"])
}

func TestListWorkflows(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Get("/workflows", ListWorkflows(cfg))

	req := httptest.NewRequest("GET", "/workflows", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	data, ok := result["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 3) // 1 flow + 2 modules
}

func TestListWorkflowsVerbose(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Get("/workflows", ListWorkflowsVerbose(cfg))

	req := httptest.NewRequest("GET", "/workflows", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	data, ok := result["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 3)

	// Check that verbose data includes params
	for _, item := range data {
		wf, ok := item.(map[string]interface{})
		require.True(t, ok)

		assert.Contains(t, wf, "name")
		assert.Contains(t, wf, "kind")
		assert.Contains(t, wf, "description")
		assert.Contains(t, wf, "params")
		assert.Contains(t, wf, "required_params")
	}

	// Check count
	count, ok := result["count"].(float64)
	require.True(t, ok)
	assert.Equal(t, float64(3), count)
}

func TestGetWorkflow(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Get("/workflows/:name", GetWorkflow(cfg))

	req := httptest.NewRequest("GET", "/workflows/test-module", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "test-module", result["name"])
	assert.Equal(t, "module", result["kind"])
	assert.Equal(t, "Test module for API testing", result["description"])
	assert.Contains(t, result, "params")
}

func TestGetWorkflow_NotFound(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Get("/workflows/:name", GetWorkflow(cfg))

	req := httptest.NewRequest("GET", "/workflows/nonexistent", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, true, result["error"])
	assert.Contains(t, result["message"], "not found")
}

func TestListWorkspaceNames(t *testing.T) {
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
	t.Cleanup(func() {
		_ = database.Close()
		database.SetDB(nil)
	})

	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx))

	now := time.Now()
	ws1 := &database.Workspace{Name: "b.example", DataSource: "local", Tags: []string{}, CreatedAt: now, UpdatedAt: now}
	ws2 := &database.Workspace{Name: "a.example", DataSource: "local", Tags: []string{}, CreatedAt: now, UpdatedAt: now}
	_, err = database.GetDB().NewInsert().Model(ws1).Exec(ctx)
	require.NoError(t, err)
	_, err = database.GetDB().NewInsert().Model(ws2).Exec(ctx)
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/workspace-names", ListWorkspaceNames(cfg))

	req := httptest.NewRequest("GET", "/workspace-names", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var names []string
	err = json.Unmarshal(body, &names)
	require.NoError(t, err)
	assert.Equal(t, []string{"a.example", "b.example"}, names)
}

func TestListWorkspaces_FilesystemIncludesWorkspaceFolders(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.sqlite")
	workspacesDir := filepath.Join(tmpDir, "workspaces")
	require.NoError(t, os.MkdirAll(filepath.Join(workspacesDir, "fs-only"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(workspacesDir, "shared"), 0755))

	cfg := &config.Config{
		BaseFolder:     tmpDir,
		WorkspacesPath: workspacesDir,
		Database: config.DatabaseConfig{
			DBEngine: "sqlite",
			DBPath:   dbPath,
		},
	}

	_, err := database.Connect(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = database.Close()
		database.SetDB(nil)
	})

	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx))

	now := time.Now()
	asset1 := &database.Asset{Workspace: "db-only", AssetValue: "http://example.com", CreatedAt: now, UpdatedAt: now}
	asset2 := &database.Asset{Workspace: "shared", AssetValue: "http://shared.example.com", CreatedAt: now, UpdatedAt: now}
	_, err = database.GetDB().NewInsert().Model(asset1).Exec(ctx)
	require.NoError(t, err)
	_, err = database.GetDB().NewInsert().Model(asset2).Exec(ctx)
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/workspaces", ListWorkspaces(cfg))

	req := httptest.NewRequest("GET", "/workspaces?filesystem=true&limit=100", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	data, ok := result["data"].([]interface{})
	require.True(t, ok)

	foundTotalAssets := map[string]float64{}
	foundDataSource := map[string]string{}
	foundTags := map[string][]interface{}{}
	for _, item := range data {
		m, ok := item.(map[string]interface{})
		require.True(t, ok)
		name, _ := m["name"].(string)
		totalAssets, _ := m["total_assets"].(float64)
		dataSource, _ := m["data_source"].(string)
		tags, _ := m["tags"].([]interface{})
		foundTotalAssets[name] = totalAssets
		foundDataSource[name] = dataSource
		foundTags[name] = tags
	}

	assert.Contains(t, foundTotalAssets, "db-only")
	assert.Contains(t, foundTotalAssets, "fs-only")
	assert.Contains(t, foundTotalAssets, "shared")
	assert.Equal(t, float64(1), foundTotalAssets["db-only"])
	assert.Equal(t, float64(0), foundTotalAssets["fs-only"])
	assert.Equal(t, float64(1), foundTotalAssets["shared"])

	assert.Equal(t, "filesystem", foundDataSource["db-only"])
	assert.Equal(t, "filesystem", foundDataSource["fs-only"])
	assert.Equal(t, "filesystem", foundDataSource["shared"])

	assert.Contains(t, foundTags["db-only"], "filesystem")
	assert.Contains(t, foundTags["fs-only"], "filesystem")
	assert.Contains(t, foundTags["shared"], "filesystem")

	assert.NotContains(t, foundTags["db-only"], "filesystem-only")
	assert.Contains(t, foundTags["fs-only"], "filesystem-only")
	assert.NotContains(t, foundTags["shared"], "filesystem-only")

	pagination, ok := result["pagination"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(3), pagination["total"])
}

func TestListArtifactsVerifyExist(t *testing.T) {
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
	t.Cleanup(func() {
		_ = database.Close()
		database.SetDB(nil)
	})

	ctx := context.Background()
	require.NoError(t, database.Migrate(ctx))

	filePath := filepath.Join(tmpDir, "out.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("ok"), 0644))

	folderPath := filepath.Join(tmpDir, "outdir")
	require.NoError(t, os.MkdirAll(folderPath, 0755))

	now := time.Now()
	art1 := &database.Artifact{
		ID:           "a1",
		RunID:        "r1",
		Workspace:    "w1",
		Name:         "file",
		ArtifactPath: filePath,
		ArtifactType: database.ArtifactTypeOutput,
		ContentType:  database.ContentTypeText,
		SizeBytes:    2,
		LineCount:    1,
		CreatedAt:    now,
	}
	art2 := &database.Artifact{
		ID:           "a2",
		RunID:        "r1",
		Workspace:    "w1",
		Name:         "folder",
		ArtifactPath: folderPath,
		ArtifactType: database.ArtifactTypeOutput,
		ContentType:  database.ContentTypeFolder,
		CreatedAt:    now,
	}
	art3 := &database.Artifact{
		ID:           "a3",
		RunID:        "r1",
		Workspace:    "w1",
		Name:         "missing",
		ArtifactPath: filepath.Join(tmpDir, "missing.txt"),
		ArtifactType: database.ArtifactTypeOutput,
		ContentType:  database.ContentTypeText,
		CreatedAt:    now,
	}

	_, err = database.GetDB().NewInsert().Model(art1).Exec(ctx)
	require.NoError(t, err)
	_, err = database.GetDB().NewInsert().Model(art2).Exec(ctx)
	require.NoError(t, err)
	_, err = database.GetDB().NewInsert().Model(art3).Exec(ctx)
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/artifacts", ListArtifacts(cfg))

	req := httptest.NewRequest("GET", "/artifacts?verify_exist=true&limit=100", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	data, ok := result["data"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 2, len(data))

	byID := map[string]map[string]interface{}{}
	for _, item := range data {
		m, ok := item.(map[string]interface{})
		require.True(t, ok)
		id, _ := m["id"].(string)
		if id != "" {
			byID[id] = m
		}
	}

	assert.Equal(t, true, byID["a1"]["path_exists"])
	assert.Equal(t, false, byID["a1"]["path_is_dir"])
	assert.Equal(t, true, byID["a2"]["path_exists"])
	assert.Equal(t, true, byID["a2"]["path_is_dir"])
	_, hasMissing := byID["a3"]
	assert.False(t, hasMissing)
}

func TestGetWorkflowVerbose(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Get("/workflows/:name", GetWorkflowVerbose(cfg))

	req := httptest.NewRequest("GET", "/workflows/scan-module?json=true", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "scan-module", result["name"])
	assert.Equal(t, "module", result["kind"])

	// Check params
	params, ok := result["params"].([]interface{})
	require.True(t, ok)
	assert.Len(t, params, 1)

	// Check steps
	steps, ok := result["steps"].([]interface{})
	require.True(t, ok)
	assert.Len(t, steps, 2)

	// Check triggers
	triggers, ok := result["triggers"].([]interface{})
	require.True(t, ok)
	assert.Len(t, triggers, 2)
}

func TestGetWorkflowVerbose_Flow(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Get("/workflows/:name", GetWorkflowVerbose(cfg))

	req := httptest.NewRequest("GET", "/workflows/test-flow?json=true", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "test-flow", result["name"])
	assert.Equal(t, "flow", result["kind"])

	// Check modules
	modules, ok := result["modules"].([]interface{})
	require.True(t, ok)
	assert.Len(t, modules, 1)
}

func TestValidateWorkflow(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Get("/workflows/:name/validate", ValidateWorkflow(cfg))

	req := httptest.NewRequest("GET", "/workflows/test-module/validate", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, true, result["valid"])
}

func TestValidateWorkflow_NotFound(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Get("/workflows/:name/validate", ValidateWorkflow(cfg))

	req := httptest.NewRequest("GET", "/workflows/nonexistent/validate", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestReloadWorkflows(t *testing.T) {
	cfg, _ := setupTestWorkflowDir(t)

	app := fiber.New()
	app.Post("/workflows/reload", ReloadWorkflows(cfg))

	req := httptest.NewRequest("POST", "/workflows/reload", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Contains(t, result["message"], "reloaded")
}

func TestGetSettings(t *testing.T) {
	cfg := &config.Config{
		BaseFolder: "/test/base",
		Server: config.ServerConfig{
			Host: "localhost",
			Port: 8811,
		},
	}

	app := fiber.New()
	app.Get("/settings", GetSettings(cfg))

	req := httptest.NewRequest("GET", "/settings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, "/test/base", result["base_folder"])
	assert.Equal(t, core.VERSION, result["version"])
}
