package server

import (
	"context"
	"net/http"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

// ProbeCache is the subset of the health cache needed by the probe handler.
type ProbeCache interface {
	Get(service string) health.StatusResult
}

// ProbeHandler serves HTTP readiness and liveness probe endpoints backed by
// a Probe instance. It delegates the actual gRPC check to the probe on each
// request so that thresholds are honoured across successive HTTP polls.
type ProbeHandler struct {
	probe   *health.Probe
	service string
}

// NewProbeHandler creates a ProbeHandler for the given service name.
func NewProbeHandler(probe *health.Probe, service string) *ProbeHandler {
	return &ProbeHandler{probe: probe, service: service}
}

// ServeHTTP handles an HTTP probe request.
//
//   - 200 OK            — probe considers the service ready
//   - 503 Service Unavailable — probe considers the service not ready
func (h *ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.probe.Check(ctx, h.service); err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// ReadyzHandler returns an http.HandlerFunc that performs a one-shot readiness
// check using the supplied HealthCheckFn without threshold tracking.
func ReadyzHandler(fn health.HealthCheckFn, service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := fn(r.Context(), service)
		if !result.IsServing() {
			msg := "not serving"
			if result.Err != nil {
				msg = result.Err.Error()
			}
			http.Error(w, msg, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	}
}

// ensure ProbeHandler satisfies http.Handler at compile time.
var _ http.Handler = (*ProbeHandler)(nil)

// suppress unused import warning for context (used via r.Context())
var _ = context.Background
