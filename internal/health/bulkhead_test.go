package health_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

func TestWithBulkhead_AllowsUnderLimit(t *testing.T) {
	cfg := health.BulkheadConfig{MaxConcurrent: 3}
	calls := 0
	fn := func(_ context.Context, _ string) health.StatusResult {
		calls++
		return health.StatusResult{Status: health.StatusServing}
	}
	wrapped := health.WithBulkhead(fn, cfg)

	result := wrapped(context.Background(), "svc")
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Status != health.StatusServing {
		t.Fatalf("expected Serving, got %v", result.Status)
	}
}

func TestWithBulkhead_ShedsOverLimit(t *testing.T) {
	cfg := health.BulkheadConfig{MaxConcurrent: 2}
	blocked := make(chan struct{})
	release := make(chan struct{})

	fn := func(_ context.Context, _ string) health.StatusResult {
		blocked <- struct{}{}
		<-release
		return health.StatusResult{Status: health.StatusServing}
	}
	wrapped := health.WithBulkhead(fn, cfg)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wrapped(context.Background(), "svc")
		}()
		<-blocked // ensure goroutine is inside fn
	}

	// Third call should be shed.
	result := wrapped(context.Background(), "svc")
	if result.Err != health.ErrBulkheadFull {
		t.Fatalf("expected ErrBulkheadFull, got %v", result.Err)
	}

	close(release)
	wg.Wait()
}

func TestWithBulkhead_AllowsAfterSlotFreed(t *testing.T) {
	cfg := health.BulkheadConfig{MaxConcurrent: 1}
	blocked := make(chan struct{}, 1)
	release := make(chan struct{})

	fn := func(_ context.Context, _ string) health.StatusResult {
		blocked <- struct{}{}
		<-release
		return health.StatusResult{Status: health.StatusServing}
	}
	wrapped := health.WithBulkhead(fn, cfg)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		wrapped(context.Background(), "svc")
	}()
	<-blocked

	// Slot occupied — shed.
	if r := wrapped(context.Background(), "svc"); r.Err != health.ErrBulkheadFull {
		t.Fatalf("expected shed, got %v", r.Err)
	}

	close(release)
	wg.Wait()

	// Slot now free — should succeed.
	result := wrapped(context.Background(), "svc")
	if result.Err != nil {
		t.Fatalf("expected success after slot freed, got %v", result.Err)
	}
}

func TestWithBulkhead_ConcurrentCounter(t *testing.T) {
	cfg := health.BulkheadConfig{MaxConcurrent: 10}
	var peak atomic.Int64
	var current atomic.Int64

	fn := func(_ context.Context, _ string) health.StatusResult {
		v := current.Add(1)
		if v > peak.Load() {
			peak.Store(v)
		}
		time.Sleep(2 * time.Millisecond)
		current.Add(-1)
		return health.StatusResult{Status: health.StatusServing}
	}
	wrapped := health.WithBulkhead(fn, cfg)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wrapped(context.Background(), "svc")
		}()
	}
	wg.Wait()

	if peak.Load() > cfg.MaxConcurrent {
		t.Fatalf("peak concurrency %d exceeded limit %d", peak.Load(), cfg.MaxConcurrent)
	}
}

func TestDefaultBulkheadConfig_HasPositiveLimit(t *testing.T) {
	cfg := health.DefaultBulkheadConfig()
	if cfg.MaxConcurrent <= 0 {
		t.Fatalf("expected positive MaxConcurrent, got %d", cfg.MaxConcurrent)
	}
}
