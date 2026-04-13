package server

import (
	"net/http"

	"github.com/your-org/grpc-health-proxy/internal/health"
	"github.com/your-org/grpc-health-proxy/internal/metrics"
)

// NewServeMux builds and returns the HTTP mux with all routes registered.
func NewServeMux(cache *health.Cache) *http.ServeMux {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler(cache)
	logged := LoggingMiddleware(RecoveryMiddleware(metrics.RequestCountingMiddleware(healthHandler)))

	mux.Handle("/health", logged)
	mux.Handle("/healthz", LoggingMiddleware(RecoveryMiddleware(http.HandlerFunc(LivenessHandler))))
	mux.Handle("/metrics", LoggingMiddleware(RecoveryMiddleware(metrics.Handler())))

	return mux
}
