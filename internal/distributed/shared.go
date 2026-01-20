package distributed

import (
	"sync"

	"github.com/j3ssie/osmedeus/v5/internal/config"
)

var (
	sharedClient *Client
	sharedOnce   sync.Once
	sharedErr    error
)

// GetSharedClient returns a singleton Redis client for distributed mode.
// Returns nil if Redis is not configured.
func GetSharedClient() (*Client, error) {
	cfg := config.Get()
	if cfg == nil || !cfg.IsRedisConfigured() {
		return nil, nil
	}

	sharedOnce.Do(func() {
		sharedClient, sharedErr = NewClientFromConfig(cfg)
	})

	return sharedClient, sharedErr
}

// ResetSharedClient resets the shared client (useful for testing)
func ResetSharedClient() {
	if sharedClient != nil {
		sharedClient.Close()
	}
	sharedClient = nil
	sharedOnce = sync.Once{}
	sharedErr = nil
}

// SetSharedClient sets the shared client (useful for testing or custom initialization)
func SetSharedClient(client *Client) {
	sharedClient = client
	sharedOnce.Do(func() {}) // Mark as initialized
}
