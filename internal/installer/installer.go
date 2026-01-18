package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/j3ssie/osmedeus/v5/internal/terminal"
)

const (
	// DatabaseFileName is the name of the database file to preserve during base installation
	DatabaseFileName = "database-osm.sqlite"
)

// Installer handles installation of workflows, base folder, and binaries
type Installer struct {
	BaseFolder     string
	WorkflowFolder string
	BinariesFolder string
	CustomHeaders  map[string]string
	Printer        *terminal.Printer
}

// NewInstaller creates a new Installer with the given paths
func NewInstaller(baseFolder, workflowFolder, binariesFolder string, customHeaders map[string]string) *Installer {
	return &Installer{
		BaseFolder:     baseFolder,
		WorkflowFolder: workflowFolder,
		BinariesFolder: binariesFolder,
		CustomHeaders:  customHeaders,
		Printer:        terminal.NewPrinter(),
	}
}

// InstallWorkflow installs workflows from the given source
// Source can be: git URL, zip URL, or local zip file
func (i *Installer) InstallWorkflow(source string) error {
	i.Printer.Info("Installing workflows from: %s", terminal.Gray(source))

	// Fetch source to temp directory
	i.Printer.Info("Fetching source...")
	tempDir, err := FetchToTemp(source, i.CustomHeaders)
	if err != nil {
		return fmt.Errorf("failed to fetch source: %w", err)
	}
	defer cleanupTempParent(tempDir)

	// Verify the source contains workflows
	if !containsWorkflows(tempDir) {
		return fmt.Errorf("source does not appear to contain workflows (no .yaml files found)")
	}

	// Remove existing workflows (no backup - complete overwrite)
	if _, err := os.Stat(i.WorkflowFolder); err == nil {
		i.Printer.Info("Removing existing workflows...")
		if err := os.RemoveAll(i.WorkflowFolder); err != nil {
			return fmt.Errorf("failed to remove existing workflows: %w", err)
		}
	}

	// Move new workflows to destination
	i.Printer.Info("Installing new workflows...")
	if err := os.MkdirAll(filepath.Dir(i.WorkflowFolder), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := os.Rename(tempDir, i.WorkflowFolder); err != nil {
		// If rename fails (cross-device), try copy
		if err := copyDir(tempDir, i.WorkflowFolder); err != nil {
			return fmt.Errorf("failed to install workflows: %w", err)
		}
	}

	i.Printer.Success("Workflows installed successfully to %s", terminal.Gray(i.WorkflowFolder))
	return nil
}

// InstallBase installs the base folder from the given source
// Backs up and restores database-osm.sqlite and external-binaries automatically
// Workflows are NOT backed up - they are completely overwritten
func (i *Installer) InstallBase(source string) error {
	i.Printer.Info("Installing base folder from: %s", terminal.Gray(source))

	// Check for existing database to backup
	dbPath := filepath.Join(i.BaseFolder, DatabaseFileName)
	var dbBackupPath string

	if _, err := os.Stat(dbPath); err == nil {
		i.Printer.Info("Backing up database...")
		dbBackupPath, err = backupFile(dbPath)
		if err != nil {
			return fmt.Errorf("failed to backup database: %w", err)
		}
		defer func() { _ = os.Remove(dbBackupPath) }()
	}

	// Check for existing external-binaries to backup
	// Use configured BinariesFolder if set, otherwise fall back to default path
	binariesPath := i.BinariesFolder
	if binariesPath == "" {
		binariesPath = filepath.Join(i.BaseFolder, "external-binaries")
	}
	var binariesBackupPath string

	if _, err := os.Stat(binariesPath); err == nil {
		i.Printer.Info("Backing up external-binaries...")
		binariesBackupPath, err = backupDir(binariesPath)
		if err != nil {
			return fmt.Errorf("failed to backup external-binaries: %w", err)
		}
		defer func() { _ = os.RemoveAll(binariesBackupPath) }()
		// Log backup count for verification
		if entries, err := os.ReadDir(binariesBackupPath); err == nil {
			i.Printer.Info("Backed up %d items from external-binaries", len(entries))
		}
	}

	// Fetch source to temp directory
	i.Printer.Info("Fetching source...")
	tempDir, err := FetchToTemp(source, i.CustomHeaders)
	if err != nil {
		return fmt.Errorf("failed to fetch source: %w", err)
	}
	defer cleanupTempParent(tempDir)

	// Remove existing workflows folder if it's configured separately from base folder
	// This ensures complete overwrite without preserving old workflows
	if i.WorkflowFolder != "" && !isSubPath(i.BaseFolder, i.WorkflowFolder) {
		if _, err := os.Stat(i.WorkflowFolder); err == nil {
			i.Printer.Info("Removing existing workflows folder (separate from base)...")
			if err := os.RemoveAll(i.WorkflowFolder); err != nil {
				i.Printer.Warning("Failed to remove existing workflows: %s", err)
			}
		}
	}

	// Remove existing base folder
	if _, err := os.Stat(i.BaseFolder); err == nil {
		i.Printer.Info("Removing existing base folder...")
		if err := os.RemoveAll(i.BaseFolder); err != nil {
			return fmt.Errorf("failed to remove existing base folder: %w", err)
		}
	}

	// Move new base to destination
	i.Printer.Info("Installing new base folder...")
	if err := os.MkdirAll(filepath.Dir(i.BaseFolder), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	if err := os.Rename(tempDir, i.BaseFolder); err != nil {
		// If rename fails (cross-device), try copy
		if err := copyDir(tempDir, i.BaseFolder); err != nil {
			return fmt.Errorf("failed to install base folder: %w", err)
		}
	}

	// Restore database if we backed it up
	if dbBackupPath != "" {
		i.Printer.Info("Restoring database...")
		if err := restoreFile(dbBackupPath, dbPath); err != nil {
			i.Printer.Warning("Failed to restore database: %s", err)
		}
	}

	// Restore external-binaries if we backed it up
	if binariesBackupPath != "" {
		i.Printer.Info("Restoring external-binaries...")
		if err := restoreDir(binariesBackupPath, binariesPath); err != nil {
			i.Printer.Warning("Failed to restore external-binaries: %s", err)
		} else {
			// Log restored count for verification
			if entries, err := os.ReadDir(binariesPath); err == nil {
				i.Printer.Info("Restored %d items to external-binaries", len(entries))
			}
		}
	}

	i.Printer.Success("Base folder installed successfully to %s", terminal.Gray(i.BaseFolder))
	return nil
}

// InstallBinaryFromRegistry installs a binary from the registry
func (i *Installer) InstallBinaryFromRegistry(name, registryPath string) error {
	i.Printer.Info("Loading binary registry...")
	registry, err := LoadRegistry(registryPath, i.CustomHeaders)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	i.Printer.Info("Installing binary: %s", name)
	if err := InstallBinary(name, registry, i.BinariesFolder, i.CustomHeaders); err != nil {
		return fmt.Errorf("failed to install %s: %w", name, err)
	}

	i.Printer.Success("Binary '%s' installed successfully to %s", name, terminal.Gray(i.BinariesFolder))
	return nil
}

// InstallAllBinaries installs all binaries from the registry
func (i *Installer) InstallAllBinaries(registryPath string) error {
	i.Printer.Info("Loading binary registry...")
	registry, err := LoadRegistry(registryPath, i.CustomHeaders)
	if err != nil {
		return fmt.Errorf("failed to load registry: %w", err)
	}

	binaries := registry.ListBinaries()
	i.Printer.Info("Installing %d binaries...", len(binaries))

	var failed []string
	for _, name := range binaries {
		i.Printer.Info("Installing: %s", name)
		if err := InstallBinary(name, registry, i.BinariesFolder, i.CustomHeaders); err != nil {
			i.Printer.Warning("Failed to install %s: %s", name, err)
			failed = append(failed, name)
			continue
		}
		i.Printer.Success("Installed: %s", name)
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to install %d binaries: %v", len(failed), failed)
	}

	i.Printer.Success("All binaries installed successfully to %s", terminal.Gray(i.BinariesFolder))
	return nil
}

// Helper functions

func containsWorkflows(dir string) bool {
	hasYAML := false
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && (filepath.Ext(path) == ".yaml" || filepath.Ext(path) == ".yml") {
			hasYAML = true
			return filepath.SkipAll
		}
		return nil
	})
	return hasYAML
}

func backupFile(path string) (string, error) {
	tempFile, err := os.CreateTemp("", "osmedeus-backup-*")
	if err != nil {
		return "", err
	}
	_ = tempFile.Close()

	if err := copyFile(path, tempFile.Name()); err != nil {
		_ = os.Remove(tempFile.Name())
		return "", err
	}

	return tempFile.Name(), nil
}

func restoreFile(backupPath, destPath string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	return copyFile(backupPath, destPath)
}

// backupDir copies a directory to a temporary location and returns the temp path
func backupDir(path string) (string, error) {
	tempDir, err := os.MkdirTemp("", "osmedeus-backup-dir-*")
	if err != nil {
		return "", err
	}

	if err := copyDir(path, tempDir); err != nil {
		_ = os.RemoveAll(tempDir)
		return "", err
	}

	return tempDir, nil
}

// restoreDir copies a backed-up directory to the destination path
func restoreDir(backupPath, destPath string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return err
	}
	return copyDir(backupPath, destPath)
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}

func cleanupTempParent(tempDir string) {
	// tempDir might be a subdirectory of the actual temp dir
	// Walk up to find the osmedeus-install-* parent
	for {
		parent := filepath.Dir(tempDir)
		base := filepath.Base(tempDir)
		if parent == tempDir || base == "" {
			break
		}
		if len(base) > 17 && base[:17] == "osmedeus-install-" {
			_ = os.RemoveAll(tempDir)
			return
		}
		if len(base) > 17 && base[:17] == "osmedeus-binary-" {
			_ = os.RemoveAll(tempDir)
			return
		}
		tempDir = parent
	}
}

// isSubPath checks if child path is inside or equal to parent path
func isSubPath(parent, child string) bool {
	if parent == "" || child == "" {
		return false
	}
	// Clean and resolve both paths
	parentAbs, err := filepath.Abs(parent)
	if err != nil {
		return false
	}
	childAbs, err := filepath.Abs(child)
	if err != nil {
		return false
	}
	// Check if child starts with parent path
	rel, err := filepath.Rel(parentAbs, childAbs)
	if err != nil {
		return false
	}
	// If relative path starts with "..", child is outside parent
	return rel != ".." && !filepath.IsAbs(rel) && (len(rel) < 2 || rel[:2] != "..")
}
