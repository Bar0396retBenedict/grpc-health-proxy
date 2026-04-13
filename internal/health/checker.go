package health

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Status represents the result of a health check.
type Status int

const (
	StatusUnknown Status = iota
	StatusHealthy
	StatusUnhealthy
)

func (s Status) String() string {
	switch s {
	case StatusHealthy:
		return "HEALTHY"
	case StatusUnhealthy:
		return "UNHEALTHY"
	default:
		return "UNKNOWN"
	}
}

// Checker performs gRPC health checks against a backend service.
type Checker struct {
	addr        string
	dialTimeout time.Duration
}

// NewChecker creates a new Checker targeting the given gRPC address.
func NewChecker(addr string, dialTimeout time.Duration) *Checker {
	return &Checker{
		addr:        addr,
		dialTimeout: dialTimeout,
	}
}

// Check performs a single gRPC health check for the given service name.
// An empty serviceName checks the overall server health.
func (c *Checker) Check(ctx context.Context, serviceName string) (Status, error) {
	dialCtx, cancel := context.WithTimeout(ctx, c.dialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, c.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return StatusUnhealthy, fmt.Errorf("dial %s: %w", c.addr, err)
	}
	defer conn.Close()

	client := grpc_health_v1.NewHealthClient(conn)
	resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: serviceName,
	})
	if err != nil {
		return StatusUnhealthy, fmt.Errorf("health check rpc: %w", err)
	}

	if resp.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING {
		return StatusHealthy, nil
	}
	return StatusUnhealthy, nil
}
