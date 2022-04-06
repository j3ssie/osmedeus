package cmd

import (
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/spf13/cobra"
    "os"
)

func init() {
    var execCmd = &cobra.Command{
        Use:   "exec",
        Short: "Execute inline osmedeus scripts",
        Long:  core.Banner(),
        RunE:  runExec,
    }

    execCmd.Flags().String("script", "", "Scripts to run (Multiple -s flags are accepted)")
    execCmd.Flags().StringP("scriptFile", "S", "", "File contain list of scripts")
    RootCmd.AddCommand(execCmd)
}

func runExec(cmd *cobra.Command, _ []string) error {
    script, _ := cmd.Flags().GetString("script")
    scriptFile, _ := cmd.Flags().GetString("scriptFile")

    var scripts []string
    if script != "" {
        scripts = append(scripts, script)
    }
    if scriptFile != "" {
        moreScripts := utils.ReadingFileUnique(scriptFile)
        if len(moreScripts) > 0 {
            scripts = append(scripts, moreScripts...)
        }
    }

    if len(scripts) == 0 {
        utils.ErrorF("No scripts provided")
        os.Exit(0)
    }
    runner, _ := core.InitRunner("example.com", options)

    for _, t := range options.Scan.Inputs {
        // start to run scripts
        options.Scan.ROptions = core.ParseInput(t, options)
        for _, rscript := range scripts {
            script = core.ResolveData(rscript, options.Scan.ROptions)
            utils.InforF("Script: %v", script)
            runner.RunScript(script)
        }
    }

    return nil
}
