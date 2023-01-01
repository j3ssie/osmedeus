package provider

import (
	"io/ioutil"

	"github.com/Shopify/yaml"
	"github.com/j3ssie/osmedeus/utils"
)

// ConfigProviders cloud config file
type ConfigProviders struct {
	Builder Builder          `yaml:"builder"`
	Clouds  []ConfigProvider `yaml:"clouds"`
}

// Builder config for builder file
type Builder struct {
	BuildRepo string `yaml:"build_repo"`
	PublicKey string `yaml:"public_key"`
	SecretKey string `yaml:"secret_key"`
}

type ConfigProvider struct {
	// core part
	Name  string `yaml:"name"`
	Token string `yaml:"token"`

	SecretKey   string `yaml:"secret_key"`
	AccessKeyId string `yaml:"access_key"`

	Provider     string `yaml:"provider"`
	DefaultImage string `yaml:"default_image"`
	Size         string `yaml:"size"`
	Region       string `yaml:"region"`
	Limit        int    `yaml:"limit"`

	// BaseImage     string `yaml:"base"`
	RedactedToken string `yaml:"-"`
	Snapshot      string `yaml:"-"`
	SnapshotID    string `yaml:"-"`
	InstanceName  string `yaml:"-"`
	PublicIP      string `yaml:"-"`
	SshKey        string `yaml:"-"`

	// for config
	ProviderFolder string            `yaml:"-"`
	ConfigFile     string            `yaml:"-"`
	BuildFile      string            `yaml:"-"`
	BuildData      map[string]string `yaml:"-"`

	// for building
	VarsFile     string `yaml:"-"`
	RunnerFile   string `yaml:"-"`
	BuildCommand string `yaml:"-"`
	BaseFolder   string `yaml:"-"`
	RawCommand   string `yaml:"-"`
}

// ParseProvider parse cloud file
func ParseProvider(cloudFile string) (ConfigProviders, error) {
	var clouds ConfigProviders
	cloudFile = utils.NormalizePath(cloudFile)

	yamlFile, err := ioutil.ReadFile(cloudFile)
	if err != nil {
		utils.ErrorF("YAML parsing err #%v ", err)
		return clouds, err
	}
	err = yaml.Unmarshal(yamlFile, &clouds)
	if err != nil {
		utils.ErrorF("Error: %v", err)
		return clouds, err
	}

	return clouds, nil
}
