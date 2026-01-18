package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-getter/v2"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// GetterClient wraps go-getter with Osmedeus configuration
type GetterClient struct {
	customHeaders map[string]string
	timeout       time.Duration
}

// GetterOption configures the GetterClient
type GetterOption func(*GetterClient)

// WithCustomHeaders sets custom HTTP headers for downloads
func WithCustomHeaders(headers map[string]string) GetterOption {
	return func(gc *GetterClient) {
		gc.customHeaders = headers
	}
}

// WithTimeout sets the download timeout
func WithTimeout(timeout time.Duration) GetterOption {
	return func(gc *GetterClient) {
		gc.timeout = timeout
	}
}

// NewGetterClient creates a configured go-getter client
func NewGetterClient(opts ...GetterOption) *GetterClient {
	gc := &GetterClient{
		timeout:       10 * time.Minute,
		customHeaders: make(map[string]string),
	}

	for _, opt := range opts {
		opt(gc)
	}

	return gc
}

// buildHTTPGetter creates an HTTP getter with authentication headers
func (gc *GetterClient) buildHTTPGetter() *getter.HttpGetter {
	headers := http.Header{}
	headers.Set("User-Agent", core.DefaultUA)

	// Add GitHub token for private repos
	if token := getGitHubToken(); token != "" {
		headers.Set("Authorization", "Bearer "+token)
	}

	// Add custom headers
	for k, v := range gc.customHeaders {
		headers.Set(k, v)
	}

	return &getter.HttpGetter{
		Header:      headers,
		ReadTimeout: gc.timeout,
	}
}

// buildGetters returns the list of getters to use
func (gc *GetterClient) buildGetters() []getter.Getter {
	return []getter.Getter{
		gc.buildHTTPGetter(),
		new(getter.GitGetter),
		new(getter.FileGetter),
	}
}

// Get downloads from source to destination with auto-extraction
func (gc *GetterClient) Get(ctx context.Context, dst, src string) (*getter.GetResult, error) {
	// Inject GitHub auth for git URLs
	src = gc.injectGitHubAuth(src)

	client := &getter.Client{
		Getters:       gc.buildGetters(),
		Decompressors: getter.Decompressors,
	}

	req := &getter.Request{
		Src:     src,
		Dst:     dst,
		GetMode: getter.ModeAny,
	}

	return client.Get(ctx, req)
}

// GetFile downloads a single file (no directory support)
func (gc *GetterClient) GetFile(ctx context.Context, dst, src string) (*getter.GetResult, error) {
	// Ensure destination directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	client := &getter.Client{
		Getters:       gc.buildGetters(),
		Decompressors: map[string]getter.Decompressor{}, // No decompression for file mode
	}

	req := &getter.Request{
		Src:     src,
		Dst:     dst,
		GetMode: getter.ModeFile,
	}

	return client.Get(ctx, req)
}

// GetDir downloads and extracts to a directory
func (gc *GetterClient) GetDir(ctx context.Context, dst, src string) (*getter.GetResult, error) {
	// Ensure destination directory exists
	if err := os.MkdirAll(dst, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dst, err)
	}

	// Inject GitHub auth for git URLs
	src = gc.injectGitHubAuth(src)

	client := &getter.Client{
		Getters:       gc.buildGetters(),
		Decompressors: getter.Decompressors,
	}

	req := &getter.Request{
		Src:     src,
		Dst:     dst,
		GetMode: getter.ModeDir,
	}

	return client.Get(ctx, req)
}

// GetWithRetry downloads with automatic retry on transient failures
func (gc *GetterClient) GetWithRetry(ctx context.Context, dst, src string, maxRetries int) (*getter.GetResult, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s...
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			logger.Get().Info("Retrying download",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
				zap.String("url", src))

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err := gc.Get(ctx, dst, src)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on certain errors (e.g., 404, auth failures)
		if isNonRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("download failed after %d retries: %w", maxRetries, lastErr)
}

// injectGitHubAuth transforms GitHub URLs to include authentication token
func (gc *GetterClient) injectGitHubAuth(src string) string {
	token := getGitHubToken()
	if token == "" {
		return src
	}

	// For git:: protocol URLs, inject token
	if strings.HasPrefix(src, "git::https://github.com") {
		return strings.Replace(src,
			"git::https://github.com",
			fmt.Sprintf("git::https://%s@github.com", token), 1)
	}

	// For raw GitHub URLs used with git getter
	if strings.Contains(src, "github.com") && !strings.Contains(src, "@github.com") {
		// Only inject for URLs that look like git repos (not release downloads)
		if strings.HasSuffix(src, ".git") || (!strings.Contains(src, "/releases/") && !strings.Contains(src, "/archive/")) {
			return strings.Replace(src, "github.com", fmt.Sprintf("%s@github.com", token), 1)
		}
	}

	return src
}

// Note: getGitHubToken is defined in download.go and reused here

// isNonRetryableError checks if an error should not be retried
func isNonRetryableError(err error) bool {
	errStr := err.Error()
	// Non-retryable HTTP status codes
	nonRetryable := []string{
		"404", "Not Found",
		"403", "Forbidden",
		"401", "Unauthorized",
	}
	for _, s := range nonRetryable {
		if strings.Contains(errStr, s) {
			return true
		}
	}
	return false
}

// GoGetterInstallOutput stores the last go-getter install output for display
var GoGetterInstallOutput string

// GetViaGoGetter downloads/clones using go-getter without requiring Go toolchain
// Supports HashiCorp go-getter URL formats:
//   - github.com/user/repo.git?ref=main&depth=1  (git clone with branch and depth)
//   - https://github.com/user/repo.git//subfolder  (subfolder extraction)
//   - https://example.com/file.tar.gz  (archive download with auto-extraction)
//
// If dest is empty, defaults to $HOME
func GetViaGoGetter(src, dest string) error {
	if dest == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dest = homeDir
	}

	// Expand environment variables and tilde in destination
	dest = ExpandPath(dest)

	// Add git:: prefix for .git URLs that don't have a protocol prefix
	// Go-getter requires this to recognize git repositories
	if strings.Contains(src, ".git") && !strings.Contains(src, "::") && !strings.HasPrefix(src, "http") {
		src = "git::" + src
	}

	// Remove destination folder if it exists (for git clone to work)
	if _, err := os.Stat(dest); err == nil {
		logger.Get().Info("Removing existing destination folder", zap.String("dest", dest))
		if err := os.RemoveAll(dest); err != nil {
			return fmt.Errorf("failed to remove existing destination %s: %w", dest, err)
		}
	}

	gc := NewGetterClient(WithTimeout(10 * time.Minute))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	logger.Get().Info("Downloading via go-getter",
		zap.String("src", src),
		zap.String("dest", dest))

	if _, err := gc.Get(ctx, dest, src); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	logger.Get().Info("Download completed successfully",
		zap.String("dest", dest))

	return nil
}

// InstallBinaryViaGoGetter installs a Go binary by cloning repo and building from source
// goPackage format: github.com/user/repo/v2/cmd/binary@version
// NOTE: This requires Go toolchain to be installed
func InstallBinaryViaGoGetter(binaryName string, goPackage string, binariesFolder string) error {
	if !IsGoInstalled() {
		return fmt.Errorf("go toolchain is not installed")
	}

	// Parse package path
	repoURL, cmdPath, version := parseGoPackagePath(goPackage)
	if repoURL == "" {
		return fmt.Errorf("invalid go package path: %s", goPackage)
	}

	logger.Get().Info("Installing via go-getter",
		zap.String("binary", binaryName),
		zap.String("repo", repoURL),
		zap.String("cmd_path", cmdPath),
		zap.String("version", version))

	// Create temp directory for clone
	tempDir, err := os.MkdirTemp("", "osmedeus-go-getter-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Use go-getter to clone the repository
	gc := NewGetterClient(WithTimeout(5 * time.Minute))

	// Format source URL for git getter
	// go-getter format: git::https://github.com/user/repo?ref=tag
	src := fmt.Sprintf("git::https://%s", repoURL)
	if version != "" && version != "latest" {
		src = fmt.Sprintf("%s?ref=%s", src, version)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	logger.Get().Info("Cloning repository", zap.String("src", src), zap.String("dest", tempDir))
	if _, err := gc.GetDir(ctx, tempDir, src); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Determine build directory
	buildDir := tempDir
	if cmdPath != "" {
		buildDir = filepath.Join(tempDir, cmdPath)
	}

	// Check if build directory exists
	if _, err := os.Stat(buildDir); err != nil {
		return fmt.Errorf("build directory not found: %s", buildDir)
	}

	// Build the binary
	outputPath := filepath.Join(tempDir, binaryName)
	cmd := exec.Command("go", "build", "-o", outputPath, ".")
	cmd.Dir = buildDir
	cmd.Env = os.Environ()

	logger.Get().Info("Building binary",
		zap.String("dir", buildDir),
		zap.String("output", outputPath))

	output, err := cmd.CombinedOutput()
	GoGetterInstallOutput = string(output)

	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("failed to build %s: %w\nOutput: %s", binaryName, err, string(output))
		}
		return fmt.Errorf("failed to build %s: %w", binaryName, err)
	}

	// Ensure binaries folder exists
	if err := os.MkdirAll(binariesFolder, 0755); err != nil {
		return fmt.Errorf("failed to create binaries folder: %w", err)
	}

	// Copy binary to binaries folder
	destPath := filepath.Join(binariesFolder, binaryName)
	if err := copyBinaryFile(outputPath, destPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	logger.Get().Info("Binary installed successfully",
		zap.String("binary", binaryName),
		zap.String("path", destPath))

	return nil
}

// parseGoPackagePath parses a Go package path into repo URL, command subpath, and version
// Supported formats:
//   - github.com/user/repo@latest
//   - github.com/user/repo/v2@v2.0.0
//   - github.com/user/repo/v2/cmd/binary@latest
//   - github.com/user/repo/cmd/binary@latest
func parseGoPackagePath(pkg string) (repoURL, cmdPath, version string) {
	// Split version
	parts := strings.SplitN(pkg, "@", 2)
	path := parts[0]
	if len(parts) > 1 {
		version = parts[1]
	} else {
		version = "latest"
	}

	// Split path components
	segments := strings.Split(path, "/")
	if len(segments) < 3 {
		return "", "", ""
	}

	// First 3 segments are always host/user/repo
	repoURL = strings.Join(segments[:3], "/")

	// Check for versioned module (v2, v3, etc.)
	if len(segments) > 3 {
		// Check if segment after repo is a version marker (v2, v3, etc.)
		if strings.HasPrefix(segments[3], "v") && len(segments[3]) <= 3 {
			// Versioned module path like github.com/user/repo/v2
			// The version marker is part of Go module path, not physical directory
			// So we keep repo URL as-is and check for cmd path after version marker
			if len(segments) > 4 {
				cmdPath = strings.Join(segments[4:], "/")
			}
		} else {
			// Non-versioned with subpath like github.com/user/repo/cmd/binary
			cmdPath = strings.Join(segments[3:], "/")
		}
	}

	return repoURL, cmdPath, version
}

// copyBinaryFile copies a binary file and sets executable permissions
func copyBinaryFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer func() { _ = destFile.Close() }()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	if err := os.Chmod(dest, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// ExpandPath expands environment variables and tilde in a path
// Handles both $HOME and ~ for home directory
func ExpandPath(path string) string {
	// Handle tilde expansion
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	} else if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			path = homeDir
		}
	}
	// Also expand environment variables like $HOME
	return os.ExpandEnv(path)
}
