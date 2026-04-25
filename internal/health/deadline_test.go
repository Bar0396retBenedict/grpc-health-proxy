package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithDeadline_CompletesBeforeDeadline(t *testing.T) {
	cfg := DeadlineConfig{Deadline: 500 * time.Millisecond, Label: "test"}
	fn := func(ctx context.Context, svc string) StatusResult {
		return StatusResult{Status: StatusServing}
	}
	result := WithDeadline(cfg, fn)(context.Background(), "svc")
	if result.Status != StatusServing {
		t.Errorf("expected Serving, got %v", result.Status)
	}
	if result.Err != nil {
		t.Errorf("unexpected error: %v", result.Err)
	}
}

func TestWithDeadline_ExceedsDeadline(t *testing.T) {
	cfg := DeadlineConfig{Deadline: 20 * time.Millisecond, Label: "test"}
	fn := func(ctx context.Context, svc string) StatusResult {
		select {
		case <-time.After(200 * time.Millisecond):
			return StatusResult{Status: StatusServing}
		case <-ctx.Done():
			return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
		}
	}
	result := WithDeadline(cfg, fn)(context.Background(), "slow-svc")
	if result.Status != StatusUnknown {
		t.Errorf("expected Unknown, got %v", result.Status)
	}
	if result.Err == nil {
		t.Error("expected deadline error, got nil")
	}
}

func TestWithDeadline_PropagatesInnerError(t *testing.T) {
	expected := errors.New("inner failure")
	cfg := DefaultDeadlineConfig()
	fn := func(ctx context.Context, svc string) StatusResult {
		return StatusResult{Status: StatusUnknown, Err: expected}
	}
	result := WithDeadline(cfg, fn)(context.Background(), "svc")
	if !errors.Is(result.Err, expected) {
		t.Errorf("expected inner error, got %v", result.Err)
	}
}

func TestWithDeadline_ZeroDeadlinePassesThrough(t *testing.T) {
	cfg := DeadlineConfig{Deadline: 0, Label: "noop"}
	fn := func(ctx context.Context, svc string) StatusResult {
		return StatusResult{Status: StatusServing}
	}
	result := WithDeadline(cfg, fn)(context.Background(), "svc")
	if result.Status != StatusServing {
		t.Errorf("expected Serving, got %v", result.Status)
	}
}

func TestWithDeadline_RespectsParentCancellation(t *testing.T) {
	cfg := DeadlineConfig{Deadline: 500 * time.Millisecond, Label: "test"}
	parentCtx, cancel := context.WithCancel(context.Background())

	fn := func(ctx context.Context, svc string) StatusResult {
		select {
		case <-time.After(300 * time.Millisecond):
			return StatusResult{Status: StatusServing}
		case <-ctx.Done():
			return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
		}
	}

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	result := WithDeadline(cfg, fn)(parentCtx, "svc")
	if result.Status != StatusUnknown {
		t.Errorf("expected Unknown after parent cancel, got %v", result.Status)
	}
}

func TestDefaultDeadlineConfig_HasPositiveDeadline(t *testing.T) {
	cfg := DefaultDeadlineConfig()
	if cfg.Deadline <= 0 {
		t.Errorf("expected positive deadline, got %v", cfg.Deadline)
	}
	if cfg.Label == "" {
		t.Error("expected non-empty label")
	}
}
