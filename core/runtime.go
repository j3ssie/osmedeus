package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/execution"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/robertkrimen/otto"
	"github.com/spf13/cast"
)

// InitVM init scripting engine
func (r *Runner) InitVM() {
	r.VM = otto.New()
	r.LoadEngineScripts()
}

func (r *Runner) ExecScript(script string) string {
	utils.DebugF("[Run-Scripts] %v", script)
	value, err := r.VM.Run(script)
	if err == nil {
		out, nerr := value.ToString()
		if nerr == nil {
			return out
		}
	}

	return ""
}

// RunOse really start the runner
func (r *Runner) RunOse(scriptName string) {
	scriptContent := scriptName
	if !strings.Contains(scriptName, "\n") {
		scriptFile := SelectScript(scriptName, r.Opt)
		if utils.FileExists(scriptFile) {
			scriptContent = utils.GetFileContent(scriptFile)
		}
	}

	if len(scriptContent) == 0 {
		utils.ErrorF("Error running script: %s", scriptName)
		return
	}

	utils.DebugF("-- Start ose:\n\n%s", scriptName)
	r.ExecScript(scriptContent)
	utils.DebugF("-- Done ose: %s", scriptName)
}

func (r *Runner) ConditionExecScript(script string) bool {
	utils.DebugF("[Run-Scripts] %v", script)
	value, err := r.VM.Run(script)

	if err == nil {
		out, nerr := value.ToBoolean()
		if nerr == nil {
			return out
		}
	}

	return false
}

func (r *Runner) LoadEngineScripts() {
	r.LoadScripts()
	r.LoadDBScripts()
	// r.LoadImportScripts()
	r.LoadExternalScripts()
	r.LoadGitScripts()
	r.LoadNotiScripts()
}

func (r *Runner) LoadScripts() string {
	var output string
	vm := r.VM

	// set attribute
	vm.Set("Target", r.Target)

	// SetVar('length', 6)
	vm.Set(SetVar, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		varName := args[0].String()
		value := args[1].String()
		r.Target[varName] = value
		return otto.Value{}
	})

	// Exit script used to exit the module
	vm.Set(Exit, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		output = fmt.Sprintf("exit(%v)", args[0].String())
		utils.InforF("Exit Detected")
		result, err := vm.ToValue(output)
		if err == nil {
			return result
		}
		return otto.Value{}
	})

	// ExecCmd execute command
	vm.Set(ExecCmd, func(call otto.FunctionCall) otto.Value {
		cmd := call.Argument(0).String()
		_, err := utils.RunCommandWithErr(cmd)
		var validate bool
		if err != nil {
			validate = true
		}
		result, err := vm.ToValue(validate)
		if err != nil {
			return otto.Value{}
		}
		return result
	})

	// Cat the file to stdout
	vm.Set(Cat, func(call otto.FunctionCall) otto.Value {
		filename := call.Argument(0).String()
		utils.InforF("Showing the content of: %v", color.HiCyanString(filename))
		utils.Cat(filename)
		result, err := vm.ToValue(true)
		if err != nil {
			return otto.Value{}
		}
		return result
	})

	// ExecCmdB execute in the background
	vm.Set(ExecCmdB, func(call otto.FunctionCall) otto.Value {
		cmd := call.Argument(0).String()
		go func() {
			utils.RunOSCommand(cmd)
		}()
		result, _ := vm.ToValue(true)
		return result
	})

	// ExecCmd execute command
	vm.Set(ExecCmdWithOutput, func(call otto.FunctionCall) otto.Value {
		utils.RunCommandSteamOutput(call.Argument(0).String())
		result, err := vm.ToValue(true)
		if err != nil {
			return otto.Value{}
		}
		return result
	})

	// ExecCmd execute command
	vm.Set(ExecContain, func(call otto.FunctionCall) otto.Value {
		out := utils.RunCmdWithOutput(call.Argument(0).String())
		expected := call.Argument(2).String()
		validate := strings.Contains(out, expected)
		result, err := vm.ToValue(validate)
		if err != nil {
			return otto.Value{}
		}
		return result
	})

	// CastToInt convert string to int
	vm.Set(CastToInt, func(call otto.FunctionCall) otto.Value {
		toInt := cast.ToInt(call.Argument(0).String())
		result, err := vm.ToValue(toInt)
		if err == nil {
			return result
		}
		return otto.Value{}
	})

	vm.Set(FileLength, func(call otto.FunctionCall) otto.Value {
		data := utils.FileLength(call.Argument(0).String())
		utils.DebugF("FileLength -- %v", data)
		result, err := vm.ToValue(data)
		if err == nil {
			return result
		}
		return otto.Value{}
	})

	vm.Set(IsFile, func(call otto.FunctionCall) otto.Value {
		data := utils.FileLength(call.Argument(0).String())
		var validate bool
		if data > 1 {
			validate = true
		}
		result, _ := vm.ToValue(validate)
		return result
	})

	// StripSlash strip last '/' of URL
	vm.Set(StripSlash, func(call otto.FunctionCall) otto.Value {
		raw := call.Argument(0).String()
		out := strings.Trim(raw, "/")
		result, err := vm.ToValue(out)
		if err != nil {
			return otto.Value{}
		}
		return result
	})

	vm.Set(ReadLines, func(call otto.FunctionCall) otto.Value {
		fileName := call.Argument(0).String()
		data := utils.ReadingLines(fileName)
		if len(data) > 0 {
			result, err := vm.ToValue(data)
			if err == nil {
				return result
			}
		}
		return otto.Value{}
	})

	// Printf simply print a string to console
	vm.Set(Printf, func(call otto.FunctionCall) otto.Value {
		fmt.Printf("%v\n", color.HiCyanString(call.Argument(0).String()))
		returnValue, _ := otto.ToValue(true)
		return returnValue
	})

	// split file to multiple
	vm.Set(SplitFile, func(call otto.FunctionCall) otto.Value {
		execution.SplitFile("size", call.ArgumentList)
		return otto.Value{}
	})

	// split file to multiple
	vm.Set(SplitFileByPart, func(call otto.FunctionCall) otto.Value {
		execution.SplitFile("part", call.ArgumentList)
		return otto.Value{}
	})

	vm.Set(Sleep, func(call otto.FunctionCall) otto.Value {
		execution.Sleep(call.Argument(0).String())
		returnValue, _ := otto.ToValue(true)
		return returnValue
	})

	vm.Set(SortU, func(call otto.FunctionCall) otto.Value {
		execution.SortU(call.Argument(0).String())
		returnValue, _ := otto.ToValue(true)
		return returnValue
	})

	vm.Set(Append, func(call otto.FunctionCall) otto.Value {
		dest := call.Argument(0).String()
		src := call.Argument(1).String()
		execution.Append(dest, src)
		returnValue, _ := otto.ToValue(true)
		return returnValue
	})

	vm.Set(Decompress, func(call otto.FunctionCall) otto.Value {
		dest := call.Argument(0).String()
		src := call.Argument(1).String()
		execution.Decompress(dest, src)
		returnValue, _ := otto.ToValue(true)
		return returnValue
	})

	vm.Set(Compress, func(call otto.FunctionCall) otto.Value {
		dest := call.Argument(0).String()
		src := call.Argument(1).String()
		execution.Compress(dest, src)
		returnValue, _ := otto.ToValue(true)
		return returnValue
	})

	vm.Set(CreateFolder, func(call otto.FunctionCall) otto.Value {
		utils.MakeDir(call.Argument(0).String())
		return otto.Value{}
	})

	vm.Set(DeleteFile, func(call otto.FunctionCall) otto.Value {
		execution.DeleteFile(call.Argument(0).String())
		return otto.Value{}
	})

	vm.Set(DeleteFolder, func(call otto.FunctionCall) otto.Value {
		execution.DeleteFolder(call.Argument(0).String())
		return otto.Value{}
	})

	vm.Set(Copy, func(call otto.FunctionCall) otto.Value {
		src := call.Argument(0).String()
		dest := call.Argument(1).String()
		execution.Copy(src, dest)
		return otto.Value{}
	})

	vm.Set(GetOSEnv, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		env := args[0].String()
		defaultValue := env
		if len(args) > 1 {
			defaultValue = args[1].String()
		}
		utils.GetOSEnv(env, defaultValue)
		return otto.Value{}
	})

	vm.Set(EmptyDir, func(call otto.FunctionCall) otto.Value {
		result, _ := vm.ToValue(utils.EmptyDir(call.Argument(0).String()))
		return result
	})

	vm.Set(EmptyFile, func(call otto.FunctionCall) otto.Value {
		result, err := vm.ToValue(utils.EmptyFile(call.Argument(0).String(), 0))
		if err != nil {
			return otto.Value{}
		}
		if len(call.ArgumentList) > 1 {
			num, _ := call.Argument(0).ToInteger()
			result, err = vm.ToValue(utils.EmptyFile(call.Argument(0).String(), int(num)))
			if err != nil {
				return otto.Value{}
			}
		}
		return result
	})

	vm.Set(RRSync, func(call otto.FunctionCall) otto.Value {
		vpsIP := call.Argument(0).String() // root@ipaddress
		src := call.Argument(1).String()   // local path
		dest := call.Argument(2).String()  // remote path
		cmd := fmt.Sprintf("rsync -e 'ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -i %s' -avzr --progress %s %s:%s", r.Opt.Cloud.SecretKey, src, vpsIP, dest)
		r.RetryCommandWithExpectString(cmd, `bytes/sec`)
		return otto.Value{}
	})

	r.VM = vm

	return output
}

func (r *Runner) RetryCommandWithExpectString(cmd string, expectString string, timeoutRaw ...string) string {
	timeout := "300s"
	if len(timeoutRaw) > 0 {
		timeout = timeoutRaw[0]
	}
	var out string

	utils.DebugF("Retry command: %s", cmd)
	for i := 0; i < r.Opt.Cloud.Retry; i++ {
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

func (r *Runner) LoadNotiScripts() string {
	var output string
	vm := r.VM
	options := r.Opt

	// script for notification
	vm.Set(StartNoti, func(call otto.FunctionCall) otto.Value {
		execution.StatusNoti("start", options)
		return otto.Value{}
	})
	vm.Set(DoneNoti, func(call otto.FunctionCall) otto.Value {
		execution.StatusNoti("done", options)
		return otto.Value{}
	})
	vm.Set(ReportNoti, func(call otto.FunctionCall) otto.Value {
		execution.ReportNoti(call.ArgumentList, options)
		return otto.Value{}
	})
	vm.Set(DiffNoti, func(call otto.FunctionCall) otto.Value {
		execution.DiffNoti(call.ArgumentList, options)
		return otto.Value{}
	})
	// CustomNoti("message here")
	vm.Set(CustomNoti, func(call otto.FunctionCall) otto.Value {
		execution.SendAttachment("custom", call.Argument(0).String(), options)
		return otto.Value{}
	})
	// NotiFile("src")
	vm.Set(NotiFile, func(call otto.FunctionCall) otto.Value {
		execution.SendFile(call.Argument(0).String(), options.Noti.SlackReportChannel, options)
		return otto.Value{}
	})
	// using webhook
	vm.Set(WebHookNoti, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		messContent := args[0].String()
		messType := "custom"
		if len(args) > 1 {
			messType = args[0].String()
			messContent = args[1].String()
		}

		execution.WebHookSendAttachment(options, messType, messContent)
		return otto.Value{}
	})

	// Telegram functions

	vm.Set(TeleMess, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		messContent := args[0].String()
		channelType := "general"
		if len(args) > 1 {
			channelType = args[0].String()
			messContent = args[1].String()
		}
		execution.TeleSendMess(options, messContent, channelType, false)
		return otto.Value{}
	})
	// send message but with inside ```
	vm.Set(TeleMessWrap, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		messContent := args[0].String()
		channelType := "general"
		if len(args) > 1 {
			channelType = args[0].String()
			messContent = args[1].String()
		}
		execution.TeleSendMess(options, messContent, channelType, true)
		return otto.Value{}
	})

	vm.Set(TeleMessByFile, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		fileName := args[0].String()
		channelType := "general"
		if len(args) > 1 {
			channelType = args[0].String()
			fileName = args[1].String()
		}
		if !utils.FileExists(fileName) {
			utils.DebugF("File %s not found", fileName)
			return otto.Value{}
		}

		messContent := utils.GetFileContent(fileName)
		if len(messContent) > 4000 {
			execution.TeleSendFile(options, fileName, channelType)
		} else {
			execution.TeleSendMess(options, messContent, channelType, true)
		}

		return otto.Value{}
	})

	vm.Set(TeleSendFile, func(call otto.FunctionCall) otto.Value {
		args := call.ArgumentList
		messContent := args[0].String()
		channelType := "general"
		if len(args) > 1 {
			channelType = args[0].String()
			messContent = args[1].String()
		}
		execution.TeleSendFile(options, messContent, channelType)
		return otto.Value{}
	})

	return output

}
