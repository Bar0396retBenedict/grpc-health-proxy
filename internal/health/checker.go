package health

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Checker performs gRPC health checks against a remote service.
type Checker struct {
	addr    string
	service string
	creds   credentials.TransportCredentials
	retry   RetryConfig
}

// NewChecker creates a Checker for the given address and service name.
func NewChecker(addr, service string, creds credentials.TransportCredentials, retry RetryConfig) *Checker {
	if creds == nil {
		creds = insecure.NewCredentials()
	}
	return &Checker{addr: addr, service: service, creds: creds, retry: retry}
}

// Check performs a single health check, retrying on transient failures.
// It returns nil when the service reports SERVING.
func (c *Checker) Check(ctx context.Context) error {
	return WithRetry(ctx, c.retry, func(ctx context.Context) error {
		conn, err := grpc.NewClient(c.addr, grpc.WithTransportCredentials(c.creds))
		if err != nil {
			return fmt.Errorf("dial %s: %w", c.addr, err)
		}
		defer conn.Close()

		client := grpc_health_v1.NewHealthClient(conn)
		resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{
			Service: c.service,
		})
		if err != nil {
			return fmt.Errorf("health check rpc: %w", err)
		}
		if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
			return fmt.Errorf("service %q status: %s", c.service, resp.GetStatus())
		}
		return nil
	})
}
