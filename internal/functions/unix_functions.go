package functions

import (
	"archive/zip"
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// sortUnix sorts a file using LC_ALL=C sort -u
// Usage: sortUnix(inputFile) -> bool (sorts in-place)
// Usage: sortUnix(inputFile, outputFile) -> bool
func (vf *vmFunc) sortUnix(call goja.FunctionCall) goja.Value {
	inputFile := call.Argument(0).String()
	logger.Get().Debug("Calling sortUnix", zap.String("inputFile", inputFile))

	if inputFile == "undefined" || inputFile == "" {
		logger.Get().Warn("sortUnix: empty input file provided")
		return vf.vm.ToValue(false)
	}

	// Default: sort in-place (output = input)
	outputFile := inputFile
	if !goja.IsUndefined(call.Argument(1)) {
		out := call.Argument(1).String()
		if out != "" && out != "undefined" {
			outputFile = out
		}
	}

	// Ensure output directory exists
	if outputFile != inputFile {
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Get().Warn("sortUnix: failed to create output directory", zap.String("dir", dir), zap.Error(err))
			return vf.vm.ToValue(false)
		}
	}

	// LC_ALL=C sort -u -o output input
	// Filter out existing locale variables and set LC_ALL=C to ensure consistent sorting
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "LC_ALL=") && !strings.HasPrefix(e, "LC_COLLATE=") && !strings.HasPrefix(e, "LANG=") {
			env = append(env, e)
		}
	}
	env = append(env, "LC_ALL=C")

	cmd := exec.Command("sort", "-u", "-o", outputFile, inputFile)
	cmd.Env = env

	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("sortUnix: sort command failed", zap.String("inputFile", inputFile), zap.Error(err))
	} else {
		logger.Get().Debug("sortUnix result", zap.String("inputFile", inputFile), zap.String("outputFile", outputFile), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// wgetUnix downloads a file using wget
// Usage: wgetUnix(url) -> bool (saves to current directory)
// Usage: wgetUnix(url, outputPath) -> bool
func (vf *vmFunc) wgetUnix(call goja.FunctionCall) goja.Value {
	url := call.Argument(0).String()
	outputPath := call.Argument(1).String()
	logger.Get().Debug("Calling wgetUnix", zap.String("url", url), zap.String("outputPath", outputPath))

	if url == "undefined" || url == "" {
		logger.Get().Warn("wgetUnix: empty URL provided")
		return vf.vm.ToValue(false)
	}

	args := []string{"-q", "--no-check-certificate"}

	if outputPath != "undefined" && outputPath != "" {
		// Ensure directory exists
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Get().Warn("wgetUnix: failed to create output directory", zap.String("dir", dir), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		args = append(args, "-O", outputPath)
	}
	args = append(args, url)

	cmd := exec.Command("wget", args...)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("wgetUnix: wget command failed", zap.String("url", url), zap.Error(err))
	} else {
		logger.Get().Debug("wgetUnix result", zap.String("url", url), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// wget downloads a file using pure Go with segmented parallel download support.
// For large files (>1MB) on servers that support Range requests, it splits the
// download into 4 parallel segments for faster throughput.
// Usage: wget(url, outputPath) -> bool
func (vf *vmFunc) wget(call goja.FunctionCall) goja.Value {
	rawURL := call.Argument(0).String()
	outputPath := call.Argument(1).String()
	logger.Get().Debug("Calling wget", zap.String("url", rawURL), zap.String("outputPath", outputPath))

	if rawURL == "undefined" || rawURL == "" {
		logger.Get().Warn("wget: empty URL provided")
		return vf.vm.ToValue(false)
	}
	if outputPath == "undefined" || outputPath == "" {
		logger.Get().Warn("wget: empty output path provided")
		return vf.vm.ToValue(false)
	}

	// Remove existing file if present
	_ = os.Remove(outputPath)

	// Ensure parent directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Get().Warn("wget: failed to create output directory", zap.String("dir", dir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// HEAD request to check Content-Length and Accept-Ranges
	headReq, err := http.NewRequest("HEAD", rawURL, nil)
	if err != nil {
		logger.Get().Warn("wget: failed to create HEAD request", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	headReq.Header.Set("User-Agent", core.DefaultUA)

	headResp, err := client.Do(headReq)
	if err != nil {
		logger.Get().Debug("wget: HEAD request failed, falling back to simple download", zap.Error(err))
		ok := wgetSimpleDownload(client, rawURL, outputPath)
		return vf.vm.ToValue(ok)
	}
	_ = headResp.Body.Close()

	contentLength, _ := strconv.ParseInt(headResp.Header.Get("Content-Length"), 10, 64)
	acceptRanges := headResp.Header.Get("Accept-Ranges")

	const minSegmentSize int64 = 1 << 20 // 1MB
	const numSegments = 4

	if acceptRanges == "bytes" && contentLength >= minSegmentSize {
		logger.Get().Debug("wget: using segmented download",
			zap.Int64("contentLength", contentLength),
			zap.Int("segments", numSegments))
		ok := wgetSegmentedDownload(client, rawURL, outputPath, contentLength, numSegments)
		if ok {
			return vf.vm.ToValue(true)
		}
		// If segmented download fails, fall back to simple
		logger.Get().Debug("wget: segmented download failed, falling back to simple download")
	}

	ok := wgetSimpleDownload(client, rawURL, outputPath)
	return vf.vm.ToValue(ok)
}

// wgetSimpleDownload downloads a file with a single GET request.
func wgetSimpleDownload(client *http.Client, rawURL, outputPath string) bool {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		logger.Get().Warn("wget: failed to create GET request", zap.Error(err))
		return false
	}
	req.Header.Set("User-Agent", core.DefaultUA)

	resp, err := client.Do(req)
	if err != nil {
		logger.Get().Warn("wget: GET request failed", zap.String("url", rawURL), zap.Error(err))
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Get().Warn("wget: bad status code", zap.Int("status", resp.StatusCode))
		return false
	}

	f, err := os.Create(outputPath)
	if err != nil {
		logger.Get().Warn("wget: failed to create output file", zap.Error(err))
		return false
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, resp.Body); err != nil {
		logger.Get().Warn("wget: failed to write file", zap.Error(err))
		_ = os.Remove(outputPath)
		return false
	}

	logger.Get().Debug("wget: simple download succeeded", zap.String("url", rawURL), zap.String("output", outputPath))
	return true
}

// wgetSegmentedDownload downloads a file in parallel segments using Range requests.
func wgetSegmentedDownload(client *http.Client, rawURL, outputPath string, contentLength int64, numSegments int) bool {
	segmentSize := contentLength / int64(numSegments)
	partFiles := make([]string, numSegments)

	g := new(errgroup.Group)

	for i := 0; i < numSegments; i++ {
		partFile := fmt.Sprintf("%s.part%d", outputPath, i)
		partFiles[i] = partFile

		start := int64(i) * segmentSize
		end := start + segmentSize - 1
		if i == numSegments-1 {
			end = contentLength - 1 // last segment gets remainder
		}

		g.Go(func() error {
			return wgetDownloadSegment(client, rawURL, partFile, start, end)
		})
	}

	if err := g.Wait(); err != nil {
		logger.Get().Warn("wget: segment download failed", zap.Error(err))
		// Cleanup part files
		for _, pf := range partFiles {
			_ = os.Remove(pf)
		}
		return false
	}

	// Assemble segments into final file
	outFile, err := os.Create(outputPath)
	if err != nil {
		logger.Get().Warn("wget: failed to create output file for assembly", zap.Error(err))
		for _, pf := range partFiles {
			_ = os.Remove(pf)
		}
		return false
	}

	for _, pf := range partFiles {
		partReader, err := os.Open(pf)
		if err != nil {
			_ = outFile.Close()
			_ = os.Remove(outputPath)
			for _, pf2 := range partFiles {
				_ = os.Remove(pf2)
			}
			return false
		}
		_, err = io.Copy(outFile, partReader)
		_ = partReader.Close()
		if err != nil {
			_ = outFile.Close()
			_ = os.Remove(outputPath)
			for _, pf2 := range partFiles {
				_ = os.Remove(pf2)
			}
			return false
		}
	}

	_ = outFile.Close()

	// Cleanup part files
	for _, pf := range partFiles {
		_ = os.Remove(pf)
	}

	logger.Get().Debug("wget: segmented download succeeded",
		zap.String("url", rawURL),
		zap.String("output", outputPath),
		zap.Int64("size", contentLength))
	return true
}

// wgetDownloadSegment downloads a single byte range segment to a part file.
func wgetDownloadSegment(client *http.Client, rawURL, partFile string, start, end int64) error {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", core.DefaultUA)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("GET segment: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for range %d-%d", resp.StatusCode, start, end)
	}

	f, err := os.Create(partFile)
	if err != nil {
		return fmt.Errorf("create part file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("write part file: %w", err)
	}

	return nil
}

// injectGitHubTokenForClone injects GitHub token into GitHub URLs for authentication
// Priority: GITHUB_API_KEY (from settings) > GH_TOKEN (from OS env)
func injectGitHubTokenForClone(repoURL string) string {
	// First: try GITHUB_API_KEY from settings (exported to env by root.go)
	token := os.Getenv("GITHUB_API_KEY")
	if token == "" {
		// Fallback: GH_TOKEN from OS environment (used by GitHub CLI)
		token = os.Getenv("GH_TOKEN")
	}
	if token == "" {
		return repoURL
	}
	// Only inject for github.com HTTPS URLs
	if strings.Contains(repoURL, "github.com") && strings.HasPrefix(repoURL, "https://") {
		repoURL = strings.Replace(repoURL, "https://github.com", "https://"+token+"@github.com", 1)
	}
	return repoURL
}

// gitClone clones a git repository
// Usage: gitClone(repo) -> bool (clones to current directory)
// Usage: gitClone(repo, dest) -> bool
func (vf *vmFunc) gitClone(call goja.FunctionCall) goja.Value {
	repo := call.Argument(0).String()
	// Inject GitHub token for private repo access
	repo = injectGitHubTokenForClone(repo)
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling gitClone", zap.String("repo", repo), zap.String("dest", dest))

	if repo == "undefined" || repo == "" {
		logger.Get().Warn("gitClone: empty repo provided")
		return vf.vm.ToValue(false)
	}

	args := []string{"clone", "--depth", "1"}
	args = append(args, repo)

	if dest != "undefined" && dest != "" {
		// Ensure parent directory exists
		dir := filepath.Dir(dest)
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Get().Warn("gitClone: failed to create parent directory", zap.String("dir", dir), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		args = append(args, dest)
	}

	cmd := exec.Command("git", args...)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("gitClone: git clone failed", zap.String("repo", repo), zap.Error(err))
	} else {
		logger.Get().Debug("gitClone result", zap.String("repo", repo), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// gitCloneSubfolder clones a git repository and extracts a specific subfolder
// Usage: git_clone_subfolder(git_url, subfolder, dest) -> bool
// Falls back to ZIP download if git command is not available (GitHub only)
func (vf *vmFunc) gitCloneSubfolder(call goja.FunctionCall) goja.Value {
	gitURL := call.Argument(0).String()
	subfolder := call.Argument(1).String()
	dest := call.Argument(2).String()

	logger.Get().Debug("Calling gitCloneSubfolder",
		zap.String("gitURL", gitURL),
		zap.String("subfolder", subfolder),
		zap.String("dest", dest))

	// Validate required parameters
	if gitURL == "undefined" || gitURL == "" {
		logger.Get().Warn("gitCloneSubfolder: empty git URL provided")
		return vf.vm.ToValue(false)
	}
	if subfolder == "undefined" || subfolder == "" {
		logger.Get().Warn("gitCloneSubfolder: empty subfolder provided")
		return vf.vm.ToValue(false)
	}
	if dest == "undefined" || dest == "" {
		logger.Get().Warn("gitCloneSubfolder: empty destination provided")
		return vf.vm.ToValue(false)
	}

	// Create temp directory for cloning
	tempDir, err := os.MkdirTemp("", "git-clone-subfolder-*")
	if err != nil {
		logger.Get().Warn("gitCloneSubfolder: failed to create temp directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	cloneSuccess := false

	// Try git clone first
	_, gitErr := exec.LookPath("git")
	if gitErr == nil {
		// Inject GitHub token for private repo access
		repoURL := injectGitHubTokenForClone(gitURL)
		cmd := exec.Command("git", "clone", "--depth", "1", repoURL, tempDir)
		if err := cmd.Run(); err == nil {
			cloneSuccess = true
			logger.Get().Debug("gitCloneSubfolder: git clone succeeded", zap.String("gitURL", gitURL))
		} else {
			logger.Get().Debug("gitCloneSubfolder: git clone failed, trying ZIP fallback", zap.Error(err))
		}
	}

	// Fallback to ZIP download for GitHub URLs
	if !cloneSuccess && isGitHubURL(gitURL) {
		zipURL := convertToGitHubZipURL(gitURL)
		if zipURL != "" {
			if err := downloadAndExtractZip(zipURL, tempDir); err == nil {
				cloneSuccess = true
				logger.Get().Debug("gitCloneSubfolder: ZIP download succeeded", zap.String("zipURL", zipURL))
			} else {
				logger.Get().Warn("gitCloneSubfolder: ZIP download failed", zap.String("zipURL", zipURL), zap.Error(err))
			}
		}
	}

	if !cloneSuccess {
		logger.Get().Warn("gitCloneSubfolder: both git clone and ZIP fallback failed", zap.String("gitURL", gitURL))
		return vf.vm.ToValue(false)
	}

	// Find the subfolder in the cloned/extracted content
	// For ZIP extracts, there's usually a root folder like "repo-main/"
	subfolderPath := findSubfolder(tempDir, subfolder)
	if subfolderPath == "" {
		logger.Get().Warn("gitCloneSubfolder: subfolder not found",
			zap.String("subfolder", subfolder),
			zap.String("tempDir", tempDir))
		return vf.vm.ToValue(false)
	}

	// Remove destination if it exists
	if err := os.RemoveAll(dest); err != nil {
		logger.Get().Warn("gitCloneSubfolder: failed to remove existing destination", zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		logger.Get().Warn("gitCloneSubfolder: failed to create parent directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Copy subfolder to destination
	if err := copyDir(subfolderPath, dest); err != nil {
		logger.Get().Warn("gitCloneSubfolder: failed to copy subfolder",
			zap.String("src", subfolderPath),
			zap.String("dest", dest),
			zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("gitCloneSubfolder completed successfully",
		zap.String("gitURL", gitURL),
		zap.String("subfolder", subfolder),
		zap.String("dest", dest))

	return vf.vm.ToValue(true)
}

// isGitHubURL checks if the URL is a GitHub repository URL
func isGitHubURL(url string) bool {
	return strings.Contains(url, "github.com")
}

// convertToGitHubZipURL converts a GitHub repo URL to a ZIP download URL
// Supports: https://github.com/user/repo, git@github.com:user/repo.git
func convertToGitHubZipURL(gitURL string) string {
	// Handle SSH URLs: git@github.com:user/repo.git
	if strings.HasPrefix(gitURL, "git@github.com:") {
		path := strings.TrimPrefix(gitURL, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		return fmt.Sprintf("https://github.com/%s/archive/refs/heads/main.zip", path)
	}

	// Handle HTTPS URLs: https://github.com/user/repo
	if strings.HasPrefix(gitURL, "https://github.com/") {
		path := strings.TrimPrefix(gitURL, "https://github.com/")
		path = strings.TrimSuffix(path, ".git")
		return fmt.Sprintf("https://github.com/%s/archive/refs/heads/main.zip", path)
	}

	// Handle http URLs
	if strings.HasPrefix(gitURL, "http://github.com/") {
		path := strings.TrimPrefix(gitURL, "http://github.com/")
		path = strings.TrimSuffix(path, ".git")
		return fmt.Sprintf("https://github.com/%s/archive/refs/heads/main.zip", path)
	}

	return ""
}

// downloadAndExtractZip downloads a ZIP file and extracts it to the destination
func downloadAndExtractZip(zipURL, dest string) error {
	// Download ZIP to temp file
	resp, err := http.Get(zipURL)
	if err != nil {
		return fmt.Errorf("failed to download ZIP: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Try master branch if main fails
		masterURL := strings.Replace(zipURL, "/main.zip", "/master.zip", 1)
		resp, err = http.Get(masterURL)
		if err != nil {
			return fmt.Errorf("failed to download ZIP (master fallback): %w", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("ZIP download failed with status: %d", resp.StatusCode)
		}
	}

	// Create temp file for ZIP
	tmpFile, err := os.CreateTemp("", "git-zip-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	// Write response to temp file
	_, err = io.Copy(tmpFile, resp.Body)
	_ = tmpFile.Close()
	if err != nil {
		return fmt.Errorf("failed to write ZIP file: %w", err)
	}

	// Extract ZIP
	return extractZip(tmpPath, dest)
}

// extractZip extracts a ZIP file to the destination directory
func extractZip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open ZIP: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Security check: prevent path traversal
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, 0755); err != nil {
				return err
			}
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		// Extract file
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			_ = outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		_ = rc.Close()
		_ = outFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// findSubfolder finds the subfolder path, handling ZIP extraction root folders
func findSubfolder(baseDir, subfolder string) string {
	// Try direct path first
	directPath := filepath.Join(baseDir, subfolder)
	if info, err := os.Stat(directPath); err == nil && info.IsDir() {
		return directPath
	}

	// For ZIP extracts, there's usually a root folder like "repo-main/"
	// Check one level deep
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if entry.IsDir() {
			nestedPath := filepath.Join(baseDir, entry.Name(), subfolder)
			if info, err := os.Stat(nestedPath); err == nil && info.IsDir() {
				return nestedPath
			}
		}
	}

	return ""
}

// copyDir recursively copies a directory from src to dst
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// zipUnix creates a zip archive using the zip command
// Usage: zipUnix(source, dest) -> bool (zip -r dest source)
func (vf *vmFunc) zipUnix(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling zipUnix", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("zipUnix: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	// Ensure output directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Get().Warn("zipUnix: failed to create output directory", zap.String("dir", dir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// zip -r dest source
	cmd := exec.Command("zip", "-r", dest, source)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("zipUnix: zip command failed", zap.String("source", source), zap.Error(err))
	} else {
		logger.Get().Debug("zipUnix result", zap.String("source", source), zap.String("dest", dest), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// unzipUnix extracts a zip archive using the unzip command
// Usage: unzipUnix(source) -> bool (extracts to current directory)
// Usage: unzipUnix(source, dest) -> bool (unzip source -d dest)
func (vf *vmFunc) unzipUnix(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling unzipUnix", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" {
		logger.Get().Warn("unzipUnix: empty source provided")
		return vf.vm.ToValue(false)
	}

	args := []string{"-o", source} // -o for overwrite without prompting

	if dest != "undefined" && dest != "" {
		// Ensure destination directory exists
		if err := os.MkdirAll(dest, 0755); err != nil {
			logger.Get().Warn("unzipUnix: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		args = append(args, "-d", dest)
	}

	cmd := exec.Command("unzip", args...)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("unzipUnix: unzip command failed", zap.String("source", source), zap.Error(err))
	} else {
		logger.Get().Debug("unzipUnix result", zap.String("source", source), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// tarUnix creates a tar.gz archive using the tar command
// Usage: tarUnix(source, dest) -> bool (tar -czf dest source)
func (vf *vmFunc) tarUnix(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling tarUnix", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" || dest == "undefined" || dest == "" {
		logger.Get().Warn("tarUnix: empty source or dest provided")
		return vf.vm.ToValue(false)
	}

	// Ensure output directory exists
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Get().Warn("tarUnix: failed to create output directory", zap.String("dir", dir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// tar -czf dest source
	cmd := exec.Command("tar", "-czf", dest, source)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("tarUnix: tar command failed", zap.String("source", source), zap.Error(err))
	} else {
		logger.Get().Debug("tarUnix result", zap.String("source", source), zap.String("dest", dest), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// untarUnix extracts a tar.gz archive using the tar command
// Usage: untarUnix(source) -> bool (extracts to current directory)
// Usage: untarUnix(source, dest) -> bool (tar -xzf source -C dest)
func (vf *vmFunc) untarUnix(call goja.FunctionCall) goja.Value {
	source := call.Argument(0).String()
	dest := call.Argument(1).String()
	logger.Get().Debug("Calling untarUnix", zap.String("source", source), zap.String("dest", dest))

	if source == "undefined" || source == "" {
		logger.Get().Warn("untarUnix: empty source provided")
		return vf.vm.ToValue(false)
	}

	args := []string{"-xzf", source}

	if dest != "undefined" && dest != "" {
		// Ensure destination directory exists
		if err := os.MkdirAll(dest, 0755); err != nil {
			logger.Get().Warn("untarUnix: failed to create destination directory", zap.String("dest", dest), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		args = append(args, "-C", dest)
	}

	cmd := exec.Command("tar", args...)
	err := cmd.Run()
	if err != nil {
		logger.Get().Warn("untarUnix: tar extract failed", zap.String("source", source), zap.Error(err))
	} else {
		logger.Get().Debug("untarUnix result", zap.String("source", source), zap.Bool("success", true))
	}
	return vf.vm.ToValue(err == nil)
}

// diffUnix compares two files using the diff command
// Usage: diffUnix(file1, file2) -> structured JSON response
// Usage: diffUnix(file1, file2, output) -> structured JSON response (also saves to output file)
// Returns: {error, message, output, file1, file2, line_count}
func (vf *vmFunc) diffUnix(call goja.FunctionCall) goja.Value {
	file1 := call.Argument(0).String()
	file2 := call.Argument(1).String()
	logger.Get().Debug("Calling diffUnix", zap.String("file1", file1), zap.String("file2", file2))

	if file1 == "undefined" || file1 == "" || file2 == "undefined" || file2 == "" {
		logger.Get().Warn("diffUnix: empty file paths provided")
		return vf.vm.ToValue(map[string]interface{}{
			"error":      "empty file paths provided",
			"message":    "error",
			"output":     "",
			"file1":      file1,
			"file2":      file2,
			"line_count": 0,
		})
	}

	// diff file1 file2 (returns non-zero exit code if files differ, which is expected)
	cmd := exec.Command("diff", file1, file2)
	output, _ := cmd.Output() // Ignore error since diff returns 1 when files differ

	outputStr := string(output)

	// Count lines in diff output
	lineCount := 0
	if len(outputStr) > 0 {
		lineCount = strings.Count(outputStr, "\n")
		if !strings.HasSuffix(outputStr, "\n") {
			lineCount++
		}
	}

	// If output file specified, save the diff
	outputPath := call.Argument(2).String()
	if outputPath != "undefined" && outputPath != "" {
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err == nil {
			_ = os.WriteFile(outputPath, output, 0644)
		}
	}

	logger.Get().Debug("diffUnix result", zap.String("file1", file1), zap.String("file2", file2), zap.Int("line_count", lineCount))
	return vf.vm.ToValue(map[string]interface{}{
		"error":      nil,
		"message":    "success",
		"output":     outputStr,
		"file1":      file1,
		"file2":      file2,
		"line_count": lineCount,
	})
}

// parseSedSyntax parses sed s/pattern/replacement/flags syntax
// Returns pattern, replacement, global flag, and error
func parseSedSyntax(syntax string) (pattern, replacement string, global bool, err error) {
	syntax = strings.TrimSpace(syntax)

	// Must start with 's'
	if len(syntax) < 4 || syntax[0] != 's' {
		return "", "", false, fmt.Errorf("invalid sed syntax: must start with 's'")
	}

	// Get the delimiter (character after 's')
	delim := syntax[1]

	// Find the parts: s/pattern/replacement/flags
	// Start after "s<delim>"
	rest := syntax[2:]

	// Find pattern end (first unescaped delimiter)
	patternEnd := findUnescapedDelim(rest, delim)
	if patternEnd == -1 {
		return "", "", false, fmt.Errorf("invalid sed syntax: missing pattern delimiter")
	}

	pattern = rest[:patternEnd]
	rest = rest[patternEnd+1:]

	// Find replacement end (next unescaped delimiter)
	replacementEnd := findUnescapedDelim(rest, delim)
	if replacementEnd == -1 {
		// No closing delimiter, rest is the replacement with no flags
		replacement = rest
		return pattern, replacement, false, nil
	}

	replacement = rest[:replacementEnd]
	flags := rest[replacementEnd+1:]

	// Check for global flag
	global = strings.Contains(flags, "g")

	return pattern, replacement, global, nil
}

// findUnescapedDelim finds the first unescaped occurrence of delimiter
func findUnescapedDelim(s string, delim byte) int {
	escaped := false
	for i := 0; i < len(s); i++ {
		if escaped {
			escaped = false
			continue
		}
		if s[i] == '\\' {
			escaped = true
			continue
		}
		if s[i] == delim {
			return i
		}
	}
	return -1
}

// sedStringReplace performs literal string replacement using sed-like syntax
// Usage: sed_string_replace('s/old/new/g', '/path/to/source', '/path/to/dest') -> bool
func (vf *vmFunc) sedStringReplace(call goja.FunctionCall) goja.Value {
	sedSyntax := call.Argument(0).String()
	sourceFile := call.Argument(1).String()
	destFile := call.Argument(2).String()

	logger.Get().Debug("Calling sedStringReplace",
		zap.String("syntax", sedSyntax),
		zap.String("source", sourceFile),
		zap.String("dest", destFile))

	if sedSyntax == "undefined" || sedSyntax == "" {
		logger.Get().Warn("sedStringReplace: empty sed syntax provided")
		return vf.vm.ToValue(false)
	}

	if sourceFile == "undefined" || sourceFile == "" {
		logger.Get().Warn("sedStringReplace: empty source file provided")
		return vf.vm.ToValue(false)
	}

	if destFile == "undefined" || destFile == "" {
		logger.Get().Warn("sedStringReplace: empty dest file provided")
		return vf.vm.ToValue(false)
	}

	// Parse sed syntax
	pattern, replacement, global, err := parseSedSyntax(sedSyntax)
	if err != nil {
		logger.Get().Warn("sedStringReplace: failed to parse sed syntax", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Open source file
	source, err := os.Open(sourceFile)
	if err != nil {
		logger.Get().Warn("sedStringReplace: failed to open source file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = source.Close() }()

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		logger.Get().Warn("sedStringReplace: failed to create dest directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Create dest file
	dest, err := os.Create(destFile)
	if err != nil {
		logger.Get().Warn("sedStringReplace: failed to create dest file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = dest.Close() }()

	// Stream process line by line
	scanner := bufio.NewScanner(source)
	writer := bufio.NewWriter(dest)

	for scanner.Scan() {
		line := scanner.Text()
		var newLine string
		if global {
			newLine = strings.ReplaceAll(line, pattern, replacement)
		} else {
			newLine = strings.Replace(line, pattern, replacement, 1)
		}
		_, _ = writer.WriteString(newLine + "\n")
	}

	if err := scanner.Err(); err != nil {
		logger.Get().Warn("sedStringReplace: error reading source file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if err := writer.Flush(); err != nil {
		logger.Get().Warn("sedStringReplace: error writing dest file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("sedStringReplace completed successfully",
		zap.String("source", sourceFile),
		zap.String("dest", destFile))

	return vf.vm.ToValue(true)
}

// sedRegexReplace performs regex replacement using sed-like syntax
// Usage: sed_regex_replace('s/[0-9]+/NUM/g', '/path/to/source', '/path/to/dest') -> bool
func (vf *vmFunc) sedRegexReplace(call goja.FunctionCall) goja.Value {
	sedSyntax := call.Argument(0).String()
	sourceFile := call.Argument(1).String()
	destFile := call.Argument(2).String()

	logger.Get().Debug("Calling sedRegexReplace",
		zap.String("syntax", sedSyntax),
		zap.String("source", sourceFile),
		zap.String("dest", destFile))

	if sedSyntax == "undefined" || sedSyntax == "" {
		logger.Get().Warn("sedRegexReplace: empty sed syntax provided")
		return vf.vm.ToValue(false)
	}

	if sourceFile == "undefined" || sourceFile == "" {
		logger.Get().Warn("sedRegexReplace: empty source file provided")
		return vf.vm.ToValue(false)
	}

	if destFile == "undefined" || destFile == "" {
		logger.Get().Warn("sedRegexReplace: empty dest file provided")
		return vf.vm.ToValue(false)
	}

	// Parse sed syntax
	pattern, replacement, global, err := parseSedSyntax(sedSyntax)
	if err != nil {
		logger.Get().Warn("sedRegexReplace: failed to parse sed syntax", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Compile regex
	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Get().Warn("sedRegexReplace: failed to compile regex", zap.String("pattern", pattern), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Open source file
	source, err := os.Open(sourceFile)
	if err != nil {
		logger.Get().Warn("sedRegexReplace: failed to open source file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = source.Close() }()

	// Ensure dest directory exists
	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		logger.Get().Warn("sedRegexReplace: failed to create dest directory", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Create dest file
	dest, err := os.Create(destFile)
	if err != nil {
		logger.Get().Warn("sedRegexReplace: failed to create dest file", zap.Error(err))
		return vf.vm.ToValue(false)
	}
	defer func() { _ = dest.Close() }()

	// Stream process line by line
	scanner := bufio.NewScanner(source)
	writer := bufio.NewWriter(dest)

	for scanner.Scan() {
		line := scanner.Text()
		var newLine string
		if global {
			newLine = re.ReplaceAllString(line, replacement)
		} else {
			// Replace only first occurrence
			loc := re.FindStringIndex(line)
			if loc != nil {
				newLine = line[:loc[0]] + re.ReplaceAllString(line[loc[0]:loc[1]], replacement) + line[loc[1]:]
			} else {
				newLine = line
			}
		}
		_, _ = writer.WriteString(newLine + "\n")
	}

	if err := scanner.Err(); err != nil {
		logger.Get().Warn("sedRegexReplace: error reading source file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	if err := writer.Flush(); err != nil {
		logger.Get().Warn("sedRegexReplace: error writing dest file", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("sedRegexReplace completed successfully",
		zap.String("source", sourceFile),
		zap.String("dest", destFile))

	return vf.vm.ToValue(true)
}
