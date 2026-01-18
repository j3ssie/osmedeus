package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	callCount := 0
	err := Do(context.Background(), DefaultConfig(), func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestDo_SuccessOnSecondAttempt(t *testing.T) {
	callCount := 0
	err := Do(context.Background(), Config{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}, func() error {
		callCount++
		if callCount == 1 {
			return Retryable(errors.New("transient error"))
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestDo_MaxRetriesExceeded(t *testing.T) {
	callCount := 0
	err := Do(context.Background(), Config{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}, func() error {
		callCount++
		return Retryable(errors.New("always fails"))
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
	if !errors.Is(err, errors.New("always fails")) {
		// Check that the error message contains our original error
		if err.Error() != "max retries (3) exceeded: always fails" {
			t.Errorf("unexpected error message: %v", err)
		}
	}
}

func TestDo_NonRetryableError(t *testing.T) {
	callCount := 0
	expectedErr := errors.New("non-retryable error")
	err := Do(context.Background(), DefaultConfig(), func() error {
		callCount++
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retry for non-retryable), got %d", callCount)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, Config{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}, func() error {
		callCount++
		return Retryable(errors.New("keep trying"))
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestDo_ContextAlreadyCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	callCount := 0
	err := Do(ctx, DefaultConfig(), func() error {
		callCount++
		return nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls (context already cancelled), got %d", callCount)
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"regular error", errors.New("regular"), false},
		{"retryable error", Retryable(errors.New("retryable")), true},
		{"wrapped retryable", errors.New("outer: " + Retryable(errors.New("inner")).Error()), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDoWithResult(t *testing.T) {
	callCount := 0
	result, err := DoWithResult(context.Background(), Config{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}, func() (string, error) {
		callCount++
		if callCount == 1 {
			return "", Retryable(errors.New("transient"))
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %q", result)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts=3, got %d", cfg.MaxAttempts)
	}
	if cfg.InitialDelay != 100*time.Millisecond {
		t.Errorf("expected InitialDelay=100ms, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 10*time.Second {
		t.Errorf("expected MaxDelay=10s, got %v", cfg.MaxDelay)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected Multiplier=2.0, got %v", cfg.Multiplier)
	}
}

func TestRetryable_NilError(t *testing.T) {
	err := Retryable(nil)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
