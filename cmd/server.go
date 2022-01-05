package cmd

import (
    "fmt"

    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/server"
    "github.com/spf13/cobra"
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
    serverCmd.Flags().BoolVar(&options.Server.PreFork, "prefork", false, "Enable Prefork mode for api server")
    RootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, _ []string) error {
    host, _ := cmd.Flags().GetString("host")
    port, _ := cmd.Flags().GetString("port")
    options.Server.Bind = fmt.Sprintf("%v:%v", host, port)
    DBInit()

    server.StartServer(options)
    return nil
}
