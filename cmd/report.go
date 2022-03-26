package cmd

import (
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/spf13/cobra"
    "io/ioutil"
    "net/http"
)

func init() {
    var reportCmd = &cobra.Command{
        Use:   "report",
        Short: "Show report of existing workspace",
        Long:  core.Banner(),
        RunE:  runReport,
    }

    var lsCmd = &cobra.Command{
        Use:     "list",
        Aliases: []string{"ls"},
        Short:   "List all current existing workspace",
        Long:    core.Banner(),
        RunE:    runReportList,
    }
    reportCmd.AddCommand(lsCmd)

    var viewCmd = &cobra.Command{
        Use:   "view",
        Short: "View all reports of existing workspace",
        Long:  core.Banner(),
        RunE:  runReportView,
    }
    reportCmd.AddCommand(viewCmd)

    var extractCmd = &cobra.Command{
        Use:   "extract",
        Short: "Extract a compressed workspace",
        Long:  core.Banner(),
        RunE:  runReportExtract,
    }
    extractCmd.Flags().StringVar(&options.Report.WorkspaceFile, "ws", "", "Workspace file to extract")
    reportCmd.AddCommand(extractCmd)

    reportCmd.PersistentFlags().BoolVar(&options.Report.Raw, "raw", false, "Show all the file in the workspace")
    reportCmd.PersistentFlags().StringVar(&options.Report.PublicIP, "ip", "", "Show downloadable file with the given IP address")
    reportCmd.PersistentFlags().BoolVar(&options.Report.Static, "static", false, "Show report file with Prefix Static")

    reportCmd.SetHelpFunc(ReportHelp)
    RootCmd.AddCommand(reportCmd)
}

func runReportList(_ *cobra.Command, _ []string) error {
    core.ListWorkspaces(options)
    return nil
}

func runReportView(_ *cobra.Command, _ []string) error {
    if len(options.Scan.Inputs) == 0 {
        core.ListWorkspaces(options)
        utils.InforF("Please select workspace to view report. Try %s", color.HiCyanString(`'osmedeus report view -t target.com'`))
        return nil
    }

    for _, target := range options.Scan.Inputs {
        core.ListSingleWorkspace(options, target)
    }
    return nil
}

func runReportExtract(_ *cobra.Command, _ []string) error {
    core.ListWorkspaces(options)
    return nil
}

func runReport(_ *cobra.Command, _ []string) error {
    if options.Report.PublicIP == "" {
        if utils.GetOSEnv("IPAddress", "127.0.0.1") == "127.0.0.1" {
            options.Report.PublicIP = utils.GetOSEnv("IPAddress", "127.0.0.1")
        }
    }

    if options.Report.PublicIP == "0" || options.Report.PublicIP == "0.0.0.0" {
        options.Report.PublicIP = getPublicIP()
    }

    return nil
}

func getPublicIP() string {
    utils.DebugF("getting Public IP Address")
    req, err := http.Get("https://api.ipify.org")
    if err != nil {
        return "127.0.0.1"
    }
    defer req.Body.Close()

    body, err := ioutil.ReadAll(req.Body)
    if err != nil {
        return "127.0.0.1"
    }
    return string(body)
}
