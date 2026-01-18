package terminal

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StepHistoryEntry represents a completed step for display
type StepHistoryEntry struct {
	Name     string
	Symbol   string
	Type     string // bash, function, etc.
	Status   string // success, failed, skipped
	Duration time.Duration
	Command  string // command or function that was executed
	Output   string // step output
}

// ProgressBar wraps charmbracelet/bubbles progress with simple API
type ProgressBar struct {
	program        *tea.Program
	model          *progressModel
	total          int
	current        int
	description    string
	startTime      time.Time
	completedSteps []StepHistoryEntry
	mu             sync.Mutex
}

type progressModel struct {
	progress       progress.Model
	spinner        spinner.Model
	description    string
	command        string
	current        int
	total          int
	startTime      time.Time
	done           bool
	completedSteps []StepHistoryEntry
}

type tickMsg time.Time
type progressMsg float64
type doneMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m progressModel) Init() tea.Cmd {
	return tea.Batch(tickCmd(), m.spinner.Tick)
}

func (m progressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if m.done {
			return m, nil
		}
		return m, tickCmd()

	case progressMsg:
		if m.progress.Percent() >= 1.0 {
			m.done = true
			return m, nil
		}
		cmd := m.progress.SetPercent(float64(msg))
		return m, cmd

	case doneMsg:
		m.done = true
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m progressModel) View() string {
	if m.done {
		return ""
	}

	elapsed := time.Since(m.startTime)
	elapsedStr := formatElapsed(elapsed)

	var output strings.Builder

	// Only move cursor up if we have completed steps to show
	// This prevents moving into previous terminal content on first render
	if len(m.completedSteps) > 0 {
		// Count lines: each completed step + its command line (if any) + output line (if any) + progress line
		linesToGoUp := 1 // progress line
		for _, step := range m.completedSteps {
			linesToGoUp++ // step line
			if step.Command != "" {
				linesToGoUp++ // command line
			}
			if step.Output != "" {
				linesToGoUp++ // output line
			}
		}
		if m.command != "" {
			linesToGoUp++ // current command line
		}
		output.WriteString(fmt.Sprintf("\033[%dA", linesToGoUp))
	} else {
		// No completed steps - just use carriage return to stay on current line
		output.WriteString("\r")
	}

	// Render completed steps with their commands and output
	for _, step := range m.completedSteps {
		output.WriteString("\033[K") // Clear line
		statusSymbol := getStepStatusSymbol(step.Status)
		durationStr := formatElapsed(step.Duration)
		output.WriteString(fmt.Sprintf("%s %s %s (%s)\n",
			statusSymbol,
			step.Symbol,
			step.Name,
			Gray(durationStr)))

		// Show command if available
		if step.Command != "" {
			output.WriteString("\033[K") // Clear line
			output.WriteString(fmt.Sprintf("  %s\n", HiGreen(truncateCommand(step.Command, 80))))
		}

		// Show output if available
		if step.Output != "" {
			output.WriteString("\033[K") // Clear line
			output.WriteString(fmt.Sprintf("    %s\n", Gray(truncateOutput(step.Output, 76))))
		}
	}

	// Current step with progress bar
	countStr := fmt.Sprintf("(%d/%d)", m.current, m.total)
	percentStr := fmt.Sprintf("%.0f%%", m.progress.Percent()*100)

	output.WriteString("\033[K") // Clear line
	output.WriteString(fmt.Sprintf("%s %s %s %s %s %s",
		Blue(m.spinner.View()),
		Blue(m.description),
		m.progress.View(),
		Yellow(percentStr),
		Gray(countStr),
		Gray(elapsedStr)))

	// Only add newline if we have command to show below
	if m.command != "" {
		output.WriteString("\n")
		output.WriteString("\033[K") // Clear line
		output.WriteString(HiGreen(truncateCommand(m.command, 80)))
	}

	return output.String()
}

// getStepStatusSymbol returns a colored status symbol
func getStepStatusSymbol(status string) string {
	switch status {
	case "success":
		return Green(SymbolSuccess)
	case "failed":
		return Red(SymbolFailed)
	case "skipped":
		return Gray(SymbolSkipped)
	default:
		return Gray(SymbolPending)
	}
}

func formatElapsed(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

func truncateCommand(cmd string, maxLen int) string {
	// For multi-line commands, only show first line
	if idx := strings.Index(cmd, "\n"); idx != -1 {
		cmd = cmd[:idx] + "..."
	}
	if len(cmd) <= maxLen {
		return cmd
	}
	return cmd[:maxLen-3] + "..."
}

func truncateOutput(out string, maxLen int) string {
	// Get first non-empty line
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			if len(line) > maxLen {
				return line[:maxLen-3] + "..."
			}
			return line
		}
	}
	return ""
}

// NewProgressBar creates a new progress bar with the given total and description
func NewProgressBar(total int, description string) *ProgressBar {
	p := progress.New(
		progress.WithSolidFill(string(lipgloss.Color("75"))), // Light blue color
		progress.WithWidth(30),
		progress.WithoutPercentage(),
	)

	// Create spinner with default style (no lipgloss)
	s := spinner.New()
	s.Spinner = spinner.Dot

	model := &progressModel{
		progress:       p,
		spinner:        s,
		description:    description,
		total:          total,
		current:        0,
		startTime:      time.Now(),
		completedSteps: []StepHistoryEntry{},
	}

	pb := &ProgressBar{
		model:          model,
		total:          total,
		current:        0,
		description:    description,
		startTime:      time.Now(),
		completedSteps: []StepHistoryEntry{},
	}

	pb.program = tea.NewProgram(model, tea.WithOutput(os.Stderr))

	// Start the program in background
	go func() {
		_, _ = pb.program.Run()
	}()

	return pb
}

// Add increments the progress bar by n
func (p *ProgressBar) Add(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.current += n
	if p.current > p.total {
		p.current = p.total
	}

	// Update the model
	p.model.current = p.current
	percent := float64(p.current) / float64(p.total)

	if p.program != nil {
		p.program.Send(progressMsg(percent))
	}
}

// SetDescription updates the progress bar description
func (p *ProgressBar) SetDescription(desc string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.description = desc
	p.model.description = desc
}

// SetCommand updates the command displayed above the progress bar
func (p *ProgressBar) SetCommand(cmd string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.model.command = cmd
}

// AddCompletedStep adds a completed step to the history display
func (p *ProgressBar) AddCompletedStep(name, symbol, stepType, status string, duration time.Duration, command, output string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	entry := StepHistoryEntry{
		Name:     name,
		Symbol:   symbol,
		Type:     stepType,
		Status:   status,
		Duration: duration,
		Command:  command,
		Output:   output,
	}

	p.completedSteps = append(p.completedSteps, entry)
	p.model.completedSteps = p.completedSteps
}

// GetCompletedSteps returns the completed steps as StepResultEntry for markdown rendering
func (p *ProgressBar) GetCompletedSteps() []StepResultEntry {
	p.mu.Lock()
	defer p.mu.Unlock()

	results := make([]StepResultEntry, len(p.completedSteps))
	for i, step := range p.completedSteps {
		results[i] = StepResultEntry{
			Name:     step.Name,
			Type:     step.Type,
			Symbol:   step.Symbol,
			Status:   step.Status,
			Duration: step.Duration,
		}
	}
	return results
}

// Finish completes the progress bar
// showOutput: if true, show step output in the final summary (set to false in silent mode)
func (p *ProgressBar) Finish(showOutput bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.program != nil {
		p.program.Send(progressMsg(1.0)) // Set to 100% before finishing
		p.program.Send(doneMsg{})
		p.program.Wait() // Wait for program to fully quit before printing
	}

	elapsed := time.Since(p.startTime)

	// Clear current line
	fmt.Fprintf(os.Stderr, "\r\033[K")

	// Print completed steps with their status, commands, and optionally output
	for _, step := range p.completedSteps {
		statusSymbol := getStepStatusSymbol(step.Status)
		durationStr := formatElapsed(step.Duration)
		fmt.Fprintf(os.Stderr, "%s %s %s (%s)\n",
			statusSymbol,
			step.Symbol,
			step.Name,
			Gray(durationStr))
		if step.Command != "" {
			fmt.Fprintf(os.Stderr, "  %s\n", HiGreen(truncateCommand(step.Command, 80)))
		}
		// Show output if showOutput is true and output exists
		if showOutput && step.Output != "" {
			fmt.Fprintf(os.Stderr, "    %s\n", Gray(truncateOutput(step.Output, 76)))
		}
	}

	// Print full progress bar (100%) in light blue (matches running progress bar)
	fullBar := strings.Repeat("█", 30)
	fmt.Fprintf(os.Stderr, "%s %s %s %s %s %s\n\n",
		Blue(SymbolSuccess),
		Blue(p.description),
		Blue(fullBar),
		Blue("100%"),
		Gray(fmt.Sprintf("(%d/%d)", p.total, p.total)),
		Gray(formatElapsed(elapsed)))
}

// Abort cancels the progress bar on interrupt (Ctrl+C)
func (p *ProgressBar) Abort() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.program != nil {
		p.program.Send(doneMsg{})
		p.program.Quit()
	}

	// Clear the progress bar line and show cancelled message
	elapsed := time.Since(p.startTime)
	fmt.Fprintf(os.Stderr, "\r\033[K\n%s %s %s\n",
		Yellow(SymbolWarning),
		p.description,
		Gray(fmt.Sprintf("cancelled after %s", formatElapsed(elapsed))),
	)
}
