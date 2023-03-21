package database

import (
	"os"
	"path/filepath"

	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	jsoniter "github.com/json-iterator/go"
)

func GetAllWorkspaces(opt libs.Options) (directories []string) {
	// Open the specified directory
	dir, err := os.Open(opt.Env.WorkspacesFolder)
	if err != nil {
		return directories
	}
	defer dir.Close()

	// Read all entries in the directory
	entries, err := dir.Readdir(-1)
	if err != nil {
		return directories
	}

	for _, entry := range entries {
		wsName := entry.Name()
		directories = append(directories, wsName)
	}

	return directories
}

func GetAllScan(opt libs.Options) (scans []Scan) {
	wss := GetAllWorkspaces(opt)
	for _, wsName := range wss {
		runtimeFile := filepath.Join(opt.Env.WorkspacesFolder, wsName, "runtime")
		if !utils.FileExists(runtimeFile) {
			continue
		}

		// parse the content
		runtimeContent := utils.GetFileContent(runtimeFile)
		wsData := Scan{}
		if err := jsoniter.UnmarshalFromString(runtimeContent, &wsData); err == nil {
			scans = append(scans, wsData)
		}

	}
	return scans
}

func GetSingleScan(wsName string, opt libs.Options) (scan Scan) {
	runtimeFile := filepath.Join(opt.Env.WorkspacesFolder, wsName, "runtime")
	if !utils.FileExists(runtimeFile) {
		return scan
	}

	// parse the content
	runtimeContent := utils.GetFileContent(runtimeFile)
	if err := jsoniter.UnmarshalFromString(runtimeContent, &scan); err == nil {
		return scan
	}
	return scan
}

func GetScanProgress(opt libs.Options) (scans []Scan) {
	rawScans := GetAllScan(opt)

	for _, scan := range rawScans {
		scan.Target = Target{}
		scans = append(scans, scan)
	}

	return scans
}

// func GetWorkspaceDetail(wsName string, opt libs.Options) (workspace Scan) {
// 	runtimeFile := filepath.Join(opt.Env.WorkspacesFolder, wsName, "runtime")
// 	if !utils.FileExists(runtimeFile) {
// 		return workspace
// 	}

// 	// parse the content
// 	runtimeContent := utils.GetFileContent(runtimeFile)
// 	if err := jsoniter.UnmarshalFromString(runtimeContent, &workspace); err == nil {
// 		return workspace
// 	}

// 	return workspace
// }
