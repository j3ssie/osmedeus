package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testAPIPort    string
	testAPIBaseURL string
	testAPIBaseDir string
)

// startRedisForAPI starts the Redis container for API tests
func startRedisForAPI(t *testing.T, log *TestLogger) func() {
	t.Helper()
	composePath := getDockerComposePath(t)

	log.Info("Starting Redis container for API tests")

	cmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v\nOutput: %s", err, output)
	}

	// Wait for Redis to be ready
	waitForRedisAPI(t, log, testRedisPort, 15*time.Second)

	return func() {
		log.Info("Stopping Redis container")
		cmd := exec.Command("docker-compose", "-f", composePath, "down", "-v")
		_ = cmd.Run()
	}
}

// waitForRedisAPI waits for Redis to accept connections
func waitForRedisAPI(t *testing.T, log *TestLogger, port string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)

	log.Info("Waiting for Redis on port %s...", port)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", "localhost:"+port, time.Second)
		if err == nil {
			_ = conn.Close()
			log.Success("Redis is ready")
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("Redis not ready after %v", timeout)
}

// cleanAndSeedDatabase cleans and seeds the database using CLI
func cleanAndSeedDatabase(t *testing.T, log *TestLogger) {
	t.Helper()
	binary := getBinaryPath(t)

	log.Info("Cleaning database...")
	cmd := exec.Command(binary, "--base-folder", testAPIBaseDir, "db", "clean", "--force")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Warn("Database clean failed (may be empty): %s", string(output))
	}

	log.Info("Seeding database with sample data...")
	cmd = exec.Command(binary, "--base-folder", testAPIBaseDir, "db", "seed")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to seed database: %v\nOutput: %s", err, output)
	}
	log.Success("Database seeded successfully")
}

// startAPIServer starts the API server for testing
func startAPIServer(t *testing.T, log *TestLogger, ctx context.Context) (*exec.Cmd, func()) {
	t.Helper()
	binary := getBinaryPath(t)
	workflowPath := getTestdataPath(t)

	log.Info("Starting API server on port %s", testAPIPort)

	cmd := exec.CommandContext(ctx, binary, "serve",
		"--master",
		"--port", testAPIPort,
		"--redis-url", testRedisURL,
		"-A", // Disable auth for testing
		"--base-folder", testAPIBaseDir,
		"-F", workflowPath,
	)

	err := cmd.Start()
	require.NoError(t, err, "Failed to start API server")

	// Wait for server to be ready
	waitForAPIServer(t, log, testAPIPort, 15*time.Second)

	return cmd, func() {
		log.Info("Stopping API server")
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}
}

// waitForAPIServer waits for the API server to respond
func waitForAPIServer(t *testing.T, log *TestLogger, port string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	url := "http://localhost:" + port + "/health"

	log.Info("Waiting for API server at %s...", url)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				log.Success("API server is ready")
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("API server not ready after %v", timeout)
}

// apiGet performs a GET request to the API
func apiGet(t *testing.T, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(testAPIBaseURL + path)
	require.NoError(t, err, "GET %s failed", path)
	return resp
}

// apiPost performs a POST request to the API
func apiPost(t *testing.T, path string, body interface{}) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reader = bytes.NewReader(jsonBody)
	}

	resp, err := http.Post(testAPIBaseURL+path, "application/json", reader)
	require.NoError(t, err, "POST %s failed", path)
	return resp
}

// apiPut performs a PUT request to the API
func apiPut(t *testing.T, path string, body interface{}) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(http.MethodPut, testAPIBaseURL+path, reader)
	require.NoError(t, err, "Failed to create PUT request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "PUT %s failed", path)
	return resp
}

// apiDelete performs a DELETE request to the API
func apiDelete(t *testing.T, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, testAPIBaseURL+path, nil)
	require.NoError(t, err, "Failed to create DELETE request")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "DELETE %s failed", path)
	return resp
}

// parseJSONResponse parses the JSON response body
func parseJSONResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "Failed to parse JSON response: %s", string(body))

	return result
}

// TestAPI_AllEndpoints tests all API endpoints with a shared setup
func TestAPI_AllEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping API e2e test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Running API E2E tests")

	testAPIPort = getFreePort(t)
	testAPIBaseURL = "http://localhost:" + testAPIPort
	testAPIBaseDir = t.TempDir()

	// Step 1: Start Redis
	log.Step("Step 1: Starting Redis")
	cleanupRedis := startRedisForAPI(t, log)
	defer cleanupRedis()

	// Step 2: Clean and seed database
	log.Step("Step 2: Cleaning and seeding database")
	cleanAndSeedDatabase(t, log)

	// Step 3: Start API server
	log.Step("Step 3: Starting API server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	_, cleanupServer := startAPIServer(t, log, ctx)
	defer cleanupServer()

	// Step 4: Run subtests
	log.Step("Step 4: Running API endpoint tests")

	t.Run("Health", func(t *testing.T) {
		testHealthEndpoints(t, log)
	})

	t.Run("Info", func(t *testing.T) {
		testInfoEndpoints(t, log)
	})

	t.Run("Workflows", func(t *testing.T) {
		testWorkflowEndpoints(t, log)
	})

	t.Run("Workspaces", func(t *testing.T) {
		testWorkspaceEndpoints(t, log)
	})

	t.Run("Assets", func(t *testing.T) {
		testAssetEndpoints(t, log)
	})

	t.Run("Stats", func(t *testing.T) {
		testStatsEndpoint(t, log)
	})

	t.Run("Schedules", func(t *testing.T) {
		testScheduleEndpoints(t, log)
	})

	t.Run("Registry", func(t *testing.T) {
		testRegistryEndpoint(t, log)
	})

	t.Run("Distributed", func(t *testing.T) {
		testDistributedEndpoints(t, log)
	})

	t.Run("Vulnerabilities", func(t *testing.T) {
		testVulnerabilityEndpoints(t, log)
	})

	t.Run("Runs", func(t *testing.T) {
		testRunEndpoints(t, log)
	})

	t.Run("EventLogs", func(t *testing.T) {
		testEventLogEndpoints(t, log)
	})

	t.Run("Functions", func(t *testing.T) {
		testFunctionEndpoints(t, log)
	})

	t.Run("Settings", func(t *testing.T) {
		testSettingsEndpoints(t, log)
	})

	t.Run("WorkflowExtras", func(t *testing.T) {
		testWorkflowExtras(t, log)
	})

	log.Success("All API E2E tests passed!")
}

func getFreePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "Failed to get free port")
	defer func() { _ = ln.Close() }()
	_, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err, "Failed to parse free port")
	return port
}

// testHealthEndpoints tests health check endpoints
func testHealthEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing health endpoints")

	// GET /health
	resp := apiGet(t, "/health")
	assert.Equal(t, 200, resp.StatusCode, "GET /health should return 200")
	body := parseJSONResponse(t, resp)
	assert.Equal(t, "ok", body["status"], "Health status should be 'ok'")

	// GET /health/ready
	resp = apiGet(t, "/health/ready")
	assert.Equal(t, 200, resp.StatusCode, "GET /health/ready should return 200")

	log.Success("Health endpoints OK")
}

// testInfoEndpoints tests server info endpoints
func testInfoEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing info endpoints")

	// Note: GET / serves the embedded UI (HTML), not JSON API
	// We only test the /server-info JSON endpoint here

	// GET /server-info
	resp := apiGet(t, "/server-info")
	assert.Equal(t, 200, resp.StatusCode, "GET /server-info should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "version", "Server-info should contain version")

	log.Success("Info endpoints OK")
}

// testWorkflowEndpoints tests workflow endpoints
func testWorkflowEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing workflow endpoints")

	// GET /osm/api/workflows (DB-based listing returns pagination)
	resp := apiGet(t, "/osm/api/workflows")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/workflows should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data array")
	// Note: DB-based listing returns "pagination" with "total", not "count"
	assert.Contains(t, body, "pagination", "Should contain pagination")

	// GET /osm/api/workflows/test-bash?json=true (need json=true to get JSON, otherwise returns YAML)
	resp = apiGet(t, "/osm/api/workflows/test-bash?json=true")
	// May return 200 (workflow found) or 404 (workflow not found)
	assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404,
		"GET /osm/api/workflows/test-bash should return 200 or 404")
	if resp.StatusCode == 200 {
		body = parseJSONResponse(t, resp)
		assert.Contains(t, body, "name", "Should contain workflow name")
	}

	log.Success("Workflow endpoints OK")
}

// testWorkspaceEndpoints tests workspace endpoints
func testWorkspaceEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing workspace endpoints")

	// GET /osm/api/workspaces (seeded data)
	resp := apiGet(t, "/osm/api/workspaces")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/workspaces should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data array")
	// Seeded data should have workspaces
	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	assert.NotEmpty(t, data, "Should have seeded workspaces")

	log.Success("Workspace endpoints OK")
}

// testAssetEndpoints tests asset endpoints
func testAssetEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing asset endpoints")

	// GET /osm/api/assets (seeded data)
	resp := apiGet(t, "/osm/api/assets")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/assets should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data array")
	// Seeded data should have 10 assets
	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	assert.NotEmpty(t, data, "Should have seeded assets")

	log.Success("Asset endpoints OK")
}

// testStatsEndpoint tests stats endpoint
func testStatsEndpoint(t *testing.T, log *TestLogger) {
	log.Info("Testing stats endpoint")

	// GET /osm/api/stats
	resp := apiGet(t, "/osm/api/stats")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/stats should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "workflows", "Should contain workflows count")
	assert.Contains(t, body, "assets", "Should contain assets count")

	log.Success("Stats endpoint OK")
}

// testScheduleEndpoints tests schedule CRUD endpoints
func testScheduleEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing schedule endpoints")

	// GET /osm/api/schedules (seeded data)
	resp := apiGet(t, "/osm/api/schedules")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/schedules should return 200")

	// POST /osm/api/schedules - Create new schedule
	schedule := map[string]interface{}{
		"name":          "test-api-schedule",
		"workflow_name": "test-bash",
		"workflow_kind": "module",
		"target":        "api-test.com",
		"schedule":      "0 0 * * *",
		"enabled":       true,
	}
	resp = apiPost(t, "/osm/api/schedules", schedule)
	assert.Equal(t, 201, resp.StatusCode, "POST /osm/api/schedules should return 201")
	body := parseJSONResponse(t, resp)
	// Schedule ID is in the "data" object returned by the API
	data, ok := body["data"].(map[string]interface{})
	require.True(t, ok, "Response should contain data object")
	scheduleID, ok := data["id"].(string)
	require.True(t, ok, "Data should contain schedule id (lowercase)")
	require.NotEmpty(t, scheduleID, "Schedule ID should not be empty")

	// GET /osm/api/schedules/:id
	resp = apiGet(t, "/osm/api/schedules/"+scheduleID)
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/schedules/:id should return 200")

	// PUT /osm/api/schedules/:id
	update := map[string]interface{}{"name": "updated-api-schedule"}
	resp = apiPut(t, "/osm/api/schedules/"+scheduleID, update)
	assert.Equal(t, 200, resp.StatusCode, "PUT /osm/api/schedules/:id should return 200")

	// POST /osm/api/schedules/:id/disable
	resp = apiPost(t, "/osm/api/schedules/"+scheduleID+"/disable", nil)
	assert.Equal(t, 200, resp.StatusCode, "POST /osm/api/schedules/:id/disable should return 200")

	// POST /osm/api/schedules/:id/enable
	resp = apiPost(t, "/osm/api/schedules/"+scheduleID+"/enable", nil)
	assert.Equal(t, 200, resp.StatusCode, "POST /osm/api/schedules/:id/enable should return 200")

	// DELETE /osm/api/schedules/:id
	resp = apiDelete(t, "/osm/api/schedules/"+scheduleID)
	assert.Equal(t, 200, resp.StatusCode, "DELETE /osm/api/schedules/:id should return 200")

	log.Success("Schedule endpoints OK")
}

// testRegistryEndpoint tests metadata registry endpoint
func testRegistryEndpoint(t *testing.T, log *TestLogger) {
	log.Info("Testing registry endpoint")

	resp := apiGet(t, "/osm/api/registry-info")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/registry-info should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "registry_url", "Should contain registry_url")
	assert.Contains(t, body, "binaries", "Should contain binaries")

	log.Success("Registry endpoint OK")
}

// testDistributedEndpoints tests distributed mode endpoints
func testDistributedEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing distributed endpoints")

	// GET /osm/api/workers (master mode enabled)
	resp := apiGet(t, "/osm/api/workers")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/workers should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "count", "Should contain worker count")

	// GET /osm/api/tasks
	resp = apiGet(t, "/osm/api/tasks")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/tasks should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "running", "Should contain running tasks")
	assert.Contains(t, body, "completed", "Should contain completed tasks")

	log.Success("Distributed endpoints OK")
}

// testVulnerabilityEndpoints tests vulnerability CRUD endpoints
func testVulnerabilityEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing vulnerability endpoints")

	// GET /osm/api/vulnerabilities (seeded data)
	resp := apiGet(t, "/osm/api/vulnerabilities")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/vulnerabilities should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data array")
	assert.Contains(t, body, "pagination", "Should contain pagination")
	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")
	assert.NotEmpty(t, data, "Should have seeded vulnerabilities")

	// GET /osm/api/vulnerabilities/summary
	resp = apiGet(t, "/osm/api/vulnerabilities/summary")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/vulnerabilities/summary should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data")
	summaryData, ok := body["data"].(map[string]interface{})
	assert.True(t, ok, "Data should be an object")
	assert.Contains(t, summaryData, "by_severity", "Should contain severity breakdown")
	assert.Contains(t, summaryData, "total", "Should contain total count")

	// POST /osm/api/vulnerabilities - Create new vulnerability
	vuln := map[string]interface{}{
		"workspace":   "test-workspace",
		"vuln_title":  "Test XSS Vulnerability",
		"vuln_info":   "XSS in search parameter",
		"vuln_desc":   "A reflected XSS vulnerability was found",
		"severity":    "high",
		"asset_type":  "endpoint",
		"asset_value": "test.example.com",
		"tags":        []string{"xss", "test"},
	}
	resp = apiPost(t, "/osm/api/vulnerabilities", vuln)
	assert.Equal(t, 201, resp.StatusCode, "POST /osm/api/vulnerabilities should return 201")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain created data")
	createdData, ok := body["data"].(map[string]interface{})
	require.True(t, ok, "Data should be an object")
	vulnID, ok := createdData["id"].(float64)
	require.True(t, ok, "Should have numeric id")
	require.NotZero(t, vulnID, "Vulnerability ID should not be zero")

	// GET /osm/api/vulnerabilities/:id
	vulnIDStr := fmt.Sprintf("%.0f", vulnID)
	resp = apiGet(t, "/osm/api/vulnerabilities/"+vulnIDStr)
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/vulnerabilities/:id should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data")

	// DELETE /osm/api/vulnerabilities/:id
	resp = apiDelete(t, "/osm/api/vulnerabilities/"+vulnIDStr)
	assert.Equal(t, 200, resp.StatusCode, "DELETE /osm/api/vulnerabilities/:id should return 200")

	// Verify deletion
	resp = apiGet(t, "/osm/api/vulnerabilities/"+vulnIDStr)
	assert.Equal(t, 404, resp.StatusCode, "GET deleted vulnerability should return 404")

	log.Success("Vulnerability endpoints OK")
}

// testRunEndpoints tests run management endpoints
func testRunEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing run endpoints")

	// GET /osm/api/runs
	// Note: Current implementation is a stub that returns empty data
	resp := apiGet(t, "/osm/api/runs")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/runs should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data array")
	data, ok := body["data"].([]interface{})
	assert.True(t, ok, "Data should be an array")

	// Use a test run ID for endpoint testing (handlers are stubs)
	testRunID := "test-run-123"

	// If we have seeded runs, use the first one
	if len(data) > 0 {
		if firstRun, ok := data[0].(map[string]interface{}); ok {
			if id, ok := firstRun["id"].(string); ok {
				testRunID = id
			}
		}
	}

	// GET /osm/api/runs/:id
	resp = apiGet(t, "/osm/api/runs/"+testRunID)
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/runs/:id should return 200")

	// GET /osm/api/runs/:id/steps
	resp = apiGet(t, "/osm/api/runs/"+testRunID+"/steps")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/runs/:id/steps should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain steps data")

	// GET /osm/api/runs/:id/artifacts
	resp = apiGet(t, "/osm/api/runs/"+testRunID+"/artifacts")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/runs/:id/artifacts should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain artifacts data")

	// POST /osm/api/runs - Create new run (dry-run mode)
	newRun := map[string]interface{}{
		"workflow_name": "test-bash",
		"target":        "test-run.example.com",
		"dry_run":       true,
	}
	resp = apiPost(t, "/osm/api/runs", newRun)
	// May return 201 (created) or 202 (accepted) or 400 (if workflow not found)
	// Accept 201, 202, or 400 as valid responses
	assert.True(t, resp.StatusCode == 201 || resp.StatusCode == 202 || resp.StatusCode == 400,
		"POST /osm/api/runs should return 201, 202, or 400 (workflow may not exist)")

	// DELETE /osm/api/runs/:id (cancel) - test with a test run ID
	resp = apiDelete(t, "/osm/api/runs/"+testRunID)
	// May return 200 (cancelled) or 400 (already completed/failed)
	assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 400,
		"DELETE /osm/api/runs/:id should return 200 or 400")

	log.Success("Run endpoints OK")
}

// testEventLogEndpoints tests event log endpoints
func testEventLogEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing event log endpoints")

	// GET /osm/api/event-logs
	resp := apiGet(t, "/osm/api/event-logs")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/event-logs should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data array")
	assert.Contains(t, body, "pagination", "Should contain pagination")

	// GET /osm/api/event-logs?workspace=example.com (filter)
	resp = apiGet(t, "/osm/api/event-logs?workspace=example.com")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/event-logs with filter should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data array")

	// GET /osm/api/event-logs with limit
	resp = apiGet(t, "/osm/api/event-logs?limit=5")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/event-logs with limit should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "data", "Should contain data array")

	log.Success("Event log endpoints OK")
}

// testFunctionEndpoints tests function evaluation endpoints
func testFunctionEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing function endpoints")

	// GET /osm/api/functions/list
	resp := apiGet(t, "/osm/api/functions/list")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/functions/list should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "functions", "Should contain functions list")
	functions, ok := body["functions"].([]interface{})
	assert.True(t, ok, "Functions should be an array")
	assert.NotEmpty(t, functions, "Should have available functions")

	// POST /osm/api/functions/eval - Evaluate a simple script
	// Note: Uses "script" field, not "expression"
	evalReq := map[string]interface{}{
		"script": "1 + 1",
	}
	resp = apiPost(t, "/osm/api/functions/eval", evalReq)
	assert.Equal(t, 200, resp.StatusCode, "POST /osm/api/functions/eval should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "result", "Should contain result")

	// POST /osm/api/functions/eval - Evaluate with target variable
	evalReq = map[string]interface{}{
		"script": "'hello ' + target",
		"target": "world",
	}
	resp = apiPost(t, "/osm/api/functions/eval", evalReq)
	assert.Equal(t, 200, resp.StatusCode, "POST /osm/api/functions/eval with target should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "result", "Should contain result")

	log.Success("Function endpoints OK")
}

// testSettingsEndpoints tests settings management endpoints
func testSettingsEndpoints(t *testing.T, log *TestLogger) {
	log.Info("Testing settings endpoints")

	// GET /osm/api/settings/yaml - returns raw YAML text, not JSON
	resp := apiGet(t, "/osm/api/settings/yaml")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/settings/yaml should return 200")
	// Read raw body - endpoint returns YAML text, not JSON
	bodyBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Should be able to read response body")
	yamlContent := string(bodyBytes)
	assert.NotEmpty(t, yamlContent, "Settings YAML should not be empty")
	// Verify it looks like YAML (contains common config keys)
	assert.Contains(t, yamlContent, "server:", "Should contain server config")

	log.Success("Settings endpoints OK")
}

// testWorkflowExtras tests additional workflow endpoints
func testWorkflowExtras(t *testing.T, log *TestLogger) {
	log.Info("Testing workflow extra endpoints")

	// GET /osm/api/workflows/tags
	resp := apiGet(t, "/osm/api/workflows/tags")
	assert.Equal(t, 200, resp.StatusCode, "GET /osm/api/workflows/tags should return 200")
	body := parseJSONResponse(t, resp)
	assert.Contains(t, body, "tags", "Should contain tags array")
	_, ok := body["tags"].([]interface{})
	assert.True(t, ok, "Tags should be an array")
	// May be empty if no workflows have tags, but should still be an array

	// POST /osm/api/workflows/refresh - Refresh workflow index
	resp = apiPost(t, "/osm/api/workflows/refresh", nil)
	assert.Equal(t, 200, resp.StatusCode, "POST /osm/api/workflows/refresh should return 200")
	body = parseJSONResponse(t, resp)
	assert.Contains(t, body, "message", "Should contain message")

	log.Success("Workflow extra endpoints OK")
}
