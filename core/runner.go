package core

import (
    "fmt"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/database"
    "github.com/j3ssie/osmedeus/execution"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/robertkrimen/otto"
    "github.com/thoas/go-funk"
    "os"
    "strings"
)

// Runner runner struct to start a job
type Runner struct {
    Input     string
    Workspace string

    InputType     string // domain, url, ip, cidr or domain-file, url-file, ip-file, cidr-file
    RequiredInput string // this should match with InputType
    IsInvalid     bool
    ForceParams   bool

    RoutineType  string // module or flow
    RoutineName  string // general
    RoutinePath  string
    RunnerSource string // cli or  api
    RunnerType   string // local or cloud

    Opt libs.Options

    // use for analytics
    DoneStep      int
    TotalSteps    int
    RunningTime   int
    CurrentModule string

    DoneFile    string
    RuntimeFile string

    RoutineModules []string
    Routines       []libs.Routine

    // RunTime *otto.Otto
    VM *otto.Otto

    TargetObj database.Target
    ScanObj   database.Scan

    Target map[string]string
}

// InitRunner init runner
func InitRunner(input string, opt libs.Options) (Runner, error) {
    var runner Runner
    runner.Input = input
    runner.Opt = opt
    runner.PrepareRoutine()
    runner.InitVM()

    runner.RunnerSource = "cli"
    // @TODO check if running in cloud
    runner.RunnerType = "local"
    return runner, nil
}

// PrepareWorkflow prepare workflow file
func (r *Runner) PrepareWorkflow() {
    r.RoutineType = "flow"
    var err error
    flows := SelectFlow(r.Opt.Scan.Flow, r.Opt)
    for _, flow := range flows {
        if strings.TrimSpace(flow) == "" {
            continue
        }
        r.Opt.Flow, err = ParseFlow(flow)
        r.RoutinePath = flow
        if err != nil {
            continue
        }

        if r.Opt.Flow.NoDB {
            r.Opt.NoDB = true
        }

        if r.Opt.Flow.Type == "general" {
            r.Opt.Flow.Input = r.Opt.Scan.Input
        }
        // default folder to look for module file if not specify
        r.Opt.Flow.DefaultType = r.Opt.Flow.Type

        r.Target["FlowPath"] = flow
        r.RequiredInput = r.Opt.Flow.Validator
        r.ForceParams = r.Opt.Flow.ForceParams

        // get more params from flow
        if len(r.Opt.Flow.Params) > 0 {
            for _, params := range r.Opt.Flow.Params {
                for k, v := range params {
                    r.Target[k] = v
                }
            }
        }

    }

    // generate routines
    for _, routine := range r.Opt.Flow.Routines {
        // select module depend on the flow type
        if routine.FlowFolder != "" {
            r.Opt.Flow.Type = routine.FlowFolder
        } else {
            r.Opt.Flow.Type = r.Opt.Flow.DefaultType
        }

        modules := SelectModules(routine.Modules, r.Opt)
        routine.RoutineName = fmt.Sprintf("flow-%s", r.Opt.Flow.Name)

        for _, module := range modules {
            parsedModule, err := ParseModules(module)
            if err != nil || parsedModule.Name == "" {
                continue
            }
            r.TotalSteps += len(parsedModule.Steps)
            routine.ParsedModules = append(routine.ParsedModules, parsedModule)
        }
        r.Routines = append(r.Routines, routine)
    }

    if len(r.Routines) == 0 {
        if r.Opt.Scan.Flow != "cloud-distributed" {
            utils.WarnF("Your workflow %v doesn't exist", color.HiRedString(r.Opt.Scan.Flow))
        }
    }
}

func (r *Runner) PrepareModule() {
    r.RoutineType = "module"
    var err error
    for _, rawModule := range r.Opt.Scan.Modules {
        var routine libs.Routine

        module := DirectSelectModule(r.Opt, rawModule)
        r.RoutinePath = rawModule
        r.Opt.Module, err = ParseModules(module)
        if err != nil || r.Opt.Module.Name == "" {
            continue
        }
        if r.Opt.Module.NoDB {
            r.Opt.NoDB = true
        }

        r.Target["FlowPath"] = "direct-module"
        routine.ParsedModules = append(routine.ParsedModules, r.Opt.Module)
        routine.RoutineName = fmt.Sprintf("module-%s", r.Opt.Flow.Name)
        r.Target["Module"] = module

        r.Routines = append(r.Routines, routine)

        r.TotalSteps += len(r.Opt.Module.Steps)
        r.RequiredInput = r.Opt.Module.Validator
    }
}

func (r *Runner) PrepareRoutine() {
    // prepare targets
    r.Target = ParseInput(r.Input, r.Opt)
    r.Workspace = r.Target["Workspace"]

    // take from -m flag
    if len(r.Opt.Scan.Modules) > 0 {
        r.RoutineName = strings.Join(r.Opt.Scan.Modules, "-")
        r.PrepareModule()
        return
    }

    // take from -f flag
    if r.Opt.Scan.Flow != "" {
        r.RoutineName = r.Opt.Scan.Flow
        r.PrepareWorkflow()
    }
}

func (r *Runner) Start() {
    err := r.Validator()
    if err != nil {
        utils.ErrorF("Input does not match the require type: %v -- %v", r.RequiredInput, r.Input)
        utils.InforF("Use '--nv' if you want to disable input validate")
        return
    }

    r.Opt.Scan.ROptions = r.Target
    utils.BlockF("Target", r.Input)
    utils.BlockF("Routine", r.RoutineName)

    // prepare some metadata files
    utils.MakeDir(r.Target["Output"])
    r.DoneFile = r.Target["Output"] + "/done"
    r.RuntimeFile = r.Target["Output"] + "/runtime"
    os.Remove(r.DoneFile)

    execution.TeleSendMess(r.Opt, fmt.Sprintf("**%s** -- Start new scan: **%s** -- **%s**", r.Opt.Noti.ClientName, r.Opt.Scan.Flow, r.Target["Workspace"]), "#status", false)

    r.DBNewTarget()
    r.DBNewScan()
    r.StartScanNoti()

    /////
    /* really start the scan here */
    r.StartRoutine()
    /////

    BackupWorkspace(r.Opt)
    r.DBDoneScan()
    r.ScanDoneNoti()

    utils.BlockF("Done", fmt.Sprintf("scan for %v -- %v", r.Input, color.HiMagentaString("%vs", r.RunningTime)))
    utils.WriteToFile(r.DoneFile, "done")
}

// StartRoutine start the scan
func (r *Runner) StartRoutine() {
    for _, routine := range r.Routines {
        for _, module := range routine.ParsedModules {
            module = ResolveReports(module, r.Opt)
            module.ForceParams = r.ForceParams
            r.Opt.Module = module

            // check exclude options
            if funk.ContainsString(r.Opt.Exclude, module.Name) {
                utils.BadBlockF(module.Name, fmt.Sprintf("Module Got Excluded"))
                continue
            }

            rTimeout := r.Opt.Timeout
            if module.MTimeout != "" {
                rTimeout = ResolveData(module.MTimeout, r.Target)
            }

            if rTimeout != "" {
                timeout := utils.CalcTimeout(rTimeout)
                if timeout < 43200 {
                    // run the module but with timeout
                    r.RunModulesWithTimeout(rTimeout, module, r.Opt)
                    continue
                }
            }

            r.RunModule(module, r.Opt)
        }
    }
}
