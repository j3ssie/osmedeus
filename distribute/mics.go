package distribute

import (
	"fmt"
	"path"
	"strings"

	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/cast"
)

// CommandBuilder build core command from API
func CommandBuilder(options libs.Options) string {
	binary := libs.BINARY

	if options.Debug {
		binary += " --debug"
	}

	if options.Update.EnableUpdate {
		binary += " --update"
	}

	if options.Threads > 0 {
		binary += " --threads-hold " + cast.ToString(options.Threads)
	}

	if options.ScanID != "" {
		binary += fmt.Sprintf(" --sid %v ", options.ScanID)
	}

	taskData := options.Cloud
	var command string
	var workspace, concurrency, timeout, params, workflow, plugin, extra string
	if options.Cloud.Extra != "" {
		extra = " " + options.Cloud.Extra + " "
	}

	if options.Cloud.Flow != "" {
		options.Cloud.Flow = utils.NormalizePath(options.Cloud.Flow)
		workflow = fmt.Sprintf(" -f '%v' ", options.Cloud.Flow)
	}

	if len(taskData.Params) > 0 {
		for _, param := range taskData.Params {
			params += fmt.Sprintf(" -p '%v'", param)
		}
	}

	if options.Cloud.Threads > 1 {
		concurrency = fmt.Sprintf(" -c %d ", options.Cloud.Threads)
	}

	// get workspace
	if taskData.Workspace == "" {
		taskData.Workspace = utils.CleanPath(taskData.Input)
	}
	if taskData.Workspace != "" {
		workspace = fmt.Sprintf(" -w '%v'", strings.TrimSpace(taskData.Workspace))
	}

	if taskData.Module != "" {
		taskData.Module = utils.NormalizePath(taskData.Module)
		plugin = fmt.Sprintf(" -m '%v' ", taskData.Module)
	}

	// override everything
	if taskData.RawCommand != "" {
		return taskData.RawCommand
	}

	// use target as a file
	if options.Cloud.TargetAsFile && utils.FileExists(taskData.Input) {
		taskData.InputsFile = taskData.Input
	}

	// mean general scan
	if taskData.Module == "" {
		command = fmt.Sprintf("%v scan %v -t %v %v%v%v%v%v", binary, workflow, taskData.Input, concurrency, timeout, workspace, params, extra)
		if taskData.InputsFile != "" {
			command = fmt.Sprintf("%v scan %v -T %v %v%v%v%v%v", binary, workflow, taskData.InputsFile, concurrency, timeout, workspace, params, extra)
		}
		command = strings.TrimSpace(command)

		return command
	}

	command = fmt.Sprintf("%v scan %v -t %v %v%v%v%v%v", binary, plugin, taskData.Input, concurrency, timeout, workspace, params, extra)
	if taskData.InputsFile != "" {
		command = fmt.Sprintf("%v scan %v -t %v %v%v%v%v%v", binary, plugin, taskData.InputsFile, concurrency, timeout, workspace, params, extra)
	}
	command = strings.TrimSpace(command)

	return command
}

// PrepareTarget change the target file destination
func PrepareTarget(target string, options libs.Options) string {
	if options.Cloud.EnableChunk {
		return target
	}

	if !utils.FileExists(target) && !utils.FolderExists(target) {
		utils.DebugF("target is not a file: %s", target)
		return target
	}

	utils.MakeDir(options.Cloud.TempTarget)
	dest := path.Join(options.Cloud.TempTarget, path.Base(target))

	utils.Copy(target, dest)
	utils.InforF("Change target %s --> %s", target, dest)
	return dest
}
