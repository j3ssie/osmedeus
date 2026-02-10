package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── constants ────────────────────────────────────────────────────────────────

const (
	canaryContainerName = "osm-canary"
	canaryAPIBase       = "http://localhost:8002"
	canaryWorkspaceRoot = "/root/workspaces-osmedeus"
)

// ── compose helpers ──────────────────────────────────────────────────────────

func getCanaryComposePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(getProjectRoot(t), "build", "docker", "docker-compose.canary.yaml")
}

// startCanaryContainer builds the canary image and starts the container.
// It returns a cleanup function that tears everything down.
func startCanaryContainer(t *testing.T, log *TestLogger) func() {
	t.Helper()
	composePath := getCanaryComposePath(t)

	log.Step("Building canary Docker image (this may take a while)")
	cmd := exec.Command("docker-compose", "-f", composePath, "build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build canary image: %v\nOutput: %s", err, output)
	}
	log.Success("Canary image built")

	log.Step("Starting canary container")
	cmd = exec.Command("docker-compose", "-f", composePath, "up", "-d")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start canary container: %v\nOutput: %s", err, output)
	}
	log.Success("Canary container started")

	return func() {
		log.Info("Tearing down canary container")
		cmd := exec.Command("docker-compose", "-f", composePath, "down", "-v")
		_ = cmd.Run()
	}
}

// waitForCanaryAPI polls the health endpoint until the API is ready.
func waitForCanaryAPI(t *testing.T, log *TestLogger, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	url := canaryAPIBase + "/health"

	log.Info("Waiting for canary API at %s ...", url)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				log.Success("Canary API is ready")
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("Canary API not ready after %v", timeout)
}

// ── docker exec helpers ──────────────────────────────────────────────────────

// dockerExec runs a command inside the canary container with a 2-minute timeout.
func dockerExec(t *testing.T, log *TestLogger, args ...string) (string, error) {
	t.Helper()
	return dockerExecLong(t, log, 2*time.Minute, args...)
}

// dockerExecLong runs a command inside the canary container with a custom timeout.
func dockerExecLong(t *testing.T, log *TestLogger, timeout time.Duration, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmdArgs := append([]string{"exec", canaryContainerName}, args...)
	log.Debug("docker %s", strings.Join(cmdArgs, " "))

	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	output, err := cmd.CombinedOutput()
	out := string(output)

	if len(out) > 500 {
		log.Debug("output (%d bytes): %s...", len(out), out[:500])
	} else if out != "" {
		log.Debug("output: %s", strings.TrimSpace(out))
	}

	return out, err
}

// ── filesystem checks inside container ───────────────────────────────────────

func fileExistsInContainer(t *testing.T, log *TestLogger, path string) bool {
	t.Helper()
	_, err := dockerExec(t, log, "test", "-f", path)
	return err == nil
}

func dirExistsInContainer(t *testing.T, log *TestLogger, path string) bool {
	t.Helper()
	_, err := dockerExec(t, log, "test", "-d", path)
	return err == nil
}

// findFilesInContainer returns files matching a pattern under dir.
func findFilesInContainer(t *testing.T, log *TestLogger, dir, pattern string) []string {
	t.Helper()
	out, err := dockerExec(t, log, "find", dir, "-name", pattern, "-type", "f")
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

// ── API helpers ──────────────────────────────────────────────────────────────

// canaryAPIGet performs a GET request against the canary API and parses JSON.
func canaryAPIGet(t *testing.T, path string) map[string]any {
	t.Helper()
	resp, err := http.Get(canaryAPIBase + path)
	require.NoError(t, err, "GET %s failed", path)
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read response for %s", path)

	var result map[string]any
	err = json.Unmarshal(body, &result)
	require.NoError(t, err, "failed to parse JSON from %s: %s", path, string(body))
	return result
}

// getRunsForWorkspace returns runs filtered by workspace name.
func getRunsForWorkspace(t *testing.T, ws string) []any {
	t.Helper()
	result := canaryAPIGet(t, "/osm/api/runs?workspace="+ws)
	data, _ := result["data"].([]any)
	return data
}

// getAssetsForWorkspace returns assets filtered by workspace name.
func getAssetsForWorkspace(t *testing.T, ws string) []any {
	t.Helper()
	result := canaryAPIGet(t, "/osm/api/assets?workspace="+ws)
	data, _ := result["data"].([]any)
	return data
}

// getVulnsForWorkspace returns vulnerabilities filtered by workspace name.
func getVulnsForWorkspace(t *testing.T, ws string) []any {
	t.Helper()
	result := canaryAPIGet(t, "/osm/api/vulnerabilities?workspace="+ws)
	data, _ := result["data"].([]any)
	return data
}

// getWorkspaces returns all workspace records.
func getWorkspaces(t *testing.T) []map[string]any {
	t.Helper()
	result := canaryAPIGet(t, "/osm/api/workspaces")
	raw, _ := result["data"].([]any)
	var out []map[string]any
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// findWorkspaceByName searches all workspaces for one matching name (exact or substring).
func findWorkspaceByName(t *testing.T, name string) map[string]any {
	t.Helper()
	for _, ws := range getWorkspaces(t) {
		wsName, _ := ws["name"].(string)
		if wsName == name || strings.Contains(wsName, name) {
			return ws
		}
	}
	return nil
}

// ── main test entry points ───────────────────────────────────────────────────

// requireCanaryAPI checks that the canary container API is reachable.
// Individual tests (TestCanary_Repo, etc.) assume the container was started
// externally by the Makefile, so they only need to verify the API is up.
func requireCanaryAPI(t *testing.T, log *TestLogger) {
	t.Helper()
	waitForCanaryAPI(t, log, 60*time.Second)
}

// TestCanary_FullSuite manages the full lifecycle: build, start, run all 3 scans, cleanup.
// Use `make test-canary-all` to run this.
func TestCanary_FullSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canary tests in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Starting canary test suite")

	// Start the canary container (build + up)
	cleanup := startCanaryContainer(t, log)
	defer cleanup()

	// Wait for the API server inside the container to be ready
	waitForCanaryAPI(t, log, 90*time.Second)

	// Run the three canary scans sequentially
	t.Run("RepoScan", func(t *testing.T) {
		testCanaryRepoScan(t)
	})
	t.Run("DomainLiteScan", func(t *testing.T) {
		testCanaryDomainLiteScan(t)
	})
	t.Run("CIDRScan", func(t *testing.T) {
		testCanaryCIDRScan(t)
	})

	log.Success("All canary tests completed")
}

// TestCanary_Repo runs only the repo scan canary test.
// Assumes the canary container is already running (e.g. via `make test-canary-repo`).
func TestCanary_Repo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canary tests in short mode")
	}
	log := NewTestLogger(t)
	requireCanaryAPI(t, log)
	testCanaryRepoScan(t)
}

// TestCanary_Domain runs only the domain-lite scan canary test.
// Assumes the canary container is already running (e.g. via `make test-canary-domain`).
func TestCanary_Domain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canary tests in short mode")
	}
	log := NewTestLogger(t)
	requireCanaryAPI(t, log)
	testCanaryDomainLiteScan(t)
}

// TestCanary_CIDR runs only the CIDR scan canary test.
// Assumes the canary container is already running (e.g. via `make test-canary-ip`).
func TestCanary_CIDR(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping canary tests in short mode")
	}
	log := NewTestLogger(t)
	requireCanaryAPI(t, log)
	testCanaryCIDRScan(t)
}

// ── Test 1: Repo Scan ────────────────────────────────────────────────────────

func testCanaryRepoScan(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Canary: Repo Scan (juice-shop)")

	targetURL := "https://github.com/juice-shop/juice-shop/archive/refs/heads/master.zip"
	expectedWS := "github.com"
	wsDir := canaryWorkspaceRoot + "/" + expectedWS

	// Run the scan
	log.Step("Running repo flow")
	out, err := dockerExecLong(t, log, 25*time.Minute,
		"osmedeus", "run", "-f", "repo", "-t", targetURL)
	if err != nil {
		log.Error("Repo scan command error: %v\nOutput: %s", err, out)
	}
	// Don't require.NoError — some steps may fail yet still produce results

	// ── Filesystem checks ────────────────────────────────────────────────
	log.Step("Verifying filesystem artifacts")

	assert.True(t, dirExistsInContainer(t, log, wsDir),
		"workspace directory %s should exist", wsDir)

	sarifFiles := findFilesInContainer(t, log, wsDir, "*.sarif")
	log.Info("Found %d SARIF files", len(sarifFiles))
	assert.NotEmpty(t, sarifFiles, "expected SARIF output files in workspace")

	mdFiles := findFilesInContainer(t, log, wsDir, "*.md")
	log.Info("Found %d markdown reports", len(mdFiles))

	// ── Database / API checks ────────────────────────────────────────────
	log.Step("Verifying database records via API")

	// Runs
	runs := getRunsForWorkspace(t, expectedWS)
	if len(runs) == 0 {
		// Fallback: search all workspaces
		ws := findWorkspaceByName(t, expectedWS)
		if ws != nil {
			wsName, _ := ws["name"].(string)
			runs = getRunsForWorkspace(t, wsName)
			log.Info("Fallback workspace name: %s", wsName)
		}
	}
	assert.NotEmpty(t, runs, "expected at least 1 run for workspace %s", expectedWS)

	if len(runs) > 0 {
		firstRun, _ := runs[0].(map[string]any)
		runUUID, _ := firstRun["run_uuid"].(string)
		if runUUID != "" {
			// Steps
			stepsResp := canaryAPIGet(t, fmt.Sprintf("/osm/api/runs/%s/steps", runUUID))
			stepsData, _ := stepsResp["data"].([]any)
			assert.NotEmpty(t, stepsData, "expected step results for run %s", runUUID)
			log.Info("Run %s has %d step results", runUUID, len(stepsData))

			// Artifacts
			artifactsResp := canaryAPIGet(t, fmt.Sprintf("/osm/api/runs/%s/artifacts", runUUID))
			artifactsData, _ := artifactsResp["data"].([]any)
			log.Info("Run %s has %d artifacts", runUUID, len(artifactsData))
		}
	}

	// Vulnerabilities
	vulns := getVulnsForWorkspace(t, expectedWS)
	log.Info("Found %d vulnerabilities for workspace %s", len(vulns), expectedWS)
	assert.NotEmpty(t, vulns, "juice-shop should produce vulnerability findings")

	// Workspace record
	ws := findWorkspaceByName(t, expectedWS)
	assert.NotNil(t, ws, "workspace record should exist in DB")
	if ws != nil {
		totalVulns, _ := ws["total_vulns"].(float64)
		log.Info("Workspace total_vulns: %.0f", totalVulns)
		assert.Greater(t, totalVulns, float64(0), "workspace should have total_vulns > 0")
	}

	log.Success("Repo scan canary passed")
}

// ── Test 2: Domain-Lite Scan ─────────────────────────────────────────────────

func testCanaryDomainLiteScan(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Canary: Domain-Lite Scan (hackerone.com)")

	target := "hackerone.com"
	expectedWS := "hackerone.com"
	wsDir := canaryWorkspaceRoot + "/" + expectedWS

	// Run the scan
	log.Step("Running domain-lite flow")
	out, err := dockerExecLong(t, log, 20*time.Minute,
		"osmedeus", "run", "-f", "domain-lite", "-t", target)
	if err != nil {
		log.Error("Domain-lite scan command error: %v\nOutput: %s", err, out)
	}

	// ── Filesystem checks ────────────────────────────────────────────────
	log.Step("Verifying filesystem artifacts")

	assert.True(t, dirExistsInContainer(t, log, wsDir),
		"workspace directory %s should exist", wsDir)

	txtFiles := findFilesInContainer(t, log, wsDir, "*.txt")
	log.Info("Found %d .txt files (subdomains, etc.)", len(txtFiles))
	assert.NotEmpty(t, txtFiles, "expected subdomain/output text files")

	jsonlFiles := findFilesInContainer(t, log, wsDir, "*.jsonl")
	log.Info("Found %d .jsonl files (fingerprints, etc.)", len(jsonlFiles))

	// ── Database / API checks ────────────────────────────────────────────
	log.Step("Verifying database records via API")

	// Runs
	runs := getRunsForWorkspace(t, expectedWS)
	if len(runs) == 0 {
		ws := findWorkspaceByName(t, expectedWS)
		if ws != nil {
			wsName, _ := ws["name"].(string)
			runs = getRunsForWorkspace(t, wsName)
			log.Info("Fallback workspace name: %s", wsName)
		}
	}
	assert.NotEmpty(t, runs, "expected at least 1 run for workspace %s", expectedWS)

	// Assets
	assets := getAssetsForWorkspace(t, expectedWS)
	log.Info("Found %d assets for workspace %s", len(assets), expectedWS)
	assert.NotEmpty(t, assets, "expected discovered subdomains/HTTP endpoints")

	// Workspace record
	ws := findWorkspaceByName(t, expectedWS)
	assert.NotNil(t, ws, "workspace record should exist in DB")

	log.Success("Domain-lite scan canary passed")
}

// ── Test 3: CIDR Scan ────────────────────────────────────────────────────────

func testCanaryCIDRScan(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Canary: CIDR Scan (IP list)")

	expectedWS := "list-of-ips-file"
	wsDir := canaryWorkspaceRoot + "/" + expectedWS

	// Create the IP list inside the container.
	// NOTE: These are sample/public IP addresses used solely for testing the
	// scan workflow pipeline (workspace creation, step execution, DB writes).
	// There is no intent to attack or exploit any of these hosts.
	log.Step("Creating target IP list")
	ipList := strings.Join([]string{
		"102.88.154.187",
		"8.222.222.78",
		"93.184.216.34",
		"104.16.132.229",
	}, "\\n")
	_, err := dockerExec(t, log, "sh", "-c",
		fmt.Sprintf("printf '%s\\n' > /tmp/list-of-ips.txt", ipList))
	require.NoError(t, err, "failed to create IP list in container")

	// Verify file was created
	assert.True(t, fileExistsInContainer(t, log, "/tmp/list-of-ips.txt"),
		"IP list file should exist in container")

	// Run the scan
	log.Step("Running cidr flow")
	out, err := dockerExecLong(t, log, 25*time.Minute,
		"osmedeus", "run", "-f", "cidr", "-t", "/tmp/list-of-ips.txt")
	if err != nil {
		log.Error("CIDR scan command error: %v\nOutput: %s", err, out)
	}

	// ── Filesystem checks ────────────────────────────────────────────────
	log.Step("Verifying filesystem artifacts")

	assert.True(t, dirExistsInContainer(t, log, wsDir),
		"workspace directory %s should exist", wsDir)

	outputFiles := findFilesInContainer(t, log, wsDir, "*")
	log.Info("Found %d output files in workspace", len(outputFiles))

	// ── Database / API checks ────────────────────────────────────────────
	log.Step("Verifying database records via API")

	// Runs
	runs := getRunsForWorkspace(t, expectedWS)
	if len(runs) == 0 {
		ws := findWorkspaceByName(t, expectedWS)
		if ws != nil {
			wsName, _ := ws["name"].(string)
			runs = getRunsForWorkspace(t, wsName)
			log.Info("Fallback workspace name: %s", wsName)
		}
	}
	assert.NotEmpty(t, runs, "expected at least 1 run for workspace %s", expectedWS)

	// Workspace record
	ws := findWorkspaceByName(t, expectedWS)
	assert.NotNil(t, ws, "workspace record should exist in DB")

	log.Success("CIDR scan canary passed")
}
