package server

import (
	"context"
	"net/http"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

// PipelineHandler serves a single service health check built from a
// health.HealthcheckFn pipeline. It writes 200 when Serving, 503 otherwise.
type PipelineHandler struct {
	service string
	fn      health.HealthcheckFn
}

// NewPipelineHandler creates an http.Handler that evaluates fn for service on
// every request. fn is typically constructed via health.Pipeline.
func NewPipelineHandler(service string, fn health.HealthcheckFn) http.Handler {
	return &PipelineHandler{service: service, fn: fn}
}

func (h *PipelineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	result := h.fn(ctx, h.service)

	w.Header().Set("Content-Type", "application/json")
	if result.IsServing() {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"SERVING"}`))
		return
	}

	w.WriteHeader(http.StatusServiceUnavailable)
	msg := "NOT_SERVING"
	if result.Error != nil {
		msg = result.Error.Error()
	}
	_, _ = w.Write([]byte(`{"status":"` + msg + `"}`))
}
