package core

import (
	"fmt"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/thoas/go-funk"
	"path"
	"path/filepath"
	"strings"

	"github.com/j3ssie/osmedeus/libs"
)

// ListFlow list all available mode
func ListFlow(options libs.Options) (result []string) {
	modePath := path.Join(options.Env.WorkFlowsFolder, "/*.yaml")
	result, err := filepath.Glob(modePath)
	if err != nil {
		return result
	}
	return result
}

// SelectFlow select flow to run
func SelectFlow(flowName string, options libs.Options) []string {
	flows := ListFlow(options)
	var selectedFlow []string

	// absolute path like -f customflows/general.yaml
	if strings.HasSuffix(flowName, ".yaml") {
		if utils.FileExists(flowName) {
			selectedFlow = append(selectedFlow, flowName)
			return selectedFlow
		}
	}

	// -f test
	if !strings.Contains(flowName, ",") {
		selectedFlow = append(selectedFlow, singleMode(flowName, flows)...)
	}

	// -f test1,test2
	flowNames := strings.Split(flowName, ",")
	for _, item := range flowNames {
		selectedFlow = append(selectedFlow, singleMode(item, flows)...)
	}

	// default custom flow folder
	if !utils.FileExists(flowName) {
		flowName = path.Join(options.Env.WorkFlowsFolder, "default-flows", flowName)
		if utils.FileExists(flowName) {
			selectedFlow = append(selectedFlow, flowName)
		} else if utils.FileExists(flowName + ".yaml") {
			flowName = flowName + ".yaml"
			selectedFlow = append(selectedFlow, flowName)
		}
	}

	selectedFlow = funk.UniqString(selectedFlow)
	return selectedFlow
}

func singleMode(modeName string, modes []string) (selectedMode []string) {
	for _, mode := range modes {
		basemodeName := strings.TrimRight(strings.TrimRight(filepath.Base(mode), "yaml"), ".")
		// select workflow file in workflow directory
		if strings.ToLower(basemodeName) == strings.ToLower(modeName) {
			selectedMode = append(selectedMode, mode)
		}
	}
	return selectedMode
}

// ListModules list all available module
func ListModules(options libs.Options) (modules []string) {
	modePath := path.Join(options.Env.WorkFlowsFolder, "general/*.yaml")
	if options.Flow.Type != "" {
		modePath = path.Join(options.Env.WorkFlowsFolder, fmt.Sprintf("%v/*.yaml", options.Flow.Type))
	}
	if strings.HasSuffix(options.Scan.Flow, ".yaml") {
		if options.Flow.Type == "" {
			options.Flow.Type = "general"
		}
		modePath = path.Join(path.Dir(options.Scan.Flow), options.Flow.Type) + "/*.yaml"
	}
	modules, err := filepath.Glob(modePath)
	if err != nil {
		return modules
	}
	return modules
}

// SelectModules return list of modules name
func SelectModules(moduleNames []string, options libs.Options) []string {
	if strings.Contains(options.Flow.Type, "{{.") {
		options.Flow.Type = ResolveData(options.Flow.Type, options.Scan.ROptions)
	}
	modules := ListModules(options)
	var selectedModules []string
	for _, item := range moduleNames {
		selectedModules = append(selectedModules, singleSelectModule(item, modules)...)
	}
	selectedModules = funk.UniqString(selectedModules)

	utils.DebugF("Select module name %v: %v", moduleNames, selectedModules)
	return selectedModules
}

func singleSelectModule(moduleName string, modules []string) (selectedModules []string) {
	for _, module := range modules {
		baseModuleName := strings.Trim(strings.TrimRight(filepath.Base(module), "yaml"), ".")
		if strings.ToLower(baseModuleName) == strings.ToLower(moduleName) {
			selectedModules = append(selectedModules, module)
		}
	}
	return selectedModules
}

// DefaultWorkflows select module from ~/.osmedeus/core/workflow/plugins/
func DefaultWorkflows(options libs.Options) []string {
	defaultModule := path.Join(options.Env.WorkFlowsFolder, "default-modules")
	modePath := path.Join(defaultModule, "/*.yaml")
	results, err := filepath.Glob(modePath)
	if err != nil {
		utils.ErrorF("No default module found in %v", defaultModule)
		return []string{}
	}
	return results
}

// DirectSelectModule select module from ~/osmedeus-base/workflow/default-modules
func DirectSelectModule(options libs.Options, moduleName string) string {
	// got absolutely path
	if utils.FileExists(moduleName) {
		return moduleName
	}

	// select in cloud folder first if we're running the cloud scan
	// ~/.osmedeus/core/workflow/cloud-modules/
	basePlugin := path.Join(options.Env.WorkFlowsFolder, "cloud-modules")
	modulePath := path.Join(basePlugin, moduleName)
	if utils.FileExists(modulePath) {
		utils.DebugF("Load module path: %v", modulePath)
		return modulePath
	}

	modulePath = path.Join(basePlugin, moduleName+".yaml")
	if utils.FileExists(modulePath) {
		utils.DebugF("Load module path: %v", modulePath)
		return modulePath
	}

	// ~/.osmedeus/core/workflow/default-modules/
	basePlugin = path.Join(options.Env.WorkFlowsFolder, "default-modules")
	modulePath = path.Join(basePlugin, moduleName)
	utils.DebugF("Load module path: %v", modulePath)
	if utils.FileExists(modulePath) {
		return modulePath
	}

	modulePath = path.Join(basePlugin, moduleName+".yaml")
	utils.DebugF("Load module path: %v", modulePath)
	if utils.FileExists(modulePath) {
		return modulePath
	}
	utils.DebugF("No plugin found with: %v", moduleName)
	return ""
}

// ListScripts list all available mode
func ListScripts(options libs.Options) (result []string) {
	modePath := path.Join(options.Env.OseFolder, "/*.js")
	result, err := filepath.Glob(modePath)
	if err != nil {
		return result
	}

	modePath = path.Join(options.Env.OseFolder, "/*/*.js")
	DepthResult, err := filepath.Glob(modePath)
	if err == nil {
		result = append(result, DepthResult...)
	}
	return result
}

func SelectScript(scriptName string, options libs.Options) string {
	scripts := ListScripts(options)
	for _, script := range scripts {
		if strings.Contains(scriptName, "/") {
			if strings.HasSuffix(script, scriptName) || strings.HasSuffix(script, scriptName+".js") {
				return script
			}
		}

		compareName := path.Base(script)
		if compareName == scriptName {
			return script
		}
		if compareName == fmt.Sprintf("%s.js", scriptName) {
			return script
		}
	}
	return ""
}
