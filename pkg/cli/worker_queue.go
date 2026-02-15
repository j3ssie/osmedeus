package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/database/repository"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/json"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// QueuedTask represents a task picked from either Redis or DB
type QueuedTask struct {
	Source        string // "redis" or "db"
	RunUUID       string
	RunID         int64
	WorkflowName  string
	WorkflowKind  string
	Target        string
	Params        map[string]interface{}
	InputIsFile   bool
	InputFilePath string
}

// QueuePollerConfig holds configuration for a QueuePoller instance.
type QueuePollerConfig struct {
	DBPollInterval    time.Duration // e.g. 5s for worker queue run, 30s for server
	RedisBRPOPTimeout time.Duration // default 5s
	Concurrency       int           // number of parallel executors
	RedisURL          string        // optional override (applied before Start)
}

// QueuePoller polls DB (and optionally Redis) for queued runs and executes them.
type QueuePoller struct {
	cfg       *config.Config
	pollerCfg QueuePollerConfig
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.Mutex
	running   bool
}

// NewQueuePoller creates a new QueuePoller. Call Start to begin processing.
func NewQueuePoller(cfg *config.Config, pollerCfg QueuePollerConfig) *QueuePoller {
	if pollerCfg.DBPollInterval == 0 {
		pollerCfg.DBPollInterval = 5 * time.Second
	}
	if pollerCfg.RedisBRPOPTimeout == 0 {
		pollerCfg.RedisBRPOPTimeout = 5 * time.Second
	}
	if pollerCfg.Concurrency <= 0 {
		pollerCfg.Concurrency = 1
	}
	return &QueuePoller{cfg: cfg, pollerCfg: pollerCfg}
}

// Start connects to the database, starts pollers and executor goroutines.
// It returns immediately after launching goroutines. Use Stop to shut down.
func (qp *QueuePoller) Start(ctx context.Context) error {
	qp.mu.Lock()
	defer qp.mu.Unlock()
	if qp.running {
		return fmt.Errorf("queue poller already running")
	}

	log := logger.Get()

	// Apply Redis URL override
	if qp.pollerCfg.RedisURL != "" {
		redisCfg, err := distributed.ParseRedisURL(qp.pollerCfg.RedisURL)
		if err != nil {
			return fmt.Errorf("invalid redis URL: %w", err)
		}
		qp.cfg.Redis = *redisCfg
	}

	// Connect to database and migrate
	dbConn, err := database.Connect(qp.cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	innerCtx, cancel := context.WithCancel(ctx)
	qp.cancel = cancel
	qp.running = true

	taskCh := make(chan QueuedTask, qp.pollerCfg.Concurrency*2)

	// Deduplication
	var seenMu sync.Mutex
	seen := make(map[string]bool)
	markSeen := func(runUUID string) bool {
		seenMu.Lock()
		defer seenMu.Unlock()
		if seen[runUUID] {
			return false
		}
		seen[runUUID] = true
		return true
	}

	runRepo := repository.NewRunRepository(dbConn)
	p := terminal.NewPrinter()

	// DB poller goroutine
	qp.wg.Add(1)
	go func() {
		defer qp.wg.Done()
		ticker := time.NewTicker(qp.pollerCfg.DBPollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-innerCtx.Done():
				return
			case <-ticker.C:
				runs, err := runRepo.ListQueued(innerCtx)
				if err != nil {
					if innerCtx.Err() != nil {
						return
					}
					log.Debug("DB poller error", zap.Error(err))
					continue
				}

				for _, run := range runs {
					if !markSeen(run.RunUUID) {
						continue
					}

					claimed, err := runRepo.ClaimQueuedRun(innerCtx, run.ID)
					if err != nil || !claimed {
						continue
					}

					taskCh <- QueuedTask{
						Source:        "db",
						RunUUID:       run.RunUUID,
						RunID:         run.ID,
						WorkflowName:  run.WorkflowName,
						WorkflowKind:  run.WorkflowKind,
						Target:        run.Target,
						Params:        run.Params,
						InputIsFile:   run.InputIsFile,
						InputFilePath: run.InputFilePath,
					}
				}
			}
		}
	}()

	// Redis poller goroutine (if configured)
	if qp.cfg.IsRedisConfigured() {
		qp.wg.Add(1)
		go func() {
			defer qp.wg.Done()
			client, err := distributed.NewClientFromConfig(qp.cfg)
			if err != nil {
				log.Debug("Redis poller: failed to connect", zap.Error(err))
				return
			}
			defer client.Close()

			if err := client.Ping(innerCtx); err != nil {
				log.Debug("Redis poller: ping failed", zap.Error(err))
				return
			}

			p.Info("Redis poller %s", terminal.Green("connected"))

			for {
				select {
				case <-innerCtx.Done():
					return
				default:
					task, err := client.PopTask(innerCtx, qp.pollerCfg.RedisBRPOPTimeout)
					if err != nil {
						if innerCtx.Err() != nil {
							return
						}
						log.Debug("Redis poller error", zap.Error(err))
						time.Sleep(time.Second)
						continue
					}
					if task == nil {
						continue
					}

					runUUID := task.ScanID
					if runUUID == "" {
						runUUID = task.ID
					}

					if !markSeen(runUUID) {
						continue
					}

					// Dedup with DB
					if task.ScanID != "" {
						var dbRun database.Run
						err := dbConn.NewSelect().
							Model(&dbRun).
							Where("run_uuid = ?", task.ScanID).
							Scan(innerCtx)
						if err == nil && (dbRun.Status == "running" || dbRun.Status == "completed") {
							log.Debug("Skipping Redis task - already processed in DB",
								zap.String("run_uuid", task.ScanID))
							continue
						}
					}

					taskCh <- QueuedTask{
						Source:        "redis",
						RunUUID:       runUUID,
						RunID:         0,
						WorkflowName:  task.WorkflowName,
						WorkflowKind:  task.WorkflowKind,
						Target:        task.Target,
						Params:        task.Params,
						InputIsFile:   task.InputIsFile,
						InputFilePath: task.InputFilePath,
					}
				}
			}
		}()
	}

	// Executor goroutines
	for i := 0; i < qp.pollerCfg.Concurrency; i++ {
		qp.wg.Add(1)
		go func(workerNum int) {
			defer qp.wg.Done()
			executeQueuedTasks(innerCtx, qp.cfg, taskCh, workerNum, p)
		}(i)
	}

	return nil
}

// Stop cancels the poller context and waits for all goroutines to finish.
func (qp *QueuePoller) Stop() {
	qp.mu.Lock()
	if !qp.running {
		qp.mu.Unlock()
		return
	}
	qp.running = false
	qp.mu.Unlock()

	if qp.cancel != nil {
		qp.cancel()
	}
	qp.wg.Wait()
}

var (
	queueConcurrency int
	queueRedisURL    string

	// Flags for worker queue new
	queueFlow       string
	queueModule     string
	queueTargets    []string
	queueTargetFile string
	queueParams     []string
)

// workerQueueCmd is the parent command for queue subcommands
var workerQueueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Manage and process queued tasks",
	Long:  UsageWorkerQueue(),
}

// workerQueueListCmd lists queued tasks
var workerQueueListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List queued tasks",
	Long:    UsageWorkerQueueList(),
	RunE:    runWorkerQueueList,
}

// workerQueueNewCmd creates new queued tasks
var workerQueueNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Queue a new task for later processing",
	Long:  UsageWorkerQueueNew(),
	RunE:  runWorkerQueueNew,
}

// workerQueueRunCmd processes queued tasks
var workerQueueRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Process queued tasks from DB and Redis",
	Long:  UsageWorkerQueueRun(),
	RunE:  runWorkerQueueRun,
}

func init() {
	// Queue list flags
	workerQueueListCmd.Flags().StringVar(&queueRedisURL, "redis-url", "", "Redis connection URL (overrides settings)")

	// Queue new flags
	workerQueueNewCmd.Flags().StringVarP(&queueFlow, "flow", "f", "", "flow workflow name")
	workerQueueNewCmd.Flags().StringVarP(&queueModule, "module", "m", "", "module workflow name")
	workerQueueNewCmd.Flags().StringArrayVarP(&queueTargets, "target", "t", nil, "target(s) to queue")
	workerQueueNewCmd.Flags().StringVarP(&queueTargetFile, "target-file", "T", "", "file containing targets")
	workerQueueNewCmd.Flags().StringArrayVarP(&queueParams, "params", "p", nil, "additional parameters (key=value)")
	workerQueueNewCmd.Flags().StringVar(&queueRedisURL, "redis-url", "", "Redis connection URL (overrides settings)")

	// Queue run flags
	workerQueueRunCmd.Flags().IntVar(&queueConcurrency, "concurrency", 1, "number of concurrent task executors")
	workerQueueRunCmd.Flags().StringVar(&queueRedisURL, "redis-url", "", "Redis connection URL (overrides settings)")

	// Register subcommands
	workerQueueCmd.AddCommand(workerQueueListCmd)
	workerQueueCmd.AddCommand(workerQueueNewCmd)
	workerQueueCmd.AddCommand(workerQueueRunCmd)

	// Register queue under worker
	workerCmd.AddCommand(workerQueueCmd)
}

// runWorkerQueueList lists all queued runs from the database
func runWorkerQueueList(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return errConfigNotLoaded
	}

	ctx := context.Background()

	// Connect to database
	dbConn, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if err := database.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Query all queued runs (all statuses, not just "queued")
	var runs []*database.Run
	err = dbConn.NewSelect().
		Model(&runs).
		Where("is_queued = ?", true).
		Order("created_at ASC").
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("failed to query queued runs: %w", err)
	}

	// JSON output
	if globalJSON {
		if len(runs) == 0 {
			fmt.Println("[]")
			return nil
		}
		jsonBytes, err := json.MarshalIndent(runs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal runs: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	if len(runs) == 0 {
		printer.Info("No queued tasks found")
		return nil
	}

	// Display table
	printer.Section("Queued Tasks")

	headers := []string{"ID", "UUID", "Workflow", "Target", "Status", "Priority", "IsFile", "Created"}
	var rows [][]string
	for _, r := range runs {
		isFileStr := ""
		if r.InputIsFile {
			isFileStr = "yes"
		}
		rows = append(rows, []string{
			fmt.Sprintf("%d", r.ID),
			truncateUUID(r.RunUUID),
			r.WorkflowName,
			truncateQueueStr(r.Target, 40),
			colorizeQueueStatus(r.Status),
			r.RunPriority,
			isFileStr,
			r.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	printMarkdownTable(headers, rows)
	fmt.Println()
	printer.Info("Total: %d queued task(s)", len(runs))
	return nil
}

// runWorkerQueueNew creates new queued tasks
func runWorkerQueueNew(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return errConfigNotLoaded
	}

	// Override Redis config from URL if provided
	if queueRedisURL != "" {
		redisCfg, err := distributed.ParseRedisURL(queueRedisURL)
		if err != nil {
			return err
		}
		cfg.Redis = *redisCfg
	}

	// Determine workflow name and kind
	workflowName := queueFlow
	workflowKind := "flow"
	if workflowName == "" && queueModule != "" {
		workflowName = queueModule
		workflowKind = "module"
	}
	if workflowName == "" {
		return fmt.Errorf("workflow name required (use -f or -m)")
	}

	// Collect targets
	var allTargets []string
	allTargets = append(allTargets, queueTargets...)
	if queueTargetFile != "" {
		fileTargets, err := readTargetsFromFile(queueTargetFile)
		if err != nil {
			return fmt.Errorf("failed to read target file: %w", err)
		}
		allTargets = append(allTargets, fileTargets...)
	}
	allTargets = deduplicateTargets(allTargets)

	if len(allTargets) == 0 {
		return fmt.Errorf("no targets specified (use -t or -T)")
	}

	// Parse params
	params := make(map[string]interface{})
	for _, flag := range queueParams {
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}

	p := terminal.NewPrinter()
	return queueRuns(context.Background(), cfg, workflowName, workflowKind, allTargets, queueTargetFile, params, p)
}

// runWorkerQueueRun processes queued tasks with dual-source polling (DB + Redis)
func runWorkerQueueRun(cmd *cobra.Command, args []string) error {
	cfg := config.Get()
	if cfg == nil {
		return errConfigNotLoaded
	}

	// Ensure external binaries are in PATH
	ensureExternalBinariesInPath(cfg)

	if queueConcurrency <= 0 {
		queueConcurrency = 1
	}

	p := terminal.NewPrinter()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		p.Info("Shutting down queue runner...")
		cancel()
	}()

	qp := NewQueuePoller(cfg, QueuePollerConfig{
		DBPollInterval: 5 * time.Second,
		Concurrency:    queueConcurrency,
		RedisURL:       queueRedisURL,
	})

	if err := qp.Start(ctx); err != nil {
		return err
	}

	p.Success("Queue runner started (concurrency: %s)", terminal.Cyan(fmt.Sprintf("%d", queueConcurrency)))
	p.Info("Polling DB every %s%s...", terminal.Cyan("5s"), redisStatusSuffix(cfg))

	// Block until context is cancelled
	<-ctx.Done()
	qp.Stop()

	p.Info("Queue runner stopped")
	return nil
}

// executeQueuedTasks processes tasks from the channel
func executeQueuedTasks(ctx context.Context, cfg *config.Config, taskCh <-chan QueuedTask, workerNum int, p *terminal.Printer) {
	loader := parser.NewLoader(cfg.WorkflowsPath)

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-taskCh:
			if !ok {
				return
			}

			p.Info("[worker-%d] Processing: %s -> %s (source: %s)",
				workerNum, terminal.Yellow(task.WorkflowName), terminal.Green(task.Target), task.Source)

			// If InputIsFile, check if file exists locally
			if task.InputIsFile && task.InputFilePath != "" {
				if _, err := os.Stat(task.InputFilePath); os.IsNotExist(err) {
					p.Warning("[worker-%d] Input file %s not found, skipping task", workerNum, terminal.Yellow(task.InputFilePath))
					updateQueuedRunStatus(ctx, task.RunUUID, "failed", "input file not found: "+task.InputFilePath)
					continue
				}
			}

			// Load workflow
			workflow, err := loader.LoadWorkflow(task.WorkflowName)
			if err != nil {
				p.Error("[worker-%d] Failed to load workflow %s: %s", workerNum, task.WorkflowName, err)
				updateQueuedRunStatus(ctx, task.RunUUID, "failed", fmt.Sprintf("failed to load workflow: %v", err))
				continue
			}

			// Convert params to string map
			params := make(map[string]string)
			params["target"] = task.Target
			for k, v := range task.Params {
				if s, ok := v.(string); ok {
					params[k] = s
				}
			}

			// Create executor
			exec := executor.NewExecutor()
			exec.SetLoader(loader)
			exec.SetDBRunUUID(task.RunUUID)

			// Execute workflow
			var result *core.WorkflowResult
			if workflow.IsFlow() {
				result, err = exec.ExecuteFlow(ctx, workflow, params, cfg)
			} else {
				result, err = exec.ExecuteModule(ctx, workflow, params, cfg)
			}

			// Determine outcome
			if err != nil {
				p.Error("[worker-%d] Task failed for %s: %s", workerNum, task.Target, err)
				updateQueuedRunStatus(ctx, task.RunUUID, "failed", err.Error())
			} else if result != nil && result.Status == core.RunStatusFailed {
				errMsg := "workflow execution failed"
				if result.Error != nil {
					errMsg = result.Error.Error()
				}
				p.Error("[worker-%d] Task failed for %s: %s", workerNum, task.Target, errMsg)
				updateQueuedRunStatus(ctx, task.RunUUID, "failed", errMsg)
			} else {
				p.Success("[worker-%d] Task completed for %s", workerNum, terminal.Green(task.Target))
				updateQueuedRunStatus(ctx, task.RunUUID, "completed", "")
			}
		}
	}
}

// updateQueuedRunStatus updates the status of a queued run in the database
func updateQueuedRunStatus(ctx context.Context, runUUID, status, errorMsg string) {
	if runUUID == "" {
		return
	}
	_ = database.UpdateRunStatus(ctx, runUUID, status, errorMsg)
}

// truncateUUID returns a truncated UUID for display
func truncateUUID(u string) string {
	if len(u) > 8 {
		return u[:8]
	}
	return u
}

// truncateQueueStr truncates a string to maxLen with ellipsis
func truncateQueueStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// colorizeQueueStatus applies color to queue status
func colorizeQueueStatus(status string) string {
	switch status {
	case "queued":
		return terminal.Yellow(status)
	case "running":
		return terminal.Cyan(status)
	case "completed":
		return terminal.Green(status)
	case "failed":
		return terminal.Red(status)
	default:
		return terminal.Gray(status)
	}
}

// redisStatusSuffix returns a suffix string describing Redis connectivity
func redisStatusSuffix(cfg *config.Config) string {
	if cfg.IsRedisConfigured() {
		return " + " + terminal.Cyan("Redis")
	}
	return ""
}
