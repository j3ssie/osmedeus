package core

import (
	"path"
	"text/template"

	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/j3ssie/osmedeus/libs"
	"github.com/j3ssie/osmedeus/utils"
	"github.com/spf13/cast"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func (r *Runner) GenMarkdownReport(markdownFile string, outputHTML string) {
	utils.DebugF("Reading markdown report from: %v", markdownFile)

	// get the markdown template content
	mdContent := utils.GetFileContent(markdownFile)
	mdContent = ResolveData(mdContent, r.Target)

	// replace all the <scanInfo /> tag
	mdContent = r.ResolveScanInfoTag(mdContent)
	// utils.DebugF("ResolveScanInfoTag:\n%v", mdContent)

	// replace all the <reports /> tag
	mdContent = r.ResolveReportsTag(mdContent)

	// replace all the <content /> tag
	mdContent = r.ResolveContentTag(mdContent)
	// fmt.Println("mdContent", mdContent)

	// generating the markdown file first
	outputMD := strings.Replace(outputHTML, ".html", ".md", -1)
	utils.WriteToFile(outputMD, mdContent)

	utils.InforF("Generate markdown report: %v", outputMD)
	utils.InforF("Generate HTML report: %v", outputHTML)
	// finally convert to HTML
	MarkDownToHTML(r.Opt, r.Input, outputMD, outputHTML)
}

func (r *Runner) ResolveScanInfoTag(rawMarkdown string) string {
	re := regexp.MustCompile(`<scanInfo\s*/>`)
	match := re.FindString(rawMarkdown)
	if len(match) > 1 {
		utils.DebugF("Replace scanInfo tag: %v", match)
		scanInfo := fmt.Sprintf(`
| <!-- -->       | <!-- -->    |
|----------------|-------------|
| Target         | **:target**     |
| Running Time   | **:runningTime**         |
| Workflow       | **:workflow**         |
| Status         | **:status**         |
| Statistics     | :statistics         |
`)

		status := "done"
		if r.ScanObj.IsRunning {
			status = "running"
		}

		statistics := fmt.Sprintf("`assets/%v`, `dns/%v`, `vulnerability/%v`, ", r.TargetObj.TotalAssets, r.TargetObj.TotalDns, r.TargetObj.TotalVulnerability)
		replacements := map[string]string{
			":target":      r.ScanObj.InputName,
			":runningTime": cast.ToString(int(r.RunningTime)/3600) + " hours",
			":workflow":    r.ScanObj.TaskName,
			":statistics":  statistics,
			":status":      status,
		}
		// generate the statistics info
		for oldStr, newStr := range replacements {
			scanInfo = strings.ReplaceAll(scanInfo, oldStr, newStr)
		}

		return strings.Replace(rawMarkdown, match, scanInfo, -1)
	}
	utils.DebugF("No scanInfo tag found")
	return rawMarkdown
}

func (r *Runner) ResolveContentTag(rawData string) string {
	finalMarkdown := rawData
	// finding all the content tags and replace it with the content
	re := regexp.MustCompile(`<content[^>]*>`)
	matchs := re.FindAllString(rawData, -1)
	for _, contentTag := range matchs {
		utils.DebugF("Replace content tag: %v", color.GreenString(contentTag))
		content := r.ResolveContentSrc(contentTag)
		finalMarkdown = strings.Replace(finalMarkdown, contentTag, content, -1)
	}

	return finalMarkdown
}

func (r *Runner) ResolveContentSrc(tag string) string {
	re := regexp.MustCompile(`src=\"(\S+)\"`)
	match := re.FindStringSubmatch(tag)
	if len(match) > 1 {
		fileContent := utils.GetFileContent(match[1])
		utils.DebugF("Replace content src: %v", color.GreenString(match[1]))

		if strings.Contains(tag, "expand=true") {
			return "```\n" + fileContent + "```"
		}

		if strings.Contains(tag, "shorten=true") || len(fileContent) > r.Opt.MDCodeBlockLimit {
			fileContent = template.HTMLEscapeString(fileContent) // sanitize file content to prevent XSS
			return extendTag(fileContent)
		}

		return "```\n" + fileContent + "```"
	}
	return ""
}

func extendTag(str string) string {
	data := "<details>\n<summary>Click to Expand</summary>\n\n" + "<pre>\n" + str + "\n</pre>" + "\n</details>"
	return data
}

func (r *Runner) ResolveReportsTag(rawMarkDown string) string {
	// rawData := utils.GetFileContent(markdownFile)
	finalMarkdown := rawMarkDown
	// finding all the reports tags
	re := regexp.MustCompile(`<reports\s*/>`)
	matchs := re.FindAllString(rawMarkDown, -1)
	if len(matchs) == 0 {
		utils.DebugF("No reports tag found")
		return rawMarkDown
	}

	for _, reportTag := range matchs {
		utils.DebugF("Replace content tag: %v", reportTag)
		mdContent := ""

		for _, report := range r.TargetObj.Reports {
			// add the link if report file is HTML file
			if report.ReportType == "html" {
				mdContent += fmt.Sprintf("### %s -- [%s](%s) \n\n", report.Module, report.ReportName, report.ReportPath)
				mdContent += "\n***\n"
				continue
			}
			// add the full content if report file is a text file
			mdContent += fmt.Sprintf("### %s -- *%s* \n\n", report.Module, report.ReportName)

			fileContent := utils.GetFileContent(report.ReportPath)
			if len(fileContent) > r.Opt.MDCodeBlockLimit {
				mdContent += extendTag(fileContent)
			} else {
				mdContent += "```\n"
				mdContent += fileContent
				mdContent += "\n```\n"
			}
			mdContent += "\n***\n\n"
		}

		finalMarkdown = strings.Replace(finalMarkdown, reportTag, mdContent, -1)
	}

	return finalMarkdown
}

func MarkDownToHTML(options libs.Options, target string, markdownFile string, outputFile string) error {
	css := path.Join(options.Env.DataFolder, "markdown/style.css")

	var input []byte
	var err error

	if input, err = os.ReadFile(markdownFile); err != nil {
		utils.ErrorF("Error reading %s: %v", markdownFile, err)
		return err
	}

	// set up options
	var extensions = parser.NoIntraEmphasis |
		parser.Tables |
		parser.FencedCode |
		parser.Autolink |
		parser.Strikethrough |
		parser.SpaceHeadings

	var renderer markdown.Renderer
	// render the data into HTML
	var htmlFlags html.Flags

	htmlFlags |= html.Smartypants
	htmlFlags |= html.UseXHTML
	htmlFlags |= html.CompletePage
	htmlFlags |= html.SmartypantsLatexDashes
	htmlFlags |= html.SmartypantsFractions

	params := html.RendererOptions{
		Flags: htmlFlags,
		CSS:   css,
	}
	renderer = html.NewRenderer(params)

	// parse and render
	var output []byte
	parser := parser.NewWithExtensions(extensions)
	// @NOTE: beware of XSS as I assume you will trust the markdown content you generate
	output = markdown.ToHTML(input, parser, renderer)
	// html := bluemonday.UGCPolicy().SanitizeBytes(output) // skip the sanitization as we prefer more beautiful output

	cssContent, err := os.ReadFile(css)
	if err != nil {
		return err
	}

	finalHTML := fmt.Sprintf(`
<html>
	<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width">
	<link rel="icon" type="image/svg+xml" href="https://www.osmedeus.org/favicon.png">
	<meta name="generator" content="Generated by https://github.com/gomarkdown/markdown">
	<meta name="description" content="A Workflow Engine for Offensive Security">
	<title>Osmedeus Executive Summary - %s</title>
	<meta property="og:title" content="Osmedeus Next Generation">
	<meta property="og:type" content="website">
	<meta property="og:url" content="https://www.osmedeus.org">
	<meta property="og:image" content="https://raw.githubusercontent.com/osmedeus/assets/main/banner.png">
	<meta property="og:description" content="A Workflow Engine for Offensive Security">
	<meta name="twitter:card" content="">
	<meta name="twitter:site" content="@OsmedeusEngine">
	<meta name="twitter:creator" content="@OsmedeusEngine">
	<style>%s</style>
	</head>

<body>
%s
</body>
</html>
	`, target, string(cssContent), string(output))

	// output the result
	var out *os.File
	if out, err = os.Create(outputFile); err != nil {
		utils.ErrorF("Error creating %s: %v", outputFile, err)
		return err
	}
	defer out.Close()

	if _, err = out.WriteString(finalHTML); err != nil {
		utils.ErrorF("Error writing output: %v", err)
		return err
	}

	return nil
}
