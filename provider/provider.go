package provider

import (
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/cast"
)

type Provider struct {
	ProviderName  string
	Token         string
	RedactedToken string
	// for aws only
	AccessKeyId       string
	SecretKey         string
	SecurityGroupID   string
	SecurityGroupName string

	Instances     []Instance `json:"-"`
	InstanceLimit int
	Available     bool
	HealthCheck   bool

	// for create snapshot
	SnapshotID    string
	SnapshotName  string
	OldSnapShotID []string `json:"-"`
	SnapshotFound bool
	SSHKeyFound   bool
	SSHPublicKey  string
	SSHPrivateKey string
	SSHKeyID      string
	SSHUser       string

	// for create
	CreatedInstance Instance `json:"-"`
	Region          string
	Size            string
	SSHKeyName      string

	// mics
	SwapSizeMap       map[string]int `json:"-"`
	IsBackgroundCheck bool

	// for building
	ProviderConfig ConfigProvider `json:"-"`
	Opt            libs.Options   `json:"-"`

	// for retry
	BackOff *backoff.ExponentialBackOff `json:"-"`
	// client of vendor
	Client interface{} `json:"-"`
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

	if providerName == "aws" {
		// token should be 'AccessKeyId,SecretKey'
		provider.AccessKeyId = strings.TrimSpace(strings.Split(token, ",")[0])
		provider.SecretKey = strings.TrimSpace(strings.Split(token, ",")[1])
		provider.Token = token
	}

	provider.InitClient()
	return provider, nil
}

// InitProviderWithConfig init provider object to easier interact with cloud provider
func InitProviderWithConfig(opt libs.Options, providerConfig ConfigProvider) (Provider, error) {
	var provider Provider
	provider.ProviderName = providerConfig.Provider
	provider.Token = providerConfig.Token

	// for aws only
	provider.AccessKeyId = providerConfig.AccessKeyId
	provider.SecretKey = providerConfig.SecretKey
	if provider.AccessKeyId != "" {
		provider.Token = provider.AccessKeyId + "," + provider.SecretKey
	}

	provider.ProviderConfig = providerConfig
	provider.Opt = opt
	if opt.Cloud.BackgroundRun {
		provider.IsBackgroundCheck = true
	}

	if err := provider.InitClient(); err != nil {
		return provider, fmt.Errorf("unable to validate token: %v", provider.Token)
	}

	provider.Prepare()
	return provider, nil
}

func (p *Provider) InitClient() (err error) {
	if p.Token == "" && p.AccessKeyId == "" {
		utils.ErrorF("empty or invalid token: %v", p.Token)
		return fmt.Errorf("empty or invalid token")
	}
	if len(p.Token) > 5 {
		p.RedactedToken = p.Token[:5] + "***" + p.Token[len(p.Token)-5:len(p.Token)]
		if !p.IsBackgroundCheck {
			utils.InforF("Init %v provider with token: %v", color.HiYellowString(p.ProviderName), color.HiCyanString(p.RedactedToken))
		}
	}

	p.Available = true
	switch p.ProviderName {
	case "do", "digitalocean":
		p.ClientDO()
	case "ln", "line", "linode":
		p.ClientLinode()
	case "aw", "aws", "asw":
		p.ClientAWS()
	default:
		p.ClientDO()
	}

	// skip balance check if health check
	if p.HealthCheck {
		return nil
	}

	switch p.ProviderName {
	case "do", "digitalocean":
		err = p.AccountDO()
	case "ln", "line", "linode":
		err = p.AccountLN()
	case "aw", "aws", "asw":
		err = p.AccountAWS()
	default:
		err = p.AccountDO()
	}

	return err
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
	case "ln", "line", "linode":
		p.DefaultLinode()
	case "aw", "aws", "asw":
		p.DefaultAWS()
	default:
		p.DefaultDO()
	}

	p.Action(GetSSHKey)
	p.Action(ListImage)

	if p.SSHKeyID != "" {
		if !p.IsBackgroundCheck {
			utils.InforF("Found SSH Key ID: %v", color.HiBlueString(p.SSHKeyID))
		}
	}
}

func (p *Provider) Action(actionName string, params ...interface{}) error {
	var err error
	var param interface{}
	if len(params) > 0 {
		param = params[0]
	}

	if !p.IsBackgroundCheck {
		utils.InforF("[%v] running action: %v", p.ProviderName, color.HiBlueString(actionName))
	}

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
			err = p.GetInstanceInfo(param)
		case BootInstance:
			err = p.BootInstance(param)
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
