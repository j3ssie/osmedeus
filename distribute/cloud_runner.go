package distribute

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/core"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/provider"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/panjf2000/ants"
)

type CloudRunner struct {
	Opt      libs.Options `json:"-"`
	Provider provider.Provider

	// for storing in local DB
	TaskName     string `json:"task_name"`
	TaskType     string `json:"task_type"`
	Input        string `json:"input"`
	RawCommand   string `json:"raw_command"`
	InstanceFile string `json:"instance_file"`

	// core entry point
	PublicIP      string `json:"public_ip"`
	DestInstance  string `json:"dest_instance"`
	SshPublicKey  string `json:"ssh_public_key"`
	SshPrivateKey string
	SSHUser       string
	BasePath      string

	InstanceID   string `json:"instance_id"`
	InstanceName string `json:"instance_name"`
	IsError      bool   `json:"is_error"`

	Target map[string]string `json:"-"`
	Runner core.Runner       `json:"-"`
}

// InitCloud init cloud runner obj
func InitCloud(options libs.Options, targets []string) {
	// init clouds object queue
	cloudConfigs := GetClouds(options)
	if options.Cloud.CheckingLimit {
		return
	}

	// really start doing something
	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(options.Concurrency, func(i interface{}) {
		target := i.(string)
		if strings.TrimSpace(target) == "" {
			wg.Done()
			return
		}

		for {
			var selectedCloud CloudRunner
			var isDone bool
			utils.DebugF("Select available accounts")
			for _, cloudInfo := range cloudConfigs {
				// don't run it again for next cloud account
				if isDone {
					break
				}
				if cloudInfo.Provider.Available {
					selectedCloud = cloudInfo
				}

				target = PrepareTarget(target, options)
				if !options.Cloud.OnlyCreateDroplet {
					utils.InforF("Initiating the scanning process of %v", color.HiMagentaString(utils.CleanPath(target)))
				}

				// really start to run scan here
				err := selectedCloud.Scan(target)
				if err != nil {
					utils.ErrorF("error start scan %s", color.HiCyanString(target))
					if options.Cloud.NoDelete {
						continue
					}
					if ok := selectedCloud.Provider.DeleteInstance(selectedCloud.InstanceID); ok == nil {
						selectedCloud.DeleteInstanceConfig()
					}
					continue
				}

				if !options.Cloud.OnlyCreateDroplet {
					utils.InforF("Completed scanning %v", color.HiCyanString(utils.CleanPath(target)))
					utils.InforF("--------------------------------")
				}
				isDone = true
			}
			break
		}
		wg.Done()

	}, ants.WithPreAlloc(true))
	defer p.Release()

	for _, target := range targets {
		wg.Add(1)
		_ = p.Invoke(target)
	}
	wg.Wait()

}

// GetClouds prepare clouds object in config file
func GetClouds(options libs.Options) []CloudRunner {
	var cloudInfos []CloudRunner

	// parse config from cloud/provider.yaml file
	providerConfigs, err := provider.ParseProvider(options.CloudConfigFile)
	if err != nil {
		return cloudInfos
	}

	options.Cloud.BuildRepo = providerConfigs.Builder.BuildRepo
	options.Cloud.SecretKey = utils.NormalizePath(providerConfigs.Builder.SecretKey)
	options.Cloud.PublicKey = utils.NormalizePath(providerConfigs.Builder.PublicKey)

	// we only get provider info from config file but replace token with --token
	if options.Cloud.IgnoreConfigFile || options.Cloud.Token != "" {
		var providerConfig provider.ConfigProvider
		if len(providerConfigs.Clouds) > 0 {
			utils.InforF("Ignore config file from: %v", options.CloudConfigFile)
			providerConfig = providerConfigs.Clouds[0]
		}
		providerConfig.Token = options.Cloud.Token
		cloudInfo := SetupProvider(options, providerConfig)
		cloudInfos = append(cloudInfos, cloudInfo)
		return cloudInfos
	}

	for _, providerConfig := range providerConfigs.Clouds {
		cloudInfo := SetupProvider(options, providerConfig)
		if len(cloudInfo.Provider.Token) > 6 {
			cloudInfo.Provider.RedactedToken = cloudInfo.Provider.Token[:5] + "***" + cloudInfo.Provider.Token[len(cloudInfo.Provider.Token)-5:]
		}
		cloudInfos = append(cloudInfos, cloudInfo)
	}

	utils.InforF("Number of cloud providers ready in queue: %v", color.HiMagentaString("%v", len(cloudInfos)))
	return cloudInfos
}

// SetupProvider setup new provider
func SetupProvider(opt libs.Options, providerConfig provider.ConfigProvider) CloudRunner {
	var cloudRunner CloudRunner
	cloudRunner.Opt = opt
	cloudRunner.Prepare()

	providerCloud, err := provider.InitProviderWithConfig(opt, providerConfig)
	if err != nil {
		return cloudRunner
	}
	cloudRunner.Provider = providerCloud

	cloudRunner.SSHUser = cloudRunner.Provider.SSHUser
	if cloudRunner.SSHUser == "" {
		cloudRunner.SSHUser = "root"
	}

	if cloudRunner.SSHUser != "root" {
		cloudRunner.BasePath = fmt.Sprintf("/home/%s", cloudRunner.SSHUser)
	}

	if opt.Cloud.IgnoreSetup {
		return cloudRunner
	}

	// check if snapshot is okay or not
	if !cloudRunner.Provider.SnapshotFound || opt.Cloud.ReBuildBaseImage {
		err = cloudRunner.Provider.BuildImage()
		if err != nil {
			utils.ErrorF("error build snapshot at %v", cloudRunner.Provider.ProviderConfig.BuildFile)
			return cloudRunner
		}
	}

	return cloudRunner
}

// Prepare some variables
func (c *CloudRunner) Prepare() {
	c.SshPrivateKey = c.Opt.Cloud.SecretKey
	c.SshPublicKey = c.Opt.Cloud.PublicKey
	c.SSHUser = "root"
	c.BasePath = "/root"

	// make sure the permission of private key is right
	os.Chmod(utils.NormalizePath(c.SshPrivateKey), 0600)

	// parse blank target to get env
	c.Target = core.ParseInput("example.com", c.Opt)
}
