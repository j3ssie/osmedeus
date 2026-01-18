package runner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
	"github.com/j3ssie/osmedeus/v5/internal/retry"
	"golang.org/x/crypto/ssh"
)

// SSHPoolKey uniquely identifies an SSH connection target
type SSHPoolKey struct {
	Host string
	Port int
	User string
}

// String returns a string representation of the pool key
func (k SSHPoolKey) String() string {
	return fmt.Sprintf("%s@%s:%d", k.User, k.Host, k.Port)
}

// pooledConnection holds a pooled SSH connection with metadata
type pooledConnection struct {
	client   *ssh.Client
	key      SSHPoolKey
	lastUsed time.Time
	refCount int32
}

// SSHPool manages a pool of SSH connections for reuse
type SSHPool struct {
	mu          sync.Mutex
	connections map[SSHPoolKey]*pooledConnection
	idleTimeout time.Duration
	stopCleanup chan struct{}
	cleanupOnce sync.Once
}

var (
	globalSSHPool *SSHPool
	poolOnce      sync.Once
)

// GetSSHPool returns the global SSH connection pool
func GetSSHPool() *SSHPool {
	poolOnce.Do(func() {
		globalSSHPool = &SSHPool{
			connections: make(map[SSHPoolKey]*pooledConnection),
			idleTimeout: 5 * time.Minute,
			stopCleanup: make(chan struct{}),
		}
		go globalSSHPool.cleanupLoop()
	})
	return globalSSHPool
}

// cleanupLoop periodically removes idle connections
func (p *SSHPool) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.cleanupIdle()
		case <-p.stopCleanup:
			return
		}
	}
}

// cleanupIdle removes connections that have been idle too long
func (p *SSHPool) cleanupIdle() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for key, conn := range p.connections {
		if atomic.LoadInt32(&conn.refCount) == 0 && now.Sub(conn.lastUsed) > p.idleTimeout {
			_ = conn.client.Close()
			delete(p.connections, key)
		}
	}
}

// Get retrieves or creates an SSH connection for the given config
func (p *SSHPool) Get(ctx context.Context, config *core.RunnerConfig) (*ssh.Client, SSHPoolKey, error) {
	port := config.Port
	if port == 0 {
		port = 22
	}

	key := SSHPoolKey{
		Host: config.Host,
		Port: port,
		User: config.User,
	}

	p.mu.Lock()

	// Check if we have an existing connection
	if conn, ok := p.connections[key]; ok {
		// Verify connection is still alive
		if _, _, err := conn.client.SendRequest("keepalive@openssh.org", true, nil); err == nil {
			atomic.AddInt32(&conn.refCount, 1)
			conn.lastUsed = time.Now()
			p.mu.Unlock()
			return conn.client, key, nil
		}
		// Connection is dead, remove it
		_ = conn.client.Close()
		delete(p.connections, key)
	}

	// Need to create new connection - release lock during dial
	p.mu.Unlock()

	// Build authentication methods
	authMethods, err := buildAuthMethods(config)
	if err != nil {
		return nil, key, err
	}

	// Build SSH config
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect with retry for transient network errors
	addr := fmt.Sprintf("%s:%d", config.Host, port)
	var client *ssh.Client
	err = retry.Do(ctx, retry.Config{
		MaxAttempts:  3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}, func() error {
		var dialErr error
		client, dialErr = ssh.Dial("tcp", addr, sshConfig)
		if dialErr != nil {
			// Network errors are retryable (connection refused, timeout, etc.)
			if isNetworkError(dialErr) {
				return retry.Retryable(dialErr)
			}
			return dialErr
		}
		return nil
	})
	if err != nil {
		return nil, key, fmt.Errorf("SSH connection failed to %s: %w", addr, err)
	}

	// Store in pool
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check someone else didn't create one while we were dialing
	if existing, ok := p.connections[key]; ok {
		// Use existing, close the one we just created
		_ = client.Close()
		atomic.AddInt32(&existing.refCount, 1)
		existing.lastUsed = time.Now()
		return existing.client, key, nil
	}

	// Store our new connection
	p.connections[key] = &pooledConnection{
		client:   client,
		key:      key,
		lastUsed: time.Now(),
		refCount: 1,
	}

	return client, key, nil
}

// Release decrements the reference count for a connection
func (p *SSHPool) Release(key SSHPoolKey) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, ok := p.connections[key]; ok {
		atomic.AddInt32(&conn.refCount, -1)
		conn.lastUsed = time.Now()
	}
}

// CloseAll closes all pooled connections
func (p *SSHPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for key, conn := range p.connections {
		_ = conn.client.Close()
		delete(p.connections, key)
	}
}

// Stop stops the cleanup goroutine
func (p *SSHPool) Stop() {
	p.cleanupOnce.Do(func() {
		close(p.stopCleanup)
	})
}

// buildAuthMethods builds SSH authentication methods from config
func buildAuthMethods(config *core.RunnerConfig) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	// Try key file first
	if config.KeyFile != "" {
		keyPath := expandPath(config.KeyFile)
		key, err := os.ReadFile(keyPath)
		if err == nil {
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	// Add password authentication if provided
	if config.Password != "" {
		authMethods = append(authMethods, ssh.Password(config.Password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no SSH authentication method available (provide key_file or password)")
	}

	return authMethods, nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// isNetworkError checks if an error is a transient network error that should be retried
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check for net.Error interface (timeout errors)
	var netErr net.Error
	if ok := errors.As(err, &netErr); ok && netErr.Timeout() {
		return true
	}

	// Check for common network error patterns in error messages
	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"no route to host",
		"network is unreachable",
		"i/o timeout",
		"temporary failure",
	}
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
