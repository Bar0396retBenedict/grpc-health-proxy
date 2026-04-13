package health

import (
	"fmt"
	"sync"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// entry holds a cached health status and any associated error.
type entry struct {
	status grpc_health_v1.HealthCheckResponse_ServingStatus
	err    error
}

// Cache is a thread-safe store for gRPC health check results keyed by service name.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]entry
}

// NewCache returns an initialised Cache.
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]entry),
	}
}

// Set stores the status and error for the given service name.
func (c *Cache) Set(service string, status grpc_health_v1.HealthCheckResponse_ServingStatus, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[service] = entry{status: status, err: err}
}

// Get returns the cached status for the given service name.
// If the service has never been set it returns UNKNOWN.
// If the cached entry contains an error, that error is returned.
func (c *Cache) Get(service string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[service]
	if !ok {
		return grpc_health_v1.HealthCheckResponse_UNKNOWN, nil
	}
	if e.err != nil {
		return e.status, fmt.Errorf("cached health error: %w", e.err)
	}
	return e.status, nil
}
