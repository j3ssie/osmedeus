package core

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/provider"
	"github.com/j3ssie/osmedeus/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/go-homedir"
)

/* Mostly calling OS commands for double-check the PATH too */

func CheckUpdate(options *libs.Options) bool {
	var shouldUpdate bool
	// ~/.osmedeus/update/metadata.json
	options.Update.UpdateConfig = path.Join(utils.NormalizePath(options.Env.RootFolder), "update")
	utils.MakeDir(options.Update.UpdateConfig)

	if options.Update.MetaDataURL == "" {
		options.Update.MetaDataURL = fmt.Sprintf("%s/public.json", libs.METADATA)
		if options.PremiumPackage {
			options.Update.MetaDataURL = fmt.Sprintf("%s/premium.json", libs.METADATA)
		}
	}

	utils.InforF("Checking metadata information from: %v", options.Update.MetaDataURL)
	metadataFile := path.Join(options.Update.UpdateConfig, "metadata.json")

	var oldMetaData libs.UpdateMetaData
	oldMetaData.CoreVersion = libs.VERSION
	oldMetaData.WorkflowVersion = "v0.0.1"

	if utils.FileExists(metadataFile) {
		oldMetaDataContent := utils.GetFileContent(metadataFile)
		if err := jsoniter.UnmarshalFromString(oldMetaDataContent, &oldMetaData); err != nil {
			utils.ErrorF("error to parse metadata: %v", metadataFile)
			return false
		}
	}

	var newMetaData libs.UpdateMetaData
	res := utils.SendGET("", options.Update.MetaDataURL)
	if res.StatusCode == 200 {
		if err := jsoniter.UnmarshalFromString(res.Body, &newMetaData); err != nil {
			utils.ErrorF("error to parse metadata: %v", options.Update.MetaDataURL)
			return false
		}

		utils.InforF("Writing metadata to: %v", color.HiCyanString(metadataFile))
		if data, err := jsoniter.MarshalToString(&newMetaData); err == nil {
			utils.WriteToFile(metadataFile, data)
		}
	} else {
		utils.ErrorF("error fetching metadata from: %v", options.Update.MetaDataURL)
		return false
	}

	utils.DebugF(res.Body)

	v1, err := version.NewVersion(oldMetaData.CoreVersion)
	if err != nil {
		utils.ErrorF("error parsing version: %v -- %v", oldMetaData.CoreVersion, err)
		return false
	}

	// get from metadata URL
	v2, err := version.NewVersion(newMetaData.CoreVersion)
	if err != nil {
		utils.ErrorF("error parsing version: %v -- %v", newMetaData.CoreVersion, err)

		return false
	}

	// Comparison example. There is also GreaterThan, Equal, and just
	// a simple Compare that returns an int allowing easy >=, <=, etc.
	if v1.LessThan(v2) {
		fmt.Printf("Your current %s %s are outdated. Latest is %s\n", libs.BINARY, color.HiMagentaString("%v", v1), color.HiGreenString("%v", v2))
		shouldUpdate = true
	} else {
		fmt.Printf("You're using %s core latest version %s updated at %s\n", libs.BINARY, color.HiMagentaString("%v", v1), color.HiGreenString("%v", newMetaData.UpdatedAt))
	}

	// check workflow version if core is updated
	if !shouldUpdate {
		wfv1, err := version.NewVersion(oldMetaData.WorkflowVersion)
		if err != nil {
			utils.ErrorF("error parsing version: %v -- %v", oldMetaData.WorkflowVersion, err)
		}
		wfv2, err := version.NewVersion(newMetaData.WorkflowVersion)
		if err != nil {
			utils.ErrorF("error parsing version: %v -- %v", newMetaData.WorkflowVersion, err)
		}
		if wfv1.LessThan(wfv2) {
			fmt.Printf("Your current %s workflow %s are outdated. Latest is %s\n", libs.BINARY, color.HiMagentaString("%v", v1), color.HiGreenString("%v", v2))
			shouldUpdate = true
		}
	}

	if shouldUpdate {
		home, _ := homedir.Dir()
		fmt.Printf("📖 Run %s again to update Check out this page for more detail: %s\n", color.HiGreenString("the same install script"), color.HiGreenString("https://docs.osmedeus.org/installation/"))
		fmt.Printf("💡 If you want a fresh install please run the command: %s\n", color.HiBlueString("rm -rf %s/osmedeus-base %s/.osmedeus", home, home))
	}

	return shouldUpdate
}

func GetUpdateURL(options libs.Options) string {
	if !options.PremiumPackage {
		utils.InforF("Getting update url of public community package")
		return libs.INSTALL
	}
	utils.InforF("💎 Getting update url of premium package")

	providerConfigs, err := provider.ParseProvider(options.CloudConfigFile)
	if err != nil {
		utils.ErrorF("error to parse provider config: %v", err)
		return ""
	}

	return providerConfigs.Builder.BuildRepo
}

func RunUpdate(options libs.Options) error {
	if options.Update.UpdateURL == "" {
		return fmt.Errorf("no update URL")
	}

	if options.Update.CleanOldData {
		utils.InforF("Cleaning old data: %s", color.HiCyanString(options.Env.BaseFolder))
		os.RemoveAll(options.Env.BaseFolder)
	}

	if options.PremiumPackage {
		utils.InforF("Running update from premium package install script")
	} else {
		utils.InforF("Running update from: %v", color.HiCyanString(options.Update.UpdateURL))
	}

	options.Update.UpdateScript = fmt.Sprintf("/tmp/%s-update.sh", libs.BINARY)
	cmd := fmt.Sprintf("rm -rf %s && wget -qO %s %s ", options.Update.UpdateScript, options.Update.UpdateScript, options.Update.UpdateURL)
	_, err := utils.RunCommandWithErr(cmd)
	if err != nil {
		utils.ErrorF("error to run update script: %v", err)
		return err
	}

	if !utils.FileExists(options.Update.UpdateScript) {
		utils.ErrorF("update script doesn't exist: %v", options.Update.UpdateScript)
		return fmt.Errorf("update script doesn't exist: %v", options.Update.UpdateScript)
	}

	if _, err := utils.RunCommandSteamOutput(fmt.Sprintf(`bash %s`, options.Update.UpdateScript)); err != nil {
		utils.ErrorF("error to run update script: %v", err)
		return err
	}
	return nil
}

func GenerateMetaData(options libs.Options) {
	utils.InforF("Generating metadata to: %v", color.HiCyanString(options.Update.GenerateMeta))
	t := time.Now()
	t.Format("2006-01-02T15:04:05")
	var updateData = libs.UpdateMetaData{
		CoreVersion:     libs.VERSION,
		WorkflowVersion: libs.VERSION,
		UpdatedAt:       t.Format("2006-01-02T15:04"),
	}

	if data, ok := jsoniter.MarshalToString(updateData); ok == nil {
		utils.InforF("Generate meta data: %v", data)
		utils.WriteToFile(options.Update.GenerateMeta, data)
	}
}

func GitUpdate(opt libs.Options) error {
	cmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git clone --depth=1 %v %v", opt.Storages["secret_key"], opt.Update.UpdateURL, opt.Update.UpdateFolder)
	_, err := utils.RunCommandWithErr(cmd)
	return err
}

func HTTPUpdate(opt libs.Options) error {
	cmd := fmt.Sprintf("wget -q %s -O %s", opt.Update.UpdateURL, opt.Update.UpdateFolder)
	_, err := utils.RunCommandWithErr(cmd)
	return err
}

func DownloadUpdate(opt libs.Options) error {
	os.RemoveAll(opt.Update.UpdateFolder)
	utils.InforF("Downloading the update folder via %v: %v", opt.Update.UpdateType, opt.Update.UpdateURL)
	var err error

	backOff := backoff.NewExponentialBackOff()
	backOff.MaxElapsedTime = 1200 * time.Second
	backOff.Multiplier = 2.0
	backOff.InitialInterval = 30 * time.Second

	operation := func() error {
		switch strings.ToLower(opt.Update.UpdateType) {
		case "git":
			err = GitUpdate(opt)
		case "s3", "http":
			err = HTTPUpdate(opt)
		default:
			err = GitUpdate(opt)
		}
		if err != nil {
			utils.ErrorF("error downloading update content: %s -- %s", opt.Update.UpdateType, opt.Update.UpdateURL)
		}
		return err
	}
	err = backoff.Retry(operation, backOff)
	if err != nil {
		utils.ErrorF("error downloading update content: %s -- %s", opt.Update.UpdateType, opt.Update.UpdateURL)
		return err
	}
	return nil
}

func Update(opt libs.Options) {
	os.RemoveAll(opt.Update.UpdateFolder)
	utils.MakeDir(opt.Update.UpdateFolder)

	updateScript := fmt.Sprintf("%s/update.sh", opt.Update.UpdateFolder)
	cmd := fmt.Sprintf("wget -q %s -O %s/install.sh", opt.Update.UpdateURL, updateScript)
	if _, err := utils.RunCommandWithErr(cmd); err != nil {
		utils.ErrorF("error downloading the update script: %v", opt.Update.UpdateURL)
		return
	}

	cmd = fmt.Sprintf("bash %s", updateScript)
	if _, err := utils.RunCommandWithErr(cmd); err != nil {
		utils.ErrorF("error running the update script: %v", updateScript)
		return
	}
}

func UpdateVuln(opt libs.Options) {
	utils.InforF("Updating Vulnerability Database only")

	// update nuclei templates
	utils.DebugF("Updating Nuclei Templates")
	nucleiTemplate := utils.NormalizePath("~/nuclei-templates")

	if utils.DirLength(nucleiTemplate) > 0 {
		os.RemoveAll(nucleiTemplate)
	}
	utils.RunOSCommand(fmt.Sprintf("git clone --depth=1 https://github.com/projectdiscovery/nuclei-templates.git %s", nucleiTemplate))
}

func UpdateBase(opt libs.Options) {
	err := DownloadUpdate(opt)
	if err != nil {
		return
	}

	// change the folder since we will update it
	if opt.Update.IsUpdateBin {
		utils.InforF("Updating External binaries")
		binPath := path.Join(opt.Update.UpdateFolder, "binaries")
		utils.Move(binPath, opt.Env.BinariesFolder)
		opt.Update.UpdateFolder = path.Join(opt.Update.UpdateFolder, fmt.Sprintf("%s-base", libs.BINARY))
	}

	// update Env
	utils.InforF("Updating Environments Data")
	utils.Move(path.Join(opt.Update.UpdateFolder, "data"), opt.Env.DataFolder)
	utils.Move(path.Join(opt.Update.UpdateFolder, "workflow"), opt.Env.WorkFlowsFolder)
	utils.Move(path.Join(opt.Update.UpdateFolder, "ose"), opt.Env.OseFolder)
	utils.Move(path.Join(opt.Update.UpdateFolder, "ui"), opt.Env.UIFolder)

	utils.Move(path.Join(opt.Update.UpdateFolder, "clouds"), opt.Env.CloudConfigFolder)
	os.Chmod(opt.Cloud.SecretKey, 0600)

	// update osmedeus core binary
	corePath, err := exec.LookPath(libs.BINARY)
	utils.InforF("Updating %v binary at %v", color.HiCyanString(libs.BINARY), color.HiCyanString(corePath))
	if err == nil {
		os.RemoveAll(corePath)
		newBin := fmt.Sprintf("%s/dist/%s", strings.TrimRight(opt.Update.UpdateFolder, "/"), libs.BINARY)
		unZipCmd := fmt.Sprintf("unzip %s/dist/%s-linux.zip -d %s/dist/", strings.TrimRight(opt.Update.UpdateFolder, "/"), libs.BINARY, strings.TrimRight(opt.Update.UpdateFolder, "/"))
		utils.RunOSCommand(unZipCmd)

		// update binaries in gopath
		goPath := utils.GetOSEnv("GOPATH", "GOPATH")
		if goPath != "GOPATH" {
			goPath = path.Join(goPath, fmt.Sprintf("bin/%s", libs.BINARY))
			os.RemoveAll(goPath)
			utils.RunOSCommand(fmt.Sprintf("cp %s %s", newBin, goPath))

			// go path but in plugins folder
			goPath = path.Join(opt.Env.BinariesFolder, "go", libs.BINARY)
			os.RemoveAll(goPath)
			utils.RunOSCommand(fmt.Sprintf("cp %s %s", newBin, goPath))
		}
		utils.Move(newBin, corePath)
	}

	// update vulnerability signatures
	utils.InforF("Updating Jaeles Signatures")
	jaelesSign := path.Join(opt.Update.UpdateFolder, "pro-signatures")
	if utils.DirLength(jaelesSign) > 0 {
		utils.RunOSCommand(fmt.Sprintf("jaeles config reload --signDir %s", jaelesSign))
		utils.Move(jaelesSign, "~/pro-signatures")
	} else {
		os.RemoveAll(utils.NormalizePath("~/pro-signatures"))
		utils.RunOSCommand(fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %s' git clone --depth=1 git@gitlab.com:j3ssie/pro-signatures ~/pro-signatures", opt.Storages["secret_key"]))
		utils.RunOSCommand(fmt.Sprintf("rm -rf ~/custom-nuclei-template && GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %s' git clone --depth=1 git@gitlab.com:j3ssie/custom-nuclei-template.git ~/custom-nuclei-template", opt.Storages["secret_key"]))
		utils.RunOSCommand("jaeles config reload --signDir ~/pro-signatures")
	}

	// update nuclei templates
	utils.InforF("Updating Nuclei Templates")
	nucleiTemplate := path.Join(opt.Update.UpdateFolder, "nuclei-templates")
	if utils.DirLength(nucleiTemplate) > 0 {
		utils.Move(nucleiTemplate, utils.NormalizePath("~/nuclei-templates"))
	} else {
		utils.RunOSCommand(fmt.Sprintf("git clone --depth=1 https://github.com/projectdiscovery/nuclei-templates.git ~/nuclei-templates"))
	}

	// clean up
	utils.InforF("Clean up update folder")
	os.RemoveAll(opt.Update.UpdateFolder)
}
