// Package health provides gRPC health checking utilities.
package health

import (
	"math"
	"time"
)

// BackoffConfig defines parameters for exponential backoff.
type BackoffConfig struct {
	// InitialInterval is the starting delay before the first retry.
	InitialInterval time.Duration
	// MaxInterval caps the delay between retries.
	MaxInterval time.Duration
	// Multiplier is applied to the current interval on each failure.
	Multiplier float64
	// MaxElapsed is the total time after which backoff stops. Zero means no limit.
	MaxElapsed time.Duration
}

// DefaultBackoffConfig returns a BackoffConfig with sensible defaults.
func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialInterval: 500 * time.Millisecond,
		MaxInterval:     30 * time.Second,
		Multiplier:      1.5,
		MaxElapsed:      0,
	}
}

// Backoff computes successive delay durations for a given attempt number
// (0-indexed). The result is clamped to MaxInterval.
func (b BackoffConfig) Backoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	interval := float64(b.InitialInterval) * math.Pow(b.Multiplier, float64(attempt))
	max := float64(b.MaxInterval)
	if interval > max {
		interval = max
	}
	return time.Duration(interval)
}

// ExceededMaxElapsed reports whether the given elapsed time has surpassed
// the configured MaxElapsed. It always returns false when MaxElapsed is zero.
func (b BackoffConfig) ExceededMaxElapsed(elapsed time.Duration) bool {
	if b.MaxElapsed == 0 {
		return false
	}
	return elapsed >= b.MaxElapsed
}
