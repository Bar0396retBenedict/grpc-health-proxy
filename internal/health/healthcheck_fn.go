package health

import (
	"context"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// CheckFn is a function that performs a single health check for a service.
type CheckFn func(ctx context.Context, service string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error)

// StatusResult holds the result of a health check.
type StatusResult struct {
	Service string
	Status  grpc_health_v1.HealthCheckResponse_ServingStatus
	Err     error
}

// IsServing returns true if the status is SERVING and there is no error.
func (r StatusResult) IsServing() bool {
	return r.Err == nil && r.Status == grpc_health_v1.HealthCheckResponse_SERVING
}

// Compose wraps a CheckFn with retry, timeout, and circuit breaker middleware
// using the provided configurations.
func Compose(fn CheckFn, retryCfg RetryConfig, timeoutCfg TimeoutConfig, cb *CircuitBreaker) CheckFn {
	fn = WithTimeout(fn, timeoutCfg)
	fn = WithRetry(fn, retryCfg)
	fn = WithCircuitBreaker(fn, cb)
	return fn
}

// ComposeDefault wraps a CheckFn with default retry, timeout, and circuit
// breaker middleware.
func ComposeDefault(fn CheckFn) CheckFn {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig)
	return Compose(fn, DefaultRetryConfig, DefaultTimeoutConfig, cb)
}
