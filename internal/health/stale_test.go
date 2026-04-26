package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errUpstream = errors.New("upstream error")

func TestWithStale_SuccessPassesThrough(t *testing.T) {
	cfg := DefaultStaleConfig()
	calls := 0
	fn := WithStale(cfg, func(_ context.Context, _ string) StatusResult {
		calls++
		return StatusResult{Status: StatusServing}
	})

	res := fn(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("expected Serving, got %v", res.Status)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestWithStale_ReturnsCachedOnError(t *testing.T) {
	cfg := DefaultStaleConfig()
	serving := true
	fn := WithStale(cfg, func(_ context.Context, _ string) StatusResult {
		if serving {
			return StatusResult{Status: StatusServing}
		}
		return StatusResult{Status: StatusUnknown, Error: errUpstream}
	})

	// Warm the cache.
	_ = fn(context.Background(), "svc")

	// Simulate upstream failure.
	serving = false
	res := fn(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("expected cached Serving, got %v", res.Status)
	}
	if res.Error != nil {
		t.Fatalf("expected no error from stale cache, got %v", res.Error)
	}
}

func TestWithStale_FallbackWhenCacheStale(t *testing.T) {
	cfg := StaleConfig{
		MaxAge:   1 * time.Millisecond,
		Fallback: StatusUnknown,
	}
	serving := true
	fn := WithStale(cfg, func(_ context.Context, _ string) StatusResult {
		if serving {
			return StatusResult{Status: StatusServing}
		}
		return StatusResult{Status: StatusUnknown, Error: errUpstream}
	})

	// Warm the cache.
	_ = fn(context.Background(), "svc")

	// Let the cache expire.
	time.Sleep(5 * time.Millisecond)

	serving = false
	res := fn(context.Background(), "svc")
	if res.Status != StatusUnknown {
		t.Fatalf("expected Unknown fallback, got %v", res.Status)
	}
	if res.Error == nil {
		t.Fatal("expected upstream error to be propagated")
	}
}

func TestWithStale_FallbackWhenNoCacheAndError(t *testing.T) {
	cfg := DefaultStaleConfig()
	fn := WithStale(cfg, func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusUnknown, Error: errUpstream}
	})

	res := fn(context.Background(), "svc")
	if res.Status != StatusUnknown {
		t.Fatalf("expected Unknown fallback, got %v", res.Status)
	}
	if !errors.Is(res.Error, errUpstream) {
		t.Fatalf("expected errUpstream, got %v", res.Error)
	}
}

func TestDefaultStaleConfig_HasPositiveMaxAge(t *testing.T) {
	cfg := DefaultStaleConfig()
	if cfg.MaxAge <= 0 {
		t.Fatalf("expected positive MaxAge, got %v", cfg.MaxAge)
	}
}
