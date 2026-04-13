package server

import (
	"net/http"

	"github.com/grpc-health-proxy/internal/health"
	"github.com/grpc-health-proxy/internal/metrics"
)

// NewServeMux builds and returns the HTTP mux with all routes registered.
func NewServeMux(cache *health.Cache, serviceName string) *http.ServeMux {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler(cache, serviceName)
	mux.Handle("/healthz", metrics.RequestCountingMiddleware(healthHandler))
	mux.Handle("/livez", metrics.RequestCountingMiddleware(http.HandlerFunc(LivenessHandler)))
	mux.Handle("/metrics", metrics.Handler())

	return mux
}
