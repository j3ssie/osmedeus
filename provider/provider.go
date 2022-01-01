package provider

import (
    "fmt"
    "github.com/cenkalti/backoff/v4"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/spf13/cast"
    "strings"
    "time"
)

type Provider struct {
    ProviderName  string
    Token         string
    RedactedToken string

    Instances     []Instance
    InstanceLimit int
    Available     bool

    // for create snapshot
    SnapshotID    string
    SnapshotName  string
    OldSnapShotID []string
    SnapshotFound bool
    SSHKeyFound   bool
    SSHPublicKey  string
    SSHPrivateKey string
    SSHKeyID      string

    // for create
    CreatedInstance Instance
    Region          string
    Size            string
    SSHKeyName      string

    // mics
    SwapSizeMap map[string]int

    // for building
    ProviderConfig ConfigProvider
    Opt            libs.Options

    // for retry
    BackOff *backoff.ExponentialBackOff
    // client of vendor
    Client interface{}
}

type Instance struct {
    InstanceID   string
    InstanceName string
    IPAddress    string

    // meta data
    Region    string
    Size      string
    Status    string
    ImageID   string
    ImageName string
    CPU       string
    // MB
    Memory string
    // GB
    Disk string

    // more for the osm
    InputName    string
    ProviderName string
    CreatedAt    string
}

// InitProvider init provider object to easier interact with cloud provider
func InitProvider(providerName string, token string) (Provider, error) {
    var provider Provider
    provider.ProviderName = providerName
    provider.Token = token
    provider.InitClient()
    return provider, nil
}

// InitProviderWithConfig init provider object to easier interact with cloud provider
func InitProviderWithConfig(opt libs.Options, providerConfig ConfigProvider) (Provider, error) {
    var provider Provider
    provider.ProviderName = providerConfig.Provider
    provider.Token = providerConfig.Token
    provider.ProviderConfig = providerConfig
    provider.Opt = opt

    provider.InitClient()
    provider.Prepare()

    return provider, nil
}

func (p *Provider) InitClient() {
    if p.Token == "" {
        utils.ErrorF("empty or invalid token: %v", p.Token)
    }
    if len(p.Token) > 5 {
        p.RedactedToken = p.Token[:5] + "***" + p.Token[len(p.Token)-5:len(p.Token)]
        utils.InforF("Init %v provider with token: %v", color.HiYellowString(p.ProviderName), color.HiCyanString(p.RedactedToken))
    }

    p.Available = true
    switch p.ProviderName {
    case "do", "digitalocean":
        p.ClientDO()
        p.AccountDO()
    case "ln", "line", "linode":
        p.ClientLinode()
        p.AccountLN()
    default:
        p.ClientDO()
        p.AccountDO()
    }
}

// Prepare setup some default variables
func (p *Provider) Prepare() {
    // get snapshot
    version := strings.ReplaceAll(strings.TrimSpace(libs.VERSION), " ", "-")
    SnapshotName := fmt.Sprintf("%s-base-%s", strings.TrimSpace(libs.SNAPSHOT), version)
    p.SnapshotName = SnapshotName

    // sshKey
    keyName := fmt.Sprintf("%s-cloud-key", strings.TrimSpace(libs.SNAPSHOT))
    p.SSHKeyName = keyName

    // for retry
    b := backoff.NewExponentialBackOff()
    // It never stops if MaxElapsedTime == 0.
    b.MaxElapsedTime = 1200 * time.Second
    b.Multiplier = 2.0
    b.InitialInterval = 30 * time.Second
    p.BackOff = b

    if p.Opt.Cloud.Retry > 0 {
        b.MaxElapsedTime = time.Duration(p.Opt.Cloud.Retry*60) * time.Second
    }

    // setup ssh key
    if p.SSHPublicKey == "" {
        p.SSHPublicKey = p.Opt.Cloud.PublicKeyContent
    }
    if p.SSHPrivateKey == "" {
        p.SSHPrivateKey = p.Opt.Cloud.SecretKeyContent
    }

    utils.DebugF("Get data of cloud provider")
    switch p.ProviderName {
    case "do", "digitalocean":
        p.DefaultDO()
        p.Action(GetSSHKey)
        p.Action(ListImage)
    case "ln", "line", "linode":
        p.DefaultLinode()
        p.Action(GetSSHKey)
        p.Action(ListImage)
    default:
        p.DefaultDO()
        p.Action(GetSSHKey)
        p.Action(ListImage)
    }

    utils.InforF("Found SSH Key ID: %v", p.SSHKeyID)
}

func (p *Provider) Action(actionName string, params ...interface{}) error {
    var err error
    var param interface{}
    if len(params) > 0 {
        param = params[0]
    }
    utils.InforF("[%v] running action: %v", p.ProviderName, color.HiBlueString(actionName))
    operation := func() error {
        switch actionName {
        case GetSSHKey:
            err = p.GetSSHKey()
        case ListInstance:
            err = p.ListInstance()
        case ListImage:
            err = p.ListSnapShot()
        case RunBuild:
            err = p.RunBuild()
        case GetInstanceInfo:
            err = p.GetInstanceInfo(cast.ToInt(param))
        case BootInstance:
            err = p.BootInstance(cast.ToInt(param))
        case CreateInstance:
            err = p.CreateInstance(cast.ToString(param))
        default:
            err = p.ListInstance()
        }
        if err != nil {
            utils.ErrorF("error running action %v: %v", color.HiCyanString(actionName), err)
        }
        return err
    }
    err = backoff.Retry(operation, p.BackOff)
    if err != nil {
        utils.ErrorF("error running action %v -- %v", actionName, p.ProviderName)
        return err
    }
    return nil
}
