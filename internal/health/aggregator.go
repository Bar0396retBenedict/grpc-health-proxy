package health

import (
	"context"
	"fmt"
	"strings"

	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

// AggregateStatus represents the combined health of multiple services.
type AggregateStatus struct {
	Overall  grpc_health_v1.HealthCheckResponse_ServingStatus
	Services map[string]grpc_health_v1.HealthCheckResponse_ServingStatus
	Errors   map[string]error
}

// IsHealthy returns true only if all services are SERVING.
func (a *AggregateStatus) IsHealthy() bool {
	return a.Overall == grpc_health_v1.HealthCheckResponse_SERVING
}

// Summary returns a human-readable summary of service statuses.
func (a *AggregateStatus) Summary() string {
	parts := make([]string, 0, len(a.Services))
	for svc, status := range a.Services {
		parts = append(parts, fmt.Sprintf("%s=%s", svc, status))
	}
	return strings.Join(parts, ", ")
}

// CheckFunc is a function that checks the health of a named service.
type CheckFunc func(ctx context.Context, service string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error)

// Aggregator checks multiple services and combines their statuses.
type Aggregator struct {
	services []string
	check    CheckFunc
}

// NewAggregator creates an Aggregator for the given services using the provided check function.
func NewAggregator(services []string, check CheckFunc) *Aggregator {
	return &Aggregator{services: services, check: check}
}

// Check runs health checks for all services and returns an AggregateStatus.
func (a *Aggregator) Check(ctx context.Context) *AggregateStatus {
	statuses := make(map[string]grpc_health_v1.HealthCheckResponse_ServingStatus, len(a.services))
	errors := make(map[string]error)

	overall := grpc_health_v1.HealthCheckResponse_SERVING

	for _, svc := range a.services {
		status, err := a.check(ctx, svc)
		statuses[svc] = status
		if err != nil {
			errors[svc] = err
			overall = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		} else if status != grpc_health_v1.HealthCheckResponse_SERVING {
			overall = grpc_health_v1.HealthCheckResponse_NOT_SERVING
		}
	}

	return &AggregateStatus{
		Overall:  overall,
		Services: statuses,
		Errors:   errors,
	}
}
