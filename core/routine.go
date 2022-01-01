package core

import (
    "context"
    "errors"
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
    "github.com/jinzhu/copier"
    "github.com/panjf2000/ants"
)

// RunModule run the module
func (r *Runner) RunModule(module libs.Module, options libs.Options) {
    // get more params
    MoreParams(module, &options)
    // get reports path
    options.Module = ResolveReports(module, options)
    r.LoadEngineScripts()

    // check if resume enable or not
    if (options.Resume || module.Resume) && !module.Forced {
        if CheckResume(options) {
            utils.BlockF(module.Name, "Resume detected")
            return
        }
    }

    r.CurrentModule = module.Name
    timeStart := time.Now()
    utils.BannerF("MODULES", fmt.Sprintf("%v - %v", module.Name, options.Module.Desc))

    // create report record first because I don't want to wait for them to show up in UI until the module done
    r.DBNewReports(module)

    // pre-run
    utils.BlockF(module.Name, "Running prepare scripts")
    r.RunScripts(module.PreRun, options)

    // main part
    utils.BlockF(module.Name, "Start run main steps")
    err := r.RunSteps(module.Steps, options)
    if err != nil {
        utils.BadBlockF(module.Name, fmt.Sprintf("got exit call"))
    }

    // post-run
    utils.BlockF(module.Name, "Running conclusion scripts")
    r.RunScripts(module.PostRun, options)

    // print the reports file
    utils.PrintLine()
    printReports(options)

    // create report record first because we don't want to wait it show up in UI until the module done
    r.DBNewReports(module)

    // estimate time
    elapsedTime := time.Since(timeStart).Seconds()
    utils.BlockF("Elapsed Time", fmt.Sprintf("Done module %v in %v", color.HiCyanString(module.Name), color.HiMagentaString("%vs", elapsedTime)))
    r.RunningTime += cast.ToInt(elapsedTime)
    utils.PrintLine()
    r.DBUpdateScan()
    r.DBUpdateTarget()
}

// RunScripts run list of scripts
func (r *Runner) RunScripts(scripts []string, options libs.Options) string {
    if options.Timeout != "" {
        timeout := utils.CalcTimeout(options.Timeout)
        utils.DebugF("Run scripts with %v seconds timeout", timeout)
        r.RunScriptsWithTimeOut(options.Timeout, scripts, options)
        return ""
    }

    for _, script := range scripts {
        outScript := r.RunScript(script, options)
        if strings.Contains(outScript, "exit") {
            return outScript
        }
    }
    return ""
}

// RunScriptsWithTimeOut run list of scripts with timeout
func (r *Runner) RunScriptsWithTimeOut(timeoutRaw string, scripts []string, options libs.Options) string {
    timeout := utils.CalcTimeout(timeoutRaw)
    utils.DebugF("Run scripts with %v seconds timeout", timeout)

    c := context.Background()
    deadline := time.Now().Add(time.Duration(timeout) * time.Second)
    c, cancel := context.WithDeadline(c, deadline)
    defer cancel()

    go func() {
        for _, script := range scripts {
            outScript := r.RunScript(script, options)
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
func (r *Runner) RunScript(script string, options libs.Options) string {
    execScript := ResolveData(script, options.Scan.ROptions)
    return r.ExecScript(execScript)
}

// RunSteps run list of steps
func (r *Runner) RunSteps(steps []libs.Step, options libs.Options) error {
    var stepOut string
    for _, step := range steps {
        r.DoneStep += 1
        if step.Timeout != "" {
            step.Timeout = ResolveData(step.Timeout, options.Scan.ROptions)
            // timeout should be: 30, 30m, 1h
            timeout := utils.CalcTimeout(step.Timeout)
            if timeout != 0 {
                stepOut, _ = r.RunStepWithTimeout(timeout, step, options)
                if strings.Contains(stepOut, "exit") {
                    return errors.New("got exit call")
                }
                continue
            }
        }

        stepOut, _ = r.RunStep(step, options)
        if strings.Contains(stepOut, "exit") {
            return errors.New("got exit call")
        }
    }
    return nil
}

// RunStepWithTimeout run step with timeout
func (r *Runner) RunStepWithTimeout(timeout int, step libs.Step, options libs.Options) (string, error) {
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
    options.Timeout = step.Timeout
    output, _ := r.RunStep(step, options)
    return output, nil
}

// RunStep really run a step
func (r *Runner) RunStep(step libs.Step, options libs.Options) (string, error) {
    var output string
    if step.Label != "" {
        utils.BlockF("Start-Step", color.HiCyanString(step.Label))
    }

    // checking required file
    err := r.CheckRequired(step.Required, r.Opt)
    if err != nil {
        return output, errors.New("missing requirements")
    }

    // check conditions and run reverse step
    err = r.CheckCondition(step.Conditions, r.Opt)
    if err != nil {
        if len(step.RCommands) == 0 && len(step.RScripts) == 0 {
            return output, errors.New("conditions not met")
        }

        // run reverse commands
        utils.InforF("Condition false, run the reverse commands")
        if len(step.RCommands) > 0 {
            RunCommands(step.RCommands, step.Std, r.Opt)
        }
        // run reverse scripts
        if len(step.RScripts) > 0 {
            output = r.RunScripts(step.RScripts, r.Opt)
            if strings.Contains(output, "exit") {
                return output, nil
            }
        }
        return output, nil
    }

    if step.Source == "" {
        if len(step.Commands) > 0 {
            RunCommands(step.Commands, step.Std, r.Opt)
        }
        if len(step.Scripts) > 0 {
            output = r.RunScripts(step.Scripts, r.Opt)
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
            err := r.CheckCondition(step.PConditions, r.Opt)
            if err == nil {
                if len(step.PScripts) > 0 {
                    r.RunScripts(step.PScripts, r.Opt)
                }
            }
        }
        if step.Label != "" {
            utils.BlockF("Done-Step", color.HiCyanString(step.Label))
        }
        return output, nil
    }

    ////// Start to run step but in loop mode

    source := ResolveData(step.Source, r.Target)
    utils.DebugF("Run step with Source: %v", source)
    data := utils.ReadingLines(source)
    if len(data) <= 0 {
        return output, errors.New("missing source")
    }
    if step.Threads != "" {
        step.Threads = ResolveData(step.Threads, r.Target)
        step.Parallel = cast.ToInt(step.Threads)
    }
    if step.Parallel == 0 {
        step.Parallel = 1
    }

    // skip concurrency part
    if step.Parallel == 1 {
        for index, line := range data {
            r.Target["line"] = line
            r.Target["line_id"] = fmt.Sprintf("%v-%v", path.Base(line), index)
            r.Target["_id_"] = fmt.Sprintf("%v", index)
            r.Target["_line_"] = execution.StripName(line)
            if len(step.Commands) > 0 {
                RunCommands(step.Commands, step.Std, options)
            }

            if len(step.Ose) > 0 {
                for _, ose := range step.Ose {
                    r.RunOse(ose)
                }
            }

            if len(step.Scripts) > 0 {
                r.RunScripts(step.Scripts, options)
            }

            // post scripts
            if len(step.PConditions) > 0 || len(step.PScripts) > 0 {
                err := r.CheckCondition(step.PConditions, options)
                if err == nil {
                    if len(step.PScripts) > 0 {
                        r.RunScripts(step.PScripts, options)
                    }
                }
            }

        }

        if step.Label != "" {
            utils.BlockF("Done-Step", color.HiCyanString(step.Label))
        }
        return output, nil
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

    var mu sync.Mutex
    for index, line := range data {
        mu.Lock()
        localOptions := libs.Options{}
        copier.Copy(&localOptions, &options)
        localOptions.Scan.ROptions["line"] = line
        localOptions.Scan.ROptions["line_id"] = fmt.Sprintf("%v-%v", path.Base(line), index)
        localOptions.Scan.ROptions["_id_"] = fmt.Sprintf("%v", index)
        localOptions.Scan.ROptions["_line_"] = execution.StripName(line)

        // make completely new Step
        localStep := libs.Step{}

        for _, cmd := range step.Commands {
            localStep.Commands = append(localStep.Commands, ResolveData(cmd, localOptions.Scan.ROptions))
        }
        for _, cmd := range step.RCommands {
            localStep.RCommands = append(localStep.RCommands, ResolveData(cmd, localOptions.Scan.ROptions))
        }

        if len(step.Ose) > 0 {
            for _, ose := range step.Ose {
                localStep.Ose = append(localStep.Ose, ResolveData(ose, localOptions.Scan.ROptions))
            }
        }

        for _, script := range step.RScripts {
            localStep.RScripts = append(localStep.RScripts, ResolveData(script, localOptions.Scan.ROptions))
        }

        for _, script := range step.Scripts {
            localStep.Scripts = append(localStep.Scripts, ResolveData(script, localOptions.Scan.ROptions))
        }

        for _, script := range step.PConditions {
            localStep.PConditions = append(localStep.PConditions, ResolveData(script, localOptions.Scan.ROptions))
        }
        for _, script := range step.PScripts {
            localStep.PScripts = append(localStep.PScripts, ResolveData(script, localOptions.Scan.ROptions))
        }

        job := stepJob{
            options: localOptions,
            step:    localStep,
        }
        wg.Add(1)
        _ = p.Invoke(job)
        mu.Unlock()
    }

    wg.Wait()
    if step.Label != "" {
        utils.BlockF("Done-Step", color.HiCyanString(step.Label))
    }
    return output, nil
}

type stepJob struct {
    options libs.Options
    step    libs.Step
}

func (r *Runner) startStepJob(j interface{}) {
    job := j.(stepJob)
    localOptions := job.options
    localStep := job.step

    err := r.CheckCondition(localStep.Conditions, localOptions)

    if err != nil {
        // run reverse commands
        if len(localStep.RCommands) > 0 {
            RunCommands(localStep.RCommands, localStep.Std, localOptions)
        }
        if len(localStep.RScripts) > 0 {
            r.RunScripts(localStep.RScripts, localOptions)
        }
    } else {
        if len(localStep.Commands) > 0 {
            RunCommands(localStep.Commands, localStep.Std, localOptions)
        }
    }

    if len(localStep.Ose) > 0 {
        for _, ose := range localStep.Ose {
            r.RunOse(ose)
        }
    }

    if len(localStep.Scripts) > 0 {
        r.RunScripts(localStep.Scripts, localOptions)
    }

    // post scripts
    if len(localStep.PConditions) > 0 || len(localStep.PScripts) > 0 {
        err := r.CheckCondition(localStep.PConditions, localOptions)
        if err == nil {
            if len(localStep.PScripts) > 0 {
                r.RunScripts(localStep.PScripts, localOptions)
            }
        }
    }
}
