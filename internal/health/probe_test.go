package health

import (
	"context"
	"errors"
	"testing"
)

func alwaysServing(_ context.Context, _ string) StatusResult {
	return StatusResult{Status: StatusServing}
}

func alwaysNotServing(_ context.Context, _ string) StatusResult {
	return StatusResult{Status: StatusNotServing}
}

func alwaysError(_ context.Context, _ string) StatusResult {
	return StatusResult{Status: StatusUnknown, Err: errors.New("dial error")}
}

func TestProbe_BecomesReadyAfterSuccessThreshold(t *testing.T) {
	p := NewProbe(ProbeConfig{SuccessThreshold: 2, FailureThreshold: 1}, alwaysServing)
	ctx := context.Background()

	if err := p.Check(ctx, "svc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Ready() {
		t.Fatal("should not be ready after only 1 success (threshold=2)")
	}
	if err := p.Check(ctx, "svc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.Ready() {
		t.Fatal("should be ready after 2 consecutive successes")
	}
}

func TestProbe_BecomesNotReadyAfterFailureThreshold(t *testing.T) {
	p := NewProbe(ProbeConfig{SuccessThreshold: 1, FailureThreshold: 2}, alwaysServing)
	ctx := context.Background()

	// Become ready first.
	_ = p.Check(ctx, "svc")
	if !p.Ready() {
		t.Fatal("expected ready")
	}

	// Switch to failing check.
	p.fn = alwaysNotServing
	_ = p.Check(ctx, "svc") // fail 1 — still ready
	if !p.Ready() {
		t.Fatal("should still be ready after 1 failure (threshold=2)")
	}
	_ = p.Check(ctx, "svc") // fail 2 — not ready
	if p.Ready() {
		t.Fatal("should not be ready after 2 consecutive failures")
	}
}

func TestProbe_CheckReturnsErrorWhenNotServing(t *testing.T) {
	p := NewProbe(DefaultProbeConfig(), alwaysNotServing)
	err := p.Check(context.Background(), "my-service")
	if err == nil {
		t.Fatal("expected error for not-serving status")
	}
}

func TestProbe_CheckReturnsUnderlyingError(t *testing.T) {
	p := NewProbe(DefaultProbeConfig(), alwaysError)
	err := p.Check(context.Background(), "my-service")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "dial error" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestProbe_DefaultConfig(t *testing.T) {
	cfg := DefaultProbeConfig()
	if cfg.SuccessThreshold != 1 {
		t.Errorf("expected SuccessThreshold=1, got %d", cfg.SuccessThreshold)
	}
	if cfg.FailureThreshold != 3 {
		t.Errorf("expected FailureThreshold=3, got %d", cfg.FailureThreshold)
	}
}

func TestProbe_ResetsSuccessCountOnFailure(t *testing.T) {
	p := NewProbe(ProbeConfig{SuccessThreshold: 3, FailureThreshold: 1}, alwaysServing)
	ctx := context.Background()

	_ = p.Check(ctx, "svc") // ok 1
	_ = p.Check(ctx, "svc") // ok 2
	p.fn = alwaysNotServing
	_ = p.Check(ctx, "svc") // fail — resets ok count
	p.fn = alwaysServing
	_ = p.Check(ctx, "svc") // ok 1 again
	if p.Ready() {
		t.Fatal("should not be ready: success count was reset")
	}
}
