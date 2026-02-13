package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/snapshot"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	snapshotOutputPath string
	snapshotSkipDB     bool
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Export and import workspace snapshots",
	Long:  UsageSnapshot(),
}

var snapshotExportCmd = &cobra.Command{
	Use:   "export <workspace>",
	Short: "Export workspace to compressed zip archive",
	Long: `Export a workspace folder to a compressed zip archive.

The archive is created with highest compression (level 9) and stored in
the snapshot folder (default: ~/osmedeus-base/snapshot/).

Examples:
  osmedeus snapshot export example.com
  osmedeus snapshot export example.com -o /tmp/backup.zip`,
	Args: cobra.ExactArgs(1),
	RunE: runSnapshotExport,
}

var snapshotImportCmd = &cobra.Command{
	Use:   "import <source>",
	Short: "Import workspace from zip file or URL",
	Long: `Import a workspace from a zip file or URL.

The source can be a local file path or a URL to download.
The workspace will be extracted to the workspaces folder and
optionally imported to the database with data_source="imported".

Examples:
  osmedeus snapshot import ~/snapshot/example.com_2026-02-13T18-20-34Z.zip
  osmedeus snapshot import https://example.com/workspace.zip
  osmedeus snapshot import ~/backup.zip --force`,
	Args: cobra.ExactArgs(1),
	RunE: runSnapshotImport,
}

var snapshotListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available snapshots",
	Long:    `List all snapshot files in the snapshot folder.`,
	RunE:    runSnapshotList,
}

func init() {
	// Export flags
	snapshotExportCmd.Flags().StringVarP(&snapshotOutputPath, "output", "o", "", "Custom output path for the snapshot")

	// Import flags
	// Note: --force flag is now global (defined in root.go)
	snapshotImportCmd.Flags().BoolVar(&snapshotSkipDB, "skip-db", false, "Skip database import (files only)")

	// Add subcommands
	snapshotCmd.AddCommand(snapshotExportCmd)
	snapshotCmd.AddCommand(snapshotImportCmd)
	snapshotCmd.AddCommand(snapshotListCmd)
}

func runSnapshotExport(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	workspaceName := args[0]
	workspacePath := filepath.Join(cfg.WorkspacesPath, workspaceName)

	// Check if workspace exists
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		printer.Error("Workspace not found: %s", workspacePath)
		return fmt.Errorf("workspace not found: %s", workspaceName)
	}

	// Determine output path
	outputPath := snapshotOutputPath
	if outputPath == "" {
		// Ensure snapshot directory exists
		if err := os.MkdirAll(cfg.SnapshotPath, 0755); err != nil {
			return fmt.Errorf("failed to create snapshot directory: %w", err)
		}
		outputPath = filepath.Join(cfg.SnapshotPath, fmt.Sprintf("%s_%s.zip", workspaceName, getTimestamp()))
	}

	printer.Info("Exporting workspace: %s", workspaceName)
	printer.Info("Source: %s", workspacePath)
	printer.Info("Destination: %s", outputPath)

	result, err := snapshot.ExportWorkspace(workspacePath, outputPath)
	if err != nil {
		printer.Error("Export failed: %s", err)
		return err
	}

	printer.Success("Snapshot created successfully!")
	printer.Info("File: %s", result.OutputPath)
	printer.Info("Size: %s", formatBytes(result.FileSize))

	return nil
}

func runSnapshotImport(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	source := args[0]

	// Show warning
	printer.Warning("Only import snapshots from %s. Imported data may conflict with existing DB records.", terminal.BoldYellow("trusted sources"))
	fmt.Println()

	// Skip confirmation when --force is provided
	if !globalForce {
		if !confirmPrompt("Continue with import?") {
			printer.Info("Import cancelled.")
			return nil
		}
	}

	printer.Info("Importing from: %s", source)

	var result *snapshot.ImportResult
	var err error

	if globalForce {
		result, err = snapshot.ForceImportWorkspace(source, cfg.WorkspacesPath, snapshotSkipDB, cfg)
	} else {
		result, err = snapshot.ImportWorkspace(source, cfg.WorkspacesPath, snapshotSkipDB, cfg)
	}

	if err != nil {
		printer.Error("Import failed: %s", err)
		return err
	}

	printer.Success("Workspace imported successfully!")
	printer.Info("Workspace: %s", result.WorkspaceName)
	printer.Info("Location: %s", result.LocalPath)
	printer.Info("Data Source: %s", result.DataSource)
	printer.Info("Files: %d", result.FilesCount)

	if snapshotSkipDB {
		printer.Info("Database import skipped (--skip-db)")
	}

	return nil
}

func runSnapshotList(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	snapshots, err := snapshot.ListSnapshots(cfg.SnapshotPath)
	if err != nil {
		printer.Error("Failed to list snapshots: %s", err)
		return err
	}

	if len(snapshots) == 0 {
		printer.Info("No snapshots found in: %s", cfg.SnapshotPath)
		return nil
	}

	printer.Section("Available Snapshots")
	fmt.Println()

	for _, s := range snapshots {
		fmt.Printf("  %-50s %10s  %s\n",
			s.Name,
			formatBytes(s.Size),
			s.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	fmt.Println()
	printer.Info("Total: %d snapshots in %s", len(snapshots), cfg.SnapshotPath)

	return nil
}

// confirmPrompt asks for user confirmation
func confirmPrompt(message string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", message)

	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// formatBytes formats bytes to human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getTimestamp returns current UTC time in ISO 8601 format (e.g. 2026-02-13T18-20-34Z)
func getTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15-04-05Z")
}

// UsageSnapshot returns usage text for snapshot command
func UsageSnapshot() string {
	return `Workspace snapshot management for backup and sharing.

Export creates a compressed zip archive of a workspace folder.
Import extracts a snapshot and optionally imports to database.

Commands:
  export <workspace>  Export workspace to zip archive
  import <source>     Import workspace from zip file or URL
  list                List available snapshots

Examples:
  osmedeus snapshot export example.com
  osmedeus snapshot import ~/backup.zip
  osmedeus snapshot list
`
}
