package functions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/runner"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

// sshExpandPath expands ~ to the user's home directory
func sshExpandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// sshConfig holds parsed SSH connection parameters
type sshConfig struct {
	User     string
	KeyFile  string
	Password string
	Port     int
}

// parseSSHConfig extracts optional SSH parameters from goja function call arguments
// starting at paramOffset. Parameters are: user, key_path, password, port
// Returns defaults: user="root", port=22
func parseSSHConfig(call goja.FunctionCall, paramOffset int) sshConfig {
	cfg := sshConfig{
		User: "root",
		Port: 22,
	}

	if len(call.Arguments) > paramOffset {
		v := call.Argument(paramOffset).String()
		if v != "undefined" && v != "" {
			cfg.User = v
		}
	}
	if len(call.Arguments) > paramOffset+1 {
		v := call.Argument(paramOffset + 1).String()
		if v != "undefined" && v != "" {
			cfg.KeyFile = v
		}
	}
	if len(call.Arguments) > paramOffset+2 {
		v := call.Argument(paramOffset + 2).String()
		if v != "undefined" && v != "" {
			cfg.Password = v
		}
	}
	if len(call.Arguments) > paramOffset+3 {
		v := call.Argument(paramOffset + 3)
		if portVal, ok := v.Export().(int64); ok && portVal > 0 {
			cfg.Port = int(portVal)
		} else if portFloat, ok := v.Export().(float64); ok && portFloat > 0 {
			cfg.Port = int(portFloat)
		}
	}

	return cfg
}

// sshExec executes a command on a remote host via SSH using the connection pool
// Usage: ssh_exec(host, command, user?, key_path?, password?, port?) -> string
func (vf *vmFunc) sshExec(call goja.FunctionCall) goja.Value {
	host := call.Argument(0).String()
	command := call.Argument(1).String()

	logger.Get().Debug("Calling "+terminal.HiGreen(FnSSHExec),
		zap.String("host", host), zap.Int("commandLength", len(command)))

	if host == "undefined" || host == "" {
		logger.Get().Warn(FnSSHExec + ": host is required")
		return vf.vm.ToValue("")
	}
	if command == "undefined" || command == "" {
		logger.Get().Warn(FnSSHExec + ": command is required")
		return vf.vm.ToValue("")
	}

	// Parse optional SSH config (user, key_path, password, port) starting at arg index 2
	cfg := parseSSHConfig(call, 2)

	// Build RunnerConfig for SSH pool
	runnerConfig := &core.RunnerConfig{
		Host:     host,
		Port:     cfg.Port,
		User:     cfg.User,
		KeyFile:  sshExpandPath(cfg.KeyFile),
		Password: cfg.Password,
	}

	// Use a 5-minute timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get pooled SSH connection
	pool := runner.GetSSHPool()
	client, poolKey, err := pool.Get(ctx, runnerConfig)
	if err != nil {
		logger.Get().Warn(FnSSHExec+": SSH connection failed",
			zap.String("host", host), zap.Error(err))
		return vf.vm.ToValue("")
	}
	defer pool.Release(poolKey)

	// Create session and run command
	session, err := client.NewSession()
	if err != nil {
		logger.Get().Warn(FnSSHExec+": failed to create SSH session",
			zap.String("host", host), zap.Error(err))
		return vf.vm.ToValue("")
	}
	defer func() { _ = session.Close() }()

	output, err := session.CombinedOutput(command)
	if err != nil {
		logger.Get().Warn(FnSSHExec+": command execution failed",
			zap.String("host", host), zap.String("command", command), zap.Error(err))
		// Return partial output even on error (non-zero exit code)
		return vf.vm.ToValue(strings.TrimSpace(string(output)))
	}

	result := strings.TrimSpace(string(output))
	logger.Get().Debug(terminal.HiGreen(FnSSHExec)+" result",
		zap.String("host", host), zap.Int("outputLength", len(result)))
	return vf.vm.ToValue(result)
}

// sshRsync copies a local file/directory to a remote host via rsync over SSH
// Usage: ssh_rsync(host, src, dest, user?, key_path?, password?, port?) -> bool
func (vf *vmFunc) sshRsync(call goja.FunctionCall) goja.Value {
	host := call.Argument(0).String()
	src := call.Argument(1).String()
	dest := call.Argument(2).String()

	logger.Get().Debug("Calling "+terminal.HiGreen(FnSSHRsync),
		zap.String("host", host), zap.String("src", src), zap.String("dest", dest))

	if host == "undefined" || host == "" {
		logger.Get().Warn(FnSSHRsync + ": host is required")
		return vf.vm.ToValue(false)
	}
	if src == "undefined" || src == "" {
		logger.Get().Warn(FnSSHRsync + ": src is required")
		return vf.vm.ToValue(false)
	}
	if dest == "undefined" || dest == "" {
		logger.Get().Warn(FnSSHRsync + ": dest is required")
		return vf.vm.ToValue(false)
	}

	// Parse optional SSH config (user, key_path, password, port) starting at arg index 3
	cfg := parseSSHConfig(call, 3)

	// Expand local source path
	src = sshExpandPath(src)

	// Build rsync destination: user@host:dest
	rsyncDest := fmt.Sprintf("%s@%s:%s", cfg.User, host, dest)

	// Use a 5-minute timeout context for rsync
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build rsync command with SSH options
	// Add ConnectTimeout to prevent SSH from hanging on unreachable hosts
	args := []string{"-avz", "-e"}

	keyPath := sshExpandPath(cfg.KeyFile)
	if cfg.Password != "" && cfg.KeyFile == "" {
		// Password-only auth: use sshpass wrapper
		sshCmd := fmt.Sprintf("ssh -p %d -o StrictHostKeyChecking=no -o ConnectTimeout=30", cfg.Port)
		args = append(args, sshCmd, src, rsyncDest)

		// Wrap with sshpass
		sshpassArgs := append([]string{"-p", cfg.Password, "rsync"}, args...)
		cmd := exec.CommandContext(ctx, "sshpass", sshpassArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Get().Warn(FnSSHRsync+": rsync with sshpass failed",
				zap.String("host", host), zap.Error(err), zap.String("output", string(output)))
			return vf.vm.ToValue(false)
		}
	} else {
		// Key-based auth (or key + password)
		if keyPath != "" {
			args = append(args, fmt.Sprintf("ssh -i %s -p %d -o StrictHostKeyChecking=no -o ConnectTimeout=30", keyPath, cfg.Port))
		} else {
			args = append(args, fmt.Sprintf("ssh -p %d -o StrictHostKeyChecking=no -o ConnectTimeout=30", cfg.Port))
		}
		args = append(args, src, rsyncDest)

		cmd := exec.CommandContext(ctx, "rsync", args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			logger.Get().Warn(FnSSHRsync+": rsync failed",
				zap.String("host", host), zap.Error(err), zap.String("output", string(output)))
			return vf.vm.ToValue(false)
		}
	}

	logger.Get().Debug(terminal.HiGreen(FnSSHRsync)+" completed",
		zap.String("host", host), zap.String("src", src), zap.String("dest", dest))
	return vf.vm.ToValue(true)
}
