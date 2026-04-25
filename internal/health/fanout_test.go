package health

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithFanout_FirstServingWins(t *testing.T) {
	targets := []string{"a", "b", "c"}
	calls := int32(0)

	fn := func(ctx context.Context, target string) StatusResult {
		atomic.AddInt32(&calls, 1)
		if target == "b" {
			return StatusResult{Status: StatusServing}
		}
		// simulate slow / unhealthy targets
		select {
		case <-ctx.Done():
			return StatusResult{Status: StatusUnknown}
		case <-time.After(500 * time.Millisecond):
			return StatusResult{Status: StatusNotServing}
		}
	}

	check := WithFanout(targets, fn, DefaultFanoutConfig())
	result := check(context.Background())

	if result.Status != StatusServing {
		t.Fatalf("expected SERVING, got %s", result.Status)
	}
}

func TestWithFanout_AllUnhealthy_ReturnsLastResult(t *testing.T) {
	targets := []string{"x", "y"}

	fn := func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusNotServing, Err: errors.New("down")}
	}

	check := WithFanout(targets, fn, DefaultFanoutConfig())
	result := check(context.Background())

	if result.Status != StatusNotServing {
		t.Fatalf("expected NOT_SERVING, got %s", result.Status)
	}
	if result.Err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestWithFanout_EmptyTargets_ReturnsUnknown(t *testing.T) {
	check := WithFanout(nil, func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	}, DefaultFanoutConfig())

	result := check(context.Background())
	if result.Status != StatusUnknown {
		t.Fatalf("expected UNKNOWN for empty targets, got %s", result.Status)
	}
}

func TestWithFanout_ContextCancelled(t *testing.T) {
	targets := []string{"slow"}

	fn := func(ctx context.Context, _ string) StatusResult {
		select {
		case <-ctx.Done():
			return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
		case <-time.After(5 * time.Second):
			return StatusResult{Status: StatusServing}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	check := WithFanout(targets, fn, DefaultFanoutConfig())
	result := check(ctx)

	if result.Status == StatusServing {
		t.Fatal("expected non-serving result when context cancelled")
	}
}

func TestDefaultFanoutConfig_HasLabel(t *testing.T) {
	cfg := DefaultFanoutConfig()
	if cfg.Label == "" {
		t.Fatal("expected non-empty label in default config")
	}
}
