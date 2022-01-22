package cmd

//import (
//    "github.com/j3ssie/osmedeus/core"
//    "github.com/spf13/cobra"
//)
//
//func init() {
//    var updateCmd = &cobra.Command{
//        Use:   "update",
//        Short: "Check latest Update",
//        Long:  core.Banner(),
//        RunE:  runUpdate,
//    }
//    updateCmd.Flags().String("repo", "", "Update repository URL")
//    RootCmd.AddCommand(updateCmd)
//}
//
//func runUpdate(cmd *cobra.Command, _ []string) error {
//    repo, _ := cmd.Flags().GetString("repo")
//    if repo != "" {
//        options.Update.UpdateURL = repo
//    }
//    core.Update(options)
//    return nil
//}
