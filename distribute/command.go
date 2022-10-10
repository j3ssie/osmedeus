package distribute

import (
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/database"
	"github.com/j3ssie/osmedeus/utils"
	"strings"
	"time"
)

func (c *CloudRunner) CloudMoreParams() {
	utils.DebugF("Parsed more params")

	// take from -m flag
	if c.Opt.Cloud.Module != "" {
		module := core.DirectSelectModule(c.Opt, c.Opt.Cloud.Module)
		parsedModule, err := core.ParseModules(module)
		if err != nil || parsedModule.Name == "" {
			return
		}

		for _, param := range parsedModule.Params {
			for k, v := range param {
				v = core.ResolveData(v, c.Target)
				if strings.HasPrefix(v, "~/") {
					v = utils.NormalizePath(v)
				}
				c.Target[k] = v
			}
		}

		return
	}

	flows := core.SelectFlow(c.Opt.Cloud.Flow, c.Opt)
	if len(flows) == 0 {
		return
	}

	parsedFlow, err := core.ParseFlow(flows[0])
	if err != nil {
		return
	}

	for _, param := range parsedFlow.Params {
		for k, v := range param {
			v = core.ResolveData(v, c.Target)
			if strings.HasPrefix(v, "~/") {
				v = utils.NormalizePath(v)
			}
			c.Target[k] = v
		}
	}

}

func (c *CloudRunner) CreateUIReport() {
	if c.Opt.NoDB {
		return
	}

	//utils.DebugF("Creating UI reports")

	// take from -m flag
	if c.Opt.Cloud.Module != "" {
		module := core.DirectSelectModule(c.Opt, c.Opt.Cloud.Module)
		parsedModule, err := core.ParseModules(module)
		if err != nil || parsedModule.Name == "" {
			return
		}

		// create record on UI
		c.Opt.Module = core.ResolveReports(parsedModule, c.Target)
		database.DBNewReports(parsedModule, &c.Runner.TargetObj)
		return
	}

	flows := core.SelectFlow(c.Opt.Cloud.Flow, c.Opt)
	if len(flows) == 0 {
		return
	}

	parsedFlow, err := core.ParseFlow(flows[0])
	if err != nil {
		return
	}

	for _, rawModules := range parsedFlow.Routines {
		// select module depend on it's flow
		if rawModules.FlowFolder != "" {
			parsedFlow.Type = rawModules.FlowFolder
		} else {
			parsedFlow.Type = parsedFlow.DefaultType
		}

		modules := core.SelectModules(rawModules.Modules, c.Opt)

		//var routine libs.Routine
		//routine.ModeName = parsedFlow.Name

		for _, module := range modules {
			parsedModule, err := core.ParseModules(module)
			if err != nil || parsedModule.Desc == "" {
				continue
			}
			// create record on UI
			c.Opt.Module = core.ResolveReports(parsedModule, c.Target)
			database.DBNewReports(c.Opt.Module, &c.Runner.TargetObj)
		}
	}
}

func (c *CloudRunner) RetryCommandWithExpectString(cmd string, expectString string, timeoutRaw ...string) string {
	timeout := "300s"
	if len(timeoutRaw) > 0 {
		timeout = timeoutRaw[0]
	}
	var out string

	utils.DebugF("Retry command: %s", cmd)
	for i := 0; i < c.Opt.Cloud.Retry; i++ {
		if timeout == "000" {
			out, _ = utils.RunOSCommand(cmd)
		} else {
			out = utils.RunCmdWithOutput(cmd, timeout)
		}

		if !strings.Contains(out, expectString) {
			utils.DebugF(out)
			time.Sleep(time.Duration(60*(i+1)) * time.Second)
			continue
		}
		return out
	}
	return out
}

func (c *CloudRunner) RetryCommandWithExcludeString(cmd string, excludeString string, timeoutRaw ...string) string {
	var out string
	timeout := "300s"
	if len(timeoutRaw) > 0 {
		timeout = timeoutRaw[0]
	}

	utils.DebugF("Retry command: %s", cmd)
	for i := 0; i < c.Opt.Cloud.Retry; i++ {
		if timeout == "000" {
			out, _ = utils.RunOSCommand(cmd)
		} else {
			out = utils.RunCmdWithOutput(cmd, timeout)
		}

		if strings.Contains(out, excludeString) {
			utils.DebugF(out)
			time.Sleep(time.Duration(60*(i+1)) * time.Second)
			continue
		}
		return out
	}
	return out
}

func (c *CloudRunner) RetryCommand(cmd string, timeoutRaw ...string) {
	timeout := "300s"
	if len(timeoutRaw) > 0 {
		timeout = timeoutRaw[0]
	}
	utils.DebugF("Retry command: %s", cmd)
	for i := 0; i < c.Opt.Cloud.Retry; i++ {
		out, err := utils.RunCommandWithErr(cmd, timeout)
		if err != nil {
			utils.DebugF(out)
			time.Sleep(time.Duration(60*(i+1)) * time.Second)
			continue
		}
		break
	}
}
