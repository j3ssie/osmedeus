package functions

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/j3ssie/osmedeus/v5/internal/logger"
	"github.com/j3ssie/osmedeus/v5/internal/terminal"
	"go.uber.org/zap"
)

var (
	sudoKeepaliveMu     sync.Mutex
	sudoKeepaliveCancel context.CancelFunc
	sudoKeepaliveActive bool
)

// AuthenticateSudo validates or refreshes the sudo credential cache.
// An empty password triggers a TTY prompt (stdin passthrough).
// A non-empty password is piped via stdin. The password is never logged.
func AuthenticateSudo(password string) bool {
	// #nosec G204 -- sudo -S -v is a fixed command used intentionally
	// to validate/refresh sudo credentials. No user input is interpolated
	// into the command itself.
	cmd := exec.Command("sudo", "-S", "-v")

	if password == "" {
		// TTY prompt: let the user type the password interactively
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
	} else {
		// Pipe the password via stdin
		pipe, err := cmd.StdinPipe()
		if err != nil {
			logger.Get().Warn("sudo_auth: failed to create stdin pipe", zap.Error(err))
			return false
		}
		go func() {
			defer func() { _ = pipe.Close() }()
			_, _ = io.WriteString(pipe, password+"\n")
		}()
	}

	if err := cmd.Run(); err != nil {
		logger.Get().Warn("sudo_auth: authentication failed", zap.Error(err))
		return false
	}
	return true
}

// StartSudoKeepalive spawns a background goroutine that refreshes
// sudo credentials every 4 minutes. It is a singleton — calling it
// multiple times is safe; only the first call has any effect.
func StartSudoKeepalive(password string) {
	sudoKeepaliveMu.Lock()
	defer sudoKeepaliveMu.Unlock()

	if sudoKeepaliveActive {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	sudoKeepaliveCancel = cancel
	sudoKeepaliveActive = true

	go func() {
		ticker := time.NewTicker(4 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = AuthenticateSudo(password)
			}
		}
	}()
}

// StopSudoKeepalive cancels the keepalive goroutine.
// Safe to call multiple times or when no keepalive is running.
func StopSudoKeepalive() {
	sudoKeepaliveMu.Lock()
	defer sudoKeepaliveMu.Unlock()

	if sudoKeepaliveCancel != nil {
		sudoKeepaliveCancel()
		sudoKeepaliveCancel = nil
	}
	sudoKeepaliveActive = false
}

// sudoAuth is the Goja wrapper exposed as sudo_auth() in workflow YAML.
//
// Usage:
//
//	sudo_auth()                     -> bool (TTY prompt, no keepalive)
//	sudo_auth(password)             -> bool (pipe password, no keepalive)
//	sudo_auth(password, keepalive)  -> bool (pipe password, start keepalive if true)
func (vf *vmFunc) sudoAuth(call goja.FunctionCall) goja.Value {
	logger.Get().Debug("Calling " + terminal.HiGreen("sudoAuth"))

	password := ""
	if len(call.Arguments) > 0 {
		p := strings.TrimSpace(call.Argument(0).String())
		if p != "undefined" {
			password = p
		}
	}

	keepalive := false
	if len(call.Arguments) > 1 {
		keepalive = call.Argument(1).ToBoolean()
	}

	ok := AuthenticateSudo(password)
	if ok && keepalive {
		StartSudoKeepalive(password)
	}

	return vf.vm.ToValue(ok)
}
