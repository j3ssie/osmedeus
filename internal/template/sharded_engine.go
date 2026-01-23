package template

import (
	"fmt"
	"hash/fnv"
	"maps"
	"strings"
	"sync"

	"github.com/flosch/pongo2/v6"
	lru "github.com/hashicorp/golang-lru/v2"
)

func init() {
	// Disable HTML autoescape - osmedeus templates are for shell commands,
	// not HTML output. Escaping breaks JSON and other structured data.
	pongo2.SetAutoescape(false)
}

// DefaultShardCount is the default number of shards (must be power of 2)
const DefaultShardCount = 16

// DefaultShardCacheSize is the default cache size per shard
const DefaultShardCacheSize = 64

// ParallelShardThreshold is the minimum number of shards needed to justify
// parallel processing overhead. Below this threshold, sequential is faster.
const ParallelShardThreshold = 2

// ShardedEngineConfig holds configuration for the sharded engine
type ShardedEngineConfig struct {
	ShardCount     int  // Number of shards (must be power of 2)
	ShardCacheSize int  // Cache size per shard
	EnablePooling  bool // Use pooled context maps
}

// DefaultShardedEngineConfig returns the default configuration
func DefaultShardedEngineConfig() ShardedEngineConfig {
	return ShardedEngineConfig{
		ShardCount:     DefaultShardCount,
		ShardCacheSize: DefaultShardCacheSize,
		EnablePooling:  true,
	}
}

// EngineShard represents a single cache shard with its own lock
type EngineShard struct {
	mu    sync.RWMutex
	cache *lru.Cache[string, *pongo2.Template]
}

// ShardedEngine is a high-performance template engine using sharded caching.
// It distributes templates across multiple shards to reduce lock contention
// under high concurrency.
type ShardedEngine struct {
	shards        []*EngineShard
	shardMask     uint32 // Used for fast shard selection (shardCount - 1)
	generators    map[string]GeneratorFunc
	generatorsMu  sync.RWMutex
	enablePooling bool
}

// NewShardedEngine creates a new sharded template engine with default config
func NewShardedEngine() *ShardedEngine {
	return NewShardedEngineWithConfig(DefaultShardedEngineConfig())
}

// NewShardedEngineWithConfig creates a new sharded template engine with custom config
func NewShardedEngineWithConfig(cfg ShardedEngineConfig) *ShardedEngine {
	// Ensure shard count is power of 2
	shardCount := cfg.ShardCount
	if shardCount <= 0 {
		shardCount = DefaultShardCount
	}
	// Round up to next power of 2
	shardCount = nextPowerOf2(shardCount)

	cacheSize := cfg.ShardCacheSize
	if cacheSize <= 0 {
		cacheSize = DefaultShardCacheSize
	}

	shards := make([]*EngineShard, shardCount)
	for i := range shards {
		cache, _ := lru.New[string, *pongo2.Template](cacheSize)
		shards[i] = &EngineShard{
			cache: cache,
		}
	}

	e := &ShardedEngine{
		shards:        shards,
		shardMask:     uint32(shardCount - 1),
		generators:    make(map[string]GeneratorFunc),
		enablePooling: cfg.EnablePooling,
	}
	e.registerBuiltinGenerators()
	return e
}

// nextPowerOf2 returns the next power of 2 >= n
func nextPowerOf2(n int) int {
	if n <= 1 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	return n + 1
}

// getShard returns the shard for the given template string using FNV-1a hash
func (e *ShardedEngine) getShard(template string) *EngineShard {
	h := fnv.New32a()
	h.Write([]byte(template))
	idx := h.Sum32() & e.shardMask
	return e.shards[idx]
}

// Render renders a template string with the given context.
// Uses sharded cache with RWMutex for improved concurrency.
func (e *ShardedEngine) Render(template string, ctx map[string]any) (string, error) {
	// Quick path: no template variables present
	if !strings.Contains(template, "{{") {
		return template, nil
	}

	shard := e.getShard(template)

	// Try cache lookup with read lock first
	shard.mu.RLock()
	tpl, ok := shard.cache.Get(template)
	shard.mu.RUnlock()

	if !ok {
		// Cache miss - need to parse and cache
		shard.mu.Lock()
		// Double-check after acquiring write lock
		tpl, ok = shard.cache.Get(template)
		if !ok {
			var err error
			tpl, err = pongo2.FromString(template)
			if err != nil {
				shard.mu.Unlock()
				return "", fmt.Errorf("template parse error: %w", err)
			}
			shard.cache.Add(template, tpl)
		}
		shard.mu.Unlock()
	}

	// Execute template OUTSIDE the lock (pongo2 templates are thread-safe once parsed)
	var processedCtx map[string]any
	if e.enablePooling {
		processedCtx = NormalizeBoolsToPooled(ctx)
		defer PutContext(processedCtx)
	} else {
		processedCtx = normalizeBoolsForTemplate(ctx)
	}

	result, err := tpl.Execute(pongo2.Context(processedCtx))
	if err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}

	// Fix incomplete expressions caused by undefined variables
	result = fixIncompleteExpressions(result)

	return result, nil
}

// RenderMap renders all template values in a map
func (e *ShardedEngine) RenderMap(m map[string]string, ctx map[string]any) (map[string]string, error) {
	result := make(map[string]string, len(m))
	for k, v := range m {
		rendered, err := e.Render(v, ctx)
		if err != nil {
			return nil, fmt.Errorf("error rendering %s: %w", k, err)
		}
		result[k] = rendered
	}
	return result, nil
}

// RenderSlice renders all template values in a slice
func (e *ShardedEngine) RenderSlice(s []string, ctx map[string]any) ([]string, error) {
	result := make([]string, len(s))
	for i, v := range s {
		rendered, err := e.Render(v, ctx)
		if err != nil {
			return nil, fmt.Errorf("error rendering index %d: %w", i, err)
		}
		result[i] = rendered
	}
	return result, nil
}

// RenderSecondary renders templates using [[ ]] delimiters
func (e *ShardedEngine) RenderSecondary(template string, ctx map[string]any) (string, error) {
	// Quick path: no secondary delimiters
	if !strings.Contains(template, "[[") {
		return template, nil
	}

	// Convert [[ ]] to {{ }} for pongo2 processing
	converted := strings.ReplaceAll(template, "[[", "{{")
	converted = strings.ReplaceAll(converted, "]]", "}}")

	return e.Render(converted, ctx)
}

// HasSecondaryVariable checks if template contains [[ ]] delimiters
func (e *ShardedEngine) HasSecondaryVariable(template string) bool {
	return strings.Contains(template, "[[") && strings.Contains(template, "]]")
}

// ExecuteGenerator executes a generator function expression
func (e *ShardedEngine) ExecuteGenerator(expr string) (string, error) {
	name, args, err := e.parseGeneratorExpr(expr)
	if err != nil {
		return "", err
	}

	e.generatorsMu.RLock()
	gen, ok := e.generators[name]
	e.generatorsMu.RUnlock()

	if !ok {
		return "", fmt.Errorf("unknown generator function: %s", name)
	}

	return gen(args...)
}

// RegisterGenerator registers a custom generator function
func (e *ShardedEngine) RegisterGenerator(name string, fn GeneratorFunc) {
	e.generatorsMu.Lock()
	e.generators[name] = fn
	e.generatorsMu.Unlock()
}

// parseGeneratorExpr parses a generator expression into function name and arguments
func (e *ShardedEngine) parseGeneratorExpr(expr string) (string, []string, error) {
	matches := generatorExprPattern.FindStringSubmatch(strings.TrimSpace(expr))
	if len(matches) != 3 {
		return "", nil, fmt.Errorf("invalid generator expression: %s", expr)
	}

	funcName := matches[1]
	argsStr := strings.TrimSpace(matches[2])

	if argsStr == "" {
		return funcName, nil, nil
	}

	args := e.parseArgs(argsStr)
	return funcName, args, nil
}

// parseArgs parses comma-separated arguments, handling quoted strings
func (e *ShardedEngine) parseArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range argsStr {
		switch {
		case (ch == '"' || ch == '\'') && !inQuote:
			inQuote = true
			quoteChar = ch
		case ch == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
		case ch == ',' && !inQuote:
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}

	return args
}

// registerBuiltinGenerators registers all built-in generator functions
func (e *ShardedEngine) registerBuiltinGenerators() {
	maps.Copy(e.generators, builtinGenerators)
}

// RenderBatch renders multiple templates in a single operation.
// Templates are grouped by shard to minimize lock acquisitions.
func (e *ShardedEngine) RenderBatch(requests []RenderRequest, ctx map[string]any) (map[string]string, error) {
	if len(requests) == 0 {
		return make(map[string]string), nil
	}

	results := make(map[string]string, len(requests))

	// Prepare normalized context once
	var processedCtx map[string]any
	if e.enablePooling {
		processedCtx = NormalizeBoolsToPooled(ctx)
		defer PutContext(processedCtx)
	} else {
		processedCtx = normalizeBoolsForTemplate(ctx)
	}

	// Group requests by shard
	shardGroups := make(map[uint32][]RenderRequest)
	for _, req := range requests {
		// Quick path for templates without variables
		if !strings.Contains(req.Template, "{{") {
			results[req.Key] = req.Template
			continue
		}
		h := fnv.New32a()
		h.Write([]byte(req.Template))
		idx := h.Sum32() & e.shardMask
		shardGroups[idx] = append(shardGroups[idx], req)
	}

	// Use parallel processing when multiple shards have work (20-40% faster startup)
	if len(shardGroups) >= ParallelShardThreshold {
		return e.renderShardGroupsParallel(shardGroups, processedCtx, results)
	}

	// Process each shard group sequentially
	for idx, reqs := range shardGroups {
		shard := e.shards[idx]
		if err := e.renderShardBatch(shard, reqs, processedCtx, results); err != nil {
			return nil, err
		}
	}

	return results, nil
}

// renderShardGroupsParallel processes multiple shard groups concurrently.
// Each shard is processed in its own goroutine, with results merged at the end.
// This provides 20-40% faster workflow startup when multiple shards have work.
func (e *ShardedEngine) renderShardGroupsParallel(groups map[uint32][]RenderRequest, ctx map[string]any, results map[string]string) (map[string]string, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstErr error

	for idx, reqs := range groups {
		wg.Add(1)
		go func(shardIdx uint32, requests []RenderRequest) {
			defer wg.Done()

			shard := e.shards[shardIdx]
			localResults := make(map[string]string, len(requests))

			if err := e.renderShardBatch(shard, requests, ctx, localResults); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			maps.Copy(results, localResults)
			mu.Unlock()
		}(idx, reqs)
	}

	wg.Wait()
	return results, firstErr
}

// renderShardBatch renders all templates for a single shard
func (e *ShardedEngine) renderShardBatch(shard *EngineShard, reqs []RenderRequest, ctx map[string]any, results map[string]string) error {
	// Collect templates that need parsing
	var toParse []RenderRequest
	templates := make(map[string]*pongo2.Template, len(reqs))

	// First pass: check cache with read lock
	shard.mu.RLock()
	for _, req := range reqs {
		if tpl, ok := shard.cache.Get(req.Template); ok {
			templates[req.Template] = tpl
		} else {
			toParse = append(toParse, req)
		}
	}
	shard.mu.RUnlock()

	// Second pass: parse and cache misses with write lock
	if len(toParse) > 0 {
		shard.mu.Lock()
		for _, req := range toParse {
			// Double-check after acquiring write lock
			if tpl, ok := shard.cache.Get(req.Template); ok {
				templates[req.Template] = tpl
				continue
			}
			tpl, err := pongo2.FromString(req.Template)
			if err != nil {
				shard.mu.Unlock()
				return fmt.Errorf("template parse error for %s: %w", req.Key, err)
			}
			shard.cache.Add(req.Template, tpl)
			templates[req.Template] = tpl
		}
		shard.mu.Unlock()
	}

	// Execute all templates outside locks
	for _, req := range reqs {
		tpl := templates[req.Template]
		result, err := tpl.Execute(pongo2.Context(ctx))
		if err != nil {
			return fmt.Errorf("template execute error for %s: %w", req.Key, err)
		}
		results[req.Key] = fixIncompleteExpressions(result)
	}

	return nil
}

// CacheStats returns statistics about the template cache
func (e *ShardedEngine) CacheStats() map[string]int {
	total := 0
	for _, shard := range e.shards {
		shard.mu.RLock()
		total += shard.cache.Len()
		shard.mu.RUnlock()
	}
	return map[string]int{
		"total_cached": total,
		"shard_count":  len(e.shards),
	}
}

// Verify interface compliance at compile time
var _ TemplateEngine = (*ShardedEngine)(nil)
var _ BatchRenderer = (*ShardedEngine)(nil)
