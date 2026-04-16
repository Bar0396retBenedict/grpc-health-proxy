package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/server"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

func TestMultiHealthHandler_AllServing(t *testing.T) {
	checker := func(_ context.Context, svc string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
		return grpc_health_v1.HealthCheckResponse_SERVING, nil
	}

	h := server.NewMultiHealthHandler([]string{"svcA", "svcB"}, checker)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz/multi", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["overall"] != "SERVING" {
		t.Errorf("expected overall SERVING, got %v", body["overall"])
	}
}

func TestMultiHealthHandler_OneUnhealthy(t *testing.T) {
	checker := func(_ context.Context, svc string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
		if svc == "bad" {
			return grpc_health_v1.HealthCheckResponse_NOT_SERVING, errors.New("unavailable")
		}
		return grpc_health_v1.HealthCheckResponse_SERVING, nil
	}

	h := server.NewMultiHealthHandler([]string{"good", "bad"}, checker)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz/multi", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["overall"] != "NOT_SERVING" {
		t.Errorf("expected overall NOT_SERVING, got %v", body["overall"])
	}
}

func TestMultiHealthHandler_ContentType(t *testing.T) {
	checker := func(_ context.Context, svc string) (grpc_health_v1.HealthCheckResponse_ServingStatus, error) {
		return grpc_health_v1.HealthCheckResponse_SERVING, nil
	}

	h := server.NewMultiHealthHandler([]string{"svc"}, checker)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz/multi", nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}
