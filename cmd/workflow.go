package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func init() {

	var workflowCmd = &cobra.Command{
		Use:     "workflow",
		Aliases: []string{"wf", "wl", "workflows", "wfs", "work", "works"},
		Short:   "Listing all available workflows",
		Long:    core.Banner(),
	}

	var workflowListCmd = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "Listing all available workflows",
		Long:    core.Banner(),
		RunE:    runWorkflow,
	}

	var workflowViewCmd = &cobra.Command{
		Use:     "view",
		Aliases: []string{"viwe", "ve", "vi", "v"},
		Short:   "View details of a workflow",
		Long:    core.Banner(),
		RunE:    runWorkflowView,
	}
	workflowViewCmd.Flags().Bool("all", false, "View all of the workflows")

	workflowCmd.AddCommand(workflowViewCmd)
	workflowCmd.AddCommand(workflowListCmd)
	workflowCmd.SetHelpFunc(UtilsHelp)
	RootCmd.AddCommand(workflowCmd)

	workflowCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if options.FullHelp {
			cmd.Help()
			os.Exit(0)
		}
	}
}

func runWorkflow(cmd *cobra.Command, _ []string) error {
	listFlows()
	fmt.Printf("\n------------------------------------------------------------\n")
	listDefaultModules()
	fmt.Printf("💡 For full help message, please run: %s or %s\n", color.GreenString("osmedeus --hh"), color.GreenString("osmedeus scan --hh"))
	return nil
}

func runWorkflowView(cmd *cobra.Command, _ []string) error {
	allFlows := core.ListFlow(options)
	viewAll, _ := cmd.Flags().GetBool("all")

	if viewAll {
		for _, flow := range allFlows {
			err := viewWorkflow(flow)
			if err != nil {
				utils.ErrorF("Error viewing workflow: %v", err)
			}
			fmt.Printf("\n------------------------------------------------------------\n\n")
		}
	} else {
		err := viewWorkflow(options.Scan.Flow)
		if err != nil {
			utils.ErrorF("Error viewing workflow: %v", err)
		}
	}

	h := color.HiCyanString("\n📄 Sample Usage:\n")
	h += color.HiGreenString(" osmedeus scan -f %v", color.HiMagentaString(options.Scan.Flow)) + color.HiGreenString(" -t ") + color.HiMagentaString("[target]") + "\n"
	h += color.HiGreenString(" osmedeus scan -f %v", color.HiMagentaString(options.Scan.Flow)) + color.HiGreenString(" -t ") + color.HiMagentaString("[target]") + color.HiGreenString(" -p ") + color.HiMagentaString("'enableSomething=false'") + "\n\n"
	fmt.Printf(h)

	fmt.Printf("💡 To list all of the workflows available, please run: %s\n", color.GreenString("osmedeus workflow ls"))
	fmt.Printf("💡 For full help message, please run: %s or %s\n", color.GreenString("osmedeus --hh"), color.GreenString("osmedeus scan --hh"))
	return nil
}

func viewWorkflow(workflowName string) error {
	fmt.Printf("📖 Viewing workflow detail: %v\n\n", color.GreenString(workflowName))
	allFlows := core.ListFlow(options)
	flows := core.SelectFlow(workflowName, options)
	if len(flows) == 0 {
		utils.ErrorF("Flow not found in any of existing workflow [%v]", color.HiYellowString(strings.Join(allFlows, ", ")))
		return fmt.Errorf("Flow %s not found", workflowName)
	}
	selectedWorkflow := flows[0]

	var content [][]string
	parsedFlow, err := core.ParseFlow(selectedWorkflow)
	if err != nil {
		utils.ErrorF("Error parsing flow: %v", selectedWorkflow)
		return err
	}

	var totalSteps, totalModules int
	parameters := make(map[string]string)
	for _, param := range parsedFlow.Params {
		for k, v := range param {
			parameters[k] = v
		}
	}

	for _, routine := range parsedFlow.Routines {
		// select module depend on the flow type
		if routine.FlowFolder != "" {
			parsedFlow.Type = routine.FlowFolder
		} else {
			parsedFlow.Type = parsedFlow.DefaultType
		}

		modules := core.SelectModules(routine.Modules, options)

		// loop through all modules to get the parameters
		for _, module := range modules {
			parsedModule, err := core.ParseModules(module)
			if err != nil || parsedModule.Name == "" {
				continue
			}
			for _, param := range parsedModule.Params {
				for k, v := range param {

					_, exist := parameters[k]
					if parsedFlow.ForceParams && exist {
						utils.DebugF("Skip override param: %v --> %v", k, v)
						continue
					}
					parameters[k] = v
				}

			}
			totalSteps += len(parsedModule.Steps)
			totalModules++
		}
	}

	var toggleFlags, skippingFlags []string
	for key, value := range parameters {
		if value == "true" {
			value = color.GreenString(value)
		} else if value == "false" {
			value = color.RedString(value)
		} else {

			value = color.CyanString(value)
		}

		if strings.HasPrefix(key, "enable") {
			toggleFlags = append(toggleFlags, fmt.Sprintf("%v=%v", key, value))
		}

		if strings.HasPrefix(key, "skip") {
			skippingFlags = append(skippingFlags, fmt.Sprintf("%v=%v", key, value))
		}
	}

	workflowInfo := fmt.Sprintf("Name: %v", color.HiCyanString(parsedFlow.Name)) + ", " + fmt.Sprintf("Total Steps: %v", color.HiCyanString("%v", totalSteps)) + ", " + fmt.Sprintf("Total Modules: %v", color.HiCyanString("%v", totalModules))
	content = append(content, []string{
		"Workflow Information", workflowInfo,
	})
	content = append(content, []string{
		"Description", parsedFlow.Desc,
	})

	content = append(content, []string{
		"Toggleable parameter", strings.Join(toggleFlags, ", "),
	})

	content = append(content, []string{
		"Skippable parameter", strings.Join(skippingFlags, ", "),
	})

	if parsedFlow.Usage != "" {
		content = append(content, []string{
			"Examples Commands", strings.TrimSpace(parsedFlow.Usage),
		})
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetColWidth(120)
	table.SetAutoWrapText(false)
	table.AppendBulk(content)
	table.Render()

	return nil
}
