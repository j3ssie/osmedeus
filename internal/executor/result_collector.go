package executor

import (
	"sync"
	"sync/atomic"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

// ResultCollector provides lock-free collection of step results for parallel execution.
// Uses pre-allocated slice with atomic index for O(1) appends without mutex contention.
type ResultCollector struct {
	results []*core.StepResult
	index   int64
	size    int
}

// NewResultCollector creates a collector pre-allocated for the expected number of steps
func NewResultCollector(stepCount int) *ResultCollector {
	return &ResultCollector{
		results: make([]*core.StepResult, stepCount),
		index:   0,
		size:    stepCount,
	}
}

// Add atomically adds a step result to the collector.
// Thread-safe without mutex - uses atomic operations.
// Returns the index where the result was stored.
func (c *ResultCollector) Add(result *core.StepResult) int {
	// Atomically get next index
	idx := int(atomic.AddInt64(&c.index, 1) - 1)

	// Bounds check (should never happen if stepCount is correct)
	if idx < c.size {
		c.results[idx] = result
	}

	return idx
}

// Results returns all collected results, filtering nil entries.
// Call this only after all goroutines have completed.
func (c *ResultCollector) Results() []*core.StepResult {
	count := int(atomic.LoadInt64(&c.index))
	if count > c.size {
		count = c.size
	}

	// Filter out nil entries (from pre-allocation)
	results := make([]*core.StepResult, 0, count)
	for i := 0; i < count; i++ {
		if c.results[i] != nil {
			results = append(results, c.results[i])
		}
	}
	return results
}

// Count returns the number of results added
func (c *ResultCollector) Count() int {
	return int(atomic.LoadInt64(&c.index))
}

// OrderedResultCollector preserves insertion order using step name mapping.
// Useful when result order must match step execution order.
type OrderedResultCollector struct {
	results  map[string]*core.StepResult
	mu       sync.RWMutex
	order    []string
	capacity int
}

// NewOrderedResultCollector creates a collector that preserves step ordering
func NewOrderedResultCollector(stepNames []string) *OrderedResultCollector {
	return &OrderedResultCollector{
		results:  make(map[string]*core.StepResult, len(stepNames)),
		order:    stepNames,
		capacity: len(stepNames),
	}
}

// Add adds a result for a specific step
func (c *OrderedResultCollector) Add(stepName string, result *core.StepResult) {
	c.mu.Lock()
	c.results[stepName] = result
	c.mu.Unlock()
}

// Results returns results in the original step order
func (c *OrderedResultCollector) Results() []*core.StepResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make([]*core.StepResult, 0, len(c.order))
	for _, name := range c.order {
		if r, ok := c.results[name]; ok && r != nil {
			results = append(results, r)
		}
	}
	return results
}
