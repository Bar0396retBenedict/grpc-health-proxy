package health

import (
	"context"
	"errors"
	"testing"
)

func TestWithWindow_BelowMinSamples_PassesThrough(t *testing.T) {
	cfg := WindowConfig{Size: 10, MinSamples: 3, FailureThreshold: 0.5}
	calls := 0
	fn := WithWindow(cfg, func(_ context.Context, _ string) StatusResult {
		calls++
		if calls == 1 {
			return StatusResult{Status: StatusNotServing}
		}
		return StatusResult{Status: StatusServing}
	})
	// First call: below MinSamples, passes primary result through.
	res := fn(context.Background(), "svc")
	if res.Status != StatusNotServing {
		t.Fatalf("expected NotServing got %s", res.Status)
	}
}

func TestWithWindow_AllServing_ReturnsServing(t *testing.T) {
	cfg := WindowConfig{Size: 5, MinSamples: 3, FailureThreshold: 0.5}
	fn := WithWindow(cfg, func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	})
	var res StatusResult
	for i := 0; i < 5; i++ {
		res = fn(context.Background(), "svc")
	}
	if res.Status != StatusServing {
		t.Fatalf("expected Serving got %s", res.Status)
	}
}

func TestWithWindow_MajorityFailures_ReturnsNotServing(t *testing.T) {
	cfg := WindowConfig{Size: 6, MinSamples: 3, FailureThreshold: 0.5}
	calls := 0
	fn := WithWindow(cfg, func(_ context.Context, _ string) StatusResult {
		calls++
		if calls%2 == 0 {
			return StatusResult{Status: StatusNotServing, Error: errors.New("fail")}
		}
		return StatusResult{Status: StatusServing}
	})
	var res StatusResult
	for i := 0; i < 6; i++ {
		res = fn(context.Background(), "svc")
	}
	if res.Status != StatusNotServing {
		t.Fatalf("expected NotServing got %s", res.Status)
	}
}

func TestWithWindow_ErrorCountsAsFailure(t *testing.T) {
	cfg := WindowConfig{Size: 4, MinSamples: 3, FailureThreshold: 0.5}
	calls := 0
	fn := WithWindow(cfg, func(_ context.Context, _ string) StatusResult {
		calls++
		if calls <= 3 {
			return StatusResult{Status: StatusServing, Error: errors.New("err")}
		}
		return StatusResult{Status: StatusServing}
	})
	var res StatusResult
	for i := 0; i < 4; i++ {
		res = fn(context.Background(), "svc")
	}
	if res.Status != StatusNotServing {
		t.Fatalf("expected NotServing due to errors, got %s", res.Status)
	}
}

func TestDefaultWindowConfig_HasPositiveValues(t *testing.T) {
	cfg := DefaultWindowConfig()
	if cfg.Size <= 0 {
		t.Errorf("Size should be positive, got %d", cfg.Size)
	}
	if cfg.MinSamples <= 0 {
		t.Errorf("MinSamples should be positive, got %d", cfg.MinSamples)
	}
	if cfg.FailureThreshold <= 0 || cfg.FailureThreshold > 1 {
		t.Errorf("FailureThreshold out of range: %f", cfg.FailureThreshold)
	}
}

func TestWithWindowDefault_Smoke(t *testing.T) {
	fn := WithWindowDefault(func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	})
	res := fn(context.Background(), "svc")
	if res.Status == StatusUnknown {
		t.Fatal("unexpected Unknown status")
	}
}
