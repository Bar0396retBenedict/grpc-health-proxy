package health

import (
	"context"
	"time"
)

// RetryConfig holds configuration for retry behaviour on health check failures.
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       200 * time.Millisecond,
	}
}

// WithRetry executes fn up to cfg.MaxAttempts times, returning the first nil
// error or the last non-nil error. The context is checked before each attempt
// so callers can cancel early.
func WithRetry(ctx context.Context, cfg RetryConfig, fn func(ctx context.Context) error) error {
	var err error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn(ctx)
		if err == nil {
			return nil
		}
		if attempt < cfg.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cfg.Delay):
			}
		}
	}
	return err
}
