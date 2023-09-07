package server

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"os"
	"path"
	"strings"
)

// UploadData data required in json form
type UploadData struct {
	Data     string `json:"data"`
	Filename string `json:"filename"`
}

// Upload testing authenticated connection
func Upload(c *fiber.Ctx) error {
	var uploadData UploadData

	err := c.BodyParser(&uploadData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	if uploadData.Filename == "" {
		uploadData.Filename = utils.RandomString(6)
	}
	tmpFile := path.Base(utils.NormalizePath(uploadData.Filename))
	baseDir := fmt.Sprintf("/tmp/%v-input/", libs.BINARY)
	if !utils.FolderExists(baseDir) {
		os.MkdirAll(baseDir, 0755)
	}

	filename := path.Join(baseDir, tmpFile)
	data := uploadData.Data

	utils.WriteToFile(filename, data)

	return c.JSON(ResponseHTTP{
		Status: 200,
		Data: fiber.Map{
			"filepath": filename,
		},
		Type:    "upload",
		Message: "New Data Uploaded",
	})

}

// SaveTargets save upload data to /tmp/osm-input/data-osm-xxx.txt
func SaveTargets(targets []string) string {
	tmpFile := fmt.Sprintf("data-%v-%v-%v.txt", utils.GetTS(), libs.BINARY, utils.RandomString(6))
	baseDir := fmt.Sprintf("/tmp/%v-input/", libs.BINARY)
	if !utils.FolderExists(baseDir) {
		os.MkdirAll(baseDir, 0755)
	}
	targetFile := path.Join(baseDir, tmpFile)
	data := strings.Join(targets, "\n")
	filename, err := utils.WriteToFile(targetFile, data)
	if err != nil {
		return ""
	}
	return filename
}

// CommandBuilder build core command from API
func CommandBuilder(taskData *TaskData) string {
	binary := fmt.Sprintf("%s scan", libs.BINARY)
	if taskData.Distributed {
		binary = fmt.Sprintf("%s cloud", libs.BINARY)
		if taskData.Chunk {
			binary = fmt.Sprintf("%s cloud --chunk", libs.BINARY)
		}
	}

	if taskData.ScanID != "" {
		binary += fmt.Sprintf(" --sid %v ", taskData.ScanID)
	}

	var command string
	var workspace, concurrency, timeout, params, workflow, plugin, scanID string

	if len(taskData.TargetsList) > 0 {
		taskData.TargetsFile = SaveTargets(taskData.TargetsList)
		utils.DebugF("Save targets list to: %v", taskData.TargetsFile)
	}

	// get workspace
	if taskData.Workspace != "" {
		taskData.Workspace = utils.CleanPath(taskData.Workspace)
		workspace = fmt.Sprintf(" -w '%v'", taskData.Workspace)
	}

	if taskData.Binary != "" {
		binary = taskData.Binary
	}

	//if taskData.RawName {
	//	binary = binary + " --rt "
	//}

	if taskData.ViewOnly {
		binary = binary + " --view-only "
	}

	// default workflow is general
	if taskData.WorkFlow != "" {
		workflow = fmt.Sprintf(" -f '%v'", taskData.WorkFlow)
	}

	// some mics options
	if taskData.Timeout != "" {
		timeout = fmt.Sprintf(" --timeout '%v'", taskData.Timeout)
	}
	if taskData.Concurrency > 0 {
		concurrency = fmt.Sprintf(" -c %v", taskData.Concurrency)
	}
	if len(taskData.Params) > 0 {
		for _, param := range taskData.Params {
			// @NOTE replace ',' from request to ';;' first because corbra auto split ','
			if strings.Contains(param, ",") {
				param = strings.Replace(param, ",", ";;", -1)
			}

			if strings.HasPrefix(param, "'") && strings.HasSuffix(param, "'") {
				params += fmt.Sprintf(" -p %s", param)
				continue
			}
			params += fmt.Sprintf(" -p '%s'", param)
		}
	}

	if taskData.PluginName != "" {
		plugin = fmt.Sprintf(" -m '%v'", taskData.PluginName)
	}

	// override everything
	if taskData.Command != "" {
		return taskData.Command
	}

	// mean general scan
	if taskData.PluginName == "" {
		command = fmt.Sprintf("%v %v -t %v %v%v%v%v", binary, workflow, taskData.Target, concurrency, timeout, workspace, params)
		if taskData.TargetsFile != "" {
			command = fmt.Sprintf("%v %v -T %v %v%v%v%v", binary, workflow, taskData.TargetsFile, concurrency, timeout, workspace, params)
		}
		command = strings.TrimSpace(command)
		if taskData.Debug {
			command = command + " --debug"
		}
		return command
	}

	command = fmt.Sprintf("%v %v -t %v %v %v%v%v%v", binary, plugin, taskData.Target, scanID, concurrency, timeout, workspace, params)
	if taskData.TargetsFile != "" {
		command = fmt.Sprintf("%v %v -t %v %v %v%v%v%v", binary, plugin, taskData.TargetsFile, scanID, concurrency, timeout, workspace, params)
	}

	command = strings.TrimSpace(command)
	if taskData.Debug {
		command = command + " --debug"
	}
	return command
}
