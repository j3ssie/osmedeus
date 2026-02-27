package executor

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	acp "github.com/coder/acp-go-sdk"
	oslogger "github.com/j3ssie/osmedeus/v5/internal/logger"
	"go.uber.org/zap"
)

// acpClientOption configures an acpClient.
type acpClientOption func(*acpClient)

// withAllowedPaths restricts ReadTextFile to files under the given directories.
func withAllowedPaths(paths ...string) acpClientOption {
	return func(c *acpClient) {
		c.allowedPaths = append(c.allowedPaths, paths...)
	}
}

// withWriteEnabled enables WriteTextFile operations.
func withWriteEnabled(enabled bool) acpClientOption {
	return func(c *acpClient) {
		c.writeEnabled = enabled
	}
}

// withStreamWriter sets a writer for real-time output streaming.
func withStreamWriter(w io.Writer) acpClientOption {
	return func(c *acpClient) {
		c.streamWriter = w
	}
}

// acpClient implements the acp.Client interface for osmedeus agent-acp steps.
// It accumulates agent output text and auto-approves permission requests.
type acpClient struct {
	mu           sync.Mutex
	output       strings.Builder
	allowedPaths []string
	writeEnabled bool
	streamWriter io.Writer
}

var _ acp.Client = (*acpClient)(nil)

func newACPClient(opts ...acpClientOption) *acpClient {
	c := &acpClient{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// collectedOutput returns the accumulated agent text output.
func (c *acpClient) collectedOutput() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.output.String()
}

// SessionUpdate receives streaming updates from the agent.
// Agent message chunks are accumulated as output text.
func (c *acpClient) SessionUpdate(_ context.Context, n acp.SessionNotification) error {
	update := n.Update

	if update.AgentMessageChunk != nil {
		if update.AgentMessageChunk.Content.Text != nil {
			text := update.AgentMessageChunk.Content.Text.Text
			c.mu.Lock()
			c.output.WriteString(text)
			c.mu.Unlock()

			if c.streamWriter != nil {
				_, _ = io.WriteString(c.streamWriter, text)
			}
		}
	}

	if update.AgentThoughtChunk != nil {
		if update.AgentThoughtChunk.Content.Text != nil {
			oslogger.Get().Debug("acp agent thought", zap.String("text", update.AgentThoughtChunk.Content.Text.Text))
		}
	}

	if update.ToolCall != nil {
		fields := []zap.Field{
			zap.String("toolCallId", string(update.ToolCall.ToolCallId)),
			zap.String("title", update.ToolCall.Title),
			zap.String("status", string(update.ToolCall.Status)),
		}
		if update.ToolCall.Kind != "" {
			fields = append(fields, zap.String("kind", string(update.ToolCall.Kind)))
		}
		if len(update.ToolCall.Locations) > 0 {
			paths := make([]string, len(update.ToolCall.Locations))
			for i, loc := range update.ToolCall.Locations {
				paths[i] = loc.Path
			}
			fields = append(fields, zap.Strings("paths", paths))
		}
		oslogger.Get().Debug("acp agent tool call", fields...)
	}

	if update.ToolCallUpdate != nil {
		fields := []zap.Field{
			zap.String("toolCallId", string(update.ToolCallUpdate.ToolCallId)),
		}
		if update.ToolCallUpdate.Status != nil {
			fields = append(fields, zap.String("status", string(*update.ToolCallUpdate.Status)))
		}
		if update.ToolCallUpdate.Title != nil {
			fields = append(fields, zap.String("title", *update.ToolCallUpdate.Title))
		}
		oslogger.Get().Debug("acp agent tool call update", fields...)
	}

	if update.Plan != nil {
		entries := make([]string, len(update.Plan.Entries))
		for i, e := range update.Plan.Entries {
			entries[i] = fmt.Sprintf("[%s] %s", string(e.Status), e.Content)
		}
		oslogger.Get().Debug("acp agent plan update", zap.Strings("entries", entries))
	}

	return nil
}

// RequestPermission auto-approves agent permission requests by selecting the
// first allow_once or allow_always option. If no allow option is found, the
// first option is selected.
func (c *acpClient) RequestPermission(_ context.Context, p acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	toolTitle := ""
	if p.ToolCall.Title != nil {
		toolTitle = *p.ToolCall.Title
	}

	optionNames := make([]string, len(p.Options))
	for i, opt := range p.Options {
		optionNames[i] = fmt.Sprintf("%s(%s)", opt.Name, opt.Kind)
	}
	oslogger.Get().Debug("acp agent requested permission",
		zap.String("toolCallId", string(p.ToolCall.ToolCallId)),
		zap.String("toolTitle", toolTitle),
		zap.Strings("options", optionNames))

	// Prefer allow_once, then allow_always, then first option
	for _, opt := range p.Options {
		if opt.Kind == acp.PermissionOptionKindAllowOnce {
			oslogger.Get().Debug("acp auto-approved permission",
				zap.String("selected", opt.Name),
				zap.String("kind", string(opt.Kind)))
			return acp.RequestPermissionResponse{
				Outcome: acp.NewRequestPermissionOutcomeSelected(opt.OptionId),
			}, nil
		}
	}
	for _, opt := range p.Options {
		if opt.Kind == acp.PermissionOptionKindAllowAlways {
			oslogger.Get().Debug("acp auto-approved permission",
				zap.String("selected", opt.Name),
				zap.String("kind", string(opt.Kind)))
			return acp.RequestPermissionResponse{
				Outcome: acp.NewRequestPermissionOutcomeSelected(opt.OptionId),
			}, nil
		}
	}
	if len(p.Options) > 0 {
		oslogger.Get().Debug("acp auto-approved permission (fallback to first option)",
			zap.String("selected", p.Options[0].Name),
			zap.String("kind", string(p.Options[0].Kind)))
		return acp.RequestPermissionResponse{
			Outcome: acp.NewRequestPermissionOutcomeSelected(p.Options[0].OptionId),
		}, nil
	}
	oslogger.Get().Debug("acp permission request cancelled (no options available)")
	return acp.RequestPermissionResponse{
		Outcome: acp.NewRequestPermissionOutcomeCancelled(),
	}, nil
}

// ReadTextFile reads a file from disk, scoped to allowed paths.
func (c *acpClient) ReadTextFile(_ context.Context, p acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	absPath, err := filepath.Abs(p.Path)
	if err != nil {
		return acp.ReadTextFileResponse{}, fmt.Errorf("invalid path: %w", err)
	}

	if !c.isPathAllowed(absPath) {
		oslogger.Get().Debug("acp ReadTextFile denied (outside allowed paths)",
			zap.String("path", p.Path),
			zap.Strings("allowedPaths", c.allowedPaths))
		return acp.ReadTextFileResponse{}, fmt.Errorf("path %q is outside allowed directories", p.Path)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		oslogger.Get().Debug("acp ReadTextFile failed",
			zap.String("path", absPath),
			zap.Error(err))
		return acp.ReadTextFileResponse{}, err
	}

	content := string(data)

	// Apply Line/Limit slicing if requested
	if p.Line != nil || p.Limit != nil {
		lines := strings.Split(content, "\n")
		start := 0
		if p.Line != nil && *p.Line > 0 {
			start = *p.Line - 1
			if start > len(lines) {
				start = len(lines)
			}
		}
		end := len(lines)
		if p.Limit != nil && *p.Limit > 0 {
			if start+*p.Limit < end {
				end = start + *p.Limit
			}
		}
		content = strings.Join(lines[start:end], "\n")
	}

	oslogger.Get().Debug("acp ReadTextFile",
		zap.String("path", absPath),
		zap.Int("bytes", len(content)))

	return acp.ReadTextFileResponse{Content: content}, nil
}

// WriteTextFile writes a file to disk if write mode is enabled.
func (c *acpClient) WriteTextFile(_ context.Context, p acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	if !c.writeEnabled {
		oslogger.Get().Debug("acp WriteTextFile denied (write disabled)",
			zap.String("path", p.Path))
		return acp.WriteTextFileResponse{}, fmt.Errorf("file writes are disabled for this agent-acp step")
	}

	absPath, err := filepath.Abs(p.Path)
	if err != nil {
		return acp.WriteTextFileResponse{}, fmt.Errorf("invalid path: %w", err)
	}

	if !c.isPathAllowed(absPath) {
		oslogger.Get().Debug("acp WriteTextFile denied (outside allowed paths)",
			zap.String("path", p.Path),
			zap.Strings("allowedPaths", c.allowedPaths))
		return acp.WriteTextFileResponse{}, fmt.Errorf("path %q is outside allowed directories", p.Path)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return acp.WriteTextFileResponse{}, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(p.Content), 0o644); err != nil {
		return acp.WriteTextFileResponse{}, fmt.Errorf("failed to write file: %w", err)
	}

	oslogger.Get().Debug("acp WriteTextFile",
		zap.String("path", absPath),
		zap.Int("bytes", len(p.Content)))

	return acp.WriteTextFileResponse{}, nil
}

// CreateTerminal returns a stub terminal ID (no-op).
func (c *acpClient) CreateTerminal(_ context.Context, p acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	oslogger.Get().Debug("acp CreateTerminal (no-op)", zap.String("command", p.Command))
	return acp.CreateTerminalResponse{TerminalId: "osmedeus-stub-term"}, nil
}

// KillTerminalCommand is a no-op.
func (c *acpClient) KillTerminalCommand(_ context.Context, _ acp.KillTerminalCommandRequest) (acp.KillTerminalCommandResponse, error) {
	oslogger.Get().Debug("acp KillTerminalCommand (no-op)")
	return acp.KillTerminalCommandResponse{}, nil
}

// ReleaseTerminal is a no-op.
func (c *acpClient) ReleaseTerminal(_ context.Context, _ acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	oslogger.Get().Debug("acp ReleaseTerminal (no-op)")
	return acp.ReleaseTerminalResponse{}, nil
}

// TerminalOutput returns empty output (no-op).
func (c *acpClient) TerminalOutput(_ context.Context, _ acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	oslogger.Get().Debug("acp TerminalOutput (no-op)")
	return acp.TerminalOutputResponse{Output: "", Truncated: false}, nil
}

// WaitForTerminalExit returns exit code 0 (no-op).
func (c *acpClient) WaitForTerminalExit(_ context.Context, _ acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	oslogger.Get().Debug("acp WaitForTerminalExit (no-op)")
	exitCode := 0
	return acp.WaitForTerminalExitResponse{ExitCode: &exitCode}, nil
}

// isPathAllowed checks if a path is under one of the allowed directories.
// If no allowed paths are configured, all paths are allowed.
func (c *acpClient) isPathAllowed(absPath string) bool {
	if len(c.allowedPaths) == 0 {
		return true
	}
	for _, allowed := range c.allowedPaths {
		allowedAbs, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		prefix := allowedAbs + string(filepath.Separator)
		if absPath == allowedAbs || strings.HasPrefix(absPath, prefix) {
			return true
		}
	}
	return false
}

// zapSlogHandler routes slog records through zap at DEBUG level,
// suppressing the ACP SDK's default INFO logs.
type zapSlogHandler struct{}

func newZapSlogHandler() *zapSlogHandler { return &zapSlogHandler{} }

func (h *zapSlogHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *zapSlogHandler) Handle(_ context.Context, r slog.Record) error {
	oslogger.Get().Debug("acp-sdk: "+r.Message, zap.String("level", r.Level.String()))
	return nil
}

func (h *zapSlogHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *zapSlogHandler) WithGroup(_ string) slog.Handler      { return h }
