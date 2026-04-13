package health_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestPoller_UpdatesCache(t *testing.T) {
	addr, stop := startFakeServer(t, grpc_health_v1.HealthCheckResponse_SERVING)
	defer stop()

	cache := NewCache()
	checker, err := NewChecker(addr, 2*time.Second)
	if err != nil {
		t.Fatalf("NewChecker: %v", err)
	}
	defer checker.Close()

	poller := NewPoller(checker, cache, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	poller.Start(ctx, "")

	// Wait for at least one poll cycle
	time.Sleep(150 * time.Millisecond)

	status, err := cache.Get("")
	if err != nil {
		t.Fatalf("unexpected error from cache: %v", err)
	}
	if status != grpc_health_v1.HealthCheckResponse_SERVING {
		t.Errorf("expected SERVING, got %v", status)
	}
}

func TestPoller_StopsOnContextCancel(t *testing.T) {
	addr, stop := startFakeServer(t, grpc_health_v1.HealthCheckResponse_SERVING)
	defer stop()

	cache := NewCache()
	checker, err := NewChecker(addr, 2*time.Second)
	if err != nil {
		t.Fatalf("NewChecker: %v", err)
	}
	defer checker.Close()

	poller := NewPoller(checker, cache, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		poller.Start(ctx, "")
		close(done)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Error("poller did not stop after context cancellation")
	}
}

func TestPoller_UnhealthyUpdatesCache(t *testing.T) {
	addr, stop := startFakeServer(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	defer stop()

	cache := NewCache()
	checker, err := NewChecker(addr, 2*time.Second)
	if err != nil {
		t.Fatalf("NewChecker: %v", err)
	}
	defer checker.Close()

	poller := NewPoller(checker, cache, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	poller.Start(ctx, "")
	time.Sleep(150 * time.Millisecond)

	_, err = cache.Get("")
	if err == nil {
		t.Error("expected error for NOT_SERVING, got nil")
	}
}
