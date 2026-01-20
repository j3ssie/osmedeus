package template

import (
	"fmt"
	"sync"
	"testing"
)

// Benchmark contexts
var (
	mediumCtx = map[string]any{
		"target":     "example.com",
		"output":     "/tmp/output",
		"threads":    10,
		"timeout":    "30s",
		"user":       "admin",
		"verbose":    true,
		"dry_run":    false,
		"workspace":  "/workspaces/example.com",
		"binaries":   "/opt/tools",
		"wordlist":   "/data/wordlists/common.txt",
	}

	largeCtx = func() map[string]any {
		ctx := make(map[string]any, 64)
		for i := range 64 {
			ctx[fmt.Sprintf("var%d", i)] = fmt.Sprintf("value%d", i)
		}
		return ctx
	}()
)

// BenchmarkEngine_Render benchmarks the standard engine single-threaded
func BenchmarkEngine_Render(b *testing.B) {
	engine := NewEngine()
	template := "Hello {{name}}! Target: {{target}}, Output: {{output}}"
	ctx := mediumCtx

	b.ResetTimer()
	for range b.N {
		_, _ = engine.Render(template, ctx)
	}
}

// BenchmarkShardedEngine_Render benchmarks the sharded engine single-threaded
func BenchmarkShardedEngine_Render(b *testing.B) {
	engine := NewShardedEngine()
	template := "Hello {{name}}! Target: {{target}}, Output: {{output}}"
	ctx := mediumCtx

	b.ResetTimer()
	for range b.N {
		_, _ = engine.Render(template, ctx)
	}
}

// BenchmarkEngine_RenderParallel benchmarks standard engine under concurrency
func BenchmarkEngine_RenderParallel(b *testing.B) {
	engine := NewEngine()
	template := "Hello {{name}}! Target: {{target}}, Output: {{output}}"
	ctx := mediumCtx

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.Render(template, ctx)
		}
	})
}

// BenchmarkShardedEngine_RenderParallel benchmarks sharded engine under concurrency
func BenchmarkShardedEngine_RenderParallel(b *testing.B) {
	engine := NewShardedEngine()
	template := "Hello {{name}}! Target: {{target}}, Output: {{output}}"
	ctx := mediumCtx

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = engine.Render(template, ctx)
		}
	})
}

// BenchmarkRenderBatch benchmarks batch rendering
func BenchmarkRenderBatch(b *testing.B) {
	engine := NewShardedEngine()
	requests := []RenderRequest{
		{Key: "command", Template: "nmap -sV {{target}} -o {{output}}/nmap.txt"},
		{Key: "input", Template: "{{output}}/subdomains.txt"},
		{Key: "output", Template: "{{workspace}}/results/{{target}}"},
		{Key: "log", Template: "Running scan on {{target}}..."},
		{Key: "timeout", Template: "{{timeout}}"},
		{Key: "threads", Template: "{{threads}}"},
	}
	ctx := mediumCtx

	b.ResetTimer()
	for range b.N {
		_, _ = engine.RenderBatch(requests, ctx)
	}
}

// BenchmarkRenderIndividual benchmarks individual rendering (for comparison)
func BenchmarkRenderIndividual(b *testing.B) {
	engine := NewShardedEngine()
	templates := []string{
		"nmap -sV {{target}} -o {{output}}/nmap.txt",
		"{{output}}/subdomains.txt",
		"{{workspace}}/results/{{target}}",
		"Running scan on {{target}}...",
		"{{timeout}}",
		"{{threads}}",
	}
	ctx := mediumCtx

	b.ResetTimer()
	for range b.N {
		for _, tmpl := range templates {
			_, _ = engine.Render(tmpl, ctx)
		}
	}
}

// BenchmarkContextPooling benchmarks context map pooling
func BenchmarkContextPooling_WithPool(b *testing.B) {
	src := mediumCtx

	b.ResetTimer()
	for range b.N {
		ctx := CloneToPooled(src)
		NormalizeBoolsInPlace(ctx)
		PutContext(ctx)
	}
}

// BenchmarkContextPooling_NoPool benchmarks without pooling (allocation each time)
func BenchmarkContextPooling_NoPool(b *testing.B) {
	src := mediumCtx

	b.ResetTimer()
	for range b.N {
		_ = normalizeBoolsForTemplate(src)
	}
}

// BenchmarkEngine_HighConcurrency simulates high concurrency workload
func BenchmarkEngine_HighConcurrency(b *testing.B) {
	engine := NewEngine()
	templates := []string{
		"Command: {{command}}",
		"Target: {{target}}",
		"Output: {{output}}",
		"Timeout: {{timeout}}",
	}

	b.ResetTimer()
	b.SetParallelism(16) // 16 goroutines

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tmpl := templates[i%len(templates)]
			_, _ = engine.Render(tmpl, mediumCtx)
			i++
		}
	})
}

// BenchmarkShardedEngine_HighConcurrency simulates high concurrency workload
func BenchmarkShardedEngine_HighConcurrency(b *testing.B) {
	engine := NewShardedEngine()
	templates := []string{
		"Command: {{command}}",
		"Target: {{target}}",
		"Output: {{output}}",
		"Timeout: {{timeout}}",
	}

	b.ResetTimer()
	b.SetParallelism(16) // 16 goroutines

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tmpl := templates[i%len(templates)]
			_, _ = engine.Render(tmpl, mediumCtx)
			i++
		}
	})
}

// BenchmarkShardedEngine_VaryingTemplates benchmarks with many different templates
func BenchmarkShardedEngine_VaryingTemplates(b *testing.B) {
	engine := NewShardedEngine()

	// Generate many unique templates
	templates := make([]string, 100)
	for i := range templates {
		templates[i] = fmt.Sprintf("Template %d: {{var%d}}", i, i%64)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tmpl := templates[i%len(templates)]
			_, _ = engine.Render(tmpl, largeCtx)
			i++
		}
	})
}

// BenchmarkCacheHitRate benchmarks cache performance
func BenchmarkCacheHitRate_ShardedEngine(b *testing.B) {
	engine := NewShardedEngine()

	// Pre-warm cache
	warmupTemplates := []string{
		"Cached template 1: {{target}}",
		"Cached template 2: {{output}}",
		"Cached template 3: {{threads}}",
	}
	for _, tmpl := range warmupTemplates {
		_, _ = engine.Render(tmpl, mediumCtx)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tmpl := warmupTemplates[i%len(warmupTemplates)]
			_, _ = engine.Render(tmpl, mediumCtx)
			i++
		}
	})
}

// Test concurrent writes don't corrupt cache
func TestShardedEngine_ConcurrentCacheWrites(t *testing.T) {
	engine := NewShardedEngine()
	ctx := mediumCtx

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Each goroutine renders unique templates
			for j := range 100 {
				tmpl := fmt.Sprintf("Unique template %d-%d: {{target}}", id, j)
				_, err := engine.Render(tmpl, ctx)
				if err != nil {
					t.Errorf("goroutine %d: %v", id, err)
				}
			}
		}(i)
	}
	wg.Wait()
}
