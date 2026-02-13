package cli

import (
	"fmt"

	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

// scanCmd is an alias for runCmd (backward compatibility)
var scanCmd = &cobra.Command{
	Use:     "scan",
	Aliases: []string{"execute"},
	Short:   "Execute a workflow (alias for 'run')",
	Long:    UsageRun(),
	RunE: func(cmd *cobra.Command, args []string) error {
		printer := terminal.NewPrinter()
		tip := "Tip: 'osmedeus scan' and 'osmedeus execute' are aliases for 'osmedeus run'"
		if terminal.IsCIMode() {
			printer.Info("%s", tip)
		} else {
			printer.Info("%s", terminal.Gray(tip))
		}
		fmt.Println()
		return runRun(cmd, args)
	},
}

func init() {
	// Copy all flags from runCmd to scanCmd for backward compatibility
	scanCmd.Flags().StringVarP(&flowName, "flow", "f", "", "flow workflow name to execute")
	scanCmd.Flags().StringArrayVarP(&moduleNames, "module", "m", nil, "module workflow(s) to execute (can specify multiple)")
	scanCmd.Flags().StringArrayVarP(&targets, "target", "t", nil, "target(s) to run against (can be specified multiple times)")
	scanCmd.Flags().StringVarP(&targetFile, "target-file", "T", "", "file containing targets (one per line)")
	scanCmd.Flags().StringArrayVarP(&paramFlags, "params", "p", nil, "additional parameters (key=value format)")
	scanCmd.Flags().StringVarP(&paramsFile, "params-file", "P", "", "file containing parameters (JSON or YAML key:value pairs)")
	scanCmd.Flags().StringVarP(&workspacePath, "workspace", "w", "", "custom workspace path (overrides {{TargetSpace}})")
	scanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be executed without running commands")
	scanCmd.Flags().IntVar(&threadsHold, "threads-hold", 0, "override thread count (0 = use tactic default)")
	scanCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "number of targets to run concurrently")
	scanCmd.Flags().StringVarP(&runTactic, "tactic", "B", "default", "run tactic: aggressive, default, gently")
	scanCmd.Flags().StringArrayVarP(&excludeModules, "exclude", "x", nil, "module(s) to exclude from execution (can be specified multiple times)")
	scanCmd.Flags().StringArrayVarP(&fuzzyExcludeModules, "fuzzy-exclude", "X", nil, "exclude modules whose name contains the given substring (can be specified multiple times)")
	scanCmd.Flags().StringVarP(&spaceName, "space", "S", "", "override {{TargetSpace}} variable")
	scanCmd.Flags().StringVarP(&workspacesFolder, "workspaces-folder", "W", "", "override {{Workspaces}} variable")
	scanCmd.Flags().StringVar(&heuristicsCheck, "heuristics-check", "basic", "heuristics check level: none, basic, advanced")
	scanCmd.Flags().BoolVarP(&distributedRun, "distributed-run", "D", false, "submit run to distributed worker queue (requires Redis)")
	scanCmd.Flags().StringVar(&redisURLRun, "redis-url", "", "Redis connection URL for distributed mode (overrides settings)")
	scanCmd.Flags().BoolVar(&repeatRun, "repeat", false, "repeat run after completion")
	scanCmd.Flags().StringVar(&repeatWaitTime, "repeat-wait-time", "1m", "wait time between repeats (e.g., 30s, 20m, 10h, 1d)")
	scanCmd.Flags().StringVar(&runTimeout, "timeout", "", "run timeout (e.g., 2h, 3h, 1d)")
	scanCmd.Flags().BoolVar(&stdModule, "std-module", false, "read module YAML from stdin")
	scanCmd.Flags().BoolVar(&emptyTarget, "empty-target", false, "run without target (generates placeholder target)")
	scanCmd.Flags().BoolVarP(&progressBar, "progress-bar", "G", false, "show progress bar during execution (enables silent mode)")
}
