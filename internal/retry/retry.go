package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

// Config configures retry behavior
type Config struct {
	MaxAttempts  int           // Maximum number of attempts (default: 3)
	InitialDelay time.Duration // Initial delay between retries (default: 100ms)
	MaxDelay     time.Duration // Maximum delay between retries (default: 10s)
	Multiplier   float64       // Delay multiplier for exponential backoff (default: 2.0)
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}
}

// RetryableError indicates the operation can be retried
type RetryableError struct {
	Err error
}

func (e RetryableError) Error() string { return e.Err.Error() }
func (e RetryableError) Unwrap() error { return e.Err }

// Retryable wraps an error to indicate it should be retried
func Retryable(err error) error {
	if err == nil {
		return nil
	}
	return RetryableError{Err: err}
}

// IsRetryable checks if error should trigger a retry
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	var re RetryableError
	return errors.As(err, &re)
}

// Do executes fn with retries using exponential backoff
func Do(ctx context.Context, cfg Config, fn func() error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.InitialDelay <= 0 {
		cfg.InitialDelay = 100 * time.Millisecond
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 10 * time.Second
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 2.0
	}

	var lastErr error

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryable(err) {
			// Non-retryable error, return immediately
			return err
		}

		// Don't sleep after last attempt
		if attempt < cfg.MaxAttempts-1 {
			delay := time.Duration(float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt)))
			delay = min(delay, cfg.MaxDelay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	// Unwrap the RetryableError for the final error message
	var re RetryableError
	if errors.As(lastErr, &re) {
		lastErr = re.Err
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxAttempts, lastErr)
}

// DoWithResult executes fn with retries and returns a result
func DoWithResult[T any](ctx context.Context, cfg Config, fn func() (T, error)) (T, error) {
	var result T
	err := Do(ctx, cfg, func() error {
		var fnErr error
		result, fnErr = fn()
		return fnErr
	})
	return result, err
}
