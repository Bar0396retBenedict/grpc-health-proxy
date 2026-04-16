package health

import (
	"context"
	"sync"

	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

// ServiceStatus holds the health status for a single gRPC service.
type ServiceStatus struct {
	Service string
	Status  grpc_health_v1.HealthCheckResponse_ServingStatus
	Err     error
}

// CheckAll runs health checks for all provided service names concurrently
// and returns a slice of ServiceStatus results.
func CheckAll(ctx context.Context, check func(ctx context.Context, service string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error), services []string) []ServiceStatus {
	results := make([]ServiceStatus, len(services))
	var wg sync.WaitGroup

	for i, svc := range services {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			status, err := check(ctx, name)
			results[idx] = ServiceStatus{
				Service: name,
				Status:  status,
				Err:     err,
			}
		}(i, svc)
	}

	wg.Wait()
	return results
}

// AllServing returns true if every ServiceStatus in the slice is SERVING
// and has no error.
func AllServing(statuses []ServiceStatus) bool {
	for _, s := range statuses {
		if s.Err != nil || s.Status != grpc_health_v1.HealthCheckResponse_SERVING {
			return false
		}
	}
	return true
}
