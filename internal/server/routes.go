package server

import (
	"fmt"
	"net/http"

	"github.com/grpc-health-proxy/internal/config"
	"github.com/grpc-health-proxy/internal/health"
)

// NewServeMux builds and returns an HTTP mux with all registered routes.
//
// Routes:
//
//	GET /healthz            — liveness probe (always 200)
//	GET /ready              — readiness based on default service ("") health
//	GET /health/{service}   — readiness for a named gRPC service
func NewServeMux(cfg *config.Config, cache *health.Cache) *http.ServeMux {
	mux := http.NewServeMux()

	// Liveness
	mux.HandleFunc("/healthz", LivenessHandler)

	// Readiness for the default (empty) service name
	defaultHandler := NewHealthHandler(cache, "")
	mux.Handle("/ready", defaultHandler)

	// Per-service readiness endpoints
	for _, svc := range cfg.Services {
		path := fmt.Sprintf("/health/%s", svc)
		mux.Handle(path, NewHealthHandler(cache, svc))
	}

	return mux
}
