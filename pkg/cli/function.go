package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

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
	funcColumnWidth  int
	funcShowExample  bool
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

func init() {
	functionEvalCmd.Flags().StringVarP(&evalScript, "eval", "e", "", "script to evaluate")
	functionEvalCmd.Flags().StringVarP(&evalTarget, "target", "t", "", "target value for {{target}} variable")
	functionEvalCmd.Flags().StringArrayVar(&evalParams, "params", nil, "additional parameters (key=value format)")
	functionEvalCmd.Flags().BoolVar(&evalStdin, "stdin", false, "read script from stdin")
	functionEvalCmd.Flags().StringVarP(&evalFunctionName, "function", "f", "", "function name to call (remaining args become function arguments)")

	functionListCmd.Flags().StringVarP(&funcSearchFilter, "search", "s", "", "filter functions by name or description")
	functionListCmd.Flags().IntVar(&funcColumnWidth, "width", 0, "max column width (wraps lines instead of truncating)")
	functionListCmd.Flags().BoolVar(&funcShowExample, "example", false, "show example usage below each function description")

	functionCmd.AddCommand(functionEvalCmd)
	functionCmd.AddCommand(functionListCmd)
}

func runFunctionEval(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

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

	// Determine script source: -f flag > positional arg > -e flag > stdin
	var script string

	// Handle -f/--function flag: build script from function name + positional args
	if evalFunctionName != "" {
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
		return fmt.Errorf("no script provided: use positional argument, -e flag, or --stdin")
	}

	// Use the resolved script
	evalScript = script

	// 1. Build context with target and params
	ctx := make(map[string]interface{})
	if evalTarget != "" {
		ctx["target"] = evalTarget
	}

	for _, p := range evalParams {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) == 2 {
			ctx[parts[0]] = parts[1]
		}
	}

	// 2. Render template variables ({{target}}, etc.)
	templateEngine := template.NewEngine()
	renderedScript, err := templateEngine.Render(evalScript, ctx)
	if err != nil {
		printer.Error("Template rendering failed: %s", err)
		return fmt.Errorf("template rendering failed: %w", err)
	}

	// Show rendered script if different from original (verbose mode)
	if verbose && renderedScript != evalScript {
		printer.Info("Rendered script: %s", renderedScript)
	}

	// 3. Execute as JavaScript using Otto runtime
	registry := functions.NewRegistry()
	result, err := registry.Execute(renderedScript, ctx)
	if err != nil {
		printer.Error("Execution failed: %s", err)
		return fmt.Errorf("execution failed: %w", err)
	}

	// 4. Print result
	if result != nil {
		fmt.Println(result)
	}

	return nil
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
	if funcColumnWidth > 0 {
		printMarkdownTableWithWidth(headers, rows, funcColumnWidth)
	} else {
		printMarkdownTable(headers, rows)
	}
	return nil
}
