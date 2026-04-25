package health

import (
	"context"
	"errors"
	"sync/atomic"
)

// ErrBulkheadFull is returned when the bulkhead has reached its concurrency limit.
var ErrBulkheadFull = errors.New("bulkhead: max concurrent calls reached")

// BulkheadConfig controls the maximum number of concurrent health check calls.
type BulkheadConfig struct {
	// MaxConcurrent is the maximum number of in-flight calls allowed at once.
	MaxConcurrent int64
}

// DefaultBulkheadConfig returns a BulkheadConfig with sensible defaults.
func DefaultBulkheadConfig() BulkheadConfig {
	return BulkheadConfig{
		MaxConcurrent: 8,
	}
}

// WithBulkhead wraps a HealthCheckFn so that at most cfg.MaxConcurrent calls
// may execute concurrently. Excess callers receive ErrBulkheadFull immediately
// rather than queuing, keeping latency predictable.
func WithBulkhead(fn HealthCheckFn, cfg BulkheadConfig) HealthCheckFn {
	var inflight atomic.Int64

	return func(ctx context.Context, service string) StatusResult {
		current := inflight.Add(1)
		defer inflight.Add(-1)

		if current > cfg.MaxConcurrent {
			return StatusResult{Status: StatusUnknown, Err: ErrBulkheadFull}
		}

		return fn(ctx, service)
	}
}
