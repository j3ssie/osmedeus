package execution

import (
    "fmt"
    "net/url"
    "os"
    "path"

    "github.com/j3ssie/osmedeus/libs"
    "github.com/j3ssie/osmedeus/utils"
    jsoniter "github.com/json-iterator/go"
)

type CdnSummary struct {
    Targets    []Cdn  `json:"targets"`
    UpdateDate string `json:"update_date"`
    Total      int    `json:"total"`
}

type Cdn struct {
    TargetName  string `json:"target_name"`
    DownloadURL string `json:"download_url"`
    Type        string `json:"type"`
}

// GetTargetFromCDN search for target from CDN Index
func GetTargetFromCDN(options libs.Options, targetWS string) string {
    utils.DebugF("Finding targets from CDN: %s", options.Cdn.Index)
    resp, err := SendGET("", options.Cdn.Index)
    if err != nil {
        return ""
    }

    var cdnSummary CdnSummary
    err = jsoniter.UnmarshalFromString(resp.Body, &cdnSummary)
    if err != nil {
        utils.ErrorF("error parsing body")
        return ""
    }

    if len(cdnSummary.Targets) == 0 {
        utils.ErrorF("index file empty")
        return ""
    }

    for _, target := range cdnSummary.Targets {
        if targetWS == target.TargetName {
            return target.DownloadURL
        }
        // in case we have typo
        if targetWS == fmt.Sprintf("%s.zip", target.TargetName) {
            return target.DownloadURL
        }
    }
    utils.ErrorF("no target match")
    return ""
}

// GetFileFromCDN get file from CDN URL
func GetFileFromCDN(options libs.Options, filename string, dest string) {
    if options.NoCdn {
        return
    }

    u, err := url.Parse(options.Cdn.URL)
    if err != nil {
        return
    }
    u.Path = path.Join(u.Path, filename)
    downloadURL := u.String()
    utils.DebugF("Downloading: %s", downloadURL)

    if !utils.FolderExists(path.Dir(dest)) {
        utils.MakeDir(path.Dir(dest))
    }

    cmd := fmt.Sprintf("wget -qO %s %s", dest, downloadURL)
    Execution(cmd, options)
    if utils.FileLength(dest) <= 0 {
        os.RemoveAll(dest)
        return
    }
}

// GetWSFromCDN get workspace zip from CDN URL
func GetWSFromCDN(options libs.Options, target string, dest string) {
    if options.NoCdn {
        return
    }

    downloadURL := GetTargetFromCDN(options, target)
    if downloadURL == "" {
        utils.ErrorF("Target not found from cdn: %v", target)
        return
    }

    utils.DebugF("Downloading: %s", downloadURL)

    if !utils.FolderExists(path.Dir(dest)) {
        utils.MakeDir(path.Dir(dest))
    }

    cmd := fmt.Sprintf("wget -qO %s %s", dest, downloadURL)
    Execution(cmd, options)
    if utils.FileLength(dest) <= 0 {
        os.RemoveAll(dest)
        return
    }
}

// DownloadFile get file from CDN URL
func DownloadFile(options libs.Options, downloadURL string, dest string) {
    utils.DebugF("Downloading: %s", downloadURL)
    if !utils.FolderExists(path.Dir(dest)) {
        utils.MakeDir(path.Dir(dest))
    }

    cmd := fmt.Sprintf("wget -qO %s %s", dest, downloadURL)
    Execution(cmd, options)

    if utils.FileLength(dest) <= 0 {
        os.RemoveAll(dest)
        return
    }

}
