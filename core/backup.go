package core

import (
    "fmt"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "os"
    "path"
)

func BackupWorkspace(options libs.Options) {
    if !options.EnableBackup {
        return
    }

    outputDir := options.Scan.ROptions["Output"]
    dest := path.Join(options.Env.BackupFolder, options.Scan.ROptions["Workspace"]) + ".zip"
    if utils.FileExists(dest) {
        os.Remove(dest)
    }

    zipCommand := fmt.Sprintf("zip -9 -q -r %s %s", dest, outputDir)
    utils.RunCmdWithOutput(zipCommand)
    if utils.FileExists(dest) {
        utils.GoodF("Backup workspace save at: %s", dest)
    }
}
