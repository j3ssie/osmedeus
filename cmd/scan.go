package cmd

import (
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/panjf2000/ants"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func init() {
	var scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Conduct a scan following a predetermined flow/module",
		Long:  core.Banner(),
		RunE:  runScan,
	}

	scanCmd.SetHelpFunc(ScanHelp)
	RootCmd.AddCommand(scanCmd)
}

func runScan(_ *cobra.Command, _ []string) error {
	DBInit()
	utils.GoodF("%v %v by %v", cases.Title(language.Und, cases.NoLower).String(libs.BINARY), libs.VERSION, color.HiMagentaString(libs.AUTHOR))
	utils.InforF("Storing the log file to: %v", color.CyanString(options.LogFile))

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
	if core.IsRootDomain(target) && options.Scan.Flow == "general" && len(options.Scan.Modules) == 0 {
		utils.WarnF("looks like you scanning a subdomain '%s' with general flow. The result might be much less than usual", color.HiCyanString(target))
		utils.WarnF("Better input should be root domain with TLD like '-t target.com'")
	}

	runner, err := core.InitRunner(target, options)
	if err != nil {
		utils.ErrorF("Error init runner with: %s", target)
		return
	}
	runner.Start()
}
