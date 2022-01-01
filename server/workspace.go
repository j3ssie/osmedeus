package server

import (
    "fmt"
    "github.com/gofiber/fiber/v2"
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/database"
    "github.com/j3ssie/osmedeus/execution"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/thoas/go-funk"
    "gorm.io/gorm/clause"
    "os"
    "path"
    "path/filepath"
    "strings"
)

// ResponseHTTP represents response body of this API
type ResponseHTTP struct {
    Status  int         `json:"status"`
    Data    interface{} `json:"data"`
    Type    string      `json:"type,omitempty"`
    Total   int         `json:"total,omitempty"`
    Message string      `json:"message"`
}

// Workspace is a function to get all books data from database
// @Summary Get all books
// @Description Get all books
// @Tags books
// @Accept json
// @Produce json
// @Success 200 {object} ResponseHTTP{}
// @Failure 503 {object} ResponseHTTP{}
// @Router /v1/books [get]
func Workspace(c *fiber.Ctx) error {
    targets := database.GetWorkspaces()
    return c.JSON(ResponseHTTP{
        Status:  200,
        Data:    targets,
        Type:    "workspaces",
        Total:   len(targets),
        Message: "List all of Workspaces",
    })
}

func WorkspaceDetail(c *fiber.Ctx) error {
    wsname := c.Params("wsname")

    var target database.Target
    DB.Preload(clause.Associations).Preload("Reports").First(&target, "workspace = ?", wsname)

    var scan database.Scan
    DB.Preload(clause.Associations).Preload("Targets").First(&scan, "target_refer = ?", target.ID)

    // make a reports map
    rawReports := target.Reports
    reports := make(map[string][]string)
    for _, report := range rawReports {
        if utils.FileLength(report.ReportPath) == 0 {
            if utils.FolderExists(report.ReportPath) && utils.FolderLength(report.ReportPath) == 0 {
                continue
            } else {
                continue
            }
        }

        //reportPath := strings.ReplaceAll(report.ReportPath, Opt.Env.WorkspacesFolder, Opt.Server.StaticPrefix)
        //reports[report.Module] = append(reports[report.Module], reportPath)

        if strings.HasPrefix(report.ReportPath, Opt.Env.WorkspacesFolder) {
            report.ReportPath = strings.ReplaceAll(report.ReportPath, Opt.Env.WorkspacesFolder, fmt.Sprintf("/%v/workspaces", Opt.Server.StaticPrefix))
        } else if strings.HasPrefix(report.ReportPath, Opt.Env.StoragesFolder) {
            report.ReportPath = strings.ReplaceAll(report.ReportPath, Opt.Env.StoragesFolder, fmt.Sprintf("/%v/storages", Opt.Server.StaticPrefix))
        }

        if !funk.Contains(reports[report.Module], report.ReportPath) {
            reports[report.Module] = append(reports[report.Module], report.ReportPath)
        }
    }

    return c.JSON(ResponseHTTP{
        Status: 200,
        Data: fiber.Map{
            "target":  target,
            "scans":   target,
            "reports": reports,
        },
        Type:    "workspace",
        Message: "Detail workspace",
    })
}

func Scan(c *fiber.Ctx) error {
    scan := database.GetScans()
    return c.JSON(ResponseHTTP{
        Status:  200,
        Data:    scan,
        Type:    "scans",
        Total:   len(scan),
        Message: "List all the scan process",
    })
}

func Process(c *fiber.Ctx) error {
    processes := execution.GetOsmProcess("")
    return c.JSON(ResponseHTTP{
        Status:  200,
        Data:    processes,
        Type:    "processes",
        Total:   len(processes),
        Message: "List all osm process",
    })
}

func RawWorkspace(c *fiber.Ctx) error {
    //processes := execution.GetOsmProcess()
    return c.JSON(ResponseHTTP{
        Status: 200,
        Data: fiber.Map{
            "storages":   fmt.Sprintf("/%s/storages/", Opt.Server.StaticPrefix),
            "workspaces": fmt.Sprintf("/%s/workspaces/", Opt.Server.StaticPrefix),
            "logs":       fmt.Sprintf("/%s/logs/", Opt.Server.StaticPrefix),
        },
        Type:    "raw",
        Message: "Raw directory",
    })
}

func ListFlows(c *fiber.Ctx) error {
    flows := core.ListFlow(Opt)
    if len(flows) == 0 {
        //color.Red("Error to list workflows: %s", options.Env.WorkFlowsFolder)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Can't list workflow",
        })
    }

    var result []map[string]string

    for _, flow := range flows {
        if flow != "" {
            item := make(map[string]string)
            item["name"] = strings.TrimSuffix(filepath.Base(flow), ".yaml")

            // get modules
            Opt.Flow.Type = strings.TrimSuffix(item["name"], path.Ext(item["name"]))
            rawModules := core.ListModules(Opt)
            var modules []string
            for _, module := range rawModules {
                if module != "" {
                    modules = append(modules, strings.TrimSuffix(filepath.Base(module), ".yaml"))
                }
            }

            item["desc"] = ""
            parsedFlow, err := core.ParseFlow(flow)
            if err == nil {
                item["desc"] = parsedFlow.Desc
            }

            item["modules"] = strings.Join(modules, ",")
            result = append(result, item)

        }
    }

    return c.JSON(ResponseHTTP{
        Status:  200,
        Data:    result,
        Total:   len(flows),
        Type:    "flows",
        Message: "Workflows Listing",
    })
}

func HelperMessage(c *fiber.Ctx) error {
    message := fmt.Sprintf(`
[*] Visit this page for complete Usage: %s
`, libs.DOCS)

    return c.JSON(ResponseHTTP{
        Status: 200,
        Data: fiber.Map{
            "version": libs.VERSION,
            "doc":     libs.DOCS,
            "message": message,
        },
        Type:    "helper",
        Message: "Helper message",
    })
}

func DeleteWorkspace(c *fiber.Ctx) error {
    wsname := c.Params("wsname")

    wsDir := path.Join(Opt.Env.WorkspacesFolder, utils.NormalizePath(wsname))
    if !utils.FolderExists(wsDir) {
        wsDir = path.Join(Opt.Env.WorkspacesFolder, utils.StripPath(wsname))
        if !utils.FolderExists(wsDir) {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
                "error": "workspace didn't exist",
            })
        }
    }

    database.CleanWorkspace(wsname)
    os.RemoveAll(wsDir)
    return c.JSON(ResponseHTTP{
        Status: 200,

        Type:    "delete",
        Message: "Workspace Deleted",
    })
}
