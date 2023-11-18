package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/server"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func init() {
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start Web Server",
		Long:  core.Banner(),
		RunE:  runServer,
	}
	serverCmd.Flags().String("host", "0.0.0.0", "IP address to bind the server")
	serverCmd.Flags().String("port", "8000", "Port")
	serverCmd.Flags().IntVar(&options.Server.PollingTime, "poll-time", 60, "Polling time to check next task")
	serverCmd.Flags().BoolVar(&options.Server.DisableSSL, "disable-ssl", false, "Disable workspaces directory listing")
	serverCmd.Flags().BoolVar(&options.Server.DisableWorkspaceListing, "disable-listing", false, "Disable workspaces directtory listing")
	serverCmd.Flags().BoolVar(&options.Server.PreFork, "prefork", false, "Enable Prefork mode for the api server")
	serverCmd.Flags().BoolVarP(&options.Server.NoAuthen, "no-auth", "A", false, "Disable authentication for the api server")

	serverCmd.SetHelpFunc(ServerHelp)
	RootCmd.AddCommand(serverCmd)
	serverCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if options.FullHelp {
			cmd.Help()
			os.Exit(0)
		}
	}
}

func runServer(cmd *cobra.Command, _ []string) error {
	host, _ := cmd.Flags().GetString("host")
	port, _ := cmd.Flags().GetString("port")
	options.Server.Bind = fmt.Sprintf("%v:%v", host, port)
	utils.GoodF("Using the %v Engine %v by %v", cases.Title(language.Und, cases.NoLower).String(libs.BINARY), color.HiCyanString(libs.VERSION), color.HiMagentaString(libs.AUTHOR))
	server.StartServer(options)
	return nil
}
