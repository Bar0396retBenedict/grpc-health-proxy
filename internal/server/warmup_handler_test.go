package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/salmanahmad/grpc-health-proxy/internal/health"
)

func makeWarmCache(t *testing.T, entries map[string]health.Status) *health.Cache {
	t.Helper()
	c := health.NewCache()
	for svc, st := range entries {
		c.Set(svc, st, nil)
	}
	return c
}

func TestWarmupHandler_AllWarmed_Returns200(t *testing.T) {
	cache := makeWarmCache(t, map[string]health.Status{
		"svc.A": health.StatusServing,
		"svc.B": health.StatusServing,
	})
	h := NewWarmupHandler(cache, []string{"svc.A", "svc.B"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/warmup", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}

	var body WarmupStatus
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.Ready {
		t.Error("expected ready=true")
	}
}

func TestWarmupHandler_UnknownService_Returns503(t *testing.T) {
	cache := health.NewCache() // all unknown
	h := NewWarmupHandler(cache, []string{"svc.A"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/warmup", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("want 503, got %d", rec.Code)
	}

	var body WarmupStatus
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if body.Ready {
		t.Error("expected ready=false")
	}
}

func TestWarmupHandler_ContentType(t *testing.T) {
	cache := makeWarmCache(t, map[string]health.Status{"svc.A": health.StatusServing})
	h := NewWarmupHandler(cache, []string{"svc.A"})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/warmup", nil))

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("want application/json, got %s", ct)
	}
}

func TestWarmupBarrier_ReturnsWhenWarmed(t *testing.T) {
	cache := health.NewCache()
	services := []string{"svc.A"}

	go func() {
		time.Sleep(30 * time.Millisecond)
		cache.Set("svc.A", health.StatusServing, nil)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := WarmupBarrier(ctx, cache, services, 10*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWarmupBarrier_CancelledContext(t *testing.T) {
	cache := health.NewCache() // stays unknown
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := WarmupBarrier(ctx, cache, []string{"svc.A"}, 5*time.Millisecond)
	if err == nil {
		t.Fatal("expected context error")
	}
}
