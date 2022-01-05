package cmd

import (
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/panjf2000/ants"
    "github.com/spf13/cobra"
    "strings"
    "sync"
)

func init() {
    var scanCmd = &cobra.Command{
        Use:   "scan",
        Short: "Do Scan based on predefined flow",
        Long:  core.Banner(),
        RunE:  runScan,
    }

    scanCmd.Flags().StringSliceVarP(&options.Scan.Modules, "module", "m", []string{}, "Target to running")
    scanCmd.Flags().StringVarP(&options.Scan.Flow, "flow", "f", "general", "Flow name for running (default: general)")
    scanCmd.Flags().StringVarP(&options.Scan.CustomWorkspace, "workspace", "w", "", "Name of workspace (default is same as target)")
    scanCmd.Flags().StringSliceVarP(&options.Scan.Params, "params", "p", []string{}, "Custom params -p='foo=bar' (Multiple -p flags are accepted)")
    scanCmd.SetHelpFunc(ScanHelp)
    RootCmd.AddCommand(scanCmd)
}

func runScan(_ *cobra.Command, _ []string) error {
    DBInit()
    utils.GoodF("%v %v by %v", strings.Title(libs.BINARY), libs.VERSION, libs.AUTHOR)
    utils.GoodF("Store log file to: %v", options.LogFile)

    var wg sync.WaitGroup
    p, _ := ants.NewPoolWithFunc(options.Concurrency, func(i interface{}) {
        // really start to scan
        CreateRunner(i)
        wg.Done()
    }, ants.WithPreAlloc(true))
    defer p.Release()

    if options.Cloud.EnableChunk {
        for _, target := range options.Scan.Inputs {
            chunkTargets := HandleChunksInputs(target)
            for _, chunkTarget := range chunkTargets {
                wg.Add(1)
                _ = p.Invoke(chunkTarget)
            }
        }
    } else {
        for _, target := range options.Scan.Inputs {
            wg.Add(1)
            _ = p.Invoke(strings.TrimSpace(target))
        }
    }

    wg.Wait()
    return nil
}

func CreateRunner(j interface{}) {
    target := j.(string)
    runner, err := core.InitRunner(target, options)
    if err != nil {
        utils.ErrorF("Error init runner with: %s", target)
        return
    }
    runner.Start()
}
