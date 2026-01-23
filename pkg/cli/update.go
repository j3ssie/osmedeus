package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/j3ssie/osmedeus/v5/internal/updater"
	"github.com/spf13/cobra"
)

var (
	updateCheck   bool
	updateYes     bool
	updateVersion string
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update osmedeus to the latest version",
	Long:  UsageUpdate(),
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheck, "check", false, "only check for updates without installing")
	updateCmd.Flags().BoolVarP(&updateYes, "yes", "y", false, "skip confirmation prompt")
	// Note: --force flag is now global (defined in root.go)
	updateCmd.Flags().StringVar(&updateVersion, "version", "", "update to a specific version (e.g., v5.1.0)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	// Parse owner/repo from REPO_URL
	owner, repo, err := updater.ParseRepoURL(core.REPO_URL)
	if err != nil {
		return fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Create updater with verbose flag
	upd := updater.NewUpdater(updater.Options{
		Owner:   owner,
		Repo:    repo,
		Verbose: verbose,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	currentVersion := core.VERSION
	printer.Info("Current version: %s", terminal.Cyan(currentVersion))

	// Check for updates
	var release *updater.Release
	var hasUpdate bool

	if updateVersion != "" {
		// Check for specific version
		printer.Info("Checking for version %s...", terminal.Cyan(updateVersion))
		release, _, err = upd.CheckSpecificVersion(ctx, updateVersion)
		if err != nil {
			return fmt.Errorf("failed to check for version: %w", err)
		}
		if release == nil {
			return fmt.Errorf("version %s not found", updateVersion)
		}
		// For specific version, hasUpdate means version exists
		hasUpdate = true
	} else {
		// Check for latest
		printer.Info("Checking for updates...")
		release, hasUpdate, err = upd.CheckForUpdate(ctx, currentVersion)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}
	}

	if release == nil {
		printer.Info("No releases found")
		return nil
	}

	if !hasUpdate && !globalForce {
		printer.Success("You are running the latest version (%s)", currentVersion)
		return nil
	}

	// Display release info
	printer.Newline()
	printer.Section("New Version Available")
	printer.KeyValue("Version", terminal.Green(release.Version))
	if !release.PublishedAt.IsZero() {
		printer.KeyValue("Published", release.PublishedAt.Format("2006-01-02"))
	}
	if release.ReleaseNotes != "" {
		printer.Newline()
		printer.SubSection("Release Notes")
		// Render release notes as markdown
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)
		if err == nil {
			rendered, err := renderer.Render(release.ReleaseNotes)
			if err == nil {
				fmt.Print(rendered)
			} else {
				fmt.Println(release.ReleaseNotes)
			}
		} else {
			fmt.Println(release.ReleaseNotes)
		}
	}
	printer.Newline()

	// Check-only mode
	if updateCheck {
		printer.Info("Run 'osmedeus update' to install this version")
		return nil
	}

	// Confirm update
	if !updateYes {
		printer.Print("Do you want to update to %s? [y/N]: ", terminal.Green(release.Version))
		if !confirmUpdatePrompt() {
			printer.Info("Update cancelled")
			return nil
		}
	}

	// Perform update
	printer.Info("Downloading %s...", release.Version)

	var result *updater.UpdateResult
	if updateVersion != "" {
		result, err = upd.UpdateToVersion(ctx, currentVersion, updateVersion, globalForce)
	} else {
		result, err = upd.Update(ctx, currentVersion, globalForce)
	}

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	if result.Updated {
		printer.Success("Successfully updated from %s to %s",
			terminal.Gray(result.OldVersion),
			terminal.Green(result.NewVersion))
		printer.Info("Restart osmedeus to use the new version")
	} else {
		printer.Info("No update was performed")
	}

	return nil
}

// confirmUpdatePrompt reads y/n confirmation from stdin
func confirmUpdatePrompt() bool {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// UsageUpdate returns the Long description for the update command
func UsageUpdate() string {
	return terminal.BoldCyan("◆ Description") + `
  Update osmedeus binary to the latest version from GitHub releases.
  Compares semantic versions and downloads the appropriate binary for
  your platform.

` + terminal.BoldCyan("▶ Options") + `
  • ` + terminal.Yellow("--check") + `    - Only check for updates without installing
  • ` + terminal.Yellow("--yes") + `      - Skip confirmation prompt
  • ` + terminal.Yellow("--force") + `    - Force update even if already on latest version
  • ` + terminal.Yellow("--version") + `  - Update to a specific version

` + terminal.BoldCyan("▷ Examples") + `
  ` + terminal.Green("# Check for updates") + `
  osmedeus update ` + terminal.Yellow("--check") + `

  ` + terminal.Green("# Update to latest version") + `
  osmedeus update

  ` + terminal.Green("# Force reinstall current version") + `
  osmedeus update ` + terminal.Yellow("--force") + `

  ` + terminal.Green("# Update to specific version") + `
  osmedeus update ` + terminal.Yellow("--version") + ` v5.1.0

` + docsFooter()
}
