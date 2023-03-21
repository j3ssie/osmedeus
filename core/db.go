package core

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/database"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/robertkrimen/otto"
)

func (r *Runner) LoadDBScripts() string {
	var output string

	r.VM.Set(TotalSubdomain, func(call otto.FunctionCall) otto.Value {
		length := utils.FileLength(call.Argument(0).String())
		r.TargetObj.TotalAssets = length
		utils.InforF("Total subdomain found: %v", color.HiMagentaString("%v", length))
		return otto.Value{}
	})

	r.VM.Set(TotalDns, func(call otto.FunctionCall) otto.Value {
		length := utils.FileLength(call.Argument(0).String())
		r.TargetObj.TotalDns = length
		utils.InforF("Total Dns: %v", color.HiMagentaString("%v", length))
		return otto.Value{}
	})

	r.VM.Set(TotalScreenShot, func(call otto.FunctionCall) otto.Value {
		length := utils.FileLength(call.Argument(0).String())
		r.TargetObj.TotalScreenShot = length
		utils.InforF("Total ScreenShot: %v", color.HiMagentaString("%v", length))
		return otto.Value{}
	})

	r.VM.Set(TotalTech, func(call otto.FunctionCall) otto.Value {
		length := utils.FileLength(call.Argument(0).String())
		r.TargetObj.TotalTech = length
		utils.InforF("Total Tech: %v", color.HiMagentaString("%v", length))
		return otto.Value{}
	})

	r.VM.Set(TotalVulnerability, func(call otto.FunctionCall) otto.Value {
		data := utils.ReadingFileUnique(call.Argument(0).String())
		var length int
		for _, line := range data {
			if !strings.Contains(line, "-info") {
				length += 1
			}
		}
		r.TargetObj.TotalVulnerability = length
		utils.InforF("Total Vulnerability: %v", color.HiMagentaString("%v", length))
		return otto.Value{}
	})

	r.VM.Set(TotalArchive, func(call otto.FunctionCall) otto.Value {
		length := utils.FileLength(call.Argument(0).String())
		r.TargetObj.TotalArchive = length
		utils.InforF("Total Archive: %v", color.HiMagentaString("%v", length))
		return otto.Value{}
	})

	r.VM.Set(TotalLink, func(call otto.FunctionCall) otto.Value {
		length := utils.FileLength(call.Argument(0).String())
		r.TargetObj.TotalLink = length
		utils.InforF("Total Link: %v", color.HiMagentaString("%v", length))
		return otto.Value{}
	})

	r.VM.Set(TotalDirb, func(call otto.FunctionCall) otto.Value {
		length := utils.FileLength(call.Argument(0).String())
		r.TargetObj.TotalDirb = length
		utils.InforF("Total Dirb: %v", color.HiMagentaString("%v", length))
		return otto.Value{}
	})

	r.VM.Set(CreateReport, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		reportPath := args[0].String()
		if utils.FileExists(reportPath) {
			return otto.Value{}
		}
		moduleName := "inline"
		if len(args) > 1 {
			moduleName = args[1].String()
		}
		reportItem := database.Report{
			ReportPath: reportPath,
			Module:     moduleName,
		}
		r.TargetObj.Reports = append(r.TargetObj.Reports, reportItem)
		return otto.Value{}
	})

	return output
}

func (r *Runner) DBNewTarget() {
	r.TargetObj = database.Target{
		InputName: r.Input,
		Workspace: r.Workspace,
		InputType: r.InputType,
	}

	r.DBRuntimeUpdate()
}

func (r *Runner) DBNewScan() {
	r.ScanObj = database.Scan{
		TaskType: r.RoutineType,
		TaskName: path.Base(r.RoutineName),

		TotalSteps: r.TotalSteps,
		InputName:  r.Input,
		InputType:  r.InputType,

		LogFile:    r.Opt.LogFile,
		Target:     r.TargetObj,
		ProcessID:  os.Getpid(),
		IsRunning:  true,
		IsDone:     false,
		IsPrepared: true,
		IsStarted:  true,
	}

	if r.RunnerType == "cloud" {
		r.ScanObj.IsCloud = true
	}

	if r.Opt.Scan.RemoteCall {
		r.ScanObj.IsCloud = true
	}

	r.ScanObj.CreatedAt = time.Now()

	if runtimeData, err := jsoniter.MarshalToString(r.ScanObj); err == nil {
		utils.WriteToFile(r.RuntimeFile, runtimeData)
	}

}

func (r *Runner) DBUpdateScan() {
	r.ScanObj.DoneStep = r.DoneStep
	r.ScanObj.CurrentModule = r.CurrentModule
	r.ScanObj.RunningTime = r.RunningTime
	r.ScanObj.ProcessID = os.Getpid()

	if r.ScanObj.DoneStep == r.ScanObj.TotalSteps {
		r.ScanObj.IsDone = true
		r.ScanObj.IsRunning = false
	} else {
		r.ScanObj.IsRunning = true
		r.ScanObj.IsDone = false

	}

	utils.DebugF("[DB] Finished %v steps in the %v module", color.HiCyanString("%v/%v", r.DoneStep, r.TotalSteps), r.CurrentModule)
	r.DBRuntimeUpdate()
}

func (r *Runner) DBDoneScan() {
	r.ScanObj.CurrentModule = "done"
	r.ScanObj.RunningTime = r.RunningTime

	r.ScanObj.DoneStep = r.TotalSteps
	r.ScanObj.IsDone = true
	r.ScanObj.IsRunning = false
	r.ScanObj.IsStarted = false
	r.ScanObj.UpdatedAt = time.Now()

	utils.DebugF("[DB] The scan has been completed: %v -- %v", color.HiCyanString(r.ScanObj.InputName), color.HiCyanString(r.ScanObj.TaskName))
	if runtimeData, err := jsoniter.MarshalToString(r.ScanObj); err == nil {
		utils.WriteToFile(r.DoneFile, runtimeData)
	}
}

func (r *Runner) DBRuntimeUpdate() {
	r.ScanObj.UpdatedAt = time.Now()
	r.ScanObj.Target = r.TargetObj
	if runtimeData, err := jsoniter.MarshalToString(r.ScanObj); err == nil {
		utils.WriteToFile(r.RuntimeFile, runtimeData)
	}
}

func (r *Runner) DBNewReports(module libs.Module) {
	r.ScanObj.CurrentModule = r.CurrentModule
	r.ScanObj.RunningTime = r.RunningTime

	var reports []string
	reports = append(reports, module.Report.Final...)
	reports = append(reports, module.Report.Noti...)
	reports = append(reports, module.Report.Diff...)

	utils.DebugF("Updating %v report records", len(reports))
	for _, report := range reports {
		reportType := "text"
		if strings.HasSuffix(report, ".html") {
			reportType = "html"
		}

		reportObj := database.Report{
			ReportName: path.Base(report),
			ModulePath: module.ModulePath,
			Module:     module.Name,
			ReportPath: report,
			ReportType: reportType,
		}

		r.TargetObj.Reports = append(r.TargetObj.Reports, reportObj)

	}
	r.ScanObj.Target = r.TargetObj
}
