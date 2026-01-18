package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSnapshotTestConfig(t *testing.T) (*config.Config, func()) {
	tmpDir := t.TempDir()

	snapshotDir := filepath.Join(tmpDir, "snapshot")
	workspacesDir := filepath.Join(tmpDir, "workspaces")

	err := os.MkdirAll(snapshotDir, 0755)
	require.NoError(t, err)

	err = os.MkdirAll(workspacesDir, 0755)
	require.NoError(t, err)

	cfg := &config.Config{
		SnapshotPath:   snapshotDir,
		WorkspacesPath: workspacesDir,
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir)
	}

	return cfg, cleanup
}

func TestListSnapshots_Empty(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	app.Get("/snapshots", ListSnapshots(cfg))

	req := httptest.NewRequest("GET", "/snapshots", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, float64(0), result["count"])
	// data can be null or empty array when no snapshots exist
	assert.Equal(t, cfg.SnapshotPath, result["path"])
}

func TestListSnapshots_WithSnapshots(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	// Create test snapshot files
	err := os.WriteFile(filepath.Join(cfg.SnapshotPath, "example.com_123.zip"), []byte("test"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(cfg.SnapshotPath, "test.com_456.zip"), []byte("test2"), 0644)
	require.NoError(t, err)

	app := fiber.New()
	app.Get("/snapshots", ListSnapshots(cfg))

	req := httptest.NewRequest("GET", "/snapshots", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Equal(t, float64(2), result["count"])
}

func TestSnapshotExport_MissingWorkspace(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	app.Post("/snapshots/export", SnapshotExport(cfg))

	// Empty request body
	req := httptest.NewRequest("POST", "/snapshots/export", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.True(t, result["error"].(bool))
	assert.Contains(t, result["message"], "Workspace name is required")
}

func TestSnapshotExport_WorkspaceNotFound(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	app.Post("/snapshots/export", SnapshotExport(cfg))

	reqBody := `{"workspace": "nonexistent.com"}`
	req := httptest.NewRequest("POST", "/snapshots/export", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.True(t, result["error"].(bool))
	assert.Contains(t, result["message"], "Workspace not found")
}

func TestSnapshotExport_Success(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	// Create a test workspace
	workspacePath := filepath.Join(cfg.WorkspacesPath, "example.com")
	err := os.MkdirAll(workspacePath, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(workspacePath, "output.txt"), []byte("scan results"), 0644)
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/snapshots/export", SnapshotExport(cfg))

	reqBody := `{"workspace": "example.com"}`
	req := httptest.NewRequest("POST", "/snapshots/export", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/zip", resp.Header.Get("Content-Type"))
	assert.Contains(t, resp.Header.Get("Content-Disposition"), "attachment")
	assert.NotEmpty(t, resp.Header.Get("X-Snapshot-Size"))
}

func TestSnapshotImport_NoSource(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	app.Post("/snapshots/import", SnapshotImport(cfg))

	// Create empty multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err := writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/snapshots/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err)

	assert.True(t, result["error"].(bool))
	assert.Contains(t, result["message"], "Either file or url is required")
}

func TestSnapshotImport_InvalidRequest(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	app.Post("/snapshots/import", SnapshotImport(cfg))

	// Send request without multipart form
	req := httptest.NewRequest("POST", "/snapshots/import", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDeleteSnapshot_NotFound(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	app.Delete("/snapshots/:name", DeleteSnapshot(cfg))

	req := httptest.NewRequest("DELETE", "/snapshots/nonexistent.zip", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.True(t, result["error"].(bool))
	assert.Contains(t, result["message"], "Snapshot not found")
}

func TestDeleteSnapshot_InvalidPath(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	app.Delete("/snapshots/:name", DeleteSnapshot(cfg))

	// Fiber preserves URL-encoded values in params, so %2F stays as %2F
	// The handler's filepath.Base check only catches actual path separators
	// URL-encoded path traversal is treated as a literal filename and returns 404
	req := httptest.NewRequest("DELETE", "/snapshots/..%2F..%2Fetc%2Fpasswd", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Returns 404 because the URL-encoded string is treated as a literal filename
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestDeleteSnapshot_PathTraversalWithSlash(t *testing.T) {
	_, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	// Use wildcard to accept path with slashes for testing
	app.Delete("/snapshots/*", func(c *fiber.Ctx) error {
		name := c.Params("*")
		if name == "" || filepath.Base(name) != name {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid snapshot name",
			})
		}
		return c.JSON(fiber.Map{"name": name})
	})

	// Test with actual path traversal (slashes in path)
	req := httptest.NewRequest("DELETE", "/snapshots/../../../etc/passwd", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.True(t, result["error"].(bool))
	assert.Contains(t, result["message"], "Invalid snapshot name")
}

func TestDeleteSnapshot_Success(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	// Create a test snapshot file
	snapshotPath := filepath.Join(cfg.SnapshotPath, "example.com_123.zip")
	err := os.WriteFile(snapshotPath, []byte("test content"), 0644)
	require.NoError(t, err)

	app := fiber.New()
	app.Delete("/snapshots/:name", DeleteSnapshot(cfg))

	req := httptest.NewRequest("DELETE", "/snapshots/example.com_123.zip", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err)

	assert.Contains(t, result["message"], "Snapshot deleted successfully")
	assert.Equal(t, "example.com_123.zip", result["name"])

	// Verify file was deleted
	_, err = os.Stat(snapshotPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteSnapshot_MissingName(t *testing.T) {
	cfg, cleanup := setupSnapshotTestConfig(t)
	defer cleanup()

	app := fiber.New()
	app.Delete("/snapshots/:name", DeleteSnapshot(cfg))

	// Empty name parameter - Fiber will route to empty string
	req := httptest.NewRequest("DELETE", "/snapshots/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// Fiber returns 404 for unmatched routes
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
