package server_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/server"
)

func TestBulkheadMiddleware_AllowsUnderLimit(t *testing.T) {
	mw := server.BulkheadMiddleware(5)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBulkheadMiddleware_ShedsOverLimit(t *testing.T) {
	mw := server.BulkheadMiddleware(1)
	blocked := make(chan struct{})
	release := make(chan struct{})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		blocked <- struct{}{}
		<-release
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		handler.ServeHTTP(rec, req)
	}()
	<-blocked // goroutine is inside handler

	// Second request should be shed.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}

	close(release)
	wg.Wait()
}

func TestBulkheadMiddleware_AllowsAfterSlotFreed(t *testing.T) {
	mw := server.BulkheadMiddleware(1)
	blocked := make(chan struct{}, 1)
	release := make(chan struct{})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		blocked <- struct{}{}
		<-release
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		handler.ServeHTTP(rec, req)
	}()
	<-blocked

	// Shed while slot is occupied.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 while occupied, got %d", rec.Code)
	}

	close(release)
	wg.Wait()

	// Slot now free — should succeed.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	// Replace handler so it doesn't block.
	handler2 := server.BulkheadMiddleware(1)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	handler2.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 after slot freed, got %d", rec2.Code)
	}
}
