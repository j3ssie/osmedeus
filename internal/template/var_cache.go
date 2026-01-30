package template

import (
	"sync"
)

// VarRefCache caches variable references extracted from template strings.
// This avoids repeated regex parsing for the same template patterns.
type VarRefCache struct {
	cache map[string]map[string]struct{}
	mu    sync.RWMutex
	size  int
}

// NewVarRefCache creates a cache with the specified maximum size
func NewVarRefCache(maxSize int) *VarRefCache {
	if maxSize <= 0 {
		maxSize = 1024
	}
	return &VarRefCache{
		cache: make(map[string]map[string]struct{}, maxSize),
		size:  maxSize,
	}
}

// Get returns cached variable references for a template, or nil if not cached
func (c *VarRefCache) Get(template string) (map[string]struct{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	refs, ok := c.cache[template]
	return refs, ok
}

// Set caches variable references for a template
func (c *VarRefCache) Set(template string, refs map[string]struct{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction: clear when at capacity
	if len(c.cache) >= c.size {
		c.cache = make(map[string]map[string]struct{}, c.size)
	}

	c.cache[template] = refs
}

// GetOrExtract returns cached refs or extracts and caches them
func (c *VarRefCache) GetOrExtract(template string, extractor func(string) map[string]struct{}) map[string]struct{} {
	if refs, ok := c.Get(template); ok {
		return refs
	}

	refs := extractor(template)
	c.Set(template, refs)
	return refs
}
