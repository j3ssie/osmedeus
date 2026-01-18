package terminal

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// StepResultEntry represents a step result for markdown table rendering
type StepResultEntry struct {
	Name     string
	Type     string // bash, function, etc.
	Symbol   string // step type symbol
	Status   string // success, failed, skipped
	Duration time.Duration
}

// RenderStepResultsMarkdown outputs step results as a markdown table
func RenderStepResultsMarkdown(w io.Writer, steps []StepResultEntry) {
	if len(steps) == 0 {
		return
	}

	// Calculate column widths for alignment
	nameWidth := 4 // "Step"
	typeWidth := 4 // "Type"
	for _, s := range steps {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
		if len(s.Type) > typeWidth {
			typeWidth = len(s.Type)
		}
	}

	// Header
	_, _ = fmt.Fprintf(w, "\n| %-*s | %-*s | Status | Duration |\n",
		nameWidth, "Step",
		typeWidth, "Type")

	// Separator
	_, _ = fmt.Fprintf(w, "|-%s-|-%s-|--------|----------|\n",
		strings.Repeat("-", nameWidth),
		strings.Repeat("-", typeWidth))

	// Rows
	for _, s := range steps {
		statusSymbol := getStatusSymbol(s.Status)
		durationStr := formatDurationShort(s.Duration)

		_, _ = fmt.Fprintf(w, "| %-*s | %-*s | %s | %s |\n",
			nameWidth, s.Name,
			typeWidth, s.Type,
			statusSymbol,
			durationStr)
	}
	_, _ = fmt.Fprintln(w)
}

// getStatusSymbol returns a colored status symbol
func getStatusSymbol(status string) string {
	switch status {
	case "success":
		return Green(SymbolSuccess) + " success"
	case "failed":
		return Red(SymbolFailed) + " failed "
	case "skipped":
		return Gray(SymbolSkipped) + " skipped"
	default:
		return Gray(SymbolPending) + " " + status
	}
}

// formatDurationShort formats duration in a compact way
func formatDurationShort(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", mins, secs)
	}
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, mins)
}
