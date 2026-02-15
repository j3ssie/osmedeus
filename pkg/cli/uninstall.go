package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	uninstallClean bool
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove Osmedeus installation (base folder, workspaces, and binary)",
	Long:  UsageUninstall(),
	RunE:  runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallClean, "clean", false, "also remove workspaces data (~/workspaces-osmedeus)")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	// Resolve folders to remove
	baseFolderPath := filepath.Join(homeDir, "osmedeus-base")
	dotOsmPath := filepath.Join(homeDir, ".osmedeus")
	workspacesPath := filepath.Join(homeDir, "workspaces-osmedeus")

	// If config is loaded, use its actual paths
	cfg := config.Get()
	if cfg != nil {
		if cfg.BaseFolder != "" {
			baseFolderPath = cfg.BaseFolder
		}
		if ws := cfg.GetWorkspacesDir(); ws != "" {
			workspacesPath = ws
		}
	}

	// Find osmedeus binaries in PATH (up to 3)
	binaries := findOsmedeusBinaries(3)

	// ── Print warning ──────────────────────────────────────────────
	fmt.Println()
	fmt.Println(terminal.BoldRed("╔════════════════════════════════════════════════════════════════╗"))
	fmt.Println(terminal.BoldRed("║         WARNING: This will PERMANENTLY remove Osmedeus        ║"))
	fmt.Println(terminal.BoldRed("╚════════════════════════════════════════════════════════════════╝"))
	fmt.Println()

	fmt.Println(terminal.BoldRed("The following will be DELETED:"))
	fmt.Println()

	printFolderLine(printer, baseFolderPath, "base folder (settings, workflows, binaries)")
	printFolderLine(printer, dotOsmPath, "initialization marker")
	if uninstallClean {
		printFolderLine(printer, workspacesPath, "workspaces data (scan results)")
	}

	if len(binaries) > 0 {
		for _, bin := range binaries {
			fmt.Printf("  %s %s  %s\n", terminal.Red("✘"), terminal.Yellow(bin), terminal.Gray("osmedeus binary"))
		}
	} else {
		fmt.Printf("  %s %s\n", terminal.Gray("○"), terminal.Gray("no osmedeus binary found in PATH"))
	}
	fmt.Println()

	if !uninstallClean {
		printer.Info("Workspaces folder %s will be kept (use %s to remove it too)",
			terminal.Yellow(workspacesPath), terminal.Cyan("--clean"))
	}
	fmt.Println()

	// ── Require --force ────────────────────────────────────────────
	if !globalForce {
		printer.Warning("This operation is IRREVERSIBLE!")
		printer.Warning("Use %s %s to confirm deletion", terminal.Cyan("--force"), terminal.Cyan("--clean"))
		fmt.Println()
		if !confirmPrompt(terminal.BoldRed("Are you sure you want to uninstall Osmedeus?")) {
			printer.Info("Uninstall cancelled.")
			return nil
		}
		fmt.Println()
	}

	// ── Perform removal ────────────────────────────────────────────
	removeFolderIfExists(printer, baseFolderPath, "base folder")
	removeFolderIfExists(printer, dotOsmPath, ".osmedeus marker")

	if uninstallClean {
		removeFolderIfExists(printer, workspacesPath, "workspaces")
	}

	// Remove binaries (up to 3)
	for _, bin := range binaries {
		if err := os.Remove(bin); err != nil {
			printer.Warning("Failed to remove binary %s: %s", terminal.Yellow(bin), err)
			if runtime.GOOS != "windows" {
				printer.Info("Try: %s", terminal.Cyan(fmt.Sprintf("sudo rm %s", bin)))
			}
		} else {
			printer.Success("Removed binary: %s", terminal.Yellow(bin))
		}
	}

	fmt.Println()
	printer.Success("Osmedeus has been uninstalled.")
	if !uninstallClean {
		printer.Info("Note: Workspaces folder was preserved at %s", terminal.Yellow(workspacesPath))
		printer.Info("To remove it manually: %s", terminal.Cyan(fmt.Sprintf("rm -rf %s", workspacesPath)))
	}

	return nil
}

// findOsmedeusBinaries searches PATH for osmedeus binaries, returning up to maxCount unique paths.
func findOsmedeusBinaries(maxCount int) []string {
	var results []string
	seen := make(map[string]bool)

	for i := 0; i < maxCount; i++ {
		path, err := exec.LookPath("osmedeus")
		if err != nil {
			break
		}

		// Resolve symlinks to get the real path
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			resolved = path
		}

		absPath, err := filepath.Abs(resolved)
		if err != nil {
			absPath = resolved
		}

		if seen[absPath] {
			break // No more unique binaries to find
		}
		seen[absPath] = true
		results = append(results, absPath)

		// Temporarily remove the directory from PATH so LookPath finds the next one
		dir := filepath.Dir(absPath)
		removeDirFromPath(dir)
	}

	return results
}

// removeDirFromPath removes a directory from the PATH environment variable.
func removeDirFromPath(dir string) {
	pathSep := string(os.PathListSeparator)
	paths := filepath.SplitList(os.Getenv("PATH"))

	var filtered []string
	for _, p := range paths {
		if p != dir {
			filtered = append(filtered, p)
		}
	}

	_ = os.Setenv("PATH", joinPaths(filtered, pathSep))
}

// joinPaths joins path segments with the given separator.
func joinPaths(paths []string, sep string) string {
	result := ""
	for i, p := range paths {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

// printFolderLine prints a folder entry with exists/missing indicator.
func printFolderLine(printer *terminal.Printer, path, description string) {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("  %s %s  %s\n", terminal.Red("✘"), terminal.Yellow(path), terminal.Gray(description))
	} else {
		fmt.Printf("  %s %s  %s\n", terminal.Gray("○"), terminal.Gray(path), terminal.Gray("(not found)"))
	}
}

// removeFolderIfExists removes a folder and prints the result.
func removeFolderIfExists(printer *terminal.Printer, path, label string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		printer.Info("Skipped %s (not found): %s", label, terminal.Gray(path))
		return
	}
	if err := os.RemoveAll(path); err != nil {
		printer.Warning("Failed to remove %s: %s", label, err)
		if runtime.GOOS != "windows" {
			printer.Info("Try: %s", terminal.Cyan(fmt.Sprintf("sudo rm -rf %s", path)))
		}
	} else {
		printer.Success("Removed %s: %s", label, terminal.Yellow(path))
	}
}
