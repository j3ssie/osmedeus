package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/console"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/heuristics"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/metrics"
	"github.com/j3ssie/osmedeus/v5/internal/notify"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/runner"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/j3ssie/osmedeus/v5/internal/utils"
	"go.uber.org/zap"
)

// TargetTypeMismatchError represents a target type validation error.
// This error is used to provide a single, nicely formatted error message
// when target type validation fails.
type TargetTypeMismatchError struct {
	Supplied     string
	ExpectedType string
}

func (e *TargetTypeMismatchError) Error() string {
	return fmt.Sprintf("target type mismatch: supplied %s, expected type %s", e.Supplied, e.ExpectedType)
}

// StepCompletedCallback is called after each step completes
type StepCompletedCallback func(ctx context.Context, runID string)

// Executor is the main workflow executor
type Executor struct {
	templateEngine        *template.Engine
	functionRegistry      *functions.Registry
	stepDispatcher        *StepDispatcher
	logger                *zap.Logger
	printer               *terminal.Printer
	dryRun                bool
	showSpinner           bool
	verbose               bool // show actual step output instead of debug logs
	silent                bool // suppress step output (only show with explicit --verbose)
	isSchedulerInvocation bool // true when invoked by scheduler (cron/event/watch), allows bypassing manual trigger check
	serverMode            bool // true when invoked via server API, enables file logging
	progressBar           *terminal.ProgressBar
	disableWorkflowState  bool                       // disable writing workflow YAML to output directory
	skipWorkspace         bool                       // skip creating workspace/output directory (for empty-target mode)
	skipValidation        bool                       // skip target type validation from dependencies.variables
	dbRunUUID             string                     // database run UUID for tracking progress
	dbRunID               int64                      // database run ID for step result foreign keys
	onStepCompleted       StepCompletedCallback      // callback after each step completes
	loader                *parser.Loader             // workflow loader for loading nested modules in flows
	consoleCapture        *console.Capture           // console output capture for run-console.log
	writeCoordinator      *database.WriteCoordinator // unified coordinator for batch database writes
}

// NewExecutor creates a new workflow executor
func NewExecutor() *Executor {
	return &Executor{
		templateEngine:   template.NewEngine(),
		functionRegistry: functions.NewRegistry(),
		stepDispatcher:   NewStepDispatcher(),
		logger:           logger.Get(),
		printer:          terminal.NewPrinter(),
		showSpinner:      false, // Disabled by default, enabled with --spinner
	}
}

func BuildBuiltinVariables(cfg *config.Config, params map[string]string) map[string]interface{} {
	if cfg == nil {
		cfg = config.DefaultConfig()
		cfg.ResolvePaths()
	}

	exec := NewExecutor()
	runUUID := uuid.New().String()
	execCtx := core.NewExecutionContext("func-eval", core.KindModule, runUUID, params["target"])
	exec.injectBuiltinVariables(cfg, params, execCtx)

	// Add temp directory for eval (cleanup is handled by the caller)
	tempDir, err := os.MkdirTemp("", "osm-tmp-")
	if err == nil {
		execCtx.SetVariable("TempDir", tempDir)
		execCtx.SetVariable("TempFile", filepath.Join(tempDir, "osm-tmp-file"))
	}

	return execCtx.GetVariables()
}

// SetSpinner enables or disables spinner display
func (e *Executor) SetSpinner(show bool) {
	e.showSpinner = show
}

// SetDryRun enables or disables dry-run mode
func (e *Executor) SetDryRun(dryRun bool) {
	e.dryRun = dryRun
	e.stepDispatcher.SetDryRun(dryRun)
}

// SetVerbose enables or disables verbose output (shows step stdout)
func (e *Executor) SetVerbose(v bool) {
	e.verbose = v
}

// SetSilent enables or disables silent mode (hides step output)
func (e *Executor) SetSilent(s bool) {
	e.silent = s
}

// SetSchedulerInvocation marks this execution as triggered by the scheduler
// When true, the manual trigger check is bypassed
func (e *Executor) SetSchedulerInvocation(v bool) {
	e.isSchedulerInvocation = v
}

// SetServerMode enables server mode which adds file logging to the workspace
func (e *Executor) SetServerMode(enabled bool) {
	e.serverMode = enabled
}

// SetDisableWorkflowState enables or disables workflow state file export
func (e *Executor) SetDisableWorkflowState(disable bool) {
	e.disableWorkflowState = disable
}

// SetSkipWorkspace enables or disables workspace folder creation
func (e *Executor) SetSkipWorkspace(skip bool) {
	e.skipWorkspace = skip
}

// SetSkipValidation enables or disables target type validation
func (e *Executor) SetSkipValidation(skip bool) {
	e.skipValidation = skip
}

// SetProgressBar sets the progress bar for execution display
func (e *Executor) SetProgressBar(pb *terminal.ProgressBar) {
	e.progressBar = pb
}

// SetDBRunUUID sets the database run UUID for progress tracking
func (e *Executor) SetDBRunUUID(runUUID string) {
	e.dbRunUUID = runUUID
}

// SetDBRunID sets the database run ID for step result foreign keys
func (e *Executor) SetDBRunID(runID int64) {
	e.dbRunID = runID
	// Initialize write coordinator with both RunID and RunUUID
	if e.dbRunUUID != "" {
		e.writeCoordinator = database.NewWriteCoordinator(runID, e.dbRunUUID, nil)
	}
}

// SetOnStepCompleted sets the callback for step completion
func (e *Executor) SetOnStepCompleted(callback StepCompletedCallback) {
	e.onStepCompleted = callback
}

// writeVerboseOutputToLog writes step output directly to console log file
// This is used when terminal output is suppressed (silent mode or progress bar)
// but we still want the log file to have complete verbose output with colors
func (e *Executor) writeVerboseOutputToLog(output string) {
	if e.consoleCapture == nil || output == "" {
		return
	}
	// Trim whitespace from output
	output = strings.TrimSpace(output)
	if output == "" {
		return // Don't print [output] if nothing to show
	}
	// Format similar to printer.VerboseOutput but write directly to file
	var sb strings.Builder
	sb.WriteString("  ")
	sb.WriteString(terminal.Gray("[output]"))
	sb.WriteString("\n")
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Skip blank lines
		if strings.TrimSpace(line) == "" {
			continue
		}
		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	e.consoleCapture.WriteToFile(sb.String())
}

// SetLoader sets the workflow loader for loading nested modules in flows
func (e *Executor) SetLoader(l *parser.Loader) {
	e.loader = l
}

// injectBuiltinVariables adds all builtin variables to the execution context
func (e *Executor) injectBuiltinVariables(cfg *config.Config, params map[string]string, execCtx *core.ExecutionContext) {
	now := time.Now()

	// Settings-based variables
	execCtx.SetVariable("BaseFolder", cfg.BaseFolder)
	execCtx.SetVariable("Binaries", cfg.BinariesPath)
	execCtx.SetVariable("Data", cfg.DataPath)
	execCtx.SetVariable("ExternalData", cfg.DataPath) // Alias for Data
	execCtx.SetVariable("ExternalConfigs", cfg.ConfigsPath)
	execCtx.SetVariable("ExternalAgentConfigs", cfg.ExternalAgentConfigsPath)
	execCtx.SetVariable("ExternalScripts", cfg.ExternalScriptsPath)
	execCtx.SetVariable("Workflows", cfg.WorkflowsPath)
	execCtx.SetVariable("MarkdownTemplates", cfg.MarkdownReportTemplatesPath)
	execCtx.SetVariable("ExternalMarkdowns", cfg.MarkdownReportTemplatesPath) // Alias for MarkdownTemplates
	execCtx.SetVariable("ExternalAgents", cfg.ExternalAgentConfigsPath)       // Alias for ExternalAgentConfigs
	execCtx.SetVariable("SnapshotsFolder", cfg.SnapshotPath)

	// Workspaces - can be overridden by -W flag
	workspacesPath := cfg.WorkspacesPath
	if workspacesFolder := params["workspaces_folder"]; workspacesFolder != "" {
		workspacesPath = workspacesFolder
	}
	execCtx.SetVariable("Workspaces", workspacesPath)

	// Thread settings based on tactic
	tactic := params["tactic"]
	if tactic == "" {
		tactic = "default"
	}
	threads, baseThreads := cfg.GetThreads(tactic)

	// Override with threads_hold if specified
	if threadsHold := params["threads_hold"]; threadsHold != "" && threadsHold != "0" {
		if customThreads, err := strconv.Atoi(threadsHold); err == nil && customThreads > 0 {
			threads = customThreads
			baseThreads = customThreads / 2
			if baseThreads < 1 {
				baseThreads = 1
			}
		}
	}
	execCtx.SetVariable("threads", threads)
	execCtx.SetVariable("baseThreads", baseThreads)

	// Version info
	execCtx.SetVariable("Version", core.VERSION)

	// Platform detection variables
	execCtx.SetVariable("PlatformOS", runtime.GOOS)
	execCtx.SetVariable("PlatformArch", runtime.GOARCH)
	execCtx.SetVariable("PlatformInDocker", DetectDocker())
	execCtx.SetVariable("PlatformInKubernetes", DetectKubernetes())
	execCtx.SetVariable("PlatformCloudProvider", DetectCloudProvider())

	// Target-based variables
	target := params["target"]
	targetFileParam := params["target_file"]
	execCtx.SetVariable("Target", target)
	execCtx.SetVariable("TargetFile", targetFileParam)

	// Default TargetSpace (can be overridden by heuristics or -S flag)
	targetSpace := sanitizeTargetSpace(target)
	execCtx.SetVariable("TargetSpace", targetSpace)

	// Heuristics check level
	heuristicsLevel := params["heuristics_check"]
	if heuristicsLevel == "" {
		heuristicsLevel = "basic"
	}
	execCtx.SetVariable("HeuristicsCheck", heuristicsLevel)

	// Run heuristics analysis if not "none"
	if heuristicsLevel != "none" {
		info, err := heuristics.Analyze(target, heuristicsLevel)
		if err == nil && info != nil {
			injectHeuristicVariables(info, execCtx)
			// Update targetSpace if heuristics detected a root domain
			if info.RootDomain != "" {
				targetSpace = info.RootDomain
				execCtx.SetVariable("TargetSpace", targetSpace)
			}
		}
	}

	// Override TargetSpace if -S flag provided (takes precedence)
	if spaceName := params["space_name"]; spaceName != "" {
		targetSpace = spaceName
		execCtx.SetVariable("TargetSpace", targetSpace)
	}

	// Set WorkspaceName for database operations
	execCtx.WorkspaceName = targetSpace
	execCtx.SetVariable("Workspace", targetSpace)

	// Output path uses final targetSpace
	output := filepath.Join(workspacesPath, targetSpace)
	execCtx.SetVariable("Output", output)

	// State files for tracking execution
	execCtx.SetVariable("StateExecutionLog", filepath.Join(output, "run-execution.log"))
	execCtx.SetVariable("StateConsoleLog", filepath.Join(output, "run-console.log"))
	execCtx.SetVariable("StateCompletedFile", filepath.Join(output, "run-completed.json"))
	execCtx.SetVariable("StateFile", filepath.Join(output, "run-state.json"))
	execCtx.SetVariable("StateWorkflowFile", filepath.Join(output, "run-workflow.yaml"))
	execCtx.SetVariable("StateWorkflowFolder", filepath.Join(output, "run-modules"))

	// Module/Workflow name variable
	execCtx.SetVariable("ModuleName", execCtx.WorkflowName)

	// Flow name variable (set when running under a flow, empty for standalone module)
	flowName := ""
	if fn, ok := params["flow_name"]; ok && fn != "" {
		flowName = fn
	}
	execCtx.SetVariable("FlowName", flowName)

	// Auto-generated variables
	execCtx.SetVariable("TaskDate", now.Format("2006-01-02"))
	execCtx.SetVariable("RunUUID", execCtx.RunUUID)
	execCtx.SetVariable("DBRunID", e.dbRunID) // Integer Run.ID for database foreign keys
	execCtx.SetVariable("TimeStamp", fmt.Sprintf("%d", now.Unix()))
	execCtx.SetVariable("CurrentTime", now.Format("2006-01-02T15:04:05"))
	execCtx.SetVariable("Today", now.Format("2006-01-02"))
	execCtx.SetVariable("RandomString", generateRandomString(6))

	// Chunk-related variables (when running in chunk mode)
	if v, ok := params["chunk_index"]; ok && v != "" {
		if chunkIndex, err := strconv.Atoi(v); err == nil {
			execCtx.SetVariable("ChunkIndex", chunkIndex)
		}
	}
	if v, ok := params["chunk_size"]; ok && v != "" {
		if chunkSize, err := strconv.Atoi(v); err == nil {
			execCtx.SetVariable("ChunkSize", chunkSize)
		}
	}
	if v, ok := params["total_chunks"]; ok && v != "" {
		if totalChunks, err := strconv.Atoi(v); err == nil {
			execCtx.SetVariable("TotalChunks", totalChunks)
		}
	}
	if v, ok := params["chunk_start"]; ok && v != "" {
		if chunkStart, err := strconv.Atoi(v); err == nil {
			execCtx.SetVariable("ChunkStart", chunkStart)
		}
	}
	if v, ok := params["chunk_end"]; ok && v != "" {
		if chunkEnd, err := strconv.Atoi(v); err == nil {
			execCtx.SetVariable("ChunkEnd", chunkEnd)
		}
	}

	// Event envelope (only for event-triggered workflows)
	if eventEnvelope, ok := params["_event_envelope"]; ok && eventEnvelope != "" {
		execCtx.SetVariable("EventEnvelope", eventEnvelope)

		// Parse envelope and extract convenience variables
		var envelope map[string]interface{}
		if err := json.Unmarshal([]byte(eventEnvelope), &envelope); err == nil {
			if topic, ok := envelope["topic"].(string); ok {
				execCtx.SetVariable("EventTopic", topic)
			}
			if source, ok := envelope["source"].(string); ok {
				execCtx.SetVariable("EventSource", source)
			}
			if dataType, ok := envelope["data_type"].(string); ok {
				execCtx.SetVariable("EventDataType", dataType)
			}
			if timestamp, ok := envelope["timestamp"].(string); ok {
				execCtx.SetVariable("EventTimestamp", timestamp)
			}
			// EventData is the parsed data object as JSON string
			if data := envelope["data"]; data != nil {
				dataJSON, err := json.Marshal(data)
				if err == nil {
					execCtx.SetVariable("EventData", string(dataJSON))
				}
			}
		}
	}
}

// setupTempDirectory creates a temporary directory and file for workflow execution
// Returns a cleanup function that removes the directory when called
func (e *Executor) setupTempDirectory(execCtx *core.ExecutionContext) (cleanup func()) {
	tempDir, err := os.MkdirTemp("", "osm-tmp-")
	if err != nil {
		e.logger.Warn("Failed to create temp directory", zap.Error(err))
		return func() {} // no-op cleanup
	}

	execCtx.SetVariable("TempDir", tempDir)

	// Create temp file path within temp dir
	tempFile := filepath.Join(tempDir, "osm-tmp-file")
	execCtx.SetVariable("TempFile", tempFile)

	return func() {
		if err := os.RemoveAll(tempDir); err != nil {
			e.logger.Warn("Failed to cleanup temp directory",
				zap.String("path", tempDir),
				zap.Error(err))
		}
	}
}

func (e *Executor) debugLogTargetVariables(execCtx *core.ExecutionContext) {
	if execCtx == nil || execCtx.Logger == nil {
		return
	}

	getStr := func(key string) string {
		v, ok := execCtx.GetVariable(key)
		if !ok || v == nil {
			return ""
		}
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}

	getInt := func(key string) int {
		v, ok := execCtx.GetVariable(key)
		if !ok || v == nil {
			return 0
		}
		if i, ok := v.(int); ok {
			return i
		}
		return 0
	}

	execCtx.Logger.Debug("Target variables",
		zap.String("workflow", execCtx.WorkflowName),
		zap.String("run_uuid", execCtx.RunUUID),
		zap.String("Target", getStr("Target")),
		zap.String("TargetSpace", getStr("TargetSpace")),
		zap.String("Output", getStr("Output")),
		zap.String("StateExecutionLog", getStr("StateExecutionLog")),
		zap.String("StateFile", getStr("StateFile")),
		zap.Int("threads", getInt("threads")),
		zap.Int("baseThreads", getInt("baseThreads")),
	)
}

// injectHeuristicVariables adds heuristic-detected variables to the execution context
func injectHeuristicVariables(info *heuristics.TargetInfo, execCtx *core.ExecutionContext) {
	if info == nil {
		return
	}

	// Set target type
	execCtx.SetVariable("TargetType", string(info.Type))

	if info.Type == heuristics.TargetTypeURL {
		execCtx.SetVariable("TargetBaseURL", info.BaseURL)
		execCtx.SetVariable("TargetRootURL", info.RootURL)
		execCtx.SetVariable("TargetHostname", info.Hostname)
		execCtx.SetVariable("TargetRootDomain", info.RootDomain)
		execCtx.SetVariable("TargetTLD", info.TLD)
		execCtx.SetVariable("TargetSLD", info.SLD)
		execCtx.SetVariable("Org", info.SLD) // Alias for TargetSLD
		execCtx.SetVariable("TargetHost", info.Host)
		execCtx.SetVariable("TargetPort", info.Port)
		execCtx.SetVariable("TargetPath", info.Path)
		execCtx.SetVariable("TargetFileExt", info.File)
		execCtx.SetVariable("TargetScheme", info.Scheme)

		// Advanced: HTTP status and content length
		if info.StatusCode > 0 {
			execCtx.SetVariable("TargetStatusCode", info.StatusCode)
			execCtx.SetVariable("TargetContentLength", info.ContentLength)
		}
	}

	if info.Type == heuristics.TargetTypeDomain {
		execCtx.SetVariable("TargetRootDomain", info.RootDomain)
		execCtx.SetVariable("TargetTLD", info.TLD)
		execCtx.SetVariable("TargetSLD", info.SLD)
		execCtx.SetVariable("Org", info.SLD) // Alias for TargetSLD
		execCtx.SetVariable("TargetIsWildcard", info.IsWildcard)

		// Advanced: resolved IP
		if info.ResolvedIP != "" {
			execCtx.SetVariable("TargetResolvedIP", info.ResolvedIP)
		}
	}

	if info.Type == heuristics.TargetTypeIP {
		execCtx.SetVariable("TargetRootDomain", info.Original)
	}
}

// sanitizeTargetSpace creates a filesystem-safe target path
func sanitizeTargetSpace(target string) string {
	// Replace unsafe characters
	sanitized := strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) {
			return '_'
		}
		return r
	}, target)

	// If too long (>30), truncate with hash
	if len(sanitized) > 30 {
		timestamp := time.Now().Unix()
		randomPart := generateRandomString(6)
		return fmt.Sprintf("%s-%s-%d", sanitized[:6], randomPart, timestamp)
	}

	return sanitized
}

// generateRandomString generates a random alphanumeric string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// parseExcludeList parses comma-separated exclude modules into a slice
func parseExcludeList(excludeModules string) []string {
	if excludeModules == "" {
		return nil
	}
	var result []string
	for _, m := range strings.Split(excludeModules, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			result = append(result, m)
		}
	}
	return result
}

// isModuleExcluded checks if a module name is in the exclude list
func isModuleExcluded(moduleName string, excludeList []string) bool {
	for _, excluded := range excludeList {
		if excluded == moduleName {
			return true
		}
	}
	return false
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

// printDryRunHeader prints a formatted header for dry-run mode
func printDryRunHeader(workflowName, workflowKind, target, tactic string, stepCount int, execCtx *core.ExecutionContext) {
	separator := strings.Repeat("═", 52)

	fmt.Println()
	fmt.Printf("%s %s %s\n", terminal.Yellow("⚠"), terminal.BoldYellow("DRY-RUN Mode"), terminal.Gray("- No commands will be executed"))
	fmt.Println(terminal.Yellow(separator))

	// Workflow info section
	fmt.Printf("%s %s\n", terminal.Cyan("◆"), terminal.Bold("Workflow Info"))
	fmt.Printf("  %s Workflow:    %s\n", terminal.Gray("│"), terminal.Cyan(workflowName))
	fmt.Printf("  %s Kind:        %s\n", terminal.Gray("│"), terminal.Gray(workflowKind))
	fmt.Printf("  %s Target:      %s\n", terminal.Gray("│"), terminal.Green(target))
	fmt.Printf("  %s Steps:       %s\n", terminal.Gray("│"), terminal.Yellow(fmt.Sprintf("%d", stepCount)))
	fmt.Printf("  %s Tactic:      %s\n", terminal.Gray("╰"), terminal.Gray(tactic))
	fmt.Println()

	// Helper to get string value from context
	getStr := func(key string) string {
		if v, ok := execCtx.GetVariable(key); ok {
			switch val := v.(type) {
			case string:
				return val
			case int:
				return fmt.Sprintf("%d", val)
			default:
				return fmt.Sprintf("%v", val)
			}
		}
		return ""
	}

	// Builtin Variables section
	fmt.Printf("%s %s\n", terminal.Cyan("✦"), terminal.Bold("Builtin Variables"))
	fmt.Printf("  %s BaseFolder:  %s\n", terminal.Gray("│"), terminal.Gray(getStr("BaseFolder")))
	fmt.Printf("  %s Binaries:    %s\n", terminal.Gray("│"), terminal.Gray(getStr("Binaries")))
	fmt.Printf("  %s Data:        %s\n", terminal.Gray("│"), terminal.Gray(getStr("Data")))
	fmt.Printf("  %s Workspaces:  %s\n", terminal.Gray("│"), terminal.Gray(getStr("Workspaces")))
	fmt.Printf("  %s Output:      %s\n", terminal.Gray("│"), terminal.Green(getStr("Output")))
	fmt.Printf("  %s threads:     %s\n", terminal.Gray("│"), terminal.Yellow(getStr("threads")))
	fmt.Printf("  %s baseThreads: %s\n", terminal.Gray("│"), terminal.Yellow(getStr("baseThreads")))
	fmt.Printf("  %s Today:       %s\n", terminal.Gray("╰"), terminal.Gray(getStr("Today")))

	fmt.Println(terminal.Yellow(separator))
	fmt.Println()
}

// getStepCommand extracts the command/script from a step for display
func getStepCommand(step *core.Step) string {
	// Handle foreach steps specially
	if step.Type == core.StepTypeForeach && step.Input != "" {
		parts := []string{fmt.Sprintf("foreach [[%s]] in %s", step.Variable, step.Input)}
		if step.VariablePreProcess != "" {
			parts = append(parts, fmt.Sprintf("pre_process: %s", step.VariablePreProcess))
		}
		// Add inner step name, type, and command
		if step.Step != nil && step.Step.Name != "" {
			innerCmd := getInnerStepCommand(step.Step)
			if innerCmd != "" {
				parts = append(parts, fmt.Sprintf("step: %s (%s) cmd: %s", step.Step.Name, step.Step.Type, innerCmd))
			} else {
				parts = append(parts, fmt.Sprintf("step: %s (%s)", step.Step.Name, step.Step.Type))
			}
		}
		threads, _ := step.Threads.Int()
		if threads <= 0 {
			threads = 1
		}
		parts = append(parts, fmt.Sprintf("threads: %d", threads))
		return strings.Join(parts, " | ")
	}

	if step.Command != "" {
		return step.Command
	}
	if len(step.Commands) > 0 {
		return strings.Join(step.Commands, "\n")
	}
	if step.Function != "" {
		return step.Function
	}
	if len(step.Functions) > 0 {
		return strings.Join(step.Functions, "\n")
	}
	return ""
}

// getStepCommandColored returns a colored command string for foreach steps (for console display)
func getStepCommandColored(step *core.Step) string {
	// Handle foreach steps specially with colors
	if step.Type == core.StepTypeForeach && step.Input != "" {
		var lines []string

		// First line: foreach [[variable]] in source | threads: N
		var firstLineParts []string
		firstLineParts = append(firstLineParts, fmt.Sprintf("foreach %s in %s",
			terminal.Magenta("[["+step.Variable+"]]"),
			terminal.Cyan(step.Input)))

		if step.VariablePreProcess != "" {
			firstLineParts = append(firstLineParts, fmt.Sprintf("pre_process: %s", terminal.Gray(step.VariablePreProcess)))
		}

		threads, _ := step.Threads.Int()
		if threads <= 0 {
			threads = 1
		}
		firstLineParts = append(firstLineParts, fmt.Sprintf("threads: %s", terminal.Yellow(fmt.Sprintf("%d", threads))))
		lines = append(lines, strings.Join(firstLineParts, terminal.Gray(" | ")))

		// Second line: step: name (type)
		if step.Step != nil && step.Step.Name != "" {
			lines = append(lines, fmt.Sprintf("  %s %s %s",
				terminal.Gray("step:"),
				terminal.HiBlue(step.Step.Name),
				terminal.Gray("("+string(step.Step.Type)+")")))

			// Add command/commands/function/functions on separate lines
			innerLines := getInnerStepCommandLines(step.Step)
			lines = append(lines, innerLines...)
		}

		return strings.Join(lines, "\n")
	}

	// For non-foreach steps, return uncolored (will be colored by printer)
	return getStepCommand(step)
}

// getInnerStepCommandLines returns colored lines for inner step commands/functions
func getInnerStepCommandLines(step *core.Step) []string {
	var lines []string

	if step.Command != "" {
		lines = append(lines, fmt.Sprintf("  %s %s",
			terminal.Gray("command:"),
			terminal.HiGreen(step.Command)))
	}
	if len(step.Commands) > 0 {
		lines = append(lines, fmt.Sprintf("  %s", terminal.Gray("commands:")))
		for _, cmd := range step.Commands {
			lines = append(lines, fmt.Sprintf("    %s %s",
				terminal.Gray("-"),
				terminal.HiGreen(cmd)))
		}
	}
	if step.Function != "" {
		lines = append(lines, fmt.Sprintf("  %s %s",
			terminal.Gray("function:"),
			terminal.HiCyan(step.Function)))
	}
	if len(step.Functions) > 0 {
		lines = append(lines, fmt.Sprintf("  %s", terminal.Gray("functions:")))
		for _, fn := range step.Functions {
			lines = append(lines, fmt.Sprintf("    %s %s",
				terminal.Gray("-"),
				terminal.HiCyan(fn)))
		}
	}

	return lines
}

// getInnerStepCommand extracts the command/function from an inner step (for logs)
func getInnerStepCommand(step *core.Step) string {
	if step.Command != "" {
		return step.Command
	}
	if len(step.Commands) > 0 {
		return strings.Join(step.Commands, "\n")
	}
	if step.Function != "" {
		return step.Function
	}
	if len(step.Functions) > 0 {
		return strings.Join(step.Functions, "\n")
	}
	return ""
}

// checkDependencies validates all workflow dependencies including function conditions
func (e *Executor) checkDependencies(deps *core.Dependencies, execCtx *core.ExecutionContext, binariesPath string) error {
	if deps == nil {
		return nil
	}

	// Skip target type validation if flag is set (still check commands and function conditions)
	if !e.skipValidation {
		// Check target type dependencies (target_types field)
		if len(deps.TargetTypes) > 0 {
			matched := false
			var unknown []string
			for _, t := range deps.TargetTypes {
				ok, err := core.MatchesTargetType(execCtx.Target, t)
				if err != nil {
					unknown = append(unknown, string(t))
					continue
				}
				if ok {
					matched = true
					break
				}
			}
			if len(unknown) > 0 {
				return fmt.Errorf("unknown target_types: %s", strings.Join(unknown, ", "))
			}
			if !matched {
				required := make([]string, 0, len(deps.TargetTypes))
				for _, t := range deps.TargetTypes {
					required = append(required, string(t))
				}
				return fmt.Errorf("target '%s' does not match any of required types: %s", execCtx.Target, strings.Join(required, ", "))
			}
		}

		// Check dependencies.variables for Target type validation
		if err := e.validateTargetVariable(deps, execCtx); err != nil {
			return err
		}
	} else {
		e.logger.Debug("Skipping target type validation (--skip-validation)")
	}

	// Check command dependencies
	if len(deps.Commands) > 0 {
		for _, cmd := range deps.Commands {
			// Try to find in external-binaries folder first, then system PATH
			if _, err := utils.LookPathWithBinaries(cmd, binariesPath); err != nil {
				return fmt.Errorf("required command not found: %s", cmd)
			}
		}
	}

	// Check function conditions
	if len(deps.FunctionsConditions) > 0 {
		ctx := execCtx.GetVariables()
		for _, condition := range deps.FunctionsConditions {
			// Render template variables in the condition
			rendered, err := e.templateEngine.Render(condition, ctx)
			if err != nil {
				return fmt.Errorf("function condition '%s' template error: %w", condition, err)
			}
			result, err := e.functionRegistry.EvaluateCondition(rendered, ctx)
			if err != nil {
				return fmt.Errorf("function condition '%s' evaluation error: %w", condition, err)
			}
			if !result {
				return fmt.Errorf("function condition failed: %s", condition)
			}
		}
	}

	return nil
}

// validateTargetVariable checks if the target matches required type from dependencies.variables.
// Supports comma-separated types (e.g., "domain,url") where matching any type is sufficient.
func (e *Executor) validateTargetVariable(deps *core.Dependencies, execCtx *core.ExecutionContext) error {
	if len(deps.Variables) == 0 {
		return nil
	}

	// Find Target variable requirement
	for _, v := range deps.Variables {
		if !strings.EqualFold(v.Name, "Target") {
			continue
		}

		if v.Type == "" {
			continue // No type specified, skip validation
		}

		// Get the target value from execution context
		target := execCtx.Target
		if target == "" {
			return nil // No target to validate
		}

		// Convert VariableType to TargetType and validate
		// Use MatchesAnyTargetType to support comma-separated types
		targetType := core.TargetType(v.Type)
		matches, err := core.MatchesAnyTargetType(target, targetType)
		if err != nil {
			return fmt.Errorf("target type validation error: %w", err)
		}

		if !matches {
			return &TargetTypeMismatchError{
				Supplied:     target,
				ExpectedType: string(v.Type),
			}
		}

		break // Only check first Target variable
	}

	return nil
}

// ExecuteModule executes a module workflow
func (e *Executor) ExecuteModule(ctx context.Context, module *core.Workflow, params map[string]string, cfg *config.Config) (*core.WorkflowResult, error) {
	e.logger.Debug("ExecuteModule called",
		zap.String("module", module.Name),
		zap.Int("step_count", len(module.Steps)),
		zap.Bool("dry_run", e.dryRun),
	)

	if !module.IsModule() {
		return nil, fmt.Errorf("workflow is not a module")
	}

	// Create cancellable context for run control plane support
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register with run control plane if we have a database run UUID
	// This enables API-based cancellation of the run
	if e.dbRunUUID != "" {
		activeRun := GetRunControlPlane().Register(e.dbRunUUID, cancel)
		defer GetRunControlPlane().Unregister(e.dbRunUUID)
		e.logger.Debug("Registered run with control plane",
			zap.String("run_uuid", e.dbRunUUID),
			zap.Time("started_at", activeRun.StartedAt),
		)
	}

	// Check if manual execution is allowed (for CLI invocation)
	if !e.isSchedulerInvocation && !module.IsManualExecutionAllowed() {
		e.logger.Info("Skipping module - manual trigger disabled",
			zap.String("module", module.Name),
		)
		return &core.WorkflowResult{
			WorkflowName: module.Name,
			WorkflowKind: core.KindModule,
			Status:       core.RunStatusSkipped,
			Message:      "Manual trigger is disabled for this module",
		}, nil
	}

	// Create execution context
	runUUID := e.dbRunUUID
	if runUUID == "" {
		runUUID = uuid.New().String() // Full UUID when not in server mode
	}
	e.logger.Debug("Created execution context",
		zap.String("run_uuid", runUUID),
		zap.String("target", params["target"]),
	)
	execCtx := core.NewExecutionContext(module.Name, core.KindModule, runUUID, params["target"])
	execCtx.Logger = logger.WithWorkflow(module.Name, runUUID)

	// Create and setup runner based on workflow configuration
	binaryPath, _ := os.Executable()
	e.logger.Debug("Creating runner",
		zap.String("binary_path", binaryPath),
	)
	r, err := runner.NewRunner(module, binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	e.logger.Debug("Runner created",
		zap.String("type", string(r.Type())),
		zap.Bool("remote", r.IsRemote()),
	)

	// Setup runner (copy binary, start container, establish SSH, etc.)
	e.logger.Debug("Setting up runner")
	if err := r.Setup(ctx); err != nil {
		return nil, fmt.Errorf("runner setup failed: %w", err)
	}
	defer func() { _ = r.Cleanup(ctx) }()

	// Set runner on step dispatcher
	e.stepDispatcher.SetRunner(r)

	// Set PID callbacks for cancellation support
	if e.dbRunUUID != "" {
		runUUID := e.dbRunUUID
		r.SetPIDCallbacks(
			func(pid int) {
				GetRunControlPlane().AddPID(runUUID, pid)
				// Update database with current PID for API visibility
				_ = database.UpdateRunPID(ctx, runUUID, pid)
			},
			func(pid int) {
				GetRunControlPlane().RemovePID(runUUID, pid)
				// Clear PID from database when process ends
				_ = database.ClearRunPID(ctx, runUUID)
			},
		)
	}

	// Set config on step dispatcher for executors that need it (e.g., LLM)
	e.stepDispatcher.SetConfig(cfg)

	// Log runner type if not host
	if r.Type() != core.RunnerTypeHost {
		execCtx.Logger.Info("Using runner",
			zap.String("type", string(r.Type())),
			zap.Bool("remote", r.IsRemote()),
		)
	}

	// Inject builtin variables
	e.logger.Debug("Injecting builtin variables",
		zap.String("target", params["target"]),
		zap.String("tactic", params["tactic"]),
	)
	e.injectBuiltinVariables(cfg, params, execCtx)
	tempCleanup := e.setupTempDirectory(execCtx)
	defer tempCleanup()
	e.debugLogTargetVariables(execCtx)

	if !e.skipWorkspace && !e.dryRun && database.GetDB() != nil {
		output, _ := execCtx.GetVariable("Output")
		stateExecutionLog, _ := execCtx.GetVariable("StateExecutionLog")
		stateCompletedFile, _ := execCtx.GetVariable("StateCompletedFile")
		stateWorkflowFile, _ := execCtx.GetVariable("StateWorkflowFile")
		stateWorkflowFolder, _ := execCtx.GetVariable("StateWorkflowFolder")

		outputStr, _ := output.(string)
		stateExecutionLogStr, _ := stateExecutionLog.(string)
		stateCompletedFileStr, _ := stateCompletedFile.(string)
		stateWorkflowFileStr, _ := stateWorkflowFile.(string)
		stateWorkflowFolderStr, _ := stateWorkflowFolder.(string)

		if execCtx.WorkspaceName != "" {
			_ = database.EnsureWorkspaceRuntime(ctx, execCtx.WorkspaceName, outputStr, execCtx.WorkflowName, stateExecutionLogStr, stateCompletedFileStr, stateWorkflowFileStr, stateWorkflowFolderStr)
		}
	}

	// Add file logging in server mode
	// Setup file logging to {{Output}}/run-execution.log
	if !e.skipWorkspace {
		if logPath, ok := execCtx.GetVariable("StateExecutionLog"); ok {
			if logStr, ok := logPath.(string); ok && logStr != "" {
				combinedLogger, err := logger.WithFileOutput(execCtx.Logger, logStr)
				if err != nil {
					e.logger.Warn("Failed to setup file logging", zap.Error(err))
				} else if combinedLogger != nil {
					execCtx.Logger = combinedLogger
				}
			}
		}
	}

	// Setup console capture to {{Output}}/run-console.log
	// Only create if not already set up by parent (flow) to avoid truncating previous module output
	createdCapture := false
	if !e.skipWorkspace && !e.dryRun && e.consoleCapture == nil {
		if logPath, ok := execCtx.GetVariable("StateConsoleLog"); ok {
			if logStr, ok := logPath.(string); ok && logStr != "" {
				var err error
				e.consoleCapture, err = console.StartCapture(logStr)
				if err != nil {
					e.logger.Warn("Failed to setup console capture", zap.Error(err))
				} else {
					createdCapture = true
				}
			}
		}
	}
	// Only defer cleanup if WE created the capture (not if inherited from flow)
	if createdCapture {
		defer func() {
			_ = e.consoleCapture.Stop()
			e.consoleCapture = nil
		}()
	}

	// Remove run-completed.json from previous run (fresh start)
	if !e.skipWorkspace {
		if completedFile, ok := execCtx.GetVariable("StateCompletedFile"); ok {
			if cfStr, ok := completedFile.(string); ok {
				RemoveRunCompleted(cfStr)
			}
		}
	}

	// Initialize result
	result := &core.WorkflowResult{
		WorkflowName: module.Name,
		WorkflowKind: core.KindModule,
		RunUUID:      runUUID,
		Target:       params["target"],
		Status:       core.RunStatusRunning,
		StartTime:    time.Now(),
		Steps:        make([]*core.StepResult, 0),
		Exports:      make(map[string]interface{}),
	}

	// Record workflow start for metrics
	metrics.RecordWorkflowStart()

	// Initialize parameters
	if err := e.initializeParams(module, params, execCtx); err != nil {
		result.Status = core.RunStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		metrics.RecordWorkflowEnd(module.Name, string(core.KindModule), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
		return result, err
	}

	// Debug log initialized parameters
	execCtx.Logger.Debug("Parameters initialized",
		zap.Int("param_count", len(params)),
		zap.Any("cli_params", params),
		zap.Int("workflow_param_defs", len(module.Params)),
	)

	// Check dependencies (including function conditions)
	if module.Dependencies != nil {
		if err := e.checkDependencies(module.Dependencies, execCtx, cfg.BinariesPath); err != nil {
			// Don't log TargetTypeMismatchError - it will be handled specially in CLI
			if _, ok := err.(*TargetTypeMismatchError); !ok {
				execCtx.Logger.Error("Dependency check failed", zap.Error(err))
			}
			result.Status = core.RunStatusFailed
			result.Error = fmt.Errorf("dependency check failed: %w", err)
			result.EndTime = time.Now()
			metrics.RecordWorkflowEnd(module.Name, string(core.KindModule), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
			return result, result.Error
		}
	}

	execCtx.Logger.Info("Starting module execution",
		zap.String("target", execCtx.Target),
		zap.Int("steps", len(module.Steps)),
	)

	// Export workflow state (write workflow YAML to output)
	if !e.skipWorkspace && !e.disableWorkflowState && !e.dryRun {
		if stateWorkflowFile, ok := execCtx.GetVariable("StateWorkflowFile"); ok {
			if swfStr, ok := stateWorkflowFile.(string); ok && swfStr != "" {
				if err := ExportWorkflowState(swfStr, module); err != nil {
					execCtx.Logger.Warn("Failed to export workflow state", zap.Error(err))
				}
			}
		}
	}

	// Set silent mode on dispatcher when progress bar is active (suppresses LLM output, etc.)
	if e.progressBar != nil {
		e.stepDispatcher.SetSilent(true)
	}

	// Show workflow info or dry-run header
	if e.dryRun {
		tactic := params["tactic"]
		if tactic == "" {
			tactic = "default"
		}
		printDryRunHeader(module.Name, string(core.KindModule), params["target"], tactic, len(module.Steps), execCtx)
	} else if e.progressBar == nil {
		e.printer.WorkflowInfo(module.Name, module.Description, module.Tags, string(module.Runner), len(module.Steps))
	}

	// Show target space folder location
	if !e.skipWorkspace && !e.dryRun && e.progressBar == nil {
		if targetSpace, ok := execCtx.GetVariable("TargetSpace"); ok {
			if tsStr, ok := targetSpace.(string); ok && tsStr != "" {
				e.printer.Info("Reserving target space folder at: %s", terminal.Cyan(tsStr))
			}
		}
	}

	// Execute steps
	e.logger.Debug("Starting step execution loop",
		zap.Int("total_steps", len(module.Steps)),
	)

	// Check if any step has dependencies - use DAG execution if so
	if hasAnyStepDependencies(module.Steps) {
		// Validate dependencies
		if err := validateStepDependencies(module.Steps); err != nil {
			result.Status = core.RunStatusFailed
			result.Error = err
			result.EndTime = time.Now()
			metrics.RecordWorkflowEnd(module.Name, string(core.KindModule), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
			return result, err
		}
		if err := detectStepCycles(module.Steps); err != nil {
			result.Status = core.RunStatusFailed
			result.Error = err
			result.EndTime = time.Now()
			metrics.RecordWorkflowEnd(module.Name, string(core.KindModule), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
			return result, err
		}

		// DAG-based execution
		if err := e.executeStepsDAG(ctx, module.Steps, execCtx, result); err != nil {
			metrics.RecordWorkflowEnd(module.Name, string(core.KindModule), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
			return result, err
		}
	} else {
		// Original sequential execution (backwards compatible)
		currentStep := 0
		for currentStep < len(module.Steps) {
			select {
			case <-ctx.Done():
				if e.progressBar != nil {
					e.progressBar.Abort()
				}
				result.Status = core.RunStatusCancelled
				result.EndTime = time.Now()
				metrics.RecordWorkflowEnd(module.Name, string(core.KindModule), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
				return result, ctx.Err()
			default:
			}

			step := &module.Steps[currentStep]

			e.logger.Debug("Executing step",
				zap.Int("step_index", currentStep),
				zap.String("step_name", step.Name),
				zap.String("step_type", string(step.Type)),
			)

			stepResult, err := e.executeStep(ctx, step, execCtx)
			result.Steps = append(result.Steps, stepResult)

			// Update progress bar with completed step
			if e.progressBar != nil {
				symbol := terminal.StepTypeSymbol(string(step.Type), string(step.StepRunner))
				status := "success"
				switch stepResult.Status {
				case core.StepStatusFailed:
					status = "failed"
				case core.StepStatusSkipped:
					status = "skipped"
				}
				e.progressBar.AddCompletedStep(step.Name, symbol, string(step.Type), status, stepResult.Duration, getStepCommand(step), stepResult.Output)
				e.progressBar.Add(1)
			}

			// Call step completed callback (for database progress tracking)
			if e.onStepCompleted != nil && e.dbRunUUID != "" {
				e.onStepCompleted(ctx, e.dbRunUUID)
			}

			e.logger.Debug("Step execution result",
				zap.String("step", step.Name),
				zap.String("status", string(stepResult.Status)),
				zap.Duration("duration", stepResult.Duration),
			)

			// Record step duration metrics
			metrics.RecordStepDuration(string(step.Type), string(stepResult.Status), stepResult.Duration.Seconds())

			if err != nil {
				execCtx.Logger.Error("Step failed",
					zap.String("step", step.Name),
					zap.Error(err),
				)

				// Record step failure metrics
				metrics.RecordStepFailure(step.Name, string(step.Type), "execution_error")

				// Check if we should continue on error
				if !e.shouldContinueOnError(step) {
					result.Status = core.RunStatusFailed
					result.Error = err
					result.EndTime = time.Now()
					metrics.RecordWorkflowEnd(module.Name, string(core.KindModule), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
					return result, err
				}
			}

			// Handle decision routing
			if stepResult.NextStep != "" {
				if stepResult.NextStep == "_end" {
					break
				}
				// Find next step by name
				nextIdx := e.findStepIndex(module.Steps, stepResult.NextStep)
				if nextIdx >= 0 {
					currentStep = nextIdx
					continue
				}
			}

			currentStep++
		}
	}

	result.Status = core.RunStatusCompleted
	result.EndTime = time.Now()
	result.Exports = execCtx.Exports

	// Count completed steps from results
	completedCount := 0
	for _, stepResult := range result.Steps {
		if stepResult.Status == core.StepStatusSuccess {
			completedCount++
		}
	}

	execCtx.Logger.Info("Module execution completed",
		zap.Int("total_steps", len(module.Steps)),
		zap.Int("completed", completedCount),
		zap.String("elapsed", formatDuration(result.EndTime.Sub(result.StartTime))),
	)

	// Record workflow completion metrics
	metrics.RecordWorkflowEnd(module.Name, string(core.KindModule), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())

	// Send webhook notification on completion
	notify.TriggerWebhooks(cfg, "workflow_"+string(result.Status), map[string]interface{}{
		"workflow": module.Name,
		"kind":     string(core.KindModule),
		"target":   execCtx.Target,
		"status":   string(result.Status),
		"duration": result.EndTime.Sub(result.StartTime).Seconds(),
	})

	// Export state on module completion
	if stateFile, ok := execCtx.GetVariable("StateFile"); ok {
		if sfStr, ok := stateFile.(string); ok && sfStr != "" {
			if err := ExportState(sfStr, result, execCtx); err != nil {
				execCtx.Logger.Warn("Failed to export state", zap.Error(err))
			}
		}
	}

	// Write run-completed.json on completion
	if completedFile, ok := execCtx.GetVariable("StateCompletedFile"); ok {
		if cfStr, ok := completedFile.(string); ok && cfStr != "" {
			if err := ExportRunCompleted(cfStr, result, execCtx); err != nil {
				execCtx.Logger.Warn("Failed to write run completed file", zap.Error(err))
			}
		}
	}

	// Register artifacts (reports from workflow + state files)
	if !e.dryRun {
		if err := RegisterArtifacts(module, execCtx, e.dbRunID, execCtx.Logger); err != nil {
			execCtx.Logger.Warn("Failed to register artifacts", zap.Error(err))
		}
	}

	// Flush write coordinator at workflow completion
	if e.writeCoordinator != nil {
		if err := e.writeCoordinator.FlushAll(context.Background()); err != nil {
			execCtx.Logger.Warn("Failed to flush write coordinator", zap.Error(err))
		}
	}

	return result, nil
}

// executeStepsDAG executes steps using DAG-based parallel execution
// Uses condition variables instead of polling for efficient step coordination
func (e *Executor) executeStepsDAG(ctx context.Context, steps []core.Step, execCtx *core.ExecutionContext, result *core.WorkflowResult) error {
	dependents, inDegree := buildStepDependencyGraph(steps)
	stepMap := buildStepMap(steps)

	var mu sync.Mutex
	cond := sync.NewCond(&mu) // Condition variable for signaling
	executed := make(map[string]bool)
	failed := make(map[string]bool)
	var firstError error
	var completedCount int32        // Atomic counter for O(1) completion check
	totalSteps := int32(len(steps)) // Total steps to execute

	// Initialize ready queue with steps that have no dependencies
	ready := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			ready = append(ready, name)
		}
	}

	// Semaphore for concurrency limit
	maxConcurrency := 8
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	// Lock-free result collection
	collector := NewResultCollector(len(steps))

	for {
		mu.Lock()

		// O(1) completion check using atomic counter
		if atomic.LoadInt32(&completedCount) == totalSteps {
			mu.Unlock()
			break
		}

		// Wait for signal instead of polling when no ready steps
		for len(ready) == 0 && atomic.LoadInt32(&completedCount) < totalSteps {
			cond.Wait()
		}

		// Re-check completion after wake
		if atomic.LoadInt32(&completedCount) == totalSteps {
			mu.Unlock()
			break
		}

		// Get next ready step
		if len(ready) == 0 {
			mu.Unlock()
			continue
		}

		stepName := ready[0]
		ready = ready[1:]

		if executed[stepName] {
			mu.Unlock()
			continue
		}

		// Check if dependency failed
		step := stepMap[stepName]
		shouldSkip := false
		for _, dep := range step.DependsOn {
			if failed[dep] {
				shouldSkip = true
				break
			}
		}

		if shouldSkip {
			executed[stepName] = true
			failed[stepName] = true
			atomic.AddInt32(&completedCount, 1)
			// Unblock dependents
			for _, dependent := range dependents[stepName] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					ready = append(ready, dependent)
				}
			}
			cond.Signal() // Wake main loop
			mu.Unlock()
			continue
		}
		mu.Unlock()

		// Execute step
		wg.Add(1)
		sem <- struct{}{}

		go func(s *core.Step, sName string) {
			defer wg.Done()
			defer func() { <-sem }()

			// Check context
			select {
			case <-ctx.Done():
				mu.Lock()
				executed[sName] = true
				atomic.AddInt32(&completedCount, 1)
				cond.Signal()
				mu.Unlock()
				return
			default:
			}

			stepResult, err := e.executeStep(ctx, s, execCtx)

			mu.Lock()
			defer mu.Unlock()

			collector.Add(stepResult)
			executed[sName] = true
			atomic.AddInt32(&completedCount, 1)

			// Update progress bar
			if e.progressBar != nil {
				symbol := terminal.StepTypeSymbol(string(s.Type), string(s.StepRunner))
				var status string
				switch stepResult.Status {
				case core.StepStatusFailed:
					status = "failed"
				case core.StepStatusSkipped:
					status = "skipped"
				default:
					status = "success"
				}
				e.progressBar.AddCompletedStep(s.Name, symbol, string(s.Type), status, stepResult.Duration, getStepCommand(s), stepResult.Output)
				e.progressBar.Add(1)
			}

			// Callback
			if e.onStepCompleted != nil && e.dbRunUUID != "" {
				e.onStepCompleted(ctx, e.dbRunUUID)
			}

			if err != nil {
				failed[sName] = true
				if firstError == nil && !e.shouldContinueOnError(s) {
					firstError = err
				}
				metrics.RecordStepFailure(s.Name, string(s.Type), "execution_error")
			}

			metrics.RecordStepDuration(string(s.Type), string(stepResult.Status), stepResult.Duration.Seconds())

			// Unblock dependents
			for _, dependent := range dependents[sName] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 && !executed[dependent] {
					ready = append(ready, dependent)
				}
			}

			cond.Signal() // Wake main loop to process newly ready steps
		}(step, stepName)
	}

	wg.Wait()

	// Collect all results from lock-free collector
	result.Steps = collector.Results()

	if firstError != nil {
		result.Status = core.RunStatusFailed
		result.Error = firstError
		result.EndTime = time.Now()
		return firstError
	}

	return nil
}

// preloadModules loads all flow modules in parallel for faster startup.
// Returns a map of module name to loaded workflow. Modules that fail to load
// are logged but not included in the result (will be loaded on-demand during execution).
func (e *Executor) preloadModules(ctx context.Context, modules []core.ModuleRef) map[string]*core.Workflow {
	result := make(map[string]*core.Workflow)
	if len(modules) == 0 || e.loader == nil {
		return result
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 8) // Limit concurrent loads to avoid file descriptor exhaustion

	for i := range modules {
		modRef := &modules[i]

		// Skip inline modules - they don't need loading from disk
		if modRef.IsInline() {
			mu.Lock()
			result[modRef.Name] = modRef.ToWorkflow()
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(ref *core.ModuleRef) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			case sem <- struct{}{}:
				defer func() { <-sem }()
			}

			workflow, err := e.loader.LoadWorkflowByPath(ref.Path)
			if err != nil {
				e.logger.Warn("Module preload failed",
					zap.String("module", ref.Name),
					zap.String("path", ref.Path),
					zap.Error(err),
				)
				return
			}

			mu.Lock()
			result[ref.Name] = workflow
			mu.Unlock()
		}(modRef)
	}

	wg.Wait()
	return result
}

// ExecuteFlow executes a flow workflow
func (e *Executor) ExecuteFlow(ctx context.Context, flow *core.Workflow, params map[string]string, cfg *config.Config) (*core.WorkflowResult, error) {
	e.logger.Debug("ExecuteFlow called",
		zap.String("flow", flow.Name),
		zap.Int("module_count", len(flow.Modules)),
		zap.Bool("dry_run", e.dryRun),
	)

	if !flow.IsFlow() {
		return nil, fmt.Errorf("workflow is not a flow")
	}

	// Create cancellable context for run registry support
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register with run control plane if we have a database run UUID
	// This enables API-based cancellation of the run
	if e.dbRunUUID != "" {
		activeRun := GetRunControlPlane().Register(e.dbRunUUID, cancel)
		defer GetRunControlPlane().Unregister(e.dbRunUUID)
		e.logger.Debug("Registered flow with control plane",
			zap.String("run_uuid", e.dbRunUUID),
			zap.Time("started_at", activeRun.StartedAt),
		)
	}

	// Check loader is configured for loading nested modules
	if e.loader == nil {
		return nil, fmt.Errorf("no loader configured for flow execution")
	}

	// Create execution context
	runUUID := e.dbRunUUID
	if runUUID == "" {
		runUUID = uuid.New().String() // Full UUID when not in server mode
	}
	e.logger.Debug("Created flow execution context",
		zap.String("run_uuid", runUUID),
		zap.String("target", params["target"]),
	)
	execCtx := core.NewExecutionContext(flow.Name, core.KindFlow, runUUID, params["target"])
	execCtx.Logger = logger.WithWorkflow(flow.Name, runUUID)

	// Inject builtin variables
	e.logger.Debug("Injecting builtin variables for flow")
	e.injectBuiltinVariables(cfg, params, execCtx)
	tempCleanup := e.setupTempDirectory(execCtx)
	defer tempCleanup()
	e.debugLogTargetVariables(execCtx)

	if !e.skipWorkspace && !e.dryRun && database.GetDB() != nil {
		output, _ := execCtx.GetVariable("Output")
		stateExecutionLog, _ := execCtx.GetVariable("StateExecutionLog")
		stateCompletedFile, _ := execCtx.GetVariable("StateCompletedFile")
		stateWorkflowFile, _ := execCtx.GetVariable("StateWorkflowFile")
		stateWorkflowFolder, _ := execCtx.GetVariable("StateWorkflowFolder")

		outputStr, _ := output.(string)
		stateExecutionLogStr, _ := stateExecutionLog.(string)
		stateCompletedFileStr, _ := stateCompletedFile.(string)
		stateWorkflowFileStr, _ := stateWorkflowFile.(string)
		stateWorkflowFolderStr, _ := stateWorkflowFolder.(string)

		if execCtx.WorkspaceName != "" {
			_ = database.EnsureWorkspaceRuntime(ctx, execCtx.WorkspaceName, outputStr, execCtx.WorkflowName, stateExecutionLogStr, stateCompletedFileStr, stateWorkflowFileStr, stateWorkflowFolderStr)
		}
	}

	// Add file logging in server mode
	// Setup file logging to {{Output}}/run-execution.log
	if !e.skipWorkspace {
		if logPath, ok := execCtx.GetVariable("StateExecutionLog"); ok {
			if logStr, ok := logPath.(string); ok && logStr != "" {
				combinedLogger, err := logger.WithFileOutput(execCtx.Logger, logStr)
				if err != nil {
					e.logger.Warn("Failed to setup file logging", zap.Error(err))
				} else if combinedLogger != nil {
					execCtx.Logger = combinedLogger
				}
			}
		}
	}

	// Setup console capture to {{Output}}/run-console.log
	if !e.skipWorkspace && !e.dryRun {
		if logPath, ok := execCtx.GetVariable("StateConsoleLog"); ok {
			if logStr, ok := logPath.(string); ok && logStr != "" {
				var err error
				e.consoleCapture, err = console.StartCapture(logStr)
				if err != nil {
					e.logger.Warn("Failed to setup console capture", zap.Error(err))
				}
			}
		}
	}
	if e.consoleCapture != nil {
		defer func() {
			_ = e.consoleCapture.Stop()
			e.consoleCapture = nil
		}()
	}

	// Remove run-completed.json from previous run (fresh start)
	if !e.skipWorkspace {
		if completedFile, ok := execCtx.GetVariable("StateCompletedFile"); ok {
			if cfStr, ok := completedFile.(string); ok {
				RemoveRunCompleted(cfStr)
			}
		}
	}

	// Set config on step dispatcher for executors that need it (e.g., LLM)
	e.stepDispatcher.SetConfig(cfg)

	// Initialize result
	result := &core.WorkflowResult{
		WorkflowName: flow.Name,
		WorkflowKind: core.KindFlow,
		RunUUID:      runUUID,
		Target:       params["target"],
		Status:       core.RunStatusRunning,
		StartTime:    time.Now(),
		Exports:      make(map[string]interface{}),
	}

	// Record workflow start for metrics
	metrics.RecordWorkflowStart()

	// Initialize parameters
	if err := e.initializeParams(flow, params, execCtx); err != nil {
		result.Status = core.RunStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		metrics.RecordWorkflowEnd(flow.Name, string(core.KindFlow), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
		return result, err
	}

	// Debug log initialized parameters
	execCtx.Logger.Debug("Parameters initialized",
		zap.Int("param_count", len(params)),
		zap.Any("cli_params", params),
		zap.Int("workflow_param_defs", len(flow.Params)),
	)

	// Check dependencies (including function conditions)
	if flow.Dependencies != nil {
		if err := e.checkDependencies(flow.Dependencies, execCtx, cfg.BinariesPath); err != nil {
			// Don't log TargetTypeMismatchError - it will be handled specially in CLI
			if _, ok := err.(*TargetTypeMismatchError); !ok {
				execCtx.Logger.Error("Dependency check failed", zap.Error(err))
			}
			result.Status = core.RunStatusFailed
			result.Error = fmt.Errorf("dependency check failed: %w", err)
			result.EndTime = time.Now()
			metrics.RecordWorkflowEnd(flow.Name, string(core.KindFlow), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
			return result, result.Error
		}
	}

	execCtx.Logger.Info("Starting flow execution",
		zap.String("target", execCtx.Target),
		zap.Int("modules", len(flow.Modules)),
	)

	// Show dry-run header for flows
	if e.dryRun {
		tactic := params["tactic"]
		if tactic == "" {
			tactic = "default"
		}
		printDryRunHeader(flow.Name, string(core.KindFlow), params["target"], tactic, len(flow.Modules), execCtx)
	}

	// Show target space folder location
	if !e.skipWorkspace && !e.dryRun && e.progressBar == nil {
		if targetSpace, ok := execCtx.GetVariable("TargetSpace"); ok {
			if tsStr, ok := targetSpace.(string); ok && tsStr != "" {
				e.printer.Info("Reserving target space folder at: %s", terminal.Cyan(tsStr))
			}
		}
	}

	// Export flow workflow state (write workflow YAML to output)
	if !e.skipWorkspace && !e.disableWorkflowState && !e.dryRun {
		if stateWorkflowFile, ok := execCtx.GetVariable("StateWorkflowFile"); ok {
			if swfStr, ok := stateWorkflowFile.(string); ok && swfStr != "" {
				if err := ExportWorkflowState(swfStr, flow); err != nil {
					execCtx.Logger.Warn("Failed to export flow workflow state", zap.Error(err))
				}
			}
		}
	}

	// Parse excluded modules
	excludeList := parseExcludeList(params["exclude_modules"])

	// Pre-load all modules in parallel for faster startup
	execCtx.Logger.Debug("Pre-loading modules", zap.Int("count", len(flow.Modules)))
	preloadStart := time.Now()
	preloaded := e.preloadModules(ctx, flow.Modules)
	execCtx.Logger.Debug("Modules pre-loaded",
		zap.Int("loaded", len(preloaded)),
		zap.Duration("duration", time.Since(preloadStart)),
	)

	// Build dependency graph using Kahn's algorithm for O(V+E) execution
	// instead of O(n²) naive loop restart
	dependents, inDegree := buildDependencyGraph(flow.Modules)
	moduleMap := buildModuleMap(flow.Modules)
	executed := make(map[string]bool)

	// Initialize ready queue with modules that have no dependencies
	ready := make([]string, 0, len(flow.Modules))
	for name, degree := range inDegree {
		if degree == 0 {
			ready = append(ready, name)
		}
	}

	// Process ready queue (Kahn's algorithm)
	for len(ready) > 0 {
		// Check for context cancellation at start of each iteration
		select {
		case <-ctx.Done():
			if e.progressBar != nil {
				e.progressBar.Abort()
			}
			result.Status = core.RunStatusCancelled
			result.EndTime = time.Now()
			execCtx.Logger.Warn("Flow execution cancelled")
			metrics.RecordWorkflowEnd(flow.Name, string(core.KindFlow), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
			return result, ctx.Err()
		default:
		}

		// Pop from ready queue
		modName := ready[0]
		ready = ready[1:]

		if executed[modName] {
			continue
		}

		modRef := moduleMap[modName]

		// Check if module is excluded
		if isModuleExcluded(modRef.Name, excludeList) {
			execCtx.Logger.Info("Skipping excluded module", zap.String("module", modRef.Name))
			executed[modRef.Name] = true
			// Unblock dependents even for excluded modules
			for _, dependent := range dependents[modRef.Name] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 && !executed[dependent] {
					ready = append(ready, dependent)
				}
			}
			continue
		}

		// Check condition
		if modRef.Condition != "" {
			ok, err := e.functionRegistry.EvaluateCondition(modRef.Condition, execCtx.GetVariables())
			if err != nil {
				execCtx.Logger.Warn("Condition evaluation failed", zap.Error(err))
				executed[modRef.Name] = true
				// Unblock dependents even for skipped modules
				for _, dependent := range dependents[modRef.Name] {
					inDegree[dependent]--
					if inDegree[dependent] == 0 && !executed[dependent] {
						ready = append(ready, dependent)
					}
				}
				continue
			}
			if !ok {
				executed[modRef.Name] = true
				// Unblock dependents even for skipped modules
				for _, dependent := range dependents[modRef.Name] {
					inDegree[dependent]--
					if inDegree[dependent] == 0 && !executed[dependent] {
						ready = append(ready, dependent)
					}
				}
				continue
			}
		}

		// Execute module
		if modRef.IsInline() {
			execCtx.Logger.Info("Executing inline module",
				zap.String("module", modRef.Name),
			)
		} else {
			execCtx.Logger.Info("Executing module",
				zap.String("module", modRef.Name),
				zap.String("path", modRef.Path),
			)
		}

		// Use preloaded module if available, else load on-demand or use inline
		var module *core.Workflow
		var err error
		if preloadedMod, ok := preloaded[modRef.Name]; ok {
			module = preloadedMod
			execCtx.Logger.Debug("Using preloaded module", zap.String("module", modRef.Name))
		} else if modRef.IsInline() {
			// Use inline module definition
			module = modRef.ToWorkflow()
			execCtx.Logger.Debug("Using inline module", zap.String("module", modRef.Name))
		} else {
			// Load the module workflow on-demand (fallback for failed preloads)
			module, err = e.loader.LoadWorkflowByPath(modRef.Path)
			if err != nil {
				execCtx.Logger.Error("Failed to load module",
					zap.String("module", modRef.Name),
					zap.String("path", modRef.Path),
					zap.Error(err))
				result.Status = core.RunStatusFailed
				result.Error = fmt.Errorf("failed to load module %s: %w", modRef.Name, err)
				result.EndTime = time.Now()
				return result, result.Error
			}
		}

		// Merge flow variables (params + exports) with module-specific params
		// This ensures nested modules inherit flow params and can access exports from previous modules
		mergedParams := make(map[string]string)

		// Start with CLI params
		for k, v := range params {
			mergedParams[k] = v
		}

		// Add flow name so modules know which flow they're running under
		mergedParams["flow_name"] = flow.Name

		// Add flow's initialized params and exports (converts interface{} to string)
		for k, v := range execCtx.GetVariables() {
			if _, exists := mergedParams[k]; !exists {
				// Skip FlowName - it will be set by injectBuiltinVariables from flow_name
				// The flow's FlowName is empty (no parent flow), but modules need it set to flow.Name
				if k == "FlowName" {
					continue
				}
				// Only add if not already set by CLI params (CLI has priority)
				if strVal, ok := v.(string); ok {
					mergedParams[k] = strVal
				} else {
					mergedParams[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		// Override with ModuleRef.Params (template-rendered)
		for k, v := range modRef.Params {
			rendered, err := e.templateEngine.Render(v, execCtx.GetVariables())
			if err != nil {
				execCtx.Logger.Warn("Failed to render param",
					zap.String("key", k),
					zap.Error(err))
				mergedParams[k] = v // Use original value on error
			} else {
				mergedParams[k] = rendered
			}
		}

		// Execute the module
		moduleResult, err := e.ExecuteModule(ctx, module, mergedParams, cfg)
		if err != nil {
			// Check for context cancellation FIRST (interrupt/timeout)
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				if e.progressBar != nil {
					e.progressBar.Abort()
				}
				result.Status = core.RunStatusCancelled
				result.EndTime = time.Now()
				execCtx.Logger.Warn("Flow execution cancelled", zap.Error(err))
				metrics.RecordWorkflowEnd(flow.Name, string(core.KindFlow), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
				return result, err
			}

			execCtx.Logger.Error("Module execution failed",
				zap.String("module", modRef.Name),
				zap.Error(err))

			// Handle on_error actions
			for _, action := range modRef.OnError {
				e.handleModuleAction(action, execCtx)
				if action.Action == "abort" {
					result.Status = core.RunStatusFailed
					result.Error = fmt.Errorf("module %s failed: %w", modRef.Name, err)
					result.EndTime = time.Now()
					metrics.RecordWorkflowEnd(flow.Name, string(core.KindFlow), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
					return result, result.Error
				}
			}
			// If no abort action, mark as executed and continue
			executed[modRef.Name] = true
			// Unblock dependents even for failed modules (if not aborted)
			for _, dependent := range dependents[modRef.Name] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 && !executed[dependent] {
					ready = append(ready, dependent)
				}
			}
			continue
		}

		// Merge module exports to flow context
		if moduleResult != nil {
			for k, v := range moduleResult.Exports {
				execCtx.MergeExports(map[string]interface{}{k: v})
			}

			// Handle on_success actions
			if moduleResult.Status == core.RunStatusCompleted {
				for _, action := range modRef.OnSuccess {
					e.handleModuleAction(action, execCtx)
				}
			}
		}

		executed[modRef.Name] = true

		// Unblock dependent modules (Kahn's algorithm propagation)
		for _, dependent := range dependents[modRef.Name] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 && !executed[dependent] {
				ready = append(ready, dependent)
			}
		}
	}

	// Check for circular dependencies (modules that were never executed)
	for name, degree := range inDegree {
		if !executed[name] && degree > 0 {
			result.Status = core.RunStatusFailed
			result.Error = fmt.Errorf("circular dependency detected: module %s has unmet dependencies", name)
			result.EndTime = time.Now()
			metrics.RecordWorkflowEnd(flow.Name, string(core.KindFlow), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
			return result, result.Error
		}
	}

	// Check if context was cancelled during execution
	if ctx.Err() != nil {
		result.Status = core.RunStatusCancelled
		result.EndTime = time.Now()
		execCtx.Logger.Warn("Flow execution cancelled")
		metrics.RecordWorkflowEnd(flow.Name, string(core.KindFlow), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())
		return result, ctx.Err()
	}

	result.Status = core.RunStatusCompleted
	result.EndTime = time.Now()
	result.Exports = execCtx.Exports

	execCtx.Logger.Info("Flow execution completed",
		zap.Duration("duration", result.EndTime.Sub(result.StartTime)),
	)

	// Record workflow completion metrics
	metrics.RecordWorkflowEnd(flow.Name, string(core.KindFlow), string(result.Status), result.EndTime.Sub(result.StartTime).Seconds())

	// Send webhook notification on flow completion
	notify.TriggerWebhooks(cfg, "workflow_"+string(result.Status), map[string]interface{}{
		"workflow": flow.Name,
		"kind":     string(core.KindFlow),
		"target":   execCtx.Target,
		"status":   string(result.Status),
		"duration": result.EndTime.Sub(result.StartTime).Seconds(),
	})

	// Export state on flow completion
	if stateFile, ok := execCtx.GetVariable("StateFile"); ok {
		if sfStr, ok := stateFile.(string); ok && sfStr != "" {
			if err := ExportState(sfStr, result, execCtx); err != nil {
				execCtx.Logger.Warn("Failed to export state", zap.Error(err))
			}
		}
	}

	// Write run-completed.json on completion
	if completedFile, ok := execCtx.GetVariable("StateCompletedFile"); ok {
		if cfStr, ok := completedFile.(string); ok && cfStr != "" {
			if err := ExportRunCompleted(cfStr, result, execCtx); err != nil {
				execCtx.Logger.Warn("Failed to write run completed file", zap.Error(err))
			}
		}
	}

	// Register artifacts (reports from workflow + state files)
	if !e.dryRun {
		if err := RegisterArtifacts(flow, execCtx, e.dbRunID, execCtx.Logger); err != nil {
			execCtx.Logger.Warn("Failed to register artifacts", zap.Error(err))
		}
	}

	// Flush write coordinator at workflow completion
	if e.writeCoordinator != nil {
		if err := e.writeCoordinator.FlushAll(context.Background()); err != nil {
			execCtx.Logger.Warn("Failed to flush write coordinator", zap.Error(err))
		}
	}

	return result, nil
}

// executeStep executes a single step
func (e *Executor) executeStep(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	stepLogger := logger.WithStep(execCtx.WorkflowName, execCtx.RunUUID, step.Name)
	stepLogger.Debug("executeStep called",
		zap.String("step_name", step.Name),
		zap.String("type", string(step.Type)),
		zap.Bool("has_pre_condition", step.PreCondition != ""),
		zap.Bool("has_exports", step.HasExports()),
	)
	stepLogger.Info("Executing step",
		zap.String("type", string(step.Type)),
		zap.Bool("dry_run", e.dryRun),
	)

	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
	}

	// Check pre-condition (skip in dry-run mode)
	if step.PreCondition != "" && !e.dryRun {
		// Render template variables in pre_condition
		renderedCondition, err := e.templateEngine.Render(step.PreCondition, execCtx.GetVariables())
		if err != nil {
			stepLogger.Warn("Failed to render pre_condition, using original",
				zap.String("condition", step.PreCondition),
				zap.Error(err))
			renderedCondition = step.PreCondition
		}
		stepLogger.Debug("Evaluating pre-condition",
			zap.String("condition", renderedCondition),
		)
		ok, err := e.functionRegistry.EvaluateCondition(renderedCondition, execCtx.GetVariables())
		if err != nil {
			stepLogger.Debug("Pre-condition evaluation failed", zap.Error(err))
			result.Status = core.StepStatusFailed
			result.Error = fmt.Errorf("pre-condition evaluation failed: %w", err)
			result.EndTime = time.Now()
			e.printer.StepFailed(step.Name, result.Error)
			return result, result.Error
		}
		stepLogger.Debug("Pre-condition evaluated", zap.Bool("result", ok))
		if !ok {
			result.Status = core.StepStatusSkipped
			result.EndTime = time.Now()
			stepLogger.Info("Step skipped (pre-condition not met)")
			e.printer.StepSkipped(step.Name)
			return result, nil
		}
	}

	// Handle dry-run mode
	if e.dryRun {
		stepLogger.Info("DRY-RUN: Would execute step",
			zap.String("step", step.Name),
			zap.String("type", string(step.Type)),
			zap.String("command", step.Command),
		)
		fmt.Printf("%s Step %d: %s (%s)\n", terminal.SymbolStart, execCtx.StepIndex+1, step.Name, terminal.TypeBadge(string(step.Type)))

		// Show pre-condition if present (skipped in dry-run)
		if step.PreCondition != "" {
			renderedCond, _ := e.templateEngine.Render(step.PreCondition, execCtx.GetVariables())
			fmt.Printf("  Pre-condition: %s %s\n", terminal.Gray(renderedCond), terminal.Yellow("(skipped in dry-run)"))
		}

		if step.Command != "" {
			// Render the command for display
			rendered, _ := e.templateEngine.Render(step.Command, execCtx.GetVariables())
			fmt.Printf("  Would execute: %s\n", terminal.Gray(rendered))
		}
		if len(step.Commands) > 0 {
			fmt.Printf("  Would execute in %s:\n", step.Type)
			for _, cmd := range step.Commands {
				rendered, _ := e.templateEngine.Render(cmd, execCtx.GetVariables())
				fmt.Printf("    %s %s\n", terminal.SymbolBullet, terminal.Gray(rendered))
			}
		}

		// Display function(s) for function steps
		if step.Function != "" {
			rendered, _ := e.templateEngine.Render(step.Function, execCtx.GetVariables())
			fmt.Printf("  Would execute: %s\n", terminal.Gray(rendered))
		}
		if len(step.Functions) > 0 {
			fmt.Printf("  Would execute functions:\n")
			for _, fn := range step.Functions {
				rendered, _ := e.templateEngine.Render(fn, execCtx.GetVariables())
				fmt.Printf("    %s %s\n", terminal.SymbolBullet, terminal.Gray(rendered))
			}
		}
		if len(step.ParallelFunctions) > 0 {
			fmt.Printf("  Would execute in parallel:\n")
			for _, fn := range step.ParallelFunctions {
				rendered, _ := e.templateEngine.Render(fn, execCtx.GetVariables())
				fmt.Printf("    %s %s\n", terminal.SymbolBullet, terminal.Gray(rendered))
			}
		}

		// In dry-run mode, set all exports to true
		if step.HasExports() {
			result.Exports = make(map[string]interface{})
			for exportName := range step.Exports {
				result.Exports[exportName] = true
			}
			fmt.Printf("  Exports: ")
			first := true
			for name := range step.Exports {
				if !first {
					fmt.Printf(", ")
				}
				fmt.Printf("%s=%s", name, terminal.Green("true"))
				first = false
			}
			fmt.Printf(" %s\n", terminal.Yellow("(dry-run)"))
		}
		fmt.Println()

		result.Status = core.StepStatusSuccess
		result.Output = "[DRY-RUN] Skipped"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		execCtx.StepIndex++
		return result, nil
	}

	// Get step type symbol, command prefix, and command for display
	stepSymbol := terminal.StepTypeSymbol(string(step.Type), string(step.StepRunner))
	cmdPrefix := terminal.StepCommandPrefix(string(step.Type))
	stepCommand := getStepCommand(step)
	stepCommandColored := getStepCommandColored(step)
	if stepCommand != "" {
		stepCommand, _ = e.templateEngine.Render(stepCommand, execCtx.GetVariables())
		stepCommandColored, _ = e.templateEngine.Render(stepCommandColored, execCtx.GetVariables())
	}

	// Show step start (skip when progress bar is active)
	if e.progressBar == nil {
		e.printer.StepStartWithCommand(step.Name, stepSymbol, stepCommandColored, cmdPrefix)
	} else {
		// Update progress bar with current step command (uncolored for progress bar)
		e.progressBar.SetCommand(stepCommand)
	}

	// Start spinner for step execution (optional, when spinner enabled and no progress bar)
	var sp *terminal.Spinner
	if e.showSpinner && e.progressBar == nil {
		sp = terminal.StepSpinner(step.Name)
		sp.Start()
	}

	// Dispatch to appropriate executor
	stepResult, err := e.stepDispatcher.Dispatch(ctx, step, execCtx)

	// Stop spinner
	if sp != nil {
		sp.Stop()
	}
	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		// Show failed step (skip when progress bar is active)
		if e.progressBar == nil {
			e.printer.StepFailedWithCommand(step.Name, stepSymbol, err, stepCommand, cmdPrefix)
		}

		// Process on_error handlers
		e.processOnError(ctx, step, execCtx, err)

		// Log step failure to state execution log file
		execCtx.Logger.Error("Step failed",
			zap.String("step_name", step.Name),
			zap.String("step_type", string(step.Type)),
			zap.String("command", stepCommand),
			zap.Duration("duration", result.Duration),
			zap.Error(err),
		)

		return result, err
	}

	// Copy results
	result.Status = stepResult.Status
	result.Output = stepResult.Output
	result.Exports = stepResult.Exports
	result.NextStep = stepResult.NextStep
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Show step result (skip when progress bar is active)
	if e.progressBar == nil {
		if result.Status == core.StepStatusSkipped {
			e.printer.StepSkippedWithCommand(step.Name, stepSymbol)
		} else {
			e.printer.StepSuccessWithCommand(step.Name, stepSymbol, formatDuration(result.Duration), stepCommand, cmdPrefix)
			// Show actual output unless silent mode
			if !e.silent && result.Output != "" {
				e.printer.VerboseOutput(result.Output)
			}
		}
	}

	// Always write verbose output to console log file (even in silent mode)
	// This ensures run-console.log has complete output with colors
	if e.consoleCapture != nil && result.Output != "" && (e.silent || e.progressBar != nil) {
		e.writeVerboseOutputToLog(result.Output)
	}

	// Merge exports into context
	if stepResult.Exports != nil {
		execCtx.MergeExports(stepResult.Exports)
	}

	// Process on_success handlers
	e.processOnSuccess(ctx, step, execCtx)

	// Evaluate decision routing
	if step.HasDecision() {
		result.NextStep = e.evaluateDecision(step.Decision, execCtx)
	}

	// Log step execution details to state execution log file
	execCtx.Logger.Info("Step completed",
		zap.String("step_name", step.Name),
		zap.String("step_type", string(step.Type)),
		zap.String("command", stepCommand),
		zap.String("status", string(result.Status)),
		zap.Duration("duration", result.Duration),
		zap.String("output", result.Output),
	)

	// Buffer step result and progress for batch insertion via write coordinator
	if e.writeCoordinator != nil && !e.dryRun {
		startedAt := result.StartTime
		completedAt := result.EndTime
		errorMsg := ""
		if result.Error != nil {
			errorMsg = result.Error.Error()
		}
		e.writeCoordinator.AddStepResult(step.Name, string(step.Type), string(result.Status),
			stepCommand, result.Output, errorMsg, result.Exports,
			result.Duration.Milliseconds(), &startedAt, &completedAt)
		e.writeCoordinator.IncrementProgress(1)
	}

	return result, nil
}

// initializeParams initializes parameters from defaults, generators, and provided values
func (e *Executor) initializeParams(workflow *core.Workflow, params map[string]string, execCtx *core.ExecutionContext) error {
	// FIRST: Set ALL provided params as variables immediately
	// This ensures event trigger vars are available before required param validation
	for name, value := range params {
		execCtx.SetVariable(name, value)
	}

	// THEN: Validate required params and apply defaults/generators
	for _, param := range workflow.Params {
		var value string
		var hasValue bool

		// Check if value was provided via CLI/params
		if v, ok := params[param.Name]; ok && v != "" {
			value = v
			hasValue = true
		}

		// Try generator if no value provided
		if !hasValue && param.Generator != "" {
			v, err := e.templateEngine.ExecuteGenerator(param.Generator)
			if err != nil {
				return fmt.Errorf("generator failed for %s: %w", param.Name, err)
			}
			value = v
			hasValue = true
		}

		// Use default if no value provided
		if !hasValue && param.HasDefault() {
			// For bool params with native bool defaults, use directly
			if param.IsBool() {
				execCtx.SetParam(param.Name, param.DefaultBool())
				continue
			}
			// Render template in default value (for string defaults)
			rendered, err := e.templateEngine.Render(param.DefaultString(), execCtx.GetVariables())
			if err != nil {
				return fmt.Errorf("template failed for %s default: %w", param.Name, err)
			}
			value = rendered
			hasValue = true
		}

		// Check if required and still no value
		if !hasValue {
			if param.Required {
				return fmt.Errorf("required parameter missing: %s", param.Name)
			}
			continue
		}

		// Convert value to appropriate type based on param.Type
		if param.IsBool() {
			// Store as actual bool for native boolean checks in pre_condition
			boolVal := strings.ToLower(value) == "true" || value == "1"
			execCtx.SetParam(param.Name, boolVal)
		} else {
			execCtx.SetParam(param.Name, value)
		}
	}

	return nil
}

// findStepIndex finds a step by name and returns its index
func (e *Executor) findStepIndex(steps []core.Step, name string) int {
	for i, step := range steps {
		if step.Name == name {
			return i
		}
	}
	return -1
}

// shouldContinueOnError checks if execution should continue after an error
func (e *Executor) shouldContinueOnError(step *core.Step) bool {
	for _, action := range step.OnError {
		if action.Action == core.ActionContinue {
			return true
		}
	}
	return false
}

// buildDependencyGraph builds the dependency graph for Kahn's algorithm
// Returns:
//   - dependents: map from module name to list of modules that depend on it
//   - inDegree: map from module name to number of unexecuted dependencies
func buildDependencyGraph(modules []core.ModuleRef) (dependents map[string][]string, inDegree map[string]int) {
	dependents = make(map[string][]string)
	inDegree = make(map[string]int)

	// Initialize all modules with in-degree 0
	for i := range modules {
		inDegree[modules[i].Name] = 0
	}

	// Build dependency relationships
	for i := range modules {
		mod := &modules[i]
		for _, dep := range mod.DependsOn {
			dependents[dep] = append(dependents[dep], mod.Name)
			inDegree[mod.Name]++
		}
	}

	return dependents, inDegree
}

// buildModuleMap creates a name -> ModuleRef lookup map
func buildModuleMap(modules []core.ModuleRef) map[string]*core.ModuleRef {
	moduleMap := make(map[string]*core.ModuleRef, len(modules))
	for i := range modules {
		moduleMap[modules[i].Name] = &modules[i]
	}
	return moduleMap
}

// hasAnyStepDependencies checks if any step has depends_on defined
func hasAnyStepDependencies(steps []core.Step) bool {
	for i := range steps {
		if len(steps[i].DependsOn) > 0 {
			return true
		}
	}
	return false
}

// buildStepDependencyGraph builds dependency graph for steps using Kahn's algorithm
func buildStepDependencyGraph(steps []core.Step) (dependents map[string][]string, inDegree map[string]int) {
	dependents = make(map[string][]string)
	inDegree = make(map[string]int)

	for i := range steps {
		inDegree[steps[i].Name] = 0
	}

	for i := range steps {
		step := &steps[i]
		for _, dep := range step.DependsOn {
			dependents[dep] = append(dependents[dep], step.Name)
			inDegree[step.Name]++
		}
	}
	return dependents, inDegree
}

// buildStepMap creates name -> *Step lookup
func buildStepMap(steps []core.Step) map[string]*core.Step {
	m := make(map[string]*core.Step, len(steps))
	for i := range steps {
		m[steps[i].Name] = &steps[i]
	}
	return m
}

// detectStepCycles detects circular dependencies using DFS
func detectStepCycles(steps []core.Step) error {
	stepMap := make(map[string][]string)
	for i := range steps {
		stepMap[steps[i].Name] = steps[i].DependsOn
	}

	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(name string, path []string) error
	dfs = func(name string, path []string) error {
		visited[name] = true
		recStack[name] = true
		path = append(path, name)

		for _, dep := range stepMap[name] {
			if !visited[dep] {
				if err := dfs(dep, path); err != nil {
					return err
				}
			} else if recStack[dep] {
				return fmt.Errorf("circular dependency: %s", strings.Join(append(path, dep), " -> "))
			}
		}
		recStack[name] = false
		return nil
	}

	for i := range steps {
		if !visited[steps[i].Name] {
			if err := dfs(steps[i].Name, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateStepDependencies checks for invalid dependency references
func validateStepDependencies(steps []core.Step) error {
	names := make(map[string]bool)
	for i := range steps {
		names[steps[i].Name] = true
	}
	for i := range steps {
		for _, dep := range steps[i].DependsOn {
			if !names[dep] {
				return fmt.Errorf("step '%s' depends on non-existent step '%s'", steps[i].Name, dep)
			}
		}
	}
	return nil
}

// evaluateDecision evaluates decision routing and returns the next step.
func (e *Executor) evaluateDecision(decision *core.DecisionConfig, execCtx *core.ExecutionContext) string {
	if decision == nil {
		return ""
	}

	vars := execCtx.GetVariables()

	// Handle switch/case syntax
	if decision.Switch != "" {
		// Render the switch expression
		switchValue, err := e.templateEngine.Render(decision.Switch, vars)
		if err != nil {
			return ""
		}
		switchValue = strings.TrimSpace(switchValue)

		// Look up the case
		if caseAction, ok := decision.Cases[switchValue]; ok {
			return caseAction.Goto
		}

		// Fall through to default
		if decision.Default != nil {
			return decision.Default.Goto
		}
	}

	return ""
}

// handleModuleAction handles a module action (for flow execution)
func (e *Executor) handleModuleAction(action core.Action, execCtx *core.ExecutionContext) {
	// Check condition if present
	if action.Condition != "" {
		ok, err := e.functionRegistry.EvaluateCondition(action.Condition, execCtx.GetVariables())
		if err != nil || !ok {
			return
		}
	}

	switch action.Action {
	case core.ActionLog:
		rendered, _ := e.templateEngine.Render(action.Message, execCtx.GetVariables())
		execCtx.Logger.Info(rendered)

	case core.ActionExport:
		execCtx.SetExport(action.Name, action.Value)
	}
}

// processOnSuccess processes on_success handlers
func (e *Executor) processOnSuccess(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) {
	for _, action := range step.OnSuccess {
		e.processAction(ctx, &action, execCtx)
	}
}

// processOnError processes on_error handlers
func (e *Executor) processOnError(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext, err error) {
	for _, action := range step.OnError {
		e.processAction(ctx, &action, execCtx)
	}
}

// processAction processes a single action
func (e *Executor) processAction(ctx context.Context, action *core.Action, execCtx *core.ExecutionContext) {
	// Check condition if present
	if action.Condition != "" {
		ok, err := e.functionRegistry.EvaluateCondition(action.Condition, execCtx.GetVariables())
		if err != nil || !ok {
			return
		}
	}

	switch action.Action {
	case core.ActionLog:
		rendered, _ := e.templateEngine.Render(action.Message, execCtx.GetVariables())
		execCtx.Logger.Info(rendered)

	case core.ActionExport:
		execCtx.SetExport(action.Name, action.Value)

	case core.ActionRun:
		// Execute embedded step
		if action.Type == core.StepTypeBash && action.Command != "" {
			step := &core.Step{
				Name:    "action-run",
				Type:    core.StepTypeBash,
				Command: action.Command,
			}
			_, _ = e.stepDispatcher.Dispatch(ctx, step, execCtx)
		}
	}
}
