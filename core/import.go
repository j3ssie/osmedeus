package core

import (
    "fmt"
    "github.com/Jeffail/gabs/v2"
    "github.com/j3ssie/osmedeus/database"
    "github.com/j3ssie/osmedeus/utils"
    jsoniter "github.com/json-iterator/go"
    "github.com/robertkrimen/otto"
    "github.com/spf13/cast"
    "path"
    "path/filepath"
    "strings"
)

func (r *Runner) LoadImportScripts() string {
    var output string

    // DB scripts

    r.VM.Set("ImportSubdomain", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportSubdomain(src)
        return otto.Value{}
    })

    r.VM.Set("ImportDns", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportDns(src)
        return otto.Value{}
    })

    r.VM.Set("ImportTech", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportTech(src)
        return otto.Value{}
    })

    r.VM.Set("ImportHTTPJson", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportHTTPJson(src)
        return otto.Value{}
    })

    r.VM.Set("ImportScreenShotJson", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportScreenShotJson(src)
        return otto.Value{}
    })

    r.VM.Set("ImportPortJson", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportPortJson(src)
        return otto.Value{}
    })

    r.VM.Set("ImportJaelesVulnJson", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportJaelesVulnJson(src)
        return otto.Value{}
    })

    r.VM.Set("ImportNucleiVulnJson", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportNucleiVulnJson(src)
        return otto.Value{}
    })

    r.VM.Set("ImportDirectoryJson", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportDirectoryJson(src)
        return otto.Value{}
    })

    r.VM.Set("ImportLinks", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportLinks(src)
        return otto.Value{}
    })

    r.VM.Set("ImportArchive", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportArchive(src)
        return otto.Value{}
    })

    r.VM.Set("ImportIPRange", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportIPRange(src)
        return otto.Value{}
    })

    r.VM.Set("ImportCred", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportCred(src)
        return otto.Value{}
    })

    r.VM.Set("ImportCert", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportCert(src)
        return otto.Value{}
    })

    r.VM.Set("ImportCloudBrute", func(call otto.FunctionCall) otto.Value {
        src := call.Argument(0).String()
        r.ImportCloudBrute(src)
        return otto.Value{}
    })

    r.VM.Set("SummaryTarget", func(call otto.FunctionCall) otto.Value {
        r.ScanObj.SummaryTarget()
        return otto.Value{}
    })

    return output
}

func (r *Runner) ImportSubdomain(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    domains := utils.ReadingLines(src)
    var objs []database.Asset

    for _, domain := range domains {
        domain = strings.TrimSpace(domain)
        if domain == "" {
            continue
        }

        obj := database.Asset{
            AssetValue: domain,
            //Dns:           nil,
            HTTP:          nil,
            Directory:     nil,
            Vulnerability: nil,
            ScanRefer:     r.ScanObj.ID,
            TargetRefer:   r.ScanObj.TargetRefer,
        }
        objs = append(objs, obj)
    }
    database.ImportAssets(objs)
}

func (r *Runner) ImportDns(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)
    for _, line := range content {
        line = strings.TrimSpace(line)
        if line == "" || !strings.Contains(line, " ") {
            continue
        }

        raw := strings.Split(line, " ")
        domain := strings.Trim(raw[0], ".")
        dnsType := strings.TrimSpace(raw[1])
        dnsValue := strings.Trim(raw[2], ".")

        if strings.TrimSpace(domain) == "" || strings.TrimSpace(dnsValue) == "" {
            continue
        }

        dnsChecksum := utils.GenHash(fmt.Sprintf("%s-%s-%s", domain, dnsType, dnsValue))

        obj := database.Dns{
            Domain:      domain,
            DnsType:     dnsType,
            DnsValue:    dnsValue,
            DnsChecksum: dnsChecksum,
            ScanRefer:   r.ScanObj.ID,
            TargetRefer: r.ScanObj.TargetRefer,
        }
        obj.Create()
    }
}

func (r *Runner) ImportTech(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)
    for _, line := range content {
        if !strings.Contains(line, ";;") {
            utils.ErrorF("Invalid format: %v", line)
            continue
        }

        // data should be domain|%v;;techs|%v
        domain := strings.TrimPrefix(strings.Split(line, ";;")[0], "domain|")
        techs := strings.TrimPrefix(strings.Split(line, ";;")[1], "techs|")
        if strings.TrimSpace(techs) == "" {
            utils.ErrorF("Invalid format: %v", line)
            continue
        }

        obj := database.Asset{
            AssetValue:  domain,
            Technology:  techs,
            IsAlive:     true,
            ScanRefer:   r.ScanObj.ID,
            TargetRefer: r.ScanObj.TargetRefer,
        }

        obj.UpdateTech()
    }
}

func (r *Runner) ImportHTTPJson(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    content := utils.ReadingLines(src)

    baseResult, _ := filepath.Abs(src)
    baseResult = path.Dir(baseResult)
    utils.DebugF("Set Base Dir for content: %v", baseResult)

    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }
        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            continue
        }

        //  {"url":"https://sso-na.tesla.com","title":"Blank Title","checksum":"527ef0b39a78caf74f54ca5b2ffb59bcfe688685","content_file":"overview/contents/https___sso-na.tesla.com.txt","status":"302","time":"0.046268348","length":"90","redirect":"https://teamchatgl.tesla.com/"}
        URL := jsonParsed.S("url").Data().(string)
        title := jsonParsed.S("title").Data().(string)
        checksum := jsonParsed.S("checksum").Data().(string)
        contentPath := jsonParsed.S("content_file").Data().(string)
        status := jsonParsed.S("status").Data().(string)
        length := jsonParsed.S("length").Data().(string)
        redirect := jsonParsed.S("redirect").Data().(string)

        if !utils.FileExists(contentPath) {
            contentPath = path.Join(baseResult, contentPath)
        }

        // base64 content
        data := "No-Content"
        if !strings.Contains(contentPath, "No-Content") {
            //utils.DebugF("Reading HTML content: %v", contentPath)
            data = utils.ImageAsBase64(contentPath)
        }

        obj := database.HTTP{
            URL:           URL,
            Title:         title,
            Checksum:      checksum,
            StatusCode:    cast.ToInt(status),
            ContentLength: cast.ToInt(length),
            HTTPContent:   data,
            Redirect:      redirect,
            ScanRefer:     r.ScanObj.ID,
            TargetRefer:   r.ScanObj.TargetRefer,
        }
        obj.CreateHTTP()
    }
}

func (r *Runner) ImportScreenShotJson(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }

    baseResult, _ := filepath.Abs(src)
    baseResult = path.Dir(baseResult)
    content := utils.ReadingLines(src)
    utils.DebugF("Set Base Dir for screenshot: %v", baseResult)

    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }
        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            continue
        }

        URL := jsonParsed.S("url").Data().(string)
        imgPath := jsonParsed.S("image").Data().(string)
        tech := jsonParsed.S("tech").Data().(string)
        if tech != "" {
            //domain, _ := utils.GetDomain(URL)
            utils.DebugF("more tech: %v", tech)
            //database.NewTech(wsObj, domain, tech)
        }

        if !utils.FileExists(imgPath) {
            imgPath = path.Join(baseResult, imgPath)
        }
        imgData := utils.ImageAsBase64(imgPath)

        //database.NewScreenshot(wsObj, URL, data)

        obj := database.HTTP{
            URL: URL,
            //Title:          tittle,
            //Checksum:       "",
            //StatusCode:     0,
            //ContentLength:  0,
            //HTTPContent:    "",
            ScreenShotData: imgData,
            ScanRefer:      r.ScanObj.ID,
            TargetRefer:    r.ScanObj.TargetRefer,
        }
        obj.CreateScreenShot()
    }
}

type PortObj struct {
    Protocol string
    PortID   string
    State    string
    Service  struct {
        Name    string
        Product string
        Cpe     string
    }
    Script struct {
        ID     string
        Output string
    }
}

func (r *Runner) ImportPortJson(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }
    content := utils.ReadingLines(src)

    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }
        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            continue
        }

        var portObj []PortObj

        ipAddress := jsonParsed.S("IPAddress").Data().(string)
        rawPorts := jsonParsed.S("Ports").Bytes()
        err = jsoniter.Unmarshal(rawPorts, &portObj)
        if err != nil || len(portObj) == 0 {
            continue
        }

        var ports []string
        for _, p := range portObj {
            info := fmt.Sprintf("%v/%v/%v", p.PortID, p.Protocol, p.Service.Product)
            ports = append(ports, strings.Trim(info, "/"))
        }
        if len(ports) == 0 {
            continue
        }

        obj := database.Dns{
            DnsValue:    ipAddress,
            DnsType:     "A",
            Ports:       strings.Join(ports, ","),
            ScanRefer:   r.ScanObj.ID,
            TargetRefer: r.ScanObj.TargetRefer,
        }
        obj.UpdatePort()
    }
}

// ImportJaelesVulnJson import new asset to DB
func (r *Runner) ImportJaelesVulnJson(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }
    content := utils.ReadingLines(src)
    baseDir := path.Dir(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }

        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            utils.ErrorF("Error parse JSON Data")
            continue
        }
        raw := jsonParsed.S("OutputFile").Data().(string)
        reportPath := raw
        if !utils.FileExists(reportPath) {
            reportPath = path.Join(baseDir, raw)
            if !utils.FileExists(reportPath) {
                reportPath = path.Join(path.Dir(baseDir), raw)
            }
        }
        vulnContent := utils.GetFileContent(reportPath)

        // parse Data as JSON
        jsonParsed, err = gabs.ParseJSON([]byte(vulnContent))
        if err != nil {
            utils.ErrorF("Error parse JSON Data")
            continue
        }

        obj := database.Vulnerability{
            URL:                jsonParsed.S("URL").Data().(string),
            VulnRequest:        jsonParsed.S("Req").Data().(string),
            VulnResponse:       jsonParsed.S("Res").Data().(string),
            DetectionString:    jsonParsed.S("DetectionString").Data().(string),
            VulnerabilityTitle: jsonParsed.S("SignName").Data().(string),
            SignatureID:        jsonParsed.S("SignID").Data().(string),
            Confidence:         jsonParsed.S("Confidence").Data().(string),
            Severity:           jsonParsed.S("Risk").Data().(string),
            Source:             "Jaeles",
            //VulnChecksum:       utils.GenHash(vulnData),
            //PluginScan: SelectScanID(workspace, pluginName, scanID),
            //Target:     SelectScanByWS(workspace),
            //Asset:      SelectAssetByData(domain),
            ScanRefer:   r.ScanObj.ID,
            TargetRefer: r.ScanObj.TargetRefer,
        }
        obj.Create()
    }
}

// ImportNucleiVulnJson import new asset to DB
func (r *Runner) ImportNucleiVulnJson(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }
    content := utils.ReadingLines(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" {
            continue
        }
        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            utils.ErrorF("Error parse JSON Data")
            continue
        }

        obj := database.Vulnerability{
            URL:             jsonParsed.S("host").Data().(string),
            VulnRequest:     utils.Base64Encode(jsonParsed.S("request").Data().(string)),
            VulnResponse:    utils.Base64Encode(jsonParsed.S("response").Data().(string)),
            DetectionString: jsonParsed.S("matched").Data().(string),

            SignatureID:        jsonParsed.S("templateID").Data().(string),
            Confidence:         "Tentative",
            VulnerabilityTitle: jsonParsed.S("info", "name").Data().(string),
            Severity:           jsonParsed.S("info", "severity").Data().(string),
            Source:             "Nuclei",
            TargetRefer:        r.ScanObj.TargetRefer,

            ScanRefer: r.ScanObj.ID,
        }
        obj.Create()
    }
}

// ImportDirectoryJson import new asset to DB
func (r *Runner) ImportDirectoryJson(src string) {
    if !utils.FileExists(src) {
        utils.ErrorF("file not found: %v", src)
        return
    }
    content := utils.ReadingLines(src)
    for _, line := range content {
        if strings.TrimSpace(line) == "" || !strings.Contains(line, "url") {
            continue
        }

        jsonParsed, err := gabs.ParseJSON([]byte(line))
        if err != nil {
            utils.ErrorF("Error parse JSON Data")
            continue
        }

        // some parser here
        URL := jsonParsed.S("url").Data().(string)
        redirect, ok := jsonParsed.S("redirectlocation").Data().(string)
        if ok {
            redirect = ""
        }

        obj := database.Directory{
            URL:           URL,
            Status:        cast.ToInt(jsonParsed.S("status").Data()),
            ContentLength: cast.ToInt(jsonParsed.S("length").Data()),
            Words:         cast.ToInt(jsonParsed.S("words").Data()),
            RedirectURL:   redirect,
            ScanRefer:     r.ScanObj.ID,
            TargetRefer:   r.ScanObj.TargetRefer,
        }
        obj.Create()
    }
}

//
//// UpdateJSONDns update dns in db
//func UpdateJSONDns(src string, options libs.Options) {
//	content := utils.ReadingLines(src)
//	if len(content) == 0 {
//		utils.ErrorF("File not found: %v", src)
//		return
//	}
//
//	utils.DebugF("Update JSON DNS: %v", src)
//	//wsObj := database.SelectScanByWS(options.Scan.ROptions["Workspace"])
//	target := options.Scan.ROptions["Workspace"]
//
//	for _, line := range content {
//		jsonParsed, err := gabs.ParseJSON([]byte(line))
//		if err != nil {
//			continue
//		}
//
//		domain, ok := jsonParsed.S("host").Data().(string)
//		if !ok {
//			continue
//		}
//
//		// filtered some unrelated domain
//		if !strings.HasSuffix(domain, fmt.Sprintf(".%s", target)) {
//			if domain != target {
//				continue
//			}
//		}
//
//		data := "N/A"
//		isError := jsonParsed.S("status_code").Data().(string)
//		if isError == "NXDOMAIN" {
//			//database.NewAssetWithDns(wsObj, domain, data)
//			continue
//		}
//		var results []string
//		a := jsonParsed.S("a")
//		if a != nil {
//			for _, record := range a.Children() {
//				data := fmt.Sprintf("A/%s", cast.ToString(record.Data()))
//				results = append(results, data)
//			}
//		}
//
//		cname := jsonParsed.S("cname")
//		if cname != nil {
//			for _, record := range cname.Children() {
//				data := fmt.Sprintf("CNAME/%s", cast.ToString(record.Data()))
//				results = append(results, data)
//			}
//		}
//
//		mx := jsonParsed.S("mx")
//		if mx != nil {
//			for _, record := range mx.Children() {
//				data := fmt.Sprintf("MX/%s", cast.ToString(record.Data()))
//				results = append(results, data)
//			}
//		}
//
//		ns := jsonParsed.S("ns")
//		if ns != nil {
//			for _, record := range ns.Children() {
//				data := fmt.Sprintf("NS/%s", cast.ToString(record.Data()))
//				results = append(results, data)
//			}
//		}
//		results = funk.UniqString(results)
//		data = strings.Join(results, ";;")
//		//database.NewAssetWithDns(wsObj, domain, data)
//	}
//}
