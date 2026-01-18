package terminal

import (
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Spinner wraps charmbracelet/bubbles spinner with simple Start/Stop API
type Spinner struct {
	program *tea.Program
	model   *spinnerModel
	mu      sync.Mutex
	running bool
}

type spinnerModel struct {
	spinner  spinner.Model
	prefix   string
	suffix   string
	style    lipgloss.Style
	quitting bool
}

type quitMsg struct{}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case quitMsg:
		m.quitting = true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf("%s%s %s", m.prefix, m.style.Render(m.spinner.View()), m.suffix)
}

// newSpinner creates a spinner with the given configuration
func newSpinner(prefix, suffix string, color string) *Spinner {
	s := spinner.New()
	s.Spinner = spinner.Dot

	// Set style based on color
	var style lipgloss.Style
	switch color {
	case "cyan":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // Cyan
	case "yellow":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	case "green":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	case "red":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
	default:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("6")) // Default cyan
	}

	model := &spinnerModel{
		spinner: s,
		prefix:  prefix,
		suffix:  suffix,
		style:   style,
	}

	return &Spinner{
		model: model,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	// Create a fresh model for each start
	newModel := &spinnerModel{
		spinner: s.model.spinner,
		prefix:  s.model.prefix,
		suffix:  s.model.suffix,
		style:   s.model.style,
	}

	s.program = tea.NewProgram(newModel, tea.WithOutput(os.Stderr))
	s.running = true

	go func() {
		_, _ = s.program.Run()
	}()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running || s.program == nil {
		s.mu.Unlock()
		return
	}

	program := s.program
	s.running = false
	s.mu.Unlock()

	// Send quit and wait for program to finish
	program.Send(quitMsg{})
	program.Wait()

	// Clear the spinner line
	fmt.Fprint(os.Stderr, "\r\033[K")
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return newSpinner("", message, "cyan")
}

// NewSpinnerWithPrefix creates a spinner with prefix and suffix
func NewSpinnerWithPrefix(prefix, suffix string) *Spinner {
	return newSpinner(prefix+" ", suffix, "cyan")
}

// StepSpinner creates a spinner for step execution
func StepSpinner(stepName string) *Spinner {
	return newSpinner(SymbolRunning+" ", stepName, "cyan")
}

// LoadingSpinner creates a spinner for loading operations
func LoadingSpinner(message string) *Spinner {
	return newSpinner("", message+"...", "cyan")
}

// ProcessingSpinner creates a spinner for processing operations
func ProcessingSpinner(message string) *Spinner {
	return newSpinner("", message, "yellow")
}
