package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/execution"

	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/cobra"
)

func init() {
	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Do some configuration from CLI",
		Long:  core.Banner(),
		RunE:  runConfig,
	}

	configCmd.Flags().StringP("action", "a", "", "Action")
	configCmd.Flags().String("pluginsRepo", "git@gitlab.com:j3ssie/osmedeus-plugins.git", "Osmedeus Plugins repository")

	// for cred action
	configCmd.Flags().String("client-name", "", "Client name for notification")
	configCmd.Flags().String("user", "", "Username")
	configCmd.Flags().String("pass", "", "Password")
	configCmd.Flags().StringP("workspace", "w", "", "Name of workspace")
	configCmd.Flags().String("master-pass", "", "Set Master password for the API Authentication")

	configCmd.SetHelpFunc(ConfigHelp)
	RootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	sort.Strings(args)
	action, _ := cmd.Flags().GetString("action")
	masterPass, _ := cmd.Flags().GetString("master-pass")
	workspace, _ := cmd.Flags().GetString("workspace")
	clientName, _ := cmd.Flags().GetString("client-name")

	// backward compatible
	if action == "" && len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "check":
		err := generalCheck()
		if err != nil {
			fmt.Printf("‼️ There is might be something wrong with your setup: %v\n", color.HiRedString("%v", err))
			return nil
		}
		break
	case "init":
		if utils.FolderExists(fmt.Sprintf("%vcore", options.Env.RootFolder)) {
			utils.TSPrintF("Look like you got properly setup.")
		}
		break
	case "cred":
		username, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("pass")
		utils.TSPrintF("Create new credentials %v:%v \n", username, password)
		break

	case "reload":
		fmt.Println("💬 Reload the configuration will replace current settings with new ones based on the current environment")
		var input string
		fmt.Printf(color.HiRedString("🌀 Do you want to proceed? (y/N): "))
		fmt.Scan(&input)
		input = strings.ToLower(input)
		if input == "yes" || input == "y" {
			utils.InforF("Delete current config and generate a new one")
			os.Remove(options.ConfigFile)
			os.Remove(options.TokenConfigFile)
			core.InitConfig(&options)
			core.ParsingConfig(&options)
		}
		break

	case "delete", "del":
		options.Scan.Input = workspace
		options.Scan.ROptions = core.ParseInput(options.Scan.Input, options)
		utils.InforF("Delete Workspace: %v", options.Scan.ROptions["Workspace"])
		os.RemoveAll(options.Scan.ROptions["Output"])
		break

	case "pull":
		for repo := range options.Storages {
			execution.PullResult(repo, options)
		}
		break

	case "set":
		if clientName != "" {
			core.SetClientName(&options, clientName)
			break
		}

		if masterPass != "" {
			core.SetMasterPassword(&options, masterPass)
			break
		}
		core.SetTactic(&options)

		break
	case "update":
		core.Update(options)
		break

	case "clean", "cl", "c":
		break
	default:
		utils.ErrorF("Unknown action: %v", color.HiRedString(action))
		if options.FullHelp {
			fmt.Println(cmd.UsageString())
		}
		fmt.Println(ConfigUsage())
	}

	return nil
}
