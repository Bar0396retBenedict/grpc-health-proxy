package health

import (
	"context"
	"testing"
	"time"
)

func TestWithJitter_ZeroMaxJitterPassesThrough(t *testing.T) {
	called := false
	inner := func(_ context.Context, svc string) StatusResult {
		called = true
		return StatusResult{Status: StatusServing}
	}

	fn := WithJitter(JitterConfig{MaxJitter: 0}, inner)
	result := fn(context.Background(), "svc")

	if !called {
		t.Fatal("expected inner fn to be called")
	}
	if result.Status != StatusServing {
		t.Fatalf("expected Serving, got %s", result.Status)
	}
}

func TestWithJitter_EventuallyCallsInner(t *testing.T) {
	called := false
	inner := func(_ context.Context, svc string) StatusResult {
		called = true
		return StatusResult{Status: StatusServing}
	}

	cfg := JitterConfig{MaxJitter: 20 * time.Millisecond}
	fn := WithJitter(cfg, inner)
	result := fn(context.Background(), "svc")

	if !called {
		t.Fatal("expected inner fn to be called after jitter delay")
	}
	if result.Status != StatusServing {
		t.Fatalf("expected Serving, got %s", result.Status)
	}
}

func TestWithJitter_ContextCancelledDuringDelay(t *testing.T) {
	inner := func(_ context.Context, svc string) StatusResult {
		t.Error("inner should not be called when context is cancelled during jitter")
		return StatusResult{Status: StatusServing}
	}

	cfg := JitterConfig{MaxJitter: 5 * time.Second}
	fn := WithJitter(cfg, inner)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	result := fn(ctx, "svc")

	if result.Status != StatusUnknown {
		t.Fatalf("expected Unknown on cancelled context, got %s", result.Status)
	}
	if result.Err == nil {
		t.Fatal("expected non-nil error on cancelled context")
	}
}

func TestDefaultJitterConfig_HasPositiveMaxJitter(t *testing.T) {
	cfg := DefaultJitterConfig()
	if cfg.MaxJitter <= 0 {
		t.Fatalf("expected positive MaxJitter, got %s", cfg.MaxJitter)
	}
}

func TestWithJitter_NegativeMaxJitterPassesThrough(t *testing.T) {
	called := false
	inner := func(_ context.Context, svc string) StatusResult {
		called = true
		return StatusResult{Status: StatusNotServing}
	}

	fn := WithJitter(JitterConfig{MaxJitter: -1 * time.Millisecond}, inner)
	result := fn(context.Background(), "svc")

	if !called {
		t.Fatal("expected inner fn to be called with negative MaxJitter")
	}
	if result.Status != StatusNotServing {
		t.Fatalf("expected NotServing, got %s", result.Status)
	}
}
