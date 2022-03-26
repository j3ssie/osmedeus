package cmd

import (
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/spf13/cobra"
)

func init() {
    var updateCmd = &cobra.Command{
        Use:   "update",
        Short: "Check latest Update",
        Long:  core.Banner(),
        RunE:  runUpdate,
    }
    updateCmd.Flags().String("meta", "", "Custom MetaData URL")
    updateCmd.Flags().BoolVar(&options.Update.ForceUpdate, "force", false, "Force Update")
    updateCmd.Flags().BoolVar(&options.Update.CleanOldData, "clean", false, "Clean Old Data")
    updateCmd.Flags().BoolVar(&options.Update.VulnUpdate, "vuln", false, "Update Vulnerability Database only")
    // generate update meta data
    updateCmd.Flags().StringVar(&options.Update.GenerateMeta, "gen", "", "Generate metadata for update")
    RootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, _ []string) error {
    meta, _ := cmd.Flags().GetString("meta")
    if meta != "" {
        options.Update.MetaDataURL = meta
    }

    if options.Update.GenerateMeta != "" {
        core.GenerateMetaData(options)
        return nil
    }

    if options.Update.VulnUpdate {
        core.UpdateVuln(options)
        return nil
    }

    var shouldUpdate bool
    options.Update.UpdateURL = core.GetUpdateURL(options)

    if options.Update.ForceUpdate {
        shouldUpdate = true
        utils.InforF("Force to Update latest release")
    } else {
        shouldUpdate = core.CheckUpdate(&options)
    }

    if shouldUpdate {
        err := core.RunUpdate(options)
        if err != nil {
            return err
        }
    }

    return nil
}
