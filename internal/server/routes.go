package server

import (
	"net/http"

	"github.com/your-org/grpc-health-proxy/internal/health"
	"github.com/your-org/grpc-health-proxy/internal/metrics"
)

// NewServeMux builds the HTTP router, wiring health and metrics endpoints.
// All routes are wrapped with the request-counting middleware.
func NewServeMux(cache *health.Cache, counters *metrics.Counters) *http.ServeMux {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler(cache)
	livenessHandler := http.HandlerFunc(LivenessHandler)

	mux.Handle("/healthz", metrics.RequestCountingMiddleware(counters, healthHandler))
	mux.Handle("/livez", metrics.RequestCountingMiddleware(counters, livenessHandler))
	mux.Handle("/metrics", counters.Handler())

	return mux
}
