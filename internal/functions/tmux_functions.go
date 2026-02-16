package functions

import (
	"fmt"
	mrand "math/rand"
	"os/exec"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// findTmuxBin checks if tmux is installed and returns its path.
// Returns empty string if tmux is not found.
func findTmuxBin() string {
	path, err := exec.LookPath("tmux")
	if err != nil {
		return ""
	}
	return path
}

// tmuxSessionExists checks whether a tmux session with the given name exists.
func tmuxSessionExists(tmuxBin, name string) bool {
	cmd := exec.Command(tmuxBin, "has-session", "-t", name)
	return cmd.Run() == nil
}

// listTmuxSessions returns a list of all tmux session names.
// Returns empty slice if no sessions exist or tmux server is not running.
func listTmuxSessions(tmuxBin string) []string {
	cmd := exec.Command(tmuxBin, "list-sessions", "-F", "#{session_name}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	raw := strings.TrimSpace(string(output))
	if raw == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	sessions := make([]string, 0, len(lines))
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name != "" {
			sessions = append(sessions, name)
		}
	}
	return sessions
}

// generateTmuxSessionName produces a session name in the format "bosm-<random8>".
func generateTmuxSessionName() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	rng := mrand.New(mrand.NewSource(time.Now().UnixNano()))

	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return "bosm-" + string(b)
}

// tmuxRun creates a detached tmux session running the given command.
// Usage: tmux_run(command, session_name?) -> string (session name) or "" on error
func (vf *vmFunc) tmuxRun(call goja.FunctionCall) goja.Value {
	command := call.Argument(0).String()

	logger.Get().Debug("Calling "+terminal.HiGreen(FnTmuxRun),
		zap.String("command", command))

	if command == "undefined" || command == "" {
		logger.Get().Warn(FnTmuxRun + ": command is required")
		return vf.vm.ToValue("")
	}

	tmuxBin := findTmuxBin()
	if tmuxBin == "" {
		logger.Get().Warn(FnTmuxRun + ": tmux is not installed")
		return vf.vm.ToValue("")
	}

	// Determine session name
	sessionName := ""
	if len(call.Arguments) > 1 {
		sessionName = call.Argument(1).String()
	}
	if sessionName == "" || sessionName == "undefined" {
		sessionName = generateTmuxSessionName()
	}

	// Create detached tmux session
	cmd := exec.Command(tmuxBin, "new-session", "-d", "-s", sessionName, command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Get().Warn(FnTmuxRun+": failed to create tmux session",
			zap.String("session", sessionName), zap.Error(err),
			zap.String("output", strings.TrimSpace(string(output))))
		return vf.vm.ToValue("")
	}

	fmt.Printf("%s %s %s\n", terminal.Green("[tmux]"), "created session:", terminal.HiGreen(sessionName))
	logger.Get().Debug(terminal.HiGreen(FnTmuxRun)+" session created",
		zap.String("session", sessionName))
	return vf.vm.ToValue(sessionName)
}

// tmuxCapture captures the current pane output from a tmux session.
// When session_name is "all", captures output from all sessions with headers.
// Usage: tmux_capture(session_name) -> string (pane content) or "" on error
func (vf *vmFunc) tmuxCapture(call goja.FunctionCall) goja.Value {
	sessionName := call.Argument(0).String()

	logger.Get().Debug("Calling "+terminal.HiGreen(FnTmuxCapture),
		zap.String("session", sessionName))

	if sessionName == "undefined" || sessionName == "" {
		logger.Get().Warn(FnTmuxCapture + ": session_name is required")
		return vf.vm.ToValue("")
	}

	tmuxBin := findTmuxBin()
	if tmuxBin == "" {
		logger.Get().Warn(FnTmuxCapture + ": tmux is not installed")
		return vf.vm.ToValue("")
	}

	// Handle "all" - capture from every session
	if sessionName == "all" {
		sessions := listTmuxSessions(tmuxBin)
		fmt.Printf("%s %s %s\n", terminal.Green("[tmux]"), "capturing all", terminal.HiGreen(fmt.Sprintf("%d session(s)", len(sessions))))

		if len(sessions) == 0 {
			return vf.vm.ToValue("")
		}

		var parts []string
		for _, name := range sessions {
			cmd := exec.Command(tmuxBin, "capture-pane", "-t", name, "-p")
			output, err := cmd.CombinedOutput()
			if err != nil {
				logger.Get().Warn(FnTmuxCapture+": failed to capture pane",
					zap.String("session", name), zap.Error(err))
				continue
			}
			captured := strings.TrimRight(string(output), "\n")
			parts = append(parts, fmt.Sprintf("=== session: %s ===\n%s", name, captured))
		}

		result := strings.Join(parts, "\n")
		logger.Get().Debug(terminal.HiGreen(FnTmuxCapture)+" captured all sessions",
			zap.Int("sessionCount", len(sessions)), zap.Int("outputLength", len(result)))
		return vf.vm.ToValue(result)
	}

	if !tmuxSessionExists(tmuxBin, sessionName) {
		logger.Get().Warn(FnTmuxCapture+": session does not exist",
			zap.String("session", sessionName))
		return vf.vm.ToValue("")
	}

	fmt.Printf("%s %s %s\n", terminal.Green("[tmux]"), "capturing session:", terminal.HiGreen(sessionName))

	cmd := exec.Command(tmuxBin, "capture-pane", "-t", sessionName, "-p")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Get().Warn(FnTmuxCapture+": failed to capture pane",
			zap.String("session", sessionName), zap.Error(err))
		return vf.vm.ToValue("")
	}

	result := strings.TrimRight(string(output), "\n")
	logger.Get().Debug(terminal.HiGreen(FnTmuxCapture)+" result",
		zap.String("session", sessionName), zap.Int("outputLength", len(result)))
	return vf.vm.ToValue(result)
}

// tmuxSend sends a command (keystrokes) to an existing tmux session.
// Usage: tmux_send(session_name, command) -> bool
func (vf *vmFunc) tmuxSend(call goja.FunctionCall) goja.Value {
	sessionName := call.Argument(0).String()
	command := call.Argument(1).String()

	logger.Get().Debug("Calling "+terminal.HiGreen(FnTmuxSend),
		zap.String("session", sessionName), zap.String("command", command))

	if sessionName == "undefined" || sessionName == "" {
		logger.Get().Warn(FnTmuxSend + ": session_name is required")
		return vf.vm.ToValue(false)
	}
	if command == "undefined" || command == "" {
		logger.Get().Warn(FnTmuxSend + ": command is required")
		return vf.vm.ToValue(false)
	}

	tmuxBin := findTmuxBin()
	if tmuxBin == "" {
		logger.Get().Warn(FnTmuxSend + ": tmux is not installed")
		return vf.vm.ToValue(false)
	}

	if !tmuxSessionExists(tmuxBin, sessionName) {
		logger.Get().Warn(FnTmuxSend+": session does not exist",
			zap.String("session", sessionName))
		return vf.vm.ToValue(false)
	}

	cmd := exec.Command(tmuxBin, "send-keys", "-t", sessionName, command, "Enter")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Get().Warn(FnTmuxSend+": failed to send keys",
			zap.String("session", sessionName), zap.Error(err),
			zap.String("output", strings.TrimSpace(string(output))))
		return vf.vm.ToValue(false)
	}

	fmt.Printf("%s %s %s\n", terminal.Green("[tmux]"), "sent command to session:", terminal.HiGreen(sessionName))
	logger.Get().Debug(terminal.HiGreen(FnTmuxSend)+" keys sent",
		zap.String("session", sessionName))
	return vf.vm.ToValue(true)
}

// tmuxKill kills (destroys) a tmux session.
// Usage: tmux_kill(session_name) -> bool
func (vf *vmFunc) tmuxKill(call goja.FunctionCall) goja.Value {
	sessionName := call.Argument(0).String()

	logger.Get().Debug("Calling "+terminal.HiGreen(FnTmuxKill),
		zap.String("session", sessionName))

	if sessionName == "undefined" || sessionName == "" {
		logger.Get().Warn(FnTmuxKill + ": session_name is required")
		return vf.vm.ToValue(false)
	}

	tmuxBin := findTmuxBin()
	if tmuxBin == "" {
		logger.Get().Warn(FnTmuxKill + ": tmux is not installed")
		return vf.vm.ToValue(false)
	}

	if !tmuxSessionExists(tmuxBin, sessionName) {
		logger.Get().Warn(FnTmuxKill+": session does not exist",
			zap.String("session", sessionName))
		return vf.vm.ToValue(false)
	}

	cmd := exec.Command(tmuxBin, "kill-session", "-t", sessionName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Get().Warn(FnTmuxKill+": failed to kill session",
			zap.String("session", sessionName), zap.Error(err),
			zap.String("output", strings.TrimSpace(string(output))))
		return vf.vm.ToValue(false)
	}

	fmt.Printf("%s %s %s\n", terminal.Green("[tmux]"), "killed session:", terminal.HiGreen(sessionName))
	logger.Get().Debug(terminal.HiGreen(FnTmuxKill)+" session killed",
		zap.String("session", sessionName))
	return vf.vm.ToValue(true)
}

// tmuxList lists all tmux session names.
// Usage: tmux_list() -> []string
func (vf *vmFunc) tmuxList(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen(FnTmuxList))

	tmuxBin := findTmuxBin()
	if tmuxBin == "" {
		logger.Get().Warn(FnTmuxList + ": tmux is not installed")
		return vf.vm.ToValue([]string{})
	}

	sessions := listTmuxSessions(tmuxBin)
	if sessions == nil {
		sessions = []string{}
	}

	fmt.Printf("%s %s %s\n", terminal.Green("[tmux]"), "found", terminal.HiGreen(fmt.Sprintf("%d session(s)", len(sessions))))
	logger.Get().Debug(terminal.HiGreen(FnTmuxList)+" result",
		zap.Int("count", len(sessions)))
	return vf.vm.ToValue(sessions)
}
