package server

import (
	"encoding/json"
	"net/http"

	"github.com/your-org/grpc-health-proxy/internal/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

type multiServiceResponse struct {
	Overall  string                    `json:"overall"`
	Services []serviceStatusEntry      `json:"services"`
}

type serviceStatusEntry struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
}

// NewMultiHealthHandler returns an HTTP handler that checks all given service
// names concurrently and responds with a JSON body summarising each result.
// It returns 200 if all services are SERVING, or 503 otherwise.
func NewMultiHealthHandler(services []string, checker func(ctx context.Context, svc string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statuses := health.CheckAll(r.Context(), checker, services)

		entries := make([]serviceStatusEntry, len(statuses))
		for i, s := range statuses {
			entry := serviceStatusEntry{
				Service: s.Service,
				Status:  s.Status.String(),
			}
			if s.Err != nil {
				entry.Error = s.Err.Error()
			}
			entries[i] = entry
		}

		overall := "SERVING"
		httpStatus := http.StatusOK
		if !health.AllServing(statuses) {
			overall = "NOT_SERVING"
			httpStatus = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		_ = json.NewEncoder(w).Encode(multiServiceResponse{
			Overall:  overall,
			Services: entries,
		})
	}
}
