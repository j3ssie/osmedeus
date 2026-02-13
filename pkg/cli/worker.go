package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var printer = terminal.NewPrinter()

var redisURL string

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
		worker, err := distributed.NewWorker(cfg)
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
	Use:   "status",
	Short: "Show worker pool status",
	Long:  UsageWorkerStatus(),
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

		if len(workers) == 0 {
			printer.Info("No workers connected")
			return nil
		}

		// Print workers
		printer.Section("Connected Workers")
		headers := []string{"ID", "Hostname", "Status", "Tasks Done", "Tasks Failed", "Last Heartbeat"}
		var rows [][]string

		for _, w := range workers {
			rows = append(rows, []string{
				w.ID,
				w.Hostname,
				w.Status,
				formatInt(w.TasksComplete),
				formatInt(w.TasksFailed),
				formatHeartbeat(w.LastHeartbeat),
			})
		}

		printMarkdownTable(headers, rows)
		return nil
	},
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
	workerStatusCmd.Flags().StringVar(&redisURL, "redis-url", "", "Redis connection URL (overrides settings)")

	workerCmd.AddCommand(workerJoinCmd)
	workerCmd.AddCommand(workerStatusCmd)
}

// formatInt formats an integer for display
func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
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
