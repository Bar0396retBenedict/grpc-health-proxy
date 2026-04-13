package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/grpc-health-proxy/internal/health"
	"github.com/grpc-health-proxy/internal/server"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestNewServeMux_HealthzRoute(t *testing.T) {
	cache := health.NewCache()
	cache.Set("", grpc_health_v1.HealthCheckResponse_SERVING, nil)

	mux := server.NewServeMux(cache, "")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /healthz, got %d", rec.Code)
	}
}

func TestNewServeMux_LivezRoute(t *testing.T) {
	cache := health.NewCache()
	mux := server.NewServeMux(cache, "")

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /livez, got %d", rec.Code)
	}
}

func TestNewServeMux_MetricsRoute(t *testing.T) {
	cache := health.NewCache()
	mux := server.NewServeMux(cache, "")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /metrics, got %d", rec.Code)
	}
}

func TestNewServeMux_UnknownRoute(t *testing.T) {
	cache := health.NewCache()
	mux := server.NewServeMux(cache, "")

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for /unknown, got %d", rec.Code)
	}
}

func TestNewServeMux_UnhealthyService(t *testing.T) {
	cache := health.NewCache()
	cache.Set("", grpc_health_v1.HealthCheckResponse_NOT_SERVING, nil)

	mux := server.NewServeMux(cache, "")

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 for /healthz when not serving, got %d", rec.Code)
	}
}
