package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/salmanahmad/grpc-health-proxy/internal/health"
)

// WarmupStatus is the JSON body returned by the warmup endpoint.
type WarmupStatus struct {
	Ready    bool              `json:"ready"`
	Services map[string]string `json:"services"`
	Elapsed  string            `json:"elapsed"`
}

// NewWarmupHandler returns an HTTP handler that reports whether the cache
// has been warmed for all given services. It responds 200 once every
// service has a non-Unknown status, and 503 otherwise.
func NewWarmupHandler(cache *health.Cache, services []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		statuses := make(map[string]string, len(services))
		ready := true

		for _, svc := range services {
			status, err := cache.Get(svc)
			if err != nil {
				statuses[svc] = "error: " + err.Error()
				ready = false
				continue
			}
			statuses[svc] = status.String()
			if status == health.StatusUnknown {
				ready = false
			}
		}

		body := WarmupStatus{
			Ready:    ready,
			Services: statuses,
			Elapsed:  time.Since(start).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(body)
	}
}

// WarmupBarrier blocks until all services in the cache have a non-Unknown
// status or ctx is cancelled. It is intended to be called during startup
// before the server begins accepting traffic.
func WarmupBarrier(ctx context.Context, cache *health.Cache, services []string, poll time.Duration) error {
	ticker := time.NewTicker(poll)
	defer ticker.Stop()
	for {
		if allWarmed(cache, services) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func allWarmed(cache *health.Cache, services []string) bool {
	for _, svc := range services {
		status, _ := cache.Get(svc)
		if status == health.StatusUnknown {
			return false
		}
	}
	return true
}
