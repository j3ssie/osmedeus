package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/utils"
)

// TaskData data required in json form
// TaskData represents the scan configuration and parameters
// @Description Configuration for executing a new scan
type TaskData struct {
	MasterPassword string `json:"password" example:"secret123" description:"Master password for authentication"`
	Command        string `json:"command" description:"Override command to execute (optional)"`

	Target      string   `json:"target" example:"example.com" description:"Target to scan"`
	TargetsList []string `json:"targets" description:"List of targets to scan (optional)"`
	WorkFlow    string   `json:"workflow" example:"general" description:"Workflow name to execute"`

	TargetAsFile     bool     `json:"as_file" example:"false" description:"Treat target as file (optional)"`
	TargetsFile      string   `json:"targets_file" example:"/path/to/targets.txt" description:"File containing targets (optional)"`
	UploadTargetFile bool     `json:"upload_targets_file" example:"false" description:"Upload target data (optional)"`
	Params           []string `json:"params" example:"[\"-deep\", \"-aggressive\"]" description:"Additional parameters (optional)"`

	ModuleName  string `json:"module" example:"subdomain" description:"Plugin name to run (optional)"`
	Workspace   string `json:"workspace" example:"my-project" description:"Workspace name (optional)"`
	Threads     int    `json:"threads" example:"10" description:"Number of concurrent threads (optional)"`
	Chunk       bool   `json:"chunk" example:"false" description:"Enable chunk mode (optional)"`
	Timeout     string `json:"timeout" example:"1h" description:"Scan timeout (optional)"`
	Concurrency int    `json:"concurrency" example:"5" description:"Concurrency level (optional)"`
	Distributed bool   `json:"distributed" example:"false" description:"Enable distributed scanning (optional)"`
	WildCard    bool   `json:"wildcard" example:"false" description:"Enable wildcard mode (optional)"`
	Debug       bool   `json:"debug" example:"false" description:"Enable debug mode (optional)"`
	Test        bool   `json:"test" example:"false" description:"Test mode without actual execution (optional)"`
}

// @Summary Start a new scan
// @Description Execute a new scan with specified configuration
// @Tags scans
// @Accept json
// @Produce json
// @Param taskData body TaskData true "Scan configuration"
// @Success 200 {object} ResponseHTTP{data=object{command=string,input=string}} "Scan started successfully"
// @Failure 400 {object} object{error=string} "Invalid JSON payload"
// @Example Request-SingleTarget
//
//	{
//	   "target": "example.com",
//	   "workflow": "general"
//	}
//
// @Example Request-MultipleTargets
//
//	{
//	   "targets": ["1.2.3.4/24", "5.6.7.8/24"],
//	   "as_file": true,
//	   "workflow": "cidr"
//	}
//
// @Router /api/osmp/execute [post]
func NewScan(c *fiber.Ctx) error {
	var taskData TaskData
	var invalid bool

	err := c.BodyParser(&taskData)
	if err != nil {
		invalid = true
	}

	if invalid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	taskData.Command = CommandBuilder(&taskData)
	// get workspace if we didn't have one
	if taskData.Workspace == "" {
		taskData.Workspace = utils.CleanPath(taskData.Target)
	}
	utils.InforF("Running command: %v", taskData.Command)

	if !taskData.Test {
		go func() {
			utils.RunOSCommand(taskData.Command)
			return
		}()
	}

	return c.JSON(ResponseHTTP{
		Status: 200,
		Data: fiber.Map{
			"command":   taskData.Command,
			"input":     taskData.Target,
			"workspace": taskData.Workspace, // this would be the folder name
		},
		Type:    "new-scan",
		Message: "New Scan Imported",
	})
}

// NewScanCloud new scan
func NewScanCloud(c *fiber.Ctx) error {
	var taskData TaskData
	var invalid bool

	err := c.BodyParser(&taskData)
	if err != nil {
		invalid = true
	}

	if invalid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	taskData.Command = CommandBuilder(&taskData)
	// get workspace if we didn't have one
	if taskData.Workspace == "" {
		taskData.Workspace = utils.CleanPath(taskData.Target)
	}
	utils.InforF("Running command: %v", taskData.Command)

	if !taskData.Test {
		go func() {
			utils.RunOSCommand(taskData.Command)
			return
		}()
	}

	return c.JSON(ResponseHTTP{
		Status: 200,
		Data: fiber.Map{
			"input":    taskData.Target,
			"workflow": taskData.WorkFlow,
		},
		Type:    "new-scan",
		Message: "New Scan Imported",
	})
}
