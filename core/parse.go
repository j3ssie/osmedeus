package core

import (
	"bytes"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/fatih/color"
	"github.com/flosch/pongo2/v6"
	"github.com/spf13/cast"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Shopify/yaml"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
)

// ResolveData resolve template from signature file
func ResolveData(format string, data map[string]string) string {
	// for backward compatibility because new template using `{{variable}}` instead of `{{.variable}}`
	if strings.Contains(format, "{{.") {
		return OldResolveData(format, data)
	}

	variable := make(map[string]interface{})
	for k, v := range data {
		variable[k] = v
	}
	if tpl, err := pongo2.FromString(format); err == nil {

		out, ok := tpl.Execute(variable)

		if ok == nil {
			return out
		}
		utils.ErrorF("Error when resolve template: %v", ok)
	}
	return format
}

// OldResolveData resolve template from signature file
func OldResolveData(format string, data map[string]string) string {
	t := template.Must(template.New("").Parse(format))

	buf := &bytes.Buffer{}
	err := t.Execute(buf, data)
	if err != nil {
		utils.ErrorF("Error render: %v -- %v", format, err)
		return format
	}
	return buf.String()
}

// ResolveSlice resolve template from signature file
func ResolveSlice(slice []string, data map[string]string) (resolveSlice []string) {
	for _, s := range slice {
		resolveSlice = append(resolveSlice, ResolveData(s, data))
	}
	return resolveSlice
}

// AltResolveVariable just like ResolveVariable but looking for [[.var]]
func AltResolveVariable(format string, data map[string]string) string {
	t := template.Must(template.New("").Delims("[[", "]]").Parse(format))
	buf := &bytes.Buffer{}
	err := t.Execute(buf, data)
	if err != nil {
		return format
	}
	return buf.String()
}

// ParseFlow parse mode file
func ParseFlow(flowFile string) (libs.Flow, error) {
	utils.DebugF("Parsing workflow at: %v", color.HiGreenString(flowFile))
	var flow libs.Flow
	yamlFile, err := ioutil.ReadFile(flowFile)
	if err != nil {
		utils.ErrorF("YAML parsing err: %v -- #%v ", flowFile, err)
		return flow, err
	}
	err = yaml.Unmarshal(yamlFile, &flow)
	if err != nil {
		utils.ErrorF("Error unmarshal: %v -- %v", flowFile, err)
		return flow, err
	}

	if flow.Usage != "" && strings.Contains(flow.Usage, "{{.this_file}}") {
		flow.Usage = strings.ReplaceAll(flow.Usage, "{{.this_file}}", flowFile)
	}
	return flow, nil
}

// ParseModules parse module file
func ParseModules(moduleFile string) (libs.Module, error) {
	utils.DebugF("Parsing module at: %v", color.HiCyanString(moduleFile))

	var module libs.Module

	yamlFile, err := ioutil.ReadFile(moduleFile)
	if err != nil {
		utils.ErrorF("YAML parsing err: %v -- #%v ", moduleFile, err)
		return module, err
	}
	err = yaml.Unmarshal(yamlFile, &module)
	if err != nil {
		utils.ErrorF("Error unmarshal: %v -- %v", moduleFile, err)
		return module, err
	}
	module.ModulePath = moduleFile
	if module.Usage != "" && strings.Contains(module.Usage, "{{this_file}}") {
		module.Usage = strings.ReplaceAll(module.Usage, "{{this_file}}", moduleFile)
	}
	return module, err
}

// ParseInputFormat format input
func ParseInputFormat(raw string, options libs.Options) map[string]string {
	target := make(map[string]string)
	target["RawFormat"] = raw

	jsonParsed, err := gabs.ParseJSON([]byte(raw))
	if err != nil {
		return target
	}

	// parse Target first if found one
	rawURL, ok := jsonParsed.ChildrenMap()["Target"]
	if !ok {
		panic("missing Target in special input")
	}

	target = ParseInput(cast.ToString(rawURL.Data()), options)

	// override the whole thing
	for k, v := range jsonParsed.ChildrenMap() {
		target[k] = fmt.Sprintf("%v", v.Data())
	}

	return target
}

// ParseInput parse input for routine
func ParseInput(raw string, options libs.Options) map[string]string {
	ROptions := ParseTarget(raw)
	if options.EnableFormatInput {
		// avoid the loophole
		options.EnableFormatInput = false
		ROptions = ParseInputFormat(raw, options)
		return ROptions
	}
	// some data stuff
	dir, err := os.Getwd()
	if err == nil {
		ROptions["CWD"] = dir
	}

	// default threads variables
	ROptions["Threads"] = cast.ToString(options.Threads)
	ROptions["threads"] = cast.ToString(options.Threads)
	ROptions["thread"] = cast.ToString(options.Threads)
	ROptions["baseThreads"] = cast.ToString(options.Threads)

	ROptions["Version"] = libs.VERSION
	ROptions["WSCDN"] = options.Cdn.WSURL
	ROptions["CDN"] = options.Cdn.URL
	ROptions["Date"] = time.Now().Format("2006-01-02")
	ROptions["TS"] = utils.GetCurrentDay()

	/* --- start to load default Env --- */
	// ~/osmedeus-base
	ROptions["BaseFolder"] = utils.NormalizePath(strings.TrimLeft(options.Env.BaseFolder, "/"))
	ROptions["Plugins"] = options.Env.BinariesFolder
	ROptions["Binaries"] = options.Env.BinariesFolder

	ROptions["Data"] = options.Env.DataFolder
	ROptions["Workflow"] = options.Env.WorkFlowsFolder
	ROptions["Scripts"] = options.Env.WorkFlowsFolder
	ROptions["Cloud"] = options.Env.CloudConfigFolder

	// ~/.osmedeus/clouds
	//ROptions["CWorkspaces"] = options.Env.CloudDataFolder
	ROptions["Workspaces"] = options.Env.WorkspacesFolder
	if options.Scan.BaseWorkspace != "" {
		ROptions["Workspaces"] = options.Scan.BaseWorkspace
	}
	ROptions["Storages"] = options.Env.StoragesFolder
	/* --- end of load default Env --- */

	ROptions["Workspace"] = utils.CleanPath(raw)
	if options.Scan.CustomWorkspace != "" {
		ROptions["Workspace"] = utils.CleanPath(options.Scan.CustomWorkspace)
	}
	ROptions["Output"] = path.Join(ROptions["Workspaces"], ROptions["Workspace"])

	// params in workflow file
	if len(options.Flow.Params) > 0 {
		for _, param := range options.Flow.Params {
			for k, v := range param {
				v = ResolveData(v, ROptions)
				if strings.HasPrefix(v, "~/") {
					v = utils.NormalizePath(v)
				}
				ROptions[k] = v
			}
		}
	}

	return ROptions
}

// ParseParams parse more params from cli
func ParseParams(rawParams []string) map[string]string {
	params := make(map[string]string)
	for _, item := range rawParams {
		if strings.Contains(item, "=") {
			data := strings.Split(item, "=")
			params[data[0]] = strings.Replace(item, data[0]+"=", "", -1)
		}
	}
	return params
}

// ParseTarget parsing target and some variable for template
func ParseTarget(raw string) map[string]string {
	target := make(map[string]string)
	if raw == "" {
		return target
	}
	target["Target"] = raw
	u, err := url.Parse(raw)

	// something wrong so parsing it again
	if err != nil || u.Scheme == "" || strings.Contains(u.Scheme, ".") {
		raw = fmt.Sprintf("https://%v", raw)
		u, err = url.Parse(raw)
		if err != nil {
			return target
		}
		// fmt.Println("parse again")
	}
	var hostname string
	var query string
	port := u.Port()
	// var domain string
	domain := u.Hostname()

	query = u.RawQuery
	if u.Port() == "" {
		if strings.Contains(u.Scheme, "https") {
			port = "443"
		} else {
			port = "80"
		}

		hostname = u.Hostname()
	} else {
		// ignore common port in Host
		if u.Port() == "443" || u.Port() == "80" {
			hostname = u.Hostname()
		} else {
			hostname = u.Hostname() + ":" + u.Port()
		}
	}

	target["Scheme"] = u.Scheme
	target["Path"] = u.Path
	target["Domain"] = domain

	target["Org"] = domain
	suffix, ok := publicsuffix.PublicSuffix(domain)
	if ok {
		target["Org"] = strings.Replace(domain, fmt.Sprintf(".%s", suffix), "", -1)
	} else {
		if strings.Contains(domain, ".") {
			parts := strings.Split(domain, ".")
			if len(parts) == 2 {
				target["Org"] = parts[0]
			} else {
				target["Org"] = parts[len(parts)-2]
			}
		}
	}

	target["Host"] = hostname
	target["Port"] = port
	target["RawQuery"] = query

	if (target["RawQuery"] != "") && (port == "80" || port == "443") {
		target["URL"] = fmt.Sprintf("%v://%v%v?%v", target["Scheme"], target["Host"], target["Path"], target["RawQuery"])
	} else if port != "80" && port != "443" {
		target["URL"] = fmt.Sprintf("%v://%v:%v%v?%v", target["Scheme"], target["Domain"], target["Port"], target["Path"], target["RawQuery"])
	} else {
		target["URL"] = fmt.Sprintf("%v://%v%v", target["Scheme"], target["Host"], target["Path"])
	}

	uu, _ := url.Parse(raw)
	target["BaseURL"] = fmt.Sprintf("%v://%v", uu.Scheme, uu.Host)
	target["Extension"] = filepath.Ext(target["BaseURL"])

	return target
}

func IsRootDomain(raw string) bool {
	suffix, ok := publicsuffix.PublicSuffix(raw)
	if ok {
		return false
	}

	input := strings.ReplaceAll(raw, fmt.Sprintf(".%s", suffix), "")
	if strings.Count(input, ".") == 1 && strings.Count(input, "/") == 0 {
		return true
	}
	return false
}
