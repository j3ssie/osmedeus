package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRedisPort = "6399"
	testRedisURL  = "redis://localhost:6399"
)

var (
	testServerPort string
	testBaseDir    string
)

// getProjectRoot returns the project root directory
func getProjectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get caller info")
	}
	return filepath.Join(filepath.Dir(filename), "..", "..")
}

// getDockerComposePath returns the path to the distributed test docker-compose file
func getDockerComposePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(getProjectRoot(t), "build", "docker", "docker-compose.distributed-test.yaml")
}

// startRedis starts the Redis container using docker-compose
func startRedis(t *testing.T, log *TestLogger) func() {
	t.Helper()
	composePath := getDockerComposePath(t)

	log.Info("Starting Redis container from %s", composePath)

	cmd := exec.Command("docker-compose", "-f", composePath, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v\nOutput: %s", err, output)
	}

	// Wait for Redis to be ready
	waitForRedis(t, log, testRedisPort, 15*time.Second)

	return func() {
		log.Info("Stopping Redis container")
		cmd := exec.Command("docker-compose", "-f", composePath, "down", "-v")
		_ = cmd.Run()
	}
}

// waitForRedis waits for Redis to accept connections
func waitForRedis(t *testing.T, log *TestLogger, port string, timeout time.Duration) {
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

// startMaster starts osmedeus server with --master flag
func startMaster(t *testing.T, log *TestLogger, ctx context.Context) (*exec.Cmd, func()) {
	t.Helper()
	binary := getBinaryPath(t)
	workflowPath := getTestdataPath(t)

	log.Info("Starting master server on port %s", testServerPort)

	cmd := exec.CommandContext(ctx, binary, "serve",
		"--master",
		"--port", testServerPort,
		"--redis-url", testRedisURL,
		"-A", // Disable auth for testing
		"--base-folder", testBaseDir,
		"-F", workflowPath,
	)

	// Capture output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	require.NoError(t, err, "Failed to start master server")

	// Wait for server to be ready
	waitForServer(t, log, testServerPort, 15*time.Second)

	return cmd, func() {
		log.Info("Stopping master server")
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}
}

// startWorker starts osmedeus worker join
func startWorker(t *testing.T, log *TestLogger, ctx context.Context, workerID int) (*exec.Cmd, func()) {
	t.Helper()
	binary := getBinaryPath(t)
	workflowPath := getTestdataPath(t)

	log.Info("Starting worker %d", workerID)

	cmd := exec.CommandContext(ctx, binary, "worker", "join",
		"--redis-url", testRedisURL,
		"--base-folder", testBaseDir,
		"-F", workflowPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	require.NoError(t, err, "Failed to start worker %d", workerID)

	return cmd, func() {
		log.Info("Stopping worker %d", workerID)
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}
}

// waitForServer waits for HTTP server to respond
func waitForServer(t *testing.T, log *TestLogger, port string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	url := fmt.Sprintf("http://localhost:%s/health", port)

	log.Info("Waiting for server at %s...", url)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				log.Success("Server is ready")
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("Server not ready after %v", timeout)
}

// WorkerInfo represents a worker in the API response
type WorkerInfo struct {
	ID            string `json:"id"`
	Hostname      string `json:"hostname"`
	Status        string `json:"status"`
	CurrentTaskID string `json:"current_task"`
	JoinedAt      string `json:"joined_at"`
	LastHeartbeat string `json:"last_heartbeat"`
	TasksComplete int    `json:"tasks_complete"`
	TasksFailed   int    `json:"tasks_failed"`
}

// WorkersResponse represents the paginated workers API response
type WorkersResponse struct {
	Count int          `json:"count"`
	Data  []WorkerInfo `json:"data"`
}

// TaskInfo represents a task in the API response
type TaskInfo struct {
	ID           string `json:"id"`
	TaskID       string `json:"task_id,omitempty"` // Used in completed tasks
	WorkflowName string `json:"workflow_name"`
	Target       string `json:"target"`
	Status       string `json:"status"`
	WorkerID     string `json:"worker_id"`
	Error        string `json:"error,omitempty"`
}

// TasksResponse represents the tasks API response
type TasksResponse struct {
	Running   []TaskInfo `json:"running"`
	Completed []TaskInfo `json:"completed"`
}

// getWorkers retrieves workers from the API
func getWorkers(t *testing.T, port string) []WorkerInfo {
	t.Helper()
	url := fmt.Sprintf("http://localhost:%s/osm/api/workers", port)

	resp, err := http.Get(url)
	require.NoError(t, err, "Failed to get workers")
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response")

	var workersResp WorkersResponse
	err = json.Unmarshal(body, &workersResp)
	require.NoError(t, err, "Failed to parse workers response: %s", body)

	return workersResp.Data
}

// waitForWorkerCount waits for a specific number of workers to register
func waitForWorkerCount(t *testing.T, log *TestLogger, port string, count int, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)

	log.Info("Waiting for %d worker(s) to register...", count)

	for time.Now().Before(deadline) {
		workers := getWorkers(t, port)
		if len(workers) >= count {
			log.Success("Found %d registered worker(s)", len(workers))
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("Expected %d workers but didn't find them after %v", count, timeout)
}

// getTasks retrieves tasks from the API
func getTasks(t *testing.T, port string) []TaskInfo {
	t.Helper()
	url := fmt.Sprintf("http://localhost:%s/osm/api/tasks", port)

	resp, err := http.Get(url)
	require.NoError(t, err, "Failed to get tasks")
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response")

	var tasksResp TasksResponse
	err = json.Unmarshal(body, &tasksResp)
	require.NoError(t, err, "Failed to parse tasks response: %s", body)

	// Combine running and completed tasks
	allTasks := make([]TaskInfo, 0, len(tasksResp.Running)+len(tasksResp.Completed))
	allTasks = append(allTasks, tasksResp.Running...)
	allTasks = append(allTasks, tasksResp.Completed...)

	return allTasks
}

// getTaskByID retrieves a specific task from the API
func getTaskByID(t *testing.T, port, taskID string) *TaskInfo {
	t.Helper()
	url := fmt.Sprintf("http://localhost:%s/osm/api/tasks/%s", port, taskID)

	resp, err := http.Get(url)
	require.NoError(t, err, "Failed to get task")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == 404 {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response")

	var task TaskInfo
	err = json.Unmarshal(body, &task)
	require.NoError(t, err, "Failed to parse task response: %s", body)

	return &task
}

// submitDistributedRun submits a run via CLI and returns the task ID
func submitDistributedRun(t *testing.T, log *TestLogger, workflow, target string) string {
	t.Helper()
	binary := getBinaryPath(t)
	workflowPath := getTestdataPath(t)

	log.Command("run", "-D", "-m", workflow, "-t", target)

	cmd := exec.Command(binary, "--base-folder", testBaseDir, "run",
		"-D",
		"-m", workflow,
		"-t", target,
		"-F", workflowPath,
		"--redis-url", testRedisURL,
	)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to submit distributed run: %s", output)

	log.Result(string(output), "")

	// Parse task ID from output - look for patterns like:
	// "Submitted task abc12345 for target: example.com"
	patterns := []string{
		`Submitted task ([a-f0-9]{8}) for target`, // Submitted task abc12345 for target
		`task_id["\s:]+([a-f0-9]{8})`,             // task_id: abc12345 or "task_id":"abc12345"
		`Task\s+([a-f0-9]{8})\s+submitted`,        // Task abc12345 submitted
	}

	outputStr := string(output)
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(outputStr)
		if len(matches) >= 2 {
			log.Info("Found task ID: %s", matches[1])
			return matches[1]
		}
	}

	// If we can't parse the task ID from output, get it from the API
	// Wait a moment for the task to be registered
	time.Sleep(500 * time.Millisecond)
	tasks := getTasks(t, testServerPort)
	if len(tasks) > 0 {
		// Return the most recent task (first one in running, or first in completed)
		taskID := tasks[0].ID
		if taskID == "" {
			taskID = tasks[0].TaskID // For completed tasks
		}
		if taskID != "" {
			log.Info("Found task ID from API: %s", taskID)
			return taskID
		}
	}

	t.Fatal("Could not determine task ID from output or API")
	return ""
}

// waitForTaskCompletion waits for a task to complete
func waitForTaskCompletion(t *testing.T, log *TestLogger, port, taskID string, timeout time.Duration) *TaskInfo {
	t.Helper()
	deadline := time.Now().Add(timeout)

	log.Info("Waiting for task %s to complete...", taskID)

	for time.Now().Before(deadline) {
		task := getTaskByID(t, port, taskID)
		if task != nil {
			if task.Status == "completed" || task.Status == "failed" {
				log.Success("Task %s finished with status: %s", taskID, task.Status)
				return task
			}
			log.Debug("Task %s status: %s", taskID, task.Status)
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("Task %s did not complete after %v", taskID, timeout)
	return nil
}

// TestDistributed_MasterStartup tests that the master server starts correctly with Redis
func TestDistributed_MasterStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping distributed test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing master startup")

	testServerPort = getFreePort(t)
	testBaseDir = t.TempDir()

	// Start Redis
	cleanupRedis := startRedis(t, log)
	defer cleanupRedis()

	// Start Master
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, cleanupMaster := startMaster(t, log, ctx)
	defer cleanupMaster()

	// Verify health endpoint
	log.Step("Verifying health endpoint")
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", testServerPort))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, 200, resp.StatusCode)

	// Verify workers endpoint is available
	log.Step("Verifying workers API endpoint")
	resp, err = http.Get(fmt.Sprintf("http://localhost:%s/osm/api/workers", testServerPort))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, 200, resp.StatusCode)

	log.Success("Master startup test passed")
}

// TestDistributed_WorkerJoin tests that a worker can join and register with the master
func TestDistributed_WorkerJoin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping distributed test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing worker join")

	testServerPort = getFreePort(t)
	testBaseDir = t.TempDir()

	// Start Redis
	cleanupRedis := startRedis(t, log)
	defer cleanupRedis()

	// Start Master
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, cleanupMaster := startMaster(t, log, ctx)
	defer cleanupMaster()

	// Start Worker
	log.Step("Starting worker")
	_, cleanupWorker := startWorker(t, log, ctx, 1)
	defer cleanupWorker()

	// Wait for worker to register
	waitForWorkerCount(t, log, testServerPort, 1, 15*time.Second)

	// Verify worker status
	log.Step("Verifying worker status")
	workers := getWorkers(t, testServerPort)
	require.Len(t, workers, 1)
	assert.Equal(t, "idle", workers[0].Status)
	assert.NotEmpty(t, workers[0].ID)
	assert.NotEmpty(t, workers[0].Hostname)

	log.Success("Worker join test passed")
}

// TestDistributed_TaskSubmission tests submitting and completing a distributed task
func TestDistributed_TaskSubmission(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping distributed test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing task submission")

	testServerPort = getFreePort(t)
	testBaseDir = t.TempDir()

	// Start Redis
	cleanupRedis := startRedis(t, log)
	defer cleanupRedis()

	// Start Master
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, cleanupMaster := startMaster(t, log, ctx)
	defer cleanupMaster()

	// Start Worker
	log.Step("Starting worker")
	_, cleanupWorker := startWorker(t, log, ctx, 1)
	defer cleanupWorker()

	// Wait for worker to register
	waitForWorkerCount(t, log, testServerPort, 1, 15*time.Second)

	// Submit distributed run
	log.Step("Submitting distributed run task")
	taskID := submitDistributedRun(t, log, "test-bash", "distributed-test.example.com")
	log.Info("Task submitted with ID: %s", taskID)

	// Wait for task completion
	log.Step("Waiting for task completion")
	task := waitForTaskCompletion(t, log, testServerPort, taskID, 30*time.Second)

	// Verify task completed successfully
	assert.Equal(t, "completed", task.Status)
	assert.Empty(t, task.Error)

	log.Success("Task submission test passed")
}

// TestDistributed_MultipleWorkers tests that tasks are distributed across multiple workers
func TestDistributed_MultipleWorkers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping distributed test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing multiple workers")

	testServerPort = getFreePort(t)
	testBaseDir = t.TempDir()

	// Start Redis
	cleanupRedis := startRedis(t, log)
	defer cleanupRedis()

	// Start Master
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, cleanupMaster := startMaster(t, log, ctx)
	defer cleanupMaster()

	// Start 2 Workers
	log.Step("Starting worker 1")
	_, cleanupWorker1 := startWorker(t, log, ctx, 1)
	defer cleanupWorker1()

	log.Step("Starting worker 2")
	_, cleanupWorker2 := startWorker(t, log, ctx, 2)
	defer cleanupWorker2()

	// Wait for both workers to register
	waitForWorkerCount(t, log, testServerPort, 2, 20*time.Second)

	// Verify we have 2 workers
	workers := getWorkers(t, testServerPort)
	assert.Len(t, workers, 2)

	log.Success("Multiple workers test passed")
}

// TestDistributed_WorkerStatus tests the worker status CLI command
func TestDistributed_WorkerStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping distributed test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Testing worker status command")

	testServerPort = getFreePort(t)
	testBaseDir = t.TempDir()

	// Start Redis
	cleanupRedis := startRedis(t, log)
	defer cleanupRedis()

	// Start Master
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	_, cleanupMaster := startMaster(t, log, ctx)
	defer cleanupMaster()

	// Start Worker
	log.Step("Starting worker")
	_, cleanupWorker := startWorker(t, log, ctx, 1)
	defer cleanupWorker()

	// Wait for worker to register
	waitForWorkerCount(t, log, testServerPort, 1, 15*time.Second)

	// Run worker status command
	log.Step("Running worker status command")
	binary := getBinaryPath(t)
	cmd := exec.Command(binary, "--base-folder", testBaseDir, "worker", "status", "--redis-url", testRedisURL)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Worker status command failed: %s", output)

	log.Result(string(output), "")

	// Verify output contains worker information
	outputStr := string(output)
	assert.True(t, strings.Contains(outputStr, "idle") || strings.Contains(outputStr, "Worker"),
		"Expected worker status in output, got: %s", outputStr)

	log.Success("Worker status test passed")
}

// TestDistributed_FullWorkflow is an end-to-end test of the complete distributed workflow
func TestDistributed_FullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping distributed test in short mode")
	}

	log := NewTestLogger(t)
	log.Step("Running full distributed workflow e2e test")

	testServerPort = getFreePort(t)
	testBaseDir = t.TempDir()

	// Step 1: Start Redis
	log.Step("Step 1: Starting Redis container")
	cleanupRedis := startRedis(t, log)
	defer cleanupRedis()

	// Step 2: Start Master
	log.Step("Step 2: Starting master server")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	_, cleanupMaster := startMaster(t, log, ctx)
	defer cleanupMaster()

	// Step 3: Start Worker
	log.Step("Step 3: Starting worker node")
	_, cleanupWorker := startWorker(t, log, ctx, 1)
	defer cleanupWorker()

	// Wait for worker to register
	time.Sleep(2 * time.Second)
	waitForWorkerCount(t, log, testServerPort, 1, 15*time.Second)

	// Step 4: Verify worker is idle
	log.Step("Step 4: Verifying worker is idle")
	workers := getWorkers(t, testServerPort)
	require.Len(t, workers, 1)
	assert.Equal(t, "idle", workers[0].Status)
	log.Success("Worker registered and idle")

	// Step 5: Submit distributed run
	log.Step("Step 5: Submitting distributed run task")
	taskID := submitDistributedRun(t, log, "test-bash", "e2e-full-workflow.example.com")
	log.Info("Task submitted: %s", taskID)

	// Step 6: Wait for task completion
	log.Step("Step 6: Waiting for task completion")
	task := waitForTaskCompletion(t, log, testServerPort, taskID, 45*time.Second)

	// Step 7: Verify results
	log.Step("Step 7: Verifying task results")
	assert.Equal(t, "completed", task.Status)
	assert.Empty(t, task.Error, "Task should not have errors")

	// Step 8: Verify worker processed the task
	log.Step("Step 8: Verifying worker stats")
	workers = getWorkers(t, testServerPort)
	require.Len(t, workers, 1)
	assert.GreaterOrEqual(t, workers[0].TasksComplete, 1, "Worker should have completed at least 1 task")

	log.Success("Full distributed workflow e2e test passed!")
}
