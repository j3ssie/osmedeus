package cmd

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/execution"
    "github.com/j3ssie/osmedeus/libs"
    jsoniter "github.com/json-iterator/go"
    "github.com/spf13/cobra"
)

func init() {
    var utilsCmd = &cobra.Command{
        Use:   "utils",
        Short: "Utils to get some information from the system",
        Long:  core.Banner(),
        RunE:  runProvider,
    }

    var psCmd = &cobra.Command{
        Use:   "ps",
        Short: "Utility to get information about running process",
        Long:  core.Banner(),
        RunE:  runPs,
    }
    psCmd.Flags().StringSlice("proc", []string{}, "Process name")

    var tmuxCmd = &cobra.Command{
        Use:   "tmux",
        Short: "Utility to get info from tmux",
        Long:  core.Banner(),
        RunE:  runTmux,
    }

    tmuxCmd.Flags().BoolVarP(&options.Tmux.ApplyAll, "all", "A", false, "Apply for all tmux sessions")
    tmuxCmd.Flags().StringVarP(&options.Tmux.SelectedWindow, "name", "n", "", "Apply for all tmux sessions")
    tmuxCmd.Flags().StringVarP(&options.Tmux.Exclude, "exclude", "e", "server", "Exclude tmux session")
    tmuxCmd.Flags().IntVarP(&options.Tmux.Limit, "limit", "l", 0, "Size of output content")

    var cronCmd = &cobra.Command{
        Use:   "cron",
        Short: "Utility to run command schedule",
        Long:  core.Banner(),
        RunE:  runCron,
    }
    cronCmd.Flags().IntVar(&options.Cron.Schedule, "sch", 0, "Number of minutes to schedule the job")
    cronCmd.Flags().BoolVar(&options.Cron.Forever, "for", false, "Keep running forever right after the command done")
    cronCmd.Flags().StringVar(&options.Cron.Command, "cmd", "", "Command to run")

    // add command
    utilsCmd.AddCommand(cronCmd)
    utilsCmd.AddCommand(tmuxCmd)
    utilsCmd.AddCommand(psCmd)
    utilsCmd.SetHelpFunc(UtilsHelp)
    RootCmd.AddCommand(utilsCmd)
}

func runPs(cmd *cobra.Command, _ []string) error {
    processes, _ := cmd.Flags().GetStringSlice("process")
    if len(processes) == 0 {
        processes = append(processes, libs.BINARY)
    }

    for _, process := range processes {
        pss := execution.GetOsmProcess(process)
        for _, ps := range pss {
            if options.JsonOutput {
                if data, err := jsoniter.MarshalToString(ps); err == nil {
                    fmt.Println(data)
                }
                continue
            }
            fmt.Printf("pid:%v %s %v\n", color.HiCyanString("%v", ps.PID), color.HiMagentaString("--"), ps.Command)
        }
    }

    return nil
}

func runTmux(_ *cobra.Command, args []string) error {
    tmux, err := core.InitTmux(options)
    if err != nil {
        return err
    }

    for _, argument := range args {
        switch argument {
        case "l", "ls", "list":
            tmux.ListTmux()
        case "t", "log", "logs", "tai", "tail":
            tmux.CatchSession()
        }
    }
    return nil
}

func runCron(_ *cobra.Command, _ []string) error {
    if options.Cron.Schedule == 0 && options.Cron.Forever == false {
        return fmt.Errorf("missing '--sche' flag")
    }
    if options.Cron.Forever {
        options.Cron.Schedule = -1
    }
    core.RunCron(options.Cron.Command, options.Cron.Schedule)
    return nil
}
