package core

import (
    "fmt"
    "github.com/cenkalti/backoff/v4"
    "github.com/fatih/color"
    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    "os"
    "os/exec"
    "path"
    "strings"
    "time"
)

/* Mostly calling OS commands for double-check the PATH too */

func UpdateMetadata(opt libs.Options) {
    // t.Format("02-Jan-2006")
    // ~/.osmedeus/update/metadata.json
    opt.Update.UpdateConfig = path.Join(opt.Env.RootFolder, "update")
    utils.MakeDir(opt.Update.UpdateConfig)

    // check if it is a default one or not
    if opt.Update.MetaDataURL == "" || opt.Update.MetaDataURL == "META_URL" {
        return
    }

    utils.DebugF("Updating metadata information from %v", opt.Update.MetaDataURL)
    metadataFile := path.Join(opt.Update.UpdateConfig, "metadata.json")
    if utils.FileExists(metadataFile) {
        os.RemoveAll(metadataFile)
    }
    cmd := fmt.Sprintf("wget --no-check-certificate -q %s -O %s", opt.Update.MetaDataURL, metadataFile)
    _, err := utils.RunCommandWithErr(cmd)
    if err != nil {
        utils.ErrorF("error getting update metadata from: %v", opt.Update.MetaDataURL)
    }
}

func GitUpdate(opt libs.Options) error {
    cmd := fmt.Sprintf("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no -i %v' git clone --depth=1 %v %v", opt.Storages["secret_key"], opt.Update.UpdateURL, opt.Update.UpdateFolder)
    _, err := utils.RunCommandWithErr(cmd)
    return err
}

func HTTPUpdate(opt libs.Options) error {
    cmd := fmt.Sprintf("wget --no-check-certificate -q %s -O %s", opt.Update.UpdateURL, opt.Update.UpdateFolder)
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
    cmd := fmt.Sprintf("wget --no-check-certificate -q %s -O %s/install.sh", opt.Update.UpdateURL, updateScript)
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
    utils.Move(path.Join(opt.Update.UpdateFolder, "scripts"), opt.Env.ScriptsFolder)

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
