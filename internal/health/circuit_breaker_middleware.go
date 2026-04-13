package health

import (
	"context"
	"fmt"
)

// CheckFunc is a function that performs a health check.
type CheckFunc func(ctx context.Context) error

// WithCircuitBreaker wraps a CheckFunc with circuit breaker protection.
// If the circuit breaker is open, the check is skipped and an error is returned.
// Results of the check are recorded to update the circuit breaker state.
func WithCircuitBreaker(cb *CircuitBreaker, fn CheckFunc) CheckFunc {
	return func(ctx context.Context) error {
		if !cb.Allow() {
			return fmt.Errorf("circuit breaker open: too many consecutive failures")
		}

		err := fn(ctx)
		if err != nil {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
		return err
	}
}

// WithRetryAndCircuitBreaker composes retry logic with circuit breaker protection.
// The circuit breaker wraps the entire retry sequence, not individual attempts.
func WithRetryAndCircuitBreaker(cb *CircuitBreaker, cfg RetryConfig, fn CheckFunc) CheckFunc {
	protected := WithCircuitBreaker(cb, fn)
	return WithRetry(cfg, protected)
}
