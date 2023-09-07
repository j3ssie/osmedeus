package distribute

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/jinzhu/copier"
	jsoniter "github.com/json-iterator/go"
)

func (c *CloudRunner) CreateInstance(target string) error {
	c.Opt.Cloud.Input = target
	c.Target = core.ParseInput(target, c.Opt)

	if c.Opt.EnableFormatInput {
		c.Opt.Cloud.Input = c.Target["Target"]
	}
	if c.Opt.Cloud.Workspace == "" {
		c.Opt.Cloud.Workspace = utils.CleanPath(c.Target["Target"])
	}

	if c.Opt.Cloud.Workspace != "" {
		c.Target["Workspace"] = c.Opt.Cloud.Workspace
		// disable changing workspace name in huntersuite to keep track the scanID
		if c.Opt.Cloud.EnableChunk {
			if strings.Contains(target, "-chunk-") {
				index := strings.Split(target, "-chunk-")[1]
				c.Target["WorkspaceChunk"] = path.Base(c.Opt.Cloud.Workspace) + "-" + index
				utils.DebugF("Changing workspace name: %v", c.Target["WorkspaceChunk"])
			}
		}
		utils.DebugF(`c.Target["Workspace"] -- %v`, c.Target["Workspace"])
	}

	// run-flow-example.com
	InstancePrefix := fmt.Sprintf("run")
	if c.Opt.Cloud.Flow != "" {
		InstancePrefix = fmt.Sprintf("run-%s", utils.CleanPath(c.Opt.Cloud.Flow))
	}
	if c.Opt.Cloud.Module != "" {
		InstancePrefix = fmt.Sprintf("run-%s", utils.CleanPath(c.Opt.Cloud.Module))
	}

	// run-flow-example.com-1
	c.InstanceName = fmt.Sprintf("%s-%s", InstancePrefix, strings.TrimSpace(c.Target["Workspace"]))
	if c.Opt.Cloud.EnableChunk {
		c.InstanceName = fmt.Sprintf("%s-%s", InstancePrefix, strings.TrimSpace(c.Target["WorkspaceChunk"]))
	}

	// make sure the droplet name is unique
	c.InstanceName = c.InstanceName + "-" + utils.RandomString(4)
	// clean up the instance name first
	if strings.Contains(c.InstanceName, "_") {
		c.InstanceName = strings.ReplaceAll(c.InstanceName, "_", "-")
	}

	// force changing instance name from cli
	if c.Opt.Cloud.InstanceName != "" {
		c.InstanceName = c.Opt.Cloud.InstanceName
	}

	// quick check for instance name
	switch c.Provider.ProviderName {
	case "ln", "line", "linode":
		if len(c.InstanceName) > 32 {
			c.InstanceName = strings.Trim(strings.Trim(strings.Trim(c.InstanceName[:20], "-"), "."), "_") + "-" + utils.RandomString(9)
		}
	}

	/* Really start to run command to create instance here */
	err := c.Provider.CreateInstance(c.InstanceName)

	// fix the naming in linode might need another check for length
	if err != nil {
		// check if account reach limit first
		if strings.Contains(err.Error(), "Account Limit reached") {
			utils.ErrorF("Account %v reach limit instance", c.Provider.RedactedToken)
			return fmt.Errorf("error creating instance")
		}

		if strings.Contains(err.Error(), "valid hostname characters are allowed") || strings.Contains(err.Error(), "[400] [label]") {
			c.InstanceName = fmt.Sprintf("runr-%s-%s", utils.GetTS(), utils.RandomString(8))
		}
		err = c.Provider.CreateInstance(c.InstanceName)
		if err != nil {
			return fmt.Errorf("error creating instance")
		}
	}

	if err != nil {
		return fmt.Errorf("error creating instance: %v", c.InstanceName)
	}

	c.PublicIP = c.Provider.CreatedInstance.IPAddress

	c.InstanceID = c.Provider.CreatedInstance.InstanceID
	c.DestInstance = fmt.Sprintf("%s@%s", c.SSHUser, c.PublicIP)
	utils.InforF("Instance created: %s -- %s -- %v", color.HiCyanString(c.InstanceName), color.HiCyanString(c.InstanceID), color.HiCyanString(c.PublicIP))
	c.Target["CIP"] = c.PublicIP
	c.Target["RemoteIP"] = c.PublicIP

	// create the JSON file for the instance
	c.WriteInstanceConfig()
	return nil

}

func (c *CloudRunner) DeleteInstanceConfig() {
	utils.DebugF("Deleting instance config %v", color.HiCyanString(c.InstanceFile))
	os.Remove(c.InstanceFile)
}

func (c *CloudRunner) WriteInstanceConfig() error {
	instanceFile := c.Opt.Env.InstancesFolder + "/" + c.InstanceName + "-" + c.PublicIP + ".json"
	c.InstanceFile = instanceFile
	utils.DebugF("Writing instance config to %v", color.HiCyanString(instanceFile))

	var instanceObj CloudRunner
	copier.Copy(&instanceObj, &c)
	instanceObj.Runner = core.Runner{}
	instanceObj.Opt = libs.Options{}

	if data, err := jsoniter.MarshalToString(instanceObj); err == nil {
		utils.WriteToFile(instanceFile, data)
	}

	return nil
}

func (c *CloudRunner) ErrorWriteInstanceConfig() error {
	instanceFile := c.Opt.Env.InstancesFolder + "/" + c.InstanceName + "-" + c.PublicIP + ".json"
	c.InstanceFile = instanceFile
	utils.DebugF("Writing instance config to %v", color.HiCyanString(instanceFile))

	var instanceObj CloudRunner
	copier.Copy(&instanceObj, &c)

	instanceObj.IsError = true
	instanceObj.Runner = core.Runner{}
	instanceObj.Opt = libs.Options{}

	if data, err := jsoniter.MarshalToString(instanceObj); err == nil {
		utils.WriteToFile(instanceFile, data)
	}

	return nil
}
