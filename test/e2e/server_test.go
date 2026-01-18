package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_Help(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing server help command")

	stdout, _, err := runCLIWithLog(t, log, "serve", "--help")
	require.NoError(t, err)

	log.Info("Asserting stdout contains server options")
	assert.Contains(t, stdout, "--port")
	assert.Contains(t, stdout, "--host")

	log.Success("server help displays all required options")
}

func TestServer_StartStop(t *testing.T) {
	log := NewTestLogger(t)
	log.Step("Testing server start and stop")

	if testing.Short() {
		log.Info("Skipping server integration test in short mode")
		t.Skip("skipping server integration test in short mode")
	}

	binary := getBinaryPath(t)
	port := "19999" // Use a high port to avoid conflicts
	log.Info("Using port: %s", port)

	// Start server in background
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("Starting server with command: %s serve --port %s -A", binary, port)
	cmd := exec.CommandContext(ctx, binary, "serve", "--port", port, "-A")
	err := cmd.Start()
	require.NoError(t, err)

	defer func() {
		if cmd.Process != nil {
			log.Info("Killing server process")
			_ = cmd.Process.Kill()
		}
	}()

	// Wait for server to start
	log.Info("Waiting 2 seconds for server to start")
	time.Sleep(2 * time.Second)

	// Test health endpoint
	healthURL := fmt.Sprintf("http://localhost:%s/health", port)
	log.Info("Testing health endpoint: %s", healthURL)
	resp, err := http.Get(healthURL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	log.Info("Asserting response status is 200 OK")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	log.Success("server started and health endpoint responded")
}
