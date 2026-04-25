package health

import (
	"context"
	"testing"
)

func TestPipeline_NoMiddleware(t *testing.T) {
	base := func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	}
	fn := Pipeline(base)
	res := fn(context.Background(), "svc")
	if res.Status != StatusServing {
		t.Fatalf("expected Serving, got %v", res.Status)
	}
}

func TestPipeline_MiddlewareOrder(t *testing.T) {
	var order []string

	makeMiddleware := func(label string) Middleware {
		return func(next HealthcheckFn) HealthcheckFn {
			return func(ctx context.Context, service string) StatusResult {
				order = append(order, label+"-before")
				res := next(ctx, service)
				order = append(order, label+"-after")
				return res
			}
		}
	}

	base := func(_ context.Context, _ string) StatusResult {
		order = append(order, "base")
		return StatusResult{Status: StatusServing}
	}

	fn := Pipeline(base, makeMiddleware("A"), makeMiddleware("B"))
	fn(context.Background(), "svc")

	want := []string{"A-before", "B-before", "base", "B-after", "A-after"}
	if len(order) != len(want) {
		t.Fatalf("order length mismatch: got %v, want %v", order, want)
	}
	for i, v := range want {
		if order[i] != v {
			t.Errorf("order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

func TestPipeline_MiddlewareCanShortCircuit(t *testing.T) {
	called := false
	base := func(_ context.Context, _ string) StatusResult {
		called = true
		return StatusResult{Status: StatusServing}
	}
	blocking := func(_ HealthcheckFn) HealthcheckFn {
		return func(_ context.Context, _ string) StatusResult {
			return StatusResult{Status: StatusNotServing}
		}
	}
	fn := Pipeline(base, blocking)
	res := fn(context.Background(), "svc")
	if called {
		t.Fatal("base should not have been called")
	}
	if res.Status != StatusNotServing {
		t.Fatalf("expected NotServing, got %v", res.Status)
	}
}
