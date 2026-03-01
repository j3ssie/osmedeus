package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	queryWorkspace      string
	querySeverity       string
	queryConfidence     string
	queryAsset          string
	queryStatus         string
	queryWorkflow       string
	queryTarget         string
	queryRunUUID        string
	queryLimit          int
	queryOffset         int
	queryColumns        string
	queryExcludeColumns string
	queryAll            bool
	queryWhere          []string
	querySearch         string
)

// queryCmd - parent command for agent-friendly queries
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query scan data (agent-friendly)",
	Long:  UsageQuery(),
}

// queryVulnsCmd - query vulnerabilities
var queryVulnsCmd = &cobra.Command{
	Use:   "vulns",
	Short: "Query vulnerabilities",
	Long:  UsageQueryVulns(),
	RunE:  runQueryVulns,
}

// queryRunsCmd - query runs
var queryRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Query workflow runs",
	Long:  UsageQueryRuns(),
	RunE:  runQueryRuns,
}

// queryStepsCmd - query steps for a run
var queryStepsCmd = &cobra.Command{
	Use:   "steps",
	Short: "Query steps for a specific run",
	Long:  UsageQuerySteps(),
	RunE:  runQuerySteps,
}

func init() {
	// Shared flags on parent (inherited by subcommands)
	queryCmd.PersistentFlags().IntVar(&queryLimit, "limit", 50, "maximum number of records to return")
	queryCmd.PersistentFlags().IntVar(&queryOffset, "offset", 0, "number of records to skip (for pagination)")
	queryCmd.PersistentFlags().StringVar(&queryColumns, "columns", "", "comma-separated columns to display")
	queryCmd.PersistentFlags().StringVar(&queryExcludeColumns, "exclude-columns", "", "comma-separated columns to exclude from output")
	queryCmd.PersistentFlags().BoolVar(&queryAll, "all", false, "show all columns including hidden ones (id, timestamps)")
	queryCmd.PersistentFlags().StringArrayVar(&queryWhere, "where", nil, "filter records by column (key=value format, can be repeated)")
	queryCmd.PersistentFlags().StringVar(&querySearch, "search", "", "search all columns for substring (case-insensitive)")

	// Vulns-specific flags
	queryVulnsCmd.Flags().StringVarP(&queryWorkspace, "workspace", "w", "", "filter by workspace name")
	queryVulnsCmd.Flags().StringVar(&querySeverity, "severity", "", "filter by severity (critical, high, medium, low, info)")
	queryVulnsCmd.Flags().StringVar(&queryConfidence, "confidence", "", "filter by confidence (confirmed, firm, tentative)")
	queryVulnsCmd.Flags().StringVar(&queryAsset, "asset", "", "filter by asset value (substring match)")

	// Runs-specific flags
	queryRunsCmd.Flags().StringVarP(&queryWorkspace, "workspace", "w", "", "filter by workspace name")
	queryRunsCmd.Flags().StringVar(&queryStatus, "status", "", "filter by status (pending, running, completed, failed, cancelled)")
	queryRunsCmd.Flags().StringVar(&queryWorkflow, "workflow", "", "filter by workflow name")
	queryRunsCmd.Flags().StringVar(&queryTarget, "target", "", "filter by target (substring match)")

	// Steps-specific flags
	queryStepsCmd.Flags().StringVarP(&queryRunUUID, "run", "r", "", "run UUID (required)")
	_ = queryStepsCmd.MarkFlagRequired("run")

	queryCmd.AddCommand(queryVulnsCmd)
	queryCmd.AddCommand(queryRunsCmd)
	queryCmd.AddCommand(queryStepsCmd)
}

// connectDB is a helper that checks disableDB, loads config, connects, and migrates.
func connectDB() error {
	if disableDB {
		return fmt.Errorf("query command unavailable: --disable-db flag is set")
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	_, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	ctx := context.Background()
	if err := database.Migrate(ctx); err != nil {
		_ = database.Close()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// validatePagination clamps limit and offset to safe ranges.
func validatePagination(limit, offset *int) {
	if *limit <= 0 {
		*limit = 50
	}
	if *limit > 10000 {
		*limit = 10000
	}
	if *offset < 0 {
		*offset = 0
	}
}

// buildQueryFilters merges typed convenience flags into the --where filter map.
// Typed flags take precedence; --where can add any additional column filters.
func buildQueryFilters(typed map[string]string) map[string]string {
	filters := parseWhereFilters(queryWhere)
	for k, v := range typed {
		if v != "" {
			filters[k] = v
		}
	}
	return filters
}

// runQueryTable is the shared implementation for all query subcommands.
// It calls GetTableRecords with the merged filters and renders the result.
func runQueryTable(tableName string, filters map[string]string) error {
	if err := connectDB(); err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	validatePagination(&queryLimit, &queryOffset)

	ctx := context.Background()
	records, err := database.GetTableRecords(ctx, tableName, queryOffset, queryLimit, filters, nil, querySearch, nil)
	if err != nil {
		return fmt.Errorf("failed to query %s: %w", tableName, err)
	}

	if globalJSON {
		jsonBytes, err := json.Marshal(records.Records)
		if err != nil {
			return fmt.Errorf("failed to format results: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	// Table output
	printer := terminal.NewPrinter()
	requestedColumns := parseColumns(queryColumns)
	columns := getEffectiveColumns(tableName, requestedColumns, queryAll)
	excludeColumns := parseExcludeColumns(queryExcludeColumns)
	hideDefaultColumns := !queryAll && len(requestedColumns) == 0 && tableDefaultColumns[tableName] == nil

	startRecord := records.Offset + 1
	endRecord := records.Offset + queryLimit
	if endRecord > records.TotalCount {
		endRecord = records.TotalCount
	}
	if records.TotalCount == 0 {
		startRecord = 0
	}

	printer.Info("%s", tableName)
	fmt.Printf("Showing records %d-%d of %d\n\n", startRecord, endRecord, records.TotalCount)

	renderTableWithTablewriter(tableName, records.Records, columns, globalWidth, hideDefaultColumns, excludeColumns)

	if records.TotalCount > endRecord {
		nextOffset := records.Offset + records.Limit
		printer.Info("Next page: add --offset %d --limit %d", nextOffset, queryLimit)
	}

	return nil
}

func runQueryVulns(cmd *cobra.Command, args []string) error {
	filters := buildQueryFilters(map[string]string{
		"workspace":   queryWorkspace,
		"severity":    querySeverity,
		"confidence":  queryConfidence,
		"asset_value": queryAsset,
	})
	return runQueryTable("vulnerabilities", filters)
}

func runQueryRuns(cmd *cobra.Command, args []string) error {
	filters := buildQueryFilters(map[string]string{
		"workspace":     queryWorkspace,
		"status":        queryStatus,
		"workflow_name": queryWorkflow,
		"target":        queryTarget,
	})
	return runQueryTable("runs", filters)
}

func runQuerySteps(cmd *cobra.Command, args []string) error {
	if err := connectDB(); err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	// Resolve run UUID to numeric run_id for step_results table
	ctx := context.Background()
	run, err := database.GetRunByID(ctx, queryRunUUID, false, false)
	if err != nil {
		return fmt.Errorf("run not found: %w", err)
	}

	filters := buildQueryFilters(map[string]string{
		"run_id": fmt.Sprintf("%d", run.ID),
	})

	validatePagination(&queryLimit, &queryOffset)

	records, err := database.GetTableRecords(ctx, "step_results", queryOffset, queryLimit, filters, nil, querySearch, nil)
	if err != nil {
		return fmt.Errorf("failed to query steps: %w", err)
	}

	if globalJSON {
		jsonBytes, err := json.Marshal(records.Records)
		if err != nil {
			return fmt.Errorf("failed to format results: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	// Table output
	printer := terminal.NewPrinter()
	requestedColumns := parseColumns(queryColumns)
	columns := getEffectiveColumns("step_results", requestedColumns, queryAll)
	excludeColumns := parseExcludeColumns(queryExcludeColumns)
	hideDefaultColumns := !queryAll && len(requestedColumns) == 0 && tableDefaultColumns["step_results"] == nil

	printer.Info("Steps for run: %s", queryRunUUID)
	fmt.Printf("Total: %d steps\n\n", records.TotalCount)

	renderTableWithTablewriter("step_results", records.Records, columns, globalWidth, hideDefaultColumns, excludeColumns)

	return nil
}
