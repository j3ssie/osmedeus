package provider

import (
    "fmt"
    "github.com/j3ssie/osmedeus/core"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "os"
    "path"
    "strings"
)

func (p *Provider) PrePareBuildData() {
    contentFile := path.Join(p.Opt.Env.CloudConfigFolder, fmt.Sprintf("providers/%s.provider", p.ProviderName))
    content := utils.GetFileContent(contentFile)

    data := make(map[string]string)
    data["snapshot_name"] = p.SnapshotName
    data["api_token"] = p.Token

    // c.Cloud.ProviderFolder --> ~/.osmedeus/provider/<osmp-name>-v4.x-randomstring
    p.ProviderConfig.ProviderFolder = path.Join(p.Opt.Env.ProviderFolder, fmt.Sprintf("%s-%s", p.SnapshotName, utils.RandomString(6)))
    utils.MakeDir(p.ProviderConfig.ProviderFolder)
    data["ProviderFolder"] = p.ProviderConfig.ProviderFolder

    data["image"] = p.ProviderConfig.DefaultImage
    data["size"] = p.ProviderConfig.Size
    data["region"] = p.ProviderConfig.Region
    data["TS"] = utils.GetTS()

    // generate packer content file to run
    providerString := core.ResolveData(content, data)
    data["Builder"] = providerString

    // ~/osmedeus-base
    data["BaseFolder"] = utils.NormalizePath(strings.TrimLeft(p.Opt.Env.BaseFolder, "/"))
    data["Plugins"] = p.Opt.Env.BinariesFolder
    data["OBin"] = p.Opt.Env.BinariesFolder
    data["Data"] = p.Opt.Env.DataFolder
    data["Cloud"] = p.Opt.Env.CloudConfigFolder
    data["Workflow"] = p.Opt.Env.WorkFlowsFolder
    // ~/.osmedeus/clouds
    data["CWorkspaces"] = p.Opt.Env.CloudDataFolder
    // ~/.osmedeus/workspaces
    data["Workspaces"] = p.Opt.Env.WorkspacesFolder
    data["Binary"] = libs.BINARY
    data["VERSION"] = libs.VERSION
    data["BuildRepo"] = p.Opt.Cloud.BuildRepo

    // for terraform
    data["ssh_public_key"] = p.Opt.Cloud.PublicKeyContent
    data["root_password"] = fmt.Sprintf("osmp-%s", utils.RandomString(8))

    //spew.Dump("data --> ", data)
    //spew.Dump("p.ProviderConfig --> ", p.ProviderConfig)

    p.ProviderConfig.BuildData = data
}

func (p *Provider) BuildImage() (err error) {
    if p.SnapshotFound && !p.Opt.Cloud.ReBuildBaseImage {
        return nil
    }

    p.PrePareBuildData()
    p.DeleteOldSnapshot()

    // p.ProviderConfig.ProviderFolder --> ~/.osmedeus/provider/<osmp-name>

    utils.DebugF("Cleaning old provider build: %s", p.ProviderConfig.ProviderFolder)
    os.RemoveAll(p.ProviderConfig.ProviderFolder)
    utils.MakeDir(p.ProviderConfig.ProviderFolder)

    // generate provision process
    setupContent := utils.GetFileContent(path.Join(p.Opt.Env.CloudConfigFolder, "setup.sh"))
    setupContent = core.ResolveData(setupContent, p.ProviderConfig.BuildData)
    setupFile := path.Join(p.ProviderConfig.ProviderFolder, "setup.sh")
    utils.WriteToFile(setupFile, setupContent)

    // generate build file
    var buildContent string
    buildContentFile := path.Join(p.Opt.Env.CloudConfigFolder, "general-build.packer")
    switch p.ProviderName {
    case "do", "digitalocean":
        buildContentFile = path.Join(p.Opt.Env.CloudConfigFolder, "digitalocean-build.packer")
        if !utils.FileExists(buildContentFile) {
            buildContentFile = path.Join(p.Opt.Env.CloudConfigFolder, "do-build.packer")
        } else {
            buildContentFile = path.Join(p.Opt.Env.CloudConfigFolder, "general-build.packer")
        }
    case "ln", "line", "linode":
        buildContentFile = path.Join(p.Opt.Env.CloudConfigFolder, "linode-build.packer")
        if !utils.FileExists(buildContentFile) {
            buildContentFile = path.Join(p.Opt.Env.CloudConfigFolder, "ln-build.packer")
        }
    default:
        buildContentFile = path.Join(p.Opt.Env.CloudConfigFolder, "general-build.packer")
    }

    buildContent = utils.GetFileContent(buildContentFile)
    if buildContent == "" {
        errStr := fmt.Sprintf("Build file content not found at: %v", buildContentFile)
        utils.ErrorF(errStr)
        return fmt.Errorf(errStr)
    }

    buildContent = core.ResolveData(buildContent, p.ProviderConfig.BuildData)
    buildFile := path.Join(p.ProviderConfig.ProviderFolder, "build.json")
    p.ProviderConfig.BuildFile = buildFile
    utils.WriteToFile(buildFile, buildContent)
    utils.InforF("Write build provision of %s to: %s", p.ProviderName, buildFile)

    // actually run building
    err = p.Action(RunBuild)
    if err != nil {
        p.SnapshotFound = false
        return err
    }

    err = p.Action(ListImage)
    return err
}

// RunBuild run the packer command
func (p *Provider) RunBuild() error {
    packerBinary := fmt.Sprintf("%s/packer", p.Opt.Env.BinariesFolder)
    if !utils.FileExists(packerBinary) {
        packerBinary = "packer"
    }

    cmd := fmt.Sprintf("%s validate %s", packerBinary, p.ProviderConfig.BuildFile)
    out, err := utils.RunCommandWithErr(cmd)
    if err != nil {
        utils.ErrorF(out)
        return err
    }
    utils.InforF("Config looks good at: %s", p.ProviderConfig.BuildFile)

    // really start to build stuff here
    utils.GoodF("Start packer build for: %s", p.ProviderConfig.BuildFile)
    cmd = fmt.Sprintf("%s build %s", packerBinary, p.ProviderConfig.BuildFile)
    out, _ = utils.RunCommandWithErr(cmd)

    if !strings.Contains(out, fmt.Sprintf("%v scan -f", libs.BINARY)) {
        if !strings.Contains(out, fmt.Sprintf("%v: command not found", libs.BINARY)) {
            utils.ErrorF(out)
            return fmt.Errorf("error running provisioning")
        }
    }
    return nil
}
