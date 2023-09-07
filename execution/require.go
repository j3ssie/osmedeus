package execution

//
//// RunRequire is main function for generator
//func RunRequire(script string, options libs.Options) bool {
//    // @NOTE: for some reason < auto translate to &lt; in golang template
//    if strings.Contains(script, "&lt;") {
//        script = strings.Replace(script, "&lt;", "<", -1)
//    }
//    utils.DebugF("[Run-Require] %v", script)
//    vm := otto.New()
//
//    vm.Set("EmptyDir", func(call otto.FunctionCall) otto.Value {
//        result, _ := vm.ToValue(utils.EmptyDir(call.Argument(0).String()))
//        return result
//    })
//
//    vm.Set("NotEmptyDir", func(call otto.FunctionCall) otto.Value {
//        result, _ := vm.ToValue(!utils.EmptyDir(call.Argument(0).String()))
//        return result
//    })
//
//    vm.Set("NotEmptyFile", func(call otto.FunctionCall) otto.Value {
//        result, _ := vm.ToValue(!utils.EmptyFile(call.Argument(0).String(), 0))
//        if len(call.ArgumentList) > 1 {
//            num, _ := call.Argument(0).ToInteger()
//            result, _ = vm.ToValue(!utils.EmptyFile(call.Argument(0).String(), int(num)))
//        }
//        return result
//    })
//
//    vm.Set("EmptyFile", func(call otto.FunctionCall) otto.Value {
//        result, _ := vm.ToValue(utils.EmptyFile(call.Argument(0).String(), 0))
//        if len(call.ArgumentList) > 1 {
//            num, _ := call.Argument(0).ToInteger()
//            result, _ = vm.ToValue(utils.EmptyFile(call.Argument(0).String(), int(num)))
//        }
//        return result
//    })
//
//    vm.Set("ExecContain", func(call otto.FunctionCall) otto.Value {
//        var validate bool
//        args := call.ArgumentList
//        cmd := args[0].String()
//        search := args[1].String()
//        out, _ := Execution(cmd, options)
//        if strings.Contains(out, search) {
//            validate = true
//        }
//        result, _ := vm.ToValue(validate)
//        return result
//    })
//
//    vm.Set("ExecMatch", func(call otto.FunctionCall) otto.Value {
//        var validate bool
//        args := call.ArgumentList
//        cmd := args[0].String()
//        search := args[1].String()
//        out, _ := Execution(cmd, options)
//        validate = RegexCount(out, search)
//        result, _ := vm.ToValue(validate)
//        return result
//    })
//
//    vm.Set("DirLength", func(call otto.FunctionCall) otto.Value {
//        validate := utils.DirLength(call.Argument(0).String())
//        result, err := vm.ToValue(validate)
//        if err != nil {
//            return otto.FalseValue()
//        }
//        return result
//    })
//
//    vm.Set("FileLength", func(call otto.FunctionCall) otto.Value {
//        validate := utils.FileLength(call.Argument(0).String())
//        result, err := vm.ToValue(validate)
//        if err != nil {
//            return otto.FalseValue()
//        }
//        return result
//    })
//
//    result, serr := vm.Run(script)
//    if serr != nil {
//        return false
//    }
//    analyzeResult, err := result.Export()
//    if err != nil || analyzeResult == nil {
//        return false
//    }
//    utils.DebugF("Required: %v -- %v", script, result)
//    return analyzeResult.(bool)
//}
//
//// RegexCount count regex string in component
//func RegexCount(component string, analyzeString string) bool {
//    r, err := regexp.Compile(analyzeString)
//    if err != nil {
//        return false
//    }
//    matches := r.FindAllStringIndex(component, -1)
//    if len(matches) > 0 {
//        return true
//    }
//    return false
//}
