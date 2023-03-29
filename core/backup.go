package core

import (
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/execution"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
)

// in workflow file
// Compress('{{Backup}}/{{Workspace}}.tar.gz', '{{Output}}')
// Decompress('{{Output}}', '{{Backup}}/{{Workspace}}.tar.gz')

func (r *Runner) BackupWorkspace() {
	utils.InforF("Backing up the workspace: %v", r.Target["Workspace"])
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

func CompressWorkspace(target string, opt libs.Options) {
	utils.InforF("Backing up the workspace: %v", color.HiCyanString(target))
	outputDir := path.Join(opt.Env.WorkspacesFolder, target)
	if utils.FolderLength(outputDir) == 0 {
		utils.ErrorF("Workspace is empty: %s", outputDir)
		return
	}

	dest := path.Join(opt.Env.BackupFolder, target) + ".tar.gz"
	if utils.FileExists(dest) {
		os.Remove(dest)
	}

	execution.Compress(dest, outputDir)
	if utils.FileExists(dest) {
		utils.InforF("The workspace has been backed up and saved in %s", color.HiMagentaString(dest))
	}
}

func ExtractBackup(src string, opt libs.Options) {
	if !utils.FileExists(src) {
		utils.ErrorF("Backup file not found: %s", src)
		return
	}

	target := strings.ReplaceAll(path.Base(src), ".tar.gz", "")
	dest := path.Join(opt.Report.ExtractFolder, target)
	if !strings.HasSuffix(dest, "/") {
		dest += "/"
	}

	if utils.FolderExists(dest) {
		utils.MakeDir(dest)
	}
	execution.Decompress(dest, src)
	utils.GoodF("Extracting the %v to %s", color.HiCyanString(target), color.HiMagentaString(dest))
}
