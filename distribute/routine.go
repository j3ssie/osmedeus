package distribute

import (
	"fmt"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
)

func (c *CloudRunner) Scan(target string) error {
	err := c.CreateInstance(target)
	if err != nil {
		utils.ErrorF("Error to create instance")
		return err
	}

	if c.Opt.Cloud.OnlyCreateDroplet {
		return nil
	}

	// parse some parameters first
	c.PrepareInput()

	// pre run before starting the scan
	c.PreRunLocal()

	if c.Opt.Cloud.EnableSyncWorkflow {
		err = c.CopyWorkflow()
		if err != nil {
			utils.ErrorF("Error to copy workflow to instance")
			return err
		}
	}

	// copy target to droplet first
	err = c.CopyTarget()
	if err != nil {
		utils.ErrorF("Error to copy input to instance")
		return err
	}

	// pre run commands
	c.PreRunRemote()

	err = c.StartScan()
	if err != nil {
		utils.ErrorF("Error to run command on instance")
		return err
	}

	// check if done file created in instance or not
	c.CheckingDone()
	c.DBDoneCloudScan()

	if !c.Opt.Cloud.DisableLocalSync {
		err = c.SyncResult()
		if err != nil {
			utils.ErrorF("Error to sync result to instance")
			return err
		}
	}

	// post run after scan done
	c.PostRunLocal()

	if c.Opt.Cloud.NoDelete {
		utils.DebugF("Disable copy result and Delete instance when run command in background")
		return nil
	}

	err = c.Provider.DeleteInstance(c.InstanceID)
	if err != nil {
		utils.ErrorF("Error to delete instance")
		return err
	}

	if c.Opt.Cloud.CopyWorkspaceToGit {
		utils.InforF("Coping workspace to git storages")
		baseCmd := fmt.Sprintf("%s scan --nn -f sync -t {{.Workspace}}", libs.BINARY)
		cmd := core.ResolveData(baseCmd, c.Target)
		utils.RunCommandWithErr(cmd)
	}

	return nil
}
