package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	buildBinaryOnce sync.Once
	buildBinaryErr  error
)

// ANSI color codes matching project logger style (internal/logger/logger.go)
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorGrey    = "\033[90m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
)

// getBinaryPath returns the path to the osmedeus binary
func getBinaryPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get caller info")
	}
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	binary := filepath.Join(projectRoot, "build", "bin", "osmedeus")

	buildBinaryOnce.Do(func() {
		_ = os.MkdirAll(filepath.Dir(binary), 0755)
		cmd := exec.Command("go", "build", "-o", binary, "./cmd/osmedeus")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			buildBinaryErr = fmt.Errorf("failed to build osmedeus binary: %w\n%s", err, string(output))
			return
		}
		if _, err := os.Stat(binary); err != nil {
			buildBinaryErr = fmt.Errorf("binary not found after build: %w", err)
		}
	})

	if buildBinaryErr != nil {
		t.Fatal(buildBinaryErr)
	}
	return binary
}

// getTestdataPath returns the path to test workflow fixtures
func getTestdataPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get caller info")
	}
	return filepath.Join(filepath.Dir(filename), "..", "testdata", "workflows")
}

// TestLogger provides verbose logging for E2E tests matching project style
type TestLogger struct {
	t        *testing.T
	testName string
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) *TestLogger {
	t.Helper()
	return &TestLogger{
		t:        t,
		testName: t.Name(),
	}
}

// formatTimestamp returns grey-colored ISO 8601 timestamp
func formatTimestamp() string {
	return colorGrey + time.Now().Format("2006-01-02T15:04:05-07:00") + colorReset
}

// formatLevel returns bold+colored level string
func formatLevel(level string) string {
	var color string
	switch level {
	case "DEBUG":
		color = colorMagenta
	case "INFO":
		color = colorCyan
	case "WARN":
		color = colorYellow
	case "ERROR":
		color = colorRed
	default:
		color = ""
	}
	return colorBold + color + level + colorReset
}

// log formats and outputs a log message
func (l *TestLogger) log(level, format string, args ...interface{}) {
	l.t.Helper()
	msg := fmt.Sprintf(format, args...)
	l.t.Logf("%s %s %s", formatTimestamp(), formatLevel(level), msg)
}

// Debug logs a debug message
func (l *TestLogger) Debug(format string, args ...interface{}) {
	l.t.Helper()
	l.log("DEBUG", format, args...)
}

// Info logs an info message
func (l *TestLogger) Info(format string, args ...interface{}) {
	l.t.Helper()
	l.log("INFO", format, args...)
}

// Warn logs a warning message
func (l *TestLogger) Warn(format string, args ...interface{}) {
	l.t.Helper()
	l.log("WARN", format, args...)
}

// Error logs an error message
func (l *TestLogger) Error(format string, args ...interface{}) {
	l.t.Helper()
	l.log("ERROR", format, args...)
}

// Step logs the start of a test step
func (l *TestLogger) Step(stepName string) {
	l.t.Helper()
	l.log("INFO", "==> Step: %s", stepName)
}

// Command logs a CLI command being executed (in blue)
func (l *TestLogger) Command(args ...string) {
	l.t.Helper()
	cmd := "osmedeus " + strings.Join(args, " ")
	l.t.Logf("%s %s %s", formatTimestamp(), formatLevel("DEBUG"),
		colorBlue+colorBold+"$ "+cmd+colorReset)
}

// Result logs command output (stdout in green, stderr in yellow)
func (l *TestLogger) Result(stdout, stderr string) {
	l.t.Helper()
	if stdout != "" {
		// Truncate long output for readability
		out := strings.TrimSpace(stdout)
		if len(out) > 200 {
			out = out[:200] + "..."
		}
		l.t.Logf("%s %s %s", formatTimestamp(), formatLevel("DEBUG"),
			colorGreen+"stdout: "+out+colorReset)
	}
	if stderr != "" {
		out := strings.TrimSpace(stderr)
		if len(out) > 200 {
			out = out[:200] + "..."
		}
		l.t.Logf("%s %s %s", formatTimestamp(), formatLevel("WARN"),
			colorYellow+"stderr: "+out+colorReset)
	}
}

// Success logs a success message (in green)
func (l *TestLogger) Success(format string, args ...interface{}) {
	l.t.Helper()
	msg := fmt.Sprintf(format, args...)
	l.t.Logf("%s %s %s", formatTimestamp(), formatLevel("INFO"),
		colorGreen+colorBold+"✓ "+msg+colorReset)
}

// runCLIWithLog executes the CLI with given args and logs verbose output
func runCLIWithLog(t *testing.T, log *TestLogger, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	binary := getBinaryPath(t)
	baseDir := t.TempDir()
	args = append([]string{"--base-folder", baseDir}, args...)

	log.Command(args...)

	cmd := exec.Command(binary, args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	log.Result(stdout, stderr)

	if err != nil {
		log.Error("Command failed: %v", err)
	}

	return stdout, stderr, err
}

func runCLIWithLogAndBase(t *testing.T, log *TestLogger, args ...string) (baseDir, stdout, stderr string, err error) {
	t.Helper()
	binary := getBinaryPath(t)
	baseDir = t.TempDir()
	args = append([]string{"--base-folder", baseDir}, args...)

	log.Command(args...)

	cmd := exec.Command(binary, args...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err = cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	log.Result(stdout, stderr)

	if err != nil {
		log.Error("Command failed: %v", err)
	}

	return baseDir, stdout, stderr, err
}
