package health

import (
	"context"
	"math/rand"
	"sync"
)

// SamplerConfig controls probabilistic sampling of health check calls.
type SamplerConfig struct {
	// SampleRate is the fraction of calls to pass through [0.0, 1.0].
	// A value of 1.0 means every call passes; 0.0 means none pass.
	SampleRate float64
	// Fallback is the status returned for skipped (non-sampled) calls.
	Fallback StatusResult
}

// DefaultSamplerConfig returns a SamplerConfig that passes every call.
func DefaultSamplerConfig() SamplerConfig {
	return SamplerConfig{
		SampleRate: 1.0,
		Fallback:   StatusResult{Status: StatusUnknown},
	}
}

type sampler struct {
	mu     sync.Mutex
	rng    *rand.Rand
	cfg    SamplerConfig
	inner  HealthCheckFn
}

// WithSampler wraps a HealthCheckFn so that only a random fraction of calls
// are forwarded to the inner function. Skipped calls return cfg.Fallback.
func WithSampler(cfg SamplerConfig, inner HealthCheckFn) HealthCheckFn {
	s := &sampler{
		rng:   rand.New(rand.NewSource(rand.Int63())), //nolint:gosec
		cfg:   cfg,
		inner: inner,
	}
	return s.check
}

func (s *sampler) check(ctx context.Context, service string) StatusResult {
	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()

	if v >= s.cfg.SampleRate {
		return s.cfg.Fallback
	}
	return s.inner(ctx, service)
}
