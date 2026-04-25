package health_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

func TestWithShadow_ReturnsPrimaryResult(t *testing.T) {
	primary := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusServing}
	}
	shadowFn := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusNotServing}
	}

	wrapped := health.WithShadow(primary, shadowFn, health.DefaultShadowConfig())
	result := wrapped(context.Background(), "svc")

	if result.Status != health.StatusServing {
		t.Errorf("expected Serving, got %s", result.Status)
	}
}

func TestWithShadow_ShadowRunsAsynchronously(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	primary := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusServing}
	}
	shadowFn := func(_ context.Context, _ string) health.StatusResult {
		defer wg.Done()
		return health.StatusResult{Status: health.StatusServing}
	}

	wrapped := health.WithShadow(primary, shadowFn, health.DefaultShadowConfig())
	wrapped(context.Background(), "svc")

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:
		// ok
	case <-time.After(time.Second):
		t.Fatal("shadow goroutine did not complete in time")
	}
}

func TestWithShadow_ShadowErrorDoesNotAffectPrimary(t *testing.T) {
	primary := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusServing}
	}
	shadowFn := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Err: errors.New("shadow boom")}
	}

	wrapped := health.WithShadow(primary, shadowFn, health.DefaultShadowConfig())
	result := wrapped(context.Background(), "svc")

	if result.Err != nil {
		t.Errorf("expected no error from primary, got %v", result.Err)
	}
}

func TestDefaultShadowConfig_HasLabel(t *testing.T) {
	cfg := health.DefaultShadowConfig()
	if cfg.Label == "" {
		t.Error("expected non-empty default label")
	}
}
