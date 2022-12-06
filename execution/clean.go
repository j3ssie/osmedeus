package execution

import (
	"bufio"
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/spf13/cast"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thoas/go-funk"

	"github.com/Jeffail/gabs/v2"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
)

// Cleaning the execution directory
func Cleaning(folder string, reports []string) {
	utils.DebugF("Cleaning result: %v", folder)
	// list all the file
	items, err := filepath.Glob(fmt.Sprintf("%v/*", folder))
	if err != nil {
		return
	}

	for _, item := range items {
		item = utils.NormalizePath(item)
		utils.DebugF("Check Cleaning: %v", item)
		if funk.Contains(reports, item) {
			utils.DebugF("Skip cleaning file: %v", item)
			continue
		}

		fi, err := os.Stat(item)
		if err != nil {
			continue
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			DeleteFolder(item)
			continue
		case mode.IsRegular():
			DeleteFile(item)
		}
	}
}

// CleanGoBuster clean output for gobuster
func CleanGoBuster(src string, output string) {
	data := utils.GetFileContent(src)
	if data == "" {
		return
	}
	result := strings.Replace(data, "Found: ", "", -1)
	utils.WriteToFile(output, result)
}

// CleanMassdns clean result of massdns to get IP address
func CleanMassdns(filename string, output string) {
	file, err := os.Open(utils.NormalizePath(filename))
	if err != nil {
		return
	}
	defer file.Close()

	outputFile, err := os.OpenFile(utils.NormalizePath(output), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.Contains(line, " A ") {
			data := strings.Split(line, " A ")
			host := strings.Trim(data[0], ".")
			ip := strings.Trim(data[1], ".")
			outputFile.WriteString(fmt.Sprintf("%v,%v\n", host, ip))
		} else if strings.Contains(line, " CNAME ") {
			data := strings.Split(line, " CNAME ")
			host := strings.Trim(data[0], ".")
			ip := strings.Trim(data[1], ".")
			outputFile.WriteString(fmt.Sprintf("%v,%v\n", host, ip))
		}
	}
}

// CleanAmass get IP range and ASN from Amass result
func CleanAmass(filename string, output string) {
	content := utils.ReadingLines(filename)
	if len(content) <= 0 {
		return
	}

	var result []string
	for _, line := range content {
		jsonParsed, err := gabs.ParseJSON([]byte(line))
		if err != nil {
			continue
		}
		asn := jsonParsed.Path("asn").String()
		cidr := jsonParsed.Path("cidr").String()
		desc := jsonParsed.Path("desc").String()
		if cidr != "" && cidr != "null" {
			result = append(result, fmt.Sprintf("AS%v,%v,%v", asn, cidr, desc))
		}
	}

	if len(result) > 0 {
		utils.WriteToFile(output, strings.Join(result, "\n"))
	}

}

// CleanSWebanalyze get to get formatted report
func CleanSWebanalyze(filename string, output string) {
	content := utils.ReadingLines(filename)
	if len(content) <= 0 {
		return
	}

	result := make(map[string][]string)
	var finalResult []string
	for _, line := range content {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.Contains(line, "tech~") || !strings.Contains(line, "|") {
			continue
		}
		data := strings.Split(line, "|")
		domain, _ := utils.GetDomain(strings.TrimLeft(data[1], "~"))
		tech := data[2]
		if strings.HasSuffix(tech, "/") {
			tech = strings.Trim(tech, "/")
		}
		result[domain] = append(result[domain], tech)
	}

	// join summary result
	for k, v := range result {
		v = funk.UniqString(v)
		sort.Strings(v)
		techs := strings.Join(v, ",")
		data := fmt.Sprintf("domain|%v;;techs|%v", k, strings.Trim(techs, ","))
		finalResult = append(finalResult, data)
	}

	if len(result) > 0 {
		utils.WriteToFile(output, strings.Join(finalResult, "\n"))
	}
}

// CleanWebanalyze get to get formatted report
func CleanWebanalyze(filename string, output string, sum string) {
	content := utils.ReadingLines(filename)
	if len(content) <= 0 {
		return
	}

	result := make(map[string][]string)
	var finalResult []string
	var techSum []string
	for _, line := range content {
		jsonParsed, err := gabs.ParseJSON([]byte(line))
		if err != nil {
			continue
		}

		URL := strings.Trim(jsonParsed.S("hostname").String(), `"`)
		u, err := url.Parse(URL)
		if err != nil {
			continue
		}
		matches := jsonParsed.S("matches").Children()
		for _, child := range matches {
			techs := ""
			app := strings.Trim(child.S("app_name").String(), `"`)
			version := strings.Trim(child.S("version").String(), `"`)
			if version != "" {
				techs += fmt.Sprintf("%v/%v,", app, version)
			} else {
				techs += fmt.Sprintf("%v,", app)
			}

			// ignore blank techs
			if techs == "" || strings.Trim(techs, ",") == "" || strings.Trim(techs, ",") == " " {
				continue
			}
			techSum = append(techSum, strings.Trim(app, ","))
			techs = strings.TrimSpace(strings.Trim(techs, ","))
			result[u.Hostname()] = append(result[u.Hostname()], techs)
		}
	}

	// join summary result
	for k, v := range result {
		techs := strings.Join(funk.UniqString(v), ",")
		data := fmt.Sprintf("domain|%v;;techs|%v", k, strings.Trim(techs, ","))
		finalResult = append(finalResult, data)
	}

	if len(techSum) > 0 {
		utils.WriteToFile(sum, strings.Join(funk.UniqString(techSum), "\n"))
	}
	if len(result) > 0 {
		utils.WriteToFile(output, strings.Join(finalResult, "\n"))
	}
}

// CleanArjun clean output of Arjun
func CleanArjun(src string, dest string) {
	src = utils.NormalizePath(src)
	if !utils.FolderExists(src) {
		return
	}

	if !strings.HasSuffix(src, "/") {
		src += "/"
	}
	src = path.Join(src, "*")
	outputs, err := filepath.Glob(src)
	if err != nil {
		return
	}

	var data []string
	for _, output := range outputs {
		content := utils.GetFileContent(output)
		jsonParsed, err := gabs.ParseJSON([]byte(content))
		if err != nil {
			continue
		}
		prefix := "[GET]"
		if strings.HasPrefix(output, "post") {
			prefix = "[POST]"
		}

		for URL, child := range jsonParsed.ChildrenMap() {
			if len(child.Children()) <= 0 {
				continue
			}
			for _, query := range child.Children() {
				line := fmt.Sprintf("%v %v?%v=FUZZ", prefix, URL, query.Data())
				data = append(data, line)
			}
		}
	}

	if len(data) > 0 {
		utils.WriteToFile(dest, strings.Join(data, "\n"))
	}
}

// CleanJSONDnsx get to get formatted report
func CleanJSONDnsx(filename string, dest string) {
	content := utils.ReadingLines(filename)
	if len(content) <= 0 {
		utils.WarnF("File not found: %s", filename)
		return
	}

	var results []string
	for _, line := range content {
		jsonParsed, err := gabs.ParseJSON([]byte(line))
		if err != nil {
			continue
		}

		domain, ok := jsonParsed.S("host").Data().(string)
		if !ok {
			continue
		}

		a := jsonParsed.S("a")
		if a != nil {
			for _, record := range a.Children() {
				data := fmt.Sprintf("%s A %s", domain, cast.ToString(record.Data()))
				results = append(results, data)
			}
		}

		cname := jsonParsed.S("cname")
		if cname != nil {
			for _, record := range cname.Children() {
				data := fmt.Sprintf("%s CNAME %s", domain, cast.ToString(record.Data()))
				results = append(results, data)
			}
		}

		mx := jsonParsed.S("mx")
		if mx != nil {
			for _, record := range mx.Children() {
				data := fmt.Sprintf("%s MX %s", domain, cast.ToString(record.Data()))
				results = append(results, data)
			}
		}

		ns := jsonParsed.S("ns")
		if ns != nil {
			for _, record := range ns.Children() {
				data := fmt.Sprintf("%s NS %s", domain, cast.ToString(record.Data()))
				results = append(results, data)
			}
		}

	}

	if len(results) > 0 {
		utils.WriteToFile(dest, strings.Join(results, "\n"))
	}
}

// CleanRustScan make rustscan data to flat format ip:port
func CleanRustScan(src string, dest string) {
	src = utils.NormalizePath(src)
	dest = utils.NormalizePath(dest)
	content := utils.ReadingLines(src)
	if len(content) <= 0 {
		utils.WarnF("File not found: %s", src)
		return
	}

	var results []string
	for _, line := range content {
		// 103.247.207.76 -> [80,80,443,443]
		if !strings.Contains(line, " -> ") {
			continue
		}

		ip := strings.Split(line, " -> ")[0]
		rPorts := strings.Split(line, " -> ")[1]
		rPorts = rPorts[1 : len(rPorts)-1]

		if !strings.Contains(rPorts, ",") {
			results = append(results, fmt.Sprintf("%s:%s", ip, rPorts))
		}

		ports := strings.Split(rPorts, ",")
		ports = funk.UniqString(ports)
		for _, port := range ports {
			results = append(results, fmt.Sprintf("%s:%s", ip, port))
		}
	}
	utils.WriteToFile(dest, strings.Join(results, "\n"))
}

type Vulnerability struct {
	SignID     string
	SignPath   string
	URL        string
	Risk       string
	Confidence string
	Request    string
	ModalID    string

	ReportPath string
	ReportFile string

	Status string
	Length string
	Words  string
	Time   string
}

func GenNucleiReport(opt libs.Options, src string, dest string, templateFile string) {
	if templateFile == "" {
		templateFile = path.Join(opt.Env.DataFolder, "nuclei-report.html")
	}

	if !utils.FileExists(src) {
		utils.WarnF("file not found: %v", src)
		return
	}
	content := utils.ReadingLines(src)
	var vulns []Vulnerability
	for index, line := range content {
		if strings.TrimSpace(line) == "" {
			continue
		}
		jsonParsed, err := gabs.ParseJSON([]byte(line))
		if err != nil {
			utils.WarnF("Error parse JSON Data")
			continue
		}

		modalID, err := utils.GetDomain(cast.ToString(jsonParsed.S("host").Data()))
		if err != nil {
			continue
		}
		modalID = fmt.Sprintf("%s-%d", strings.ReplaceAll(modalID, ".", "-"), index)

		vulns = append(vulns, Vulnerability{
			Request:    cast.ToString(jsonParsed.S("request").Data()),
			URL:        cast.ToString(jsonParsed.S("matched-at").Data()),
			SignID:     cast.ToString(jsonParsed.S("template-id").Data()),
			Risk:       cast.ToString(jsonParsed.S("info", "severity").Data()),
			ModalID:    modalID,
			Confidence: "Tentative",
		})
	}

	utils.DebugF("Reading vuln %v from: %v ", len(vulns), src)
	if len(vulns) == 0 {
		utils.WarnF("No Vulnerability found %v", src)
		return
	}

	// read template file
	tmpl := utils.GetFileContent(templateFile)
	if strings.TrimSpace(tmpl) == "" {
		utils.WarnF("empty template data: %v", templateFile)
		return
	}

	variable := make(map[string]interface{})
	variable["Title"] = "Nuclei Summary Report"
	variable["Vulnerabilities"] = vulns
	variable["CurrentDay"] = utils.GetCurrentDay()
	variable["Version"] = libs.VERSION
	variable["Src"] = filepath.Base(src)

	tpl, err := pongo2.FromString(tmpl)
	if err != nil {
		utils.WarnF("error render data: %v", err)
		return
	}
	out, ok := tpl.Execute(variable)
	if ok == nil {
		utils.DebugF("Writing Nuclei HTML report to: %v", dest)
		utils.WriteToFile(dest, out)
	}
}

// CleanJSONHttpx get to get formatted report
func CleanJSONHttpx(filename string, dest string) {
	content := utils.ReadingLines(filename)
	if len(content) <= 0 {
		utils.WarnF("File not found: %s", filename)
		return
	}

	var results []string
	for _, line := range content {
		jsonParsed, err := gabs.ParseJSON([]byte(line))
		if err != nil {
			continue
		}

		URL := cast.ToString(jsonParsed.S("url").Data())
		//domain, _ := utils.GetDomain(URL)
		bodyHash := cast.ToString(jsonParsed.S("body-sha256").Data())
		headerHash := cast.ToString(jsonParsed.S("header-sha256").Data())
		hash := utils.GenHash(fmt.Sprintf("%s-%s", headerHash, bodyHash))

		title := "No-Title"
		rawTitle := jsonParsed.S("title")
		if rawTitle != nil {
			title = cast.ToString(rawTitle.Data())
		}

		rawTechs := jsonParsed.S("tech")
		techs := "No-Tech"
		if rawTechs != nil {
			techs = strings.Join(cast.ToStringSlice(rawTechs.Data()), ";;")

		}

		data := fmt.Sprintf("%s,%s,%s,%s", URL, hash, title, techs)
		results = append(results, data)

	}

	if len(results) > 0 {
		utils.WriteToFile(dest, strings.Join(results, "\n"))
	}
}

// CleanFFUFJson get to get formatted report
func CleanFFUFJson(filename string, dest string) {
	content := utils.ReadingLines(filename)
	if len(content) <= 0 {
		utils.WarnF("File not found: %s", filename)
		return
	}

	var results []string
	for _, line := range content {
		jsonParsed, err := gabs.ParseJSON([]byte(line))
		if err != nil {
			continue
		}

		resultsJson := jsonParsed.S("results")
		if resultsJson == nil {
			continue
		}

		for _, item := range resultsJson.Children() {
			//.url,.status,.length,.words,.lines,.redirectlocation
			endpoint := cast.ToString(item.S("url").Data())
			status := cast.ToString(item.S("status").Data())
			length := cast.ToString(item.S("length").Data())
			words := cast.ToString(item.S("words").Data())
			lines := cast.ToString(item.S("lines").Data())
			redirectLocation := cast.ToString(item.S("redirectlocation").Data())

			data := fmt.Sprintf("%s,%s,%s,%s,%s,%s", endpoint, status, length, words, lines, redirectLocation)
			results = append(results, data)
		}
	}

	if len(results) > 0 {
		utils.WriteToFile(dest, strings.Join(results, "\n"))
	}
}
