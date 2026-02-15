package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// GoInstallOutput stores the last go install output for display
var GoInstallOutput string

// IsGoInstalled checks if Go toolchain is available in PATH
func IsGoInstalled() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

// GetGoVersion returns the installed Go version
func GetGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetGoBinPath returns the GOBIN or GOPATH/bin directory
func GetGoBinPath() (string, error) {
	// Try GOBIN first
	cmd := exec.Command("go", "env", "GOBIN")
	output, err := cmd.Output()
	if err == nil {
		gobin := strings.TrimSpace(string(output))
		if gobin != "" {
			return gobin, nil
		}
	}

	// Fall back to GOPATH/bin
	cmd = exec.Command("go", "env", "GOPATH")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get GOPATH: %w", err)
	}
	gopath := strings.TrimSpace(string(output))
	if gopath == "" {
		return "", fmt.Errorf("GOPATH is not set")
	}
	return filepath.Join(gopath, "bin"), nil
}

// InstallBinaryViaGo installs a binary using `go install` and optionally copies it to binariesFolder
func InstallBinaryViaGo(binaryName string, goPackage string, binariesFolder string) error {
	if !IsGoInstalled() {
		return fmt.Errorf("go toolchain is not installed")
	}

	// Ensure package has a version suffix
	pkg := goPackage
	if !strings.Contains(pkg, "@") {
		pkg = pkg + "@latest"
	}

	logger.Get().Info("Installing via go install",
		zap.String("binary", binaryName),
		zap.String("package", pkg))

	cmd := exec.Command("go", "install", pkg)
	cmd.Env = os.Environ()

	// Capture output for display
	output, err := cmd.CombinedOutput()
	GoInstallOutput = string(output)

	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("failed to install %s via go install: %w\nOutput: %s", pkg, err, string(output))
		}
		return fmt.Errorf("failed to install %s via go install: %w", pkg, err)
	}

	// Symlink binary from GOBIN to binaries folder if specified
	if binariesFolder != "" {
		if err := symlinkGoBinaryToFolder(binaryName, binariesFolder); err != nil {
			return fmt.Errorf("failed to symlink binary to folder: %w", err)
		}
	}

	return nil
}

// symlinkGoBinaryToFolder finds a binary installed by go and symlinks it to the target folder
func symlinkGoBinaryToFolder(binaryName string, binariesFolder string) error {
	goBinPath, err := GetGoBinPath()
	if err != nil {
		return err
	}

	srcPath := filepath.Join(goBinPath, binaryName)
	if _, err := os.Stat(srcPath); err != nil {
		// Try finding via which
		cmd := exec.Command("which", binaryName)
		output, whichErr := cmd.Output()
		if whichErr != nil {
			return fmt.Errorf("binary %s not found after go install (checked %s and PATH)", binaryName, srcPath)
		}
		srcPath = strings.TrimSpace(string(output))
	}

	// Ensure binaries folder exists
	if err := os.MkdirAll(binariesFolder, 0755); err != nil {
		return fmt.Errorf("failed to create binaries folder: %w", err)
	}

	destPath := filepath.Join(binariesFolder, binaryName)

	logger.Get().Info("Symlinking binary to external folder",
		zap.String("src", srcPath),
		zap.String("dest", destPath))

	return symlinkOrCopyFile(srcPath, destPath)
}

// GetGoPackageName returns the go_install package path for a binary
// Returns the GoInstall field if set, otherwise returns empty string
func GetGoPackageName(entry BinaryEntry, binaryName string) string {
	if entry.GoInstall != "" {
		return entry.GoInstall
	}
	return ""
}

// ExtractBinaryNameFromGoPackage extracts binary name from a go package path
// e.g., "github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest" -> "nuclei"
func ExtractBinaryNameFromGoPackage(pkg string) string {
	// Remove version suffix
	pkg = strings.Split(pkg, "@")[0]
	// Get last path component
	parts := strings.Split(pkg, "/")
	return parts[len(parts)-1]
}
