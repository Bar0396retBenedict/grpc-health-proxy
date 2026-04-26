package health

import (
	"context"
	"sync"
	"time"
)

// DefaultStaleConfig returns a StaleConfig with sensible defaults.
func DefaultStaleConfig() StaleConfig {
	return StaleConfig{
		MaxAge:   30 * time.Second,
		Fallback: StatusUnknown,
	}
}

// StaleConfig controls how long a cached result is considered fresh.
type StaleConfig struct {
	// MaxAge is the maximum age of a cached result before it is considered stale.
	MaxAge time.Duration
	// Fallback is the status returned when the cached result is stale.
	Fallback Status
}

type staleEntry struct {
	result StatusResult
	at     time.Time
}

// WithStale wraps a HealthCheckFn and returns the last known result if the
// upstream check fails, provided the cached result is not older than MaxAge.
// If the cache is stale or empty and the check fails, Fallback is returned.
func WithStale(cfg StaleConfig, fn HealthCheckFn) HealthCheckFn {
	var mu sync.Mutex
	var last *staleEntry

	return func(ctx context.Context, service string) StatusResult {
		result := fn(ctx, service)

		mu.Lock()
		defer mu.Unlock()

		if result.Error == nil {
			last = &staleEntry{result: result, at: time.Now()}
			return result
		}

		// Check failed — serve stale if available and fresh enough.
		if last != nil && time.Since(last.at) <= cfg.MaxAge {
			return last.result
		}

		return StatusResult{Status: cfg.Fallback, Error: result.Error}
	}
}
