package core

import (
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/execution"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"os"
	"path"
	"strings"
)

func (r *Runner) BackupWorkspace() {
	outputDir := r.Target["Output"]
	dest := path.Join(r.Opt.Env.BackupFolder, r.Target["Workspace"]) + ".tar.gz"
	if utils.FileExists(dest) {
		os.Remove(dest)
	}

	execution.Compress(dest, outputDir)
	if utils.FileExists(dest) {
		utils.GoodF("Backup workspace save at %s", color.HiMagentaString(dest))
	}
}

func ExtractBackup(src string, opt libs.Options) {
	if !utils.FileExists(src) {
		utils.ErrorF("Backup file not found: %s", src)
		return
	}

	target := strings.ReplaceAll(path.Base(src), ".tar.gz", "")
	dest := opt.Report.ExtractFolder
	if !strings.HasSuffix(dest, "/") {
		dest += "/"
	}

	if utils.FolderExists(dest) {
		utils.MakeDir(dest)
	}
	execution.Decompress(dest, src)
	utils.GoodF("Extracting the %v to %s", color.HiCyanString(target), color.HiMagentaString(dest))
}
