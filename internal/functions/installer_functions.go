package functions

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/dop251/goja"
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
