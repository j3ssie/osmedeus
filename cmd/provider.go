package cmd

import (
	"fmt"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/distribute"
	"github.com/j3ssie/osmedeus/provider"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/panjf2000/ants"
	"github.com/spf13/cobra"
	"os"
	"sync"
)

func init() {
	var providerCmd = &cobra.Command{
		Use:   "provider",
		Short: "Cloud utils for Distributed Mode",
		Long:  core.Banner(),
		RunE:  runProvider,
	}

	providerCmd.PersistentFlags().StringVar(&options.Cloud.RawCommand, "cmd", "", "raw command")
	providerCmd.PersistentFlags().StringVar(&options.Cloud.CloudWait, "wait", "30m", "timeout to wait before next queue check")
	providerCmd.PersistentFlags().BoolVar(&options.Cloud.CheckingLimit, "check", false, "Only check for limit of config")
	providerCmd.PersistentFlags().StringVar(&options.Cloud.InstanceName, "name", "", "override instance name")
	providerCmd.PersistentFlags().BoolVar(&options.Cloud.BackgroundRun, "bg", false, "Send command to instance and run it in background")
	providerCmd.PersistentFlags().BoolVar(&options.Cloud.IgnoreConfigFile, "ic", false, "Ignore token in the config file")
	providerCmd.PersistentFlags().IntVar(&options.Cloud.Retry, "retry", 8, "Number of retry when command is error")
	providerCmd.PersistentFlags().StringSlice("id", []string{}, "Instance IDs that will be delete")

	var providerWizard = &cobra.Command{
		Use:     "wizard",
		Aliases: []string{"wi", "wiz", "wizazrd"},
		Short:   "Start a cloud config wizard",
		Long:    core.Banner(),
		RunE:    runCloudInit,
	}
	providerWizard.PersistentFlags().BoolVar(&options.Cloud.AddNewProvider, "add", false, "Open wizard to add new provider only")

	var providerBuild = &cobra.Command{
		Use:   "build",
		Short: "Build snapshot image",
		Long:  core.Banner(),
		RunE:  runProviderBuild,
	}
	var providerCreate = &cobra.Command{
		Use:   "create",
		Short: "Create cloud instance based on image",
		Long:  core.Banner(),
		RunE:  runProviderCreate,
	}

	var providerHealth = &cobra.Command{
		Use:   "health",
		Short: "Run a health check on running cloud instances",
		Long:  core.Banner(),
		RunE:  runCloudHealth,
	}

	var providerValidate = &cobra.Command{
		Use:   "validate",
		Short: "Run validate of the existing cloud configs",
		Long:  core.Banner(),
		RunE:  runProviderValidate,
	}
	var providerList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all instances",
		Long:    core.Banner(),
		RunE:    runProviderListing,
	}
	var providerDel = &cobra.Command{
		Use:     "delete",
		Aliases: []string{"del"},
		Short:   "Delete instances",
		Long:    core.Banner(),
		RunE:    runProviderDelete,
	}
	providerCmd.AddCommand(providerWizard)
	providerCmd.AddCommand(providerList)
	providerCmd.AddCommand(providerDel)
	providerCmd.AddCommand(providerValidate)
	providerCmd.AddCommand(providerHealth)
	providerCmd.AddCommand(providerCreate)
	providerCmd.AddCommand(providerBuild)
	providerCmd.SetHelpFunc(CloudHelp)
	RootCmd.AddCommand(providerCmd)
}

func runCloudHealth(_ *cobra.Command, _ []string) error {
	DBInit()
	distribute.CheckingCloudInstance(options)
	return nil
}

func runCloudInit(_ *cobra.Command, _ []string) error {
	DBInit()
	// interactive mode to show config file here
	distribute.InitCloudSetup(options)
	return nil
}

func runProvider(_ *cobra.Command, args []string) error {
	DBInit()
	if len(args) == 0 {
		fmt.Println(CloudUsage())
	}
	return nil
}

func runProviderBuild(_ *cobra.Command, _ []string) error {
	options.Cloud.OnlyCreateDroplet = true
	options.Cloud.ReBuildBaseImage = true

	// building multiple tokens
	options.Cloud.TokensFile = utils.NormalizePath(options.Cloud.TokensFile)
	if options.Cloud.TokensFile != "" {
		tokens := utils.ReadingFileUnique(options.Cloud.TokensFile)
		if len(tokens) == 0 {
			utils.ErrorF("token file not found: %v", options.Cloud.TokensFile)
			return nil
		}

		var wg sync.WaitGroup
		p, _ := ants.NewPoolWithFunc(options.Concurrency, func(i interface{}) {
			lOptions := options
			lOptions.Cloud.Token = i.(string)

			distribute.InitCloud(lOptions, lOptions.Scan.Inputs)
			wg.Done()
		}, ants.WithPreAlloc(true))
		defer p.Release()

		for _, token := range tokens {
			wg.Add(1)
			_ = p.Invoke(token)
		}
		wg.Wait()
		return nil

	}

	distribute.InitCloud(options, options.Scan.Inputs)
	return nil
}

func runProviderCreate(_ *cobra.Command, _ []string) error {
	options.Cloud.OnlyCreateDroplet = true
	if len(options.Scan.Inputs) == 0 {
		options.Scan.Inputs = append(options.Scan.Inputs, utils.RandomString(4))
	}

	distribute.InitCloud(options, options.Scan.Inputs)
	return nil
}

func runProviderValidate(_ *cobra.Command, _ []string) error {
	cloudValidate()
	return nil
}

func runProviderListing(_ *cobra.Command, _ []string) error {
	cloudRunners := distribute.GetClouds(options)
	cloudListing(cloudRunners)
	return nil
}

func runProviderDelete(cmd *cobra.Command, _ []string) error {
	cloudRunners := distribute.GetClouds(options)
	InstanceIDs, _ := cmd.Flags().GetStringSlice("id")

	for _, InstanceID := range InstanceIDs {
		for _, cloudRunner := range cloudRunners {
			cloudRunner.Provider.DeleteInstance(InstanceID)
		}
	}

	cloudListing(cloudRunners)
	return nil
}

func cloudListing(cloudRunners []distribute.CloudRunner) {
	var content [][]string
	for _, cloudRunner := range cloudRunners {
		cloudRunner.Provider.Action(provider.ListInstance)
		for _, instance := range cloudRunner.Provider.Instances {
			row := []string{
				cloudRunner.Provider.ProviderName,
				cloudRunner.Provider.RedactedToken,
				instance.InstanceID,
				instance.InstanceName,
				instance.IPAddress,
			}
			content = append(content, row)
		}
	}
	table := tablewriter.NewWriter(os.Stderr)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Provider", "Token", "Instance ID", "Instance Name", "IP Address"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(content) // Add Bulk Data
	table.Render()
}

func cloudValidate() {
	cloudRunners := distribute.GetClouds(options)

	var content [][]string
	for _, cloudRunner := range cloudRunners {
		row := []string{
			cloudRunner.Provider.ProviderName,
			cloudRunner.Provider.RedactedToken,
			cloudRunner.Provider.SSHKeyID,
			cloudRunner.Provider.SnapshotID,
		}
		content = append(content, row)
	}
	table := tablewriter.NewWriter(os.Stderr)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Provider", "Token", "SSH Key ID", "Osmedeus Snapshot ID"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(content) // Add Bulk Data
	table.Render()

}
