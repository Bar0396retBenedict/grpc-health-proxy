// Package health provides health checking primitives.
package health

import (
	"context"
	"math/rand"
	"time"
)

// DefaultJitterConfig returns a JitterConfig with sensible defaults.
func DefaultJitterConfig() JitterConfig {
	return JitterConfig{
		MaxJitter: 500 * time.Millisecond,
	}
}

// JitterConfig controls how much random delay is added before invoking a check.
type JitterConfig struct {
	// MaxJitter is the upper bound of the random delay added before each check.
	// A value of zero disables jitter entirely.
	MaxJitter time.Duration
}

// WithJitter wraps a HealthCheckFn and adds a random delay in [0, cfg.MaxJitter)
// before each invocation. This spreads concurrent checks across time to avoid
// thundering-herd problems against a shared upstream.
func WithJitter(cfg JitterConfig, fn HealthCheckFn) HealthCheckFn {
	if cfg.MaxJitter <= 0 {
		return fn
	}
	return func(ctx context.Context, service string) StatusResult {
		// #nosec G404 — non-cryptographic randomness is fine for jitter.
		delay := time.Duration(rand.Int63n(int64(cfg.MaxJitter)))
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
		}
		return fn(ctx, service)
	}
}
