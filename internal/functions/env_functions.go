package functions

import (
	"os"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// osGetenv gets an environment variable
// Usage: os_getenv(name) -> string
// name: environment variable name
// Returns: string value or empty string if not set
func (vf *vmFunc) osGetenv(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("os_getenv"))

	if len(call.Arguments) < 1 {
		logger.Get().Warn("os_getenv: requires 1 argument")
		return vf.vm.ToValue("")
	}

	name := call.Argument(0).String()
	if name == "" || name == "undefined" {
		logger.Get().Warn("os_getenv: name cannot be empty")
		return vf.vm.ToValue("")
	}

	logger.Get().Debug(terminal.HiGreen("os_getenv")+" params",
		zap.String("name", name))

	value := os.Getenv(name)
	return vf.vm.ToValue(value)
}

// osSetenv sets an environment variable
// Usage: os_setenv(name, value) -> bool
// name: environment variable name
// value: value to set
// Returns: true on success, false on failure
func (vf *vmFunc) osSetenv(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("os_setenv"))

	if len(call.Arguments) < 2 {
		logger.Get().Warn("os_setenv: requires 2 arguments")
		return vf.vm.ToValue(false)
	}

	name := call.Argument(0).String()
	value := call.Argument(1).String()

	if name == "" || name == "undefined" {
		logger.Get().Warn("os_setenv: name cannot be empty")
		return vf.vm.ToValue(false)
	}

	// Allow "undefined" as value to unset (set to empty)
	if value == "undefined" {
		value = ""
	}

	logger.Get().Debug(terminal.HiGreen("os_setenv")+" params",
		zap.String("name", name),
		zap.String("value", value))

	err := os.Setenv(name, value)
	if err != nil {
		logger.Get().Warn("os_setenv: failed to set env var", zap.Error(err))
		return vf.vm.ToValue(false)
	}

	return vf.vm.ToValue(true)
}
