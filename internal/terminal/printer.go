package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Printer provides formatted output helpers for terminal display
type Printer struct {
	highlighter *Highlighter
}

// NewPrinter creates a new printer
func NewPrinter() *Printer {
	return &Printer{
		highlighter: NewHighlighter(),
	}
}

// printJSONL outputs a JSON line for CI mode
func printJSONL(data map[string]interface{}) {
	jsonBytes, _ := json.Marshal(data)
	fmt.Println(string(jsonBytes))
}

// Info prints an info message with symbol
func (p *Printer) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if IsCIMode() {
		printJSONL(map[string]interface{}{"type": "info", "message": msg})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", InfoSymbol(), msg)
}

// Success prints a success message with symbol
func (p *Printer) Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if IsCIMode() {
		printJSONL(map[string]interface{}{"type": "success", "message": msg})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", StepSuccessSymbol(), msg)
}

// Warning prints a warning message with symbol
func (p *Printer) Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if IsCIMode() {
		printJSONL(map[string]interface{}{"type": "warning", "message": msg})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", WarningSymbol(), msg)
}

// Error prints an error message with colored prefix
func (p *Printer) Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if IsCIMode() {
		printJSONL(map[string]interface{}{"type": "error", "message": msg})
		return
	}
	fmt.Fprintf(os.Stderr, "%s %s\n", Red("Error:"), msg)
}

// SecurityWarning prints a security warning with red label and yellow message
func (p *Printer) SecurityWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(os.Stdout, "%s %s %s\n", WarningSymbol(), BoldRed("Security Warning:"), Yellow(msg))
}

// Installing prints an installing message with symbol
func (p *Printer) Installing(name string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s '%s'\n", Cyan(SymbolStart), "installing", Cyan(name))
}

// GrayOutput prints output text in gray (for verbose subprocess output)
func (p *Printer) GrayOutput(output string) {
	if output == "" {
		return
	}
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	for _, line := range lines {
		if line != "" {
			_, _ = fmt.Fprintf(os.Stdout, "  %s\n", Gray(line))
		}
	}
}

// StepStart prints step start message
func (p *Printer) StepStart(stepName string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", StepStartSymbol(), HiBlue(stepName))
}

// StepRunning prints step running message
func (p *Printer) StepRunning(stepName string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", StepRunningSymbol(), HiBlue(stepName))
}

// StepSuccess prints step success message with duration
func (p *Printer) StepSuccess(stepName, duration string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s %s\n", StepSuccessSymbol(), HiBlue(stepName), Gray("("+duration+")"))
}

// StepFailed prints step failed message with error
func (p *Printer) StepFailed(stepName string, err error) {
	errMsg := ""
	if err != nil {
		errMsg = ": " + err.Error()
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s %s%s\n", StepFailedSymbol(), HiBlue(stepName), Red(errMsg))
}

// StepSkipped prints step skipped message
func (p *Printer) StepSkipped(stepName string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s %s\n", StepSkippedSymbol(), Gray(stepName), Gray("(skipped)"))
}

// StepStartWithCommand prints step start with type symbol and command
func (p *Printer) StepStartWithCommand(stepName, typeSymbol, command, cmdPrefix string) {
	if IsCIMode() {
		printJSONL(map[string]interface{}{
			"type":    "step_start",
			"step":    stepName,
			"command": command,
		})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s\n", StepStartSymbol(), typeSymbol, Gray("(starting)"), HiBlue(stepName))
	if command != "" {
		_, _ = fmt.Fprintf(os.Stdout, "  %s\n", HiGreen(formatMultilineCommand(command, cmdPrefix)))
	}
}

// StepSuccessWithCommand prints step success with type symbol and command
func (p *Printer) StepSuccessWithCommand(stepName, typeSymbol, duration, command, cmdPrefix string) {
	if IsCIMode() {
		printJSONL(map[string]interface{}{
			"type":     "step_success",
			"step":     stepName,
			"duration": duration,
			"command":  command,
		})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s\n", StepSuccessSymbol(), typeSymbol, Gray("(finished in "+duration+")"), HiBlue(stepName))
	// Command already shown when step started, no need to repeat
}

// StepFailedWithCommand prints step failed with type symbol and command
func (p *Printer) StepFailedWithCommand(stepName, typeSymbol string, err error, command, cmdPrefix string) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	if IsCIMode() {
		printJSONL(map[string]interface{}{
			"type":    "step_failed",
			"step":    stepName,
			"error":   errMsg,
			"command": command,
		})
		return
	}
	if errMsg != "" {
		errMsg = ": " + errMsg
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s%s\n", StepFailedSymbol(), typeSymbol, Red("(failed)"), HiBlue(stepName), Red(errMsg))
	if command != "" {
		_, _ = fmt.Fprintf(os.Stdout, "  %s\n", Red(formatMultilineCommand(command, cmdPrefix)))
	}
}

// StepSkippedWithCommand prints step skipped with type symbol
func (p *Printer) StepSkippedWithCommand(stepName, typeSymbol string) {
	if IsCIMode() {
		printJSONL(map[string]interface{}{
			"type": "step_skipped",
			"step": stepName,
		})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "%s %s %s %s\n", StepSkippedSymbol(), typeSymbol, Gray("(skipped)"), Gray(stepName))
}

// WorkflowInfo prints workflow metadata (for normal mode)
func (p *Printer) WorkflowInfo(name, description string, tags []string, runnerType string, totalSteps int) {
	if IsCIMode() {
		printJSONL(map[string]interface{}{
			"type":        "workflow_start",
			"workflow":    name,
			"description": description,
			"tags":        tags,
			"runner":      runnerType,
			"total_steps": totalSteps,
		})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "\n%s %s %s\n", Yellow(SymbolStar), Bold("Executing:"), name)
	if description != "" {
		_, _ = fmt.Fprintf(os.Stdout, "  %s %s\n", Gray("Description:"), description)
	}
	if len(tags) > 0 {
		_, _ = fmt.Fprintf(os.Stdout, "  %s %s\n", Gray("Tags:"), strings.Join(tags, ", "))
	}
	if runnerType != "" && runnerType != "host" {
		_, _ = fmt.Fprintf(os.Stdout, "  %s %s\n", Gray("Runner:"), runnerType)
	}
	_, _ = fmt.Fprintf(os.Stdout, "  %s %d\n\n", Gray("Total Steps:"), totalSteps)
}

// Section prints a section header with symbol
func (p *Printer) Section(title string) {
	if IsCIMode() {
		printJSONL(map[string]interface{}{"type": "section", "title": title})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "\n%s %s\n", SectionSymbol(), Bold(title))
}

// SubSection prints a subsection header with symbol
func (p *Printer) SubSection(title string) {
	if IsCIMode() {
		printJSONL(map[string]interface{}{"type": "subsection", "title": title})
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "\n%s %s\n", SubSectionSymbol(), BoldCyan(title))
}

// KeyValue prints a key-value pair
func (p *Printer) KeyValue(key, value string) {
	_, _ = fmt.Fprintf(os.Stdout, "  %s: %s\n", Gray(key), value)
}

// KeyValueColored prints a key-value pair with colored value
func (p *Printer) KeyValueColored(key, value string, colorFn func(string) string) {
	_, _ = fmt.Fprintf(os.Stdout, "  %s: %s\n", Gray(key), colorFn(value))
}

// Bullet prints a bullet point
func (p *Printer) Bullet(text string) {
	_, _ = fmt.Fprintf(os.Stdout, "  %s %s\n", SymbolBullet, text)
}

// BulletColored prints a colored bullet point
func (p *Printer) BulletColored(text string, colorFn func(string) string) {
	_, _ = fmt.Fprintf(os.Stdout, "  %s %s\n", SymbolBullet, colorFn(text))
}

// Divider prints a divider line
func (p *Printer) Divider() {
	if IsCIMode() {
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, Gray(strings.Repeat("─", 40)))
}

// DoubleDivider prints a double divider line
func (p *Printer) DoubleDivider() {
	if IsCIMode() {
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, Gray(strings.Repeat("═", 40)))
}

// Newline prints an empty line
func (p *Printer) Newline() {
	if IsCIMode() {
		return
	}
	_, _ = fmt.Fprintln(os.Stdout)
}

// Print prints plain text
func (p *Printer) Print(format string, args ...interface{}) {
	if IsCIMode() {
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

// Println prints plain text with newline
func (p *Printer) Println(format string, args ...interface{}) {
	if IsCIMode() {
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// Flow prints a flow indicator with name
func (p *Printer) Flow(name string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", FlowSymbol(), Cyan(name))
}

// Module prints a module indicator with name
func (p *Printer) Module(name string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", ModuleSymbol(), Yellow(name))
}

// Highlighter returns the syntax highlighter
func (p *Printer) Highlighter() *Highlighter {
	return p.highlighter
}

// HighlightYAML prints highlighted YAML content
func (p *Printer) HighlightYAML(content string) {
	highlighted, err := p.highlighter.HighlightYAML(content)
	if err != nil {
		_, _ = fmt.Fprint(os.Stdout, content)
		return
	}
	_, _ = fmt.Fprint(os.Stdout, highlighted)
}

// HighlightMarkdown prints highlighted Markdown content
func (p *Printer) HighlightMarkdown(content string) {
	highlighted, err := p.highlighter.HighlightMarkdown(content)
	if err != nil {
		_, _ = fmt.Fprint(os.Stdout, content)
		return
	}
	_, _ = fmt.Fprint(os.Stdout, highlighted)
}

// formatMultilineCommand formats a multi-line command with proper indentation
func formatMultilineCommand(cmd, cmdPrefix string) string {
	lines := strings.Split(strings.TrimSuffix(cmd, "\n"), "\n")
	if len(lines) <= 1 {
		return cmdPrefix + " " + cmd
	}

	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, cmdPrefix+" "+line)
		}
	}
	return strings.Join(result, "\n  ")
}

// StatusBadge returns a colored status badge
func StatusBadge(status string) string {
	switch strings.ToLower(status) {
	case "completed", "success", "done", "passed":
		return Green(status)
	case "failed", "error":
		return Red(status)
	case "running", "in_progress", "active":
		return Cyan(status)
	case "pending", "waiting", "queued":
		return Yellow(status)
	case "cancelled", "skipped":
		return Gray(status)
	default:
		return status
	}
}

// TypeBadge returns a colored type badge
func TypeBadge(typ string) string {
	switch strings.ToLower(typ) {
	case "flow":
		return Cyan(typ)
	case "module":
		return Yellow(typ)
	case "bash":
		return Green(typ)
	case "function":
		return Magenta(typ)
	case "parallel":
		return Blue(typ)
	case "foreach":
		return Blue(typ)
	default:
		return typ
	}
}

// VerboseOutput prints step output with indentation (for verbose mode)
func (p *Printer) VerboseOutput(output string) {
	if output == "" {
		return
	}
	_, _ = fmt.Fprintf(os.Stdout, "  %s\n", Gray("[output]"))
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	for _, line := range lines {
		_, _ = fmt.Fprintf(os.Stdout, "  %s\n", line)
	}
}

// VerboseInfo prints a verbose info message with source file reference
func (p *Printer) VerboseInfo(msg, source string) {
	_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", InfoSymbol(), HiCyan(msg))
	if source != "" {
		_, _ = fmt.Fprintf(os.Stdout, "  %s\n", Gray(source))
	}
}
