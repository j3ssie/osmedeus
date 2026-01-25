package functions

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/installer"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// goGetter downloads files/repos using go-getter
// Usage: go_getter(url, dest) -> bool
// url: source URL (supports git repos, archives, files)
// dest: destination path
func (vf *vmFunc) goGetter(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("go_getter"))

	if len(call.Arguments) < 2 {
		logger.Get().Warn("go_getter: requires 2 arguments")
		return vf.vm.ToValue(false)
	}

	url := call.Argument(0).String()
	dest := call.Argument(1).String()

	if url == "" || url == "undefined" {
		logger.Get().Warn("go_getter: url cannot be empty")
		return vf.vm.ToValue(false)
	}

	if dest == "" || dest == "undefined" {
		logger.Get().Warn("go_getter: dest cannot be empty")
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug(terminal.HiGreen("go_getter")+" params",
		zap.String("url", url),
		zap.String("dest", dest))

	err := installer.GetViaGoGetter(url, dest)
	if err != nil {
		logger.Get().Warn("go_getter: download failed", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}

// goGetterWithSSHKey downloads git repos using go-getter with SSH key authentication
// Usage: go_getter_with_sshkey(ssh_key_path, git_url, dest) -> bool
// ssh_key_path: path to SSH private key file
// git_url: git repository URL (will be prefixed with git:: if needed)
// dest: destination path
// The SSH key is base64 encoded and appended as ?sshkey= query parameter
func (vf *vmFunc) goGetterWithSSHKey(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("go_getter_with_sshkey"))

	if len(call.Arguments) < 3 {
		logger.Get().Warn("go_getter_with_sshkey: requires 3 arguments (ssh_key_path, git_url, dest)")
		return vf.vm.ToValue(false)
	}

	sshKeyPath := call.Argument(0).String()
	gitURL := call.Argument(1).String()
	dest := call.Argument(2).String()

	// Validate arguments
	if sshKeyPath == "" || sshKeyPath == "undefined" {
		logger.Get().Warn("go_getter_with_sshkey: ssh_key_path cannot be empty")
		return vf.vm.ToValue(false)
	}
	if gitURL == "" || gitURL == "undefined" {
		logger.Get().Warn("go_getter_with_sshkey: git_url cannot be empty")
		return vf.vm.ToValue(false)
	}
	if dest == "" || dest == "undefined" {
		logger.Get().Warn("go_getter_with_sshkey: dest cannot be empty")
		return vf.vm.ToValue(false)
	}

	// Expand path (handle ~ and env vars)
	sshKeyPath = installer.ExpandPath(sshKeyPath)

	// Read SSH key file
	keyContent, err := os.ReadFile(sshKeyPath)
	if err != nil {
		logger.Get().Warn("go_getter_with_sshkey: failed to read SSH key file",
			zap.String("path", sshKeyPath),
			zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Base64 encode the key
	encodedKey := base64.StdEncoding.EncodeToString(keyContent)

	// Build the URL with sshkey parameter
	// go-getter format: git::git@github.com:user/repo.git?sshkey=<base64>
	separator := "?"
	if strings.Contains(gitURL, "?") {
		separator = "&"
	}

	// Add git:: prefix if not present and URL looks like SSH
	if strings.HasPrefix(gitURL, "git@") && !strings.HasPrefix(gitURL, "git::") {
		gitURL = "git::" + gitURL
	}

	fullURL := gitURL + separator + "sshkey=" + encodedKey

	logger.Get().Debug(terminal.HiGreen("go_getter_with_sshkey")+" params",
		zap.String("ssh_key_path", sshKeyPath),
		zap.String("git_url", gitURL),
		zap.String("dest", dest),
		zap.Int("key_length", len(keyContent)))

	// Call go-getter with the modified URL
	err = installer.GetViaGoGetter(fullURL, dest)
	if err != nil {
		logger.Get().Warn("go_getter_with_sshkey: download failed", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}

// nixInstall installs a package using Nix
// Usage: nix_install(package, dest?) -> bool
// package: Nix package name (e.g., "nuclei", "subfinder")
// dest: optional destination folder for binary copy (default: no copy)
func (vf *vmFunc) nixInstall(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("nix_install"))

	if len(call.Arguments) < 1 {
		logger.Get().Warn("nix_install: requires at least 1 argument")
		return vf.vm.ToValue(false)
	}

	pkg := call.Argument(0).String()
	if pkg == "" || pkg == "undefined" {
		logger.Get().Warn("nix_install: package cannot be empty")
		return vf.vm.ToValue(false)
	}

	// Optional destination folder
	dest := ""
	if len(call.Arguments) >= 2 {
		destArg := call.Argument(1).String()
		if destArg != "" && destArg != "undefined" {
			dest = destArg
		}
	}

	logger.Get().Debug(terminal.HiGreen("nix_install")+" params",
		zap.String("package", pkg),
		zap.String("dest", dest))

	// Check if Nix is installed
	if !installer.IsNixInstalled() {
		logger.Get().Warn("nix_install: Nix is not installed")
		return vf.vm.ToValue(false)
	}

	// Install the package via Nix
	err := installer.InstallBinaryViaNix(pkg, pkg, dest)
	if err != nil {
		logger.Get().Warn("nix_install: installation failed", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}

// filepathInstaller installs a local file (archive or binary) to the binaries folder
// Usage: filepath_installer(local_path, tool_name, dest?) -> bool
// local_path: path to a local .tar.gz, .zip, .gz, or executable binary file
// tool_name: name for the installed binary (used when archive contains files with different names)
// dest: optional destination directory
//
//	Defaults to external_binaries_path ({{base_folder}}/external-binaries) from config
func (vf *vmFunc) filepathInstaller(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen(FnFilepathInstaller))

	if len(call.Arguments) < 2 {
		logger.Get().Warn("filepath_installer: requires at least 2 arguments: local_path, tool_name")
		return vf.vm.ToValue(false)
	}

	localPath := call.Argument(0).String()
	toolName := call.Argument(1).String()

	// Validate required args
	if localPath == "" || localPath == "undefined" {
		logger.Get().Warn("filepath_installer: local_path cannot be empty")
		return vf.vm.ToValue(false)
	}
	if toolName == "" || toolName == "undefined" {
		logger.Get().Warn("filepath_installer: tool_name cannot be empty")
		return vf.vm.ToValue(false)
	}

	// Optional destination directory
	dest := ""
	if len(call.Arguments) >= 3 {
		dest = call.Argument(2).String()
		if dest == "undefined" {
			dest = ""
		}
	}

	// Default destination: external_binaries_path ({{base_folder}}/external-binaries)
	// Priority: 1) explicit dest param, 2) Binaries from VM context, 3) config.BinariesPath
	if dest == "" {
		if val := vf.vm.Get("Binaries"); val != nil && val != goja.Undefined() {
			dest = val.String()
		}
	}

	// Fallback to global config's external_binaries_path
	if dest == "" {
		cfg := config.Get()
		if cfg != nil && cfg.BinariesPath != "" {
			dest = cfg.BinariesPath
		}
	}

	if dest == "" {
		logger.Get().Warn("filepath_installer: destination directory not specified and Binaries path not available (ensure config is loaded)")
		return vf.vm.ToValue(false)
	}

	// Check if source file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		logger.Get().Warn("filepath_installer: source file not found", zap.String("path", localPath))
		return vf.vm.ToValue(false)
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(dest, 0755); err != nil {
		logger.Get().Warn("filepath_installer: failed to create destination directory",
			zap.String("dest", dest), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	destPath := filepath.Join(dest, toolName)

	// Detect archive type
	archiveType := installer.DetectArchiveType(localPath)

	logger.Get().Debug(terminal.HiGreen(FnFilepathInstaller)+" params",
		zap.String("local_path", localPath),
		zap.String("tool_name", toolName),
		zap.String("dest", dest),
		zap.String("archive_type", archiveType))

	switch archiveType {
	case "tar.gz":
		// Extract to temp dir, find binary, copy to dest
		if err := extractAndInstallBinary(localPath, toolName, destPath, installer.ExtractTarGz); err != nil {
			logger.Get().Warn("filepath_installer: failed to extract tar.gz",
				zap.String("path", localPath), zap.Error(err))
			return vf.vm.ToValue(false)
		}

	case "zip":
		// Extract to temp dir, find binary, copy to dest
		if err := extractAndInstallBinary(localPath, toolName, destPath, installer.ExtractZip); err != nil {
			logger.Get().Warn("filepath_installer: failed to extract zip",
				zap.String("path", localPath), zap.Error(err))
			return vf.vm.ToValue(false)
		}

	case "gz":
		// Single gzip file - extract directly to destination
		if err := installer.ExtractGz(localPath, destPath); err != nil {
			logger.Get().Warn("filepath_installer: failed to extract gz",
				zap.String("path", localPath), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		// ExtractGz already sets 0755 permissions

	default:
		// Assume it's a binary - copy directly
		if err := copyBinaryFile(localPath, destPath); err != nil {
			logger.Get().Warn("filepath_installer: failed to copy binary",
				zap.String("src", localPath), zap.String("dest", destPath), zap.Error(err))
			return vf.vm.ToValue(false)
		}
		// Set executable permissions
		if err := os.Chmod(destPath, 0755); err != nil {
			logger.Get().Warn("filepath_installer: failed to set executable permissions",
				zap.String("path", destPath), zap.Error(err))
			return vf.vm.ToValue(false)
		}
	}

	logger.Get().Info("filepath_installer: successfully installed",
		zap.String("tool", toolName), zap.String("dest", destPath))
	return vf.vm.ToValue(true)
}

// extractAndInstallBinary extracts an archive to a temp dir, finds the binary, and copies it to destPath
func extractAndInstallBinary(archivePath, toolName, destPath string, extractFn func(string, string) error) error {
	// Create temp directory for extraction
	tmpDir, err := os.MkdirTemp("", "filepath_installer_*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// Extract archive
	if err := extractFn(archivePath, tmpDir); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Find the binary - try exact name first, then search
	binaryPath := findBinaryInDir(tmpDir, toolName)
	if binaryPath == "" {
		return fmt.Errorf("could not find binary '%s' in extracted archive", toolName)
	}

	// Copy binary to destination
	if err := copyBinaryFile(binaryPath, destPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}

	// Set executable permissions
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}

	return nil
}

// findBinaryInDir searches for a binary in the extracted directory
func findBinaryInDir(dir, toolName string) string {
	// First, try exact match at root level
	exactPath := filepath.Join(dir, toolName)
	if info, err := os.Stat(exactPath); err == nil && !info.IsDir() {
		return exactPath
	}

	// Walk directory to find the binary
	var found string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// Match by name
		if info.Name() == toolName {
			found = path
			return filepath.SkipAll
		}
		return nil
	})

	return found
}

// copyBinaryFile copies a file from src to dst
func copyBinaryFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
