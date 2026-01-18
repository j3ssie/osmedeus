package utils

import (
	"os"
	"os/exec"
	"path/filepath"
)

// LookPathWithBinaries searches for cmd in binariesPath first, then system PATH.
// This ensures external-binaries folder is checked during dependency validation,
// even before the shell's PATH is modified.
func LookPathWithBinaries(cmd, binariesPath string) (string, error) {
	// First, check if cmd exists in the binaries path
	if binariesPath != "" {
		binaryPath := filepath.Join(binariesPath, cmd)
		if info, err := os.Stat(binaryPath); err == nil && !info.IsDir() {
			// Check if it's executable
			if info.Mode()&0111 != 0 {
				return binaryPath, nil
			}
		}
	}

	// Fall back to standard PATH lookup
	return exec.LookPath(cmd)
}
