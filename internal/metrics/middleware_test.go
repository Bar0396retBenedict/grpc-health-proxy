package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestCountingMiddleware(t *testing.T) {
	c := &Counters{}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := RequestCountingMiddleware(c, inner)

	for i := 0; i < 3; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("iteration %d: want 200, got %d", i, rec.Code)
		}
	}

	if got := c.HTTPRequestTotal.Load(); got != 3 {
		t.Errorf("http_request_total: want 3, got %d", got)
	}
}

func TestRequestCountingMiddleware_DoesNotAffectInnerResponse(t *testing.T) {
	c := &Counters{}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	handler := RequestCountingMiddleware(c, inner)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("inner status should be preserved: want 418, got %d", rec.Code)
	}
	if c.HTTPRequestTotal.Load() != 1 {
		t.Error("counter should still be incremented")
	}
}
