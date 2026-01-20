package template

import (
	"maps"
	"sync"
)

// DefaultContextSize is the default pre-allocated size for context maps.
// This is based on typical workflow context sizes (Target, Output, threads, etc.)
const DefaultContextSize = 64

// contextPool is a sync.Pool for reusing context maps.
// This reduces GC pressure by avoiding repeated map allocations during rendering.
var contextPool = sync.Pool{
	New: func() any {
		return make(map[string]any, DefaultContextSize)
	},
}

// GetContext retrieves a context map from the pool.
// The returned map is empty but pre-allocated with DefaultContextSize capacity.
// Caller must call PutContext when done to return it to the pool.
func GetContext() map[string]any {
	return contextPool.Get().(map[string]any)
}

// PutContext clears the map and returns it to the pool.
// The map should not be used after calling this function.
func PutContext(ctx map[string]any) {
	if ctx == nil {
		return
	}
	// Clear the map for reuse
	clear(ctx)
	contextPool.Put(ctx)
}

// CloneToPooled copies the source map into a pooled map.
// Returns a new map from the pool with all key-value pairs from src.
// Caller must call PutContext when done with the returned map.
func CloneToPooled(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := GetContext()
	maps.Copy(dst, src)
	return dst
}

// NormalizeBoolsToPooled normalizes bool values and returns a pooled map.
// Bool values are converted to lowercase strings ("true"/"false") for
// consistent template output (pongo2 outputs "True"/"False" by default).
// Caller must call PutContext when done with the returned map.
func NormalizeBoolsToPooled(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := GetContext()
	for k, v := range src {
		if b, ok := v.(bool); ok {
			if b {
				dst[k] = "true"
			} else {
				dst[k] = "false"
			}
		} else {
			dst[k] = v
		}
	}
	return dst
}

// NormalizeBoolsInPlace normalizes bool values in the map in-place.
// This modifies the original map and is useful when you don't need
// to preserve the original values.
func NormalizeBoolsInPlace(ctx map[string]any) {
	for k, v := range ctx {
		if b, ok := v.(bool); ok {
			if b {
				ctx[k] = "true"
			} else {
				ctx[k] = "false"
			}
		}
	}
}
