package distribute

import (
	"path"
	"time"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/database"
	"github.com/j3ssie/osmedeus/utils"
	jsoniter "github.com/json-iterator/go"
)

func (c *CloudRunner) DBNewTarget() {
	targetFolder := path.Join(c.Opt.Env.WorkspacesFolder, c.Target["Workspace"])
	utils.MakeDir(targetFolder)

	c.Runner.ScanObj = database.Scan{
		IsCloud:   true,
		InputName: c.Target["Workspace"],
		TaskType:  "distributed",
	}

	c.Runner.ScanObj.CreatedAt = time.Now()
	c.DBRuntimeUpdate()
}

func (c *CloudRunner) DBRuntimeUpdate() {
	runtimeFile := path.Join(c.Opt.Env.WorkspacesFolder, c.Target["Workspace"], "runtime")
	c.Runner.ScanObj.UpdatedAt = time.Now()

	if runtimeData, err := jsoniter.MarshalToString(c.Runner.ScanObj); err == nil {
		utils.InforF("Updating runtime file: %s", color.HiCyanString(runtimeFile))
		utils.WriteToFile(runtimeFile, runtimeData)
	}
}
