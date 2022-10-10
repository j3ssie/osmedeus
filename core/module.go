package core

import (
	"context"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/execution"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/panjf2000/ants"
)

// RunModule run the module
func (r *Runner) RunModule(module libs.Module) {
	// get reports path
	module = ResolveReports(module, r.Params)

	// check if resume enable or not
	if (r.Opt.Resume || module.Resume) && !module.Forced {
		if CheckResume(module) {
			utils.BlockF(module.Name, "Resume detected")
			return
		}
	}

	r.CurrentModule = module.Name
	timeStart := time.Now()
	utils.BannerF("Module-Started", fmt.Sprintf("%v - %v", module.Name, module.Desc))

	// create report record first because I don't want to wait for them to show up in UI until the module done
	r.DBNewReports(module)

	// pre-run
	utils.InforF("Running prepare scripts for module %v", color.CyanString(module.Name))
	r.RunScripts(module.PreRun)

	// main part
	utils.BlockF(module.Name, "Start run main steps")
	err := r.RunSteps(module.Steps)
	if err != nil {
		utils.BadBlockF(module.Name, fmt.Sprintf("got exit call"))
	}

	// post-run
	utils.InforF("Running prepare scripts for module %v", color.CyanString(module.Name))
	r.RunScripts(module.PostRun)

	// print the reports file
	utils.PrintLine()
	printReports(module)

	// create report record first because we don't want to wait it show up in UI until the module done
	r.DBNewReports(module)

	// estimate time
	elapsedTime := time.Since(timeStart).Seconds()
	utils.BlockF("Module-Ended", fmt.Sprintf("Elapsed Time for the module %v in %v", color.HiCyanString(module.Name), color.HiMagentaString("%vs", elapsedTime)))
	r.RunningTime += cast.ToInt(elapsedTime)
	utils.PrintLine()
	r.DBUpdateScan()
	r.DBUpdateTarget()
}

// RunScripts run list of scripts
func (r *Runner) RunScripts(scripts []string) string {
	if r.Opt.Timeout != "" {
		timeout := utils.CalcTimeout(r.Opt.Timeout)
		utils.DebugF("Run scripts with %v seconds timeout", timeout)
		r.RunScriptsWithTimeOut(r.Opt.Timeout, scripts)
		return ""
	}

	for _, script := range scripts {
		outScript := r.RunScript(script)
		if strings.Contains(outScript, "exit") {
			return outScript
		}
	}
	return ""
}

// RunScriptsWithTimeOut run list of scripts with timeout
func (r *Runner) RunScriptsWithTimeOut(timeoutRaw string, scripts []string) string {
	timeout := utils.CalcTimeout(timeoutRaw)
	utils.DebugF("Run scripts with %v seconds timeout", timeout)

	c := context.Background()
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	c, cancel := context.WithDeadline(c, deadline)
	defer cancel()

	go func() {
		for _, script := range scripts {
			outScript := r.RunScript(script)
			if strings.Contains(outScript, "exit") {
				return
			}
		}
		cancel()
	}()

	select {
	case <-c.Done():
		utils.DebugF("Scripts done")
		return ""
	case <-time.After(time.Duration(timeout) * time.Second):
		utils.BadBlockF("timeout", fmt.Sprintf("Scripts got timeout after %v", color.HiMagentaString(timeoutRaw)))
	}
	return ""
}

// RunScript really run a script
func (r *Runner) RunScript(script string) string {
	return r.ExecScript(script)
}

// RunSteps run list of steps
func (r *Runner) RunSteps(steps []libs.Step) error {
	var stepOut string
	for _, step := range steps {
		r.DoneStep += 1

		if step.Timeout != "" {
			// timeout should be: 30, 30m, 1h
			timeout := utils.CalcTimeout(step.Timeout)
			if timeout != 0 {
				stepOut, _ = r.RunStepWithTimeout(timeout, step)
				if strings.Contains(stepOut, "exit") {
					return fmt.Errorf("got exit call")
				}
				continue
			}
		}

		stepOut, _ = r.RunStep(step)
		if strings.Contains(stepOut, "exit") {
			return fmt.Errorf("got exit call")
		}
	}
	return nil
}

// RunStepWithTimeout run step with timeout
func (r *Runner) RunStepWithTimeout(timeout int, step libs.Step) (out string, err error) {
	utils.DebugF("Run step with %v seconds timeout", timeout)
	prefix := fmt.Sprintf("timeout -k 1m %vs ", timeout)

	// prepare the os command with prefix timeout first
	var preFixCommands []string
	for _, command := range step.Commands {
		preFixCommand := command
		if !strings.Contains(command, "timeout") {
			preFixCommand = prefix + command
		}
		preFixCommands = append(preFixCommands, preFixCommand)
	}
	step.Commands = preFixCommands

	// override global timeout
	r.Opt.Timeout = step.Timeout
	return r.RunStep(step)
}

func (r *Runner) RunStep(step libs.Step) (string, error) {
	var output string
	if step.Label != "" {
		utils.BlockF("Start-Step", color.HiCyanString(step.Label))
	}

	// checking required file
	err := r.CheckRequired(step.Required)
	if err != nil {
		return output, fmt.Errorf("missing requirements")
	}

	// check conditions and run reverse step
	err = r.CheckCondition(step.Conditions)
	if err != nil {
		if len(step.RCommands) == 0 && len(step.RScripts) == 0 {
			return output, fmt.Errorf("conditions not met")
		}

		// run reverse commands
		utils.InforF("Condition false, run the reverse commands")
		if len(step.RCommands) > 0 {
			r.RunCommands(step.RCommands, step.Std)
		}
		// run reverse scripts
		if len(step.RScripts) > 0 {
			output = r.RunScripts(step.RScripts)
			if strings.Contains(output, "exit") {
				return output, nil
			}
		}
		return output, nil
	}

	// run the step in loop mode
	if step.Source != "" {
		return r.RunStepWithSource(step)
	}
	//

	if len(step.Commands) > 0 {
		r.RunCommands(step.Commands, step.Std)
	}
	if len(step.Scripts) > 0 {
		output = r.RunScripts(step.Scripts)
		if strings.Contains(output, "exit") {
			return output, nil
		}
	}

	// run ose here
	if len(step.Ose) > 0 {
		for _, ose := range step.Ose {
			r.RunOse(ose)
		}
	}

	// post scripts
	if len(step.PConditions) > 0 || len(step.PScripts) > 0 {
		err := r.CheckCondition(step.PConditions)
		if err == nil {
			if len(step.PScripts) > 0 {
				r.RunScripts(step.PScripts)
			}
		}
	}
	if step.Label != "" {
		utils.BlockF("Done-Step", color.HiCyanString(step.Label))
	}
	return output, nil

}

// RunStepWithSource really run a step
func (r *Runner) RunStepWithSource(step libs.Step) (out string, err error) {
	////// Start to run step but in loop mode
	utils.DebugF("Run step with Source: %v", step.Source)
	data := utils.ReadingLines(step.Source)
	if len(data) <= 0 {
		return out, fmt.Errorf("missing source")
	}
	if step.Threads != "" {
		step.Parallel = cast.ToInt(step.Threads)
	}
	if step.Parallel == 0 {
		step.Parallel = 1
	}

	// prepare the data first
	var newGeneratedSteps []libs.Step
	for index, line := range data {
		customParams := make(map[string]string)
		customParams["line"] = line
		customParams["line_id"] = fmt.Sprintf("%v-%v", path.Base(line), index)
		customParams["_id_"] = fmt.Sprintf("%v", index)
		customParams["_line_"] = execution.StripName(line)

		// make completely new Step
		localStep := libs.Step{}

		for _, cmd := range step.Commands {
			localStep.Commands = append(localStep.Commands, AltResolveVariable(cmd, customParams))
		}
		for _, cmd := range step.RCommands {
			localStep.RCommands = append(localStep.RCommands, AltResolveVariable(cmd, customParams))
		}

		if len(step.Ose) > 0 {
			for _, ose := range step.Ose {
				localStep.Ose = append(localStep.Ose, AltResolveVariable(ose, customParams))
			}
		}

		for _, script := range step.RScripts {
			localStep.RScripts = append(localStep.RScripts, AltResolveVariable(script, customParams))
		}

		for _, script := range step.Scripts {
			localStep.Scripts = append(localStep.Scripts, AltResolveVariable(script, customParams))
		}

		for _, script := range step.PConditions {
			localStep.PConditions = append(localStep.PConditions, AltResolveVariable(script, customParams))
		}
		for _, script := range step.PScripts {
			localStep.PScripts = append(localStep.PScripts, AltResolveVariable(script, customParams))
		}

		newGeneratedSteps = append(newGeneratedSteps, localStep)
	}

	// skip concurrency part
	if step.Parallel == 1 {
		for _, newGeneratedStep := range newGeneratedSteps {
			out, err = r.RunStep(newGeneratedStep)
			if err != nil {
				continue
			}
		}
		if step.Label != "" {
			utils.BlockF("Done-Step", color.HiCyanString(step.Label))
		}
	}

	/////////////
	// run multiple steps in concurrency mode

	utils.DebugF("Run step in Parallel: %v", step.Parallel)
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(step.Parallel, func(i interface{}) {
		r.startStepJob(i)
		wg.Done()
	}, ants.WithPreAlloc(true))
	defer p.Release()

	for _, newGeneratedStep := range newGeneratedSteps {
		wg.Add(1)
		err = p.Invoke(newGeneratedStep)
		if err != nil {
			utils.ErrorF("Error in parallel: %v", err)
		}
	}

	wg.Wait()
	if step.Label != "" {
		utils.BlockF("Done-Step", color.HiCyanString(step.Label))
	}
	return out, nil
}

func (r *Runner) startStepJob(j interface{}) {
	localStep := j.(libs.Step)

	err := r.CheckCondition(localStep.Conditions)

	if err != nil {
		// run reverse commands
		if len(localStep.RCommands) > 0 {
			r.RunCommands(localStep.RCommands, localStep.Std)
		}
		if len(localStep.RScripts) > 0 {
			r.RunScripts(localStep.RScripts)
		}
	} else {
		if len(localStep.Commands) > 0 {
			r.RunCommands(localStep.Commands, localStep.Std)
		}
	}

	if len(localStep.Ose) > 0 {
		for _, ose := range localStep.Ose {
			r.RunOse(ose)
		}
	}

	if len(localStep.Scripts) > 0 {
		r.RunScripts(localStep.Scripts)
	}

	// post scripts
	if len(localStep.PConditions) > 0 || len(localStep.PScripts) > 0 {
		err := r.CheckCondition(localStep.PConditions)
		if err == nil {
			if len(localStep.PScripts) > 0 {
				r.RunScripts(localStep.PScripts)
			}
		}
	}
}
