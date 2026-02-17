package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	assetsWorkspace      string
	assetsSource         string
	assetsType           string
	assetsStats          bool
	assetsLimit          int
	assetsOffset         int
	assetsColumns        string
	assetsExcludeColumns string
	assetsAll            bool
)

var assetsCmd = &cobra.Command{
	Use:     "assets [search]",
	Aliases: []string{"asset"},
	Short:   "Query and list discovered assets",
	Long:    UsageAssets(),
	RunE:    runAssets,
}

func init() {
	assetsCmd.Flags().StringVarP(&assetsWorkspace, "workspace", "w", "", "filter by workspace name")
	assetsCmd.Flags().StringVar(&assetsSource, "source", "", "filter by source field (e.g., httpx, subfinder)")
	assetsCmd.Flags().StringVar(&assetsType, "type", "", "filter by asset_type field (e.g., web, subdomain)")
	assetsCmd.Flags().BoolVar(&assetsStats, "stats", false, "show asset statistics (unique technologies, sources, remarks, types)")
	assetsCmd.Flags().IntVar(&assetsLimit, "limit", 50, "maximum number of records to return")
	assetsCmd.Flags().IntVar(&assetsOffset, "offset", 0, "number of records to skip (for pagination)")
	assetsCmd.Flags().StringVar(&assetsColumns, "columns", "", "comma-separated columns to display")
	assetsCmd.Flags().StringVar(&assetsExcludeColumns, "exclude-columns", "", "comma-separated columns to exclude from output")
	assetsCmd.Flags().BoolVar(&assetsAll, "all", false, "show all columns including hidden ones (id, timestamps)")
}

func runAssets(cmd *cobra.Command, args []string) error {
	if disableDB {
		return fmt.Errorf("assets command unavailable: --disable-db flag is set")
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Connect to database
	_, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = database.Close() }()

	ctx := context.Background()
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if assetsStats {
		return runAssetsStats(ctx)
	}
	return runAssetsList(ctx, args)
}

func runAssetsList(ctx context.Context, args []string) error {
	// Validate pagination
	if assetsLimit <= 0 {
		assetsLimit = 50
	}
	if assetsLimit > 10000 {
		assetsLimit = 10000
	}
	if assetsOffset < 0 {
		assetsOffset = 0
	}

	// Build filters
	filters := make(map[string]string)
	if assetsSource != "" {
		filters["source"] = assetsSource
	}
	if assetsType != "" {
		filters["asset_type"] = assetsType
	}
	if assetsWorkspace != "" {
		filters["workspace"] = assetsWorkspace
	}

	// Search term from positional arg
	search := ""
	if len(args) > 0 {
		search = args[0]
	}

	records, err := database.GetTableRecords(ctx, "assets", assetsOffset, assetsLimit, filters, search, database.AssetHeavyColumns)
	if err != nil {
		return fmt.Errorf("failed to get assets: %w", err)
	}

	// JSON output
	if globalJSON {
		jsonBytes, err := json.Marshal(records.Records)
		if err != nil {
			return fmt.Errorf("failed to format records: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	// Table output
	printer := terminal.NewPrinter()

	requestedColumns := parseColumns(assetsColumns)
	columns := getEffectiveColumns("assets", requestedColumns, assetsAll)
	excludeColumns := parseExcludeColumns(assetsExcludeColumns)
	hideDefaultColumns := !assetsAll && len(requestedColumns) == 0 && tableDefaultColumns["assets"] == nil

	startRecord := records.Offset + 1
	endRecord := records.Offset + assetsLimit
	if endRecord > records.TotalCount {
		endRecord = records.TotalCount
	}
	if records.TotalCount == 0 {
		startRecord = 0
	}

	printer.Info("Assets")
	fmt.Printf("Showing records %d-%d of %d\n\n", startRecord, endRecord, records.TotalCount)

	renderTableWithTablewriter("assets", records.Records, columns, globalWidth, hideDefaultColumns, excludeColumns)

	if records.TotalCount > endRecord {
		nextOffset := records.Offset + records.Limit
		printer.Info("Next page: osmedeus assets --offset %d --limit %d", nextOffset, assetsLimit)
	}

	return nil
}

func runAssetsStats(ctx context.Context) error {
	stats, err := database.GetAssetStats(ctx, assetsWorkspace)
	if err != nil {
		return fmt.Errorf("failed to get asset stats: %w", err)
	}

	// JSON output
	if globalJSON {
		jsonBytes, err := json.Marshal(stats)
		if err != nil {
			return fmt.Errorf("failed to format stats: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	// Table output
	printer := terminal.NewPrinter()
	if assetsWorkspace != "" {
		printer.Info("Asset Statistics (workspace: %s)", assetsWorkspace)
	} else {
		printer.Info("Asset Statistics")
	}
	fmt.Println()

	printStatCategory("Technologies", stats.Technologies)
	printStatCategory("Sources", stats.Sources)
	printStatCategory("Remarks", stats.Remarks)
	printStatCategory("Asset Types", stats.AssetTypes)

	return nil
}

func printStatCategory(name string, items []string) {
	fmt.Printf("%s (%d):\n", terminal.Bold(name), len(items))
	if len(items) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, item := range items {
			fmt.Printf("  %s %s\n", terminal.SymbolBullet, strings.TrimSpace(item))
		}
	}
	fmt.Println()
}
