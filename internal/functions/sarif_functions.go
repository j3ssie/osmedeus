package functions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/database"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// SARIF document types for JSON unmarshalling

type sarifDocument struct {
	Runs []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name  string      `json:"name"`
	Rules []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	ShortDescription sarifMessage       `json:"shortDescription"`
	FullDescription  sarifMessage       `json:"fullDescription"`
	Help             sarifMessage       `json:"help"`
	DefaultConfig    sarifDefaultConfig `json:"defaultConfiguration"`
	Properties       sarifRuleProps     `json:"properties"`
}

type sarifDefaultConfig struct {
	Level string `json:"level"`
}

type sarifRuleProps struct {
	Tags []string `json:"tags"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int          `json:"startLine"`
	EndLine     int          `json:"endLine"`
	StartColumn int          `json:"startColumn"`
	EndColumn   int          `json:"endColumn"`
	Snippet     sarifSnippet `json:"snippet"`
}

type sarifSnippet struct {
	Text string `json:"text"`
}

// parseSARIFFile reads and parses a SARIF JSON file
func parseSARIFFile(filePath string) (*sarifDocument, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var doc sarifDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse SARIF JSON: %w", err)
	}

	return &doc, nil
}

// mapSARIFLevel maps SARIF level to osmedeus severity
func mapSARIFLevel(level string) string {
	switch strings.ToLower(level) {
	case "error":
		return "high"
	case "warning":
		return "medium"
	case "note":
		return "low"
	default:
		return "info"
	}
}

// severityOrder returns sort priority (lower = more severe)
func severityOrder(sev string) int {
	switch sev {
	case "critical":
		return 0
	case "high":
		return 1
	case "medium":
		return 2
	case "low":
		return 3
	default:
		return 4
	}
}

// buildRuleMap creates a ruleID -> sarifRule lookup from all runs
func buildRuleMap(runs []sarifRun) map[string]sarifRule {
	m := make(map[string]sarifRule)
	for _, run := range runs {
		for _, rule := range run.Tool.Driver.Rules {
			m[rule.ID] = rule
		}
	}
	return m
}

// resolveTitle returns the best title for a result given its rule
func resolveTitle(ruleID string, rule sarifRule) string {
	if rule.Name != "" {
		return rule.Name
	}
	if rule.ShortDescription.Text != "" {
		return rule.ShortDescription.Text
	}
	return ruleID
}

// resolveSeverity returns the severity for a result, using result level then rule default
func resolveSeverity(result sarifResult, rule sarifRule) string {
	if result.Level != "" {
		return mapSARIFLevel(result.Level)
	}
	if rule.DefaultConfig.Level != "" {
		return mapSARIFLevel(rule.DefaultConfig.Level)
	}
	return "info"
}

// resolveLocation returns URI and location string for a result
func resolveLocation(result sarifResult) (uri string, location string) {
	if len(result.Locations) == 0 {
		return "", ""
	}
	loc := result.Locations[0]
	uri = loc.PhysicalLocation.ArtifactLocation.URI
	if loc.PhysicalLocation.Region.StartLine > 0 {
		location = fmt.Sprintf("%s:%d", uri, loc.PhysicalLocation.Region.StartLine)
	} else {
		location = uri
	}
	return uri, location
}

// extractCWETags extracts CWE-related tags from rule properties
func extractCWETags(rule sarifRule) []string {
	var tags []string
	for _, tag := range rule.Properties.Tags {
		tags = append(tags, tag)
	}
	return tags
}

// dbImportSARIF imports vulnerabilities from a SARIF file into the database
// Usage: db_import_sarif(workspace, file_path) -> map (stats)
func (vf *vmFunc) dbImportSARIF(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("dbImportSARIF"))

	if len(call.Arguments) < 2 {
		return vf.errorValue("db_import_sarif requires 2 arguments: workspace, file_path")
	}

	workspace := call.Argument(0).String()
	filePath := call.Argument(1).String()

	if workspace == "" || workspace == "undefined" {
		return vf.errorValue("workspace cannot be empty")
	}

	if filePath == "" || filePath == "undefined" {
		return vf.errorValue("file_path cannot be empty")
	}

	db := database.GetDB()
	if db == nil {
		return vf.errorValue("database not connected")
	}

	doc, err := parseSARIFFile(filePath)
	if err != nil {
		return vf.errorValue(fmt.Sprintf("failed to parse SARIF file: %v", err))
	}

	ruleMap := buildRuleMap(doc.Runs)
	ctx := context.Background()
	stats := database.ImportStats{}
	now := time.Now()

	for _, run := range doc.Runs {
		for _, result := range run.Results {
			rule := ruleMap[result.RuleID]
			uri, _ := resolveLocation(result)
			if uri == "" {
				uri = result.RuleID
			}

			severity := resolveSeverity(result, rule)
			title := resolveTitle(result.RuleID, rule)
			tags := extractCWETags(rule)

			// Get snippet for POC
			var snippet string
			if len(result.Locations) > 0 {
				snippet = result.Locations[0].PhysicalLocation.Region.Snippet.Text
			}

			// Serialize result to JSON for raw storage
			rawJSON, _ := json.Marshal(result)

			vuln := database.Vulnerability{
				Workspace:   workspace,
				VulnInfo:    result.RuleID,
				VulnTitle:   title,
				VulnDesc:    result.Message.Text,
				VulnPOC:     snippet,
				Severity:    severity,
				Confidence:  "Firm",
				AssetType:   "repo",
				AssetValue:  uri,
				Tags:        tags,
				RawVulnJSON: string(rawJSON),
				LastSeenAt:  now,
			}

			// Upsert: match on workspace + vuln_info + asset_value
			var existing database.Vulnerability
			selectErr := db.NewSelect().Model(&existing).
				Where("workspace = ?", workspace).
				Where("vuln_info = ?", vuln.VulnInfo).
				Where("asset_value = ?", vuln.AssetValue).
				Scan(ctx)

			if selectErr != nil {
				// New vulnerability - insert
				vuln.CreatedAt = now
				vuln.UpdatedAt = now
				_, insertErr := db.NewInsert().Model(&vuln).Exec(ctx)
				if insertErr != nil {
					logger.Get().Debug("failed to insert SARIF vulnerability", zap.Error(insertErr))
					stats.Errors++
					continue
				}
				stats.New++
			} else if hasVulnChanged(&existing, &vuln) {
				// Changed - full update
				vuln.ID = existing.ID
				vuln.CreatedAt = existing.CreatedAt
				vuln.UpdatedAt = now
				_, updateErr := db.NewUpdate().Model(&vuln).WherePK().Exec(ctx)
				if updateErr != nil {
					logger.Get().Debug("failed to update SARIF vulnerability", zap.Error(updateErr))
					stats.Errors++
					continue
				}
				stats.Updated++
			} else {
				// Unchanged - only update last_seen_at
				_, updateErr := db.NewUpdate().Model((*database.Vulnerability)(nil)).
					Set("last_seen_at = ?", now).
					Where("id = ?", existing.ID).
					Exec(ctx)
				if updateErr != nil {
					logger.Get().Debug("failed to update last_seen_at", zap.Error(updateErr))
					stats.Errors++
					continue
				}
				stats.Unchanged++
			}
		}
	}

	total := stats.New + stats.Updated + stats.Unchanged
	logger.Get().Debug("dbImportSARIF completed",
		zap.String("workspace", workspace),
		zap.String("file", filePath),
		zap.Int("new", stats.New),
		zap.Int("updated", stats.Updated),
		zap.Int("unchanged", stats.Unchanged),
		zap.Int("errors", stats.Errors),
		zap.Int("total", total))

	return vf.vm.ToValue(map[string]interface{}{
		"new":       stats.New,
		"updated":   stats.Updated,
		"unchanged": stats.Unchanged,
		"errors":    stats.Errors,
		"total":     total,
	})
}

// sarifFinding is an intermediate struct for markdown conversion
type sarifFinding struct {
	Severity    string
	Location    string
	Title       string
	Description string
}

// convertSARIFToMarkdown parses a SARIF file and writes a markdown report
// Usage: convert_sarif_to_markdown(input_path, output_path) -> bool
func (vf *vmFunc) convertSARIFToMarkdown(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		logger.Get().Warn("convertSARIFToMarkdown: requires 2 arguments (input_path, output_path)")
		return vf.vm.ToValue(false)
	}

	inputPath := call.Argument(0).String()
	outputPath := call.Argument(1).String()
	logger.Get().Debug("Calling convertSARIFToMarkdown", zap.String("input", inputPath), zap.String("output", outputPath))

	if inputPath == "undefined" || inputPath == "" {
		logger.Get().Warn("convertSARIFToMarkdown: empty input path provided")
		return vf.vm.ToValue(false)
	}

	if outputPath == "undefined" || outputPath == "" {
		logger.Get().Warn("convertSARIFToMarkdown: empty output path provided")
		return vf.vm.ToValue(false)
	}

	doc, err := parseSARIFFile(inputPath)
	if err != nil {
		logger.Get().Warn("convertSARIFToMarkdown: failed to parse SARIF", zap.String("path", inputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	ruleMap := buildRuleMap(doc.Runs)

	// Determine tool name
	toolName := "Unknown"
	for _, run := range doc.Runs {
		if run.Tool.Driver.Name != "" {
			toolName = run.Tool.Driver.Name
			break
		}
	}

	// Collect findings
	var findings []sarifFinding
	severityCounts := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	for _, run := range doc.Runs {
		for _, result := range run.Results {
			rule := ruleMap[result.RuleID]
			_, location := resolveLocation(result)
			if location == "" {
				location = result.RuleID
			}

			severity := resolveSeverity(result, rule)
			title := resolveTitle(result.RuleID, rule)
			desc := result.Message.Text

			// Truncate description to 200 chars for table readability
			if len(desc) > 200 {
				desc = desc[:200] + "..."
			}

			// Escape pipe characters for markdown table
			desc = strings.ReplaceAll(desc, "|", "\\|")
			desc = strings.ReplaceAll(desc, "\n", " ")
			title = strings.ReplaceAll(title, "|", "\\|")
			location = strings.ReplaceAll(location, "|", "\\|")

			findings = append(findings, sarifFinding{
				Severity:    severity,
				Location:    location,
				Title:       title,
				Description: desc,
			})

			severityCounts[severity]++
		}
	}

	// Sort by severity (high -> medium -> low -> info)
	sort.Slice(findings, func(i, j int) bool {
		return severityOrder(findings[i].Severity) < severityOrder(findings[j].Severity)
	})

	// Build markdown
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# SARIF Results — %s\n\n", toolName))

	total := len(findings)
	sb.WriteString(fmt.Sprintf("**Total: %d findings** | Critical: %d | High: %d | Medium: %d | Low: %d\n\n",
		total, severityCounts["critical"], severityCounts["high"], severityCounts["medium"], severityCounts["low"]))

	if total > 0 {
		sb.WriteString("| Severity | Location | Title | Description |\n")
		sb.WriteString("| --- | --- | --- | --- |\n")

		for _, f := range findings {
			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n",
				f.Severity, f.Location, f.Title, f.Description))
		}
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		logger.Get().Warn("convertSARIFToMarkdown: failed to create output directory",
			zap.String("path", outputDir), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	// Write to output file
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		logger.Get().Warn("convertSARIFToMarkdown: failed to write output",
			zap.String("path", outputPath), zap.Error(err))
		return vf.vm.ToValue(false)
	}

	logger.Get().Debug("convertSARIFToMarkdown result",
		zap.String("input", inputPath),
		zap.String("output", outputPath),
		zap.Int("findings", total))
	return vf.vm.ToValue(true)
}
