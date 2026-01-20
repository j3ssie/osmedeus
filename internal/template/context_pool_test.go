package template

import (
	"testing"
)

func TestContextPool_GetAndPut(t *testing.T) {
	// Get a context from the pool
	ctx := GetContext()
	if ctx == nil {
		t.Fatal("GetContext() returned nil")
	}

	// Verify it's empty
	if len(ctx) != 0 {
		t.Errorf("GetContext() returned non-empty map with %d entries", len(ctx))
	}

	// Add some values
	ctx["key1"] = "value1"
	ctx["key2"] = 42

	// Return to pool
	PutContext(ctx)

	// Get another context - it should be empty (cleared on put)
	ctx2 := GetContext()
	if len(ctx2) != 0 {
		t.Errorf("Second GetContext() returned non-empty map with %d entries", len(ctx2))
	}

	PutContext(ctx2)
}

func TestCloneToPooled(t *testing.T) {
	src := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	dst := CloneToPooled(src)
	if dst == nil {
		t.Fatal("CloneToPooled() returned nil")
	}

	// Verify all keys are copied
	if len(dst) != len(src) {
		t.Errorf("CloneToPooled() length = %d, want %d", len(dst), len(src))
	}

	for k, v := range src {
		if dst[k] != v {
			t.Errorf("CloneToPooled()[%s] = %v, want %v", k, dst[k], v)
		}
	}

	// Verify it's a different map (not same reference)
	dst["new_key"] = "new_value"
	if _, ok := src["new_key"]; ok {
		t.Error("CloneToPooled() returned same map reference")
	}

	PutContext(dst)
}

func TestCloneToPooled_Nil(t *testing.T) {
	dst := CloneToPooled(nil)
	if dst != nil {
		t.Errorf("CloneToPooled(nil) = %v, want nil", dst)
	}
}

func TestNormalizeBoolsToPooled(t *testing.T) {
	src := map[string]any{
		"bool_true":  true,
		"bool_false": false,
		"string":     "hello",
		"number":     42,
	}

	dst := NormalizeBoolsToPooled(src)
	if dst == nil {
		t.Fatal("NormalizeBoolsToPooled() returned nil")
	}

	// Verify bool normalization
	if dst["bool_true"] != "true" {
		t.Errorf("NormalizeBoolsToPooled()[bool_true] = %v, want \"true\"", dst["bool_true"])
	}
	if dst["bool_false"] != "false" {
		t.Errorf("NormalizeBoolsToPooled()[bool_false] = %v, want \"false\"", dst["bool_false"])
	}

	// Verify non-bool values are unchanged
	if dst["string"] != "hello" {
		t.Errorf("NormalizeBoolsToPooled()[string] = %v, want \"hello\"", dst["string"])
	}
	if dst["number"] != 42 {
		t.Errorf("NormalizeBoolsToPooled()[number] = %v, want 42", dst["number"])
	}

	PutContext(dst)
}

func TestNormalizeBoolsToPooled_Nil(t *testing.T) {
	dst := NormalizeBoolsToPooled(nil)
	if dst != nil {
		t.Errorf("NormalizeBoolsToPooled(nil) = %v, want nil", dst)
	}
}

func TestNormalizeBoolsInPlace(t *testing.T) {
	ctx := map[string]any{
		"bool_true":  true,
		"bool_false": false,
		"string":     "hello",
	}

	NormalizeBoolsInPlace(ctx)

	if ctx["bool_true"] != "true" {
		t.Errorf("NormalizeBoolsInPlace()[bool_true] = %v, want \"true\"", ctx["bool_true"])
	}
	if ctx["bool_false"] != "false" {
		t.Errorf("NormalizeBoolsInPlace()[bool_false] = %v, want \"false\"", ctx["bool_false"])
	}
	if ctx["string"] != "hello" {
		t.Errorf("NormalizeBoolsInPlace()[string] = %v, want \"hello\"", ctx["string"])
	}
}

func TestPutContext_Nil(t *testing.T) {
	// Should not panic
	PutContext(nil)
}
