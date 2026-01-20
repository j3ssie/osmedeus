package template

import (
	"fmt"
	"sync"
	"testing"
)

func TestShardedEngine_Render(t *testing.T) {
	engine := NewShardedEngine()

	tests := []struct {
		name     string
		template string
		ctx      map[string]any
		want     string
		wantErr  bool
	}{
		{
			name:     "simple variable",
			template: "Hello {{name}}!",
			ctx:      map[string]any{"name": "World"},
			want:     "Hello World!",
		},
		{
			name:     "multiple variables",
			template: "{{greeting}} {{name}}!",
			ctx:      map[string]any{"greeting": "Hello", "name": "World"},
			want:     "Hello World!",
		},
		{
			name:     "no variables",
			template: "Hello World!",
			ctx:      map[string]any{},
			want:     "Hello World!",
		},
		{
			name:     "bool true",
			template: "Value: {{value}}",
			ctx:      map[string]any{"value": true},
			want:     "Value: true",
		},
		{
			name:     "bool false",
			template: "Value: {{value}}",
			ctx:      map[string]any{"value": false},
			want:     "Value: false",
		},
		{
			name:     "undefined variable",
			template: "Value: {{undefined}}",
			ctx:      map[string]any{},
			want:     "Value: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Render(tt.template, tt.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShardedEngine_RenderConcurrent(t *testing.T) {
	engine := NewShardedEngine()

	const numGoroutines = 100
	const numIterations = 100

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numIterations)

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range numIterations {
				template := fmt.Sprintf("Hello {{name}}! Iteration %d-%d", id, j)
				ctx := map[string]any{"name": fmt.Sprintf("User%d", id)}
				expected := fmt.Sprintf("Hello User%d! Iteration %d-%d", id, id, j)

				result, err := engine.Render(template, ctx)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d iteration %d: %w", id, j, err)
					return
				}
				if result != expected {
					errors <- fmt.Errorf("goroutine %d iteration %d: got %q, want %q", id, j, result, expected)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

func TestShardedEngine_RenderBatch(t *testing.T) {
	engine := NewShardedEngine()

	requests := []RenderRequest{
		{Key: "greeting", Template: "Hello {{name}}!"},
		{Key: "farewell", Template: "Goodbye {{name}}!"},
		{Key: "static", Template: "No variables here"},
		{Key: "number", Template: "Count: {{count}}"},
	}

	ctx := map[string]any{
		"name":  "World",
		"count": 42,
	}

	results, err := engine.RenderBatch(requests, ctx)
	if err != nil {
		t.Fatalf("RenderBatch() error = %v", err)
	}

	expected := map[string]string{
		"greeting": "Hello World!",
		"farewell": "Goodbye World!",
		"static":   "No variables here",
		"number":   "Count: 42",
	}

	for key, want := range expected {
		if got := results[key]; got != want {
			t.Errorf("RenderBatch()[%s] = %q, want %q", key, got, want)
		}
	}
}

func TestShardedEngine_CacheStats(t *testing.T) {
	engine := NewShardedEngine()

	// Render some templates to populate cache
	templates := []string{
		"Template {{a}}",
		"Template {{b}}",
		"Template {{c}}",
	}

	ctx := map[string]any{"a": "1", "b": "2", "c": "3"}
	for _, tmpl := range templates {
		_, _ = engine.Render(tmpl, ctx)
	}

	stats := engine.CacheStats()
	if stats["total_cached"] != 3 {
		t.Errorf("CacheStats() total_cached = %d, want 3", stats["total_cached"])
	}
	if stats["shard_count"] != DefaultShardCount {
		t.Errorf("CacheStats() shard_count = %d, want %d", stats["shard_count"], DefaultShardCount)
	}
}

func TestShardedEngine_RenderSecondary(t *testing.T) {
	engine := NewShardedEngine()

	tests := []struct {
		name     string
		template string
		ctx      map[string]any
		want     string
	}{
		{
			name:     "secondary delimiters",
			template: "Value: [[value]]",
			ctx:      map[string]any{"value": "test"},
			want:     "Value: test",
		},
		{
			name:     "mixed delimiters",
			template: "Primary: {{primary}}, Secondary: [[secondary]]",
			ctx:      map[string]any{"primary": "a", "secondary": "b"},
			want:     "Primary: a, Secondary: b",
		},
		{
			name:     "no secondary",
			template: "No secondary here",
			ctx:      map[string]any{},
			want:     "No secondary here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.RenderSecondary(tt.template, tt.ctx)
			if err != nil {
				t.Errorf("RenderSecondary() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("RenderSecondary() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShardedEngine_ExecuteGenerator(t *testing.T) {
	engine := NewShardedEngine()

	// Test UUID generator (just check it doesn't error and returns something)
	result, err := engine.ExecuteGenerator("uuid()")
	if err != nil {
		t.Errorf("ExecuteGenerator(uuid()) error = %v", err)
	}
	if len(result) != 36 { // UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
		t.Errorf("ExecuteGenerator(uuid()) returned invalid UUID: %s", result)
	}

	// Test concat generator
	result, err = engine.ExecuteGenerator("concat(Hello, World)")
	if err != nil {
		t.Errorf("ExecuteGenerator(concat()) error = %v", err)
	}
	if result != "HelloWorld" {
		t.Errorf("ExecuteGenerator(concat()) = %q, want %q", result, "HelloWorld")
	}
}

func TestShardedEngineConfig(t *testing.T) {
	cfg := ShardedEngineConfig{
		ShardCount:     8,
		ShardCacheSize: 32,
		EnablePooling:  false,
	}

	engine := NewShardedEngineWithConfig(cfg)

	// Verify shard count (should be 8 as it's already power of 2)
	stats := engine.CacheStats()
	if stats["shard_count"] != 8 {
		t.Errorf("Custom config: shard_count = %d, want 8", stats["shard_count"])
	}
}

func TestNextPowerOf2(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{15, 16},
		{16, 16},
		{17, 32},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("input=%d", tt.input), func(t *testing.T) {
			got := nextPowerOf2(tt.input)
			if got != tt.want {
				t.Errorf("nextPowerOf2(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
