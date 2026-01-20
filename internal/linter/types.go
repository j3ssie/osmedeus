package linter

import (
	"github.com/goccy/go-yaml/ast"
	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// Severity represents the severity level of a lint issue
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

// String returns the string representation of the severity
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// ParseSeverity parses a string into a Severity
func ParseSeverity(s string) Severity {
	switch s {
	case "info":
		return SeverityInfo
	case "warning":
		return SeverityWarning
	case "error":
		return SeverityError
	default:
		return SeverityWarning
	}
}

// LintIssue represents a single lint issue found in a workflow
type LintIssue struct {
	Rule       string   // Rule name (e.g., "unused-variable")
	Severity   Severity // Issue severity level
	Message    string   // Human-readable description of the issue
	Suggestion string   // Optional fix suggestion
	Line       int      // 1-based line number
	Column     int      // 1-based column number
	Field      string   // YAML path (e.g., "steps[0].bash")
}

// LinterRule is the interface that all lint rules must implement
type LinterRule interface {
	// Name returns the unique identifier for this rule
	Name() string
	// Description returns a human-readable description of what this rule checks
	Description() string
	// Severity returns the default severity level for issues from this rule
	Severity() Severity
	// Check performs the lint check and returns any issues found
	Check(ast *WorkflowAST) []LintIssue
}

// WorkflowAST holds parsed workflow with line/column information
type WorkflowAST struct {
	// Workflow is the parsed workflow struct
	Workflow *core.Workflow
	// FilePath is the path to the source file
	FilePath string
	// Source is the raw YAML content
	Source []byte
	// Root is the raw YAML AST node
	Root ast.Node
	// NodeMap maps YAML paths to AST nodes for line tracking
	NodeMap map[string]ast.Node
}

// LintResult holds the complete result of linting a workflow
type LintResult struct {
	FilePath string
	Issues   []LintIssue
	Errors   int
	Warnings int
	Infos    int
}

// HasErrors returns true if there are any error-level issues
func (r *LintResult) HasErrors() bool {
	return r.Errors > 0
}

// HasIssues returns true if there are any issues at all
func (r *LintResult) HasIssues() bool {
	return len(r.Issues) > 0
}

// Summary returns a summary string of the lint result
func (r *LintResult) Summary() string {
	if !r.HasIssues() {
		return "No issues found"
	}
	return ""
}

// OutputFormat specifies the output format for lint results
type OutputFormat string

const (
	FormatPretty OutputFormat = "pretty" // Colored terminal output with context
	FormatJSON   OutputFormat = "json"   // Machine-readable JSON
	FormatGitHub OutputFormat = "github" // GitHub Actions annotations
)

// ParseOutputFormat parses a string into an OutputFormat
func ParseOutputFormat(s string) OutputFormat {
	switch s {
	case "json":
		return FormatJSON
	case "github":
		return FormatGitHub
	default:
		return FormatPretty
	}
}
