package distribute

import (
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/database"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/provider"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/panjf2000/ants"
    "os"
    "strings"
    "sync"
)

type CloudRunner struct {
    ProviderName string
    Opt          libs.Options
    Provider     provider.Provider

    // for storing in local DB
    TaskName      string
    TaskType      string
    Input         string
    CloudInstance database.CloudInstance

    // core entry point
    PublicIP      string
    DestInstance  string
    SshPublicKey  string
    SshPrivateKey string

    InstanceID   string
    InstanceName string

    Available bool
    Target    map[string]string
    Runner    core.Runner
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

                //// got the valid cloud with token here
                //if selectedCloud.ProviderName != "linode" {
                //	if !selectedCloud.Usage() {
                //		utils.ErrorF("Current cloud reach limit usage: %v", selectedCloud.Cloud.Name)
                //		timeout := utils.CalcTimeout(options.Cloud.CloudWait)
                //		time.Sleep(time.Duration(timeout) * time.Second)
                //		continue
                //	}
                //}

                target = PrepareTarget(target, options)
                utils.BlockF("start-scan", utils.CleanPath(target))

                // really start to run scan here
                err := selectedCloud.Scan(target)
                if err != nil {
                    utils.ErrorF("error start scan %s", target)
                    selectedCloud.Provider.DeleteInstance(selectedCloud.InstanceID)
                    continue
                }

                utils.BlockF("done-scan", utils.CleanPath(target))
                utils.InforF("--------------------------------")
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

    // return the cloud back
}

// GetClouds prepare clouds object in config file
func GetClouds(options libs.Options) []CloudRunner {
    var cloudInfos []CloudRunner

    // ~/osmedeus-plugins/cloud/provider.yaml
    //cloudConfigFile := path.Join(options.Env.CloudConfigFolder, "provider.yaml")

    // parse config from cloud/config.yaml file
    providerConfigs, err := provider.ParseProvider(options.CloudConfigFile)
    utils.DebugF("Parsing cloud config from: %s", options.CloudConfigFile)
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

    utils.InforF("Number of cloud provider prepared in queue: %v", len(cloudInfos))
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

    // make sure the permission of private key is right
    os.Chmod(utils.NormalizePath(c.SshPrivateKey), 0600)

    // parse blank target to get env
    c.Target = core.ParseInput("example.com", c.Opt)
}
