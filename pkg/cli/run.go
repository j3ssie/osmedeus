package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/distributed"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/fileio"
	"github.com/j3ssie/osmedeus/v5/internal/heuristics"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	flowName             string
	moduleNames          []string
	targets              []string
	targetFile           string
	paramFlags           []string
	paramsFile           string
	workspacePath        string
	dryRun               bool
	threadsHold          int
	runTactic            string
	excludeModules       []string
	spaceName            string
	workspacesFolder     string
	heuristicsCheck      string
	distributedRun       bool
	redisURLRun          string
	concurrency          int
	repeatRun            bool
	repeatWaitTime       string
	runTimeout           string
	stdModule            bool
	moduleURL            string
	emptyTarget          bool
	progressBar          bool
	disableWorkflowState bool

	// Chunk mode flags
	chunkSize    int
	chunkCount   int
	chunkPart    int
	chunkThreads int

	// Validation flags
	skipValidation bool

	// Server registration flag
	serverURL string

	// activeChunkInfo holds chunk info during execution (nil when not chunking)
	activeChunkInfo *ChunkInfo

	// explicitFlags tracks which CLI flags were explicitly set by the user
	// Used to determine precedence when applying workflow preferences
	explicitFlags map[string]bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a workflow",
	Long:  UsageRun(),
	RunE:  runRun,
}

func init() {
	runCmd.Flags().StringVarP(&flowName, "flow", "f", "", "flow workflow name to execute")
	runCmd.Flags().StringArrayVarP(&moduleNames, "module", "m", nil, "module workflow(s) to execute (can specify multiple)")
	runCmd.Flags().StringArrayVarP(&targets, "target", "t", nil, "target(s) to run against (can be specified multiple times)")
	runCmd.Flags().StringVarP(&targetFile, "target-file", "T", "", "file containing targets (one per line)")
	runCmd.Flags().StringArrayVarP(&paramFlags, "params", "p", nil, "additional parameters (key=value format)")
	runCmd.Flags().StringVarP(&paramsFile, "params-file", "P", "", "file containing parameters (JSON or YAML key:value pairs)")
	runCmd.Flags().StringVarP(&workspacePath, "workspace", "w", "", "custom workspace path (overrides {{TargetSpace}})")
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be executed without running commands")
	runCmd.Flags().IntVar(&threadsHold, "threads-hold", 0, "override thread count (0 = use tactic default)")
	runCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "number of targets to run concurrently")
	runCmd.Flags().StringVarP(&runTactic, "tactic", "B", "default", "run tactic: aggressive, default, gently")
	runCmd.Flags().StringArrayVarP(&excludeModules, "exclude", "x", nil, "module(s) to exclude from execution (can be specified multiple times)")
	runCmd.Flags().StringVarP(&spaceName, "space", "S", "", "override {{TargetSpace}} variable")
	runCmd.Flags().StringVarP(&workspacesFolder, "workspaces-folder", "W", "", "override {{Workspaces}} variable")
	runCmd.Flags().StringVar(&heuristicsCheck, "heuristics-check", "basic", "heuristics check level: none, basic, advanced")
	runCmd.Flags().BoolVarP(&distributedRun, "distributed-run", "D", false, "submit run to distributed worker queue (requires Redis)")
	runCmd.Flags().StringVar(&redisURLRun, "redis-url", "", "Redis connection URL for distributed mode (overrides settings)")
	runCmd.Flags().BoolVar(&repeatRun, "repeat", false, "repeat run after completion")
	runCmd.Flags().StringVar(&repeatWaitTime, "repeat-wait-time", "1m", "wait time between repeats (e.g., 30s, 20m, 10h, 1d)")
	runCmd.Flags().StringVar(&runTimeout, "timeout", "", "run timeout (e.g., 2h, 3h, 1d)")
	runCmd.Flags().BoolVar(&stdModule, "std-module", false, "read module YAML from stdin")
	runCmd.Flags().StringVar(&moduleURL, "module-url", "", "URL to fetch module YAML from (supports GitHub private repos)")
	runCmd.Flags().BoolVar(&emptyTarget, "empty-target", false, "run without target (generates placeholder target)")
	runCmd.Flags().BoolVarP(&progressBar, "progress-bar", "G", false, "show progress bar during execution (enables silent mode)")
	runCmd.Flags().BoolVar(&disableWorkflowState, "disable-workflow-state", false, "disable writing workflow YAML to output directory")

	// Chunk mode flags
	runCmd.Flags().IntVar(&chunkSize, "chunk-size", 0, "split targets into chunks of N targets each (0 = disabled)")
	runCmd.Flags().IntVar(&chunkCount, "chunk-count", 0, "split targets into N equal chunks (0 = disabled)")
	runCmd.Flags().IntVar(&chunkPart, "chunk-part", -1, "execute only chunk M (0-indexed, requires --chunk-size or --chunk-count)")
	runCmd.Flags().IntVar(&chunkThreads, "chunk-threads", 0, "override concurrency within chunk (0 = use -c value)")

	// Validation flags
	runCmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "skip target type validation from dependencies.variables")

	// Server registration flag
	runCmd.Flags().StringVar(&serverURL, "server-url", "", "Server URL for cron trigger registration (e.g., http://localhost:8002)")
}

// captureExplicitFlags records which CLI flags were explicitly set by the user
// This is used to determine precedence when applying workflow preferences
func captureExplicitFlags(cmd *cobra.Command) {
	explicitFlags = make(map[string]bool)

	// Run command flags
	runFlagNames := []string{
		"heuristics-check", "repeat", "repeat-wait-time",
	}
	for _, name := range runFlagNames {
		if f := cmd.Flags().Lookup(name); f != nil {
			explicitFlags[name] = f.Changed
		}
	}

	// Global/persistent flags (from root command)
	globalFlagNames := []string{
		"silent", "disable-logging", "disable-notification", "ci-output-format",
	}
	for _, name := range globalFlagNames {
		// Check both local and inherited persistent flags
		if f := cmd.Flags().Lookup(name); f != nil {
			explicitFlags[name] = f.Changed
		} else if f := cmd.InheritedFlags().Lookup(name); f != nil {
			explicitFlags[name] = f.Changed
		}
	}
}

// applyWorkflowPreferences applies workflow preferences to CLI variables
// Only applies if the corresponding CLI flag was NOT explicitly set by the user
func applyWorkflowPreferences(prefs *core.Preferences, printer *terminal.Printer) {
	if prefs == nil {
		return
	}

	applied := []string{}

	// disable_notifications -> disableNotification (global)
	if prefs.DisableNotifications != nil && !explicitFlags["disable-notification"] {
		disableNotification = *prefs.DisableNotifications
		if *prefs.DisableNotifications {
			applied = append(applied, "disable_notifications")
		}
	}

	// disable_logging -> disableLogging (global)
	if prefs.DisableLogging != nil && !explicitFlags["disable-logging"] {
		disableLogging = *prefs.DisableLogging
		if *prefs.DisableLogging {
			applied = append(applied, "disable_logging")
		}
	}

	// heuristics_check -> heuristicsCheck
	if prefs.HeuristicsCheck != nil && !explicitFlags["heuristics-check"] {
		heuristicsCheck = *prefs.HeuristicsCheck
		applied = append(applied, "heuristics_check="+*prefs.HeuristicsCheck)
	}

	// ci_output_format -> ciOutputFormat (global)
	if prefs.CIOutputFormat != nil && !explicitFlags["ci-output-format"] {
		ciOutputFormat = *prefs.CIOutputFormat
		if *prefs.CIOutputFormat {
			terminal.SetCIMode(true)
			terminal.SetColorEnabled(false)
			applied = append(applied, "ci_output_format")
		}
	}

	// silent -> silent (global)
	if prefs.Silent != nil && !explicitFlags["silent"] {
		silent = *prefs.Silent
		if *prefs.Silent {
			applied = append(applied, "silent")
		}
	}

	// repeat -> repeatRun
	if prefs.Repeat != nil && !explicitFlags["repeat"] {
		repeatRun = *prefs.Repeat
		if *prefs.Repeat {
			applied = append(applied, "repeat")
		}
	}

	// repeat_wait_time -> repeatWaitTime
	if prefs.RepeatWaitTime != nil && !explicitFlags["repeat-wait-time"] {
		repeatWaitTime = *prefs.RepeatWaitTime
		applied = append(applied, "repeat_wait_time="+*prefs.RepeatWaitTime)
	}

	// Log applied preferences if verbose
	if len(applied) > 0 && verbose {
		printer.Info("Applied workflow preferences: %s", strings.Join(applied, ", "))
	}
}

// handleTargetTypeMismatchError checks if err is a TargetTypeMismatchError and prints it formatted.
// Returns true if it was handled, false otherwise.
func handleTargetTypeMismatchError(err error) bool {
	var ttmErr *executor.TargetTypeMismatchError
	if errors.As(err, &ttmErr) {
		fmt.Println()
		fmt.Printf("%s %s\n", terminal.Red("✘"), terminal.BoldRed("Target type mismatch"))
		fmt.Printf("  Supplied: %s\n", ttmErr.Supplied)
		fmt.Printf("  %s %s\n", terminal.HiBlue("Required Params:"), ttmErr.ExpectedType)
		fmt.Println()
		fmt.Printf("  %s %s\n", terminal.Yellow("💡"), terminal.HiBlue("\"Target\" in Required Params is supplied via -t flag (e.g., -t example.com) or each line from -T list-of-targets.txt"))
		fmt.Printf("  %s %s\n", terminal.Yellow("💡"), terminal.Yellow("Hint: Use --skip-validation to bypass this check"))
		fmt.Println()
		return true
	}
	return false
}

func runRun(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	// Capture which CLI flags were explicitly set (for preference merging)
	captureExplicitFlags(cmd)

	// Print greeting message (skip in CI mode)
	if !ciOutputFormat {
		printer.Println("%s Initiating Osmedeus %s - Crafted with %s by %s",
			terminal.Yellow(terminal.SymbolLightning),
			terminal.Cyan(core.VERSION),
			terminal.Red("<3"),
			terminal.Yellow(core.AUTHOR))
		printer.Newline()
	}

	// Auto-enable no-log when progress bar is enabled
	if progressBar {
		// Re-initialize logger with silent mode to suppress log output
		logCfg := logger.DefaultConfig()
		logCfg.Level = "error"
		logCfg.Silent = true
		_ = logger.Init(logCfg)
	}

	// Get logger after potential re-initialization
	log := logger.Get()

	// Validate flags
	if flowName == "" && len(moduleNames) == 0 && !stdModule && moduleURL == "" {
		printer.Warning("No workflow specified. Using default flow: general")
		printer.Info("Tip: Use -f <flow_name> or -m <module_name> to select a workflow")
		fmt.Println()
		flowName = "general"
	}
	if flowName != "" && len(moduleNames) > 0 {
		return fmt.Errorf("only one of --flow or --module can be specified")
	}
	if stdModule && (flowName != "" || len(moduleNames) > 0) {
		return fmt.Errorf("--std-module cannot be combined with --flow or --module")
	}
	if moduleURL != "" && (flowName != "" || len(moduleNames) > 0 || stdModule) {
		return fmt.Errorf("--module-url cannot be combined with --flow, --module, or --std-module")
	}

	// Parse timeout duration
	var timeoutDuration time.Duration
	if runTimeout != "" {
		var err error
		timeoutDuration, err = parseRunDuration(runTimeout)
		if err != nil {
			return fmt.Errorf("invalid timeout: %w", err)
		}
		printer.Info("Run timeout: %s", runTimeout)
	}

	// Parse repeat wait time
	var waitDuration time.Duration
	if repeatRun {
		var err error
		waitDuration, err = parseRunDuration(repeatWaitTime)
		if err != nil {
			return fmt.Errorf("invalid repeat-wait-time: %w", err)
		}
		printer.Info("Repeat mode enabled, wait time: %s", repeatWaitTime)
	}

	cfg := config.Get()
	if cfg == nil {
		return fmt.Errorf("configuration not loaded")
	}

	// Ensure external binaries are in PATH at runtime
	// This helps when users haven't reloaded their shell after installation
	ensureExternalBinariesInPath(cfg)

	// Collect all targets from flags, file, and stdin
	log.Debug("Collecting targets",
		zap.Strings("flag_targets", targets),
		zap.String("target_file", targetFile),
	)
	allTargets, err := collectTargets()
	if err != nil {
		return err
	}

	log.Debug("Targets collected",
		zap.Int("count", len(allTargets)),
		zap.Strings("targets", allTargets),
	)

	if len(allTargets) == 0 {
		if emptyTarget {
			// Generate placeholder target
			allTargets = []string{generateEmptyTarget()}
			printer.Info("Using generated target: %s", allTargets[0])
		} else {
			return fmt.Errorf("no targets specified. Use -t, -T, pipe targets via stdin, or use --empty-target")
		}
	}

	// Validate chunk flags mutual exclusivity
	if chunkSize > 0 && chunkCount > 0 {
		return fmt.Errorf("cannot use both --chunk-size and --chunk-count")
	}

	// Convert --chunk-count to --chunk-size
	if chunkCount > 0 && len(allTargets) > 0 {
		chunkSize = (len(allTargets) + chunkCount - 1) / chunkCount // ceiling division
	}

	// Apply chunking if enabled
	activeChunkInfo = nil // Reset chunk info
	if chunkSize > 0 {
		chunkedTargets, info, err := chunkTargets(allTargets, chunkSize, chunkPart)
		if err != nil {
			if err.Error() == "chunk-info" {
				// Info mode - display chunk breakdown
				printer.Info("Chunk Info: %d total targets, %d chunks of size %d",
					len(allTargets), info.Total, info.Size)
				for i := 0; i < info.Total; i++ {
					start := i * info.Size
					end := start + info.Size
					if end > len(allTargets) {
						end = len(allTargets)
					}
					printer.Info("  Chunk %d: targets %d-%d (%d targets)",
						i, start, end-1, end-start)
				}
				return nil
			}
			return err
		}
		allTargets = chunkedTargets
		activeChunkInfo = info
		printer.Info("Processing chunk %d/%d (%d targets, indices %d-%d)",
			info.Index+1, info.Total, len(allTargets), info.Start, info.End-1)
	}

	// Apply chunk-threads override
	if chunkThreads > 0 && chunkSize > 0 {
		concurrency = chunkThreads
	}

	// Handle distributed run mode
	if distributedRun {
		return runDistributedRun(cfg, allTargets, printer)
	}

	loader := parser.NewLoader(cfg.WorkflowsPath)

	// Execute workflow for each target (with concurrency)
	if concurrency <= 0 {
		concurrency = 1
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Main run loop (handles repeat)
	iteration := 0
	for {
		iteration++
		if repeatRun && iteration > 1 {
			printer.Section(fmt.Sprintf("Repeat Iteration %d", iteration))
		}

		// Create context with timeout if specified
		var ctx context.Context
		var cancel context.CancelFunc
		if timeoutDuration > 0 {
			ctx, cancel = context.WithTimeout(context.Background(), timeoutDuration)
		} else {
			ctx, cancel = context.WithCancel(context.Background())
		}

		// Handle interrupt signals in goroutine
		go func() {
			select {
			case <-sigChan:
				log.Warn("Received interrupt signal, cancelling...")
				printer.Warning("Interrupt received, cancelling run...")
				cancel()
			case <-ctx.Done():
			}
		}()

		var lastErr error

		if stdModule {
			// Stdin module mode - read workflow from stdin
			workflow, err := readWorkflowFromStdin()
			if err != nil {
				return err
			}

			printer.Success("Workflow loaded from stdin: %s (%s)", workflow.Name, terminal.TypeBadge(string(workflow.Kind)))

			// Apply workflow preferences (if any) - CLI flags take precedence
			applyWorkflowPreferences(workflow.Preferences, printer)

			if workflow.IsFlow() {
				return fmt.Errorf("--std-module only supports module workflows, got flow")
			}

			// Execute for all targets (nil loader since flows not supported for stdin)
			lastErr = executeSingleWorkflowDirect(ctx, workflow, allTargets, cfg, printer, log, nil)
		} else if moduleURL != "" {
			// URL module mode - fetch workflow from URL
			workflow, err := fetchWorkflowFromURL(moduleURL)
			if err != nil {
				return fmt.Errorf("failed to fetch workflow from URL: %w", err)
			}

			printer.Success("Workflow fetched from URL: %s (%s)", workflow.Name, terminal.TypeBadge(string(workflow.Kind)))

			// Apply workflow preferences (if any) - CLI flags take precedence
			applyWorkflowPreferences(workflow.Preferences, printer)

			if workflow.IsFlow() {
				return fmt.Errorf("--module-url only supports module workflows, got flow")
			}

			// Execute for all targets (nil loader since flows not supported for URL modules)
			lastErr = executeSingleWorkflowDirect(ctx, workflow, allTargets, cfg, printer, log, nil)
		} else if flowName != "" {
			// Flow mode - single workflow
			lastErr = executeSingleWorkflow(ctx, loader, flowName, allTargets, cfg, printer, log)
		} else {
			// Module mode - run each module in sequence
			for i, moduleName := range moduleNames {
				if len(moduleNames) > 1 {
					printer.Section(fmt.Sprintf("Module %d/%d: %s", i+1, len(moduleNames), moduleName))
				}

				moduleErr := executeSingleWorkflow(ctx, loader, moduleName, allTargets, cfg, printer, log)
				if moduleErr != nil {
					if ctx.Err() != nil {
						// Context cancelled (timeout or interrupt)
						lastErr = moduleErr
						break
					}
					// Skip printing for TargetTypeMismatchError (already printed)
					var ttmErr *executor.TargetTypeMismatchError
					if !errors.As(moduleErr, &ttmErr) {
						printer.Error("Module %s failed: %s", moduleName, moduleErr)
					}
					lastErr = moduleErr
					// Continue to next module
				}
			}
		}

		// Check context error before calling cancel (cancel sets Canceled)
		ctxErr := ctx.Err()
		cancel()

		// Check if timeout was exceeded
		if ctxErr == context.DeadlineExceeded {
			printer.Error("Run timed out after %s", runTimeout)
			return fmt.Errorf("run timed out after %s", runTimeout)
		}

		// Check if interrupted by signal (not by our own cancel)
		if ctxErr == context.Canceled {
			return fmt.Errorf("run cancelled")
		}

		// Handle repeat
		if !repeatRun {
			return lastErr
		}

		printer.Info("Run iteration %d completed. Waiting %s before next iteration...", iteration, repeatWaitTime)
		printer.Info("Press Ctrl+C to stop repeat mode")

		// Wait with interrupt handling
		select {
		case <-time.After(waitDuration):
			// Continue to next iteration
		case <-sigChan:
			printer.Info("Interrupt received, stopping repeat mode")
			return nil
		}
	}
}

// executeSingleWorkflow loads and executes a single workflow against all targets
func executeSingleWorkflow(ctx context.Context, loader *parser.Loader, workflowName string, allTargets []string, cfg *config.Config, printer *terminal.Printer, log *zap.Logger) error {
	log.Debug("Loading workflow",
		zap.String("workflow_name", workflowName),
	)

	var sp *terminal.Spinner
	if showSpinner {
		sp = terminal.LoadingSpinner("Loading workflow " + workflowName)
		sp.Start()
	}

	workflow, err := loader.LoadWorkflow(workflowName)
	if sp != nil {
		sp.Stop()
	}

	if err != nil {
		printer.Error("Failed to load workflow: %s", err)
		return fmt.Errorf("failed to load workflow: %w", err)
	}

	printer.Success("Workflow loaded: %s (%s)", workflow.Name, terminal.TypeBadge(string(workflow.Kind)))
	log.Info("Workflow loaded",
		zap.String("name", workflow.Name),
		zap.String("kind", string(workflow.Kind)),
	)

	// Apply workflow preferences (if any) - CLI flags take precedence
	applyWorkflowPreferences(workflow.Preferences, printer)

	// Register cron triggers with server (async, best-effort)
	if len(allTargets) > 0 {
		params := map[string]string{"target": allTargets[0], "tactic": runTactic}
		go registerCronTriggersWithServer(ctx, workflow, allTargets[0], params, cfg, printer, log)
	}

	// Show target count and concurrency
	if len(allTargets) > 1 {
		printer.Info("Running against %d targets (concurrency: %d)", len(allTargets), concurrency)
	}

	log.Debug("Starting concurrent execution",
		zap.Int("target_count", len(allTargets)),
		zap.Int("concurrency", concurrency),
		zap.String("tactic", runTactic),
		zap.Bool("dry_run", dryRun),
	)

	results, lastErr := executeRunsConcurrentlyWithContext(ctx, workflow, allTargets, cfg, concurrency, loader)

	// Print summary for multiple targets
	if len(allTargets) > 1 {
		printMultiTargetSummary(results, len(allTargets))
	}

	return lastErr
}

// executeRunsConcurrentlyWithContext runs workflows for multiple targets with controlled concurrency and context
func executeRunsConcurrentlyWithContext(ctx context.Context, workflow *core.Workflow, targets []string, cfg *config.Config, maxConcurrency int, loader *parser.Loader) ([]*core.WorkflowResult, error) {
	printer := terminal.NewPrinter()

	type scanResult struct {
		index  int
		result *core.WorkflowResult
		err    error
	}

	sem := make(chan struct{}, maxConcurrency) // Semaphore for concurrency control
	results := make(chan scanResult, len(targets))
	var wg sync.WaitGroup

	for i, target := range targets {
		wg.Add(1)
		go func(idx int, t string) {
			defer wg.Done()

			// Check if context is cancelled
			select {
			case <-ctx.Done():
				results <- scanResult{index: idx, result: nil, err: ctx.Err()}
				return
			default:
			}

			// Acquire semaphore
			sem <- struct{}{}
			defer func() { <-sem }()

			if len(targets) > 1 {
				printer.Info("[%d/%d] Starting: %s", idx+1, len(targets), t)
			}

			result, err := executeRunForTargetWithContext(ctx, workflow, t, cfg, loader)
			results <- scanResult{index: idx, result: result, err: err}
		}(i, target)
	}

	// Close results channel when all done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	allResults := make([]*core.WorkflowResult, len(targets))
	var lastErr error

	for r := range results {
		allResults[r.index] = r.result
		if r.err != nil {
			// Skip printing for TargetTypeMismatchError (already printed)
			var ttmErr *executor.TargetTypeMismatchError
			if !errors.As(r.err, &ttmErr) {
				printer.Error("Failed for target %s: %s", targets[r.index], r.err)
			}
			lastErr = r.err
		}
	}

	return allResults, lastErr
}

// executeRunForTargetWithContext executes the workflow for a single target with context support
func executeRunForTargetWithContext(ctx context.Context, workflow *core.Workflow, target string, cfg *config.Config, loader *parser.Loader) (*core.WorkflowResult, error) {
	log := logger.Get()

	log.Debug("Starting run for target",
		zap.String("target", target),
		zap.String("workflow", workflow.Name),
	)

	// Parse parameters: file params first, then CLI params override
	params := make(map[string]string)

	// Load params from file if specified
	if paramsFile != "" {
		fileParams, err := loadParamsFromFile(paramsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load params file: %w", err)
		}
		for k, v := range fileParams {
			params[k] = v
		}
		log.Debug("Loaded params from file",
			zap.String("file", paramsFile),
			zap.Int("count", len(fileParams)),
		)
	}

	// CLI params (-p) override file params
	for k, v := range parseParams(paramFlags) {
		params[k] = v
	}

	// Set built-in params (override user params)
	params["target"] = target
	params["target_file"] = targetFile
	params["tactic"] = runTactic
	params["threads_hold"] = fmt.Sprintf("%d", threadsHold)
	params["exclude_modules"] = strings.Join(excludeModules, ",")
	params["space_name"] = spaceName
	params["workspaces_folder"] = workspacesFolder
	params["heuristics_check"] = heuristicsCheck

	// Add chunk params if chunking is active
	if activeChunkInfo != nil {
		params["chunk_index"] = fmt.Sprintf("%d", activeChunkInfo.Index)
		params["chunk_size"] = fmt.Sprintf("%d", activeChunkInfo.Size)
		params["total_chunks"] = fmt.Sprintf("%d", activeChunkInfo.Total)
		params["chunk_start"] = fmt.Sprintf("%d", activeChunkInfo.Start)
		params["chunk_end"] = fmt.Sprintf("%d", activeChunkInfo.End)
	}

	log.Debug("Run parameters configured",
		zap.String("target", target),
		zap.String("tactic", runTactic),
		zap.Int("threads_hold", threadsHold),
		zap.Strings("exclude_modules", excludeModules),
		zap.Int("param_count", len(params)),
	)

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Create run record in database (skip for dry-run)
	var runUUID string
	var runID int64
	if !dryRun {
		runUUID, runID = createCLIRunRecord(ctx, cfg, workflow, target, params, loader)
	}

	// Create executor
	log.Debug("Creating executor",
		zap.Bool("dry_run", dryRun),
		zap.Bool("spinner", showSpinner),
		zap.Bool("verbose", verbose),
		zap.Bool("progress_bar", progressBar),
	)
	exec := executor.NewExecutor()
	exec.SetDryRun(dryRun)
	exec.SetDisableWorkflowState(disableWorkflowState)
	exec.SetSkipValidation(skipValidation)
	exec.SetSpinner(showSpinner)
	exec.SetVerbose(verbose) // Show actual step output in verbose mode
	exec.SetSilent(silent)   // Hide step output in silent mode
	if loader != nil {
		exec.SetLoader(loader) // Set loader for flow execution (loading nested modules)
	}

	// Set up database progress tracking
	if runUUID != "" {
		exec.SetDBRunUUID(runUUID)
		exec.SetDBRunID(runID)
		exec.SetOnStepCompleted(func(stepCtx context.Context, dbRunUUID string) {
			_ = database.IncrementRunCompletedSteps(stepCtx, dbRunUUID)
		})
	}

	// Create progress bar if enabled
	var pb *terminal.ProgressBar
	if progressBar && !dryRun {
		pb = terminal.NewProgressBar(len(workflow.Steps), workflow.Name)
		exec.SetProgressBar(pb)
	}

	// Print dry-run header if enabled
	if dryRun {
		// Calculate thread values
		threads, baseThreads := cfg.GetThreads(runTactic)
		if threadsHold > 0 {
			threads = threadsHold
			baseThreads = threadsHold / 2
			if baseThreads < 1 {
				baseThreads = 1
			}
		}

		separator := strings.Repeat("═", 40)

		fmt.Println()
		fmt.Printf("%s %s %s\n", terminal.Yellow("⚠"), terminal.BoldYellow("DRY-RUN Mode"), terminal.Gray("- No commands will be executed"))
		fmt.Println(terminal.Yellow(separator))
		fmt.Printf("%s Workflow: %s\n", terminal.Cyan("│"), terminal.Bold(workflow.Name))
		fmt.Printf("%s Target:   %s\n", terminal.Cyan("│"), terminal.Cyan(target))
		fmt.Printf("%s Steps:    %s\n", terminal.Cyan("│"), terminal.Gray(fmt.Sprintf("%d", len(workflow.Steps))))
		fmt.Printf("%s Tactic:   %s\n", terminal.Cyan("│"), terminal.Gray(runTactic))
		fmt.Println()
		fmt.Printf("%s %s\n", terminal.Cyan("✦"), terminal.Bold("Builtin Variables"))
		fmt.Printf("  %s BaseFolder:     %s\n", terminal.Gray("│"), terminal.Gray(cfg.BaseFolder))
		fmt.Printf("  %s Binaries:       %s\n", terminal.Gray("│"), terminal.Gray(cfg.BinariesPath))
		fmt.Printf("  %s Data:           %s\n", terminal.Gray("│"), terminal.Gray(cfg.DataPath))
		fmt.Printf("  %s Workspaces:     %s\n", terminal.Gray("│"), terminal.Gray(cfg.WorkspacesPath))
		fmt.Printf("  %s Output:         %s\n", terminal.Gray("│"), terminal.Gray(cfg.WorkspacesPath+"/"+target))
		fmt.Printf("  %s threads:        %s\n", terminal.Gray("│"), terminal.Gray(fmt.Sprintf("%d", threads)))
		fmt.Printf("  %s baseThreads:    %s\n", terminal.Gray("│"), terminal.Gray(fmt.Sprintf("%d", baseThreads)))
		fmt.Printf("  %s Today:          %s\n", terminal.Gray("│"), terminal.Gray(time.Now().Format("2006-01-02")))
		fmt.Println(terminal.Yellow(separator))
		fmt.Println()
	}

	// Execute workflow
	log.Debug("Executing workflow",
		zap.String("workflow", workflow.Name),
		zap.String("kind", string(workflow.Kind)),
		zap.String("target", target),
		zap.Bool("is_flow", workflow.IsFlow()),
	)
	var result *core.WorkflowResult
	var err error
	if workflow.IsFlow() {
		result, err = exec.ExecuteFlow(ctx, workflow, params, cfg)
	} else {
		result, err = exec.ExecuteModule(ctx, workflow, params, cfg)
	}

	if err != nil {
		// Abort progress bar on error
		if pb != nil {
			pb.Abort()
		}

		// Handle target type mismatch error specially (print once, skip logging)
		if handleTargetTypeMismatchError(err) {
			// Update run status to failed in database
			// Use a fresh context for cleanup - the original ctx may be cancelled
			if runUUID != "" {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
				_ = database.UpdateRunStatus(cleanupCtx, runUUID, "failed", err.Error())
				cleanupCancel()
			}
			return nil, err
		}

		// Determine status based on error type
		status := "failed"
		if errors.Is(err, context.Canceled) {
			status = "cancelled"
			log.Warn("Workflow execution cancelled",
				zap.String("workflow", workflow.Name),
				zap.String("target", target),
			)
		} else {
			log.Error("Workflow execution failed",
				zap.String("workflow", workflow.Name),
				zap.String("target", target),
				zap.Error(err),
			)
		}

		// Update run status in database
		// Use a fresh context for cleanup - the original ctx may be cancelled
		if runUUID != "" {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
			if updateErr := database.UpdateRunStatus(cleanupCtx, runUUID, status, err.Error()); updateErr != nil {
				log.Error("Failed to update run status on cancellation",
					zap.String("run_uuid", runUUID),
					zap.String("status", status),
					zap.Error(updateErr),
				)
			}
			cleanupCancel()
		}
		return nil, err
	}

	// Finish progress bar on success
	if pb != nil {
		pb.Finish(!silent) // Show output unless silent mode
	}

	log.Debug("Workflow execution completed",
		zap.String("workflow", workflow.Name),
		zap.String("target", target),
		zap.String("status", string(result.Status)),
		zap.Int("step_results", len(result.Steps)),
		zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
	)

	// Update run status to completed in database
	if runUUID != "" {
		_ = database.UpdateRunStatus(ctx, runUUID, "completed", "")
	}

	// Print result summary for this target (skip if progress bar was used - it shows its own summary)
	if pb == nil {
		printResultSummary(result)
	}

	return result, nil
}

// calculateTotalSteps returns the appropriate step count based on workflow kind.
// For module workflows, it returns len(Steps).
// For flow workflows, it loads each module and sums their step counts.
func calculateTotalSteps(workflow *core.Workflow, loader *parser.Loader) int {
	if workflow.Kind != core.KindFlow {
		return len(workflow.Steps)
	}

	// Flow workflow: sum steps from all modules
	if loader == nil {
		return len(workflow.Modules)
	}

	log := logger.Get()
	totalSteps := 0

	for _, modRef := range workflow.Modules {
		if modRef.Path == "" {
			totalSteps++
			continue
		}

		module, err := loader.LoadWorkflowByPath(modRef.Path)
		if err != nil {
			log.Warn("Failed to load module for step counting",
				zap.String("module", modRef.Name),
				zap.String("path", modRef.Path),
				zap.Error(err),
			)
			totalSteps++
			continue
		}

		totalSteps += len(module.Steps)
	}

	return totalSteps
}

// computeWorkspace computes the workspace name from target and params
// This mirrors the executor's logic for computing TargetSpace
func computeWorkspace(target string, params map[string]string) string {
	// If -S flag provided, use it directly
	if spaceName := params["space_name"]; spaceName != "" {
		return spaceName
	}

	// Use heuristics to extract root domain/host (matches executor behavior)
	heuristicsLevel := params["heuristics_check"]
	if heuristicsLevel == "" {
		heuristicsLevel = "basic"
	}
	if heuristicsLevel != "none" {
		info, err := heuristics.Analyze(target, heuristicsLevel)
		if err == nil && info != nil && info.RootDomain != "" {
			return sanitizeTargetForWorkspace(info.RootDomain)
		}
	}

	// Otherwise, sanitize the target for filesystem safety
	return sanitizeTargetForWorkspace(target)
}

// sanitizeTargetForWorkspace creates a filesystem-safe workspace name from target
// This mirrors the executor's sanitizeTargetSpace function
func sanitizeTargetForWorkspace(target string) string {
	sanitized := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) {
			return '_'
		}
		return r
	}, target)
	// Limit length to avoid filesystem issues
	if len(sanitized) > 200 {
		sanitized = sanitized[:200]
	}
	return sanitized
}

// createCLIRunRecord creates a run record in the database for CLI executions
// Returns the RunUUID (string) and Run.ID (int64)
func createCLIRunRecord(ctx context.Context, cfg *config.Config, workflow *core.Workflow, target string, params map[string]string, loader *parser.Loader) (string, int64) {
	log := logger.Get()

	// Connect to database
	_, err := database.Connect(cfg)
	if err != nil {
		log.Debug("Failed to connect to database for run record", zap.Error(err))
		return "", 0
	}

	// Migrate database schema if needed
	if err := database.Migrate(ctx); err != nil {
		log.Debug("Failed to migrate database for run record", zap.Error(err))
		return "", 0
	}

	now := time.Now()
	runUUID := uuid.New().String()

	// Convert params to interface map
	paramsInterface := make(map[string]interface{})
	for k, v := range params {
		paramsInterface[k] = v
	}

	// Compute workspace from target and params
	workspace := computeWorkspace(target, params)

	run := &database.Run{
		RunUUID:      runUUID,
		WorkflowName: workflow.Name,
		WorkflowKind: string(workflow.Kind),
		Target:       target,
		Params:       paramsInterface,
		Status:       "running",
		TriggerType:  "cli",
		StartedAt:    &now,
		TotalSteps:   calculateTotalSteps(workflow, loader),
		Workspace:    workspace,
	}

	if err := database.CreateRun(ctx, run); err != nil {
		log.Debug("Failed to create run record", zap.Error(err))
		return "", 0
	}

	log.Debug("Created run record", zap.String("run_uuid", runUUID), zap.Int64("run_id", run.ID))
	return runUUID, run.ID
}

// collectTargets gathers targets from all input sources: flags, file, and stdin
// When stdModule is true, stdin is reserved for the workflow YAML, not targets
func collectTargets() ([]string, error) {
	var allTargets []string

	// 1. Add targets from -t flags
	allTargets = append(allTargets, targets...)

	// 2. Read targets from file if -T is provided
	if targetFile != "" {
		fileTargets, err := readTargetsFromFile(targetFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read target file: %w", err)
		}
		allTargets = append(allTargets, fileTargets...)
	}

	// 3. Read targets from stdin if piped (skip if --std-module is used - stdin is for workflow)
	if !stdModule {
		stdinTargets, err := readTargetsFromStdin()
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
		allTargets = append(allTargets, stdinTargets...)
	}

	// Deduplicate and filter empty lines
	return deduplicateTargets(allTargets), nil
}

// readTargetsFromFile reads targets from a file, one per line.
// Uses memory-mapped I/O for large files (>1MB) for 40-60% faster loading.
func readTargetsFromFile(path string) ([]string, error) {
	return fileio.ReadLinesFiltered(path)
}

// readTargetsFromStdin reads targets from stdin if data is piped
func readTargetsFromStdin() ([]string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, nil // Ignore stat errors, just skip stdin
	}

	// Check if stdin has piped data (not a terminal)
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, nil // No piped data
	}

	var result []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}
	return result, scanner.Err()
}

// deduplicateTargets removes duplicates and empty strings
func deduplicateTargets(inputTargets []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, t := range inputTargets {
		t = strings.TrimSpace(t)
		if t != "" && !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}
	return result
}

// ChunkInfo holds metadata about the current chunk
type ChunkInfo struct {
	Index int
	Size  int
	Total int
	Start int
	End   int
}

// chunkTargets splits targets into chunks and returns the specified chunk
func chunkTargets(allTargets []string, size, part int) ([]string, *ChunkInfo, error) {
	if size <= 0 {
		return allTargets, nil, nil // No chunking
	}

	total := len(allTargets)
	if total == 0 {
		return allTargets, nil, nil // Empty targets, no chunking needed
	}

	totalChunks := (total + size - 1) / size // Ceiling division

	if part < 0 {
		// Info mode - return error with chunk details
		return nil, &ChunkInfo{Total: totalChunks, Size: size},
			fmt.Errorf("chunk-info")
	}

	if part >= totalChunks {
		return nil, nil, fmt.Errorf("chunk-part %d exceeds total chunks %d (0-indexed)",
			part, totalChunks)
	}

	start := part * size
	end := start + size
	if end > total {
		end = total
	}

	info := &ChunkInfo{
		Index: part,
		Size:  size,
		Total: totalChunks,
		Start: start,
		End:   end,
	}

	return allTargets[start:end], info, nil
}

// printMultiTargetSummary prints a summary for multiple target execution
func printMultiTargetSummary(results []*core.WorkflowResult, totalTargets int) {
	succeeded := 0
	failed := 0
	for _, r := range results {
		if r.Status == core.RunStatusCompleted {
			succeeded++
		} else {
			failed++
		}
	}
	skipped := totalTargets - len(results)

	// CI mode: output JSON
	if ciOutputFormat {
		ciResults := make([]map[string]interface{}, 0, len(results))
		for _, r := range results {
			if r != nil {
				ciResults = append(ciResults, formatResultForCI(r))
			}
		}
		output := map[string]interface{}{
			"type":          "multi_target_summary",
			"total_targets": totalTargets,
			"succeeded":     succeeded,
			"failed":        failed,
			"skipped":       skipped,
			"results":       ciResults,
		}
		jsonBytes, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonBytes))
		return
	}

	printer := terminal.NewPrinter()
	printer.Section("Multi-Target Summary")

	printer.KeyValue("Total Targets", fmt.Sprintf("%d", totalTargets))
	printer.KeyValueColored("Succeeded", fmt.Sprintf("%d", succeeded), terminal.Green)
	if failed > 0 {
		printer.KeyValueColored("Failed", fmt.Sprintf("%d", failed), terminal.Red)
	}
	if skipped > 0 {
		printer.KeyValueColored("Skipped", fmt.Sprintf("%d", skipped), terminal.Yellow)
	}

	fmt.Println()
}

// parseParams parses key=value parameter flags
func parseParams(flags []string) map[string]string {
	params := make(map[string]string)
	for _, flag := range flags {
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}
	return params
}

// loadParamsFromFile reads parameters from a JSON or YAML file
func loadParamsFromFile(path string) (map[string]string, error) {
	log := logger.Get()

	log.Debug("Loading params from file", zap.String("path", path))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read params file: %w", err)
	}

	// Detect format by extension
	ext := strings.ToLower(filepath.Ext(path))

	var rawParams map[string]interface{}

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &rawParams); err != nil {
			return nil, fmt.Errorf("failed to parse JSON params file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &rawParams); err != nil {
			return nil, fmt.Errorf("failed to parse YAML params file: %w", err)
		}
	default:
		// Try JSON first, then YAML
		if err := json.Unmarshal(data, &rawParams); err != nil {
			if err := yaml.Unmarshal(data, &rawParams); err != nil {
				return nil, fmt.Errorf("failed to parse params file (tried JSON and YAML): %w", err)
			}
		}
	}

	// Convert all values to strings
	params := make(map[string]string)
	for k, v := range rawParams {
		switch val := v.(type) {
		case string:
			params[k] = val
		case int, int64, float64:
			params[k] = fmt.Sprintf("%v", val)
		case bool:
			params[k] = fmt.Sprintf("%v", val)
		default:
			// For complex types, marshal back to JSON string
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				params[k] = fmt.Sprintf("%v", val)
			} else {
				params[k] = string(jsonBytes)
			}
		}
	}

	log.Debug("Params loaded from file",
		zap.String("path", path),
		zap.Int("count", len(params)),
	)

	return params, nil
}

// printResultSummary prints a summary of the workflow result
func printResultSummary(result *core.WorkflowResult) {
	// Calculate step counts from Steps slice
	var completedSteps, failedSteps, skippedSteps int
	for _, step := range result.Steps {
		switch step.Status {
		case core.StepStatusSuccess:
			completedSteps++
		case core.StepStatusFailed:
			failedSteps++
		case core.StepStatusSkipped:
			skippedSteps++
		}
	}

	// Log execution summary to state execution log file
	logger.Get().Info("Execution Summary",
		zap.String("workflow", result.WorkflowName),
		zap.String("run_uuid", result.RunUUID),
		zap.String("target", result.Target),
		zap.String("status", string(result.Status)),
		zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
		zap.Int("total_steps", len(result.Steps)),
		zap.Int("completed_steps", completedSteps),
		zap.Int("failed_steps", failedSteps),
		zap.Int("skipped_steps", skippedSteps),
	)

	// CI mode: output JSON
	if ciOutputFormat {
		output := formatResultForCI(result)
		jsonBytes, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonBytes))
		return
	}

	printer := terminal.NewPrinter()

	printer.Section("Execution Summary")
	printer.KeyValue("Workflow", result.WorkflowName)
	printer.KeyValue("Run ID", result.RunUUID)
	printer.KeyValue("Target", result.Target)
	printer.KeyValue("Status", terminal.StatusBadge(string(result.Status)))
	printer.KeyValue("Duration", formatDuration(result.EndTime.Sub(result.StartTime)))

	if len(result.Steps) > 0 {
		fmt.Println()
		fmt.Println(terminal.ResultSymbol() + " " + terminal.Bold("Step Results:"))
		table := terminal.NewTable(os.Stdout, []string{"Status", "Step", "Duration"})

		for _, step := range result.Steps {
			table.Append([]string{
				terminal.StepSymbol(string(step.Status)),
				step.StepName,
				formatDuration(step.Duration),
			})
		}
		table.Render()
	}

	if len(result.Artifacts) > 0 {
		fmt.Println()
		fmt.Println(terminal.ListSymbol() + " " + terminal.Bold("Artifacts:"))
		for _, artifact := range result.Artifacts {
			printer.Bullet(artifact)
		}
	}

	fmt.Println()
}

// formatResultForCI formats a workflow result for CI JSON output
func formatResultForCI(result *core.WorkflowResult) map[string]interface{} {
	steps := make([]map[string]interface{}, 0, len(result.Steps))
	for _, step := range result.Steps {
		steps = append(steps, map[string]interface{}{
			"name":     step.StepName,
			"status":   string(step.Status),
			"duration": formatDuration(step.Duration),
		})
	}

	return map[string]interface{}{
		"workflow":   result.WorkflowName,
		"run_uuid":   result.RunUUID,
		"target":     result.Target,
		"status":     string(result.Status),
		"duration":   formatDuration(result.EndTime.Sub(result.StartTime)),
		"start_time": result.StartTime.Format(time.RFC3339),
		"end_time":   result.EndTime.Format(time.RFC3339),
		"steps":      steps,
		"artifacts":  result.Artifacts,
	}
}

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm %ds", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}

// parseRunDuration parses duration strings like "30s", "2h", "1d"
// Extends time.ParseDuration to support days (d)
func parseRunDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}

	// Handle days specially (not supported by time.ParseDuration)
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}

// generateEmptyTarget creates a placeholder target name for --empty-target mode
func generateEmptyTarget() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	random := make([]byte, 6)
	for i := range random {
		random[i] = chars[rand.Intn(len(chars))]
	}
	return fmt.Sprintf("empty-%s-%d", string(random), time.Now().Unix())
}

// readWorkflowFromStdin reads and parses a workflow YAML from stdin
func readWorkflowFromStdin() (*core.Workflow, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat stdin: %w", err)
	}

	// Check if stdin has piped data
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("no data piped to stdin")
	}

	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	if len(content) == 0 {
		return nil, fmt.Errorf("stdin is empty")
	}

	workflow, err := parser.ParseContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	// Validate the workflow
	p := parser.NewParser()
	if err := p.Validate(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	return workflow, nil
}

// fetchWorkflowFromURL fetches workflow YAML from a URL with GitHub auth fallback
func fetchWorkflowFromURL(urlStr string) (*core.Workflow, error) {
	log := logger.Get()

	// Validate URL format
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("URL must use http or https scheme")
	}

	log.Debug("Fetching workflow from URL", zap.String("url", urlStr))

	// First attempt: fetch without auth
	content, err := fetchURLContent(urlStr, nil)
	if err == nil {
		// Success without auth
		return parseAndValidateWorkflow(content)
	}

	// If failed and is a GitHub URL, retry with auth
	if isGitHubURLForFetch(urlStr) {
		token := getGitHubTokenForFetch()
		if token != "" {
			log.Debug("Retrying with GitHub authentication")

			// Transform URL to GitHub API format for private repos
			apiURL := transformToGitHubAPIURL(urlStr)
			headers := map[string]string{
				"Authorization": "Bearer " + token,
				"Accept":        "application/vnd.github.v3.raw",
			}

			content, err = fetchURLContent(apiURL, headers)
			if err == nil {
				return parseAndValidateWorkflow(content)
			}
		}
	}

	return nil, fmt.Errorf("failed to fetch workflow: %w", err)
}

// fetchURLContent fetches content from a URL with optional headers
func fetchURLContent(urlStr string, headers map[string]string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", core.DefaultUA)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return content, nil
}

// isGitHubURLForFetch checks if the URL is a GitHub URL
func isGitHubURLForFetch(urlStr string) bool {
	return strings.Contains(urlStr, "github.com") ||
		strings.Contains(urlStr, "raw.githubusercontent.com") ||
		strings.Contains(urlStr, "api.github.com")
}

// getGitHubTokenForFetch returns the GitHub token from settings or environment
// Priority: GITHUB_API_KEY (from settings) > GH_TOKEN (from OS env)
func getGitHubTokenForFetch() string {
	// First: try GITHUB_API_KEY from settings (exported to env by root.go)
	if token := os.Getenv("GITHUB_API_KEY"); token != "" {
		return token
	}
	// Fallback: GH_TOKEN from OS environment (used by GitHub CLI)
	if token := os.Getenv("GH_TOKEN"); token != "" {
		return token
	}
	return ""
}

// transformToGitHubAPIURL transforms GitHub URLs to API format for private repo access
// Supports:
//   - https://github.com/owner/repo/blob/branch/path/file.yaml
//   - https://raw.githubusercontent.com/owner/repo/branch/path/file.yaml
//   - https://api.github.com/... (unchanged)
func transformToGitHubAPIURL(urlStr string) string {
	// Already an API URL, return as-is
	if strings.Contains(urlStr, "api.github.com") {
		return urlStr
	}

	// Pattern: https://github.com/owner/repo/blob/branch/path/to/file.yaml
	githubBlobPattern := regexp.MustCompile(`^https://github\.com/([^/]+)/([^/]+)/blob/([^/]+)/(.+)$`)
	if matches := githubBlobPattern.FindStringSubmatch(urlStr); matches != nil {
		owner, repo, branch, path := matches[1], matches[2], matches[3], matches[4]
		return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch)
	}

	// Pattern: https://raw.githubusercontent.com/owner/repo/branch/path/to/file.yaml
	rawGitHubPattern := regexp.MustCompile(`^https://raw\.githubusercontent\.com/([^/]+)/([^/]+)/([^/]+)/(.+)$`)
	if matches := rawGitHubPattern.FindStringSubmatch(urlStr); matches != nil {
		owner, repo, branch, path := matches[1], matches[2], matches[3], matches[4]
		return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch)
	}

	// Not a recognized format, return original
	return urlStr
}

// parseAndValidateWorkflow parses and validates workflow content
func parseAndValidateWorkflow(content []byte) (*core.Workflow, error) {
	if len(content) == 0 {
		return nil, fmt.Errorf("workflow content is empty")
	}

	workflow, err := parser.ParseContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	// Validate the workflow
	p := parser.NewParser()
	if err := p.Validate(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	return workflow, nil
}

// executeSingleWorkflowDirect executes a pre-loaded workflow against all targets
// loader can be nil for module workflows loaded from stdin (flows not supported for stdin)
func executeSingleWorkflowDirect(ctx context.Context, workflow *core.Workflow, allTargets []string, cfg *config.Config, printer *terminal.Printer, log *zap.Logger, loader *parser.Loader) error {
	// Show target count and concurrency
	if len(allTargets) > 1 {
		printer.Info("Running against %d targets (concurrency: %d)", len(allTargets), concurrency)
	}

	results, lastErr := executeRunsConcurrentlyWithContext(ctx, workflow, allTargets, cfg, concurrency, loader)

	// Print summary for multiple targets
	if len(allTargets) > 1 {
		printMultiTargetSummary(results, len(allTargets))
	}

	return lastErr
}

// runDistributedRun submits run tasks to the distributed worker queue
func runDistributedRun(cfg *config.Config, allTargets []string, printer *terminal.Printer) error {
	// Override Redis config from URL if provided
	if redisURLRun != "" {
		redisCfg, err := distributed.ParseRedisURL(redisURLRun)
		if err != nil {
			return fmt.Errorf("invalid redis URL: %w", err)
		}
		cfg.Redis = *redisCfg
	}

	// Check Redis is configured
	if !cfg.IsRedisConfigured() {
		return fmt.Errorf("redis not configured. Add redis section to osm-settings.yaml or use --redis-url")
	}

	// Determine workflow name and kind
	workflowName := flowName
	workflowKind := "flow"
	if workflowName == "" && len(moduleNames) > 0 {
		workflowName = moduleNames[0] // Use first module for distributed run
		workflowKind = "module"
	}

	if workflowName == "" {
		return fmt.Errorf("workflow name required (use -f or -m)")
	}

	// Create master client to submit tasks
	master, err := distributed.NewMaster(cfg)
	if err != nil {
		return fmt.Errorf("failed to create master client: %w", err)
	}

	ctx := context.Background()

	// Parse additional params
	params := make(map[string]interface{})
	for _, flag := range paramFlags {
		parts := strings.SplitN(flag, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}

	// Submit task for each target
	printer.Section("Submitting Distributed Tasks")

	var taskIDs []string
	for _, target := range allTargets {
		task := &distributed.Task{
			WorkflowName: workflowName,
			WorkflowKind: workflowKind,
			Target:       target,
			Params:       params,
		}

		if err := master.SubmitTask(ctx, task); err != nil {
			printer.Error("Failed to submit task for %s: %s", target, err)
			continue
		}

		taskIDs = append(taskIDs, task.ID)
		printer.Success("Submitted task %s for target: %s", task.ID, target)
	}

	// Print summary
	fmt.Println()
	printer.Info("Submitted %d tasks to the distributed queue", len(taskIDs))
	printer.Info("Use 'osmedeus worker status' to check worker availability")
	printer.Info("Tasks will be processed by available workers")

	return nil
}

// registerCronTriggersWithServer registers workflow cron triggers with the server.
// Best-effort: failures are logged but don't block execution.
func registerCronTriggersWithServer(ctx context.Context, workflow *core.Workflow, target string, params map[string]string, cfg *config.Config, printer *terminal.Printer, log *zap.Logger) {
	cronTriggers := workflow.GetCronTriggers()
	if len(cronTriggers) == 0 {
		return
	}

	// Determine server URL (CLI flag takes precedence over config)
	url := serverURL
	if url == "" {
		url = cfg.Server.GetServerURL()
	}
	if url == "" {
		log.Debug("No server URL configured, skipping cron trigger registration")
		return
	}

	client := NewScheduleClient(cfg)
	client.SetBaseURL(url)

	if !client.IsServerAvailable() {
		log.Debug("Server not available, skipping cron trigger registration",
			zap.String("url", url),
		)
		return
	}

	registered := 0
	for _, trigger := range cronTriggers {
		if err := client.RegisterCronTrigger(ctx, workflow, &trigger, target, params); err != nil {
			log.Warn("Failed to register cron trigger",
				zap.String("trigger", trigger.Name),
				zap.String("schedule", trigger.Schedule),
				zap.Error(err),
			)
		} else {
			registered++
			log.Debug("Registered cron trigger",
				zap.String("trigger", trigger.Name),
				zap.String("schedule", trigger.Schedule),
			)
		}
	}

	if registered > 0 && !silent {
		printer.Info("Registered %d cron trigger(s) with server at %s", registered, url)
	}
}

// ensureExternalBinariesInPath adds the external-binaries folder to PATH if it exists
// and is not already present. This ensures installed tools are available even if
// the user hasn't reloaded their shell after running `osmedeus install binary`.
func ensureExternalBinariesInPath(cfg *config.Config) {
	if cfg.BinariesPath == "" {
		return
	}

	// Check if directory exists
	if _, err := os.Stat(cfg.BinariesPath); os.IsNotExist(err) {
		return
	}

	// Get current PATH
	currentPath := os.Getenv("PATH")

	// Check if already in PATH
	pathSep := string(os.PathListSeparator)
	paths := strings.Split(currentPath, pathSep)
	for _, p := range paths {
		if p == cfg.BinariesPath {
			return // Already in PATH
		}
	}

	// Prepend external-binaries to PATH
	newPath := cfg.BinariesPath + pathSep + currentPath
	_ = os.Setenv("PATH", newPath)

	log := logger.Get()
	log.Debug("Added external-binaries to PATH", zap.String("path", cfg.BinariesPath))
}
