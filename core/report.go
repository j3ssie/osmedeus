package core

import (
    "fmt"
    "github.com/Jeffail/gabs/v2"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/olekukonko/tablewriter"
    "github.com/spf13/cast"
    "io/ioutil"
    "os"
    "path"
    "path/filepath"
    "strings"
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
                path.Base(ws.Name()), status, flowName, progress, wsFolder,
            }
            content = append(content, row)
        }
    }

    table := tablewriter.NewWriter(os.Stderr)
    table.SetAutoFormatHeaders(false)
    table.SetHeader([]string{"Workspace Name", "Flow", "Status", "Progress", "Workspace Path"})
    table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
    table.SetColWidth(120)
    table.AppendBulk(content)
    table.Render()

    fmt.Println(color.HiGreenString("📁 Total Workspaces: ") + color.HiMagentaString("%v", len(content)))
    return content
}

func ListSingleWorkspace(options libs.Options, target string) (content [][]string) {
    workspaces, err := ioutil.ReadDir(utils.NormalizePath(options.Env.WorkspacesFolder))
    if err != nil {
        utils.ErrorF("Error reading workspaces folder: %s", err)
        return content
    }

    header := []string{"Workspace Name", "Module", "Report Name", "Report Path"}

    for _, ws := range workspaces {
        if !ws.IsDir() {
            continue
        }
        if target != path.Base(ws.Name()) {
            continue
        }
        wsFolder := path.Join(utils.NormalizePath(options.Env.WorkspacesFolder), ws.Name())

        runtimeFile := path.Join(wsFolder, "runtime")
        if utils.FileExists(runtimeFile) && !options.Report.Raw {
            utils.InforF("Reading information from: %v", runtimeFile)
            runtimeContent := utils.GetFileContent(runtimeFile)

            if jsonParsed, ok := gabs.ParseJSON([]byte(runtimeContent)); ok == nil {
                reports := jsonParsed.S("target", "reports").Children()

                for _, report := range reports {
                    moduleName := cast.ToString(report.S("module").Data())
                    reportName := cast.ToString(report.S("report_name").Data())
                    reportPath := cast.ToString(report.S("report_path").Data())

                    row := []string{
                        ws.Name(), moduleName, processReport(options, reportName), processReport(options, reportPath),
                    }

                    content = append(content, row)
                }
            }
            continue
        }

        header = []string{"Workspace Name", "Report Name", "Report Path"}
        if options.Report.Static {
            header = []string{"Workspace Name", "Report Name", "Report URL"}
        }
        filepath.Walk(wsFolder, func(reportPath string, _ os.FileInfo, err error) error {
            reportName := path.Base(reportPath)
            row := []string{
                ws.Name(), processReport(options, reportName), processReport(options, reportPath),
            }
            content = append(content, row)
            return nil
        })
    }

    table := tablewriter.NewWriter(os.Stderr)
    table.SetAutoFormatHeaders(false)
    table.SetHeader(header)
    table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
    table.SetColWidth(120)
    table.AppendBulk(content)
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
