package provider

import (
	"os"
	"math/rand"
	"time"

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

// ConfigProvider cloud config file for each provider from ~/osmedeus-base/cloud/provider.yaml
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
	Username     string `yaml:"username"`

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

// ParseProvider parse cloud file from ~/osmedeus-base/cloud/provider.yaml
func ParseProvider(cloudFile string) (ConfigProviders, error) {
	var clouds ConfigProviders
	cloudFile = utils.NormalizePath(cloudFile)

	yamlFile, err := os.ReadFile(cloudFile)
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

// mics function

const (
	lowerChars  = "abcdefghijklmnopqrstuvwxyz"
	upperChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberChars = "0123456789"
	symbolChars = "!@#$%^&*()+-=[]{}<>?~"
)

func GeneratePassword(length int) string {
	rand.Seed(time.Now().UnixNano())

	var passwordChars []byte

	// Add at least one of each character type
	passwordChars = append(passwordChars, randomChar(lowerChars))
	passwordChars = append(passwordChars, randomChar(upperChars))
	passwordChars = append(passwordChars, randomChar(numberChars))
	passwordChars = append(passwordChars, randomChar(symbolChars))

	// Add remaining characters randomly
	for i := len(passwordChars); i < length; i++ {
		charType := rand.Intn(4) // 0 for lower, 1 for upper, 2 for number, 3 for symbol
		switch charType {
		case 0:
			passwordChars = append(passwordChars, randomChar(lowerChars))
		case 1:
			passwordChars = append(passwordChars, randomChar(upperChars))
		case 2:
			passwordChars = append(passwordChars, randomChar(numberChars))
		case 3:
			passwordChars = append(passwordChars, randomChar(symbolChars))
		}
	}

	// Shuffle the characters randomly
	rand.Shuffle(len(passwordChars), func(i, j int) {
		passwordChars[i], passwordChars[j] = passwordChars[j], passwordChars[i]
	})

	return string(passwordChars)
}

func randomChar(charset string) byte {
	return charset[rand.Intn(len(charset))]
}
