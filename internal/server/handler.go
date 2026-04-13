package server

import (
	"encoding/json"
	"net/http"

	"github.com/grpc-health-proxy/internal/health"
)

// HealthHandler handles HTTP health check requests and maps them
// to gRPC health check results stored in the cache.
type HealthHandler struct {
	cache   *health.Cache
	service string
}

// NewHealthHandler creates a new HealthHandler for the given service name.
func NewHealthHandler(cache *health.Cache, service string) *HealthHandler {
	return &HealthHandler{
		cache:   cache,
		service: service,
	}
}

type healthResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

// ServeHTTP responds with the current health status of the configured service.
// Returns 200 OK when healthy, 503 Service Unavailable otherwise.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status, err := h.cache.Get(h.service)

	resp := healthResponse{
		Service: h.service,
		Status:  status.String(),
	}

	if err != nil {
		resp.Error = err.Error()
	}

	if status.IsServing() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// LivenessHandler always returns 200 OK — used for basic liveness probes.
func LivenessHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
