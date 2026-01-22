package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/j3ssie/osmedeus/v5/internal/config"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/linter"
	"github.com/j3ssie/osmedeus/v5/internal/parser"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

// workflowCmd represents the workflow command
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage workflows",
	Long:  UsageWorkflow(),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default to 'list' when no subcommand is specified
		return workflowListCmd.RunE(cmd, args)
	},
}

// workflowListCmd lists available workflows
var workflowListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available workflows",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		loader := parser.NewLoader(cfg.WorkflowsPath)

		// List flows
		flows, err := loader.ListFlows()
		if err != nil {
			return fmt.Errorf("failed to list flows: %w", err)
		}

		// List modules
		modules, err := loader.ListModules()
		if err != nil {
			return fmt.Errorf("failed to list modules: %w", err)
		}

		// Table format output
		printer := terminal.NewPrinter()
		fmt.Println()
		title := fmt.Sprintf("Available Workflows (%s)", terminal.Gray(cfg.WorkflowsPath))
		if len(filterTags) > 0 {
			title = fmt.Sprintf("Available Workflows - Filtered by tags: %s", terminal.Cyan(strings.Join(filterTags, ", ")))
		}
		fmt.Printf("%s %s\n", terminal.InfoSymbol(), terminal.Bold(title))

		// Combined table for all workflows
		if len(flows) > 0 || len(modules) > 0 {
			// Collect all workflow data first to calculate column widths
			type workflowRow struct {
				name        string
				colorName   string // name with color codes
				wfType      string
				desc        string
				reqParams   string
				steps       string // step/module count
				tags        string
				targetTypes string
			}
			var rows []workflowRow
			uniqueTags := make(map[string]bool) // Collect unique tags across all workflows

			// Track workflows with errors for verbose mode
			type workflowError struct {
				name string
				err  error
			}
			var workflowErrors []workflowError
			hasTargetRequired := false // Track if any workflow has required Target

			// Load flows
			for _, f := range flows {
				wf, err := loader.LoadWorkflow(f)
				if err != nil {
					workflowErrors = append(workflowErrors, workflowError{name: f, err: err})
					continue
				}
				// Skip hidden workflows
				if wf.Hidden {
					continue
				}
				// Skip if tags filter is specified and workflow doesn't match
				if len(filterTags) > 0 && !hasMatchingTags(wf, filterTags) {
					continue
				}
				desc := "-"
				if wf.Description != "" {
					desc = truncateString(wf.Description, 50)
				}
				reqParams := getRequiredParams(wf)
				// Track if any workflow has required Target
				if workflowHasRequiredTarget(wf) {
					hasTargetRequired = true
				}
				steps := "-"
				// Count modules for flows
				if len(wf.Modules) > 0 {
					steps = fmt.Sprintf("%d modules", len(wf.Modules))
				}
				tags := "-"
				if len(wf.Tags) > 0 {
					tags = truncateString(strings.Join(wf.Tags, ", "), 25)
					// Collect unique tags (from raw tags, not truncated)
					for _, tag := range wf.Tags {
						uniqueTags[tag] = true
					}
				}
				targetTypes := getTargetTypes(wf)
				name := f
				colorName := f
				if f == "general" {
					name = f + " (default)"
					colorName = terminal.Green(name)
				}
				rows = append(rows, workflowRow{name, colorName, "flow", desc, reqParams, steps, tags, targetTypes})
			}

			// Load modules
			for _, m := range modules {
				wf, err := loader.LoadWorkflow(m)
				if err != nil {
					workflowErrors = append(workflowErrors, workflowError{name: m, err: err})
					continue
				}
				// Skip hidden workflows
				if wf.Hidden {
					continue
				}
				// Skip if tags filter is specified and workflow doesn't match
				if len(filterTags) > 0 && !hasMatchingTags(wf, filterTags) {
					continue
				}
				desc := "-"
				if wf.Description != "" {
					desc = truncateString(wf.Description, 50)
				}
				reqParams := getRequiredParams(wf)
				// Track if any workflow has required Target
				if workflowHasRequiredTarget(wf) {
					hasTargetRequired = true
				}
				steps := "-"
				// Count steps for modules
				if len(wf.Steps) > 0 {
					steps = fmt.Sprintf("%d steps", len(wf.Steps))
				}
				tags := "-"
				if len(wf.Tags) > 0 {
					tags = truncateString(strings.Join(wf.Tags, ", "), 25)
					// Collect unique tags (from raw tags, not truncated)
					for _, tag := range wf.Tags {
						uniqueTags[tag] = true
					}
				}
				targetTypes := getTargetTypes(wf)
				rows = append(rows, workflowRow{m, m, "module", desc, reqParams, steps, tags, targetTypes})
			}

			// Check if any workflows matched the filter
			if len(rows) == 0 && len(filterTags) > 0 {
				fmt.Println()
				printer.Warning("No workflows found with tags: %s", strings.Join(filterTags, ", "))
				fmt.Println()
				return nil
			}

			// Calculate max widths
			nameWidth := len("Name")
			typeWidth := len("Type")
			descWidth := len("Description")
			paramsWidth := len("Required Params")
			stepsWidth := len("Steps")
			tagsWidth := len("Tags")
			targetTypesWidth := len("Target Types")

			for _, r := range rows {
				if len(r.name) > nameWidth {
					nameWidth = len(r.name)
				}
				if len(r.wfType) > typeWidth {
					typeWidth = len(r.wfType)
				}
				if len(r.desc) > descWidth {
					descWidth = len(r.desc)
				}
				if len(r.reqParams) > paramsWidth {
					paramsWidth = len(r.reqParams)
				}
				if len(r.steps) > stepsWidth {
					stepsWidth = len(r.steps)
				}
				if showTags && len(r.tags) > tagsWidth {
					tagsWidth = len(r.tags)
				}
				if len(r.targetTypes) > targetTypesWidth {
					targetTypesWidth = len(r.targetTypes)
				}
			}

			// Print markdown table with styled headers
			fmt.Println()
			if showTags {
				// With Tags column
				fmt.Printf("| %s%s | %s%s | %s%s | %s%s | %s%s | %s%s | %s%s |\n",
					terminal.Bold("Name"), strings.Repeat(" ", nameWidth-4),
					terminal.Bold("Type"), strings.Repeat(" ", typeWidth-4),
					terminal.Bold("Description"), strings.Repeat(" ", descWidth-11),
					terminal.Bold("Required Params"), strings.Repeat(" ", paramsWidth-15),
					terminal.Bold("Steps"), strings.Repeat(" ", stepsWidth-5),
					terminal.Bold("Target Types"), strings.Repeat(" ", targetTypesWidth-12),
					terminal.Bold("Tags"), strings.Repeat(" ", tagsWidth-4))
				fmt.Printf("|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|\n",
					strings.Repeat("-", nameWidth), strings.Repeat("-", typeWidth),
					strings.Repeat("-", descWidth), strings.Repeat("-", paramsWidth),
					strings.Repeat("-", stepsWidth), strings.Repeat("-", targetTypesWidth),
					strings.Repeat("-", tagsWidth))
			} else {
				// Without Tags column (default)
				fmt.Printf("| %s%s | %s%s | %s%s | %s%s | %s%s | %s%s |\n",
					terminal.Bold("Name"), strings.Repeat(" ", nameWidth-4),
					terminal.Bold("Type"), strings.Repeat(" ", typeWidth-4),
					terminal.Bold("Description"), strings.Repeat(" ", descWidth-11),
					terminal.Bold("Required Params"), strings.Repeat(" ", paramsWidth-15),
					terminal.Bold("Steps"), strings.Repeat(" ", stepsWidth-5),
					terminal.Bold("Target Types"), strings.Repeat(" ", targetTypesWidth-12))
				fmt.Printf("|-%s-|-%s-|-%s-|-%s-|-%s-|-%s-|\n",
					strings.Repeat("-", nameWidth), strings.Repeat("-", typeWidth),
					strings.Repeat("-", descWidth), strings.Repeat("-", paramsWidth),
					strings.Repeat("-", stepsWidth), strings.Repeat("-", targetTypesWidth))
			}

			for _, r := range rows {
				// Calculate padding needed (color codes don't take visual space)
				namePad := nameWidth - len(r.name)
				var colorType string
				switch r.wfType {
				case "flow":
					colorType = terminal.Cyan(r.wfType)
				default:
					colorType = terminal.Yellow(r.wfType)
				}
				typePad := typeWidth - len(r.wfType)
				colorSteps := terminal.Gray(r.steps)
				stepsPad := stepsWidth - len(r.steps)
				colorTargetTypes := terminal.Magenta(r.targetTypes)
				targetTypesPad := targetTypesWidth - len(r.targetTypes)
				// Highlight "Target" in blue within reqParams
				colorReqParams := strings.ReplaceAll(r.reqParams, "Target", terminal.HiBlue("Target"))
				paramsPad := paramsWidth - len(r.reqParams)

				if showTags {
					colorTags := terminal.Gray(r.tags)
					tagsPad := tagsWidth - len(r.tags)
					fmt.Printf("| %s%s | %s%s | %-*s | %s%s | %s%s | %s%s | %s%s |\n",
						r.colorName, strings.Repeat(" ", namePad),
						colorType, strings.Repeat(" ", typePad),
						descWidth, r.desc,
						colorReqParams, strings.Repeat(" ", paramsPad),
						colorSteps, strings.Repeat(" ", stepsPad),
						colorTargetTypes, strings.Repeat(" ", targetTypesPad),
						colorTags, strings.Repeat(" ", tagsPad))
				} else {
					fmt.Printf("| %s%s | %s%s | %-*s | %s%s | %s%s | %s%s |\n",
						r.colorName, strings.Repeat(" ", namePad),
						colorType, strings.Repeat(" ", typePad),
						descWidth, r.desc,
						colorReqParams, strings.Repeat(" ", paramsPad),
						colorSteps, strings.Repeat(" ", stepsPad),
						colorTargetTypes, strings.Repeat(" ", targetTypesPad))
				}
			}

			fmt.Println()

			// Show tip about Target parameter if any workflow requires it
			if hasTargetRequired {
				fmt.Printf("%s %s\n\n", terminal.HiBlue("ℹ"), terminal.HiBlue("\"Target\" in Required Params is supplied via -t flag (e.g., -t example.com) or each line from -T list-of-targets.txt"))
			}

			// Count flows and modules in filtered results
			flowCount := 0
			moduleCount := 0
			for _, r := range rows {
				switch r.wfType {
				case "flow":
					flowCount++
				default:
					moduleCount++
				}
			}

			// Convert unique tags map to sorted slice
			var tagList []string
			for tag := range uniqueTags {
				tagList = append(tagList, tag)
			}
			// Sort tags alphabetically for consistent display
			if len(tagList) > 1 {
				for i := 0; i < len(tagList)-1; i++ {
					for j := i + 1; j < len(tagList); j++ {
						if tagList[i] > tagList[j] {
							tagList[i], tagList[j] = tagList[j], tagList[i]
						}
					}
				}
			}

			// Summary with colors
			if len(filterTags) > 0 {
				summaryParts := []string{
					terminal.Green(fmt.Sprintf("%d", flowCount)) + " flows",
					terminal.Yellow(fmt.Sprintf("%d", moduleCount)) + " modules",
				}
				fmt.Printf("◆ Matching: %s (filtered by tags: %s)\n",
					strings.Join(summaryParts, ", "),
					terminal.Cyan(strings.Join(filterTags, ", ")))
			} else {
				summaryParts := []string{
					terminal.Green(fmt.Sprintf("%d", flowCount)) + " flows",
					terminal.Yellow(fmt.Sprintf("%d", moduleCount)) + " modules",
				}
				summaryParts = append(summaryParts, terminal.Cyan(fmt.Sprintf("%d", len(tagList)))+" unique tags")
				fmt.Printf("◆ Total: %s\n", strings.Join(summaryParts, ", "))
			}

			// Show available tags
			if len(tagList) > 0 {
				tagsDisplay := strings.Join(tagList, ", ")
				if len(tagsDisplay) > 80 {
					tagsDisplay = tagsDisplay[:77] + "..."
				}
				fmt.Printf("◇ Available tags: %s\n", terminal.Gray(tagsDisplay))
			}

			// View workflow details hint
			binaryPath := os.Args[0]
			fmt.Println()
			fmt.Println("◌ " + terminal.Bold("View workflow details:"))
			fmt.Printf("  %s workflow show %s\n", terminal.Cyan(binaryPath), terminal.Yellow("<workflow_name>"))

			// Validate workflow hint
			fmt.Println()
			fmt.Println("◌ " + terminal.Bold("Validate workflow:"))
			fmt.Printf("  %s workflow validate %s\n", terminal.Cyan(binaryPath), terminal.Yellow("<workflow_name>"))

			// Filter by tags hint
			fmt.Println()
			fmt.Println("◌ " + terminal.Bold("Filter by tags:"))
			fmt.Printf("  %s workflow ls --tags %s\n", terminal.Cyan(binaryPath), terminal.Yellow("recon,fast"))
			fmt.Printf("  %s workflow ls --show-tags\n", terminal.Cyan(binaryPath))

			// Example run usage
			fmt.Println()
			fmt.Println("◌ " + terminal.Bold("Example Run Usage:"))
			fmt.Printf("  %s run -f %s -t <target>\n", terminal.Cyan(binaryPath), terminal.Yellow("<flow_name>"))
			fmt.Printf("  %s run -f general -T list_of_targets.txt\n", terminal.Cyan(binaryPath))
			fmt.Printf("  %s run --threads-hold 10 -t sample.com\n", terminal.Cyan(binaryPath))
			fmt.Printf("  %s run -t sample.com -x %s\n", terminal.Cyan(binaryPath), terminal.Yellow("<exclude_module>"))
			fmt.Printf("  %s run -m %s -t <target> --params %s\n",
				terminal.Cyan(binaryPath),
				terminal.Yellow("<module_name>"),
				terminal.Gray("<key=value>"))

			fmt.Println()
			fmt.Printf("%s Tip: %s %s\n", terminal.Gray(terminal.SymbolLightning), terminal.Cyan("osmedeus run --help"), terminal.Gray("for more usage and options"))
			fmt.Println()

			// Show workflow errors if verbose mode
			if showVerbose && len(workflowErrors) > 0 {
				fmt.Printf("%s Workflows with Errors (%d):\n", terminal.WarningSymbol(), len(workflowErrors))
				for _, we := range workflowErrors {
					fmt.Printf("  %s\n", terminal.Yellow(we.name))
					fmt.Printf("    └─ %s\n", terminal.Gray(we.err.Error()))
				}
				fmt.Println()
			}
		} else {
			printer.Warning("No workflows found in %s", cfg.WorkflowsPath)
		}

		return nil
	},
}

// truncateString truncates a string to maxLen and adds "..." if needed
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getRequiredParams returns a comma-separated list of required parameter names
func getRequiredParams(wf *core.Workflow) string {
	var required []string

	// Check dependencies.variables for required Target
	if wf.Dependencies != nil {
		for _, v := range wf.Dependencies.Variables {
			if v.Required && strings.EqualFold(v.Name, "Target") {
				required = append(required, "Target")
				break
			}
		}
	}

	// Check params (existing logic, but skip if already added Target)
	for _, p := range wf.Params {
		if p.Required {
			// Skip if it's a target param and we already added from dependencies
			if strings.EqualFold(p.Name, "Target") && containsTargetParam(required) {
				continue
			}
			required = append(required, p.Name)
		}
	}

	if len(required) == 0 {
		return "-"
	}
	return strings.Join(required, ", ")
}

// containsTargetParam checks if required list already has Target entry
func containsTargetParam(required []string) bool {
	for _, r := range required {
		if strings.EqualFold(r, "Target") {
			return true
		}
	}
	return false
}

// workflowHasRequiredTarget checks if a workflow has a required Target in dependencies.variables
func workflowHasRequiredTarget(wf *core.Workflow) bool {
	if wf.Dependencies == nil {
		return false
	}
	for _, v := range wf.Dependencies.Variables {
		if v.Required && strings.EqualFold(v.Name, "Target") {
			return true
		}
	}
	return false
}

// getTargetTypes returns a comma-separated list of target types from dependencies
func getTargetTypes(wf *core.Workflow) string {
	if wf.Dependencies == nil {
		return "-"
	}

	var types []string
	seen := make(map[string]bool)

	// First check dependencies.target_types if available
	for _, t := range wf.Dependencies.TargetTypes {
		typeStr := string(t)
		if typeStr != "" && !seen[typeStr] {
			types = append(types, typeStr)
			seen[typeStr] = true
		}
	}

	// Then check dependencies.variables[].type
	for _, v := range wf.Dependencies.Variables {
		typeStr := string(v.Type)
		if typeStr != "" && !seen[typeStr] {
			types = append(types, typeStr)
			seen[typeStr] = true
		}
	}

	if len(types) == 0 {
		return "-"
	}
	return strings.Join(types, ", ")
}

// stripAnsi removes ANSI escape codes for length calculation
func stripAnsi(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// isToggleParam detects boolean/toggle parameters by:
// - Type == "bool"
// - Name patterns: enableX, enable_X, skipX, skip_X, disableX, useX, verboseX
// - Default value is bool (true/false)
func isToggleParam(p core.Param) bool {
	// Check explicit type
	if p.Type == "bool" {
		return true
	}

	// Check name patterns (case-insensitive)
	name := strings.ToLower(p.Name)
	togglePrefixes := []string{"enable", "skip", "disable", "use", "verbose"}
	for _, prefix := range togglePrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
		// Also check with underscore: enable_xxx, skip_xxx
		if strings.HasPrefix(name, prefix+"_") {
			return true
		}
	}

	// Check if default value is boolean
	if p.Default != nil {
		switch p.Default.(type) {
		case bool:
			return true
		}
		// Also check string representation
		defaultStr := strings.ToLower(p.DefaultString())
		if defaultStr == "true" || defaultStr == "false" {
			return true
		}
	}

	return false
}

// isSpeedControlParam detects performance/speed parameters by:
// - Name contains: threads, timeout, rate, concurrency, delay, limit, workers, parallel, batch, interval, retry
// - Name ends with: depth, parallel
// - Default value matches time pattern: \d+[hms]
func isSpeedControlParam(p core.Param) bool {
	name := strings.ToLower(p.Name)
	speedPatterns := []string{
		"threads", "timeout", "rate", "concurrency", "delay",
		"limit", "workers", "parallel", "batch", "interval", "retry",
	}

	for _, pattern := range speedPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	// Check suffix patterns for depth and parallel
	speedSuffixes := []string{"depth", "parallel"}
	for _, suffix := range speedSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	// Check for time pattern in default value (e.g., 8h, 30m, 1800)
	if p.Default != nil {
		defaultStr := p.DefaultString()
		// Match patterns like: 8h, 30m, 1800, 60s
		timePatternRegex := regexp.MustCompile(`^\d+[hms]?$`)
		if timePatternRegex.MatchString(defaultStr) {
			// Also verify it's numeric or has time suffix
			if len(defaultStr) > 0 {
				lastChar := defaultStr[len(defaultStr)-1]
				// If it ends with h, m, or s, it's a time value
				if lastChar == 'h' || lastChar == 'm' || lastChar == 's' {
					return true
				}
				// If purely numeric with reasonable size, could be timeout/limit
				if p.Type == "int" || p.Type == "" {
					// Check if numeric only
					numericRegex := regexp.MustCompile(`^\d+$`)
					if numericRegex.MatchString(defaultStr) {
						// Large numbers (>100) are likely timeouts/limits
						val := 0
						_, _ = fmt.Sscanf(defaultStr, "%d", &val)
						if val > 100 {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// isConfigParam detects configuration parameters by:
// - Name ends with: Config, config, Cfg, cfg
func isConfigParam(p core.Param) bool {
	name := strings.ToLower(p.Name)
	configSuffixes := []string{"config", "cfg"}
	for _, suffix := range configSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}

// categorizeParams groups params into Toggle, Speed, Config, and General categories
func categorizeParams(params []core.Param) (toggle, speed, config, general []core.Param) {
	for _, p := range params {
		if isToggleParam(p) {
			toggle = append(toggle, p)
		} else if isSpeedControlParam(p) {
			speed = append(speed, p)
		} else if isConfigParam(p) {
			config = append(config, p)
		} else {
			general = append(general, p)
		}
	}
	return
}

// printToggleParams prints toggle parameters with green highlighting
func printToggleParams(params []core.Param) {
	if len(params) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("◐ " + terminal.Bold("Toggle Parameters:"))
	var rows [][]string
	for _, p := range params {
		required := terminal.Gray("no")
		if p.Required {
			required = terminal.Green("yes")
		}
		defaultVal := p.DefaultString()
		if defaultVal == "" {
			defaultVal = "-"
		}
		// Color the default value based on true/false
		coloredDefault := defaultVal
		if strings.ToLower(defaultVal) == "true" {
			coloredDefault = terminal.Green(defaultVal)
		} else if strings.ToLower(defaultVal) == "false" {
			coloredDefault = terminal.Gray(defaultVal)
		}
		rows = append(rows, []string{terminal.Green(p.Name), coloredDefault, required})
	}
	printMarkdownTable([]string{"Name", "Default", "Required"}, rows)
}

// printSpeedControlParams prints speed/performance parameters with yellow highlighting
func printSpeedControlParams(params []core.Param) {
	if len(params) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("◎ " + terminal.Bold("Speed Control Parameters:"))
	var rows [][]string
	for _, p := range params {
		required := terminal.Gray("no")
		if p.Required {
			required = terminal.Green("yes")
		}
		defaultVal := p.DefaultString()
		if defaultVal == "" {
			defaultVal = "-"
		}
		// Color numeric values in yellow
		coloredDefault := terminal.Yellow(defaultVal)
		rows = append(rows, []string{terminal.Yellow(p.Name), coloredDefault, required})
	}
	printMarkdownTable([]string{"Name", "Default", "Required"}, rows)
}

// printConfigParams prints config parameters with magenta highlighting
func printConfigParams(params []core.Param) {
	if len(params) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("⚙ " + terminal.Bold("Config Parameters:"))
	var rows [][]string
	for _, p := range params {
		required := terminal.Gray("no")
		if p.Required {
			required = terminal.Green("yes")
		}
		defaultVal := p.DefaultString()
		if defaultVal == "" {
			defaultVal = "-"
		}
		coloredDefault := terminal.Magenta(defaultVal)
		if defaultVal == "-" {
			coloredDefault = defaultVal
		}
		rows = append(rows, []string{terminal.Magenta(p.Name), coloredDefault, required})
	}
	printMarkdownTable([]string{"Name", "Default", "Required"}, rows)
}

// printGeneralParams prints general parameters with cyan highlighting
func printGeneralParams(params []core.Param) {
	if len(params) == 0 {
		return
	}
	fmt.Println()
	fmt.Println("● " + terminal.Bold("General Parameters:"))
	var rows [][]string
	for _, p := range params {
		required := terminal.Gray("no")
		if p.Required {
			required = terminal.Green("yes")
		}
		defaultVal := p.DefaultString()
		if defaultVal == "" {
			defaultVal = "-"
		}
		coloredDefault := terminal.Cyan(defaultVal)
		if defaultVal == "-" {
			coloredDefault = defaultVal
		}
		rows = append(rows, []string{terminal.Cyan(p.Name), coloredDefault, required})
	}
	printMarkdownTable([]string{"Name", "Default", "Required"}, rows)
}

// printMarkdownTable prints an aligned markdown table (supports colored cells)
func printMarkdownTable(headers []string, rows [][]string) {
	// Calculate column widths (using display length, not byte length)
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(stripAnsi(h))
	}
	for _, row := range rows {
		for i, cell := range row {
			displayLen := len(stripAnsi(cell))
			if i < len(widths) && displayLen > widths[i] {
				widths[i] = displayLen
			}
		}
	}

	// Print header
	fmt.Print("|")
	for i, h := range headers {
		fmt.Printf(" %-*s |", widths[i], h)
	}
	fmt.Println()

	// Print separator
	fmt.Print("|")
	for _, w := range widths {
		fmt.Printf("-%s-|", strings.Repeat("-", w))
	}
	fmt.Println()

	// Print rows (with ANSI-aware padding)
	for _, row := range rows {
		fmt.Print("|")
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			// Calculate padding needed (display width vs actual string length)
			displayLen := len(stripAnsi(cell))
			padding := widths[i] - displayLen
			fmt.Printf(" %s%s |", cell, strings.Repeat(" ", padding))
		}
		fmt.Println()
	}
}

// wrapText wraps text to maxWidth characters, preserving existing newlines
func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	var result []string
	// Split on existing newlines first
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(stripAnsi(line)) <= maxWidth {
			result = append(result, line)
			continue
		}

		// Wrap long lines
		remaining := line
		for len(stripAnsi(remaining)) > maxWidth {
			// Find break point - try to break at space
			displayLen := 0
			lastSpace := -1
			byteIdx := 0

			for byteIdx < len(remaining) && displayLen < maxWidth {
				if remaining[byteIdx] == '\x1b' {
					// Skip ANSI escape sequence
					for byteIdx < len(remaining) && remaining[byteIdx] != 'm' {
						byteIdx++
					}
					if byteIdx < len(remaining) {
						byteIdx++
					}
					continue
				}
				if remaining[byteIdx] == ' ' {
					lastSpace = byteIdx
				}
				displayLen++
				byteIdx++
			}
			breakPoint := byteIdx

			// Prefer breaking at space if found
			if lastSpace > 0 && lastSpace > breakPoint/2 {
				breakPoint = lastSpace
			}

			result = append(result, remaining[:breakPoint])
			remaining = strings.TrimLeft(remaining[breakPoint:], " ")
		}
		if remaining != "" {
			result = append(result, remaining)
		}
	}

	return result
}

// printMarkdownTableWithWidth prints a table with column width wrapping
func printMarkdownTableWithWidth(headers []string, rows [][]string, maxWidth int) {
	// First pass: wrap all cells and calculate column widths
	type wrappedRow struct {
		cells     [][]string // Each cell is a slice of lines
		maxHeight int
	}

	var wrappedRows []wrappedRow
	widths := make([]int, len(headers))

	// Initialize widths with header lengths
	for i, h := range headers {
		widths[i] = len(stripAnsi(h))
	}

	// Wrap each cell and track widths
	for _, row := range rows {
		wr := wrappedRow{cells: make([][]string, len(headers))}
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			wrapped := wrapText(cell, maxWidth)
			wr.cells[i] = wrapped
			if len(wrapped) > wr.maxHeight {
				wr.maxHeight = len(wrapped)
			}
			// Track max width for this column
			for _, line := range wrapped {
				lineWidth := len(stripAnsi(line))
				if lineWidth > widths[i] {
					widths[i] = lineWidth
				}
			}
		}
		wrappedRows = append(wrappedRows, wr)
	}

	// Print header
	fmt.Print("|")
	for i, h := range headers {
		fmt.Printf(" %-*s |", widths[i], h)
	}
	fmt.Println()

	// Print separator
	fmt.Print("|")
	for _, w := range widths {
		fmt.Printf("-%s-|", strings.Repeat("-", w))
	}
	fmt.Println()

	// Print rows (with multi-line support)
	for _, wr := range wrappedRows {
		for lineIdx := 0; lineIdx < wr.maxHeight; lineIdx++ {
			fmt.Print("|")
			for colIdx := range headers {
				cell := ""
				if lineIdx < len(wr.cells[colIdx]) {
					cell = wr.cells[colIdx][lineIdx]
				}
				// Calculate padding needed (display width vs actual string length)
				displayLen := len(stripAnsi(cell))
				padding := widths[colIdx] - displayLen
				fmt.Printf(" %s%s |", cell, strings.Repeat(" ", padding))
			}
			fmt.Println()
		}
	}
}

// workflowShowCmd shows workflow details
var workflowShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show workflow details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		printer := terminal.NewPrinter()
		loader := parser.NewLoader(cfg.WorkflowsPath)
		workflow, err := loader.LoadWorkflow(args[0])
		if err != nil {
			// Print detailed error with formatting
			fmt.Println()
			printer.Error("Failed to load workflow: %s", args[0])
			fmt.Println()
			fmt.Println(err.Error())
			fmt.Println()
			return err
		}

		// Show raw YAML with syntax highlighting when --yaml flag is set
		if showYaml {
			content, err := os.ReadFile(workflow.FilePath)
			if err != nil {
				return fmt.Errorf("failed to read workflow file: %w", err)
			}

			// Render with Glamour for syntax highlighting
			markdown := "```yaml\n" + string(content) + "\n```"
			renderer, err := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(120),
			)
			if err == nil {
				rendered, renderErr := renderer.Render(markdown)
				if renderErr == nil {
					fmt.Print(rendered)
					return nil
				}
			}
			// Fallback to plain YAML if rendering fails
			fmt.Println(string(content))
			return nil
		}

		// Default: Table format output
		// Metadata section
		fmt.Println()
		fmt.Println("❯ " + terminal.Bold("Metadata:"))
		printer.KeyValue("Name", workflow.Name)
		printer.KeyValue("Kind", terminal.TypeBadge(string(workflow.Kind)))
		printer.KeyValue("Description", workflow.Description)
		printer.KeyValue("File", terminal.Gray(workflow.FilePath))

		// Show parameters (categorized)
		if len(workflow.Params) > 0 {
			toggle, speed, config, general := categorizeParams(workflow.Params)
			printToggleParams(toggle)
			printSpeedControlParams(speed)
			printConfigParams(config)
			printGeneralParams(general)
		}

		// Show steps (for modules)
		if workflow.IsModule() && len(workflow.Steps) > 0 {
			fmt.Println()
			fmt.Println("◼ " + terminal.Bold("Steps:"))
			var rows [][]string
			for i, s := range workflow.Steps {
				coloredType := terminal.TypeBadge(string(s.Type))
				rows = append(rows, []string{fmt.Sprintf("%d", i+1), s.Name, coloredType})
			}
			printMarkdownTable([]string{"#", "Name", "Type"}, rows)
		}

		// Show modules (for flows)
		if workflow.IsFlow() && len(workflow.Modules) > 0 {
			fmt.Println()
			fmt.Println("◼ " + terminal.Bold("Modules:"))
			var rows [][]string
			for i, m := range workflow.Modules {
				deps := strings.Join(m.DependsOn, ", ")
				if deps == "" {
					deps = "-"
				}
				rows = append(rows, []string{fmt.Sprintf("%d", i+1), m.Name, m.Path, deps})
			}
			printMarkdownTable([]string{"#", "Name", "Path", "Depends On"}, rows)
		}

		// Show triggers
		if len(workflow.Triggers) > 0 {
			fmt.Println()
			fmt.Println("◼ " + terminal.Bold("Triggers:"))
			var rows [][]string
			for _, t := range workflow.Triggers {
				status := terminal.Gray("disabled")
				if t.Enabled {
					status = terminal.Green("enabled")
				}
				rows = append(rows, []string{t.Name, string(t.On), status})
			}
			printMarkdownTable([]string{"Name", "Type", "Status"}, rows)
		}

		// Show builtin variables
		fmt.Println()
		fmt.Println("◆ " + terminal.Bold("Builtin Variables:"))

		if showVerbose {
			// Verbose: show full aligned markdown table with default values and descriptions
			// Get actual default values from config where possible
			baseFolder := "~/osmedeus-base"
			binariesPath := "~/osmedeus-base/external-binaries"
			dataPath := "~/osmedeus-base/data"
			workspacesPath := "~/workspaces-osmedeus"
			workflowsPath := cfg.WorkflowsPath
			if cfg.BaseFolder != "" {
				baseFolder = cfg.BaseFolder
			}
			if cfg.BinariesPath != "" {
				binariesPath = cfg.BinariesPath
			}
			if cfg.DataPath != "" {
				dataPath = cfg.DataPath
			}
			if cfg.WorkspacesPath != "" {
				workspacesPath = cfg.WorkspacesPath
			}

			// Truncate paths for display
			truncatePath := func(p string, maxLen int) string {
				if len(p) <= maxLen {
					return p
				}
				return "..." + p[len(p)-maxLen+3:]
			}

			builtinVars := [][]string{
				// Path variables
				{terminal.Cyan("{{BaseFolder}}"), "Base installation folder", terminal.Gray(truncatePath(baseFolder, 28))},
				{terminal.Cyan("{{Binaries}}"), "Path to binaries", terminal.Gray(truncatePath(binariesPath, 28))},
				{terminal.Cyan("{{Data}}"), "Path to data files", terminal.Gray(truncatePath(dataPath, 28))},
				{terminal.Cyan("{{ExternalConfigs}}"), "Path to external configs", terminal.Gray("{{BaseFolder}}/configs")},
				{terminal.Cyan("{{ExternalScripts}}"), "Path to external scripts", terminal.Gray("{{BaseFolder}}/scripts")},
				{terminal.Cyan("{{Workspaces}}"), "Path to workspaces", terminal.Gray(truncatePath(workspacesPath, 28))},
				{terminal.Cyan("{{Workflows}}"), "Path to workflows", terminal.Gray(truncatePath(workflowsPath, 28))},
				{terminal.Cyan("{{ExternalMarkdowns}}"), "Path to markdown templates", terminal.Gray("{{BaseFolder}}/markdown-report-templates")},
				{terminal.Cyan("{{ExternalAgents}}"), "Path to agent configs", terminal.Gray("{{BaseFolder}}/external-agent-configs")},

				// Target variables
				{terminal.Cyan("{{Target}}"), "Current scan target", terminal.Yellow("<from -t flag>")},
				{terminal.Cyan("{{TargetFile}}"), "File containing targets", terminal.Yellow("<from -T flag>")},
				{terminal.Cyan("{{TargetSpace}}"), "Sanitized target path", terminal.Yellow("<sanitized target>")},
				{terminal.Cyan("{{Output}}"), "Output directory for target", terminal.Gray("{{Workspaces}}/{{TargetSpace}}")},

				// Target type heuristics
				{terminal.Cyan("{{TargetType}}"), "Target type (url/domain/ip/cidr/file)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetRootDomain}}"), "Root domain (domain/URL targets)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetTLD}}"), "Top-level domain (domain/URL targets)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetSLD}}"), "Second-level domain (domain/URL)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{Org}}"), "Alias for TargetSLD", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetBaseURL}}"), "Base URL (URL targets only)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetRootURL}}"), "Root URL (URL targets only)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetHostname}}"), "Hostname (URL targets only)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetHost}}"), "Host with port (URL targets only)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetPort}}"), "Port number (URL targets only)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetPath}}"), "URL path (URL targets only)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetScheme}}"), "URL scheme (URL targets only)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetIsWildcard}}"), "Is wildcard (domain targets only)", terminal.Yellow("<detected>")},
				{terminal.Cyan("{{TargetResolvedIP}}"), "Resolved IP (domain targets only)", terminal.Yellow("<if resolved>")},
				{terminal.Cyan("{{TargetStatusCode}}"), "HTTP status (URL targets only)", terminal.Yellow("<if fetched>")},
				{terminal.Cyan("{{TargetContentLength}}"), "Content length (URL targets only)", terminal.Yellow("<if fetched>")},

				// State files
				{terminal.Cyan("{{StateExecutionLog}}"), "Path to execution log", terminal.Gray("{{Output}}/run-execution.log")},
				{terminal.Cyan("{{StateCompletedFile}}"), "Path to run completed JSON", terminal.Gray("{{Output}}/run-completed.json")},
				{terminal.Cyan("{{StateFile}}"), "Path to run state JSON", terminal.Gray("{{Output}}/run-state.json")},
				{terminal.Cyan("{{StateWorkflowFile}}"), "Path to workflow YAML", terminal.Gray("{{Output}}/run-workflow.yaml")},
				{terminal.Cyan("{{StateWorkflowFolder}}"), "Path to workflow modules", terminal.Gray("{{Output}}/run-modules")},

				// Thread/performance variables
				{terminal.Cyan("{{threads}}"), "Thread count (tactic based)", terminal.Yellow("10 (default tactic)")},
				{terminal.Cyan("{{baseThreads}}"), "Base thread count", terminal.Yellow("10")},

				// Metadata variables
				{terminal.Cyan("{{Version}}"), "Osmedeus version", terminal.Gray("<osmedeus version>")},
				{terminal.Cyan("{{RunUUID}}"), "Unique run identifier (UUID)", terminal.Yellow("<generated>")},
				{terminal.Cyan("{{TaskDate}}"), "Task date (YYYY-MM-DD)", terminal.Yellow("<current date>")},
				{terminal.Cyan("{{Today}}"), "Current date (YYYY-MM-DD)", terminal.Yellow("<current date>")},
				{terminal.Cyan("{{TimeStamp}}"), "Unix timestamp", terminal.Yellow("<unix timestamp>")},
				{terminal.Cyan("{{CurrentTime}}"), "Current time (ISO 8601)", terminal.Yellow("<ISO 8601 time>")},
				{terminal.Cyan("{{RandomString}}"), "Random 8-char alphanumeric", terminal.Yellow("<generated>")},
			}
			printMarkdownTable([]string{"Variable", "Description", "Default Value"}, builtinVars)
		} else {
			// Compact: show variables in columns
			vars := []string{
				"{{BaseFolder}}", "{{Binaries}}", "{{Data}}", "{{Workspaces}}",
				"{{Target}}", "{{Output}}", "{{RunUUID}}", "{{Today}}",
				"{{threads}}", "{{Version}}", "{{RandomString}}",
			}
			for i, v := range vars {
				fmt.Printf("  %s", terminal.Cyan(v))
				if (i+1)%4 == 0 {
					fmt.Println()
				}
			}
			fmt.Println()
			fmt.Println()
			fmt.Printf("  %s\n", terminal.Gray("Tip: Use --verbose to show all variables with descriptions and default values"))
		}

		fmt.Println()
		return nil
	},
}

// workflowValidateCmd validates a workflow
var workflowValidateCmd = &cobra.Command{
	Use:     "validate [name|path|folder]",
	Aliases: []string{"val", "lint", "fmt"},
	Short:   "Validate and lint workflow(s) - accepts workflow name, file path, or folder",
	Long: `Validate and lint workflow YAML file(s).

Accepts:
  - Workflow name (looks up in workflows directory)
  - Path to a YAML file
  - Path to a folder (recursively validates all workflow YAMLs)

The linter checks for:
  - Missing required fields (name, kind, type)
  - Undefined variables (referenced but not defined)
  - Unused variables (exported but never used)
  - Invalid goto/depends_on references
  - Circular dependencies
  - Empty steps
  - Duplicate step names

Examples:
  osmedeus workflow validate test-echo
  osmedeus workflow lint ./my-workflow.yaml
  osmedeus workflow validate /path/to/workflows/
  osmedeus workflow validate . --fail-fast
  osmedeus workflow lint my-workflow.yaml --check --format json
  osmedeus workflow validate . --disable unused-variable
  osmedeus workflow lint my-workflow.yaml --severity error`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()
		if cfg == nil {
			return fmt.Errorf("configuration not loaded")
		}

		printer := terminal.NewPrinter()
		input := args[0]

		inputType, resolvedPath := classifyInput(input)

		switch inputType {
		case "file":
			return lintFile(resolvedPath, cfg, printer)
		case "folder":
			return lintFolder(resolvedPath, cfg, printer, validateFailFast)
		case "name":
			return lintByName(input, cfg, printer)
		}

		return nil
	},
}

var showVerbose bool
var filterTags []string
var showYaml bool
var showTags bool
var validateFailFast bool

// Linter flags
var lintCheck bool
var lintFormat string
var lintDisable []string
var lintSeverity string

func init() {
	workflowShowCmd.Flags().BoolVarP(&showVerbose, "verbose", "v", false, "show detailed variable descriptions")
	workflowShowCmd.Flags().BoolVar(&showYaml, "yaml", false, "show raw YAML instead of table format")
	workflowListCmd.Flags().StringSliceVar(&filterTags, "tags", []string{}, "filter workflows by tags (comma-separated)")
	workflowListCmd.Flags().BoolVar(&showTags, "show-tags", false, "show tags column in output")
	workflowListCmd.Flags().BoolVarP(&showVerbose, "verbose", "v", false, "show workflows with errors")
	workflowValidateCmd.Flags().BoolVar(&validateFailFast, "fail-fast", false, "stop on first validation failure")
	workflowValidateCmd.Flags().BoolVar(&lintCheck, "check", false, "exit with error code if issues found (for CI)")
	workflowValidateCmd.Flags().StringVar(&lintFormat, "format", "pretty", "output format: pretty, json, github")
	workflowValidateCmd.Flags().StringSliceVar(&lintDisable, "disable", []string{}, "disable specific rules (comma-separated)")
	workflowValidateCmd.Flags().StringVar(&lintSeverity, "severity", "info", "minimum severity level: info, warning, error")
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.AddCommand(workflowShowCmd)
	workflowCmd.AddCommand(workflowValidateCmd)
}

// hasMatchingTags checks if a workflow has any of the specified tags
func hasMatchingTags(wf *core.Workflow, tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	for _, filterTag := range tags {
		for _, wfTag := range wf.Tags {
			if strings.EqualFold(wfTag, filterTag) {
				return true
			}
		}
	}
	return false
}

// classifyInput determines if input is a file path, folder path, or workflow name
func classifyInput(input string) (inputType string, resolvedPath string) {
	if info, err := os.Stat(input); err == nil {
		absPath, _ := filepath.Abs(input)
		if info.IsDir() {
			return "folder", absPath
		}
		return "file", absPath
	}

	// Check if it looks like a path
	if strings.Contains(input, string(filepath.Separator)) ||
		strings.Contains(input, "/") ||
		strings.HasSuffix(input, ".yaml") ||
		strings.HasSuffix(input, ".yml") {
		absPath, _ := filepath.Abs(input)
		return "file", absPath
	}

	return "name", input
}

// isWorkflowYAML checks if file contains kind: module or kind: flow
func isWorkflowYAML(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "kind:") {
		return false
	}

	kindPattern := regexp.MustCompile(`(?m)^kind:\s*['"]?(module|flow)['"]?\s*$`)
	return kindPattern.Match(content)
}

// findWorkflowFiles recursively finds workflow YAML files in directory
func findWorkflowFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible files
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}
		if isWorkflowYAML(path) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// createLinter creates a linter with the current CLI options
func createLinter() *linter.Linter {
	opts := linter.LinterOptions{
		DisabledRules: lintDisable,
		MinSeverity:   linter.ParseSeverity(lintSeverity),
	}
	return linter.NewLinter(opts)
}

// lintByName lints a workflow by name
func lintByName(name string, cfg *config.Config, printer *terminal.Printer) error {
	loader := parser.NewLoader(cfg.WorkflowsPath)
	workflow, err := loader.LoadWorkflow(name)
	if err != nil {
		fmt.Println()
		printer.Error("Failed to load workflow: %s", name)
		fmt.Println()
		fmt.Println(err.Error())
		fmt.Println()
		return err
	}

	// Use the FilePath from the parsed workflow
	if workflow.FilePath == "" {
		return fmt.Errorf("could not determine workflow file path for: %s", name)
	}

	return lintWorkflowFile(workflow.FilePath, workflow, printer)
}

// lintFile lints a single workflow file
func lintFile(path string, _ *config.Config, printer *terminal.Printer) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	if !isWorkflowYAML(path) {
		printer.Warning("File does not contain 'kind: module' or 'kind: flow': %s", path)
		return fmt.Errorf("not a workflow file")
	}

	// Parse the workflow first
	p := parser.NewParser()
	workflow, err := p.Parse(path)
	if err != nil {
		printer.Error("Parse error: %s", err)
		return err
	}

	return lintWorkflowFile(path, workflow, printer)
}

// lintWorkflowFile lints a parsed workflow file
func lintWorkflowFile(path string, workflow *core.Workflow, printer *terminal.Printer) error {
	// Read source for context display
	source, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create and run linter
	l := createLinter()
	result, err := l.LintContent(source, path)
	if err != nil {
		printer.Error("Lint error: %s", err)
		return err
	}

	// Format and print results
	format := linter.ParseOutputFormat(lintFormat)
	formatter := linter.GetFormatter(format, true)

	if result.HasIssues() {
		output := formatter.Format(result, source)
		fmt.Println()
		fmt.Print(output)
		fmt.Println(formatter.FormatSummary([]*linter.LintResult{result}))
		fmt.Println()
	} else {
		printer.Success("Workflow '%s' (%s) passed all lint checks", workflow.Name, workflow.Kind)
	}

	// Return error if --check mode and errors found
	if lintCheck && result.HasErrors() {
		return fmt.Errorf("lint check failed with %d error(s)", result.Errors)
	}

	return nil
}

// lintFolder lints all workflow files in a folder
func lintFolder(dir string, _ *config.Config, printer *terminal.Printer, failFast bool) error {
	fmt.Println()
	printer.Info("Scanning for workflow files in: %s", terminal.Cyan(dir))

	files, err := findWorkflowFiles(dir)
	if err != nil {
		return fmt.Errorf("failed to scan directory: %w", err)
	}

	if len(files) == 0 {
		printer.Warning("No workflow YAML files found in %s", dir)
		return nil
	}

	printer.Info("Found %d workflow file(s)", len(files))
	fmt.Println()

	l := createLinter()
	format := linter.ParseOutputFormat(lintFormat)
	formatter := linter.GetFormatter(format, true)

	var results []*linter.LintResult
	var hasErrors bool

	for _, file := range files {
		source, readErr := os.ReadFile(file)
		if readErr != nil {
			printer.Warning("Could not read %s: %s", file, readErr)
			continue
		}

		result, lintErr := l.LintContent(source, file)
		if lintErr != nil {
			// Parse errors become error issues
			result = &linter.LintResult{
				FilePath: file,
				Issues: []linter.LintIssue{{
					Rule:     "parse-error",
					Severity: linter.SeverityError,
					Message:  lintErr.Error(),
					Line:     1,
					Column:   1,
				}},
				Errors: 1,
			}
		}

		results = append(results, result)

		if result.HasIssues() {
			output := formatter.Format(result, source)
			fmt.Print(output)
		}

		if result.HasErrors() {
			hasErrors = true
			if failFast {
				fmt.Println()
				fmt.Println(formatter.FormatSummary(results))
				return fmt.Errorf("lint check failed")
			}
		}
	}

	fmt.Println()
	fmt.Println(formatter.FormatSummary(results))
	fmt.Println()

	// Summary stats
	totalErrors := linter.TotalErrors(results)
	totalWarnings := linter.TotalWarnings(results)
	filesWithIssues := 0
	for _, r := range results {
		if r.HasIssues() {
			filesWithIssues++
		}
	}

	printer.Info("Linted %d file(s): %s errors, %s warnings in %d file(s)",
		len(files),
		terminal.Red(fmt.Sprintf("%d", totalErrors)),
		terminal.Yellow(fmt.Sprintf("%d", totalWarnings)),
		filesWithIssues)

	if lintCheck && hasErrors {
		return fmt.Errorf("lint check failed with %d error(s)", totalErrors)
	}

	return nil
}
