package cmd

import (
    "github.com/j3ssie/osmedeus/core"
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
    RootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, _ []string) error {
    meta, _ := cmd.Flags().GetString("meta")
    if meta != "" {
        options.Update.UpdateURL = meta
    }
    core.UpdateMetadata(&options)
    return nil
}
