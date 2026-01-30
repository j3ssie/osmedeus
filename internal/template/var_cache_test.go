package template

import (
	"sync"
	"testing"
)

func TestVarRefCache(t *testing.T) {
	cache := NewVarRefCache(10)

	refs := map[string]struct{}{"Target": {}, "Output": {}}
	cache.Set("{{Target}}/{{Output}}", refs)

	got, ok := cache.Get("{{Target}}/{{Output}}")
	if !ok {
		t.Fatal("expected cache hit")
	}

	if len(got) != 2 {
		t.Errorf("expected 2 refs, got %d", len(got))
	}
}

func TestVarRefCacheEviction(t *testing.T) {
	cache := NewVarRefCache(2)

	cache.Set("a", map[string]struct{}{"a": {}})
	cache.Set("b", map[string]struct{}{"b": {}})
	cache.Set("c", map[string]struct{}{"c": {}}) // Should trigger eviction

	// After eviction, cache should be cleared and only have "c"
	if _, ok := cache.Get("c"); !ok {
		t.Error("expected 'c' to be in cache after eviction")
	}
}

func TestVarRefCacheConcurrency(t *testing.T) {
	cache := NewVarRefCache(100)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := string(rune('a' + n%26))
			cache.Set(key, map[string]struct{}{key: {}})
			cache.Get(key)
		}(i)
	}
	wg.Wait()
}

func TestExtractVariablesMethod(t *testing.T) {
	e := NewEngine()

	tests := []struct {
		template string
		expected []string
	}{
		{"{{Target}}", []string{"Target"}},
		{"{{Target}}/{{Output}}/file.txt", []string{"Target", "Output"}},
		{"no variables here", nil},
		{"", nil},
		{"{{ Target }}", []string{"Target"}}, // with spaces
	}

	for _, tt := range tests {
		vars := e.ExtractVariablesSet(tt.template)
		if tt.expected == nil {
			if len(vars) > 0 {
				t.Errorf("ExtractVariablesSet(%q) = %v, want nil", tt.template, vars)
			}
			continue
		}

		if len(vars) != len(tt.expected) {
			t.Errorf("ExtractVariablesSet(%q) got %d vars, want %d", tt.template, len(vars), len(tt.expected))
		}

		for _, exp := range tt.expected {
			if _, ok := vars[exp]; !ok {
				t.Errorf("ExtractVariablesSet(%q) missing variable %q", tt.template, exp)
			}
		}
	}
}

func TestRenderLazy(t *testing.T) {
	e := NewEngine()

	// Large context simulating real workflow
	ctx := make(map[string]any, 100)
	for i := 0; i < 100; i++ {
		ctx[string(rune('A'+i%26))+string(rune('a'+i%26))] = "/some/path/value"
	}
	ctx["Target"] = "example.com"
	ctx["Output"] = "/output/dir"

	template := "{{Target}}/{{Output}}/results.txt"

	result, err := e.RenderLazy(template, ctx)
	if err != nil {
		t.Fatalf("RenderLazy failed: %v", err)
	}

	expected := "example.com//output/dir/results.txt"
	if result != expected {
		t.Errorf("RenderLazy = %q, want %q", result, expected)
	}
}

func TestRenderLazyNoVariables(t *testing.T) {
	e := NewEngine()

	ctx := map[string]any{"Target": "example.com"}
	result, err := e.RenderLazy("no variables", ctx)
	if err != nil {
		t.Fatalf("RenderLazy failed: %v", err)
	}

	if result != "no variables" {
		t.Errorf("RenderLazy = %q, want %q", result, "no variables")
	}
}

func BenchmarkRenderLazyVsRegular(b *testing.B) {
	e := NewEngine()

	// Large context simulating real workflow
	ctx := make(map[string]any, 100)
	for i := 0; i < 100; i++ {
		ctx[string(rune('A'+i%26))+string(rune('a'+i%26))] = "/some/path/value"
	}
	ctx["Target"] = "example.com"
	ctx["Output"] = "/output/dir"

	template := "{{Target}}/{{Output}}/results.txt"

	b.Run("Regular", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = e.Render(template, ctx)
		}
	})

	b.Run("Lazy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = e.RenderLazy(template, ctx)
		}
	})
}
