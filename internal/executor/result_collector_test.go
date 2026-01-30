package executor

import (
	"sync"
	"testing"
	"time"

	"github.com/j3ssie/osmedeus/v5/internal/core"
)

func TestResultCollectorConcurrent(t *testing.T) {
	stepCount := 100
	collector := NewResultCollector(stepCount)

	var wg sync.WaitGroup
	for i := 0; i < stepCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			result := &core.StepResult{
				StepName: "step-" + string(rune('a'+idx%26)),
				Status:   core.StepStatusSuccess,
			}
			collector.Add(result)
		}(i)
	}

	wg.Wait()

	results := collector.Results()
	if len(results) != stepCount {
		t.Errorf("expected %d results, got %d", stepCount, len(results))
	}

	if collector.Count() != stepCount {
		t.Errorf("expected count %d, got %d", stepCount, collector.Count())
	}
}

func TestResultCollectorOrder(t *testing.T) {
	names := []string{"step-a", "step-b", "step-c"}
	collector := NewOrderedResultCollector(names)

	// Add in reverse order
	collector.Add("step-c", &core.StepResult{StepName: "step-c"})
	collector.Add("step-a", &core.StepResult{StepName: "step-a"})
	collector.Add("step-b", &core.StepResult{StepName: "step-b"})

	results := collector.Results()

	// Should be in original order
	for i, r := range results {
		if r.StepName != names[i] {
			t.Errorf("position %d: expected %s, got %s", i, names[i], r.StepName)
		}
	}
}

func BenchmarkResultCollectorLockFree(b *testing.B) {
	b.Run("lock-free", func(b *testing.B) {
		collector := NewResultCollector(b.N)
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				collector.Add(&core.StepResult{
					StepName:  "test",
					Status:    core.StepStatusSuccess,
					StartTime: time.Now(),
				})
			}()
		}
		wg.Wait()
	})

	b.Run("mutex", func(b *testing.B) {
		var mu sync.Mutex
		results := make([]*core.StepResult, 0, b.N)
		var wg sync.WaitGroup
		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				mu.Lock()
				results = append(results, &core.StepResult{
					StepName:  "test",
					Status:    core.StepStatusSuccess,
					StartTime: time.Now(),
				})
				mu.Unlock()
			}()
		}
		wg.Wait()
	})
}
