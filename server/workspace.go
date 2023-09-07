package server

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/database"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/thoas/go-funk"
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
// @Router /v1/workspaces [get]
func ListWorkspaces(c *fiber.Ctx) error {
	workspaces := database.GetAllScan(Opt)
	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    workspaces,
		Type:    "workspaces",
		Total:   len(workspaces),
		Message: "List all of Workspaces",
	})
}

func WorkspaceDetail(c *fiber.Ctx) error {
	wsname := c.Params("wsname")
	workspace := database.GetSingleScan(wsname, Opt)

	// make a reports map
	rawReports := workspace.Target.Reports
	reports := make(map[string][]string)

	for _, report := range rawReports {
		// replace the home folder first
		if strings.Contains(report.ReportPath, workspace.Target.Workspace) {
			// /root/workspaces-osmedeus/
			homeFolder := strings.Split(report.ReportPath, workspace.Target.Workspace)[0]
			report.ReportPath = strings.ReplaceAll(report.ReportPath, homeFolder, Opt.Env.WorkspacesFolder+"/")
		}

		if utils.FileLength(report.ReportPath) == 0 {
			if utils.FolderExists(report.ReportPath) && utils.FolderLength(report.ReportPath) == 0 {
				continue
			} else {
				continue
			}
		}

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
			"workspace": workspace,
			"reports":   reports,
		},
		Type:    "workspace",
		Message: "Workspace Detail ",
	})
}

func ListAllScan(c *fiber.Ctx) error {
	scan := database.GetScanProgress(Opt)
	return c.JSON(ResponseHTTP{
		Status:  200,
		Data:    scan,
		Type:    "scans",
		Total:   len(scan),
		Message: "List all the scan process",
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

	os.RemoveAll(wsDir)
	return c.JSON(ResponseHTTP{
		Status:  200,
		Type:    "delete",
		Message: "Workspace Deleted",
	})
}
