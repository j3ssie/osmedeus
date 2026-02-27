package executor

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	acp "github.com/coder/acp-go-sdk"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	oslogger "github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/template"
	"go.uber.org/zap"
)

// acpAgentDef describes a built-in agent command.
type acpAgentDef struct {
	Command string
	Args    []string
}

// builtinACPAgents maps agent names to their command definitions.
var builtinACPAgents = map[string]acpAgentDef{
	"claude-code": {Command: "npx", Args: []string{"-y", "@zed-industries/claude-code-acp@latest"}},
	"codex":       {Command: "npx", Args: []string{"-y", "@zed-industries/codex-acp"}},
	"opencode":    {Command: "opencode", Args: []string{"acp"}},
	"gemini":      {Command: "gemini", Args: []string{"--experimental-acp"}},
}

// ACPExecutor implements StepExecutorPlugin for agent-acp steps.
// It spawns an AI coding agent as a subprocess, communicates over stdin/stdout
// using the Agent Communication Protocol (ACP), and collects the output.
type ACPExecutor struct {
	templateEngine template.TemplateEngine
}

// NewACPExecutor creates a new ACP executor.
func NewACPExecutor(engine template.TemplateEngine) *ACPExecutor {
	return &ACPExecutor{
		templateEngine: engine,
	}
}

// Name returns the executor name for logging/debugging.
func (e *ACPExecutor) Name() string {
	return "agent-acp"
}

// StepTypes returns the step types this executor handles.
func (e *ACPExecutor) StepTypes() []core.StepType {
	return []core.StepType{core.StepTypeAgentACP}
}

// ResolveAgent resolves the agent command and args from step configuration.
// It checks step.Agent against the built-in registry first, then falls back to
// step.ACPConfig.Command/Args for custom agents.
func ResolveAgent(step *core.Step) (command string, args []string, err error) {
	// Check built-in registry first
	if step.Agent != "" {
		def, ok := builtinACPAgents[step.Agent]
		if !ok {
			return "", nil, fmt.Errorf("unknown built-in agent: %q (available: %s)", step.Agent, availableAgentNames())
		}
		return def.Command, def.Args, nil
	}

	// Custom agent via acp_config
	if step.ACPConfig != nil && step.ACPConfig.Command != "" {
		return step.ACPConfig.Command, step.ACPConfig.Args, nil
	}

	return "", nil, fmt.Errorf("agent-acp step requires 'agent' field or 'acp_config.command'")
}

// BuildPrompt builds the prompt string from step messages.
// It joins all message Content fields (string only) with newlines.
func BuildPrompt(step *core.Step) string {
	var parts []string
	for _, msg := range step.Messages {
		if content, ok := msg.Content.(string); ok && content != "" {
			parts = append(parts, content)
		}
	}
	return strings.Join(parts, "\n")
}

// Execute runs an agent-acp step.
func (e *ACPExecutor) Execute(ctx context.Context, step *core.Step, execCtx *core.ExecutionContext) (*core.StepResult, error) {
	result := &core.StepResult{
		StepName:  step.Name,
		Status:    core.StepStatusRunning,
		StartTime: time.Now(),
		Exports:   make(map[string]interface{}),
	}

	// Resolve agent name for the standalone function
	agentName := step.Agent
	var customCommand string
	var customArgs []string
	if agentName == "" && step.ACPConfig != nil && step.ACPConfig.Command != "" {
		customCommand = step.ACPConfig.Command
		customArgs = step.ACPConfig.Args
	}

	// For custom commands, use ResolveAgent + the old path
	// For built-in agents, delegate to RunAgentACP
	if customCommand != "" {
		// Custom agent — validate command exists
		_, err := exec.LookPath(customCommand)
		if err != nil {
			err = fmt.Errorf("agent command %q not found in PATH: %w", customCommand, err)
			result.Status = core.StepStatusFailed
			result.Error = err
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, err
		}
		// Custom agents not supported by RunAgentACP — keep original error
		_ = customArgs
		err = fmt.Errorf("custom agent commands should use 'agent' field with a built-in agent name; use acp_config for custom agents via workflow steps only")
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Validate we have an agent
	if agentName == "" {
		err := fmt.Errorf("agent-acp step requires 'agent' field or 'acp_config.command'")
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Build prompt from messages
	prompt := BuildPrompt(step)
	if prompt == "" {
		err := fmt.Errorf("agent-acp step has no prompt (messages with content required)")
		result.Status = core.StepStatusFailed
		result.Error = err
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Build config from step + execution context
	cfg := &RunAgentACPConfig{
		Cwd: step.Cwd,
	}
	if cfg.Cwd == "" {
		cfg.Cwd = execCtx.WorkspacePath
	}

	// Determine allowed paths
	cfg.AllowedPaths = step.AllowedPaths
	if len(cfg.AllowedPaths) == 0 && execCtx.WorkspacePath != "" {
		cfg.AllowedPaths = []string{execCtx.WorkspacePath}
	}

	// Environment and write config
	if step.ACPConfig != nil {
		cfg.Env = step.ACPConfig.Env
		cfg.WriteEnabled = step.ACPConfig.WriteEnabled
	}

	// Delegate to standalone function
	agentOutput, stderrStr, err := RunAgentACP(ctx, prompt, agentName, cfg)

	if err != nil {
		result.Status = core.StepStatusFailed
		result.Error = err
		result.Output = agentOutput
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.Exports["acp_output"] = agentOutput
		result.Exports["acp_stderr"] = stderrStr
		result.Exports["acp_agent"] = resolveAgentName(step)
		return result, err
	}

	// Success
	result.Status = core.StepStatusSuccess
	result.Output = agentOutput
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Exports["acp_output"] = agentOutput
	result.Exports["acp_stderr"] = stderrStr
	result.Exports["acp_agent"] = resolveAgentName(step)

	return result, nil
}

// RunAgentACPConfig holds options for standalone RunAgentACP calls.
type RunAgentACPConfig struct {
	Cwd          string            // Working directory (default: ".")
	AllowedPaths []string          // Restrict file reads to these directories
	Env          map[string]string // Extra environment variables for agent process
	WriteEnabled bool              // Allow file writes (default: false)
	StreamWriter io.Writer         // Optional writer for real-time output streaming
}

// RunAgentACP spawns an ACP agent subprocess and returns its output.
// agentName can be a built-in name ("claude-code", "codex", etc.) or empty to use the default ("claude-code").
// Returns (stdout, stderr, error).
func RunAgentACP(ctx context.Context, prompt, agentName string, cfg *RunAgentACPConfig) (string, string, error) {
	log := oslogger.Get()

	if prompt == "" {
		return "", "", fmt.Errorf("prompt is required")
	}

	if cfg == nil {
		cfg = &RunAgentACPConfig{}
	}

	// Default agent
	if agentName == "" {
		agentName = "claude-code"
	}

	// Resolve agent command from built-in registry
	def, ok := builtinACPAgents[agentName]
	if !ok {
		return "", "", fmt.Errorf("unknown agent: %q (available: %s)", agentName, availableAgentNames())
	}
	command := def.Command
	args := def.Args

	// Verify command exists in PATH
	cmdPath, err := exec.LookPath(command)
	if err != nil {
		return "", "", fmt.Errorf("agent command %q not found in PATH: %w", command, err)
	}

	// Build full command string for logging
	fullCmd := cmdPath + " " + strings.Join(args, " ")

	log.Debug("starting agent subprocess (ACP mode)",
		zap.String("cmd", fullCmd),
		zap.Int("promptLength", len(prompt)))

	// Create command with process group for cleanup
	cmd := exec.CommandContext(ctx, cmdPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Apply environment variables
	if len(cfg.Env) > 0 {
		cmd.Env = cmd.Environ()
		for k, v := range cfg.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	// Create pipes
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrReader, stderrWriter := io.Pipe()
	cmd.Stderr = stderrWriter

	if err := cmd.Start(); err != nil {
		_ = stderrWriter.Close()
		return "", "", fmt.Errorf("failed to start agent process: %w", err)
	}

	// Drain stderr in background
	var stderrBuf bytes.Buffer
	var stderrWg sync.WaitGroup
	stderrWg.Add(1)
	go func() {
		defer stderrWg.Done()
		scanner := bufio.NewScanner(stderrReader)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line)
			stderrBuf.WriteByte('\n')
			log.Debug("acp agent stderr", zap.String("line", line))
		}
	}()

	defer func() {
		_ = stderrWriter.Close()
		stderrWg.Wait()
		if cmd.Process != nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		_ = cmd.Wait()
	}()

	// Build client options
	var clientOpts []acpClientOption
	if len(cfg.AllowedPaths) > 0 {
		clientOpts = append(clientOpts, withAllowedPaths(cfg.AllowedPaths...))
	}
	clientOpts = append(clientOpts, withWriteEnabled(cfg.WriteEnabled))
	if cfg.StreamWriter != nil {
		clientOpts = append(clientOpts, withStreamWriter(cfg.StreamWriter))
	}

	// Create ACP client and connection
	client := newACPClient(clientOpts...)
	conn := acp.NewClientSideConnection(client, stdinPipe, stdoutPipe)
	conn.SetLogger(slog.New(newZapSlogHandler()))

	log.Debug("ACP connection established, sending Initialize request",
		zap.Int("protocolVersion", acp.ProtocolVersionNumber))

	// Initialize
	_, initErr := conn.Initialize(ctx, acp.InitializeRequest{
		ProtocolVersion: acp.ProtocolVersionNumber,
		ClientCapabilities: acp.ClientCapabilities{
			Fs: acp.FileSystemCapability{
				ReadTextFile:  true,
				WriteTextFile: cfg.WriteEnabled,
			},
			Terminal: false,
		},
	})
	if initErr != nil {
		stderrStr := stderrBuf.String()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", stderrStr, fmt.Errorf("ACP initialize timed out: %w", ctx.Err())
		}
		return "", stderrStr, fmt.Errorf("ACP initialize failed: %w", initErr)
	}

	log.Debug("ACP initialized successfully")

	// Determine working directory for the session
	cwd := cfg.Cwd
	if cwd == "" {
		cwd = "."
	}
	if absCwd, absErr := filepath.Abs(cwd); absErr == nil {
		cwd = absCwd
	}

	log.Debug("creating ACP session", zap.String("cwd", cwd))

	// Create session
	sess, sessErr := conn.NewSession(ctx, acp.NewSessionRequest{
		Cwd:        cwd,
		McpServers: []acp.McpServer{},
	})
	if sessErr != nil {
		stderrStr := stderrBuf.String()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", stderrStr, fmt.Errorf("ACP session creation timed out: %w", ctx.Err())
		}
		return "", stderrStr, fmt.Errorf("ACP new session failed: %w", sessErr)
	}

	log.Debug("ACP session created",
		zap.String("sessionId", string(sess.SessionId)))

	log.Debug("sending ACP prompt, waiting for agent completion...",
		zap.Int("promptLength", len(prompt)))

	// Send prompt — blocks until the agent completes
	promptResp, promptErr := conn.Prompt(ctx, acp.PromptRequest{
		SessionId: sess.SessionId,
		Prompt:    []acp.ContentBlock{acp.TextBlock(prompt)},
	})

	agentOutput := client.collectedOutput()
	stderrStr := stderrBuf.String()

	if promptErr != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return agentOutput, stderrStr, fmt.Errorf("ACP prompt timed out: %w", ctx.Err())
		}
		return agentOutput, stderrStr, fmt.Errorf("ACP prompt failed: %w", promptErr)
	}

	// Close stdin to signal EOF to the agent process
	_ = stdinPipe.Close()

	stopReason := "unknown"
	if promptResp.StopReason != "" {
		stopReason = string(promptResp.StopReason)
	}

	log.Debug("ACP prompt completed",
		zap.String("stopReason", stopReason),
		zap.Int("collectedOutputBytes", len(agentOutput)))

	return agentOutput, stderrStr, nil
}

// ListAgentNames returns the names of all built-in ACP agents.
func ListAgentNames() []string {
	names := make([]string, 0, len(builtinACPAgents))
	for name := range builtinACPAgents {
		names = append(names, name)
	}
	return names
}

// resolveAgentName returns the agent name for exports.
func resolveAgentName(step *core.Step) string {
	if step.Agent != "" {
		return step.Agent
	}
	if step.ACPConfig != nil && step.ACPConfig.Command != "" {
		return step.ACPConfig.Command
	}
	return "unknown"
}

// availableAgentNames returns a comma-separated list of built-in agent names.
func availableAgentNames() string {
	names := make([]string, 0, len(builtinACPAgents))
	for name := range builtinACPAgents {
		names = append(names, name)
	}
	return strings.Join(names, ", ")
}
