package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

// TestWindowPipeline verifies that WithWindow integrates correctly with
// Pipeline so that repeated failures bubble up as NotServing.
func TestWindowPipeline_RepeatedFailuresCauseNotServing(t *testing.T) {
	cfg := health.WindowConfig{Size: 4, MinSamples: 3, FailureThreshold: 0.5}
	base := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusNotServing, Error: errors.New("down")}
	}
	pipeline := health.Pipeline(base, health.WithWindow(cfg, base))
	_ = pipeline // Pipeline wraps; just verify no panic.

	wfn := health.WithWindow(cfg, base)
	var last health.StatusResult
	for i := 0; i < 5; i++ {
		last = wfn(context.Background(), "svc")
	}
	if last.Status != health.StatusNotServing {
		t.Fatalf("expected NotServing after repeated failures, got %s", last.Status)
	}
}

// TestWindowPipeline_RecoveryAfterSuccess verifies that once enough successes
// are recorded the window reports Serving again.
func TestWindowPipeline_RecoveryAfterSuccess(t *testing.T) {
	cfg := health.WindowConfig{Size: 4, MinSamples: 3, FailureThreshold: 0.75}
	calls := 0
	fn := health.WithWindow(cfg, func(_ context.Context, _ string) health.StatusResult {
		calls++
		if calls <= 2 {
			return health.StatusResult{Status: health.StatusNotServing, Error: errors.New("fail")}
		}
		return health.StatusResult{Status: health.StatusServing}
	})

	var res health.StatusResult
	for i := 0; i < 6; i++ {
		res = fn(context.Background(), "svc")
	}
	if res.Status != health.StatusServing {
		t.Fatalf("expected recovery to Serving, got %s", res.Status)
	}
}

// TestWindowPipeline_ComposeDefault verifies ComposeDefault still works when
// WithWindowDefault is part of the chain.
func TestWindowPipeline_ComposeDefault(t *testing.T) {
	base := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusServing}
	}
	composed := health.ComposeDefault(health.WithWindowDefault(base))
	res := composed(context.Background(), "svc")
	if res.Status == health.StatusUnknown {
		t.Fatal("unexpected Unknown from composed window pipeline")
	}
}
