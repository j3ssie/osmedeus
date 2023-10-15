package cmd

import (
	"os"

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
	updateCmd.Flags().Bool("F", false, "Shortcut for force update and clean old data at the same time")
	updateCmd.Flags().BoolVar(&options.Update.ForceUpdate, "force", false, "Force update")
	updateCmd.Flags().BoolVar(&options.Update.CleanOldData, "clean", false, "Clean up old Data")
	updateCmd.Flags().BoolVar(&options.Update.VulnUpdate, "vuln", false, "Update Vulnerability Database only")
	updateCmd.Flags().StringVar(&options.Update.UpdateURL, "update-url", "", "The script URL to download update")
	// generate update meta data
	updateCmd.Flags().StringVar(&options.Update.GenerateMeta, "gen", "", "Generate metadata for update")
	RootCmd.AddCommand(updateCmd)
	updateCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if options.FullHelp {
			cmd.Help()
			os.Exit(0)
		}
	}
}

func runUpdate(cmd *cobra.Command, _ []string) error {
	meta, _ := cmd.Flags().GetString("meta")
	if meta != "" {
		options.Update.MetaDataURL = meta
	}

	forcedUpdateandClean, _ := cmd.Flags().GetBool("F")
	if forcedUpdateandClean {
		options.Update.ForceUpdate = true
		options.Update.CleanOldData = true
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
	if options.Update.UpdateURL == "" {
		options.Update.UpdateURL = core.GetUpdateURL(options)
	}

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
