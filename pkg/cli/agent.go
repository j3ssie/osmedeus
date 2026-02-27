package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

var (
	agentName    string
	agentCwd     string
	agentStdin   bool
	agentTimeout string
	agentList    bool
)

// agentCmd runs an ACP agent interactively from the terminal.
var agentCmd = &cobra.Command{
	Use:   "agent [message]",
	Short: "Run an ACP agent interactively",
	Long:  UsageAgent(),
	RunE:  runAgent,
}

func init() {
	agentCmd.Flags().StringVar(&agentName, "agent", "claude-code", "agent to use (see --list for available agents)")
	agentCmd.Flags().StringVar(&agentCwd, "cwd", "", "working directory for the agent (default: current directory)")
	agentCmd.Flags().BoolVar(&agentStdin, "stdin", false, "read message from stdin")
	agentCmd.Flags().StringVar(&agentTimeout, "timeout", "30m", "timeout duration (e.g., 30m, 1h)")
	agentCmd.Flags().BoolVar(&agentList, "list", false, "list available agents")
}

func runAgent(cmd *cobra.Command, args []string) error {
	printer := terminal.NewPrinter()

	// List agents
	if agentList {
		names := executor.ListAgentNames()
		sort.Strings(names)
		printer.Section("Available ACP Agents")
		fmt.Println()
		for _, name := range names {
			fmt.Printf("  %s %s\n", terminal.SymbolBullet, terminal.Cyan(name))
		}
		fmt.Println()
		return nil
	}

	// Resolve message
	message, err := resolveAgentMessage(args)
	if err != nil {
		printer.Error("%s", err)
		return err
	}

	// Parse timeout
	timeout, err := parseRunDuration(agentTimeout)
	if err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Build config
	cfg := &executor.RunAgentACPConfig{
		Cwd:          agentCwd,
		StreamWriter: os.Stdout,
	}

	output, _, err := executor.RunAgentACP(ctx, message, agentName, cfg)
	if err != nil {
		printer.Error("Agent failed: %s", err)
		return err
	}

	// If there was no stream writer, print output at the end
	if cfg.StreamWriter == nil && output != "" {
		fmt.Println(output)
	}

	return nil
}

// resolveAgentMessage determines the message from positional args, --stdin, or piped stdin.
func resolveAgentMessage(args []string) (string, error) {
	// Positional argument (not "-")
	if len(args) > 0 && args[0] != "-" {
		return strings.Join(args, " "), nil
	}

	// --stdin flag or "-" argument
	if agentStdin || (len(args) > 0 && args[0] == "-") {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("failed to read from stdin: %w", err)
		}
		msg := strings.TrimSpace(string(data))
		if msg == "" {
			return "", fmt.Errorf("empty message from stdin")
		}
		return msg, nil
	}

	return "", fmt.Errorf("no message provided: use positional argument, --stdin, or pipe with -")
}
