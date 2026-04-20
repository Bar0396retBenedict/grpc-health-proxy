package health

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiterConfig controls how many health check calls are allowed per interval.
type RateLimiterConfig struct {
	// MaxCalls is the maximum number of calls allowed within the Interval.
	MaxCalls int
	// Interval is the sliding window duration.
	Interval time.Duration
}

// DefaultRateLimiterConfig returns a sensible default: 10 calls per second.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		MaxCalls: 10,
		Interval: time.Second,
	}
}

// ErrRateLimited is returned when a health check is dropped by the rate limiter.
var ErrRateLimited = fmt.Errorf("health check rate limited")

type rateLimiter struct {
	cfg        RateLimiterConfig
	mu         sync.Mutex
	timestamps []time.Time
}

// WithRateLimit wraps a HealthCheckFn and enforces a sliding-window rate limit.
// If the limit is exceeded the call is dropped and StatusUnknown is returned.
func WithRateLimit(cfg RateLimiterConfig, fn HealthCheckFn) HealthCheckFn {
	rl := &rateLimiter{cfg: cfg}
	return func(ctx context.Context, service string) StatusResult {
		if !rl.allow() {
			return StatusResult{Status: StatusUnknown, Err: ErrRateLimited}
		}
		return fn(ctx, service)
	}
}

// allow returns true when the call is permitted under the sliding window.
func (rl *rateLimiter) allow() bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := now.Add(-rl.cfg.Interval)
	valid := rl.timestamps[:0]
	for _, t := range rl.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.timestamps = valid

	if len(rl.timestamps) >= rl.cfg.MaxCalls {
		return false
	}
	rl.timestamps = append(rl.timestamps, now)
	return true
}
