package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

func makeReporter(service string, status health.Status, err error) *health.Reporter {
	cache := health.NewCache()
	cache.Set(service, status, err)
	h := health.NewHistory(10)
	h.Record(service, health.NewEvent(service, status, err, time.Now()))
	return health.NewReporter(cache, h)
}

func TestReportHandler_ServingStatus(t *testing.T) {
	rep := makeReporter("myservice", health.StatusServing, nil)
	h := NewReportHandler(rep)

	req := httptest.NewRequest(http.MethodGet, "/report?service=myservice", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "Serving" {
		t.Errorf("expected Serving, got %v", body["status"])
	}
	if body["service"] != "myservice" {
		t.Errorf("expected myservice, got %v", body["service"])
	}
}

func TestReportHandler_ContentType(t *testing.T) {
	rep := makeReporter("", health.StatusUnknown, nil)
	h := NewReportHandler(rep)

	req := httptest.NewRequest(http.MethodGet, "/report", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}
}

func TestReportHandler_EventsInResponse(t *testing.T) {
	rep := makeReporter("svc", health.StatusServing, nil)
	h := NewReportHandler(rep)

	req := httptest.NewRequest(http.MethodGet, "/report?service=svc&events=2", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	var body map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&body)
	events, ok := body["events"].([]interface{})
	if !ok {
		t.Fatal("expected events array")
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}
