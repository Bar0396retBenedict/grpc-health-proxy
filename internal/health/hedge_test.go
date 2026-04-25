package health

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithHedge_FastPrimaryNoHedge(t *testing.T) {
	var calls int64
	fn := func(_ context.Context, _ string) StatusResult {
		atomic.AddInt64(&calls, 1)
		return StatusResult{Status: StatusServing}
	}
	cfg := HedgeConfig{Delay: 100 * time.Millisecond, MaxHedges: 1}
	result := WithHedge(cfg, fn)(context.Background(), "svc")
	if result.Status != StatusServing {
		t.Fatalf("expected Serving, got %v", result.Status)
	}
	if atomic.LoadInt64(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", atomic.LoadInt64(&calls))
	}
}

func TestWithHedge_SlowPrimaryTriggersHedge(t *testing.T) {
	var calls int64
	fn := func(ctx context.Context, _ string) StatusResult {
		n := atomic.AddInt64(&calls, 1)
		if n == 1 {
			// first call blocks until context is cancelled
			<-ctx.Done()
			return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
		}
		return StatusResult{Status: StatusServing}
	}
	cfg := HedgeConfig{Delay: 10 * time.Millisecond, MaxHedges: 1}
	result := WithHedge(cfg, fn)(context.Background(), "svc")
	if result.Status != StatusServing {
		t.Fatalf("expected Serving from hedge, got %v", result.Status)
	}
}

func TestWithHedge_AllFail_ReturnsLastError(t *testing.T) {
	sentinel := errors.New("boom")
	fn := func(ctx context.Context, _ string) StatusResult {
		select {
		case <-time.After(5 * time.Millisecond):
			return StatusResult{Status: StatusUnknown, Err: sentinel}
		case <-ctx.Done():
			return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
		}
	}
	cfg := HedgeConfig{Delay: 2 * time.Millisecond, MaxHedges: 1}
	result := WithHedge(cfg, fn)(context.Background(), "svc")
	if result.Err == nil {
		t.Fatal("expected error")
	}
}

func TestWithHedge_ZeroDelayDisabled(t *testing.T) {
	var calls int64
	fn := func(_ context.Context, _ string) StatusResult {
		atomic.AddInt64(&calls, 1)
		return StatusResult{Status: StatusServing}
	}
	cfg := HedgeConfig{Delay: 0, MaxHedges: 1}
	WithHedge(cfg, fn)(context.Background(), "svc")
	if atomic.LoadInt64(&calls) != 1 {
		t.Fatalf("expected exactly 1 call when delay=0, got %d", atomic.LoadInt64(&calls))
	}
}

func TestWithHedge_ContextCancelled(t *testing.T) {
	fn := func(ctx context.Context, _ string) StatusResult {
		<-ctx.Done()
		return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
	}
	cfg := DefaultHedgeConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	result := WithHedge(cfg, fn)(ctx, "svc")
	if result.Err == nil {
		t.Fatal("expected error on context cancellation")
	}
}

func TestDefaultHedgeConfig_HasPositiveValues(t *testing.T) {
	cfg := DefaultHedgeConfig()
	if cfg.Delay <= 0 {
		t.Errorf("expected positive Delay, got %v", cfg.Delay)
	}
	if cfg.MaxHedges <= 0 {
		t.Errorf("expected positive MaxHedges, got %d", cfg.MaxHedges)
	}
}
