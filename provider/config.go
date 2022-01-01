package provider

import (
    "github.com/Shopify/yaml"
    "github.com/j3ssie/osmedeus/utils"
    "io/ioutil"
)

// ConfigProviders cloud config file
type ConfigProviders struct {
    Clouds []ConfigProvider `yaml:"clouds"`
}

// ConfigProvider single element in cloud config file
type ConfigProvider struct {
    // core part
    Name         string `yaml:"name"`
    Token        string `yaml:"token"`
    Provider     string `yaml:"provider"`
    DefaultImage string `yaml:"default_image"`
    Size         string `yaml:"size"`
    Region       string `yaml:"region"`
    Limit        int    `yaml:"limit"`

    // BaseImage     string `yaml:"base"`
    Snapshot     string `yaml:"snapshot"`
    SnapshotID   string `yaml:"snapshot_id"`
    InstanceName string `yaml:"instance"`
    PublicIP     string `yaml:"ip"`
    SshKey       string `yaml:"ssh_key"`

    // for config
    ProviderFolder string
    ConfigFile     string
    BuildFile      string
    BuildData      map[string]string

    // for building
    VarsFile     string
    RunnerFile   string
    BuildCommand string
    BaseFolder   string
    RawCommand   string
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
