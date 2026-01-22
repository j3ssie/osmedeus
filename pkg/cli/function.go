package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/functions"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	evalScript       string
	evalTarget       string
	evalParams       []string
	evalStdin        bool
	evalFunctionName string
	funcSearchFilter string
	funcShowExample  bool

	// Bulk processing flags
	funcTargetsFile  string
	funcFunctionFile string
	funcConcurrency  int

	// Repeat flags
	funcRepeat         bool
	funcRepeatWaitTime string
)

// functionCmd is the parent command for function operations
var functionCmd = &cobra.Command{
	Use:     "function",
	Aliases: []string{"func"},
	Short:   "Execute and test utility functions",
	Long:    UsageFunction(),
}

// functionEvalCmd evaluates a script with template rendering and function execution
var functionEvalCmd = &cobra.Command{
	Use:     "eval",
	Aliases: []string{"e"},
	Short:   "Evaluate a script with template rendering and function execution",
	Long:    UsageFunctionEval(),
	RunE:    runFunctionEval,
}

// functionListCmd lists all available functions
var functionListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all available utility functions",
	RunE:    runFunctionList,
}

// evalCmd is a top-level shorthand for 'func eval'
var evalCmd = &cobra.Command{
	Use:     "eval",
	Aliases: []string{"e", "ev", "evl", "evla"},
	Short:   "Evaluate a script (shorthand for 'func eval')",
	Long:    UsageFunctionEval(),
	RunE:    runFunctionEval,
}

func init() {
	functionEvalCmd.Flags().StringVarP(&evalScript, "eval", "e", "", "script to evaluate")
	functionEvalCmd.Flags().StringVarP(&evalTarget, "target", "t", "", "target value for {{target}} variable")
	functionEvalCmd.Flags().StringArrayVar(&evalParams, "params", nil, "additional parameters (key=value format)")
	functionEvalCmd.Flags().BoolVar(&evalStdin, "stdin", false, "read script from stdin")
	functionEvalCmd.Flags().StringVarP(&evalFunctionName, "function", "f", "", "function name to call (remaining args become function arguments)")

	// Bulk processing flags
	functionEvalCmd.Flags().StringVarP(&funcTargetsFile, "targets", "T", "", "file containing targets (one per line)")
	functionEvalCmd.Flags().StringVar(&funcFunctionFile, "function-file", "", "file containing the function/script to execute")
	functionEvalCmd.Flags().IntVarP(&funcConcurrency, "concurrency", "c", 1, "number of concurrent executions")

	// Repeat flags
	functionEvalCmd.Flags().BoolVar(&funcRepeat, "repeat", false, "repeat run after completion")
	functionEvalCmd.Flags().StringVar(&funcRepeatWaitTime, "repeat-wait-time", "5s", "wait time between repeats (e.g., 30s, 20m, 10h, 1d)")

	functionListCmd.Flags().StringVarP(&funcSearchFilter, "search", "s", "", "filter functions by name or description")
	// Note: --width flag is now global (defined in root.go)
	functionListCmd.Flags().BoolVar(&funcShowExample, "example", false, "show example usage below each function description")

	// evalCmd flags (same as functionEvalCmd - it's a shorthand)
	evalCmd.Flags().StringVarP(&evalScript, "eval", "e", "", "script to evaluate")
	evalCmd.Flags().StringVarP(&evalTarget, "target", "t", "", "target value for {{target}} variable")
	evalCmd.Flags().StringArrayVar(&evalParams, "params", nil, "additional parameters (key=value format)")
	evalCmd.Flags().BoolVar(&evalStdin, "stdin", false, "read script from stdin")
	evalCmd.Flags().StringVarP(&evalFunctionName, "function", "f", "", "function name to call (remaining args become function arguments)")
	evalCmd.Flags().StringVarP(&funcTargetsFile, "targets", "T", "", "file containing targets (one per line)")
	evalCmd.Flags().StringVar(&funcFunctionFile, "function-file", "", "file containing the function/script to execute")
	evalCmd.Flags().IntVarP(&funcConcurrency, "concurrency", "c", 1, "number of concurrent executions")
	evalCmd.Flags().BoolVar(&funcRepeat, "repeat", false, "repeat run after completion")
	evalCmd.Flags().StringVar(&funcRepeatWaitTime, "repeat-wait-time", "5s", "wait time between repeats (e.g., 30s, 20m, 10h, 1d)")

	functionCmd.AddCommand(functionEvalCmd)
	functionCmd.AddCommand(functionListCmd)
}

func runFunctionEval(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	// Parse repeat wait time
	var waitDuration time.Duration
	if funcRepeat {
		var err error
		waitDuration, err = parseRunDuration(funcRepeatWaitTime)
		if err != nil {
			return fmt.Errorf("invalid repeat-wait-time: %w", err)
		}
		printer.Info("Repeat mode enabled, wait time: %s", funcRepeatWaitTime)
	}

	// Setup signal handling for repeat mode
	sigChan := make(chan os.Signal, 1)
	if funcRepeat {
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigChan)
	}

	// Connect to database for db_* functions (skip if --disable-db is set)
	if !disableDB {
		cfg := config.Get()
		if cfg != nil {
			// Try to connect to database - don't fail if DB not available
			if _, dbErr := database.Connect(cfg); dbErr != nil {
				if verbose {
					printer.Info("Database connection warning: %s", dbErr)
				}
			}
		}
	}

	// Determine script source: --function-file > -f flag > positional arg > -e flag > stdin
	var script string

	// Read script from function file if provided
	if funcFunctionFile != "" {
		data, err := os.ReadFile(funcFunctionFile)
		if err != nil {
			printer.Error("Failed to read function file: %s", err)
			return fmt.Errorf("failed to read function file: %w", err)
		}
		script = strings.TrimSpace(string(data))
	} else if evalFunctionName != "" {
		// Handle -f/--function flag: build script from function name + positional args
		var quotedArgs []string
		for _, arg := range args {
			quotedArgs = append(quotedArgs, fmt.Sprintf("%q", arg))
		}
		script = fmt.Sprintf("%s(%s)", evalFunctionName, strings.Join(quotedArgs, ", "))
	} else if len(args) > 0 && args[0] != "-" {
		if len(args) > 1 {
			// Multiple args: treat first as function name, rest as arguments
			// e.g., "func_name arg1 arg2" → "func_name("arg1", "arg2")"
			var quotedArgs []string
			for _, arg := range args[1:] {
				quotedArgs = append(quotedArgs, fmt.Sprintf("%q", arg))
			}
			script = fmt.Sprintf("%s(%s)", args[0], strings.Join(quotedArgs, ", "))
		} else {
			// Single arg: use as-is (could be full expression or function name with no args)
			script = args[0]
		}
	} else if evalScript != "" {
		// Script provided via -e flag
		script = evalScript
	} else if evalStdin || (len(args) > 0 && args[0] == "-") {
		// Read script from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			printer.Error("Failed to read from stdin: %s", err)
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		script = strings.TrimSpace(string(data))
	}

	if script == "" {
		return fmt.Errorf("no script provided: use positional argument, -e flag, --function-file, or --stdin")
	}

	// Main execution loop
	iteration := 0
	for {
		iteration++
		if funcRepeat && iteration > 1 {
			printer.Section(fmt.Sprintf("Repeat Iteration %d", iteration))
		}

		var lastErr error
		// Bulk processing mode: process multiple targets from file
		if funcTargetsFile != "" {
			lastErr = runBulkFunctionEval(printer, script)
		} else {
			// Single target execution (existing behavior)
			lastErr = executeFunctionForTarget(printer, script, evalTarget)
		}

		if !funcRepeat {
			return lastErr
		}

		printer.Info("Iteration %d completed. Waiting %s before next iteration...", iteration, funcRepeatWaitTime)
		printer.Info("Press Ctrl+C to stop repeat mode")

		select {
		case <-time.After(waitDuration):
			// Continue to next iteration
		case <-sigChan:
			printer.Info("Interrupt received, stopping repeat mode")
			return nil
		}
	}
}

// runBulkFunctionEval processes the script for multiple targets from a file
func runBulkFunctionEval(printer *terminal.Printer, script string) error {
	// Read targets from file
	targets, err := readFuncTargetsFromFile(funcTargetsFile)
	if err != nil {
		printer.Error("Failed to read targets file: %s", err)
		return fmt.Errorf("failed to read targets file: %w", err)
	}

	if len(targets) == 0 {
		printer.Warning("No targets found in file: %s", funcTargetsFile)
		return nil
	}

	// Deduplicate targets
	targets = deduplicateFuncTargets(targets)

	if verbose {
		printer.Info("Processing %d targets with concurrency %d", len(targets), funcConcurrency)
	}

	// Ensure concurrency is at least 1
	maxConcurrency := funcConcurrency
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}

	// Process targets concurrently
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var errCount int
	var errMu sync.Mutex

	for _, target := range targets {
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire semaphore
			defer func() { <-sem }() // Release semaphore

			if err := executeFunctionForTarget(printer, script, t); err != nil {
				errMu.Lock()
				errCount++
				errMu.Unlock()
				if verbose {
					printer.Error("Failed for target %s: %s", t, err)
				}
			}
		}(target)
	}

	wg.Wait()

	if errCount > 0 {
		printer.Warning("Completed with %d errors out of %d targets", errCount, len(targets))
	} else if verbose {
		printer.Success("Successfully processed %d targets", len(targets))
	}

	return nil
}

// executeFunctionForTarget executes the script for a single target
func executeFunctionForTarget(printer *terminal.Printer, script, target string) error {
	// Build context with target and params
	ctx := make(map[string]interface{})
	if target != "" {
		ctx["target"] = target
	}

	for _, p := range evalParams {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) == 2 {
			ctx[parts[0]] = parts[1]
		}
	}

	// Render template variables ({{target}}, etc.)
	templateEngine := template.NewEngine()
	renderedScript, err := templateEngine.Render(script, ctx)
	if err != nil {
		printer.Error("Template rendering failed: %s", err)
		return fmt.Errorf("template rendering failed: %w", err)
	}

	// Show rendered script if different from original (verbose mode)
	if verbose && renderedScript != script {
		printer.Info("Rendered script: %s", renderedScript)
	}

	// Execute as JavaScript using Otto runtime
	registry := functions.NewRegistry()
	result, err := registry.Execute(renderedScript, ctx)
	if err != nil {
		printer.Error("Execution failed: %s", err)
		return fmt.Errorf("execution failed: %w", err)
	}

	// Print result
	if result != nil {
		fmt.Println(result)
	}

	return nil
}

// readFuncTargetsFromFile reads targets from a file, one per line
func readFuncTargetsFromFile(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var result []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}
	return result, scanner.Err()
}

// deduplicateFuncTargets removes duplicates and empty strings
func deduplicateFuncTargets(inputTargets []string) []string {
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

func runFunctionList(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	printer.Section("Available Utility Functions")
	fmt.Println()

	// Use positional arg as search if --search not provided
	if funcSearchFilter == "" && len(args) > 0 {
		funcSearchFilter = args[0]
	}

	// Get function registry and category order
	registry := functions.FunctionRegistry()
	categories := functions.CategoryOrder()

	// Build rows for all functions
	var rows [][]string
	searchLower := strings.ToLower(funcSearchFilter)

	for _, cat := range categories {
		if funcs, ok := registry[cat.Key]; ok {
			for _, fn := range funcs {
				// Apply search filter if specified (matches name, description, or category)
				if funcSearchFilter != "" {
					nameLower := strings.ToLower(fn.Name)
					descLower := strings.ToLower(fn.Description)
					catLower := strings.ToLower(cat.Title)
					if !strings.Contains(nameLower, searchLower) &&
						!strings.Contains(descLower, searchLower) &&
						!strings.Contains(catLower, searchLower) {
						continue
					}
				}
				// Build description, optionally with example
				desc := fn.Description
				if funcShowExample && fn.Example != "" {
					desc = fn.Description + "\n" + terminal.Gray("e.g. "+fn.Example)
				}
				rows = append(rows, []string{
					terminal.Yellow(cat.ShortTitle),
					terminal.Cyan(fn.Signature),
					desc,
					terminal.Magenta(fn.ReturnType),
				})
			}
		}
	}

	if len(rows) == 0 {
		if funcSearchFilter != "" {
			printer.Info("No functions matching '%s'", funcSearchFilter)
		} else {
			printer.Info("No functions available")
		}
		return nil
	}

	if funcSearchFilter != "" {
		printer.Info("Found %d function(s) matching '%s':", len(rows), funcSearchFilter)
		fmt.Println()
	}

	headers := []string{"Category", "Function", "Description", "Returns"}
	if globalWidth > 0 {
		printMarkdownTableWithWidth(headers, rows, globalWidth)
	} else {
		printMarkdownTable(headers, rows)
	}
	return nil
}
