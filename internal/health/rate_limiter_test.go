package health

import (
	"context"
	"testing"
	"time"
)

func TestWithRateLimit_AllowsUnderLimit(t *testing.T) {
	cfg := RateLimiterConfig{MaxCalls: 3, Interval: time.Second}
	calls := 0
	fn := WithRateLimit(cfg, func(_ context.Context, _ string) StatusResult {
		calls++
		return StatusResult{Status: StatusServing}
	})

	for i := 0; i < 3; i++ {
		res := fn(context.Background(), "svc")
		if res.Status != StatusServing {
			t.Fatalf("expected serving, got %v", res.Status)
		}
	}
	if calls != 3 {
		t.Fatalf("expected 3 inner calls, got %d", calls)
	}
}

func TestWithRateLimit_BlocksOverLimit(t *testing.T) {
	cfg := RateLimiterConfig{MaxCalls: 2, Interval: time.Second}
	fn := WithRateLimit(cfg, func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	})

	fn(context.Background(), "svc")
	fn(context.Background(), "svc")
	res := fn(context.Background(), "svc") // third call should be blocked

	if res.Status != StatusUnknown {
		t.Fatalf("expected unknown (rate limited), got %v", res.Status)
	}
	if res.Err != ErrRateLimited {
		t.Fatalf("expected ErrRateLimited, got %v", res.Err)
	}
}

func TestWithRateLimit_AllowsAfterWindowExpires(t *testing.T) {
	cfg := RateLimiterConfig{MaxCalls: 1, Interval: 50 * time.Millisecond}
	fn := WithRateLimit(cfg, func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	})

	res := fn(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("first call: expected serving, got %v", res.Status)
	}

	res = fn(context.Background(), "svc")
	if res.Err != ErrRateLimited {
		t.Fatal("second call should be rate limited")
	}

	time.Sleep(60 * time.Millisecond)

	res = fn(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("after window: expected serving, got %v", res.Status)
	}
}

func TestDefaultRateLimiterConfig_HasPositiveValues(t *testing.T) {
	cfg := DefaultRateLimiterConfig()
	if cfg.MaxCalls <= 0 {
		t.Errorf("expected positive MaxCalls, got %d", cfg.MaxCalls)
	}
	if cfg.Interval <= 0 {
		t.Errorf("expected positive Interval, got %v", cfg.Interval)
	}
}

func TestWithRateLimit_IndependentInstances(t *testing.T) {
	cfg := RateLimiterConfig{MaxCalls: 1, Interval: time.Second}

	fn1 := WithRateLimit(cfg, func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	})
	fn2 := WithRateLimit(cfg, func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	})

	fn1(context.Background(), "svc")
	res := fn2(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatal("fn2 should not be affected by fn1's rate limit state")
	}
}
