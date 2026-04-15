package health_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

func TestWithThrottle_FirstCallAlwaysPasses(t *testing.T) {
	calls := atomic.Int32{}
	fn := func(_ context.Context, _ string) (bool, error) {
		calls.Add(1)
		return true, nil
	}
	throttled := health.WithThrottle(health.DefaultThrottleConfig(), fn)

	ok, err := throttled(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected healthy")
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
}

func TestWithThrottle_SecondCallBlockedWithinInterval(t *testing.T) {
	calls := atomic.Int32{}
	fn := func(_ context.Context, _ string) (bool, error) {
		calls.Add(1)
		return true, nil
	}
	cfg := health.ThrottleConfig{MinInterval: 200 * time.Millisecond}
	throttled := health.WithThrottle(cfg, fn)

	_, _ = throttled(context.Background(), "svc")
	_, err := throttled(context.Background(), "svc")

	if !errors.Is(err, health.ErrThrottled) {
		t.Fatalf("expected ErrThrottled, got %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 underlying call, got %d", calls.Load())
	}
}

func TestWithThrottle_CallAllowedAfterInterval(t *testing.T) {
	calls := atomic.Int32{}
	fn := func(_ context.Context, _ string) (bool, error) {
		calls.Add(1)
		return true, nil
	}
	cfg := health.ThrottleConfig{MinInterval: 50 * time.Millisecond}
	throttled := health.WithThrottle(cfg, fn)

	_, _ = throttled(context.Background(), "svc")
	time.Sleep(60 * time.Millisecond)
	_, err := throttled(context.Background(), "svc")

	if err != nil {
		t.Fatalf("expected no error after interval, got %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", calls.Load())
	}
}

func TestWithThrottle_IndependentInstances(t *testing.T) {
	fn := func(_ context.Context, _ string) (bool, error) { return true, nil }
	cfg := health.ThrottleConfig{MinInterval: 200 * time.Millisecond}

	a := health.WithThrottle(cfg, fn)
	b := health.WithThrottle(cfg, fn)

	_, _ = a(context.Background(), "svc")
	_, err := b(context.Background(), "svc")
	if err != nil {
		t.Fatalf("separate throttle instances should not share state: %v", err)
	}
}

func TestWithThrottle_DefaultConfig(t *testing.T) {
	cfg := health.DefaultThrottleConfig()
	if cfg.MinInterval <= 0 {
		t.Fatal("default MinInterval must be positive")
	}
}
