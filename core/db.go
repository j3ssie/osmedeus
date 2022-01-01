package core

import (
    "github.com/j3ssie/osmedeus/database"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    jsoniter "github.com/json-iterator/go"
    "github.com/robertkrimen/otto"
    "github.com/spf13/cast"
    "os"
    "path"
    "strings"
)

func (r *Runner) LoadDBScripts() string {
    var output string

    r.VM.Set("TotalSubdomain", func(call otto.FunctionCall) otto.Value {
        length := utils.FileLength(call.Argument(0).String())
        r.TargetObj.TotalAssets = length
        return otto.Value{}
    })

    r.VM.Set("TotalDns", func(call otto.FunctionCall) otto.Value {
        length := utils.FileLength(call.Argument(0).String())
        r.TargetObj.TotalDns = length
        return otto.Value{}
    })

    r.VM.Set("TotalScreenShot", func(call otto.FunctionCall) otto.Value {
        length := utils.FileLength(call.Argument(0).String())
        r.TargetObj.TotalScreenShot = length
        return otto.Value{}
    })

    r.VM.Set("TotalTech", func(call otto.FunctionCall) otto.Value {
        length := utils.FileLength(call.Argument(0).String())
        r.TargetObj.TotalTech = length
        return otto.Value{}
    })

    r.VM.Set("TotalVulnerability", func(call otto.FunctionCall) otto.Value {
        data := utils.ReadingFileUnique(call.Argument(0).String())
        var length int
        for _, line := range data {
            if !strings.Contains(line, "-info") {
                length += 1
            }
        }
        r.TargetObj.TotalVulnerability = length
        return otto.Value{}
    })

    r.VM.Set("TotalArchive", func(call otto.FunctionCall) otto.Value {
        length := utils.FileLength(call.Argument(0).String())
        r.TargetObj.TotalArchive = length
        return otto.Value{}
    })

    r.VM.Set("TotalLink", func(call otto.FunctionCall) otto.Value {
        length := utils.FileLength(call.Argument(0).String())
        r.TargetObj.TotalLink = length
        return otto.Value{}
    })

    r.VM.Set("TotalDirb", func(call otto.FunctionCall) otto.Value {
        length := utils.FileLength(call.Argument(0).String())
        r.TargetObj.TotalDirb = length
        return otto.Value{}
    })

    // CreateReport('report', 'subdomain')
    r.VM.Set("CreateReport", func(call otto.FunctionCall) otto.Value {
        args := call.ArgumentList
        report := args[0].String()
        if utils.FileExists(report) {
            return otto.Value{}
        }
        moduleName := "inline"
        if len(args) > 1 {
            moduleName = args[1].String()
        }

        reportObj := database.Report{
            ReportName:  path.Base(report),
            Module:      moduleName,
            ReportPath:  report,
            ReportType:  "",
            TargetRefer: r.TargetObj.ID,
        }
        database.NewReport(&reportObj)
        return otto.Value{}
    })

    return output
}

func (r *Runner) DBNewTarget() {
    if r.Opt.NoDB {
        return
    }

    //// this is just sample org
    //org := database.Org{
    //	Name: "Sample Org",
    //	Desc: "Sample Desc",
    //	//Targets: nil,
    //}
    //database.NewOrg(&org)

    r.TargetObj = database.Target{
        InputName: r.Input,
        Workspace: r.Workspace,
        InputType: r.InputType,
        //OrgRefer:  org.ID,
    }
    database.DBUpdateTarget(&r.TargetObj)
}

func (r *Runner) DBNewScan() {
    if r.Opt.NoDB {
        return
    }

    r.ScanObj = database.Scan{
        TaskType: r.RoutineType,
        TaskName: path.Base(r.RoutineName),

        // this should be user id as uuid
        //UID: r.RunnerSource,

        TotalSteps: r.TotalSteps,
        InputName:  r.Input,
        InputType:  r.InputType,

        LogFile:    r.Opt.LogFile,
        Target:     r.TargetObj,
        ProcessID:  os.Getpid(),
        IsRunning:  true,
        IsDone:     false,
        IsPrepared: true,
    }

    if r.Opt.ScanID != "" {
        utils.InforF("Continue scanning on scan id: %v", r.Opt.ScanID)
        r.ScanObj.ID = cast.ToUint(r.Opt.ScanID)
    }

    if r.RunnerType == "cloud" {
        r.ScanObj.IsCloud = true
    }

    database.DBNewScan(&r.ScanObj)
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

    utils.DebugF("[DB] Done module %v with %v/%v steps ", r.CurrentModule, r.DoneStep, r.TotalSteps)
    database.DBUpdateScan(&r.ScanObj)

    r.ScanObj.Target = r.TargetObj
    runtimeData, err := jsoniter.MarshalToString(r.ScanObj)
    if err == nil {
        utils.WriteToFile(r.RuntimeFile, runtimeData)
    }
}

func (r *Runner) DBDoneScan() {
    r.ScanObj.CurrentModule = "done"
    r.ScanObj.RunningTime = r.RunningTime

    r.ScanObj.DoneStep = r.TotalSteps
    r.ScanObj.IsDone = true
    r.ScanObj.IsRunning = false

    utils.DebugF("[DB] Done the scan: %v -- %v", r.ScanObj.InputName, r.ScanObj.TaskName)
    database.DBUpdateScan(&r.ScanObj)
}

func (r *Runner) DBUpdateTarget() {
    database.DBUpdateTarget(&r.TargetObj)
}

func (r *Runner) DBNewReports(module libs.Module) {
    if r.Opt.NoDB {
        return
    }

    r.ScanObj.CurrentModule = r.CurrentModule
    r.ScanObj.RunningTime = r.RunningTime
    database.DBUpdateScan(&r.ScanObj)

    var reports []string
    reports = append(reports, module.Report.Final...)
    reports = append(reports, module.Report.Noti...)
    reports = append(reports, module.Report.Diff...)

    utils.DebugF("Updating reports")
    for _, report := range reports {
        reportObj := database.Report{
            ReportName:  path.Base(report),
            ModulePath:  module.ModulePath,
            Module:      module.Name,
            ReportPath:  report,
            ReportType:  "",
            TargetRefer: r.TargetObj.ID,
        }

        database.NewReport(&reportObj)
        r.TargetObj.Reports = append(r.TargetObj.Reports, reportObj)
    }
}
