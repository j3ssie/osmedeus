package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	orgCreateDescription string
	orgCreateUUID        string
	orgCreateTags        []string
	orgDeletePurge       bool
	orgAssignWorkspaces  []string
	orgUseClear          bool
)

var orgCmd = &cobra.Command{
	Use:     "org",
	Aliases: []string{"orgs", "tenant"},
	Short:   "Group workspaces under an org for cross-workspace queries",
	Long:    UsageOrg(),
	RunE:    runOrgList,
}

var orgListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all orgs",
	RunE:  runOrgList,
}

var orgCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new org",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgCreate,
}

var orgShowCmd = &cobra.Command{
	Use:   "show <name|uuid>",
	Short: "Show an org's details and data counts",
	Args:  cobra.ExactArgs(1),
	RunE:  runOrgShow,
}

var orgUseCmd = &cobra.Command{
	Use:   "use <name|uuid>",
	Short: "Set the active org for subsequent commands",
	Long: "Persists the active org so later commands scope to it without --org.\n" +
		"Also prints an export line you can eval to set OSMEDEUS_ORG_UUID for the current shell:\n\n" +
		"  eval $(osmedeus org use acme)\n\n" +
		"Run with --clear to go back to the unfiltered all-orgs view.",
	Args: cobra.MaximumNArgs(1),
	RunE: runOrgUse,
}

var orgAssignCmd = &cobra.Command{
	Use:   "assign <name|uuid> -w <workspace>...",
	Short: "Move existing workspaces into an org",
	Long: "Assigns whole workspaces to an org, cascading the stamp to every asset,\n" +
		"vulnerability and run in those workspaces. This is how data that predates\n" +
		"the org gets grouped without re-scanning.",
	Args: cobra.ExactArgs(1),
	RunE: runOrgAssign,
}

var orgRenameCmd = &cobra.Command{
	Use:   "rename <name|uuid> <new-name>",
	Short: "Rename an org",
	Args:  cobra.ExactArgs(2),
	RunE:  runOrgRename,
}

var orgDeleteCmd = &cobra.Command{
	Use:   "delete <name|uuid>",
	Short: "Delete an org",
	Long: "Deletes an org. By default its workspaces, assets, vulnerabilities and runs\n" +
		"are reassigned to the default org — nothing is lost, only the grouping.\n" +
		"Pass --purge to delete that data along with the org.",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	RunE:    runOrgDelete,
}

func init() {
	orgCreateCmd.Flags().StringVarP(&orgCreateDescription, "description", "d", "", "org description")
	orgCreateCmd.Flags().StringVar(&orgCreateUUID, "uuid", "", "explicit UUID (default: generated)")
	orgCreateCmd.Flags().StringArrayVar(&orgCreateTags, "tag", nil, "tag to attach (repeatable)")

	orgUseCmd.Flags().BoolVar(&orgUseClear, "clear", false, "clear the active org and show all orgs again")

	orgAssignCmd.Flags().StringArrayVarP(&orgAssignWorkspaces, "workspace", "w", nil, "workspace to assign (repeatable)")

	orgDeleteCmd.Flags().BoolVar(&orgDeletePurge, "purge", false, "also delete the org's workspaces, assets, vulnerabilities and runs")

	orgCmd.AddCommand(orgListCmd)
	orgCmd.AddCommand(orgCreateCmd)
	orgCmd.AddCommand(orgShowCmd)
	orgCmd.AddCommand(orgUseCmd)
	orgCmd.AddCommand(orgAssignCmd)
	orgCmd.AddCommand(orgRenameCmd)
	orgCmd.AddCommand(orgDeleteCmd)
}

// withOrgDB opens the database, migrates, and runs fn.
func withOrgDB(fn func(ctx context.Context) error) error {
	if err := connectDB(); err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	return fn(context.Background())
}

func runOrgList(cmd *cobra.Command, args []string) error {
	return withOrgDB(func(ctx context.Context) error {
		orgs, err := database.ListOrgs(ctx)
		if err != nil {
			return fmt.Errorf("failed to list orgs: %w", err)
		}

		type orgEntry struct {
			UUID        string   `json:"uuid"`
			Name        string   `json:"name"`
			Description string   `json:"description,omitempty"`
			Tags        []string `json:"tags,omitempty"`
			Workspaces  int      `json:"workspaces"`
			Assets      int      `json:"assets"`
			Vulns       int      `json:"vulns"`
			IsDefault   bool     `json:"is_default"`
			IsActive    bool     `json:"is_active"`
		}

		activeUUID, _ := resolveOrgUUID(ctx)

		entries := make([]orgEntry, 0, len(orgs))
		for i := range orgs {
			org := &orgs[i]
			stats, err := database.GetOrgStats(ctx, org.UUID)
			if err != nil {
				return fmt.Errorf("failed to get stats for org %s: %w", org.Name, err)
			}
			entries = append(entries, orgEntry{
				UUID:        org.UUID,
				Name:        org.Name,
				Description: org.Description,
				Tags:        org.Tags,
				Workspaces:  stats.TotalWorkspaces,
				Assets:      stats.TotalAssets,
				Vulns:       stats.TotalVulns,
				IsDefault:   org.IsDefault(),
				IsActive:    org.UUID == activeUUID,
			})
		}

		if globalJSON {
			return outputJSON(map[string]interface{}{"orgs": entries, "total": len(entries)})
		}

		printer := terminal.NewPrinter()
		printer.Newline()
		printer.Info("%s org(s)", terminal.BoldCyan(fmt.Sprintf("%d", len(entries))))
		printer.Newline()

		const nameWidth, wsWidth, assetWidth, vulnWidth = 22, 10, 10, 8
		fmt.Printf("  %s  %s  %s  %s  %s\n",
			terminal.Bold(padRight("NAME", nameWidth)),
			terminal.Bold(padRight("WORKSPACES", wsWidth)),
			terminal.Bold(padRight("ASSETS", assetWidth)),
			terminal.Bold(padRight("VULNS", vulnWidth)),
			terminal.Bold("UUID"))

		for _, e := range entries {
			name := e.Name
			if e.IsActive {
				name = "* " + name
			}
			fmt.Printf("  %s  %s  %s  %s  %s\n",
				colorPadRight(name, nameWidth, terminal.Cyan),
				padRight(fmt.Sprintf("%d", e.Workspaces), wsWidth),
				padRight(fmt.Sprintf("%d", e.Assets), assetWidth),
				padRight(fmt.Sprintf("%d", e.Vulns), vulnWidth),
				terminal.Gray(e.UUID))
		}

		binaryPath := filepath.Base(os.Args[0])
		printer.Newline()
		if activeUUID != "" {
			printer.Info("%s marks the active org. Clear it with %s",
				terminal.Cyan("*"), terminal.Gray(binaryPath+" org use --clear"))
		} else {
			printer.Info("No active org - commands show data across all orgs.")
		}
		printer.Info("Group workspaces: %s", terminal.Gray(binaryPath+" org assign <name> -w <workspace>"))
		return nil
	})
}

func runOrgCreate(cmd *cobra.Command, args []string) error {
	return withOrgDB(func(ctx context.Context) error {
		org, err := database.CreateOrg(ctx, args[0], orgCreateDescription, orgCreateUUID, orgCreateTags)
		if err != nil {
			return err
		}

		if globalJSON {
			return outputJSON(org)
		}

		printer := terminal.NewPrinter()
		printer.Success("Created org %s (%s)", terminal.BoldGreen(org.Name), terminal.Gray(org.UUID))

		binaryPath := filepath.Base(os.Args[0])
		printer.Info("Assign workspaces: %s",
			terminal.Gray(fmt.Sprintf("%s org assign %s -w <workspace>", binaryPath, org.Name)))
		return nil
	})
}

func runOrgShow(cmd *cobra.Command, args []string) error {
	return withOrgDB(func(ctx context.Context) error {
		org, err := database.ResolveOrgRef(ctx, args[0])
		if err != nil {
			return err
		}
		stats, err := database.GetOrgStats(ctx, org.UUID)
		if err != nil {
			return fmt.Errorf("failed to get org stats: %w", err)
		}

		if globalJSON {
			return outputJSON(map[string]interface{}{"org": org, "stats": stats})
		}

		printer := terminal.NewPrinter()
		printer.Newline()
		fmt.Printf("  %s %s\n", terminal.Bold("Name:"), terminal.BoldCyan(org.Name))
		fmt.Printf("  %s %s\n", terminal.Bold("UUID:"), terminal.Gray(org.UUID))
		if org.Description != "" {
			fmt.Printf("  %s %s\n", terminal.Bold("Description:"), org.Description)
		}
		if len(org.Tags) > 0 {
			fmt.Printf("  %s %s\n", terminal.Bold("Tags:"), strings.Join(org.Tags, ", "))
		}
		if org.IsDefault() {
			fmt.Printf("  %s %s\n", terminal.Bold("Default:"), terminal.Green("yes"))
		}
		fmt.Printf("  %s %s\n", terminal.Bold("Created:"), org.CreatedAt.Format("2006-01-02 15:04:05"))

		printer.Newline()
		fmt.Printf("  %s %d workspaces, %d assets, %d vulnerabilities, %d runs\n",
			terminal.Bold("Data:"),
			stats.TotalWorkspaces, stats.TotalAssets, stats.TotalVulns, stats.TotalRuns)

		if len(stats.Workspaces) > 0 {
			printer.Newline()
			fmt.Printf("  %s\n", terminal.Bold("Workspaces:"))
			sorted := append([]string(nil), stats.Workspaces...)
			sort.Strings(sorted)
			for _, ws := range sorted {
				fmt.Printf("    %s\n", terminal.Cyan(ws))
			}
		}

		binaryPath := filepath.Base(os.Args[0])
		printer.Newline()
		printer.Info("Query its assets: %s",
			terminal.Gray(fmt.Sprintf("%s assets --org %s", binaryPath, org.Name)))
		return nil
	})
}

func runOrgUse(cmd *cobra.Command, args []string) error {
	if orgUseClear {
		if err := clearActiveOrg(); err != nil {
			return fmt.Errorf("failed to clear active org: %w", err)
		}
		fmt.Println("unset OSMEDEUS_ORG_UUID")
		fmt.Fprintf(os.Stderr, "%s Active org cleared - commands now span all orgs.\n",
			terminal.StepSuccessSymbol())
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("requires an org name or UUID, or --clear")
	}

	return withOrgDB(func(ctx context.Context) error {
		org, err := database.ResolveOrgRef(ctx, args[0])
		if err != nil {
			return err
		}
		if err := writeActiveOrg(org.UUID); err != nil {
			return fmt.Errorf("failed to persist active org: %w", err)
		}

		// The export line goes to stdout so `eval $(...)` picks it up; everything
		// else goes to stderr so eval does not swallow it.
		fmt.Printf("export OSMEDEUS_ORG_UUID=%s\n", org.UUID)
		fmt.Fprintf(os.Stderr, "%s Active org: %s (%s)\n",
			terminal.StepSuccessSymbol(), terminal.BoldGreen(org.Name), org.UUID)
		fmt.Fprintf(os.Stderr, "  Saved to %s - picked up automatically by future commands.\n",
			terminal.Gray(activeOrgFilePath()))
		return nil
	})
}

func runOrgAssign(cmd *cobra.Command, args []string) error {
	if len(orgAssignWorkspaces) == 0 {
		return fmt.Errorf("no workspaces given: pass -w <workspace> at least once")
	}

	return withOrgDB(func(ctx context.Context) error {
		org, err := database.ResolveOrgRef(ctx, args[0])
		if err != nil {
			return err
		}

		counts, err := database.AssignWorkspacesToOrg(ctx, org.UUID, orgAssignWorkspaces)
		if err != nil {
			return err
		}

		if globalJSON {
			return outputJSON(map[string]interface{}{
				"org":        org,
				"workspaces": orgAssignWorkspaces,
				"updated":    counts,
			})
		}

		printer := terminal.NewPrinter()
		printer.Success("Assigned %d workspace(s) to %s",
			len(orgAssignWorkspaces), terminal.BoldGreen(org.Name))
		for _, table := range []string{"workspaces", "assets", "vulnerabilities", "runs"} {
			fmt.Printf("  %s %d rows\n", terminal.Bold(padRight(table+":", 18)), counts[table])
		}

		if counts["workspaces"] == 0 {
			printer.Warning("No workspace rows matched - check the names with %s",
				terminal.Gray(filepath.Base(os.Args[0])+" db -t workspaces"))
		}
		return nil
	})
}

func runOrgRename(cmd *cobra.Command, args []string) error {
	return withOrgDB(func(ctx context.Context) error {
		org, err := database.ResolveOrgRef(ctx, args[0])
		if err != nil {
			return err
		}

		oldName := org.Name
		updated, err := database.UpdateOrg(ctx, org.UUID, args[1], "", nil)
		if err != nil {
			return err
		}

		if globalJSON {
			return outputJSON(updated)
		}

		terminal.NewPrinter().Success("Renamed org %s to %s",
			terminal.Gray(oldName), terminal.BoldGreen(updated.Name))
		return nil
	})
}

func runOrgDelete(cmd *cobra.Command, args []string) error {
	return withOrgDB(func(ctx context.Context) error {
		org, err := database.ResolveOrgRef(ctx, args[0])
		if err != nil {
			return err
		}

		printer := terminal.NewPrinter()

		// Only the purge confirmation needs the counts, so the query stays out of
		// the default (reassign) path.
		if orgDeletePurge && !globalForce {
			stats, err := database.GetOrgStats(ctx, org.UUID)
			if err != nil {
				return fmt.Errorf("failed to get org stats: %w", err)
			}
			printer.Warning("--purge will permanently DELETE %d workspaces, %d assets, %d vulnerabilities and %d runs.",
				stats.TotalWorkspaces, stats.TotalAssets, stats.TotalVulns, stats.TotalRuns)
			if !confirmPrompt(fmt.Sprintf("Delete org %q and all its data?", org.Name)) {
				printer.Info("Aborted.")
				return nil
			}
		}

		if err := database.DeleteOrg(ctx, org.UUID, orgDeletePurge); err != nil {
			return err
		}

		// A deleted org must not stay selected.
		if active := readActiveOrg(); active == org.UUID {
			_ = clearActiveOrg()
		}

		if orgDeletePurge {
			printer.Success("Deleted org %s and its data", terminal.BoldGreen(org.Name))
		} else {
			printer.Success("Deleted org %s - its data moved to the %s org",
				terminal.BoldGreen(org.Name), terminal.Cyan(database.DefaultOrgName))
		}
		return nil
	})
}
