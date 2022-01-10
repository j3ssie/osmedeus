package cmd

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/distribute"
    "github.com/j3ssie/osmedeus/provider"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/panjf2000/ants"
    "github.com/spf13/cobra"
    "sort"
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

    var providerBuild = &cobra.Command{
        Use:   "build",
        Short: "Build cloud image",
        Long:  core.Banner(),
        RunE:  runProviderBuild,
    }
    var providerCreate = &cobra.Command{
        Use:   "create",
        Short: "Create cloud instance based on image",
        Long:  core.Banner(),
        RunE:  runProviderCreate,
    }

    var healthCmd = &cobra.Command{
        Use:   "health",
        Short: "Cloud Utility to check cloud instance health",
        Long:  core.Banner(),
        RunE:  runCloudHealth,
    }

    var providerValidate = &cobra.Command{
        Use:   "validate",
        Short: "Run various action on cloud provider",
        Long:  core.Banner(),
        RunE:  runProviderValidate,
    }
    providerCmd.AddCommand(providerValidate)

    providerCmd.AddCommand(healthCmd)
    providerCmd.AddCommand(providerCreate)
    providerCmd.AddCommand(providerBuild)
    providerCmd.AddCommand(providerValidate)
    providerCmd.SetHelpFunc(CloudHelp)
    RootCmd.AddCommand(providerCmd)
}

func runCloudHealth(_ *cobra.Command, _ []string) error {
    DBInit()
    distribute.CheckingCloudInstance(options)
    return nil
}

func runProvider(_ *cobra.Command, args []string) error {
    DBInit()
    if len(args) == 0 {
        fmt.Println(CloudUsage())
    }
    err := checkCloud()
    if err != nil {
        fmt.Println(color.YellowString("⚠️️ There is might be something wrong with your cloud setup: %v\n", err))
        return err
    }

    return nil
}

func runProviderBuild(_ *cobra.Command, _ []string) error {
    options.Cloud.OnlyCreateDroplet = true
    options.Cloud.ReBuildBaseImage = true

    // building multiple tokens
    if options.Cloud.TokensFile != "" && utils.FileExists(options.Cloud.TokensFile) {
        tokens := utils.ReadingFileUnique(options.Cloud.TokensFile)
        if len(tokens) > 0 {
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
            return nil
        }
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

func runProviderValidate(_ *cobra.Command, actions []string) error {
    //options.Cloud.IgnoreSetup = true
    clouds := distribute.GetClouds(options)
    sort.Strings(actions)

    for _, cloud := range clouds {
        for _, action := range actions {
            err := cloud.Provider.Action(provider.ListInstance)
            if err != nil {
                utils.ErrorF("error running %v ", action)
            }
            for _, instance := range cloud.Provider.Instances {
                fmt.Printf("[%v]: %v -- %v", instance.ProviderName, instance.InstanceID, instance.IPAddress)
            }
        }
    }
    return nil
}
