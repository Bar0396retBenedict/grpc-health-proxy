package health

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Checker performs gRPC health checks against a remote service.
type Checker struct {
	addr    string
	cb      *CircuitBreaker
	dialOpt grpc.DialOption
}

// NewChecker creates a Checker targeting the given gRPC address.
func NewChecker(addr string) *Checker {
	return &Checker{
		addr:    addr,
		cb:      NewCircuitBreaker(DefaultCircuitBreakerConfig()),
		dialOpt: grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
}

// Check performs a single gRPC health check for the named service.
// An empty service name checks the overall server health.
func (c *Checker) Check(ctx context.Context, service string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	if !c.cb.Allow() {
		return grpc_health_v1.HealthCheckResponse_UNKNOWN,
			fmt.Errorf("circuit breaker open: too many recent failures for %s", c.addr)
	}

	conn, err := grpc.NewClient(c.addr, c.dialOpt)
	if err != nil {
		c.cb.RecordFailure()
		return grpc_health_v1.HealthCheckResponse_UNKNOWN, fmt.Errorf("dial %s: %w", c.addr, err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
	if err != nil {
		c.cb.RecordFailure()
		return grpc_health_v1.HealthCheckResponse_UNKNOWN, fmt.Errorf("health check %s: %w", service, err)
	}

	c.cb.RecordSuccess()
	return resp.GetStatus(), nil
}
