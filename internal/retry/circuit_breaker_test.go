package retry

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCircuitBreakerClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())

	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("expected state Closed, got %v", cb.State())
	}
}

func TestCircuitBreakerOpensAfterFailures(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold:    3,
		SuccessThreshold:    2,
		RecoveryTimeout:     100 * time.Millisecond,
		MaxHalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	testErr := errors.New("test error")

	// Fail 3 times to open the circuit
	for i := 0; i < 3; i++ {
		_ = cb.Execute(func() error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Errorf("expected state Open after %d failures, got %v", cfg.FailureThreshold, cb.State())
	}

	// Next call should return ErrCircuitOpen
	err := cb.Execute(func() error {
		return nil
	})

	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		RecoveryTimeout:     50 * time.Millisecond,
		MaxHalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Fail to open
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return errors.New("fail")
		})
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected Open state, got %v", cb.State())
	}

	// Wait for recovery timeout
	time.Sleep(60 * time.Millisecond)

	// Should transition to half-open and allow request
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("expected success in half-open, got %v", err)
	}

	// Need another success to close
	_ = cb.Execute(func() error {
		return nil
	})

	if cb.State() != StateClosed {
		t.Errorf("expected Closed after recovery, got %v", cb.State())
	}
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		RecoveryTimeout:     50 * time.Millisecond,
		MaxHalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return errors.New("fail")
		})
	}

	time.Sleep(60 * time.Millisecond)

	// Fail in half-open state
	_ = cb.Execute(func() error {
		return errors.New("still failing")
	})

	// Should go back to Open
	if cb.State() != StateOpen {
		t.Errorf("expected Open after half-open failure, got %v", cb.State())
	}
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold:    10,
		SuccessThreshold:    5,
		RecoveryTimeout:     100 * time.Millisecond,
		MaxHalfOpenRequests: 2,
	}
	cb := NewCircuitBreaker(cfg)

	var successCount int64
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cb.Execute(func() error {
				return nil
			})
			if err == nil {
				atomic.AddInt64(&successCount, 1)
			}
		}()
	}

	wg.Wait()

	if successCount != 100 {
		t.Errorf("expected 100 successes, got %d", successCount)
	}

	if cb.State() != StateClosed {
		t.Errorf("expected Closed, got %v", cb.State())
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cfg := CircuitBreakerConfig{
		FailureThreshold:    2,
		SuccessThreshold:    2,
		RecoveryTimeout:     time.Hour, // Long timeout
		MaxHalfOpenRequests: 1,
	}
	cb := NewCircuitBreaker(cfg)

	// Open the circuit
	for i := 0; i < 2; i++ {
		_ = cb.Execute(func() error {
			return errors.New("fail")
		})
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected Open, got %v", cb.State())
	}

	// Reset should close it immediately
	cb.Reset()

	if cb.State() != StateClosed {
		t.Errorf("expected Closed after reset, got %v", cb.State())
	}

	// Should work again
	err := cb.Execute(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("expected success after reset, got %v", err)
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.expected)
		}
	}
}
