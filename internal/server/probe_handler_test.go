package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

func servingFn(_ context.Context, _ string) health.StatusResult {
	return health.StatusResult{Status: health.StatusServing}
}

func notServingFn(_ context.Context, _ string) health.StatusResult {
	return health.StatusResult{Status: health.StatusNotServing}
}

func errFn(_ context.Context, _ string) health.StatusResult {
	return health.StatusResult{Status: health.StatusUnknown, Err: errors.New("connection refused")}
}

func TestProbeHandler_Healthy(t *testing.T) {
	probe := health.NewProbe(health.ProbeConfig{SuccessThreshold: 1, FailureThreshold: 1}, servingFn)
	h := NewProbeHandler(probe, "my-service")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); body != "ok\n" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestProbeHandler_Unhealthy(t *testing.T) {
	probe := health.NewProbe(health.ProbeConfig{SuccessThreshold: 1, FailureThreshold: 1}, notServingFn)
	h := NewProbeHandler(probe, "my-service")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestProbeHandler_ErrorPropagated(t *testing.T) {
	probe := health.NewProbe(health.ProbeConfig{SuccessThreshold: 1, FailureThreshold: 1}, errFn)
	h := NewProbeHandler(probe, "my-service")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
	if body := rec.Body.String(); body == "" {
		t.Error("expected non-empty error body")
	}
}

func TestReadyzHandler_Healthy(t *testing.T) {
	h := ReadyzHandler(servingFn, "svc")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestReadyzHandler_Unhealthy(t *testing.T) {
	h := ReadyzHandler(notServingFn, "svc")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestReadyzHandler_ErrorBody(t *testing.T) {
	h := ReadyzHandler(errFn, "svc")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
	if body := rec.Body.String(); body == "" {
		t.Error("expected error message in body")
	}
}
