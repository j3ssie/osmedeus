package core

import (
    "context"
    "fmt"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/thoas/go-funk"
    "strings"
    "sync"
    "time"
)

func (r *Runner) RunModulesWithTimeout(timeoutRaw string, module libs.Module, options libs.Options) {
    timeout := utils.CalcTimeout(timeoutRaw)
    utils.InforF("Run module %v with %v seconds timeout", color.HiCyanString(module.Name), timeout)

    c := context.Background()
    deadline := time.Now().Add(time.Duration(timeout) * time.Second)
    c, cancel := context.WithDeadline(c, deadline)
    defer cancel()

    go func() {
        r.RunModule(module, options)
        cancel()
    }()

    select {
    case <-c.Done():
        utils.DebugF("Module done")
        return
    case <-time.After(time.Duration(timeout) * time.Second):
        utils.BadBlockF("timeout", fmt.Sprintf("Module got timeout after %v", color.HiMagentaString(timeoutRaw)))
    }
    return
}

// CheckResume check resume report
func CheckResume(options libs.Options) bool {
    for _, report := range options.Module.Report.Final {
        if !strings.Contains(report, ".osmedeus/storages") && !utils.FileExists(report) {
            return false
        }
    }
    return true
}

// ResolveReports resolve real path of reports
func ResolveReports(module libs.Module, options libs.Options) libs.Module {
    var final []string
    var noti []string
    var diff []string
    for _, report := range module.Report.Final {
        final = append(final, ResolveData(report, options.Scan.ROptions))
    }

    for _, report := range module.Report.Noti {
        noti = append(noti, ResolveData(report, options.Scan.ROptions))
    }

    for _, report := range module.Report.Diff {
        diff = append(diff, ResolveData(report, options.Scan.ROptions))
    }

    module.Report.Final = final
    module.Report.Noti = noti
    module.Report.Diff = diff
    return module
}

// CheckRequired check if required file exist or not
func (r *Runner) CheckRequired(requires []string, options libs.Options) error {
    if len(requires) == 0 {
        return nil
    }
    for _, require := range requires {
        require = ResolveData(require, options.Scan.ROptions)

        if strings.Contains(require, "(") && strings.Contains(require, ")") {
            validate := r.ConditionExecScript(require)
            if !validate {
                utils.DebugF("Missing Requirement: %v", require)
                return fmt.Errorf("condition not met: %s", require)
            }
            continue
        }

        require = utils.NormalizePath(require)
        if !utils.FileExists(require) && utils.EmptyFile(require, 0) {
            if !utils.FolderExists(require) && utils.DirLength(require) > 0 {
                utils.DebugF("Missing %v", require)
                return fmt.Errorf("missing requirement")
            }
        }
    }
    return nil
}

// CheckCondition check if required file exist or not
func (r *Runner) CheckCondition(conditions []string, options libs.Options) error {
    if len(conditions) == 0 {
        return nil
    }
    for _, require := range conditions {
        require = ResolveData(require, options.Scan.ROptions)
        validate := r.ConditionExecScript(require)
        if !validate {
            return fmt.Errorf("condition not met: %s", require)
        }
    }
    return nil
}

// RunCommands run list of commands in parallel
func RunCommands(commands []string, std string, options libs.Options) string {
    var wg sync.WaitGroup
    var output string
    var err error

    for _, rawCommand := range commands {
        command := ResolveData(rawCommand, options.Scan.ROptions)
        wg.Add(1)
        // don't run too much command at once
        go func() {
            defer wg.Done()
            var out string
            if std != "" {
                out, err = utils.RunOSCommand(command)
            } else {
                err = utils.RunCommandWithoutOutput(command)
            }

            if err != nil {
                utils.DebugF("error running command: %v", command)
            }

            if out != "" {
                output += out
            }
        }()
    }
    wg.Wait()

    if std != "" {
        utils.WriteToFile(std, output)
    }
    return output
}

//  print all report
func printReports(options libs.Options) {
    var files []string
    files = append(files, options.Module.Report.Final...)
    files = append(files, options.Module.Report.Noti...)
    files = append(files, options.Module.Report.Diff...)

    reports := funk.UniqString(files)
    utils.BannerF("REPORT", options.Module.Name)
    for _, report := range reports {
        if !utils.FileExists(report) && utils.EmptyFile(report, 0) {
            if !utils.FolderExists(report) && utils.EmptyDir(report) {
                continue
            }
        }
        utils.BlockF("report", report)
    }
}
