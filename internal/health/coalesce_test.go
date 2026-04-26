package health

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithCoalesce_ZeroWindowPassesThrough(t *testing.T) {
	calls := 0
	fn := func(_ context.Context, _ string) StatusResult {
		calls++
		return StatusResult{Status: StatusServing}
	}
	coalesced := WithCoalesce(CoalesceConfig{Window: 0}, fn)
	res := coalesced(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("expected Serving, got %v", res.Status)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestWithCoalesce_ConcurrentCallsCoalesced(t *testing.T) {
	var callCount atomic.Int32

	fn := func(_ context.Context, _ string) StatusResult {
		time.Sleep(10 * time.Millisecond)
		callCount.Add(1)
		return StatusResult{Status: StatusServing}
	}

	cfg := CoalesceConfig{Window: 50 * time.Millisecond}
	coalesced := WithCoalesce(cfg, fn)

	const goroutines = 10
	var wg sync.WaitGroup
	results := make([]StatusResult, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = coalesced(context.Background(), "svc")
		}(i)
	}
	wg.Wait()

	for i, r := range results {
		if r.Status != StatusServing {
			t.Errorf("goroutine %d: expected Serving, got %v", i, r.Status)
		}
	}

	if n := callCount.Load(); n > 3 {
		t.Errorf("expected coalesced calls (<= 3), got %d", n)
	}
}

func TestWithCoalesce_ContextCancelledDuringWindow(t *testing.T) {
	fn := func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	}

	cfg := CoalesceConfig{Window: 500 * time.Millisecond}
	coalesced := WithCoalesce(cfg, fn)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	res := coalesced(ctx, "svc")
	if res.Err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestDefaultCoalesceConfig_HasPositiveWindow(t *testing.T) {
	cfg := DefaultCoalesceConfig()
	if cfg.Window <= 0 {
		t.Fatalf("expected positive window, got %v", cfg.Window)
	}
}

func TestWithCoalesce_SequentialCallsEachExecute(t *testing.T) {
	var callCount atomic.Int32

	fn := func(_ context.Context, _ string) StatusResult {
		callCount.Add(1)
		return StatusResult{Status: StatusServing}
	}

	cfg := CoalesceConfig{Window: 10 * time.Millisecond}
	coalesced := WithCoalesce(cfg, fn)

	for i := 0; i < 3; i++ {
		res := coalesced(context.Background(), "svc")
		if res.Status != StatusServing {
			t.Fatalf("call %d: expected Serving", i)
		}
		time.Sleep(20 * time.Millisecond)
	}

	if n := callCount.Load(); n != 3 {
		t.Errorf("expected 3 sequential calls, got %d", n)
	}
}
