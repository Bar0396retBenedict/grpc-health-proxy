package health_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

func TestWithMirror_BothServing_Agreement(t *testing.T) {
	serving := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusServing}
	}

	check := health.WithMirror(serving, serving)
	result := check(context.Background(), "svc")

	if !result.Agreement {
		t.Error("expected agreement when both are serving")
	}
}

func TestWithMirror_Disagreement(t *testing.T) {
	primary := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusServing}
	}
	mirrorFn := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusNotServing}
	}

	check := health.WithMirror(primary, mirrorFn)
	result := check(context.Background(), "svc")

	if result.Agreement {
		t.Error("expected disagreement between serving and not-serving")
	}
	if result.Primary.Status != health.StatusServing {
		t.Errorf("expected primary Serving, got %s", result.Primary.Status)
	}
	if result.Mirror.Status != health.StatusNotServing {
		t.Errorf("expected mirror NotServing, got %s", result.Mirror.Status)
	}
}

func TestWithMirror_ErrorBreaksAgreement(t *testing.T) {
	serving := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Status: health.StatusServing}
	}
	errFn := func(_ context.Context, _ string) health.StatusResult {
		return health.StatusResult{Err: errors.New("mirror error")}
	}

	check := health.WithMirror(serving, errFn)
	result := check(context.Background(), "svc")

	if result.Agreement {
		t.Error("expected no agreement when mirror returns error")
	}
	if result.Mirror.Err == nil {
		t.Error("expected mirror error to be captured")
	}
}

func TestWithMirror_RunsConcurrently(t *testing.T) {
	gate := make(chan struct{})
	blocking := func(_ context.Context, _ string) health.StatusResult {
		<-gate
		return health.StatusResult{Status: health.StatusServing}
	}

	check := health.WithMirror(blocking, blocking)

	go func() { close(gate) }()
	result := check(context.Background(), "svc")

	if result.Primary.Status != health.StatusServing {
		t.Error("expected primary to complete")
	}
}
