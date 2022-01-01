package distribute

import (
    "fmt"
    "github.com/j3ssie/osmedeus/database"
    "path"
)

func (c *CloudRunner) DBNewTarget() {
    if c.Opt.NoDB {
        return
    }

    c.Runner.TargetObj = database.Target{
        InputName: c.Target["Input"],
        Workspace: c.Target["Workspace"],
    }
    database.DBUpdateTarget(&c.Runner.TargetObj)
}

func (c *CloudRunner) DBNewScanLocal() {
    if c.Opt.NoDB {
        return
    }

    c.Runner.ScanObj = database.Scan{
        TaskType: fmt.Sprintf("cloud-%s", c.TaskType),
        TaskName: path.Base(c.TaskName),
        //UID:      "cloud-runner",

        TotalSteps: 0,
        InputName:  c.Input,
        IsCloud:    true,
        LogFile:    c.Opt.LogFile,
        Target:     c.Runner.TargetObj,
    }
    database.DBNewScan(&c.Runner.ScanObj)
}

func (c *CloudRunner) DBNewCloudInstance() {
    if c.Opt.NoDB {
        return
    }

    c.CloudInstance = database.CloudInstance{
        Token:      c.Provider.Token,
        Provider:   c.Provider.ProviderName,
        SnapShotID: c.Provider.SnapshotID,
        InstanceID: c.InstanceID,
        IPAddress:  c.PublicIP,
        InputName:  c.Input,
        Status:     "running",
        Target:     c.Runner.TargetObj,
    }

    if c.Opt.Cloud.EnableChunk {
        c.CloudInstance.IsChunk = true
    }
    database.DBUpdateCloudInstance(&c.CloudInstance)
}

func (c *CloudRunner) DBErrorCloudScan() {
    c.CloudInstance.Status = "error"
    c.CloudInstance.IsError = true
    database.DBUpdateCloudInstance(&c.CloudInstance)
}

func (c *CloudRunner) DBDoneCloudScan() {
    if c.Opt.NoDB {
        return
    }

    c.CloudInstance.Status = "done"
    database.DBUpdateCloudInstance(&c.CloudInstance)
}
