// Package health provides gRPC health checking primitives.
package health

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// CheckFn is the signature for a health-check function.
type CheckFn func(ctx context.Context, service string) (healthy bool, err error)

// CheckerConfig holds options for NewChecker.
type CheckerConfig struct {
	Addrstring
	DialTLSC    credentials.TransportCredentials
}

// NewChecker returns a CheckFn that queries the gRPC health service at the
// address specified in cfg.
func NewChecker(cfg CheckerConfig) CheckFn {
	return func(ctx context.Context, service string) (bool, error) {
		dialCtx, cancel := context.WithTimeout(ctx, cfg.DialTimeout)
		defer cancel()

		creds := cfg.TLSCreds
		if creds == nil {
			creds = insecure.NewCredentials()
		}

		conn, err := grpc.DialContext(dialCtx, cfg.Addr, grpc.WithTransportCredentials(creds)) //nolint:staticcheck
		if err != nil {
			return false, fmt.Errorf("dial %s: %w", cfg.Addr, err)
		}
		defer conn.Close()

		client := grpc_health_v1.NewHealthClient(conn)
		resp, err := client.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: service})
		if err != nil {
			return false, fmt.Errorf("health check rpc: %w", err)
		}

		return resp.GetStatus() == grpc_health_v1.HealthCheckResponse_SERVING, nil
	}
}
