package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/j3ssie/osmedeus/v5/public"
	"go.uber.org/zap"
)

// DefaultRegistryURL is the default URL for the binary registry
const DefaultRegistryURL = "https://raw.githubusercontent.com/osmedeus/osmedeus-base/main/registry-metadata.json"

// BinaryEntry represents a single binary's download/install information
// Supports both download URLs and commands per OS/architecture
type BinaryEntry struct {
	Desc                string            `json:"desc,omitempty"`
	RepoLink            string            `json:"repo_link,omitempty"`
	Version             string            `json:"version,omitempty"`
	PackageManager      string            `json:"package-manager,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
	ValidateCommand     string            `json:"valide-command,omitempty"`
	NixPackage          string            `json:"nix_package,omitempty"`
	GoInstall           string            `json:"go_install,omitempty"` // Go package path for go install (e.g., "github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest")
	Linux               map[string]string `json:"linux,omitempty"`
	Darwin              map[string]string `json:"darwin,omitempty"`
	Windows             map[string]string `json:"windows,omitempty"`
	CommandLinux        map[string]string `json:"command-linux,omitempty"`
	CommandDarwin       map[string]string `json:"command-darwin,omitempty"`
	CommandDual         map[string]string `json:"command-dual,omitempty"`
	MultiCommandsLinux  []string          `json:"multi-commands-linux,omitempty"`
	MultiCommandsDarwin []string          `json:"multi-commands-darwin,omitempty"`
}

// BinaryRegistry is a map of binary name to BinaryEntry
type BinaryRegistry map[string]BinaryEntry

// IsGoGetterPackage returns true if entry uses go-getter package manager
func (entry *BinaryEntry) IsGoGetterPackage() bool {
	return entry.PackageManager == "go-getter"
}

// GetGoGetterPackagePath returns the Go package path for go-getter entries
func (entry *BinaryEntry) GetGoGetterPackagePath() string {
	if !entry.IsGoGetterPackage() {
		return ""
	}
	if entry.CommandDual != nil {
		if pkg, ok := entry.CommandDual["dual"]; ok {
			return pkg
		}
	}
	return ""
}

// LoadRegistry loads a binary registry from a file path or URL
// If no path is provided, uses embedded registry (falls back to GitHub URL if embedded fails)
// Optional customHeaders map adds custom HTTP headers for URL fetches
func LoadRegistry(pathOrURL string, customHeaders map[string]string) (BinaryRegistry, error) {
	var data []byte
	var err error

	if pathOrURL == "" {
		// Try embedded registry first
		data, err = public.GetRegistryMetadata()
		if err != nil {
			// Fall back to remote URL
			data, err = fetchURL(DefaultRegistryURL, customHeaders)
		}
	} else if IsURL(pathOrURL) {
		data, err = fetchURL(pathOrURL, customHeaders)
	} else {
		data, err = os.ReadFile(pathOrURL)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load registry: %w", err)
	}

	// Parse into raw map first to filter metadata keys (prefixed with _)
	var rawRegistry map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawRegistry); err != nil {
		return nil, fmt.Errorf("failed to parse registry JSON: %w", err)
	}

	// Filter and parse binary entries
	registry := make(BinaryRegistry)
	for name, rawEntry := range rawRegistry {
		// Skip metadata keys (start with underscore)
		if strings.HasPrefix(name, "_") {
			continue
		}

		var entry BinaryEntry
		if err := json.Unmarshal(rawEntry, &entry); err != nil {
			return nil, fmt.Errorf("failed to parse entry '%s': %w", name, err)
		}
		registry[name] = entry
	}

	return registry, nil
}

func fetchURL(url string, customHeaders map[string]string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request with User-Agent header
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", core.DefaultUA)

	// Auto-inject GitHub token for GitHub URLs (helps with rate limiting and private repos)
	if isGitHubURL(url) {
		if token := getGitHubToken(); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	// Add custom headers (can override auto-injected headers if needed)
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// DetectPackageManager returns the system's package manager
func DetectPackageManager() string {
	if runtime.GOOS == "darwin" {
		return "brew"
	}

	// Linux: check common package managers in order of preference
	managers := []string{"apt", "dnf", "yum", "pacman", "zypper", "apk"}
	for _, mgr := range managers {
		if _, err := exec.LookPath(mgr); err == nil {
			return mgr
		}
	}
	return "apt" // fallback
}

// SubstitutePackageManager replaces <auto_detect_package_manager> placeholder
func SubstitutePackageManager(command string) string {
	if !strings.Contains(command, "<auto_detect_package_manager>") {
		return command
	}
	return strings.ReplaceAll(command, "<auto_detect_package_manager>", DetectPackageManager())
}

// GetBinaryInfo returns download URL or command(s) for current OS/arch
// Returns: url, commands (slice), error
// Priority: multi-commands > command-dual > command-<os>.dual > <os>.dual > command-<os>.<arch> > <os>.<arch>
func (entry *BinaryEntry) GetBinaryInfo() (url string, commands []string, err error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// 1. Check multi-commands first
	var multiCmds []string
	switch osName {
	case "linux":
		multiCmds = entry.MultiCommandsLinux
	case "darwin":
		multiCmds = entry.MultiCommandsDarwin
	}
	if len(multiCmds) > 0 {
		// Substitute package manager in all commands
		result := make([]string, len(multiCmds))
		for i, cmd := range multiCmds {
			result[i] = SubstitutePackageManager(cmd)
		}
		return "", result, nil
	}

	// 2. Check command-dual (applies to any OS)
	if entry.CommandDual != nil {
		if cmd, ok := entry.CommandDual["dual"]; ok {
			return "", []string{SubstitutePackageManager(cmd)}, nil
		}
	}

	// 3. Check command-<os> with dual key or arch-specific
	var commandMap map[string]string
	switch osName {
	case "linux":
		commandMap = entry.CommandLinux
	case "darwin":
		commandMap = entry.CommandDarwin
	}
	if commandMap != nil {
		// Check dual key first (any architecture)
		if cmd, ok := commandMap["dual"]; ok {
			return "", []string{SubstitutePackageManager(cmd)}, nil
		}
		// Then check arch-specific
		if cmd, ok := commandMap[arch]; ok {
			return "", []string{SubstitutePackageManager(cmd)}, nil
		}
	}

	// 4. Check OS map for dual key (command) or arch-specific URL
	var urlMap map[string]string
	switch osName {
	case "linux":
		urlMap = entry.Linux
	case "darwin":
		urlMap = entry.Darwin
	case "windows":
		urlMap = entry.Windows
	}
	if urlMap != nil {
		// Check dual key first (it's a command, not URL)
		if cmd, ok := urlMap["dual"]; ok {
			return "", []string{SubstitutePackageManager(cmd)}, nil
		}
		// Then check arch-specific URL
		if downloadURL, ok := urlMap[arch]; ok {
			return downloadURL, nil, nil
		}
	}

	return "", nil, fmt.Errorf("no download/command available for %s/%s", osName, arch)
}

// coreUnixTools lists system utilities that should never be copied to external-binaries.
// These are typically provided by the OS and copying them is unnecessary.
var coreUnixTools = map[string]bool{
	// Core utilities
	"cat": true, "cp": true, "mv": true, "rm": true, "ls": true,
	"mkdir": true, "rmdir": true, "touch": true, "chmod": true, "chown": true,
	// Text processing
	"grep": true, "sed": true, "awk": true, "sort": true, "uniq": true,
	"head": true, "tail": true, "cut": true, "tr": true, "wc": true,
	// Network
	"curl": true, "wget": true, "ping": true, "ssh": true, "scp": true,
	// Build tools
	"make": true, "gcc": true, "cc": true, "clang": true,
	// Other common tools
	"find": true, "xargs": true, "tar": true, "gzip": true, "gunzip": true,
	"zip": true, "unzip": true, "diff": true, "patch": true,
	"which": true, "whoami": true, "hostname": true, "date": true,
	"env": true, "echo": true, "printf": true, "tee": true,
	"bash": true, "sh": true, "zsh": true,
	// Version control
	"git": true,
}

// IsCoreUnixTool returns true if the binary name is a core Unix tool
func IsCoreUnixTool(name string) bool {
	return coreUnixTools[name]
}

// IsBinaryInPath checks if a binary exists and is executable in $PATH
func IsBinaryInPath(name string) bool {
	path, err := exec.LookPath(name)
	if err != nil {
		return false
	}
	// Verify it's actually executable
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}

// IsBinaryInstalled checks if a binary is installed using validate command or PATH lookup
// If entry has a ValidateCommand, run it and check exit code (0 = installed)
// If ValidateCommand is empty, fall back to checking if binary name is in PATH
func IsBinaryInstalled(name string, entry *BinaryEntry) bool {
	// If validate command is provided and not empty, use it
	if entry != nil && entry.ValidateCommand != "" {
		// @NOTE: This is intentional - ValidateCommand comes from the binary registry
		// configuration which is a trusted source for installation validation commands.
		cmd := exec.Command("sh", "-c", entry.ValidateCommand)
		err := cmd.Run()
		return err == nil // exit code 0 means installed
	}
	// Fall back to default PATH check
	return IsBinaryInPath(name)
}

// InstallBinary installs a single binary from the registry
// Skips installation if the binary is already available in PATH
// Optional customHeaders map adds custom HTTP headers for downloads
func InstallBinary(name string, registry BinaryRegistry, binariesFolder string, customHeaders map[string]string) error {
	// Check if binary already exists in PATH
	if IsBinaryInPath(name) {
		if IsCoreUnixTool(name) {
			// Core Unix tools: show exists, don't copy
			fmt.Printf("[%s] Binary '%s' already available in PATH (system tool, skipping copy)\n",
				terminal.Gray(terminal.SymbolBowtie), terminal.HiBlue(name))
			return nil
		}
		// Non-core tools: show exists, still symlink to external-binaries
		fmt.Printf("[%s] Binary '%s' already available in PATH, symlinking to external-binaries\n",
			terminal.Gray(terminal.SymbolBowtie), terminal.HiBlue(name))
		entry, ok := registry[name]
		if ok {
			_ = SymlinkInstalledBinaryToFolder(name, &entry, binariesFolder)
		}
		return nil
	}

	entry, ok := registry[name]
	if !ok {
		return fmt.Errorf("binary '%s' not found in registry", name)
	}

	// Check for go-getter package manager
	if entry.IsGoGetterPackage() {
		pkg := entry.GetGoGetterPackagePath()
		if pkg == "" {
			return fmt.Errorf("go-getter entry '%s' missing package path in command-dual", name)
		}

		// Check if package path contains space-separated destination
		// Format: "github.com/repo.git?ref=main&depth=1 $HOME/destination"
		if strings.Contains(pkg, " ") {
			parts := strings.SplitN(pkg, " ", 2)
			src := strings.TrimSpace(parts[0])
			dest := strings.TrimSpace(parts[1])

			// Expand environment variables and tilde in destination
			dest = ExpandPath(dest)

			logger.Get().Info("Installing via go-getter",
				zap.String("name", name),
				zap.String("src", src),
				zap.String("dest", dest))
			return GetViaGoGetter(src, dest)
		}

		// No destination specified - use go-getter to build Go binary
		logger.Get().Info("Installing binary via go-getter",
			zap.String("name", name),
			zap.String("package", pkg))
		return InstallBinaryViaGoGetter(name, pkg, binariesFolder)
	}

	url, commands, err := entry.GetBinaryInfo()
	if err != nil {
		return err
	}

	// Execute commands if present
	if len(commands) > 0 {
		logger.Get().Info("Installing binary via command(s)",
			zap.String("name", name),
			zap.Int("command_count", len(commands)))
		if err := executeCommands(commands); err != nil {
			return err
		}

		// After successful command execution, symlink the binary to external-binaries
		if err := SymlinkInstalledBinaryToFolder(name, &entry, binariesFolder); err != nil {
			logger.Get().Warn("Failed to symlink binary to external-binaries folder",
				zap.String("name", name),
				zap.Error(err))
			// Don't return error - installation succeeded, this is just an optimization
		}
		return nil
	}

	logger.Get().Info("Downloading binary",
		zap.String("name", name),
		zap.String("url", url))
	logger.Get().Info("Installing to", zap.String("path", binariesFolder))
	return downloadAndExtractBinary(name, url, binariesFolder, customHeaders)
}

// executeCommand runs a shell command for installing a binary
func executeCommand(command string) error {
	var cmd *exec.Cmd
	// @NOTE: This is intentional - installation commands come from the binary registry
	// configuration which is a trusted source for binary installation procedures.
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	if os.Getenv("OSMEDEUS_SILENT") == "1" {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// executeCommands runs multiple shell commands sequentially
func executeCommands(commands []string) error {
	for i, command := range commands {
		logger.Get().Info("Running command",
			zap.Int("step", i+1),
			zap.Int("total", len(commands)),
			zap.String("command", command))

		if err := executeCommand(command); err != nil {
			return fmt.Errorf("command %d failed: %w", i+1, err)
		}
	}
	return nil
}

// downloadAndExtractBinary downloads and extracts a binary to the binaries folder
// Optional customHeaders map adds custom HTTP headers for the download
func downloadAndExtractBinary(name, url, binariesFolder string, customHeaders map[string]string) error {
	// Ensure binaries folder exists
	if err := os.MkdirAll(binariesFolder, 0755); err != nil {
		return fmt.Errorf("failed to create binaries folder: %w", err)
	}

	// Create temp directory for download
	tempDir, err := os.MkdirTemp("", "osmedeus-binary-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Determine filename from URL
	urlParts := strings.Split(url, "/")
	filename := urlParts[len(urlParts)-1]
	downloadPath := filepath.Join(tempDir, filename)

	// Download the file
	if err := DownloadFile(url, downloadPath, customHeaders); err != nil {
		return err
	}

	// Extract based on archive type
	archiveType := DetectArchiveType(filename)
	extractDir := filepath.Join(tempDir, "extracted")

	logger.Get().Info("Archive type detected", zap.String("type", archiveType))

	switch archiveType {
	case "zip":
		logger.Get().Info("Extracting zip", zap.String("dest", extractDir))
		if err := ExtractZip(downloadPath, extractDir); err != nil {
			return err
		}
	case "tar.gz":
		logger.Get().Info("Extracting tar.gz", zap.String("dest", extractDir))
		if err := ExtractTarGz(downloadPath, extractDir); err != nil {
			return err
		}
	case "gz":
		// Single file gz - extract directly to binaries folder
		destPath := filepath.Join(binariesFolder, name)
		logger.Get().Info("Extracting gz", zap.String("dest", destPath))
		if err := ExtractGz(downloadPath, destPath); err != nil {
			return err
		}
		return nil
	default:
		// Assume it's a raw binary
		destPath := filepath.Join(binariesFolder, name)
		logger.Get().Info("Copying raw binary", zap.String("dest", destPath))
		if err := copyFile(downloadPath, destPath); err != nil {
			return err
		}
		if err := os.Chmod(destPath, 0755); err != nil {
			return err
		}
		return nil
	}

	// Find and copy the binary from extracted directory
	return copyBinaryFromExtracted(name, extractDir, binariesFolder)
}

// copyBinaryFromExtracted finds the binary in extracted directory and copies it
func copyBinaryFromExtracted(name, extractDir, binariesFolder string) error {
	// Look for the binary file
	var binaryPath string

	err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		baseName := filepath.Base(path)
		// Match exact name or name with extension
		if baseName == name || strings.TrimSuffix(baseName, filepath.Ext(baseName)) == name {
			// Check if it's executable (on Unix) or ends with .exe (Windows)
			if info.Mode()&0111 != 0 || strings.HasSuffix(baseName, ".exe") || !strings.Contains(baseName, ".") {
				binaryPath = path
				return filepath.SkipAll
			}
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return fmt.Errorf("failed to search extracted directory: %w", err)
	}

	if binaryPath == "" {
		// If not found by name, look for any executable
		err := filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			// Skip common non-binary files
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".md" || ext == ".txt" || ext == ".json" || ext == ".yaml" || ext == ".yml" {
				return nil
			}
			if info.Mode()&0111 != 0 {
				binaryPath = path
				return filepath.SkipAll
			}
			return nil
		})
		if err != nil && err != filepath.SkipAll {
			return err
		}
	}

	if binaryPath == "" {
		return fmt.Errorf("binary '%s' not found in extracted archive", name)
	}

	destPath := filepath.Join(binariesFolder, name)
	if err := copyFile(binaryPath, destPath); err != nil {
		return err
	}

	return os.Chmod(destPath, 0755)
}

func copyFile(src, dest string) error {
	// Get source file info to preserve permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = sourceFile.Close() }()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Preserve original file permissions
	return os.Chmod(dest, srcInfo.Mode())
}

// symlinkOrCopyFile creates a symlink from src to dest, falling back to copy if symlink fails
func symlinkOrCopyFile(src, dest string) error {
	// Remove existing file/symlink at dest
	if _, err := os.Lstat(dest); err == nil {
		if err := os.Remove(dest); err != nil {
			return fmt.Errorf("failed to remove existing destination: %w", err)
		}
	}
	// Try symlink first, fall back to copy
	if err := os.Symlink(src, dest); err != nil {
		logger.Get().Warn("Symlink failed, falling back to copy",
			zap.String("src", src), zap.String("dest", dest), zap.Error(err))
		return copyFile(src, dest)
	}
	return nil
}

// SymlinkInstalledBinaryToFolder finds a binary using LookPath and symlinks it to the destination folder
// Uses the validate command (valide-command) from registry to locate the binary
// Returns nil if binary not found (installation may have failed) or symlink succeeds
func SymlinkInstalledBinaryToFolder(name string, entry *BinaryEntry, destFolder string) error {
	// Determine which command to look for
	lookupCmd := name
	if entry != nil && entry.ValidateCommand != "" {
		lookupCmd = entry.ValidateCommand
	}

	// Find the binary path using LookPath (equivalent to 'which')
	binaryPath, err := exec.LookPath(lookupCmd)
	if err != nil {
		logger.Get().Warn("Could not find installed binary in PATH",
			zap.String("name", name),
			zap.String("lookup", lookupCmd),
			zap.Error(err))
		return nil // Not an error - binary might not be in PATH yet
	}

	// Ensure destination folder exists
	if err := os.MkdirAll(destFolder, 0755); err != nil {
		return fmt.Errorf("failed to create destination folder: %w", err)
	}

	// Copy to destination
	destPath := filepath.Join(destFolder, name)

	// Skip if source and dest are the same
	if binaryPath == destPath {
		logger.Get().Debug("Binary already in destination folder",
			zap.String("path", destPath))
		return nil
	}

	logger.Get().Info("Symlinking binary to external-binaries",
		zap.String("from", binaryPath),
		zap.String("to", destPath))

	if err := symlinkOrCopyFile(binaryPath, destPath); err != nil {
		return fmt.Errorf("failed to symlink binary: %w", err)
	}

	return nil
}

// ListBinaries returns all binary names in the registry
func (r BinaryRegistry) ListBinaries() []string {
	names := make([]string, 0, len(r))
	for name := range r {
		names = append(names, name)
	}
	return names
}
