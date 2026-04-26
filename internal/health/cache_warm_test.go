package health

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWarmCache_PopulatesAllServices(t *testing.T) {
	cache := NewCache()
	cfg := WarmupConfig{
		Timeout:  2 * time.Second,
		Services: []string{"svc.A", "svc.B"},
	}

	fn := func(_ context.Context, svc string) StatusResult {
		return StatusResult{Status: StatusServing}
	}

	err := WarmCache(context.Background(), cfg, cache, fn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, svc := range cfg.Services {
		status, _ := cache.Get(svc)
		if status != StatusServing {
			t.Errorf("service %s: want Serving, got %s", svc, status)
		}
	}
}

func TestWarmCache_StoresErrorStatus(t *testing.T) {
	cache := NewCache()
	cfg := WarmupConfig{
		Timeout:  2 * time.Second,
		Services: []string{"svc.Broken"},
	}

	wantErr := errors.New("dial error")
	fn := func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusUnknown, Err: wantErr}
	}

	_ = WarmCache(context.Background(), cfg, cache, fn)

	_, gotErr := cache.Get("svc.Broken")
	if !errors.Is(gotErr, wantErr) {
		t.Errorf("want %v, got %v", wantErr, gotErr)
	}
}

func TestWarmCache_EmptyServicesIsNoop(t *testing.T) {
	cache := NewCache()
	cfg := WarmupConfig{Timeout: time.Second, Services: nil}

	err := WarmCache(context.Background(), cfg, cache, func(_ context.Context, _ string) StatusResult {
		t.Fatal("fn should not be called")
		return StatusResult{}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWarmCache_TimesOut(t *testing.T) {
	cache := NewCache()
	cfg := WarmupConfig{
		Timeout:  50 * time.Millisecond,
		Services: []string{"svc.Slow"},
	}

	var called atomic.Int32
	fn := func(ctx context.Context, _ string) StatusResult {
		called.Add(1)
		select {
		case <-ctx.Done():
		case <-time.After(10 * time.Second):
		}
		return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
	}

	err := WarmCache(context.Background(), cfg, cache, fn)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if called.Load() != 1 {
		t.Errorf("expected fn called once, got %d", called.Load())
	}
}

func TestDefaultWarmupConfig_HasPositiveTimeout(t *testing.T) {
	cfg := DefaultWarmupConfig()
	if cfg.Timeout <= 0 {
		t.Errorf("expected positive timeout, got %v", cfg.Timeout)
	}
}
