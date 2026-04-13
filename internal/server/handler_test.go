package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-health-proxy/internal/health"
	"github.com/grpc-health-proxy/internal/server"
)

func TestHealthHandler_Healthy(t *testing.T) {
	cache := health.NewCache()
	cache.Set("myservice", health.StatusServing, nil)

	h := server.NewHealthHandler(cache, "myservice")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/myservice", nil)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "SERVING" {
		t.Errorf("expected status SERVING, got %s", body["status"])
	}
}

func TestHealthHandler_Unhealthy(t *testing.T) {
	cache := health.NewCache()
	cache.Set("myservice", health.StatusNotServing, nil)

	h := server.NewHealthHandler(cache, "myservice")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/myservice", nil)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestHealthHandler_UnknownService(t *testing.T) {
	cache := health.NewCache()
	// nothing set — should default to unknown / not serving

	h := server.NewHealthHandler(cache, "unknown")
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health/unknown", nil)

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for unknown service, got %d", rec.Code)
	}
}

func TestLivenessHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	server.LivenessHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
