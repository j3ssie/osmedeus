package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var printer = terminal.NewPrinter()

var redisURL string
var getPublicIP bool

// Worker status display flags
var (
	workerColumns        string
	workerExcludeColumns string
	workerSearch         string
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker node commands for distributed scanning",
	Long:  UsageWorker(),
}

// workerJoinCmd joins the worker pool
var workerJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join as a worker node",
	Long:  UsageWorkerJoin(),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return errConfigNotLoaded
		}

		// Override Redis config from URL if provided
		if redisURL != "" {
			redisCfg, err := distributed.ParseRedisURL(redisURL)
			if err != nil {
				return err
			}
			cfg.Redis = *redisCfg
		}

		// Check Redis is configured
		if !cfg.IsRedisConfigured() {
			return errRedisNotConfigured
		}

		// Ensure external-binaries are in PATH so workflow steps can find tools
		ensureExternalBinariesInPath(cfg)

		// Create worker
		worker, err := distributed.NewWorker(cfg, &distributed.WorkerOptions{
			GetPublicIP: getPublicIP,
		})
		if err != nil {
			return err
		}

		// Setup graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			<-sigCh
			cancel()
		}()

		// Run worker
		return worker.Run(ctx)
	},
}

// workerStatusCmd shows worker status
var workerStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"ls"},
	Short:   "Show worker pool status",
	Long:    UsageWorkerStatus(),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return errConfigNotLoaded
		}

		// Override Redis config from URL if provided
		if redisURL != "" {
			redisCfg, err := distributed.ParseRedisURL(redisURL)
			if err != nil {
				return err
			}
			cfg.Redis = *redisCfg
		}

		// Check Redis is configured
		if !cfg.IsRedisConfigured() {
			return errRedisNotConfigured
		}

		// Create master client to query workers
		master, err := distributed.NewMaster(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		workers, err := master.ListWorkers(ctx)
		if err != nil {
			return err
		}

		// JSON output mode
		if globalJSON {
			if len(workers) == 0 {
				fmt.Println("[]")
				return nil
			}
			jsonBytes, err := json.MarshalIndent(workers, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal workers: %w", err)
			}
			fmt.Println(string(jsonBytes))
			return nil
		}

		if len(workers) == 0 {
			printer.Info("No workers connected")
			return nil
		}

		// Print workers
		printer.Section("Connected Workers")

		allHeaders := []string{"ID", "Alias", "Hostname", "IP Address", "Public IP", "SSH", "Status", "Tasks Done", "Tasks Failed", "Last Heartbeat"}
		defaultHeaders := []string{"ID", "Hostname", "Public IP", "SSH", "Status", "Tasks Done", "Tasks Failed", "Last Heartbeat"}

		// Build all rows (raw values for search, colored for display)
		type workerRow struct {
			raw     []string // for search matching
			display []string // for display (with colors)
		}
		var allRows []workerRow
		for _, w := range workers {
			sshRaw := "no"
			sshDisplay := "no"
			if w.SSHEnabled {
				sshRaw = "yes"
				sshDisplay = terminal.Green("yes")
			}
			allRows = append(allRows, workerRow{
				raw: []string{w.ID, w.Alias, w.Hostname, w.IPAddress, w.PublicIP, sshRaw, w.Status,
					formatInt(w.TasksComplete), formatInt(w.TasksFailed), formatHeartbeat(w.LastHeartbeat)},
				display: []string{w.ID, w.Alias, w.Hostname, w.IPAddress, w.PublicIP, sshDisplay, colorizeWorkerStatus(w.Status),
					formatInt(w.TasksComplete), formatInt(w.TasksFailed), formatHeartbeat(w.LastHeartbeat)},
			})
		}

		// Search filter (on raw values, before column selection)
		if workerSearch != "" {
			searchLower := strings.ToLower(workerSearch)
			var filtered []workerRow
			for _, row := range allRows {
				match := false
				for _, cell := range row.raw {
					if strings.Contains(strings.ToLower(cell), searchLower) {
						match = true
						break
					}
				}
				if match {
					filtered = append(filtered, row)
				}
			}
			allRows = filtered
		}

		if len(allRows) == 0 {
			printer.Info("No workers matching search filter")
			return nil
		}

		// Determine active headers
		headers := resolveWorkerColumns(allHeaders, defaultHeaders, workerColumns, workerExcludeColumns)

		// Map headers to column indices and project rows
		colIndices := mapHeaderIndices(allHeaders, headers)
		var rows [][]string
		for _, row := range allRows {
			projected := make([]string, len(colIndices))
			for i, idx := range colIndices {
				projected[i] = row.display[idx]
			}
			rows = append(rows, projected)
		}

		printMarkdownTable(headers, rows)
		return nil
	},
}

// workerEvalCmd evaluates a function expression with distributed hooks
var workerEvalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Evaluate a function expression with distributed hooks registered",
	Long:  UsageWorkerEval(),
	RunE:  runWorkerEval,
}

func runWorkerEval(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return errConfigNotLoaded
	}

	// Override Redis config from URL if provided
	if redisURL != "" {
		redisCfg, err := distributed.ParseRedisURL(redisURL)
		if err != nil {
			return err
		}
		cfg.Redis = *redisCfg
	}

	// Check Redis is configured
	if !cfg.IsRedisConfigured() {
		return errRedisNotConfigured
	}

	// Create lightweight Redis client (no full Worker)
	client, err := distributed.NewClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create redis client: %w", err)
	}
	defer client.Close()

	// Ping Redis to verify connection
	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Generate a worker ID for hook registration
	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("%s-eval-%s", hostname, uuid.NewString()[:8])

	// Set worker mode BEFORE hook registration (required for ShouldUseRedisDataQueues)
	config.SetWorkerMode(true, workerID)

	// Register distributed hooks and defer cleanup
	cleanup := distributed.RegisterDistributedHooksFromClient(client, workerID)
	defer cleanup()

	// Verify distributed hooks are active
	if !config.ShouldUseRedisDataQueues() {
		printer.Error("Distributed hooks not active (Redis=%v, WorkerMode=%v)",
			cfg.IsDistributedMode(), config.IsWorkerMode())
		return fmt.Errorf("distributed hooks failed to activate")
	}

	// Setup lazy DB for db_* functions
	if !disableDB {
		database.SetLazyConfig(cfg)
	}

	// Resolve script from args/flags/stdin
	script, err := resolveEvalScript(args)
	if err != nil {
		printer.Error("%s", err)
		return err
	}

	printer.Info("Evaluating with distributed hooks (worker: %s)", terminal.Cyan(workerID))

	// Execute the script
	return executeFunctionForTarget(printer, script, evalTarget)
}

// workerSetCmd updates a worker field
var workerSetCmd = &cobra.Command{
	Use:   "set <worker-id-or-alias> <field> <value>",
	Short: "Update a worker field",
	Long:  UsageWorkerSet(),
	Args:  cobra.ExactArgs(3),
	RunE:  runWorkerSet,
}

func runWorkerSet(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return errConfigNotLoaded
	}

	// Override Redis config from URL if provided
	if redisURL != "" {
		redisCfg, err := distributed.ParseRedisURL(redisURL)
		if err != nil {
			return err
		}
		cfg.Redis = *redisCfg
	}

	if !cfg.IsRedisConfigured() {
		return errRedisNotConfigured
	}

	client, err := distributed.NewClientFromConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create redis client: %w", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	identifier := args[0]
	field := strings.ToLower(args[1])
	value := args[2]

	// Resolve worker: try by ID first, then by alias
	worker, err := client.GetWorker(ctx, identifier)
	if err != nil {
		return fmt.Errorf("failed to get worker: %w", err)
	}
	if worker == nil {
		worker, err = client.GetWorkerByAlias(ctx, identifier)
		if err != nil {
			return fmt.Errorf("failed to get worker by alias: %w", err)
		}
	}
	if worker == nil {
		return fmt.Errorf("worker not found: %s", identifier)
	}

	// Normalize field name variants
	field = strings.ReplaceAll(field, "-", "")
	field = strings.ReplaceAll(field, "_", "")

	switch field {
	case "alias":
		worker.Alias = value
	case "publicip":
		worker.PublicIP = value
	case "sshenabled":
		worker.SSHEnabled = value == "true" || value == "1"
	case "sshkeyspath":
		worker.SSHKeysPath = value
	default:
		return fmt.Errorf("unknown field %q (valid: alias, public-ip, ssh-enabled, ssh-keys-path)", args[1])
	}

	if err := client.RegisterWorker(ctx, worker); err != nil {
		return fmt.Errorf("failed to update worker: %w", err)
	}

	printer.Success("Updated %s=%s for worker %s", args[1], value, worker.ID)
	return nil
}

var errConfigNotLoaded = &exitError{message: "configuration not loaded", code: 1}
var errRedisNotConfigured = &exitError{message: "redis not configured. Add redis section to osm-settings.yaml or use --redis-url", code: 1}

type exitError struct {
	message string
	code    int
}

func (e *exitError) Error() string {
	return e.message
}

func init() {
	workerJoinCmd.Flags().StringVar(&redisURL, "redis-url", "", "Redis connection URL (overrides settings)")
	workerJoinCmd.Flags().BoolVar(&getPublicIP, "get-public-ip", false, "Fetch public IP from ipinfo.io on join")
	workerStatusCmd.Flags().StringVar(&redisURL, "redis-url", "", "Redis connection URL (overrides settings)")
	workerStatusCmd.Flags().StringVar(&workerColumns, "columns", "", "comma-separated columns to display (e.g., 'ID,Hostname,Status')")
	workerStatusCmd.Flags().StringVar(&workerExcludeColumns, "exclude-columns", "", "comma-separated columns to exclude")
	workerStatusCmd.Flags().StringVarP(&workerSearch, "search", "s", "", "filter workers by substring match (case-insensitive, searches all fields)")

	workerEvalCmd.Flags().StringVar(&redisURL, "redis-url", "", "Redis connection URL (overrides settings)")
	workerEvalCmd.Flags().StringVarP(&evalScript, "eval", "e", "", "script to evaluate")
	workerEvalCmd.Flags().StringVarP(&evalTarget, "target", "t", "", "target value for {{target}} variable")
	workerEvalCmd.Flags().StringArrayVar(&evalParams, "params", nil, "additional parameters (key=value format)")
	workerEvalCmd.Flags().BoolVar(&evalStdin, "stdin", false, "read script from stdin")

	workerSetCmd.Flags().StringVar(&redisURL, "redis-url", "", "Redis connection URL (overrides settings)")

	workerCmd.AddCommand(workerJoinCmd)
	workerCmd.AddCommand(workerStatusCmd)
	workerCmd.AddCommand(workerEvalCmd)
	workerCmd.AddCommand(workerSetCmd)
}

// formatInt formats an integer for display
func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}

// colorizeWorkerStatus applies color to a worker status string
func colorizeWorkerStatus(status string) string {
	switch status {
	case "idle":
		return terminal.Green(status)
	case "busy":
		return terminal.Yellow(status)
	case "offline":
		return terminal.Red(status)
	default:
		return terminal.Gray(status)
	}
}

// resolveWorkerColumns determines which columns to display based on flags
func resolveWorkerColumns(allHeaders, defaultHeaders []string, columnsFlag, excludeFlag string) []string {
	if columnsFlag != "" {
		// Use only the specified columns
		cols := parseColumns(columnsFlag)
		// Validate against allHeaders
		headerSet := make(map[string]bool)
		for _, h := range allHeaders {
			headerSet[strings.ToLower(h)] = true
		}
		var valid []string
		for _, col := range cols {
			for _, h := range allHeaders {
				if strings.EqualFold(col, h) {
					valid = append(valid, h)
					break
				}
			}
		}
		if len(valid) > 0 {
			return valid
		}
		return defaultHeaders
	}

	if excludeFlag != "" {
		excludeMap := parseExcludeColumns(excludeFlag)
		var result []string
		for _, h := range allHeaders {
			if !excludeMap[h] {
				result = append(result, h)
			}
		}
		return result
	}

	return defaultHeaders
}

// mapHeaderIndices returns the indices of selected headers within allHeaders
func mapHeaderIndices(allHeaders, selected []string) []int {
	indexMap := make(map[string]int)
	for i, h := range allHeaders {
		indexMap[h] = i
	}
	var indices []int
	for _, h := range selected {
		if idx, ok := indexMap[h]; ok {
			indices = append(indices, idx)
		}
	}
	return indices
}

// formatHeartbeat formats a time as a relative duration
func formatHeartbeat(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}
