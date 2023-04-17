package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cast"
)

func ListWorkspaces(options libs.Options) (content [][]string) {
	workspaces, err := ioutil.ReadDir(utils.NormalizePath(options.Env.WorkspacesFolder))
	if err != nil {
		utils.ErrorF("Error reading workspaces folder: %s", err)
		return content
	}

	for _, ws := range workspaces {
		if ws.IsDir() {
			status := "unknown"
			flowName := "unknown"
			progress := "N/A"
			wsFolder := path.Join(utils.NormalizePath(options.Env.WorkspacesFolder), ws.Name())

			if utils.DirLength(wsFolder) == 0 {
				continue
			}

			if utils.FileExists(path.Join(wsFolder, "done")) {
				status = "done"
			}

			runtimeFile := path.Join(wsFolder, "runtime")
			if utils.FileExists(runtimeFile) {
				utils.DebugF("Reading information from: %v", runtimeFile)
				runtimeContent := utils.GetFileContent(runtimeFile)

				if jsonParsed, ok := gabs.ParseJSON([]byte(runtimeContent)); ok == nil {
					flowName = cast.ToString(jsonParsed.S("task_name").Data())
					doneStep := cast.ToString(jsonParsed.S("done_step").Data())
					totalSteps := cast.ToString(jsonParsed.S("total_steps").Data())
					isRunning := cast.ToString(jsonParsed.S("is_running").Data())

					if isRunning == "true" {
						status = "running"
					}

					progress = color.HiCyanString(fmt.Sprintf(doneStep + "/" + totalSteps))
				}
			}

			row := []string{
				path.Base(color.HiMagentaString(ws.Name())), status, flowName, progress, wsFolder,
			}
			content = append(content, row)
		}
	}

	table := tablewriter.NewWriter(os.Stderr)
	table.SetAutoFormatHeaders(false)
	table.SetHeader([]string{"Workspace Name", "Status", "Routine", "Progress", "Workspace Path"})
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetColWidth(120)
	table.AppendBulk(content)
	table.Render()

	fmt.Println(color.HiGreenString("📁 Total Workspaces: ") + color.HiMagentaString("%v", len(content)))
	usage := color.HiWhiteString("💡 How to view report:") + color.HiGreenString(" osmedeus report view -t %v", color.HiMagentaString("[targetName]"))
	fmt.Println(usage)

	return content
}

func ListSingleWorkspace(options libs.Options, target string) (content [][]string) {
	workspaces, err := ioutil.ReadDir(utils.NormalizePath(options.Env.WorkspacesFolder))
	if err != nil {
		utils.ErrorF("Error reading workspaces folder: %s", err)
		return content
	}

	header := []string{"Module", "Report Name", "Report Path"}
	for _, ws := range workspaces {
		if !ws.IsDir() {
			continue
		}
		// compare target name with workspace name
		if target != path.Base(ws.Name()) {
			continue
		}
		wsFolder := path.Join(utils.NormalizePath(options.Env.WorkspacesFolder), ws.Name())

		runtimeFile := path.Join(wsFolder, "runtime")
		// only listing file that in report part
		if utils.FileExists(runtimeFile) && !options.Report.Raw {
			utils.InforF("Reading information from: %v", runtimeFile)
			runtimeContent := utils.GetFileContent(runtimeFile)

			isImported := false
			if !strings.Contains(runtimeContent, options.Env.WorkspacesFolder) {
				isImported = true
			}

			// replace the workspace folder if it doesn't exist
			if !strings.Contains(runtimeContent, options.Env.WorkspacesFolder) {
				homeFolder := "/root/.osmedeus/workspaces/"
				if strings.Contains(runtimeContent, "/root/workspaces-osmedeus/") {
					homeFolder = "/root/workspaces-osmedeus/"
				}
				runtimeContent = strings.ReplaceAll(runtimeContent, homeFolder, options.Env.WorkspacesFolder+"/")
			}

			if strings.Contains(runtimeContent, "/root/.osmedeus/workspaces") {
				runtimeContent = strings.ReplaceAll(runtimeContent, "/root/.osmedeus/workspaces", options.Env.WorkspacesFolder)
			}

			row := []string{"==> Workspace Name", color.HiGreenString(ws.Name()), color.HiGreenString(wsFolder)}
			content = append(content, row)

			if jsonParsed, ok := gabs.ParseJSON([]byte(runtimeContent)); ok == nil {
				reports := jsonParsed.S("target", "reports").Children()

				for _, report := range reports {
					moduleName := cast.ToString(report.S("module").Data())
					reportName := cast.ToString(report.S("report_name").Data())
					reportPath := cast.ToString(report.S("report_path").Data())

					if !utils.FileExists(reportPath) {
						// /root/.osmedeus/workspaces
						continue
					}

					row := []string{
						moduleName, processReport(options, reportName), processReport(options, reportPath),
					}
					content = append(content, row)
				}
				sep := []string{"==> --------", color.HiGreenString("----------"), color.HiGreenString("-----------")}
				content = append(content, sep)

				if len(content) <= 2 && isImported {
					utils.WarnF("Workspace folder not found in runtime file, you might extract it from different machine that being run on different user")
					utils.WarnF("💡 If you still have problem, please view it as raw format: %s", color.HiGreenString("osmedeus view -t %s --raw", target))
				}

			}
			continue
		}

		// --raw flag: list all avaliable file
		header = []string{"Report Name", "Report Path"}
		if options.Report.Static {
			header = []string{"Report Name", "Report URL"}
		}
		row := []string{"==> Workspace Name", color.HiGreenString(ws.Name())}
		content = append(content, row)

		filepath.Walk(wsFolder, func(reportPath string, _ os.FileInfo, err error) error {
			reportName := path.Base(reportPath)
			row := []string{processReport(options, reportName), processReport(options, reportPath)}
			content = append(content, row)
			return nil
		})

		sep := []string{"==> --------", color.HiGreenString("-----------")}
		content = append(content, sep)

	}

	table := tablewriter.NewWriter(os.Stderr)
	table.SetAutoFormatHeaders(true)
	table.SetHeader(header)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetColWidth(120)
	table.AppendBulk(content)
	table.SetHeaderLine(true)
	table.Render()

	return content
}

func processReport(options libs.Options, reportPath string) string {
	if options.Report.Static {
		base := fmt.Sprintf("https://%s:8000/%s/workspaces", options.Report.PublicIP, options.Server.StaticPrefix)
		reportPath = strings.ReplaceAll(reportPath, options.Env.WorkspacesFolder, base)
	}

	if strings.HasSuffix(reportPath, ".html") {
		reportPath = color.HiCyanString(reportPath)
	}
	if strings.HasSuffix(reportPath, ".json") {
		reportPath = color.HiBlueString(reportPath)
	}

	return reportPath
}
