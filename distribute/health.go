package distribute

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/jinzhu/copier"
	jsoniter "github.com/json-iterator/go"
	"github.com/panjf2000/ants"
)

func GetAllInstances(opt libs.Options) (instances []CloudRunner) {
	utils.InforF("Getting all instances from %s", opt.Env.InstancesFolder)

	// Read all entries in the directory
	entries, err := ioutil.ReadDir(opt.Env.InstancesFolder)
	if err != nil {
		utils.ErrorF("error reading instances folder: %v", err)
		return instances
	}

	for _, entry := range entries {
		instanceFile := filepath.Join(opt.Env.InstancesFolder, entry.Name())

		if !utils.FileExists(instanceFile) {
			continue
		}

		instance := CloudRunner{}
		runtimeContent := utils.GetFileContent(instanceFile)
		err := jsoniter.UnmarshalFromString(runtimeContent, &instance)
		if err != nil {
			utils.ErrorF("error marshal data: %v", err)
		}
		instances = append(instances, instance)
	}

	return instances
}

func CheckingCloudInstance(opt libs.Options) {
	if opt.Cloud.Retry == 0 {
		opt.Cloud.Retry = 8
	}

	// select all cloud instance
	instances := GetAllInstances(opt)
	if len(instances) == 0 {
		utils.WarnF("no active cloud instance running")
		return
	}

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(opt.Concurrency*5, func(i interface{}) {
		// really start to scan
		instance := i.(CloudRunner)

		var options libs.Options
		copier.Copy(&options, &opt)
		instance.Opt = opt
		instance.Provider.IsBackgroundCheck = true
		instance.Provider.InitClient()
		instance.Prepare()
		instance.HealthCheck()
		wg.Done()
	}, ants.WithPreAlloc(true))
	defer p.Release()

	utils.InforF("Evaluating the status of %v cloud instances", color.HiMagentaString("%v", len(instances)))
	for _, instance := range instances {
		wg.Add(1)
		_ = p.Invoke(instance)
	}
	wg.Wait()

}

func (c *CloudRunner) HealthCheck() bool {
	utils.InforF("[%v] Checking the instance: %v -- %v", color.HiYellowString(c.Provider.ProviderName), color.HiCyanString(c.PublicIP), color.HiCyanString(path.Base(c.InstanceName)))
	c.Opt.Cloud.Input = c.Input

	counter := 0
	for i := 0; i < c.Opt.Cloud.Retry; i++ {
		// sync in result first
		c.SyncResult()
		if !c.IsRunning() || c.IsPanic() {
			counter += 1
			utils.DebugF("retry[%v]: An error has been detected at the cloud instance: %v -- %v", counter, c.Provider.ProviderName, c.PublicIP)
		}

		if counter == c.Opt.Cloud.Retry-1 {
			utils.ErrorF("An error has been detected at the cloud instance: %v -- %v", c.Provider.ProviderName, c.PublicIP)
			err := c.Provider.DeleteInstance(c.InstanceID)
			if err == nil {
				utils.InforF("Instance deleted %s -- %s", color.HiYellowString(c.PublicIP), color.HiYellowString(c.InstanceID))
				instanceFile := c.Opt.Env.InstancesFolder + "/" + c.InstanceName + "-" + c.PublicIP + ".json"
				os.Remove(instanceFile)
			}
			return false
		}
		time.Sleep(10 * time.Second)
	}

	utils.InforF("[%v] Instance is still running well: %v -- %v", color.HiYellowString(c.Provider.ProviderName), color.HiCyanString(c.PublicIP), color.HiCyanString(path.Base(c.InstanceName)))
	return true
}

// IsRunning checking if cloud instance is running or not
func (c *CloudRunner) IsRunning() bool {
	utils.DebugF("Checking running process at: %v", c.PublicIP)
	cmd := fmt.Sprintf("%s utils ps --json", libs.BINARY)

	// ignore checking process if you're running custom command '--no-ps'
	if c.Opt.Cloud.IgnoreProcess {
		return true
	}

	out, err := c.SSHExec(cmd)
	if err == nil && strings.Contains(out, "pid") {
		return true
	}

	// retry checking process
	for i := 0; i < c.Opt.Cloud.Retry; i += 2 {
		out, err := c.SSHExec(cmd)
		if err == nil && strings.Contains(out, "pid") {
			return true
		}
	}

	utils.DebugF(out)
	utils.ErrorF("no process running at %v", c.PublicIP)
	return false
}

// IsPanic checking if cloud instance has any panic error
func (c *CloudRunner) IsPanic() bool {
	utils.DebugF("Checking panic error at: %v", c.PublicIP)
	cmd := fmt.Sprintf("%s utils tmux logs -A -l 30", libs.BINARY)
	out, err := c.SSHExec(cmd)

	if err == nil {
		if strings.Contains(out, "out of memory") || strings.Contains(out, "runtime.(*") || strings.Contains(out, "[panic]") {
			utils.DebugF(out)
			utils.ErrorF("Fatal panic detected at: %s", c.PublicIP)
			return true
		}
	}

	return false
}
