// Package health provides health checking primitives.
package health

import (
	"context"
	"sync"
	"time"
)

// ThrottleConfig controls how often a check function may be called.
type ThrottleConfig struct {
	// MinInterval is the minimum duration between successive calls.
	MinInterval time.Duration
}

// DefaultThrottleConfig returns a ThrottleConfig with sensible defaults.
func DefaultThrottleConfig() ThrottleConfig {
	return ThrottleConfig{
		MinInterval: 500 * time.Millisecond,
	}
}

// throttle holds state for a single throttled function.
type throttle struct {
	cfg    ThrottleConfig
	mu     sync.Mutex
	lastAt time.Time
}

// WithThrottle wraps a CheckFn so that it is called at most once per
// ThrottleConfig.MinInterval. If called before the interval has elapsed the
// previous result (cached by the caller) is returned via ErrThrottled.
var ErrThrottled = throttleError("check throttled: called too frequently")

type throttleError string

func (e throttleError) Error() string { return string(e) }

// WithThrottle returns a new CheckFn that enforces a minimum call interval.
func WithThrottle(cfg ThrottleConfig, fn CheckFn) CheckFn {
	t := &throttle{cfg: cfg}
	return func(ctx context.Context, service string) (bool, error) {
		t.mu.Lock()
		now := time.Now()
		if !t.lastAt.IsZero() && now.Sub(t.lastAt) < t.cfg.MinInterval {
			t.mu.Unlock()
			return false, ErrThrottled
		}
		t.lastAt = now
		t.mu.Unlock()
		return fn(ctx, service)
	}
}
