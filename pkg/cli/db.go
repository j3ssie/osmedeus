package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	dbTable          string
	dbOffset         int
	dbLimit          int
	dbNoTUI          bool
	dbWhere          []string
	dbColumns        string
	dbSearch         string
	dbAll            bool
	dbIndexForce     bool
	dbListColumns    bool
	dbExcludeColumns string
	dbRefresh        string
)

// defaultHiddenColumns are columns hidden by default for all tables
var defaultHiddenColumns = []string{"id", "created_at", "updated_at", "completed_at"}

// tableDefaultColumns defines default columns for specific tables
var tableDefaultColumns = map[string][]string{
	"runs":            {"run_uuid", "workflow_name", "target", "workspace", "trigger_type", "status", "completed_steps", "total_steps"},
	"step_results":    {"step_name", "step_type", "status", "duration_ms", "command"},
	"artifacts":       {"name", "path", "type", "size_bytes", "line_count"},
	"assets":          {"asset_value", "host_ip", "title", "status_code", "last_seen_at", "technologies"},
	"event_logs":      {"topic", "source", "processed", "data_type", "workspace", "data"},
	"schedules":       {"name", "workflow_name", "trigger_type", "schedule", "is_enabled", "run_count"},
	"workspaces":      {"name", "data_source", "total_assets", "total_ips", "total_vulns", "risk_score"},
	"vulnerabilities": {"vuln_title", "severity", "confidence", "asset_value", "last_seen_at", "workspace"},
}

// dbCmd - parent command for database management
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  UsageDB(),
	RunE:  runDBList,
}

// dbSeedCmd - seed database with sample data
var dbSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed database with sample data",
	Long:  UsageDBSeed(),
	RunE:  runDBSeed,
}

// dbCleanCmd - clean all data from database
var dbCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove all data from database",
	Long:  UsageDBClean(),
	RunE:  runDBClean,
}

// dbMigrateCmd - run database migrations
var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  UsageDBMigrate(),
	RunE:  runDBMigrate,
}

// dbListCmd - list database tables
var dbListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List database tables and row counts",
	Long:    UsageDBList(),
	RunE:    runDBList,
}

// dbIndexCmd - parent command for indexing resources
var dbIndexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index resources from filesystem to database",
	Long:  `Index resources from the filesystem into the database for faster querying.`,
}

// dbIndexWorkflowCmd - index workflows from filesystem
var dbIndexWorkflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Index workflows from filesystem to database",
	Long: `Scan the workflows directory and index all workflows into the database.
This enables faster workflow listing and filtering by tags, kind, etc.

The command will:
- Add new workflows found on disk
- Update workflows that have changed (checksum mismatch)
- Remove workflows from DB that no longer exist on disk

Use --force to re-index all workflows regardless of checksum.`,
	RunE: runIndexWorkflows,
}

func init() {
	// Note: --force flag is now global (defined in root.go)

	// Use persistent flags on dbCmd so they work on both `db` and `db ls`
	dbCmd.PersistentFlags().StringVarP(&dbTable, "table", "t", "", "table name to list records from (runs, step_results, artifacts, assets, event_logs, schedules, workspaces)")
	dbCmd.PersistentFlags().IntVar(&dbOffset, "offset", 0, "number of records to skip (for pagination)")
	dbCmd.PersistentFlags().IntVar(&dbLimit, "limit", 50, "maximum number of records to return")
	// Note: --json and --width flags are now global (defined in root.go)
	dbCmd.PersistentFlags().BoolVar(&dbNoTUI, "no-tui", false, "disable interactive TUI mode, use plain text output")
	dbCmd.PersistentFlags().StringArrayVar(&dbWhere, "where", nil, "filter records (key=value format, can be repeated) - only with --no-tui")
	dbCmd.PersistentFlags().StringVar(&dbColumns, "columns", "", "comma-separated columns to display (default: all) - only with --no-tui")
	dbCmd.PersistentFlags().StringVar(&dbSearch, "search", "", "search all columns for substring (case-insensitive) - only with --no-tui")
	dbCmd.PersistentFlags().BoolVar(&dbAll, "all", false, "show all columns including hidden ones (id, timestamps) - only with --no-tui")
	dbCmd.PersistentFlags().BoolVar(&dbListColumns, "list-columns", false, "list all available columns for the specified table")
	dbCmd.PersistentFlags().StringVar(&dbExcludeColumns, "exclude-columns", "", "comma-separated column names to exclude from output")
	dbCmd.PersistentFlags().StringVar(&dbRefresh, "refresh", "", "auto-refresh interval (e.g., 5s, 1m, 30s)")

	dbIndexWorkflowCmd.Flags().BoolVar(&dbIndexForce, "force", false, "force re-index all workflows regardless of checksum")

	dbIndexCmd.AddCommand(dbIndexWorkflowCmd)

	dbCmd.AddCommand(dbSeedCmd)
	dbCmd.AddCommand(dbCleanCmd)
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbListCmd)
	dbCmd.AddCommand(dbIndexCmd)
}

// runDBSeed seeds the database with sample data
func runDBSeed(cmd *cobra.Command, args []string) error {
	if disableDB {
		return fmt.Errorf("database commands unavailable: --disable-db flag is set")
	}

	printer := terminal.NewPrinter()
	cfg := config.Get()

	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	printer.Info("Connecting to database...")

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = database.Close() }()

	// Run migrations first to ensure tables exist
	ctx := context.Background()
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	printer.Info("Seeding database with sample data...")

	// Seed the database
	if err := database.SeedDatabase(ctx); err != nil {
		return fmt.Errorf("failed to seed database: %w", err)
	}

	printer.Success("Database seeded successfully")
	printer.Info("Database: %s", getDatabaseInfo(cfg, db))

	return nil
}

// runDBClean removes all data from the database
func runDBClean(cmd *cobra.Command, args []string) error {
	if disableDB {
		return fmt.Errorf("database commands unavailable: --disable-db flag is set")
	}

	printer := terminal.NewPrinter()
	cfg := config.Get()

	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	if !globalForce {
		printer.Warning("This will delete ALL data from the database!")
		printer.Warning("Use --force to skip this confirmation")
		return fmt.Errorf("operation aborted: use --force to confirm")
	}

	ctx := context.Background()

	if cfg.IsSQLite() {
		// SQLite: Delete the file and recreate
		dbPath := cfg.GetDBPath()

		// Close existing connection if any
		_ = database.Close()

		// Delete the database file
		if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete database file: %w", err)
		}
		printer.Info("Deleted database file: %s", dbPath)

		// Reconnect and migrate
		_, err := database.Connect(cfg)
		if err != nil {
			return fmt.Errorf("failed to reconnect to database: %w", err)
		}
		defer func() { _ = database.Close() }()

		if err := database.Migrate(ctx); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}
		printer.Success("Database recreated with fresh schema")
	} else {
		// PostgreSQL: Clean tables and run migrate
		printer.Info("Connecting to database...")

		db, err := database.Connect(cfg)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer func() { _ = database.Close() }()

		printer.Info("Cleaning database...")
		if err := database.CleanDatabase(ctx); err != nil {
			return fmt.Errorf("failed to clean database: %w", err)
		}

		printer.Info("Running migrations...")
		if err := database.Migrate(ctx); err != nil {
			return fmt.Errorf("failed to run migrations: %w", err)
		}

		printer.Success("Database cleaned and schema updated")
		printer.Info("Database: %s", getDatabaseInfo(cfg, db))
	}

	return nil
}

// runDBMigrate runs database migrations
func runDBMigrate(cmd *cobra.Command, args []string) error {
	if disableDB {
		return fmt.Errorf("database commands unavailable: --disable-db flag is set")
	}

	printer := terminal.NewPrinter()
	cfg := config.Get()

	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	printer.Info("Connecting to database...")

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = database.Close() }()

	ctx := context.Background()

	printer.Info("Running migrations...")

	// Run migrations
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	printer.Success("Database migrations completed")
	printer.Info("Database: %s", getDatabaseInfo(cfg, db))

	return nil
}

// runDBList lists all database tables with row counts or records from a specific table
func runDBList(cmd *cobra.Command, args []string) error {
	if disableDB {
		return fmt.Errorf("database commands unavailable: --disable-db flag is set")
	}

	printer := terminal.NewPrinter()
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

	// Ensure tables exist
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Handle --list-columns flag
	if dbListColumns {
		if dbTable == "" {
			return fmt.Errorf("--list-columns requires --table/-t flag")
		}
		columns := database.GetAllTableColumns(dbTable)
		if len(columns) == 0 {
			return fmt.Errorf("unknown table or no columns found: %s", dbTable)
		}
		printer.Info("Available columns for table '%s':", dbTable)
		for _, col := range columns {
			fmt.Printf("  %s\n", col)
		}
		return nil
	}

	// JSON mode bypasses TUI entirely
	if globalJSON {
		if dbTable != "" {
			return listTableRecordsJSON(ctx)
		}
		return listAllTablesJSON(ctx)
	}

	// No-TUI mode or specific table with flags: use plain text output
	if dbNoTUI || dbTable != "" {
		if dbTable != "" {
			return listTableRecords(ctx, cfg, printer)
		}
		return listAllTables(ctx, cfg, printer)
	}

	// Default: use interactive TUI
	return runDBListTUI(ctx)
}

// runDBListTUI starts the interactive database TUI
func runDBListTUI(ctx context.Context) error {
	// Get all tables
	tables, err := database.ListTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	// Convert to terminal.TableInfo
	tuiTables := make([]terminal.TableInfo, len(tables))
	for i, t := range tables {
		tuiTables[i] = terminal.TableInfo{
			Name:     t.Name,
			RowCount: t.RowCount,
		}
	}

	// Record fetcher wraps database.GetTableRecords
	recordFetcher := func(ctx context.Context, tableName string, offset, limit int, filters map[string]string, search string) (*terminal.TableRecords, error) {
		result, err := database.GetTableRecords(ctx, tableName, offset, limit, filters, search)
		if err != nil {
			return nil, err
		}
		return &terminal.TableRecords{
			Table:      result.Table,
			TotalCount: result.TotalCount,
			Offset:     result.Offset,
			Limit:      result.Limit,
			Records:    result.Records,
		}, nil
	}

	// Column fetcher wraps database.GetTableColumns (for display)
	columnFetcher := func(tableName string) []string {
		return database.GetTableColumns(tableName)
	}

	// All columns fetcher wraps database.GetAllTableColumns (for column selection)
	allColumnsFetcher := func(tableName string) []string {
		return database.GetAllTableColumns(tableName)
	}

	// Create and run TUI (pass dbLimit for page size)
	tui := terminal.NewDBTUI(tuiTables, recordFetcher, columnFetcher, allColumnsFetcher, dbLimit)
	return tui.Run()
}

// listAllTablesJSON outputs all tables as JSON
func listAllTablesJSON(ctx context.Context) error {
	tables, err := database.ListTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}
	jsonBytes, err := json.Marshal(tables)
	if err != nil {
		return fmt.Errorf("failed to format tables: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

// listTableRecordsJSON outputs table records as JSON only
func listTableRecordsJSON(ctx context.Context) error {
	// Validate limit
	if dbLimit <= 0 {
		dbLimit = 50
	}
	if dbLimit > 10000 {
		dbLimit = 10000
	}
	if dbOffset < 0 {
		dbOffset = 0
	}

	filters := parseWhereFilters(dbWhere)
	records, err := database.GetTableRecords(ctx, dbTable, dbOffset, dbLimit, filters, dbSearch)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	jsonBytes, err := json.Marshal(records.Records)
	if err != nil {
		return fmt.Errorf("failed to format records: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

// listAllTables lists all database tables with their row counts
func listAllTables(ctx context.Context, cfg *config.Config, printer *terminal.Printer) error {
	tables, err := database.ListTables(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	printer.Info("Database: %s", getDatabaseInfo(cfg, nil))
	fmt.Println()
	fmt.Printf("%-20s %s\n", "Table", "Rows")
	fmt.Println("─────────────────────────────")
	for _, t := range tables {
		fmt.Printf("%-20s %d\n", t.Name, t.RowCount)
	}
	fmt.Println()
	printer.Info("Use --table <name> to list records from a specific table")

	return nil
}

// runDBRefreshLoop continuously refreshes the table display at the specified interval
func runDBRefreshLoop(ctx context.Context, cfg *config.Config, printer *terminal.Printer, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	for {
		fmt.Print("\033[2J\033[H") // Clear screen
		if err := listTableRecordsOnce(ctx, cfg, printer); err != nil {
			printer.Error("Query failed: %s", err)
		}
		fmt.Printf("\n%s Refreshing every %s. Press Ctrl+C to stop.\n", terminal.Gray("⟳"), interval)

		select {
		case <-ticker.C:
			continue
		case <-sigChan:
			fmt.Print("\033[2J\033[H")
			printer.Info("Refresh stopped")
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// listTableRecords lists records from a specific table with pagination
func listTableRecords(ctx context.Context, cfg *config.Config, printer *terminal.Printer) error {
	// Check if refresh mode is enabled
	if dbRefresh != "" {
		interval, err := time.ParseDuration(dbRefresh)
		if err != nil {
			return fmt.Errorf("invalid refresh interval: %w", err)
		}
		if interval < time.Second {
			return fmt.Errorf("refresh interval must be at least 1s")
		}
		return runDBRefreshLoop(ctx, cfg, printer, interval)
	}
	return listTableRecordsOnce(ctx, cfg, printer)
}

// listTableRecordsOnce performs a single query and displays the results
func listTableRecordsOnce(ctx context.Context, cfg *config.Config, printer *terminal.Printer) error {
	// Validate limit
	if dbLimit <= 0 {
		dbLimit = 50
	}
	if dbLimit > 10000 {
		dbLimit = 10000
	}

	// Validate offset
	if dbOffset < 0 {
		dbOffset = 0
	}

	// Parse filters, columns, and exclude columns
	filters := parseWhereFilters(dbWhere)
	requestedColumns := parseColumns(dbColumns)
	columns := getEffectiveColumns(dbTable, requestedColumns, dbAll)
	excludeColumns := parseExcludeColumns(dbExcludeColumns)

	// Determine if we should hide default columns (when no specific columns requested and not --all)
	hideDefaultColumns := !dbAll && len(requestedColumns) == 0 && tableDefaultColumns[dbTable] == nil

	records, err := database.GetTableRecords(ctx, dbTable, dbOffset, dbLimit, filters, dbSearch)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	// JSON-only output mode
	if globalJSON {
		jsonBytes, err := json.Marshal(records.Records)
		if err != nil {
			return fmt.Errorf("failed to format records: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	// Calculate pagination info
	startRecord := records.Offset + 1
	endRecord := records.Offset + dbLimit
	if endRecord > records.TotalCount {
		endRecord = records.TotalCount
	}
	if records.TotalCount == 0 {
		startRecord = 0
	}

	printer.Info("Table: %s", records.Table)
	fmt.Printf("Showing records %d-%d of %d\n\n", startRecord, endRecord, records.TotalCount)

	// Output as markdown table
	tableStr := formatAsMarkdownTable(records.Records, columns, globalWidth, hideDefaultColumns, excludeColumns)

	// Render with glamour for styled markdown table
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
	)
	if err == nil {
		rendered, renderErr := renderer.Render(tableStr)
		if renderErr == nil {
			fmt.Print(rendered)
		} else {
			fmt.Println(tableStr)
		}
	} else {
		fmt.Println(tableStr)
	}

	// Show pagination hints
	if records.TotalCount > endRecord {
		nextOffset := records.Offset + records.Limit
		printer.Info("Next page: osmedeus db list -t %s --offset %d --limit %d", dbTable, nextOffset, dbLimit)
	}

	return nil
}

// parseWhereFilters parses --where flags into a map
func parseWhereFilters(whereFlags []string) map[string]string {
	filters := make(map[string]string)
	for _, w := range whereFlags {
		parts := strings.SplitN(w, "=", 2)
		if len(parts) == 2 {
			filters[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return filters
}

// parseColumns parses --columns flag into a slice
func parseColumns(columnsFlag string) []string {
	if columnsFlag == "" {
		return nil
	}
	cols := strings.Split(columnsFlag, ",")
	for i := range cols {
		cols[i] = strings.TrimSpace(cols[i])
	}
	return cols
}

// parseExcludeColumns parses --exclude-columns flag into a map for O(1) lookup
func parseExcludeColumns(excludeFlag string) map[string]bool {
	excludeMap := make(map[string]bool)
	if excludeFlag == "" {
		return excludeMap
	}
	cols := strings.Split(excludeFlag, ",")
	for _, col := range cols {
		excludeMap[strings.TrimSpace(col)] = true
	}
	return excludeMap
}

// getEffectiveColumns determines which columns to display based on flags and table defaults
func getEffectiveColumns(tableName string, requestedColumns []string, showAll bool) []string {
	// If user specified columns, use them as-is
	if len(requestedColumns) > 0 {
		return requestedColumns
	}

	// If --all flag, return nil (show all columns)
	if showAll {
		return nil
	}

	// Check for table-specific defaults
	if defaults, ok := tableDefaultColumns[tableName]; ok {
		return defaults
	}

	// Return nil to indicate "all except hidden"
	return nil
}

// isHiddenColumn checks if a column should be hidden by default
func isHiddenColumn(col string) bool {
	for _, hidden := range defaultHiddenColumns {
		if col == hidden {
			return true
		}
	}
	return false
}

// formatAsMarkdownTable formats records as a markdown table
func formatAsMarkdownTable(records interface{}, columns []string, maxWidth int, hideDefaultColumns bool, excludeColumns map[string]bool) string {
	// Convert records to []map[string]interface{}
	jsonBytes, _ := json.Marshal(records)
	var data []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return "No records found."
	}

	if len(data) == 0 {
		return "No records found."
	}

	// Get headers (all keys or selected columns)
	var headers []string
	if len(columns) > 0 {
		// Filter out excluded columns from specified columns
		for _, col := range columns {
			if !excludeColumns[col] {
				headers = append(headers, col)
			}
		}
	} else {
		for key := range data[0] {
			// Skip hidden columns if hideDefaultColumns is true
			if hideDefaultColumns && isHiddenColumn(key) {
				continue
			}
			// Skip excluded columns
			if excludeColumns[key] {
				continue
			}
			headers = append(headers, key)
		}
		sort.Strings(headers)
	}

	// Build markdown table
	var sb strings.Builder

	// Header row
	sb.WriteString("| ")
	sb.WriteString(strings.Join(headers, " | "))
	sb.WriteString(" |\n")

	// Separator row
	sb.WriteString("|")
	for range headers {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// Data rows
	for _, row := range data {
		sb.WriteString("| ")
		for i, h := range headers {
			val := formatTableValue(row[h], maxWidth)
			if i > 0 {
				sb.WriteString(" | ")
			}
			sb.WriteString(val)
		}
		sb.WriteString(" |\n")
	}

	return sb.String()
}

// formatTableValue converts a value to string for markdown display
func formatTableValue(v interface{}, maxWidth int) string {
	if v == nil {
		return ""
	}

	var s string
	switch val := v.(type) {
	case string:
		// Escape pipe characters and newlines
		s = strings.ReplaceAll(val, "|", "\\|")
		s = strings.ReplaceAll(s, "\n", " ")
	case map[string]interface{}, []interface{}:
		// Compact JSON for complex types
		b, _ := json.Marshal(val)
		s = string(b)
	default:
		s = fmt.Sprintf("%v", val)
	}

	// Apply width limit
	if maxWidth > 0 && len(s) > maxWidth {
		if maxWidth > 3 {
			s = s[:maxWidth-3] + "..."
		} else {
			s = s[:maxWidth]
		}
	}

	return s
}

// getDatabaseInfo returns a human-readable database info string
func getDatabaseInfo(cfg *config.Config, db interface{}) string {
	if cfg.IsPostgres() {
		return fmt.Sprintf("PostgreSQL @ %s:%d/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	}
	return fmt.Sprintf("SQLite @ %s", cfg.GetDBPath())
}

// runIndexWorkflows indexes workflows from filesystem to database
func runIndexWorkflows(cmd *cobra.Command, args []string) error {
	if disableDB {
		return fmt.Errorf("database commands unavailable: --disable-db flag is set")
	}

	printer := terminal.NewPrinter()
	cfg := config.Get()

	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	printer.Info("Connecting to database...")

	// Connect to database
	_, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = database.Close() }()

	ctx := context.Background()

	// Ensure tables exist
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	printer.Info("Indexing workflows from: %s", cfg.WorkflowsPath)
	if dbIndexForce {
		printer.Info("Force mode: re-indexing all workflows")
	}

	// Index workflows
	result, err := database.IndexWorkflowsFromFilesystem(ctx, cfg.WorkflowsPath, dbIndexForce)
	if err != nil {
		return fmt.Errorf("failed to index workflows: %w", err)
	}

	// Print results
	printer.Success("Workflow indexing completed")
	fmt.Println()
	fmt.Printf("  Added:   %d\n", result.Added)
	fmt.Printf("  Updated: %d\n", result.Updated)
	fmt.Printf("  Removed: %d\n", result.Removed)

	if len(result.Errors) > 0 {
		fmt.Println()
		printer.Warning("Errors encountered:")
		for _, e := range result.Errors {
			printer.Bullet(e)
		}
	}

	// Show total count
	count, err := database.GetWorkflowCount(ctx)
	if err == nil {
		fmt.Println()
		printer.Info("Total workflows indexed: %d", count)
	}

	return nil
}
