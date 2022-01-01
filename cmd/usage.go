package cmd

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/spf13/cobra"
)

// RootUsage base help
func RootUsage() {
    var h string
    h += ScanUsage()
    h += CloudUsage()
    h += UtilsUsage()

    fmt.Println(h)
}

func ScanUsage() string {
    h := color.HiCyanString("\nScan Usage:\n")
    h += "  osmedeus scan -f [flowName] -t [target] \n"
    h += "  osmedeus scan -m [modulePath] -T [targetsFile] \n"
    h += "  osmedeus scan -f /path/to/flow.yaml -t [target] \n"
    h += "  osmedeus scan -m /path/to/module.yaml -t [target] --params 'port=9200'\n"
    h += "  osmedeus scan -m /path/to/module.yaml -t [target] -l /tmp/log.log\n"
    h += "  cat targets | osmedeus scan -f sample\n"

    h += color.HiCyanString("\nPractical Scan Usage:\n")
    h += "  osmedeus scan -T list_of_targets.txt -W custom_workspaces\n"
    h += "  osmedeus scan -t target.com -w workspace_name --debug\n"
    h += "  osmedeus scan -f general -t www.sample.com\n"
    h += "  osmedeus scan -f gdirb -T list_of_target.txt\n"
    h += "  osmedeus scan -m ~/.osmedeus/core/workflow/test/dirbscan.yaml -t list_of_urls.txt\n"

    return h
}

func UtilsUsage() string {
    h := color.HiCyanString("\nUtilities Usage:\n")
    h += "  osmedeus health \n"
    h += "  osmedeus version --json \n"
    h += "  osmedeus utils tmux ls \n"
    h += "  osmedeus utils tmux logs -A -l 10 \n"
    h += "  osmedeus utils ps \n"
    h += "  osmedeus utils ps --proc 'jaeles' \n"
    h += "  osmedeus utils cron --cmd 'osmdeus scan -t example.com' --sch 60\n"
    h += "  osmedeus utils cron --for --cmd 'osmedeus scan -t example.com'\n"
    return h
}

func ConfigUsage() string {
    h := color.HiCyanString("\nConfig Usage:\n")
    h += "  osmedeus config [action] [OPTIONS] \n"
    h += "  osmedeus config init -p https://github.com/j3ssie/osmedeus-plugins\n"
    h += "  osmedeus config --user newusser --pass newpassword\n"
    h += "  osmedeus config reload \n"
    h += "  osmedeus config update \n"
    h += "  osmedeus config clean \n"
    h += "  osmedeus config delete -t woskapce \n"
    h += "  osmedeus config delete -w workspace_name \n"
    return h
}

func CloudUsage() string {
    h := color.HiCyanString("\nProvider Usage:\n")
    h += "  osmedeus provider build \n"
    h += "  osmedeus provider build --token xxx --rebuild --ic\n"
    h += "  osmedeus provider create --name 'sample' \n"
    h += "  osmedeus provider health --debug \n"

    h += color.HiCyanString("\nCloud Usage:\n")
    h += "  osmedeus cloud -f [flowName] -t [target] \n"
    h += "  osmedeus cloud -m [modulePath] -t [target] \n"
    h += "  osmedeus cloud -c 10 -f [flowName] -T [targetsFile] \n"
    h += "  osmedeus cloud --token xxx -G -c 10 -f [flowName] -T [targetsFile] \n"
    h += "  osmedeus cloud --chunk -c 10 -f [flowName] -t [targetsFile] \n"

    return h
}

// ScanHelp scan help message
func ScanHelp(cmd *cobra.Command, _ []string) {
    fmt.Println(core.Banner())
    fmt.Println(cmd.UsageString())
    h := ScanUsage()
    fmt.Println(h)
    fmt.Printf("Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}

// CloudHelp scan help message
func CloudHelp(cmd *cobra.Command, _ []string) {
    fmt.Println(core.Banner())
    fmt.Println(cmd.UsageString())
    h := CloudUsage()
    fmt.Println(h)
    fmt.Printf("Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}

// ConfigHelp config help message
func ConfigHelp(cmd *cobra.Command, _ []string) {
    fmt.Println(core.Banner())
    fmt.Println(cmd.UsageString())
    h := ConfigUsage()
    fmt.Println(h)
    fmt.Printf("Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}

// UtilsHelp utils help message
func UtilsHelp(cmd *cobra.Command, _ []string) {
    fmt.Println(core.Banner())
    fmt.Println(cmd.UsageString())
    h := UtilsUsage()
    fmt.Println(h)
    fmt.Printf("Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}

// RootHelp print help message
func RootHelp(cmd *cobra.Command, _ []string) {
    fmt.Println(core.Banner())
    fmt.Println(cmd.UsageString())
    RootUsage()
    fmt.Printf("Documentation can be found here: %s\n", color.GreenString(libs.DOCS))
}
