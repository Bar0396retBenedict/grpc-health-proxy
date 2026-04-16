package health

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/health/grpc_health_v1"
)

func servingFn(_ context.Context, _ string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	return grpc_health_v1.HealthCheckResponse_SERVING, nil
}

func failingFn(_ context.Context, _ string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	return grpc_health_v1.HealthCheckResponse_NOT_SERVING, errors.New("unavailable")
}

func TestStatusResult_IsServing_True(t *testing.T) {
	r := StatusResult{Status: grpc_health_v1.HealthCheckResponse_SERVING}
	if !r.IsServing() {
		t.Fatal("expected IsServing to be true")
	}
}

func TestStatusResult_IsServing_FalseOnError(t *testing.T) {
	r := StatusResult{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
		Err:    errors.New("err"),
	}
	if r.IsServing() {
		t.Fatal("expected IsServing to be false when error is set")
	}
}

func TestStatusResult_IsServing_FalseOnNotServing(t *testing.T) {
	r := StatusResult{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}
	if r.IsServing() {
		t.Fatal("expected IsServing to be false for NOT_SERVING")
	}
}

func TestComposeDefault_Serving(t *testing.T) {
	fn := ComposeDefault(servingFn)
	status, err := fn(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected SERVING, got %v", status)
	}
}

func TestCompose_CustomConfig(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig)
	fn := Compose(servingFn, DefaultRetryConfig, DefaultTimeoutConfig, cb)
	status, err := fn(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected SERVING, got %v", status)
	}
}

func TestComposeDefault_FailingFn(t *testing.T) {
	fn := ComposeDefault(failingFn)
	_, err := fn(context.Background(), "svc")
	if err == nil {
		t.Fatal("expected error from failing fn")
	}
}
