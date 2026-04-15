package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithTimeout_CompletesBeforeDeadline(t *testing.T) {
	inner := func(ctx context.Context, svc string) (bool, error) {
		return true, nil
	}

	wrapped := WithTimeout(DefaultTimeoutConfig(), inner)
	ok, err := wrapped(context.Background(), "svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
}

func TestWithTimeout_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := func(ctx context.Context, svc string) (bool, error) {
		return false, sentinel
	}

	wrapped := WithTimeout(DefaultTimeoutConfig(), inner)
	ok, err := wrapped(context.Background(), "svc")
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false")
	}
}

func TestWithTimeout_ExceedsDeadline(t *testing.T) {
	cfg := TimeoutConfig{Timeout: 20 * time.Millisecond}

	inner := func(ctx context.Context, svc string) (bool, error) {
		select {
		case <-time.After(500 * time.Millisecond):
			return true, nil
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}

	wrapped := WithTimeout(cfg, inner)
	start := time.Now()
	ok, err := wrapped(context.Background(), "slow-svc")
	elapsed := time.Since(start)

	if ok {
		t.Fatal("expected ok=false on timeout")
	}
	if err == nil {
		t.Fatal("expected non-nil error on timeout")
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("wrapped call took too long: %v", elapsed)
	}
}

func TestWithTimeout_RespectsParentCancellation(t *testing.T) {
	inner := func(ctx context.Context, svc string) (bool, error) {
		select {
		case <-time.After(500 * time.Millisecond):
			return true, nil
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}

	wrapped := WithTimeout(DefaultTimeoutConfig(), inner)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	ok, err := wrapped(ctx, "svc")
	if ok {
		t.Fatal("expected ok=false when parent cancelled")
	}
	if err == nil {
		t.Fatal("expected error when parent cancelled")
	}
}

func TestDefaultTimeoutConfig_HasPositiveTimeout(t *testing.T) {
	cfg := DefaultTimeoutConfig()
	if cfg.Timeout <= 0 {
		t.Fatalf("expected positive timeout, got %v", cfg.Timeout)
	}
}
