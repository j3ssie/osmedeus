package server

import (
    "github.com/gofiber/fiber/v2"
    "github.com/j3ssie/osmedeus/database"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/spf13/cast"
)

// TaskData data required in json form
type TaskData struct {
    MasterPassword string `json:"password"`
    Binary         string `json:"binary"`
    // override everything below
    Command string `json:"command"`

    // these two not be blank when run with plugins
    WorkFlow   string `json:"workflow"`
    PluginName string `json:"plugin"`
    ScanID     string `json:"scan_id"`

    // for select scan + task
    Workspace   string   `json:"workspace"`
    Target      string   `json:"target"`
    TargetsList []string `json:"targets"`
    TargetsFile string   `json:"targets_file"`

    AliveAssets bool `json:"alive_assets"` // skip targets part and select the assets from DB
    AllAssets   bool `json:"all_assets"`   // skip targets part and select the assets from DB

    // just more mics info for custom command later
    Params      []string `json:"params"`
    Timeout     string   `json:"timeout"`
    Concurrency int      `json:"concurrency"`

    // enable distributed scan
    Distributed bool `json:"distributed"`

    // for chunk mode only
    Threads      int  `json:"threads"`
    Chunk        bool `json:"chunk"`
    TargetAsFile bool `json:"as_file"`

    // only select record not run the command
    RawName  bool `json:"RawName"`
    WildCard bool `json:"wildcard"`
    ViewOnly bool `json:"view_only"`
    Debug    bool `json:"debug"`
    Test     bool `json:"test"`
}

// NewScan new scan
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

    CreatePreparedScan(&taskData)
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
            "command": taskData.Command,
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

    CreatePreparedScan(&taskData)
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
            "scan_id":  taskData.ScanID,
            "workflow": taskData.WorkFlow,
        },
        Type:    "new-scan",
        Message: "New Scan Imported",
    })
}

func CreatePreparedScan(taskData *TaskData) {
    scanObj := database.Scan{
        InputName:  taskData.Target,
        TaskName:   taskData.WorkFlow,
        IsRunning:  false,
        IsPrepared: false,
    }
    database.DBNewScan(&scanObj)
    taskData.ScanID = cast.ToString(scanObj.ID)
}
