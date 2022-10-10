package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/execution"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
	"path"
	"sort"
)

func init() {
	var healthCmd = &cobra.Command{
		Use:     "health",
		Aliases: []string{"hea", "heal", "health", "healht"},
		Short:   "Run diagnostics to check configurations",
		Long:    core.Banner(),
		RunE:    runHealth,
	}
	RootCmd.AddCommand(healthCmd)
}

func runHealth(_ *cobra.Command, args []string) error {
	if options.PremiumPackage {
		fmt.Printf("💠 Osmedeus %s: Run diagnostics to check if everything okay\n", libs.VERSION)
	} else {
		fmt.Printf("🚀 Osmedeus %s: Run diagnostics to check if everything okay\n", libs.VERSION)
	}

	sort.Strings(args)
	var err error
	for _, arg := range args {
		switch arg {
		case "store", "git", "storages", "stora":
			err = checkStorages()
		case "cloud", "dist", "provider":
			err = checkCloud()
		case "all", "a", "full":
			err = checkStorages()
			if err != nil {
				fmt.Println(color.YellowString("⚠️️ There is might be something wrong with your storages: %v\n", err))
			}
			err = checkCloud()
			if err != nil {
				fmt.Println(color.YellowString("%s If you install osmedeus on a single machine then it's okay to ignore the cloud setup\n", "[!] Cloud config setup incorrectly."))
			}
			err = generalCheck()
			if err != nil {
				fmt.Printf("‼️ There is might be something wrong with your setup: %v\n", color.HiRedString("%v", err))
				return nil
			}
			break
		}
		if err != nil {
			fmt.Println(color.YellowString("⚠️️ There is might be something wrong with your cloud or storages setup: %v\n", err))
			return nil
		}
	}
	if len(args) > 0 {
		return nil
	}

	err = generalCheck()
	if err != nil {
		fmt.Printf("‼️ There is might be something wrong with your setup: %v\n", err)
		return nil
	}

	err = listFlows()
	if err != nil {
		fmt.Printf("‼️ There is might be something wrong with your setup: %v\n", err)
		return nil
	}
	fmt.Printf(color.GreenString("\n🦾 It’s all good. Happy Hacking 🦾\n"))
	return nil
}

func checkCloud() error {
	// check packer program
	if _, err := utils.RunCommandWithErr("packer -h"); err != nil {
		if _, err := utils.RunCommandWithErr(fmt.Sprintf("%s -h", path.Join(options.Env.BinariesFolder, "packer"))); err != nil {
			color.Red("[-] Packer program setup incorrectly")
			return fmt.Errorf("error checking core programs: %v", fmt.Sprintf("%s -h", path.Join(options.Env.BinariesFolder, "packer")))
		}
	}

	// check config files
	if !utils.FileExists(options.CloudConfigFile) {
		return fmt.Errorf("distributed cloud config doesn't exist: %v", path.Join(options.Env.CloudConfigFolder, "provider.yaml"))
	}
	if utils.DirLength(path.Join(options.Env.CloudConfigFolder, "providers")) == 0 {
		return fmt.Errorf("providers file doesn't exist: %v", path.Join(options.Env.CloudConfigFolder, "providers"))
	}

	// check SSH Keys
	if !utils.FileExists(options.Cloud.SecretKey) {
		keysDir := path.Dir(options.Cloud.SecretKey)
		os.RemoveAll(keysDir)
		utils.MakeDir(keysDir)
		utils.DebugF("Generate SSH Key at: %v", options.Cloud.SecretKey)
		if _, err := utils.RunCommandWithErr(fmt.Sprintf(`ssh-keygen -t ed25519 -f %s -q -N ''`, options.Cloud.SecretKey)); err != nil {
			color.Red("[-] error generated SSH Key for cloud config at: %v", options.Cloud.SecretKey)
			return fmt.Errorf("[-] error generated SSH Key for cloud config at: %v", options.Cloud.SecretKey)
		}
	}
	if !utils.FileExists(options.Cloud.PublicKey) {
		return fmt.Errorf("providers SSH Key missing: %v", options.Cloud.PublicKey)
	}

	fmt.Printf("[+] Health Check Cloud Config: %s\n", color.GreenString("✔"))
	return nil
}

func checkStorages() error {
	utils.DebugF("Checking storages setup")
	if !execution.ValidGitURL(options.Storages["summary_repo"]) {
		return fmt.Errorf("invalid git summary: %v", options.Storages["summary_repo"])
	}

	utils.DebugF("Check if your summary directory is exist or not: %v", options.Env.StoragesFolder)
	if utils.DirLength(options.Env.StoragesFolder) < 1 {
		return fmt.Errorf("storages folder doesn't exist: %v", options.Env.StoragesFolder)
	}

	utils.DebugF("Check the secret key for git usage: %v", options.Storages["secret_key"])
	if !utils.FileExists(options.Storages["secret_key"]) {
		return fmt.Errorf("secret key for git command doesn't exist: %v", options.Storages["secret_key"])
	}

	fmt.Printf("[+] Health Check Storages Config: %s\n", color.GreenString("✔"))
	return nil
}

func generalCheck() error {
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
	return nil
}

func listFlows() error {
	flows := core.ListFlow(options)
	if len(flows) == 0 {
		color.Red("[-] Error to list workflows: %s", options.Env.WorkFlowsFolder)
		return fmt.Errorf("[-] Error to list workflows: %s", options.Env.WorkFlowsFolder)
	}
	fmt.Printf("[+] Health Check Workflows: %s\n", color.GreenString("✔"))

	var content [][]string
	for _, flow := range flows {
		parsedFlow, err := core.ParseFlow(flow)
		if err != nil {
			utils.ErrorF("Error parsing flow: %v", flow)
			continue
		}
		row := []string{
			parsedFlow.Name, parsedFlow.Desc,
		}
		content = append(content, row)
	}
	fmt.Printf("\nFound %v available workflows at: %s \n\n", color.HiGreenString("%v", len(content)), color.HiCyanString(options.Env.WorkFlowsFolder))

	table := tablewriter.NewWriter(os.Stdout)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Flow Name", "Description"})
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetColWidth(120)
	table.AppendBulk(content) // Add Bulk Data
	table.Render()

	h := "\nUsage:\n"
	h += color.HiCyanString("  osmedeus scan -f [flowName] -t [target] \n")
	fmt.Printf(h)
	return nil
}
