package distribute

import (
    "bufio"
    "fmt"
    "github.com/Shopify/yaml"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/provider"
    "github.com/j3ssie/osmedeus/utils"
    "github.com/olekukonko/tablewriter"
    "golang.org/x/term"
    "os"
    "strings"
    "syscall"
)

func InitCloudSetup(opt libs.Options) {
    var supportedProvider = []string{"digitalocean", "linode"}
    fmt.Println("🔮 Start cloud setup wizard 🔮")
    fmt.Println("Currently only these providers are supported: ", color.HiYellowString("%v", supportedProvider))

    var configProviders provider.ConfigProviders
    configProviders.Builder.BuildRepo = opt.Cloud.BuildRepo
    if configProviders.Builder.BuildRepo == "" {
        configProviders.Builder.BuildRepo = StringPrompt("🌀 Enter premium install script URL (e.g: https://long-url-here/x/premium.sh)?", "")
    }
    configProviders.Builder.PublicKey = opt.Cloud.PublicKey
    configProviders.Builder.SecretKey = opt.Cloud.SecretKey

    for {
        configProvider := generateProvider()
        configProviders.Clouds = append(configProviders.Clouds, configProvider)
        if stop := StringPrompt("🌀 Do you want to add more provider (y/N)?", "n"); stop != "y" {
            break
        }
    }

    data, err := yaml.Marshal(&configProviders)
    if err != nil {
        return
    }

    //utils.DebugF(string(data))
    if override := StringPrompt("🧙 Do you want to override the old config at "+color.HiGreenString(opt.CloudConfigFile)+" (Y/n)?", "y"); override != "n" {
        _, err := utils.WriteToFile(opt.CloudConfigFile, string(data))
        if err != nil {
            utils.WarnF("error to write provider config: %v", opt.CloudConfigFile)
        }
    }

    if isValidate := StringPrompt("❄️  Do you want to validate your provider config (Y/n)?", "y"); isValidate == "n" {
        return
    }
    fmt.Printf("💡 You also can manually rebuild the snapshot with the command: %s\n", color.HiCyanString("%s provider build --rebuild", libs.BINARY))

    var cloudRunners []CloudRunner
    for _, configProvider := range configProviders.Clouds {
        cloudRunner, err := ValidateProvider(opt, configProvider)
        if err != nil {
            utils.ErrorF("error validate config: %v -- %v", configProvider.Name, configProvider.RedactedToken)
            continue
        }
        cloudRunners = append(cloudRunners, cloudRunner)
    }

    content := [][]string{}
    for _, cloudRunner := range cloudRunners {
        row := []string{
            cloudRunner.Provider.ProviderName,
            cloudRunner.Provider.RedactedToken,
            cloudRunner.Provider.SSHKeyID,
            cloudRunner.Provider.SnapshotID,
        }
        content = append(content, row)
    }
    table := tablewriter.NewWriter(os.Stderr)
    table.SetAutoFormatHeaders(false)
    table.SetHeader([]string{"Provider", "Token", "SSH Key ID", "Osmedeus Snapshot ID"})
    table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
    table.SetCenterSeparator("|")
    table.AppendBulk(content) // Add Bulk Data
    table.Render()

}

func generateProvider() provider.ConfigProvider {
    configProvider := provider.ConfigProvider{
        Token:        "",
        Provider:     "digitalocean",
        DefaultImage: "debian-10-x64",
        Size:         "s-2vcpu-4gb",
        Region:       "sfo3",
        Limit:        0,
    }

    //#   provider: "digitalocean"
    //#   name: "do-osmp"
    //#   default_image: "debian-10-x64"
    //#   size: "s-2vcpu-4gb"
    //#   region: "sfo3"

    //#   provider: "linode"
    //#   name: "linode-osmp"
    //#   default_image: "linode/debian10"
    //#   size: "g6-standard-1"
    //#   region: "us-east"

    configProvider.Provider = StringPrompt("🌀 What is your cloud provider?", configProvider.Provider)

    switch configProvider.Provider {
    case "do", "digitalocean":
        configProvider.Provider = "digitalocean"
        fmt.Printf("==> provider selected: %s\n", color.HiBlueString("digitalocean"))
    case "ln", "line", "linode":
        configProvider = provider.ConfigProvider{
            Token:        "",
            Provider:     "linode",
            DefaultImage: "linode/debian10",
            Size:         "g6-standard-1",
            Region:       "us-east",
            Limit:        0,
        }
        fmt.Printf("==> provider selected: %s\n", color.HiBlueString("linode"))

    default:
        configProvider.Provider = "digitalocean"
        fmt.Printf("==> provider selected: %s\n", color.HiBlueString("digitalocean"))

    }

    configProvider.Token = credentials()
    configProvider.Name = fmt.Sprintf("digitalocean-%s", utils.RandomString(6))
    configProvider.DefaultImage = StringPrompt("\n🌀 Choose "+color.HiGreenString("base image")+" for building Osmedeus Image?", configProvider.DefaultImage)
    configProvider.Size = StringPrompt("🌀 Choose "+color.HiGreenString("instance type")+" for running the scan?", configProvider.Size)
    configProvider.Region = StringPrompt("🌀 Choose "+color.HiGreenString("instance region")+" for running the scan?", configProvider.Region)

    return configProvider
}

func credentials() string {
    fmt.Printf("🌀 Enter your %s? ", color.HiGreenString("API token"))
    var token string
    for {
        byteToken, err := term.ReadPassword(int(syscall.Stdin))
        if err == nil && len(byteToken) > 6 {
            token = strings.TrimSpace(string(byteToken))
            break
        }
        utils.WarnF("Looks like your token is invalid. Please try again: %v", token)

    }

    redactedToken := token[:5] + "***" + token[len(token)-5:]
    fmt.Printf("Your token has been saved: %v", color.HiBlueString(redactedToken))
    return token
}

// StringPrompt asks for a string value using the label
func StringPrompt(label string, alt string) string {
    var s string
    r := bufio.NewReader(os.Stdin)
    for {
        fmt.Fprintf(os.Stderr, fmt.Sprintf("%v (default: %s): ", label, color.HiCyanString(alt)))
        s, _ = r.ReadString('\n')
        s = strings.TrimSpace(strings.ToLower(s))
        if s == "" {
            if alt != "" {
                return alt
            }
            utils.WarnF("Blank input doesn't allow, please specify one")
        }

        if s != "" {
            break
        }
    }
    return strings.TrimSpace(s)
}

// ValidateProvider setup new provider
func ValidateProvider(opt libs.Options, providerConfig provider.ConfigProvider) (CloudRunner, error) {
    var cloudRunner CloudRunner
    cloudRunner.Opt = opt
    cloudRunner.Prepare()

    providerCloud, err := provider.InitProviderWithConfig(opt, providerConfig)
    if err != nil {
        return cloudRunner, err
    }
    cloudRunner.Provider = providerCloud

    // check if snapshot is okay or not
    if !cloudRunner.Provider.SnapshotFound {
        utils.InforF("No Snapshot found: %v", cloudRunner.Provider.SnapshotName)
        err = cloudRunner.Provider.BuildImage()
        if err != nil {
            utils.ErrorF("error build snapshot at %v", cloudRunner.Provider.ProviderConfig.BuildFile)
            return cloudRunner, err
        }
    }

    return cloudRunner, nil
}
