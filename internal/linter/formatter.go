package linter

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// Formatter formats lint results for output
type Formatter interface {
	// Format formats the lint result for a single file
	Format(result *LintResult, source []byte) string
	// FormatSummary formats a summary of multiple results
	FormatSummary(results []*LintResult) string
}

// PrettyFormatter provides colored terminal output with source context
type PrettyFormatter struct {
	ShowContext bool   // Show source line with issue
	NoColor     bool   // Disable colored output
	BaseDir     string // Base directory for relative paths
}

// NewPrettyFormatter creates a new pretty formatter
func NewPrettyFormatter(showContext bool) *PrettyFormatter {
	return &PrettyFormatter{
		ShowContext: showContext,
	}
}

// Format formats lint issues with colored output and source context
func (f *PrettyFormatter) Format(result *LintResult, source []byte) string {
	if len(result.Issues) == 0 {
		return ""
	}

	var sb strings.Builder
	lines := splitLines(source)
	displayPath := f.getDisplayPath(result.FilePath)

	for _, issue := range result.Issues {
		// Header: path:line:col: severity[rule]: message
		severityStr := f.colorSeverity(issue.Severity)
		sb.WriteString(fmt.Sprintf("%s:%d:%d: %s[%s]: %s\n",
			displayPath, issue.Line, issue.Column,
			severityStr, issue.Rule, issue.Message))

		// Source context
		if f.ShowContext && issue.Line > 0 && issue.Line <= len(lines) {
			sourceLine := lines[issue.Line-1]
			sb.WriteString(fmt.Sprintf("   %d | %s\n", issue.Line, sourceLine))

			// Pointer to the issue position
			if issue.Column > 0 {
				padding := len(fmt.Sprintf("   %d | ", issue.Line))
				pointer := strings.Repeat(" ", padding+issue.Column-1) + f.colorPointer("^")
				sb.WriteString(pointer + "\n")
			}
		}

		// Suggestion
		if issue.Suggestion != "" {
			sb.WriteString(f.colorSuggestion("   Suggestion: "+issue.Suggestion) + "\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatSummary formats a summary of lint results
func (f *PrettyFormatter) FormatSummary(results []*LintResult) string {
	totalErrors := TotalErrors(results)
	totalWarnings := TotalWarnings(results)
	totalInfos := TotalInfos(results)
	filesWithIssues := 0
	for _, r := range results {
		if r.HasIssues() {
			filesWithIssues++
		}
	}

	if totalErrors == 0 && totalWarnings == 0 && totalInfos == 0 {
		return f.colorSuccess("No issues found")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %s, %s, %s in %d file(s)\n",
		f.colorError(fmt.Sprintf("%d error(s)", totalErrors)),
		f.colorWarning(fmt.Sprintf("%d warning(s)", totalWarnings)),
		f.colorInfo(fmt.Sprintf("%d info(s)", totalInfos)),
		filesWithIssues))
	sb.WriteString(f.colorMuted("\nNote: The linter shows best practices for writing workflows. You can still execute workflows normally even with linter warnings."))
	return sb.String()
}

func (f *PrettyFormatter) getDisplayPath(path string) string {
	if f.BaseDir != "" {
		if rel, err := filepath.Rel(f.BaseDir, path); err == nil {
			return rel
		}
	}
	return path
}

func (f *PrettyFormatter) colorSeverity(s Severity) string {
	if f.NoColor {
		return s.String()
	}
	switch s {
	case SeverityError:
		return "\033[31m" + s.String() + "\033[0m" // Red
	case SeverityWarning:
		return "\033[33m" + s.String() + "\033[0m" // Yellow
	case SeverityInfo:
		return "\033[36m" + s.String() + "\033[0m" // Cyan
	default:
		return s.String()
	}
}

func (f *PrettyFormatter) colorPointer(s string) string {
	if f.NoColor {
		return s
	}
	return "\033[31m" + s + "\033[0m" // Red
}

func (f *PrettyFormatter) colorSuggestion(s string) string {
	if f.NoColor {
		return s
	}
	return "\033[90m" + s + "\033[0m" // Gray
}

func (f *PrettyFormatter) colorError(s string) string {
	if f.NoColor {
		return s
	}
	return "\033[31m" + s + "\033[0m" // Red
}

func (f *PrettyFormatter) colorWarning(s string) string {
	if f.NoColor {
		return s
	}
	return "\033[33m" + s + "\033[0m" // Yellow
}

func (f *PrettyFormatter) colorSuccess(s string) string {
	if f.NoColor {
		return s
	}
	return "\033[32m" + s + "\033[0m" // Green
}

func (f *PrettyFormatter) colorInfo(s string) string {
	if f.NoColor {
		return s
	}
	return "\033[36m" + s + "\033[0m" // Cyan
}

func (f *PrettyFormatter) colorMuted(s string) string {
	if f.NoColor {
		return s
	}
	return "\033[90m" + s + "\033[0m" // Gray
}

// JSONFormatter provides machine-readable JSON output
type JSONFormatter struct{}

// JSONOutput represents the JSON output structure
type JSONOutput struct {
	File    string      `json:"file"`
	Issues  []JSONIssue `json:"issues"`
	Summary JSONSummary `json:"summary"`
}

// JSONIssue represents a single issue in JSON format
type JSONIssue struct {
	Rule       string `json:"rule"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Field      string `json:"field,omitempty"`
}

// JSONSummary represents the summary in JSON format
type JSONSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Info     int `json:"info"`
}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format formats lint result as JSON
func (f *JSONFormatter) Format(result *LintResult, _ []byte) string {
	output := JSONOutput{
		File:   result.FilePath,
		Issues: make([]JSONIssue, len(result.Issues)),
		Summary: JSONSummary{
			Errors:   result.Errors,
			Warnings: result.Warnings,
			Info:     result.Infos,
		},
	}

	for i, issue := range result.Issues {
		output.Issues[i] = JSONIssue{
			Rule:       issue.Rule,
			Severity:   issue.Severity.String(),
			Message:    issue.Message,
			Suggestion: issue.Suggestion,
			Line:       issue.Line,
			Column:     issue.Column,
			Field:      issue.Field,
		}
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(data)
}

// FormatSummary formats a combined summary as JSON
func (f *JSONFormatter) FormatSummary(results []*LintResult) string {
	type combinedOutput struct {
		TotalErrors   int          `json:"total_errors"`
		TotalWarnings int          `json:"total_warnings"`
		TotalFiles    int          `json:"total_files"`
		Files         []JSONOutput `json:"files"`
	}

	combined := combinedOutput{
		TotalErrors:   TotalErrors(results),
		TotalWarnings: TotalWarnings(results),
		TotalFiles:    len(results),
		Files:         make([]JSONOutput, len(results)),
	}

	for i, result := range results {
		issues := make([]JSONIssue, len(result.Issues))
		for j, issue := range result.Issues {
			issues[j] = JSONIssue{
				Rule:       issue.Rule,
				Severity:   issue.Severity.String(),
				Message:    issue.Message,
				Suggestion: issue.Suggestion,
				Line:       issue.Line,
				Column:     issue.Column,
				Field:      issue.Field,
			}
		}
		combined.Files[i] = JSONOutput{
			File:   result.FilePath,
			Issues: issues,
			Summary: JSONSummary{
				Errors:   result.Errors,
				Warnings: result.Warnings,
				Info:     result.Infos,
			},
		}
	}

	data, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error())
	}
	return string(data)
}

// GitHubFormatter provides GitHub Actions annotation format
type GitHubFormatter struct {
	BaseDir string
}

// NewGitHubFormatter creates a new GitHub Actions formatter
func NewGitHubFormatter() *GitHubFormatter {
	return &GitHubFormatter{}
}

// Format formats lint result as GitHub Actions annotations
func (f *GitHubFormatter) Format(result *LintResult, _ []byte) string {
	if len(result.Issues) == 0 {
		return ""
	}

	var sb strings.Builder
	displayPath := f.getDisplayPath(result.FilePath)

	for _, issue := range result.Issues {
		// GitHub annotation format: ::severity file=path,line=N,col=N::message
		level := f.severityToGitHub(issue.Severity)
		message := issue.Message
		if issue.Suggestion != "" {
			message += " Suggestion: " + issue.Suggestion
		}

		sb.WriteString(fmt.Sprintf("::%s file=%s,line=%d,col=%d::[%s] %s\n",
			level, displayPath, issue.Line, issue.Column, issue.Rule, message))
	}

	return sb.String()
}

// FormatSummary formats a summary (GitHub format doesn't have a special summary)
func (f *GitHubFormatter) FormatSummary(results []*LintResult) string {
	totalErrors := TotalErrors(results)
	totalWarnings := TotalWarnings(results)

	if totalErrors == 0 && totalWarnings == 0 {
		return "::notice::Workflow linting passed with no issues"
	}

	return fmt.Sprintf("::notice::Found %d error(s), %d warning(s)", totalErrors, totalWarnings)
}

func (f *GitHubFormatter) getDisplayPath(path string) string {
	if f.BaseDir != "" {
		if rel, err := filepath.Rel(f.BaseDir, path); err == nil {
			return rel
		}
	}
	return path
}

func (f *GitHubFormatter) severityToGitHub(s Severity) string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	default:
		return "notice"
	}
}

// GetFormatter returns the appropriate formatter for the given format
func GetFormatter(format OutputFormat, showContext bool) Formatter {
	switch format {
	case FormatJSON:
		return NewJSONFormatter()
	case FormatGitHub:
		return NewGitHubFormatter()
	default:
		return NewPrettyFormatter(showContext)
	}
}
