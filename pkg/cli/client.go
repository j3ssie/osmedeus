package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/j3ssie/osmedeus/v5/internal/client"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	// Client connection flags
	clientRemoteURL string
	clientAuthKey   string

	// Fetch subcommand flags
	clientFetchTable     string
	clientFetchWorkspace string
	clientFetchSeverity  string
	clientFetchStatus    string
	clientFetchLimit     int
	clientFetchOffset    int
	clientFetchRefresh   string

	// Run subcommand flags
	clientRunFlow   string
	clientRunModule string
	clientRunTarget string
	clientRunCancel string

	// Exec subcommand flags
	clientExecScript string
	clientExecTarget string
)

// clientCmd is the parent command for interacting with a remote osmedeus server
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Interact with a remote osmedeus server",
	Long:  UsageClient(),
}

// clientFetchCmd fetches data from the server
var clientFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch data from remote server (assets, runs, vulns, etc.)",
	Long:  UsageClientFetch(),
	RunE:  runClientFetch,
}

// clientRunCmd creates or cancels runs on the server
var clientRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Create or cancel a run on remote server",
	Long:  UsageClientRun(),
	RunE:  runClientRun,
}

// clientExecCmd executes a function remotely
var clientExecCmd = &cobra.Command{
	Use:   "exec [script]",
	Short: "Execute a function on remote server",
	Long:  UsageClientExec(),
	Args:  cobra.MaximumNArgs(1),
	RunE:  runClientExec,
}

func init() {
	// Add persistent flags for client connection (available to all subcommands)
	clientCmd.PersistentFlags().StringVar(&clientRemoteURL, "remote-url", "", "Remote server URL (default: $OSM_REMOTE_URL)")
	clientCmd.PersistentFlags().StringVar(&clientAuthKey, "auth-key", "", "API authentication key (default: $OSM_REMOTE_AUTH_KEY)")

	// Fetch subcommand flags
	clientFetchCmd.Flags().StringVarP(&clientFetchTable, "table", "t", "assets", "Table to fetch data from: runs, step_results, artifacts, assets, event_logs, schedules, workspaces, vulnerabilities, asset_diffs, vuln_diffs")
	clientFetchCmd.Flags().StringVarP(&clientFetchWorkspace, "workspace", "w", "", "Filter by workspace name")
	clientFetchCmd.Flags().StringVar(&clientFetchSeverity, "severity", "", "Filter by severity (for vulnerabilities only)")
	clientFetchCmd.Flags().StringVar(&clientFetchStatus, "status", "", "Filter by status (for runs only)")
	clientFetchCmd.Flags().IntVar(&clientFetchLimit, "limit", 20, "Maximum records to return")
	clientFetchCmd.Flags().IntVar(&clientFetchOffset, "offset", 0, "Number of records to skip (pagination)")
	clientFetchCmd.Flags().StringVar(&clientFetchRefresh, "refresh", "", "auto-refresh interval (e.g., 5s, 1m, 30s)")

	// Run subcommand flags
	clientRunCmd.Flags().StringVarP(&clientRunFlow, "flow", "f", "", "Flow workflow name")
	clientRunCmd.Flags().StringVarP(&clientRunModule, "module", "m", "", "Module workflow name")
	clientRunCmd.Flags().StringVarP(&clientRunTarget, "target", "T", "", "Target to run against")
	clientRunCmd.Flags().StringVar(&clientRunCancel, "cancel", "", "Run ID to cancel (switches to cancel mode)")

	// Exec subcommand flags
	clientExecCmd.Flags().StringVarP(&clientExecScript, "script", "s", "", "Script to execute (alternative to positional arg)")
	clientExecCmd.Flags().StringVarP(&clientExecTarget, "target", "t", "", "Target variable for script")

	// Build command hierarchy
	clientCmd.AddCommand(clientFetchCmd)
	clientCmd.AddCommand(clientRunCmd)
	clientCmd.AddCommand(clientExecCmd)
}

// createClient creates a new API client with the configured URL and auth key
func createClient() (*client.Client, error) {
	return client.NewClient(clientRemoteURL, clientAuthKey)
}

// outputJSON outputs data as JSON to stdout
func outputJSON(data interface{}) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// printClientTable prints data in a styled markdown table format
func printClientTable(tableName string, data interface{}, columns []string, pagination client.Pagination, recordCount int) {
	printer := terminal.NewPrinter()
	printer.Info("Table: %s", tableName)

	// Calculate pagination info
	startRecord := pagination.Offset + 1
	endRecord := pagination.Offset + recordCount
	if pagination.Total == 0 {
		startRecord = 0
	}
	fmt.Printf("Showing records %d-%d of %d\n\n", startRecord, endRecord, pagination.Total)

	// Format as markdown table
	tableStr := formatClientMarkdownTable(data, columns, globalWidth)

	// Render with glamour
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
}

// formatClientMarkdownTable formats records as a markdown table
func formatClientMarkdownTable(records interface{}, columns []string, maxWidth int) string {
	// Convert records to []map[string]interface{}
	jsonBytes, _ := json.Marshal(records)
	var data []map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		return "No records found."
	}

	if len(data) == 0 {
		return "No records found."
	}

	// Use specified columns or extract from first record
	var headers []string
	if len(columns) > 0 {
		headers = columns
	} else {
		for key := range data[0] {
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
			val := formatClientTableValue(row[h], maxWidth, h)
			if i > 0 {
				sb.WriteString(" | ")
			}
			sb.WriteString(val)
		}
		sb.WriteString(" |\n")
	}

	return sb.String()
}

// formatClientTableValue converts a value to string for markdown display
func formatClientTableValue(v interface{}, maxWidth int, columnName string) string {
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

	// Colorize status column
	if columnName == "status" {
		return terminal.ColorizeStatus(s)
	}
	return s
}

// validFetchTables lists all valid table names for the fetch command
var validFetchTables = []string{
	"runs", "step_results", "artifacts", "assets", "event_logs",
	"schedules", "workspaces", "vulnerabilities", "asset_diffs", "vuln_diffs",
}

// clientTableDefaultColumns defines default columns for each table in client fetch output
var clientTableDefaultColumns = map[string][]string{
	"runs":            {"run_uuid", "workflow_name", "target", "status", "completed_steps", "total_steps"},
	"step_results":    {"step_name", "step_type", "status", "duration_ms", "command"},
	"artifacts":       {"name", "artifact_path", "artifact_type", "size_bytes"},
	"assets":          {"asset_value", "host_ip", "title", "status_code"},
	"event_logs":      {"topic", "source", "processed", "workspace"},
	"schedules":       {"name", "workflow_name", "trigger_type", "schedule", "is_enabled"},
	"workspaces":      {"name", "data_source", "total_assets"},
	"vulnerabilities": {"vuln_title", "severity", "asset_value", "workspace"},
	"asset_diffs":     {"workspace_name", "from_time", "to_time", "total_added", "total_removed"},
	"vuln_diffs":      {"workspace_name", "from_time", "to_time", "total_added", "total_removed"},
}

// runClientFetchRefreshLoop continuously refreshes the fetch display at the specified interval
func runClientFetchRefreshLoop(cmd *cobra.Command, args []string, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	for {
		fmt.Print("\033[2J\033[H") // Clear screen
		if err := runClientFetchOnce(cmd, args); err != nil {
			terminal.NewPrinter().Error("Fetch failed: %s", err)
		}
		fmt.Printf("\n%s Refreshing every %s. Press Ctrl+C to stop.\n", terminal.Gray("⟳"), interval)

		select {
		case <-ticker.C:
			continue
		case <-sigChan:
			fmt.Print("\033[2J\033[H")
			terminal.NewPrinter().Info("Refresh stopped")
			return nil
		}
	}
}

// runClientFetch handles the fetch subcommand
func runClientFetch(cmd *cobra.Command, args []string) error {
	// Check if refresh mode is enabled
	if clientFetchRefresh != "" {
		interval, err := time.ParseDuration(clientFetchRefresh)
		if err != nil {
			return fmt.Errorf("invalid refresh interval: %w", err)
		}
		if interval < time.Second {
			return fmt.Errorf("refresh interval must be at least 1s")
		}
		return runClientFetchRefreshLoop(cmd, args, interval)
	}
	return runClientFetchOnce(cmd, args)
}

// runClientFetchOnce performs a single fetch and displays the results
func runClientFetchOnce(cmd *cobra.Command, args []string) error {
	c, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	query := url.Values{}
	query.Set("limit", strconv.Itoa(clientFetchLimit))
	query.Set("offset", strconv.Itoa(clientFetchOffset))

	var result interface{}
	var endpoint string

	// Validate table name
	validTable := false
	for _, t := range validFetchTables {
		if clientFetchTable == t {
			validTable = true
			break
		}
	}
	if !validTable {
		return fmt.Errorf("invalid table: %s (valid tables: %s)", clientFetchTable, strings.Join(validFetchTables, ", "))
	}

	switch clientFetchTable {
	case "runs":
		endpoint = "/runs"
		if clientFetchStatus != "" {
			query.Set("status", clientFetchStatus)
		}
		var resp client.RunsResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printRunsTable(resp)
			return nil
		}

	case "step_results":
		endpoint = "/step-results"
		if clientFetchWorkspace != "" {
			query.Set("workspace", clientFetchWorkspace)
		}
		if clientFetchStatus != "" {
			query.Set("status", clientFetchStatus)
		}
		var resp client.StepResultsResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printStepResultsTable(resp)
			return nil
		}

	case "artifacts":
		endpoint = "/artifacts"
		if clientFetchWorkspace != "" {
			query.Set("workspace", clientFetchWorkspace)
		}
		var resp client.ArtifactsResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printArtifactsTable(resp)
			return nil
		}

	case "assets":
		endpoint = "/assets"
		if clientFetchWorkspace != "" {
			query.Set("workspace", clientFetchWorkspace)
		}
		var resp client.AssetsResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printAssetsTable(resp)
			return nil
		}

	case "event_logs":
		endpoint = "/event-logs"
		if clientFetchWorkspace != "" {
			query.Set("workspace", clientFetchWorkspace)
		}
		var resp client.EventLogsResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printEventLogsTable(resp)
			return nil
		}

	case "schedules":
		endpoint = "/schedules"
		var resp client.SchedulesResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printSchedulesTable(resp)
			return nil
		}

	case "workspaces":
		endpoint = "/workspaces"
		var resp client.WorkspacesResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printWorkspacesTable(resp)
			return nil
		}

	case "vulnerabilities":
		endpoint = "/vulnerabilities"
		if clientFetchWorkspace != "" {
			query.Set("workspace", clientFetchWorkspace)
		}
		if clientFetchSeverity != "" {
			query.Set("severity", clientFetchSeverity)
		}
		var resp client.VulnerabilitiesResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printVulnerabilitiesTable(resp)
			return nil
		}

	case "asset_diffs":
		endpoint = "/assets/diffs"
		if clientFetchWorkspace != "" {
			query.Set("workspace", clientFetchWorkspace)
		}
		var resp client.AssetDiffsResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printAssetDiffsTable(resp)
			return nil
		}

	case "vuln_diffs":
		endpoint = "/vulnerabilities/diffs"
		if clientFetchWorkspace != "" {
			query.Set("workspace", clientFetchWorkspace)
		}
		var resp client.VulnDiffsResponse
		if err := c.Get(ctx, endpoint, query, &resp); err != nil {
			return err
		}
		result = resp
		if !globalJSON {
			printVulnDiffsTable(resp)
			return nil
		}
	}

	return outputJSON(result)
}

// runClientRun handles the run subcommand (create or cancel)
func runClientRun(cmd *cobra.Command, args []string) error {
	c, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	// Cancel mode: --cancel flag is set
	if clientRunCancel != "" {
		var resp client.CancelRunResponse
		if err := c.Delete(ctx, "/runs/"+clientRunCancel, &resp); err != nil {
			return err
		}

		if globalJSON {
			return outputJSON(resp)
		}

		printer := terminal.NewPrinter()
		printer.Success("Run cancelled: %s", resp.RunUUID)
		return nil
	}

	// Create mode: create a new run
	if clientRunFlow == "" && clientRunModule == "" {
		return fmt.Errorf("either --flow or --module is required (or use --cancel to cancel a run)")
	}
	if clientRunFlow != "" && clientRunModule != "" {
		return fmt.Errorf("only one of --flow or --module can be specified")
	}
	if clientRunTarget == "" {
		return fmt.Errorf("--target is required")
	}

	req := client.CreateRunRequest{
		Flow:   clientRunFlow,
		Module: clientRunModule,
		Target: clientRunTarget,
	}

	var resp client.CreateRunResponse
	if err := c.Post(ctx, "/runs", req, &resp); err != nil {
		return err
	}

	if globalJSON {
		return outputJSON(resp)
	}

	// Print human-readable output
	printer := terminal.NewPrinter()
	printer.Success("Run created successfully")
	fmt.Printf("  Workflow: %s (%s)\n", resp.Workflow, resp.Kind)
	fmt.Printf("  Target:   %s\n", resp.Target)
	fmt.Printf("  Job ID:   %s\n", resp.JobID)
	if resp.RunUUID != "" {
		fmt.Printf("  Run UUID: %s\n", resp.RunUUID)
	}
	fmt.Printf("  Status:   %s\n", resp.Status)
	fmt.Printf("  Poll URL: %s\n", resp.PollURL)

	return nil
}

// runClientExec handles the exec subcommand
func runClientExec(cmd *cobra.Command, args []string) error {
	// Get script from args or flag
	script := clientExecScript
	if len(args) > 0 {
		script = args[0]
	}
	if script == "" {
		return fmt.Errorf("script is required (as argument or via --script)")
	}

	c, err := createClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	req := client.FunctionEvalRequest{
		Script: script,
		Target: clientExecTarget,
	}

	var resp client.FunctionEvalResponse
	if err := c.Post(ctx, "/functions/eval", req, &resp); err != nil {
		return err
	}

	if globalJSON {
		return outputJSON(resp)
	}

	// Print human-readable output
	if resp.Error {
		return fmt.Errorf("execution failed: %s", resp.Message)
	}

	// Print result
	switch v := resp.Result.(type) {
	case string:
		fmt.Println(v)
	case nil:
		// No output
	default:
		output, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(output))
	}

	return nil
}

// printAssetsTable prints assets in a table format
func printAssetsTable(resp client.AssetsResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No assets found")
		return
	}
	printClientTable("assets", resp.Data, clientTableDefaultColumns["assets"], resp.Pagination, len(resp.Data))
}

// printWorkspacesTable prints workspaces in a table format
func printWorkspacesTable(resp client.WorkspacesResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No workspaces found")
		return
	}
	printClientTable("workspaces", resp.Data, clientTableDefaultColumns["workspaces"], resp.Pagination, len(resp.Data))
}

// printVulnerabilitiesTable prints vulnerabilities in a table format
func printVulnerabilitiesTable(resp client.VulnerabilitiesResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No vulnerabilities found")
		return
	}
	printClientTable("vulnerabilities", resp.Data, clientTableDefaultColumns["vulnerabilities"], resp.Pagination, len(resp.Data))
}

// printRunsTable prints runs in a table format
func printRunsTable(resp client.RunsResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No runs found")
		return
	}
	printClientTable("runs", resp.Data, clientTableDefaultColumns["runs"], resp.Pagination, len(resp.Data))
}

// printStepResultsTable prints step results in a table format
func printStepResultsTable(resp client.StepResultsResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No step results found")
		return
	}
	printClientTable("step_results", resp.Data, clientTableDefaultColumns["step_results"], resp.Pagination, len(resp.Data))
}

// printArtifactsTable prints artifacts in a table format
func printArtifactsTable(resp client.ArtifactsResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No artifacts found")
		return
	}
	printClientTable("artifacts", resp.Data, clientTableDefaultColumns["artifacts"], resp.Pagination, len(resp.Data))
}

// printEventLogsTable prints event logs in a table format
func printEventLogsTable(resp client.EventLogsResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No event logs found")
		return
	}
	printClientTable("event_logs", resp.Data, clientTableDefaultColumns["event_logs"], resp.Pagination, len(resp.Data))
}

// printSchedulesTable prints schedules in a table format
func printSchedulesTable(resp client.SchedulesResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No schedules found")
		return
	}
	printClientTable("schedules", resp.Data, clientTableDefaultColumns["schedules"], resp.Pagination, len(resp.Data))
}

// printAssetDiffsTable prints asset diffs in a table format
func printAssetDiffsTable(resp client.AssetDiffsResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No asset diffs found")
		return
	}
	printClientTable("asset_diffs", resp.Data, clientTableDefaultColumns["asset_diffs"], resp.Pagination, len(resp.Data))
}

// printVulnDiffsTable prints vulnerability diffs in a table format
func printVulnDiffsTable(resp client.VulnDiffsResponse) {
	if len(resp.Data) == 0 {
		terminal.NewPrinter().Info("No vulnerability diffs found")
		return
	}
	printClientTable("vuln_diffs", resp.Data, clientTableDefaultColumns["vuln_diffs"], resp.Pagination, len(resp.Data))
}
