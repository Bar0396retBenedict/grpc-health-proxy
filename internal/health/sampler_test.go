package health

import (
	"context"
	"testing"
)

func TestWithSampler_AlwaysPassesAtRate1(t *testing.T) {
	called := 0
	inner := func(_ context.Context, _ string) StatusResult {
		called++
		return StatusResult{Status: StatusServing}
	}

	fn := WithSampler(SamplerConfig{SampleRate: 1.0}, inner)

	for i := 0; i < 20; i++ {
		res := fn(context.Background(), "svc")
		if res.Status != StatusServing {
			t.Fatalf("expected Serving, got %v", res.Status)
		}
	}
	if called != 20 {
		t.Errorf("expected inner called 20 times, got %d", called)
	}
}

func TestWithSampler_NeverPassesAtRate0(t *testing.T) {
	called := 0
	inner := func(_ context.Context, _ string) StatusResult {
		called++
		return StatusResult{Status: StatusServing}
	}

	fallback := StatusResult{Status: StatusUnknown}
	fn := WithSampler(SamplerConfig{SampleRate: 0.0, Fallback: fallback}, inner)

	for i := 0; i < 20; i++ {
		res := fn(context.Background(), "svc")
		if res.Status != StatusUnknown {
			t.Fatalf("expected Unknown fallback, got %v", res.Status)
		}
	}
	if called != 0 {
		t.Errorf("expected inner never called, got %d", called)
	}
}

func TestWithSampler_PartialRate(t *testing.T) {
	called := 0
	inner := func(_ context.Context, _ string) StatusResult {
		called++
		return StatusResult{Status: StatusServing}
	}

	fn := WithSampler(SamplerConfig{SampleRate: 0.5, Fallback: StatusResult{Status: StatusUnknown}}, inner)

	const iterations = 10000
	for i := 0; i < iterations; i++ {
		fn(context.Background(), "svc")
	}

	rate := float64(called) / iterations
	if rate < 0.40 || rate > 0.60 {
		t.Errorf("expected ~50%% pass-through, got %.2f%%", rate*100)
	}
}

func TestDefaultSamplerConfig_HasFullRate(t *testing.T) {
	cfg := DefaultSamplerConfig()
	if cfg.SampleRate != 1.0 {
		t.Errorf("expected SampleRate 1.0, got %f", cfg.SampleRate)
	}
	if cfg.Fallback.Status != StatusUnknown {
		t.Errorf("expected fallback Unknown, got %v", cfg.Fallback.Status)
	}
}

func TestWithSampler_FallbackReturnedWhenSkipped(t *testing.T) {
	inner := func(_ context.Context, _ string) StatusResult {
		return StatusResult{Status: StatusServing}
	}

	fallback := StatusResult{Status: StatusNotServing}
	fn := WithSampler(SamplerConfig{SampleRate: 0.0, Fallback: fallback}, inner)

	res := fn(context.Background(), "svc")
	if res.Status != StatusNotServing {
		t.Errorf("expected NotServing fallback, got %v", res.Status)
	}
}
