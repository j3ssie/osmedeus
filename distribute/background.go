package distribute

import (
    "fmt"
    "github.com/j3ssie/osmedeus/database"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/provider"
    "github.com/j3ssie/osmedeus/utils"
    "strings"
)

func CheckingCloudInstance(opt libs.Options) {
    // select all cloud instance
    if opt.Cloud.Retry == 0 {
        opt.Cloud.Retry = 8
    }

    instances := database.GetRunningClouds()
    if len(instances) == 0 {
        utils.WarnF("no active cloud instance running")
        return
    }

    utils.InforF("Checking health of %v cloud instances", len(instances))
    for _, instance := range instances {
        providerObj, err := provider.InitProvider(instance.Provider, instance.Token)
        if err != nil {
            continue
        }

        cloud := CloudRunner{
            Opt:      opt,
            Provider: providerObj,
        }
        cloud.Prepare()
        cloud.PublicIP = instance.IPAddress
        cloud.InstanceID = instance.InstanceID

        cloud.HealthCheck()
    }
}

func (c *CloudRunner) HealthCheck() bool {
    if !c.IsRunning() || c.IsPanic() {
        utils.ErrorF("error detected at cloud instance: %v -- %v", c.Provider.ProviderName, c.PublicIP)

        c.DBErrorCloudScan()
        // instance should be deleted
        err := c.Provider.DeleteInstance(c.InstanceID)
        if err == nil {
            utils.InforF("Instance error detected at: %s", c.PublicIP)
        }

        // return error to the scan
        return false
    }

    return true
}

// IsRunning checking if cloud instance is running or not
func (c *CloudRunner) IsRunning() bool {
    utils.DebugF("Checking running process at: %v", c.PublicIP)
    cmd := fmt.Sprintf("%s utils ps --json", libs.BINARY)

    out, err := c.SSHExec(cmd)
    if err == nil && strings.Contains(out, "pid") {
        return true
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
