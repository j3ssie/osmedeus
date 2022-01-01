package cmd

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/execution"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/spf13/cobra"
    "path"
)

func init() {
    var execCmd = &cobra.Command{
        Use:     "health",
        Aliases: []string{"hea", "heal", "health", "healht"},
        Short:   "Run diagnostics to check configurations",
        Long:    core.Banner(),
        RunE:    runHealth,
    }
    RootCmd.AddCommand(execCmd)
}

func runHealth(_ *cobra.Command, _ []string) error {
    if options.PremiumPackage {
        fmt.Printf("💠 Osmedeus %s: Run diagnostics to check if everything okay\n", libs.VERSION)
    } else {
        fmt.Printf("🚀 Osmedeus %s: Run diagnostics to check if everything okay\n", libs.VERSION)
    }

    err := checkCorePrograms(options)
    if err != nil {
        fmt.Printf("‼️ There is might be something wrong with your setup: %v\n", err)
        return nil
    }

    err = listFlows(options)
    if err != nil {
        fmt.Printf("‼️ There is might be something wrong with your setup: %v\n", err)
        return nil
    }

    fmt.Printf(color.GreenString("\n🦾 It’s all good. Happy Hacking 🦾\n"))
    return nil
}

func checkCorePrograms(options libs.Options) error {
    exist := utils.FolderExists(options.Env.BaseFolder)
    if !exist {
        color.Red("[-] Core folder setup incorrect: %v", options.Env.BaseFolder)
        return fmt.Errorf("error running diagnostics")
    }

    // check core programs
    var err error
    _, err = utils.RunCommandWithErr("jaeles -h")
    if err != nil {
        color.Red("[-] Core program setup incorrectly")
        return fmt.Errorf("error checking core programs: %v", "jaeles")

    }
    _, err = utils.RunCommandWithErr("amass -h")
    if err != nil {
        color.Red("[-] Core program setup incorrectly")
        return fmt.Errorf("error checking core programs: %v", "amass")

    }
    _, err = utils.RunCommandWithErr(fmt.Sprintf("%s -h", path.Join(options.Env.BinariesFolder, "httprobe")))
    if err != nil {
        color.Red("[-] Core program setup incorrectly")
        return fmt.Errorf("error checking core programs: %v", fmt.Sprintf("%s -h", path.Join(options.Env.BinariesFolder, "httprobe")))

    }
    fmt.Printf("[+] Health Check Core Programs: %s\n", color.GreenString("✔"))

    // Check core signatures
    okVuln := false
    if utils.DirLength("~/.jaeles/base-signatures/") > 0 || utils.DirLength("~/pro-signatures/") > 0 {
        okVuln = true
    }

    if utils.DirLength("~/nuclei-templates") > 0 {
        okVuln = true
    }

    if okVuln {
        fmt.Printf("[+] Health Check Vulnerability scanning config: %s\n", color.GreenString("✔"))
    } else {
        color.Red("vulnerability scanning config setup incorrectly")
        return fmt.Errorf("vulnerability scanning config setup incorrectly")
    }

    // check data folder
    if utils.FolderExists(options.Env.DataFolder) {
        fmt.Printf("[+] Health Check Data Config: %s\n", color.GreenString("✔"))
    } else {
        color.Red("[-] Data setup incorrectly: %v", options.Env.DataFolder)
        return fmt.Errorf("[-] Data setup incorrectly: %v", options.Env.DataFolder)
    }

    // check cloud config
    var okCloud bool
    if utils.FileExists(path.Join(options.Env.CloudConfigFolder, "config.yaml")) {
        okCloud = true
        if utils.DirLength(path.Join(options.Env.CloudConfigFolder, "providers")) == 0 {
            okCloud = false
        }
        if utils.DirLength(path.Join(options.Env.CloudConfigFolder, "ssh")) < 2 {
            okCloud = false
        }
        if okCloud {
            fmt.Printf("[+] Health Check Cloud Config: %s\n", color.GreenString("✔"))
        } else {
            fmt.Printf(color.YellowString("%s If you install osmedeus on a single machine then it's okay to ignore the cloud setup\n", "[!] Cloud config setup incorrectly."))
        }

    }

    if execution.ValidGitURL(options.Storages["summary_repo"]) {
        if utils.DirLength(options.Storages["summary_storage"]) > 1 {
            fmt.Printf("[+] Health Check Storages Config: %s\n", color.GreenString("✔"))
        }
    }

    return nil
}

func listFlows(options libs.Options) error {
    flows := core.ListFlow(options)
    if len(flows) == 0 {
        color.Red("[-] Error to list workflows: %s", options.Env.WorkFlowsFolder)
        return fmt.Errorf("[-] Error to list workflows: %s", options.Env.WorkFlowsFolder)
    }
    fmt.Printf("[+] Health Check Workflows: %s\n", color.GreenString("✔"))

    fmt.Printf("\nChecking available workflow at: %s \n\n", color.HiBlueString(options.Env.WorkFlowsFolder))
    for _, flow := range flows {
        parsedFlow, err := core.ParseFlow(flow)
        if err != nil {
            utils.ErrorF("Error parsing flow: %v", flow)
        }
        fmt.Printf("%10s - %s\n", parsedFlow.Name, parsedFlow.Desc)
    }
    h := "\nUsage:\n"
    h += "  osmedeus scan -f [flowName] -t [target] \n"
    fmt.Printf(h)
    return nil
}
