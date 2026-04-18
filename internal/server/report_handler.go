package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

// ReportHandler serves JSON health reports for a service.
type ReportHandler struct {
	reporter *health.Reporter
}

// NewReportHandler creates a handler backed by the given reporter.
func NewReportHandler(reporter *health.Reporter) *ReportHandler {
	return &ReportHandler{reporter: reporter}
}

func (h *ReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	if service == "" {
		service = ""
	}

	n := 5
	if v := r.URL.Query().Get("events"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}

	rep := h.reporter.Report(service)
	events := h.reporter.RecentEvents(service, n)

	type eventJSON struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
		Time   string `json:"time"`
	}
	type response struct {
		Service string      `json:"service"`
		Status  string      `json:"status"`
		Error   string      `json:"error,omitempty"`
		Events  []eventJSON `json:"events"`
	}

	evJSON := make([]eventJSON, 0, len(events))
	for _, e := range events {
		ej := eventJSON{Status: e.Status.String(), Time: e.OccurredAt.Format("2006-01-02T15:04:05Z07:00")}
		if e.Err != nil {
			ej.Error = e.Err.Error()
		}
		evJSON = append(evJSON, ej)
	}

	resp := response{Service: rep.Service, Status: rep.Status.String(), Events: evJSON}
	if rep.Err != nil {
		resp.Error = rep.Err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
