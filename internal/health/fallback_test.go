package health

import (
	"context"
	"errors"
	"testing"
)

func primaryServing(_ context.Context, _ string) StatusResult {
	return StatusResult{Status: StatusServing}
}

func primaryError(_ context.Context, _ string) StatusResult {
	return StatusResult{Status: StatusUnknown, Error: errors.New("primary down")}
}

func primaryUnknown(_ context.Context, _ string) StatusResult {
	return StatusResult{Status: StatusUnknown}
}

func fallbackServing(_ context.Context, _ string) StatusResult {
	return StatusResult{Status: StatusServing}
}

func TestWithFallback_PrimarySucceeds(t *testing.T) {
	cfg := DefaultFallbackConfig()
	fn := WithFallback(cfg, primaryServing, primaryError)
	res := fn(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("expected Serving, got %v", res.Status)
	}
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
}

func TestWithFallback_PrimaryErrorUsesFallback(t *testing.T) {
	cfg := DefaultFallbackConfig()
	fn := WithFallback(cfg, primaryError, fallbackServing)
	res := fn(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("expected fallback Serving, got %v", res.Status)
	}
}

func TestWithFallback_UnknownNotFallingBackByDefault(t *testing.T) {
	cfg := DefaultFallbackConfig()
	fn := WithFallback(cfg, primaryUnknown, fallbackServing)
	res := fn(context.Background(), "svc")
	if res.Status != StatusUnknown {
		t.Fatalf("expected Unknown without fallback, got %v", res.Status)
	}
}

func TestWithFallback_UnknownFallsBackWhenEnabled(t *testing.T) {
	cfg := DefaultFallbackConfig()
	cfg.FallbackOnUnknown = true
	fn := WithFallback(cfg, primaryUnknown, fallbackServing)
	res := fn(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("expected fallback Serving, got %v", res.Status)
	}
}

func TestDefaultFallbackConfig_HasLabel(t *testing.T) {
	cfg := DefaultFallbackConfig()
	if cfg.Label == "" {
		t.Fatal("expected non-empty Label")
	}
	if cfg.FallbackOnUnknown {
		t.Fatal("FallbackOnUnknown should default to false")
	}
}
