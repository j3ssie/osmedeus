package linter

import (
	"sort"
)

// LinterOptions configures the linter behavior
type LinterOptions struct {
	DisabledRules []string // Rule names to skip
	MinSeverity   Severity // Only report issues >= this severity
}

// DefaultOptions returns the default linter options
func DefaultOptions() LinterOptions {
	return LinterOptions{
		DisabledRules: nil,
		MinSeverity:   SeverityInfo,
	}
}

// Linter is the main workflow linting engine
type Linter struct {
	rules   []LinterRule
	options LinterOptions
}

// NewLinter creates a new linter with the given options
func NewLinter(opts LinterOptions) *Linter {
	l := &Linter{
		rules:   GetDefaultRules(),
		options: opts,
	}
	return l
}

// NewDefaultLinter creates a linter with default options and all built-in rules
func NewDefaultLinter() *Linter {
	return NewLinter(DefaultOptions())
}

// RegisterRule adds a custom rule to the linter
func (l *Linter) RegisterRule(rule LinterRule) {
	l.rules = append(l.rules, rule)
}

// SetRules replaces all rules with the given set
func (l *Linter) SetRules(rules []LinterRule) {
	l.rules = rules
}

// GetRules returns all registered rules
func (l *Linter) GetRules() []LinterRule {
	return l.rules
}

// isRuleDisabled checks if a rule is in the disabled list
func (l *Linter) isRuleDisabled(ruleName string) bool {
	for _, disabled := range l.options.DisabledRules {
		if disabled == ruleName {
			return true
		}
	}
	return false
}

// Lint lints a workflow file and returns the result
func (l *Linter) Lint(path string) (*LintResult, error) {
	ast, err := ParseWorkflowAST(path)
	if err != nil {
		return nil, err
	}

	return l.LintWorkflow(ast), nil
}

// LintContent lints workflow content from bytes
func (l *Linter) LintContent(content []byte, filename string) (*LintResult, error) {
	ast, err := ParseWorkflowASTFromContent(content, filename)
	if err != nil {
		return nil, err
	}

	return l.LintWorkflow(ast), nil
}

// LintWorkflow lints a pre-parsed workflow AST
func (l *Linter) LintWorkflow(ast *WorkflowAST) *LintResult {
	var allIssues []LintIssue

	// Run all enabled rules
	for _, rule := range l.rules {
		if l.isRuleDisabled(rule.Name()) {
			continue
		}

		issues := rule.Check(ast)
		for _, issue := range issues {
			// Filter by minimum severity
			if issue.Severity >= l.options.MinSeverity {
				allIssues = append(allIssues, issue)
			}
		}
	}

	// Sort issues by line number, then column
	sort.Slice(allIssues, func(i, j int) bool {
		if allIssues[i].Line != allIssues[j].Line {
			return allIssues[i].Line < allIssues[j].Line
		}
		return allIssues[i].Column < allIssues[j].Column
	})

	// Count by severity
	result := &LintResult{
		FilePath: ast.FilePath,
		Issues:   allIssues,
	}

	for _, issue := range allIssues {
		switch issue.Severity {
		case SeverityError:
			result.Errors++
		case SeverityWarning:
			result.Warnings++
		case SeverityInfo:
			result.Infos++
		}
	}

	return result
}

// LintMultiple lints multiple workflow files and returns combined results
func (l *Linter) LintMultiple(paths []string) ([]*LintResult, error) {
	var results []*LintResult

	for _, path := range paths {
		result, err := l.Lint(path)
		if err != nil {
			// Include parse errors as a result with no issues but with error
			results = append(results, &LintResult{
				FilePath: path,
				Issues: []LintIssue{{
					Rule:     "parse-error",
					Severity: SeverityError,
					Message:  err.Error(),
					Line:     1,
					Column:   1,
				}},
				Errors: 1,
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// TotalErrors returns the total number of errors across all results
func TotalErrors(results []*LintResult) int {
	total := 0
	for _, r := range results {
		total += r.Errors
	}
	return total
}

// TotalWarnings returns the total number of warnings across all results
func TotalWarnings(results []*LintResult) int {
	total := 0
	for _, r := range results {
		total += r.Warnings
	}
	return total
}

// TotalInfos returns the total number of infos across all results
func TotalInfos(results []*LintResult) int {
	total := 0
	for _, r := range results {
		total += r.Infos
	}
	return total
}

// TotalIssues returns the total number of issues across all results
func TotalIssues(results []*LintResult) int {
	total := 0
	for _, r := range results {
		total += len(r.Issues)
	}
	return total
}
