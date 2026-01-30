package retry

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is in Open state
var ErrCircuitOpen = errors.New("circuit breaker is open")

// ErrTooManyRequests is returned when the circuit breaker is in HalfOpen state
// and max concurrent requests limit is reached
var ErrTooManyRequests = errors.New("too many requests in half-open state")

// State represents the state of the circuit breaker
type State int32

const (
	StateClosed   State = iota // Normal operation, requests allowed
	StateOpen                  // Circuit tripped, requests blocked
	StateHalfOpen              // Testing if service recovered
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig configures the circuit breaker behavior
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit
	FailureThreshold int64
	// SuccessThreshold is the number of consecutive successes in half-open state to close the circuit
	SuccessThreshold int64
	// RecoveryTimeout is how long to wait before transitioning from Open to HalfOpen
	RecoveryTimeout time.Duration
	// MaxHalfOpenRequests is the max concurrent requests allowed in half-open state
	MaxHalfOpenRequests int64
	// OnStateChange is called when the circuit breaker state changes
	OnStateChange func(from, to State)
}

// DefaultCircuitBreakerConfig returns sensible defaults for a circuit breaker
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:    5,
		SuccessThreshold:    2,
		RecoveryTimeout:     30 * time.Second,
		MaxHalfOpenRequests: 1,
	}
}

// CircuitBreaker implements the circuit breaker pattern to prevent
// resource exhaustion when downstream services are failing
type CircuitBreaker struct {
	cfg CircuitBreakerConfig

	state            int32 // atomic State
	failureCount     int64 // atomic
	successCount     int64 // atomic
	halfOpenRequests int64 // atomic - current requests in half-open state
	lastFailureTime  int64 // atomic - unix nano timestamp

	mu sync.Mutex // protects state transitions
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 2
	}
	if cfg.RecoveryTimeout <= 0 {
		cfg.RecoveryTimeout = 30 * time.Second
	}
	if cfg.MaxHalfOpenRequests <= 0 {
		cfg.MaxHalfOpenRequests = 1
	}

	return &CircuitBreaker{
		cfg:   cfg,
		state: int32(StateClosed),
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() State {
	return State(atomic.LoadInt32(&cb.state))
}

// Execute runs the given function through the circuit breaker.
// Returns ErrCircuitOpen if the circuit is open and recovery timeout hasn't passed.
// Returns ErrTooManyRequests if in half-open state with too many concurrent requests.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// Check if we can proceed
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function
	err := fn()

	// Record the result
	cb.afterRequest(err)

	return err
}

// ExecuteWithContext runs the given function with context through the circuit breaker
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, fn func() error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return cb.Execute(fn)
}

// beforeRequest checks if the request should be allowed
func (cb *CircuitBreaker) beforeRequest() error {
	state := cb.State()

	switch state {
	case StateClosed:
		return nil

	case StateOpen:
		// Check if recovery timeout has passed
		lastFailure := time.Unix(0, atomic.LoadInt64(&cb.lastFailureTime))
		if time.Since(lastFailure) < cb.cfg.RecoveryTimeout {
			return ErrCircuitOpen
		}

		// Try to transition to half-open
		cb.mu.Lock()
		if cb.State() == StateOpen {
			cb.setState(StateHalfOpen)
			atomic.StoreInt64(&cb.halfOpenRequests, 0)
			atomic.StoreInt64(&cb.successCount, 0)
		}
		cb.mu.Unlock()

		// Now in half-open, fall through to check request limit
		return cb.checkHalfOpenLimit()

	case StateHalfOpen:
		return cb.checkHalfOpenLimit()

	default:
		return nil
	}
}

// checkHalfOpenLimit checks if we can accept another request in half-open state
func (cb *CircuitBreaker) checkHalfOpenLimit() error {
	current := atomic.AddInt64(&cb.halfOpenRequests, 1)
	if current > cb.cfg.MaxHalfOpenRequests {
		atomic.AddInt64(&cb.halfOpenRequests, -1)
		return ErrTooManyRequests
	}
	return nil
}

// afterRequest records the result of the request
func (cb *CircuitBreaker) afterRequest(err error) {
	state := cb.State()

	if err != nil {
		cb.recordFailure(state)
	} else {
		cb.recordSuccess(state)
	}
}

// recordFailure handles a failed request
func (cb *CircuitBreaker) recordFailure(state State) {
	atomic.StoreInt64(&cb.lastFailureTime, time.Now().UnixNano())

	switch state {
	case StateClosed:
		failures := atomic.AddInt64(&cb.failureCount, 1)
		if failures >= cb.cfg.FailureThreshold {
			cb.mu.Lock()
			if cb.State() == StateClosed {
				cb.setState(StateOpen)
			}
			cb.mu.Unlock()
		}

	case StateHalfOpen:
		atomic.AddInt64(&cb.halfOpenRequests, -1)
		cb.mu.Lock()
		if cb.State() == StateHalfOpen {
			cb.setState(StateOpen)
		}
		cb.mu.Unlock()
	}
}

// recordSuccess handles a successful request
func (cb *CircuitBreaker) recordSuccess(state State) {
	switch state {
	case StateClosed:
		atomic.StoreInt64(&cb.failureCount, 0)

	case StateHalfOpen:
		atomic.AddInt64(&cb.halfOpenRequests, -1)
		successes := atomic.AddInt64(&cb.successCount, 1)
		if successes >= cb.cfg.SuccessThreshold {
			cb.mu.Lock()
			if cb.State() == StateHalfOpen {
				cb.setState(StateClosed)
				atomic.StoreInt64(&cb.failureCount, 0)
			}
			cb.mu.Unlock()
		}
	}
}

// setState changes the circuit breaker state and calls the callback if set
func (cb *CircuitBreaker) setState(newState State) {
	oldState := State(atomic.SwapInt32(&cb.state, int32(newState)))
	if cb.cfg.OnStateChange != nil && oldState != newState {
		cb.cfg.OnStateChange(oldState, newState)
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.setState(StateClosed)
	atomic.StoreInt64(&cb.failureCount, 0)
	atomic.StoreInt64(&cb.successCount, 0)
	atomic.StoreInt64(&cb.halfOpenRequests, 0)
}

// Counts returns the current failure and success counts
func (cb *CircuitBreaker) Counts() (failures, successes int64) {
	return atomic.LoadInt64(&cb.failureCount), atomic.LoadInt64(&cb.successCount)
}

// DoWithCircuitBreaker executes fn with both retry and circuit breaker protection.
// The circuit breaker wraps the retry logic.
func DoWithCircuitBreaker(ctx context.Context, cb *CircuitBreaker, retryCfg Config, fn func() error) error {
	return cb.ExecuteWithContext(ctx, func() error {
		return Do(ctx, retryCfg, fn)
	})
}

// DoWithResultAndCircuitBreaker executes fn with both retry and circuit breaker protection, returning a result.
func DoWithResultAndCircuitBreaker[T any](ctx context.Context, cb *CircuitBreaker, retryCfg Config, fn func() (T, error)) (T, error) {
	var result T
	err := cb.ExecuteWithContext(ctx, func() error {
		var fnErr error
		result, fnErr = DoWithResult(ctx, retryCfg, fn)
		return fnErr
	})
	return result, err
}
