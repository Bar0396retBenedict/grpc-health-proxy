package health_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/yourorg/grpc-health-proxy/internal/health"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// fakeHealthServer implements grpc_health_v1.HealthServer for testing.
type fakeHealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	status grpc_health_v1.HealthCheckResponse_ServingStatus
}

func (f *fakeHealthServer) Check(
	_ context.Context,
	_ *grpc_health_v1.HealthCheckRequest,
) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: f.status}, nil
}

func startFakeServer(t *testing.T, status grpc_health_v1.HealthCheckResponse_ServingStatus) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, &fakeHealthServer{status: status})
	go srv.Serve(lis) //nolint:errcheck
	t.Cleanup(srv.GracefulStop)
	return lis.Addr().String()
}

func TestChecker_Healthy(t *testing.T) {
	addr := startFakeServer(t, grpc_health_v1.HealthCheckResponse_SERVING)
	checker := health.NewChecker(addr, 2*time.Second)

	status, err := checker.Check(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != health.StatusHealthy {
		t.Errorf("expected HEALTHY, got %s", status)
	}
}

func TestChecker_Unhealthy(t *testing.T) {
	addr := startFakeServer(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	checker := health.NewChecker(addr, 2*time.Second)

	status, err := checker.Check(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != health.StatusUnhealthy {
		t.Errorf("expected UNHEALTHY, got %s", status)
	}
}

func TestChecker_DialFailure(t *testing.T) {
	checker := health.NewChecker("127.0.0.1:1", 200*time.Millisecond)

	status, err := checker.Check(context.Background(), "")
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if status != health.StatusUnhealthy {
		t.Errorf("expected UNHEALTHY on dial failure, got %s", status)
	}
}
