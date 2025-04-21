package core

import (
	"fmt"

	"github.com/fatih/color"

	"github.com/j3ssie/osmedeus/utils"
)

func (r *Runner) GenerateFlowChart() {
	// spew.Dump(r.Routines)

	moduleGroups := make(map[string][]string)
	for index, routine := range r.Routines {
		moduleGroup := []string{}
		for _, module := range routine.ParsedModules {
			moduleGroup = append(moduleGroup, module.Name)
		}
		moduleGroups[fmt.Sprintf("%d", index)] = moduleGroup
	}

	// Generate flowchart
	flowchart := "flowchart LR\n"
	for index, modules := range moduleGroups {
		if len(modules) == 0 {
			continue
		}

		// Create nodes for each module in the group
		for i, module := range modules {
			currentNode := fmt.Sprintf("\t%s_%d[%s]", index, i, module)
			if len(modules) > 1 {
				currentNode = fmt.Sprintf("\t%s_%d(%s)", index, i, module)
			}
			flowchart += currentNode + "\n"

			// Connect to the next module in the same group
			if i < len(modules)-1 {
				flowchart += fmt.Sprintf("\t%s_%d --> %s_%d\n", index, i, index, i+1)
			}

			// If this is the last module in the group, connect to the first module of the next group
			if i == len(modules)-1 {
				nextGroupIndex := fmt.Sprintf("%d", StringToInt(index)+1)
				if nextModules, exists := moduleGroups[nextGroupIndex]; exists && len(nextModules) > 0 {
					flowchart += fmt.Sprintf("\t%s_%d --> %s_0\n", index, i, nextGroupIndex)
				}
			}
		}
	}

	// fmt.Println(flowchart)

	flowChartFile := r.WorkspaceFolder + "/scan-flowchart.mermaid"
	utils.InforF("Generate flowchart to: %v", color.HiCyanString("%v", flowChartFile))
	utils.WriteToFile(flowChartFile, flowchart)
}

func StringToInt(s string) int {
	val := 0
	fmt.Sscanf(s, "%d", &val)
	return val
}
