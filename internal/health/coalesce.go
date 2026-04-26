package health

import (
	"context"
	"sync"
	"time"
)

// DefaultCoalesceConfig returns a CoalesceConfig with sensible defaults.
func DefaultCoalesceConfig() CoalesceConfig {
	return CoalesceConfig{
		Window: 100 * time.Millisecond,
	}
}

// CoalesceConfig controls the coalescing window.
type CoalesceConfig struct {
	// Window is how long to wait for additional callers before executing.
	Window time.Duration
}

// WithCoalesce wraps a HealthCheckFn so that concurrent calls within Window
// are deduplicated: only one upstream call is made and all waiters receive
// the same result.
func WithCoalesce(cfg CoalesceConfig, fn HealthCheckFn) HealthCheckFn {
	if cfg.Window <= 0 {
		return fn
	}

	var (
		mu      sync.Mutex
		inflight *coalesceCall
	)

	return func(ctx context.Context, service string) StatusResult {
		mu.Lock()
		if inflight == nil {
			call := &coalesceCall{
				done: make(chan struct{}),
			}
			inflight = call
			mu.Unlock()

			// Wait for the coalesce window, then execute.
			select {
			case <-time.After(cfg.Window):
			case <-ctx.Done():
				mu.Lock()
				inflight = nil
				mu.Unlock()
				call.result = StatusResult{Status: StatusUnknown, Err: ctx.Err()}
				close(call.done)
				return call.result
			}

			call.result = fn(ctx, service)

			mu.Lock()
			inflight = nil
			mu.Unlock()

			close(call.done)
			return call.result
		}

		// Join the in-flight call.
		call := inflight
		mu.Unlock()

		select {
		case <-call.done:
			return call.result
		case <-ctx.Done():
			return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
		}
	}
}

type coalesceCall struct {
	done   chan struct{}
	result StatusResult
}
