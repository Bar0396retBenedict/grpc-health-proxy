package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

func TestWindowMiddleware_PassesRequestThrough(t *testing.T) {
	cfg := health.DefaultWindowConfig()
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := WindowMiddleware(cfg, "svc", inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
}

func TestWindowMiddleware_SetsWindowSuccessHeader(t *testing.T) {
	cfg := health.DefaultWindowConfig()
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := WindowMiddleware(cfg, "svc", inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Window-Success"); got == "" {
		t.Error("expected X-Window-Success header to be set")
	}
}

func TestWindowMiddleware_ServerErrorMarkedFailure(t *testing.T) {
	cfg := health.DefaultWindowConfig()
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	h := WindowMiddleware(cfg, "svc", inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Window-Success"); got != "false" {
		t.Errorf("expected X-Window-Success=false, got %q", got)
	}
}

func TestWindowMiddleware_SuccessMarkedTrue(t *testing.T) {
	cfg := health.DefaultWindowConfig()
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := WindowMiddleware(cfg, "svc", inner)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Window-Success"); got != "true" {
		t.Errorf("expected X-Window-Success=true, got %q", got)
	}
}
