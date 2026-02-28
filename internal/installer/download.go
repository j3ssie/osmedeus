package installer

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// SourceType represents the type of installation source
type SourceType int

const (
	SourceTypeUnknown SourceType = iota
	SourceTypeGit
	SourceTypeZipURL
	SourceTypeTarGzURL
	SourceTypeLocalZip
	SourceTypeLocalTarGz
	SourceTypeLocalFolder
)

// DetectSourceType determines the type of source from the input string
func DetectSourceType(source string) SourceType {
	source = strings.TrimSpace(source)

	// Check for local files/folders first
	if info, err := os.Stat(source); err == nil {
		if info.IsDir() {
			return SourceTypeLocalFolder
		}
		if strings.HasSuffix(strings.ToLower(source), ".zip") {
			return SourceTypeLocalZip
		}
		if strings.HasSuffix(strings.ToLower(source), ".tar.gz") || strings.HasSuffix(strings.ToLower(source), ".tgz") {
			return SourceTypeLocalTarGz
		}
	}

	// Check for URLs
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		lowerSource := strings.ToLower(source)
		if strings.HasSuffix(lowerSource, ".zip") {
			return SourceTypeZipURL
		}
		if strings.HasSuffix(lowerSource, ".tar.gz") || strings.HasSuffix(lowerSource, ".tgz") {
			return SourceTypeTarGzURL
		}
		// Assume git for other URLs (github.com, gitlab.com, etc.)
		return SourceTypeGit
	}

	// Check for git URLs without http prefix
	if strings.HasSuffix(source, ".git") || strings.Contains(source, "github.com") || strings.Contains(source, "gitlab.com") {
		return SourceTypeGit
	}

	return SourceTypeUnknown
}

// IsURL checks if the source is a URL
func IsURL(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

// IsGitURL checks if the source is a git repository URL
func IsGitURL(source string) bool {
	return DetectSourceType(source) == SourceTypeGit
}

// IsZipFile checks if the source is a zip file (local or URL)
func IsZipFile(source string) bool {
	st := DetectSourceType(source)
	return st == SourceTypeLocalZip || st == SourceTypeZipURL
}

// isGitHubURL checks if a URL is a GitHub URL that can benefit from authentication
func isGitHubURL(url string) bool {
	return strings.Contains(url, "github.com") || strings.Contains(url, "raw.githubusercontent.com")
}

// downloadMaxRetries is the number of retry attempts for transient download failures
const downloadMaxRetries = 3

// DownloadFile downloads a file from a URL to the destination path with retry and fallback.
// It retries transient failures with exponential backoff, then falls back to wget/curl.
// Optional customHeaders map adds custom HTTP headers to the request.
func DownloadFile(rawURL, dest string, customHeaders map[string]string) error {
	// Ensure destination directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	var lastErr error

	for attempt := 0; attempt <= downloadMaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * time.Second // 2s, 4s, 8s
			logger.Get().Info("Retrying download",
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
				zap.String("url", rawURL))
			time.Sleep(backoff)
		}

		lastErr = downloadFileOnce(rawURL, dest, customHeaders)
		if lastErr == nil {
			return nil
		}

		// Don't retry on non-retryable errors (404, 403, 401)
		if isNonRetryableError(lastErr) {
			return lastErr
		}

		// Clean up partial file before retry
		_ = os.Remove(dest)
	}

	// All Go HTTP retries exhausted — fall back to wget/curl
	logger.Get().Warn("Go HTTP download failed after retries, falling back to wget/curl",
		zap.String("url", rawURL),
		zap.Error(lastErr))
	_ = os.Remove(dest)

	if err := downloadFileWithExternalTool(rawURL, dest, customHeaders); err == nil {
		return nil
	}

	return fmt.Errorf("download failed after %d retries and wget/curl fallback: %w", downloadMaxRetries, lastErr)
}

// downloadFileOnce performs a single HTTP download attempt with content-length validation.
func downloadFileOnce(rawURL, dest string, customHeaders map[string]string) error {
	client := &http.Client{
		Timeout: 10 * time.Minute,
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for %s: %w", rawURL, err)
	}
	req.Header.Set("User-Agent", core.DefaultUA)

	// Auto-inject GitHub token for GitHub URLs
	if isGitHubURL(rawURL) {
		if token := getGitHubToken(); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	// Add custom headers
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", rawURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: status %d", rawURL, resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dest, err)
	}
	defer func() { _ = out.Close() }()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", dest, err)
	}

	// Validate content length if the server provided one
	if resp.ContentLength > 0 && written != resp.ContentLength {
		return fmt.Errorf("incomplete download of %s: got %d bytes, expected %d", rawURL, written, resp.ContentLength)
	}

	return nil
}

// validateDownloadURL ensures the URL is safe for use as an argument to external tools.
func validateDownloadURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q: only http and https are allowed", parsed.Scheme)
	}
	// Reject shell metacharacters to prevent injection when passed as exec args
	for _, ch := range []string{";", "|", "&", "`", "$", "(", ")", "\n", "\r"} {
		if strings.Contains(rawURL, ch) {
			return fmt.Errorf("URL contains disallowed character %q", ch)
		}
	}
	return nil
}

// downloadFileWithExternalTool attempts to download using wget, then curl as fallback.
func downloadFileWithExternalTool(rawURL, dest string, customHeaders map[string]string) error {
	if err := validateDownloadURL(rawURL); err != nil {
		return fmt.Errorf("URL validation failed: %w", err)
	}

	// Build auth header args for GitHub URLs
	var authHeaderWget []string
	var authHeaderCurl []string
	if isGitHubURL(rawURL) {
		if token := getGitHubToken(); token != "" {
			authHeaderWget = []string{"--header", "Authorization: Bearer " + token}
			authHeaderCurl = []string{"-H", "Authorization: Bearer " + token}
		}
	}

	// Add custom headers
	for key, value := range customHeaders {
		authHeaderWget = append(authHeaderWget, "--header", key+": "+value)
		authHeaderCurl = append(authHeaderCurl, "-H", key+": "+value)
	}

	// Try wget first
	if wgetPath, err := exec.LookPath("wget"); err == nil {
		args := append([]string{"-q", "-O", dest}, authHeaderWget...)
		args = append(args, rawURL)
		cmd := exec.Command(wgetPath, args...)
		if output, err := cmd.CombinedOutput(); err == nil {
			logger.Get().Info("Download succeeded via wget", zap.String("url", rawURL))
			return nil
		} else {
			logger.Get().Debug("wget fallback failed",
				zap.String("url", rawURL),
				zap.String("output", string(output)),
				zap.Error(err))
		}
	}

	// Fallback to curl
	if curlPath, err := exec.LookPath("curl"); err == nil {
		args := append([]string{"-fsSL", "-o", dest}, authHeaderCurl...)
		args = append(args, rawURL)
		cmd := exec.Command(curlPath, args...)
		if output, err := cmd.CombinedOutput(); err == nil {
			logger.Get().Info("Download succeeded via curl", zap.String("url", rawURL))
			return nil
		} else {
			logger.Get().Debug("curl fallback failed",
				zap.String("url", rawURL),
				zap.String("output", string(output)),
				zap.Error(err))
		}
	}

	return fmt.Errorf("neither wget nor curl succeeded for %s", rawURL)
}

// getGitHubToken returns the GitHub token from settings or environment
// Priority: GITHUB_API_KEY (from settings) > GH_TOKEN (from OS env)
func getGitHubToken() string {
	// First: try GITHUB_API_KEY from settings (exported to env by root.go)
	if token := os.Getenv("GITHUB_API_KEY"); token != "" {
		return token
	}
	// Fallback: GH_TOKEN from OS environment (used by GitHub CLI)
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}
	return ""
}

// injectGitHubToken injects token into GitHub URLs for authentication
// Converts: https://github.com/owner/repo.git -> https://{token}@github.com/owner/repo.git
func injectGitHubToken(repoURL string) string {
	token := getGitHubToken()
	if token == "" {
		return repoURL
	}
	// Only inject for github.com HTTPS URLs
	if strings.Contains(repoURL, "github.com") && strings.HasPrefix(repoURL, "https://") {
		repoURL = strings.Replace(repoURL, "https://github.com", "https://"+token+"@github.com", 1)
	}
	return repoURL
}

// GitClone clones a git repository to the destination directory
func GitClone(repoURL, dest string) error {
	// Inject GitHub token for private repo access
	repoURL = injectGitHubToken(repoURL)
	// Ensure destination parent directory exists
	parentDir := filepath.Dir(dest)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", parentDir, err)
	}

	// Remove destination if it exists
	if _, err := os.Stat(dest); err == nil {
		if err := os.RemoveAll(dest); err != nil {
			return fmt.Errorf("failed to remove existing directory %s: %w", dest, err)
		}
	}

	// Clone with depth 1 for faster download
	cmd := exec.Command("git", "clone", "--depth", "1", repoURL, dest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// FetchToTemp downloads or clones a source to a temporary directory
// Returns the path to the temp directory and any error
// Optional customHeaders map adds custom HTTP headers for URL downloads
func FetchToTemp(source string, customHeaders map[string]string) (string, error) {
	sourceType := DetectSourceType(source)

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "osmedeus-install-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	switch sourceType {
	case SourceTypeGit:
		destDir := filepath.Join(tempDir, "repo")
		if err := GitClone(source, destDir); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", err
		}
		return destDir, nil

	case SourceTypeZipURL:
		zipPath := filepath.Join(tempDir, "download.zip")
		if err := DownloadFile(source, zipPath, customHeaders); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", err
		}
		extractDir := filepath.Join(tempDir, "extracted")
		if err := ExtractZip(zipPath, extractDir); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", err
		}
		// Check if there's a single top-level directory and return that
		return findContentRoot(extractDir), nil

	case SourceTypeTarGzURL:
		tarPath := filepath.Join(tempDir, "download.tar.gz")
		if err := DownloadFile(source, tarPath, customHeaders); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", err
		}
		extractDir := filepath.Join(tempDir, "extracted")
		if err := ExtractTarGz(tarPath, extractDir); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", err
		}
		return findContentRoot(extractDir), nil

	case SourceTypeLocalZip:
		extractDir := filepath.Join(tempDir, "extracted")
		if err := ExtractZip(source, extractDir); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", err
		}
		return findContentRoot(extractDir), nil

	case SourceTypeLocalTarGz:
		extractDir := filepath.Join(tempDir, "extracted")
		if err := ExtractTarGz(source, extractDir); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", err
		}
		return findContentRoot(extractDir), nil

	case SourceTypeLocalFolder:
		// Copy local folder to temp directory
		destDir := filepath.Join(tempDir, "folder")
		if err := copyDir(source, destDir); err != nil {
			_ = os.RemoveAll(tempDir)
			return "", fmt.Errorf("failed to copy local folder: %w", err)
		}
		return destDir, nil

	default:
		_ = os.RemoveAll(tempDir)
		return "", fmt.Errorf("unknown source type: %s", source)
	}
}

// findContentRoot checks if extracted directory has a single subdirectory
// and returns that instead (common pattern in archives)
func findContentRoot(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return dir
	}

	// Filter out hidden files
	var visibleEntries []os.DirEntry
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), ".") {
			visibleEntries = append(visibleEntries, entry)
		}
	}

	// If there's exactly one directory, use that as the root
	if len(visibleEntries) == 1 && visibleEntries[0].IsDir() {
		return filepath.Join(dir, visibleEntries[0].Name())
	}

	return dir
}
