package health

import (
	"context"
	"fmt"
	"time"
)

// ProbeConfig holds configuration for a readiness/liveness probe.
type ProbeConfig struct {
	// InitialDelay is the time to wait before the first probe.
	InitialDelay time.Duration
	// SuccessThreshold is the number of consecutive successes required.
	SuccessThreshold int
	// FailureThreshold is the number of consecutive failures before reporting unhealthy.
	FailureThreshold int
}

// DefaultProbeConfig returns a ProbeConfig with sensible defaults.
func DefaultProbeConfig() ProbeConfig {
	return ProbeConfig{
		InitialDelay:     0,
		SuccessThreshold: 1,
		FailureThreshold: 3,
	}
}

// Probe wraps a HealthCheckFn and tracks consecutive success/failure counts
// to implement Kubernetes-style readiness probe semantics.
type Probe struct {
	cfg            ProbeConfig
	fn             HealthCheckFn
	consecutiveFail int
	consecutiveOK   int
	ready          bool
}

// NewProbe creates a Probe with the given config and check function.
func NewProbe(cfg ProbeConfig, fn HealthCheckFn) *Probe {
	if cfg.SuccessThreshold <= 0 {
		cfg.SuccessThreshold = 1
	}
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 1
	}
	return &Probe{cfg: cfg, fn: fn}
}

// Check runs the underlying health check and updates the probe state.
// It returns an error if the probe considers the target not ready.
func (p *Probe) Check(ctx context.Context, service string) error {
	result := p.fn(ctx, service)
	if result.IsServing() {
		p.consecutiveFail = 0
		p.consecutiveOK++
		if p.consecutiveOK >= p.cfg.SuccessThreshold {
			p.ready = true
		}
		return nil
	}
	p.consecutiveOK = 0
	p.consecutiveFail++
	if p.consecutiveFail >= p.cfg.FailureThreshold {
		p.ready = false
	}
	err := result.Err
	if err == nil {
		err = fmt.Errorf("service %q is %s", service, result.Status)
	}
	return err
}

// Ready reports whether the probe currently considers the target ready.
func (p *Probe) Ready() bool {
	return p.ready
}
