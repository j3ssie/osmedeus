package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall"

	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/executor"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"github.com/spf13/cobra"
)

// runStatusCmd - show status of a specific run
var runStatusCmd = &cobra.Command{
	Use:   "status <run-uuid>",
	Short: "Show status of a workflow run",
	Long:  UsageRunStatus(),
	Args:  cobra.ExactArgs(1),
	RunE:  runRunStatus,
}

// runCancelCmd - cancel a running workflow
var runCancelCmd = &cobra.Command{
	Use:   "cancel <run-uuid>",
	Short: "Cancel a running workflow",
	Long:  UsageRunCancel(),
	Args:  cobra.ExactArgs(1),
	RunE:  runRunCancel,
}

func init() {
	runCmd.AddCommand(runStatusCmd)
	runCmd.AddCommand(runCancelCmd)
}

func runRunStatus(cmd *cobra.Command, args []string) error {
	if err := connectDB(); err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	runUUID := args[0]
	ctx := context.Background()

	run, err := database.GetRunByID(ctx, runUUID, false, false)
	if err != nil {
		return fmt.Errorf("run not found: %w", err)
	}

	if globalJSON {
		jsonBytes, err := json.Marshal(run)
		if err != nil {
			return fmt.Errorf("failed to format run: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	// Key-value display
	printer := terminal.NewPrinter()
	printer.Info("Run Details")
	fmt.Println()

	printField("run_uuid", run.RunUUID)
	printField("workflow", run.WorkflowName)
	printField("kind", run.WorkflowKind)
	printField("target", run.Target)
	printField("workspace", run.Workspace)
	printField("status", terminal.ColorizeStatus(run.Status))
	printField("progress", fmt.Sprintf("%d/%d steps", run.CompletedSteps, run.TotalSteps))
	printField("trigger_type", run.TriggerType)
	if run.StartedAt != nil {
		printField("started_at", run.StartedAt.String())
	}
	if run.CompletedAt != nil {
		printField("completed_at", run.CompletedAt.String())
	}
	if run.ErrorMessage != "" {
		printField("error", run.ErrorMessage)
	}
	if run.CurrentPID > 0 {
		printField("pid", fmt.Sprintf("%d", run.CurrentPID))
	}

	return nil
}

// printField prints a key-value pair with aligned formatting.
func printField(key, value string) {
	fmt.Printf("  %-16s %s\n", key+":", value)
}

func runRunCancel(cmd *cobra.Command, args []string) error {
	if err := connectDB(); err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	runUUID := args[0]
	ctx := context.Background()

	run, err := database.GetRunByID(ctx, runUUID, false, false)
	if err != nil {
		return fmt.Errorf("run not found: %w", err)
	}

	if run.Status != "pending" && run.Status != "running" {
		return fmt.Errorf("cannot cancel run with status '%s'", run.Status)
	}

	var killedPIDs []int
	var killMethod string

	// Try to cancel via control plane first
	controlPlane := executor.GetRunControlPlane()
	controlPlanePIDs, controlPlaneErr := controlPlane.Cancel(run.RunUUID)

	if controlPlaneErr == nil && len(controlPlanePIDs) > 0 {
		killedPIDs = controlPlanePIDs
		killMethod = "control_plane"
	} else if run.CurrentPID > 0 {
		if cliKillProcessAndChildren(run.CurrentPID) {
			killedPIDs = []int{run.CurrentPID}
			killMethod = "database_pid"
		}
	}

	// Update database status
	if err := database.UpdateRunStatus(ctx, run.RunUUID, "cancelled", "Cancelled via CLI"); err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	if globalJSON {
		response := map[string]interface{}{
			"message":     "Run cancelled successfully",
			"run_uuid":    run.RunUUID,
			"killed_pids": killedPIDs,
			"kill_method": killMethod,
		}
		jsonBytes, err := json.Marshal(response)
		if err != nil {
			return fmt.Errorf("failed to format response: %w", err)
		}
		fmt.Println(string(jsonBytes))
		return nil
	}

	printer := terminal.NewPrinter()
	printer.Success("Run %s cancelled", run.RunUUID)
	if len(killedPIDs) > 0 {
		fmt.Printf("  Terminated %d process(es) via %s\n", len(killedPIDs), killMethod)
	} else {
		fmt.Println("  No active processes found to terminate")
	}

	return nil
}

// cliKillProcessAndChildren kills a process and all its children using SIGKILL.
// Returns true if the kill signal was sent successfully.
func cliKillProcessAndChildren(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Try to kill the process group first (negative PID kills all processes in the group)
	err := syscall.Kill(-pid, syscall.SIGKILL)
	if err != nil {
		// Process group kill failed, try killing just the process
		err = syscall.Kill(pid, syscall.SIGKILL)
		if err != nil {
			return false
		}
	}
	return true
}
