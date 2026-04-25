package health

import (
	"context"
	"fmt"
	"time"
)

// DeadlineConfig configures the deadline enforcement wrapper.
type DeadlineConfig struct {
	// Deadline is the absolute duration from now after which the check is cancelled.
	Deadline time.Duration
	// Label is used for error messages to identify the deadline source.
	Label string
}

// DefaultDeadlineConfig returns a sensible default deadline configuration.
func DefaultDeadlineConfig() DeadlineConfig {
	return DeadlineConfig{
		Deadline: 10 * time.Second,
		Label:    "deadline",
	}
}

// WithDeadline wraps a HealthCheckFn and enforces an absolute deadline on each
// invocation. Unlike WithTimeout, which sets a per-call timeout relative to
// invocation time, WithDeadline sets a fixed duration from the moment the
// wrapper is called. If the parent context already has a shorter deadline,
// that deadline takes precedence.
func WithDeadline(cfg DeadlineConfig, fn HealthCheckFn) HealthCheckFn {
	return func(ctx context.Context, service string) StatusResult {
		if cfg.Deadline <= 0 {
			return fn(ctx, service)
		}

		deadlineCtx, cancel := context.WithDeadline(ctx, time.Now().Add(cfg.Deadline))
		defer cancel()

		type result struct {
			sr StatusResult
		}
		ch := make(chan result, 1)

		go func() {
			ch <- result{sr: fn(deadlineCtx, service)}
		}()

		select {
		case r := <-ch:
			return r.sr
		case <-deadlineCtx.Done():
			return StatusResult{
				Status: StatusUnknown,
				Err:    fmt.Errorf("%s: deadline exceeded for service %q: %w", cfg.Label, service, deadlineCtx.Err()),
			}
		}
	}
}
