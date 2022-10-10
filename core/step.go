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
		r.RunModule(module)
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
func CheckResume(module libs.Module) bool {
	for _, report := range module.Report.Final {
		if !strings.Contains(report, ".osmedeus/storages") && !utils.FileExists(report) {
			return false
		}
	}
	return true
}

// ResolveReports resolve real path of reports
func ResolveReports(module libs.Module, params map[string]string) libs.Module {
	var final []string
	var noti []string
	var diff []string
	for _, report := range module.Report.Final {
		final = append(final, ResolveData(report, params))
	}

	for _, report := range module.Report.Noti {
		noti = append(noti, ResolveData(report, params))
	}

	for _, report := range module.Report.Diff {
		diff = append(diff, ResolveData(report, params))
	}

	module.Report.Final = final
	module.Report.Noti = noti
	module.Report.Diff = diff
	return module
}

//  print all report
func printReports(module libs.Module) {
	var files []string
	files = append(files, module.Report.Final...)
	files = append(files, module.Report.Noti...)
	files = append(files, module.Report.Diff...)

	reports := funk.UniqString(files)
	utils.BannerF("Report", module.Name)
	for _, report := range reports {
		if !utils.FileExists(report) && utils.EmptyFile(report, 0) {
			if !utils.FolderExists(report) && utils.EmptyDir(report) {
				continue
			}
		}
		utils.BlockF("report", report)
	}
}

// CheckRequired check if required file exist or not
func (r *Runner) CheckRequired(requires []string) error {
	if len(requires) == 0 {
		return nil
	}

	utils.DebugF("Checking require: %v", requires)
	for _, require := range requires {
		//require = ResolveData(require, options.Scan.ROptions)

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
			if !utils.FolderExists(require) && utils.DirLength(require) == 0 {
				utils.DebugF("Missing %v", require)
				return fmt.Errorf("missing requirement")
			}
		}
	}
	return nil
}

// CheckCondition check if required file exist or not
func (r *Runner) CheckCondition(conditions []string) error {
	if len(conditions) == 0 {
		return nil
	}
	for _, require := range conditions {
		if !r.ConditionExecScript(require) {
			return fmt.Errorf("condition not met: %s", require)
		}
	}
	return nil
}

// RunCommands run list of commands in parallel
func (r *Runner) RunCommands(commands []string, std string) string {
	var wg sync.WaitGroup
	var output string
	var err error

	for _, command := range commands {
		wg.Add(1)
		// don't run too much command at once
		go func(command string) {
			defer wg.Done()
			var out string
			if std != "" {
				out, err = utils.RunOSCommand(command)
			} else {
				err = utils.RunCommandWithoutOutput(command)
			}

			if err != nil {
				utils.DebugF("error running command: %v -- %v", command, err)
			}

			if out != "" {
				output += out
			}
		}(command)
	}
	wg.Wait()

	if std != "" {
		utils.WriteToFile(std, output)
	}
	return output
}
