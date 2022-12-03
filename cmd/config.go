package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/database"
	"github.com/j3ssie/osmedeus/execution"
	"os"
	"sort"

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
	configCmd.Flags().String("user", "", "Username")
	configCmd.Flags().String("pass", "", "Password")
	configCmd.Flags().StringP("workspace", "w", "", "Name of workspace")
	configCmd.SetHelpFunc(ConfigHelp)
	RootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	sort.Strings(args)
	action, _ := cmd.Flags().GetString("action")
	workspace, _ := cmd.Flags().GetString("workspace")
	DBInit()

	// backward compatible
	if action == "" && len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "init":
		if utils.FolderExists(fmt.Sprintf("%vcore", options.Env.RootFolder)) {
			utils.GoodF("Look like you got properly setup.")
		}
		break
	case "cred":
		username, _ := cmd.Flags().GetString("user")
		password, _ := cmd.Flags().GetString("pass")
		//database.CreateUser(username, password)
		utils.GoodF("Create new credentials %v:%v \n", username, password)
		break

	case "reload":
		core.ReloadConfig(&options)
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
		core.SetTactic(&options)
		break
	case "update":
		core.Update(options)
		break

	case "clean", "cl", "c":
		database.ClearDB()
		break
	default:
		utils.ErrorF("Unknown action: %v", color.HiRedString(action))
		fmt.Println(ConfigUsage())
	}

	return nil
}
