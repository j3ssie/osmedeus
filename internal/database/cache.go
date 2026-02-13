package database

import (
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
)

// CacheConfig holds cache configuration
type CacheConfig struct {
	MaxCost     int64         // Maximum cost (memory) for cache
	NumCounters int64         // Number of counters for admission
	BufferItems int64         // Number of keys per Get buffer
	TTL         time.Duration // Default TTL for cached items
}

// DefaultCacheConfig returns sensible defaults
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		MaxCost:     8 << 20,         // 8MB (sufficient for thousands of ~1KB workflow entries)
		NumCounters: 1e5,             // 100K counters (ristretto recommends 10x expected items)
		BufferItems: 64,              // 64 keys per buffer
		TTL:         5 * time.Minute, // 5 minute TTL
	}
}

// Cache wraps ristretto cache with type-safe methods
type Cache struct {
	cache *ristretto.Cache
	ttl   time.Duration
}

var (
	globalCache *Cache
	cacheOnce   sync.Once
	cacheMu     sync.RWMutex
)

// Key prefixes for different cache types
const (
	keyPrefixWorkflowMeta    = "wf:"     // wf:{name}
	keyPrefixTechSummary     = "tech:"   // tech:{workspace}
	keyPrefixStatusSummary   = "status:" // status:{workspace}
	keyPrefixSeveritySummary = "sev:"    // sev:{workspace}
)

// Summary cache TTL (shorter than workflow meta since stats change more frequently)
const summaryCacheTTL = 2 * time.Minute

// CachedSummary holds cached summary data with timestamp
type CachedSummary struct {
	Data      map[string]int `json:"data"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// InitCache initializes the global cache
func InitCache(cfg *CacheConfig) error {
	var initErr error

	cacheOnce.Do(func() {
		if cfg == nil {
			cfg = DefaultCacheConfig()
		}

		cache, err := ristretto.NewCache(&ristretto.Config{
			NumCounters: cfg.NumCounters,
			MaxCost:     cfg.MaxCost,
			BufferItems: cfg.BufferItems,
		})
		if err != nil {
			initErr = err
			return
		}

		globalCache = &Cache{
			cache: cache,
			ttl:   cfg.TTL,
		}
	})

	return initErr
}

// GetCache returns the global cache instance
func GetCache() *Cache {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return globalCache
}

// GetWorkflowMeta retrieves WorkflowMeta from cache
func (c *Cache) GetWorkflowMeta(name string) (*WorkflowMeta, bool) {
	if c == nil || c.cache == nil {
		return nil, false
	}

	key := keyPrefixWorkflowMeta + name
	value, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	meta, ok := value.(*WorkflowMeta)
	if !ok {
		return nil, false
	}

	return meta, true
}

// SetWorkflowMeta stores WorkflowMeta in cache
func (c *Cache) SetWorkflowMeta(name string, meta *WorkflowMeta) {
	if c == nil || c.cache == nil || meta == nil {
		return
	}

	key := keyPrefixWorkflowMeta + name
	// Cost is estimated as 1KB per workflow meta entry
	cost := int64(1024)
	c.cache.SetWithTTL(key, meta, cost, c.ttl)
}

// InvalidateWorkflowMeta removes a workflow from cache
func (c *Cache) InvalidateWorkflowMeta(name string) {
	if c == nil || c.cache == nil {
		return
	}

	key := keyPrefixWorkflowMeta + name
	c.cache.Del(key)
}

// InvalidateAllWorkflows clears all workflow entries
// Note: ristretto doesn't support prefix-based deletion,
// so we clear the entire cache
func (c *Cache) InvalidateAllWorkflows() {
	if c == nil || c.cache == nil {
		return
	}

	c.cache.Clear()
}

// ============================================================================
// Tech Summary Cache
// ============================================================================

// GetTechSummary retrieves cached tech summary for a workspace
func (c *Cache) GetTechSummary(workspace string) (map[string]int, bool) {
	if c == nil || c.cache == nil {
		return nil, false
	}

	key := keyPrefixTechSummary + workspace
	if val, found := c.cache.Get(key); found {
		if summary, ok := val.(*CachedSummary); ok {
			return summary.Data, true
		}
	}
	return nil, false
}

// SetTechSummary stores tech summary with shorter TTL
func (c *Cache) SetTechSummary(workspace string, data map[string]int) {
	if c == nil || c.cache == nil || data == nil {
		return
	}

	key := keyPrefixTechSummary + workspace
	summary := &CachedSummary{Data: data, UpdatedAt: time.Now()}
	// Cost estimated as 64 bytes per entry
	cost := int64(len(data) * 64)
	if cost < 256 {
		cost = 256 // Minimum cost
	}
	c.cache.SetWithTTL(key, summary, cost, summaryCacheTTL)
}

// ============================================================================
// Status Summary Cache
// ============================================================================

// GetStatusSummary retrieves cached status summary for a workspace
func (c *Cache) GetStatusSummary(workspace string) (map[string]int, bool) {
	if c == nil || c.cache == nil {
		return nil, false
	}

	key := keyPrefixStatusSummary + workspace
	if val, found := c.cache.Get(key); found {
		if summary, ok := val.(*CachedSummary); ok {
			return summary.Data, true
		}
	}
	return nil, false
}

// SetStatusSummary stores status summary with shorter TTL
func (c *Cache) SetStatusSummary(workspace string, data map[string]int) {
	if c == nil || c.cache == nil || data == nil {
		return
	}

	key := keyPrefixStatusSummary + workspace
	summary := &CachedSummary{Data: data, UpdatedAt: time.Now()}
	cost := int64(len(data) * 64)
	if cost < 256 {
		cost = 256
	}
	c.cache.SetWithTTL(key, summary, cost, summaryCacheTTL)
}

// ============================================================================
// Severity Summary Cache
// ============================================================================

// GetSeveritySummary retrieves cached severity summary for a workspace
func (c *Cache) GetSeveritySummary(workspace string) (map[string]int, bool) {
	if c == nil || c.cache == nil {
		return nil, false
	}

	key := keyPrefixSeveritySummary + workspace
	if val, found := c.cache.Get(key); found {
		if summary, ok := val.(*CachedSummary); ok {
			return summary.Data, true
		}
	}
	return nil, false
}

// SetSeveritySummary stores severity summary with shorter TTL
func (c *Cache) SetSeveritySummary(workspace string, data map[string]int) {
	if c == nil || c.cache == nil || data == nil {
		return
	}

	key := keyPrefixSeveritySummary + workspace
	summary := &CachedSummary{Data: data, UpdatedAt: time.Now()}
	cost := int64(len(data) * 64)
	if cost < 256 {
		cost = 256
	}
	c.cache.SetWithTTL(key, summary, cost, summaryCacheTTL)
}

// ============================================================================
// Workspace Invalidation
// ============================================================================

// InvalidateWorkspace clears all cached data for a workspace.
// Call this when workspace data changes (new assets, vulnerabilities, etc.)
func (c *Cache) InvalidateWorkspace(workspace string) {
	if c == nil || c.cache == nil {
		return
	}

	c.cache.Del(keyPrefixTechSummary + workspace)
	c.cache.Del(keyPrefixStatusSummary + workspace)
	c.cache.Del(keyPrefixSeveritySummary + workspace)
}

// InvalidateWorkspaceSummaries clears all summary caches for a workspace.
// Alias for InvalidateWorkspace for semantic clarity.
func (c *Cache) InvalidateWorkspaceSummaries(workspace string) {
	c.InvalidateWorkspace(workspace)
}

// Close closes the cache and releases resources
func (c *Cache) Close() {
	if c == nil || c.cache == nil {
		return
	}

	c.cache.Close()
}

// closeGlobalCache closes the global cache instance
func closeGlobalCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	if globalCache != nil {
		globalCache.Close()
		globalCache = nil
	}

	// Reset the sync.Once to allow re-initialization
	cacheOnce = sync.Once{}
}
