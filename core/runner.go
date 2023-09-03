package core

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/Shopify/yaml"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/database"
	"github.com/j3ssie/osmedeus/execution"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/panjf2000/ants"
	"github.com/robertkrimen/otto"
	"github.com/thoas/go-funk"
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

	DoneFile        string
	RuntimeFile     string
	WorkspaceFolder string

	RoutineModules []string
	Reports        []string
	Routines       []libs.Routine

	VM        *otto.Otto
	TargetObj database.Target
	ScanObj   database.Scan

	Target map[string]string
	// this is same as targets but won't change during the execution time
	Params map[string]string
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
	allFlows := ListAllFlowName(r.Opt)

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
			utils.WarnF("Please select one of these flow: %v", color.HiMagentaString(strings.Join(allFlows, ", ")))
		}
	}
}

func (r *Runner) PrepareModule() {
	allModules := ListModuleName(r.Opt)
	r.RoutineType = "module"

	var err error
	for _, rawModule := range r.Opt.Scan.Modules {
		var routine libs.Routine

		module := DirectSelectModule(r.Opt, rawModule)
		r.RoutinePath = rawModule
		r.Opt.Module, err = ParseModules(module)
		if err != nil || r.Opt.Module.Name == "" {
			utils.WarnF("Your module %v doesn't exist", color.HiRedString(r.Opt.Scan.Modules[0]))
			utils.WarnF("Please select one of these module: %v", color.HiMagentaString(strings.Join(allModules, ", ")))
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

// PrepareParams prepare global params
func (r *Runner) PrepareParams() {
	r.Params = r.Target

	// looking for more params from each module
	for _, routine := range r.Routines {
		for _, module := range routine.ParsedModules {
			// params from module file
			if len(module.Params) > 0 {
				for _, param := range module.Params {
					for k, v := range param {
						// skip params if override: false
						_, exist := r.Params[k]
						if r.ForceParams && exist {
							utils.DebugF("Skip override param: %v --> %v", k, v)
							continue
						}

						v = ResolveData(v, r.Params)
						if strings.HasPrefix(v, "~/") {
							v = utils.NormalizePath(v)
						}
						r.Params[k] = v
					}
				}
			}

			if len(r.Opt.Scan.ParamsFile) > 0 {
				var params map[string]string
				yamlFile, err := os.ReadFile(r.Opt.Scan.ParamsFile)
				if err != nil {
					utils.ErrorF("YAML parsing err: %v -- #%v ", r.Opt.Scan.ParamsFile, err)
					return
				}
				err = yaml.Unmarshal(yamlFile, &params)
				if err != nil {
					utils.ErrorF("Error unmarshal: %v -- %v", params, err)
					return
				}
				if len(params) > 0 {
					for k, v := range params {
						v = ResolveData(v, r.Params)
						r.Params[k] = v
					}
				}
			}
			// more params from -p flag which will override everything
			if len(r.Opt.Scan.Params) > 0 {
				params := ParseParams(r.Opt.Scan.Params)
				if len(params) > 0 {
					for k, v := range params {
						v = ResolveData(v, r.Params)
						r.Params[k] = v
					}
				}
			}
		}
	}

	r.ResolveRoutine()

}

// ResolveRoutine resolve the module name first
func (r *Runner) ResolveRoutine() {
	var routines []libs.Routine

	for _, rawRoutine := range r.Routines {

		var routine libs.Routine
		for _, module := range rawRoutine.ParsedModules {
			module = ResolveReports(module, r.Params)

			r.Reports = append(r.Reports, module.Report.Final...)
			module.PreRun = ResolveSlice(module.PreRun, r.Params)

			// steps
			for i, step := range module.Steps {
				module.Steps[i].Timeout = ResolveData(step.Timeout, r.Params)
				module.Steps[i].Threads = ResolveData(step.Threads, r.Params)
				module.Steps[i].Label = ResolveData(step.Label, r.Params)
				module.Steps[i].Std = ResolveData(step.Std, r.Params)
				module.Steps[i].Source = ResolveData(step.Source, r.Params)

				module.Steps[i].Conditions = ResolveSlice(step.Conditions, r.Params)
				module.Steps[i].Required = ResolveSlice(step.Required, r.Params)

				module.Steps[i].Commands = ResolveSlice(step.Commands, r.Params)
				module.Steps[i].Scripts = ResolveSlice(step.Scripts, r.Params)

				module.Steps[i].RCommands = ResolveSlice(step.RCommands, r.Params)
				module.Steps[i].RScripts = ResolveSlice(step.RScripts, r.Params)
				module.Steps[i].PConditions = ResolveSlice(step.PConditions, r.Params)
				module.Steps[i].PScripts = ResolveSlice(step.PScripts, r.Params)
				module.Steps[i].Ose = ResolveSlice(step.Ose, r.Params)
			}

			module.PostRun = ResolveSlice(module.PostRun, r.Params)
			routine.ParsedModules = append(routine.ParsedModules, module)
		}

		routines = append(routines, routine)
	}
	r.Routines = routines
}

func (r *Runner) Start() {
	err := r.Validator()
	if err != nil {
		utils.ErrorF("Input does not match the require type: %v -- %v", r.RequiredInput, r.Input)
		utils.InforF("Adding %v flag if you want to disable input validate", color.HiCyanString(`'--nv'`))
		return
	}
	utils.InforF("Running %s tactic with baseline threads hold as %s", color.YellowString(r.Opt.Tactics), color.HiMagentaString("%v", r.Opt.Threads))

	r.Opt.Scan.ROptions = r.Target
	// prepare some metadata files
	utils.MakeDir(r.Target["Output"])
	r.DoneFile = r.Target["Output"] + "/done"
	r.RuntimeFile = r.Target["Output"] + "/runtime"
	r.WorkspaceFolder = r.Target["Output"]
	os.Remove(r.DoneFile)

	utils.InforF("Running the routine %v on %v", color.HiYellowString(r.RoutineName), color.CyanString(r.Input))
	utils.InforF("Detailed runtime file can be found on %v", color.CyanString(r.RuntimeFile))
	execution.TeleSendMess(r.Opt, fmt.Sprintf("%s -- Start new scan: %s -- %s", r.Opt.Noti.ClientName, r.Opt.Scan.Flow, r.Target["Workspace"]), "#status", false)

	r.DBNewTarget()
	r.DBNewScan()
	r.LoadEngineScripts()

	r.PrepareParams()

	/////
	/* really start the scan here */
	r.StartRoutines()
	/////

	r.DBDoneScan()
	utils.BlockF("Finished", fmt.Sprintf("The scan for %v was completed within %v", color.HiCyanString(r.Input), color.HiMagentaString("%vs", r.RunningTime)))

	if r.Opt.EnableBackup {
		r.BackupWorkspace()
	}
}

// StartRoutines start the scan
func (r *Runner) StartRoutines() {
	for _, routine := range r.Routines {
		// start each section of modules
		r.RunRoutine(routine.ParsedModules)
	}
}

func (r *Runner) RunRoutine(modules []libs.Module) {
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(r.Opt.Concurrency*10, func(m interface{}) {
		module := m.(libs.Module)
		r.RunModule(module)
		wg.Done()
	}, ants.WithPreAlloc(true))
	defer p.Release()

	for _, module := range modules {
		if funk.ContainsString(r.Opt.Exclude, module.Name) {
			utils.BadBlockF("Module-Excluded", fmt.Sprintf("Module %v has been excluded", color.CyanString(module.Name)))
			continue
		}

		p.Invoke(module)
		wg.Add(1)
	}

	wg.Wait()
}
