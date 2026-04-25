package server

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

// mirrorResponse is the JSON payload returned by the mirror endpoint.
type mirrorResponse struct {
	Primary   string `json:"primary"`
	Mirror    string `json:"mirror"`
	Agreement bool   `json:"agreement"`
	Error     string `json:"error,omitempty"`
}

// NewMirrorHandler returns an HTTP handler that runs a mirrored health check
// and reports both outcomes as JSON. A 200 is returned only when both sides
// agree on a serving status; otherwise 503 is returned.
func NewMirrorHandler(service string, check func(ctx interface{ Done() <-chan struct{} }, svc string) health.MirrorResult) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result := check(r.Context(), service)

		resp := mirrorResponse{
			Primary:   result.Primary.Status.String(),
			Mirror:    result.Mirror.Status.String(),
			Agreement: result.Agreement,
		}

		if result.Primary.Err != nil {
			resp.Error = result.Primary.Err.Error()
		} else if result.Mirror.Err != nil {
			resp.Error = result.Mirror.Err.Error()
		}

		w.Header().Set("Content-Type", "application/json")

		if result.Agreement && result.Primary.Status == health.StatusServing {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		_ = json.NewEncoder(w).Encode(resp)
	}
}
