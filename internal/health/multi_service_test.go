package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

func TestCheckAll_AllServing(t *testing.T) {
	services := []string{"svcA", "svcB", "svcC"}

	checker := func(_ context.Context, svc string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
		return grpc_health_v1.HealthCheckResponse_SERVING, nil
	}

	results := health.CheckAll(context.Background(), checker, services)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("unexpected error for %s: %v", r.Service, r.Err)
		}
		if r.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			t.Errorf("expected SERVING for %s, got %v", r.Service, r.Status)
		}
	}
}

func TestCheckAll_OneUnhealthy(t *testing.T) {
	services := []string{"ok", "bad"}

	checker := func(_ context.Context, svc string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
		if svc == "bad" {
			return grpc_health_v1.HealthCheckResponse_NOT_SERVING, errors.New("down")
		}
		return grpc_health_v1.HealthCheckResponse_SERVING, nil
	}

	results := health.CheckAll(context.Background(), checker, services)

	if health.AllServing(results) {
		t.Error("expected AllServing to return false")
	}
}

func TestAllServing_Empty(t *testing.T) {
	if !health.AllServing([]health.ServiceStatus{}) {
		t.Error("expected AllServing to return true for empty slice")
	}
}

func TestAllServing_WithError(t *testing.T) {
	statuses := []health.ServiceStatus{
		{Service: "svc", Status: grpc_health_v1.HealthCheckResponse_SERVING, Err: errors.New("oops")},
	}
	if health.AllServing(statuses) {
		t.Error("expected AllServing to return false when error is present")
	}
}

func TestCheckAll_ContextCancelled(t *testing.T) {
	services := []string{"svc"}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	checker := func(ctx context.Context, svc string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
		if ctx.Err() != nil {
			return grpc_health_v1.HealthCheckResponse_UNKNOWN, ctx.Err()
		}
		return grpc_health_v1.HealthCheckResponse_SERVING, nil
	}

	results := health.CheckAll(ctx, checker, services)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Err == nil {
		t.Error("expected error due to cancelled context")
	}
}
