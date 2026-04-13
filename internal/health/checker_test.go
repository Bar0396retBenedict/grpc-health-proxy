package health

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func startFakeServer(t *testing.T, status grpc_health_v1.HealthCheckResponse_ServingStatus) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", status)
	healthSrv.SetServingStatus("my.Service", status)
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	go srv.Serve(lis) //nolint:errcheck
	t.Cleanup(srv.GracefulStop)
	return lis.Addr().String()
}

func TestChecker_Healthy(t *testing.T) {
	addr := startFakeServer(t, grpc_health_v1.HealthCheckResponse_SERVING)
	checker := NewChecker(addr)

	status, err := checker.Check(context.Background(), "my.Service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Fatalf("expected SERVING, got %v", status)
	}
}

func TestChecker_Unhealthy(t *testing.T) {
	addr := startFakeServer(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	checker := NewChecker(addr)

	status, err := checker.Check(context.Background(), "my.Service")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != grpc_health_v1.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("expected NOT_SERVING, got %v", status)
	}
}

func TestChecker_DialFailure(t *testing.T) {
	checker := NewChecker("127.0.0.1:1")
	// Override circuit breaker threshold to open quickly.
	checker.cb = NewCircuitBreaker(CircuitBreakerConfig{FailureThreshold: 1, SuccessThreshold: 1, OpenDuration: 60})

	ctx, cancel := context.WithTimeout(context.Background(), 0)
	defer cancel()
	<-ctx.Done()

	_, err := checker.Check(ctx, "")
	if err == nil {
		t.Fatal("expected error for unreachable address")
	}
}

func TestChecker_CircuitBreakerBlocks(t *testing.T) {
	checker := NewChecker("127.0.0.1:1")
	cfg := CircuitBreakerConfig{FailureThreshold: 1, SuccessThreshold: 1, OpenDuration: 10_000_000_000}
	checker.cb = NewCircuitBreaker(cfg)
	// Force the breaker open.
	checker.cb.RecordFailure()

	_, err := checker.Check(context.Background(), "")
	if err == nil {
		t.Fatal("expected circuit breaker to block the request")
	}
}
