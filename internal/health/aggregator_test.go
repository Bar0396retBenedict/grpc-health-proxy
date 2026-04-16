package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

func serving(_ context.Context, _ string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	return grpc_health_v1.HealthCheckResponse_SERVING, nil
}

func notServing(_ context.Context, _ string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	return grpc_health_v1.HealthCheckResponse_NOT_SERVING, nil
}

func errCheck(_ context.Context, _ string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	return grpc_health_v1.HealthCheckResponse_UNKNOWN, errors.New("dial error")
}

func TestAggregator_AllServing(t *testing.T) {
	a := health.NewAggregator([]string{"svc1", "svc2"}, serving)
	result := a.Check(context.Background())
	if !result.IsHealthy() {
		t.Errorf("expected healthy, got %v", result.Overall)
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected no errors, got %v", result.Errors)
	}
}

func TestAggregator_OneNotServing(t *testing.T) {
	calls := 0
	check := func(ctx context.Context, svc string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
		calls++
		if svc == "bad" {
			return notServing(ctx, svc)
		}
		return serving(ctx, svc)
	}
	a := health.NewAggregator([]string{"good", "bad"}, check)
	result := a.Check(context.Background())
	if result.IsHealthy() {
		t.Error("expected unhealthy")
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestAggregator_ErrorMakesUnhealthy(t *testing.T) {
	a := health.NewAggregator([]string{"svc"}, errCheck)
	result := a.Check(context.Background())
	if result.IsHealthy() {
		t.Error("expected unhealthy on error")
	}
	if _, ok := result.Errors["svc"]; !ok {
		t.Error("expected error recorded for svc")
	}
}

func TestAggregator_EmptyServices(t *testing.T) {
	a := health.NewAggregator([]string{}, serving)
	result := a.Check(context.Background())
	if !result.IsHealthy() {
		t.Error("expected healthy with no services")
	}
}

func TestAggregateStatus_Summary(t *testing.T) {
	a := health.NewAggregator([]string{"alpha"}, serving)
	result := a.Check(context.Background())
	summary := result.Summary()
	if summary == "" {
		t.Error("expected non-empty summary")
	}
}
