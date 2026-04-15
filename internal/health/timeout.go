package health

import (
	"context"
	"fmt"
	"time"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	// Timeout is the maximum duration allowed for a single health check call.
	Timeout time.Duration
}

// DefaultTimeoutConfig returns a TimeoutConfig with sensible defaults.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout: 5 * time.Second,
	}
}

// CheckFunc is a function that performs a health check for a given service.
type CheckFunc func(ctx context.Context, service string) (bool, error)

// WithTimeout wraps a CheckFunc so that each invocation is bounded by a
// deadline derived from cfg.Timeout. If the deadline is exceeded the call
// returns false and a descriptive error; the underlying context is always
// cancelled to release resources.
func WithTimeout(cfg TimeoutConfig, next CheckFunc) CheckFunc {
	return func(ctx context.Context, service string) (bool, error) {
		ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()

		type result struct {
			ok  bool
			err error
		}

		ch := make(chan result, 1)
		go func() {
			ok, err := next(ctx, service)
			ch <- result{ok, err}
		}()

		select {
		case res := <-ch:
			return res.ok, res.err
		case <-ctx.Done():
			return false, fmt.Errorf("health check timed out after %s: %w", cfg.Timeout, ctx.Err())
		}
	}
}
