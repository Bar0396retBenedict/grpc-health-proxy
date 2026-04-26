package health

import (
	"context"
	"sync"
	"time"
)

// WindowConfig controls the sliding-window success-rate checker.
type WindowConfig struct {
	// Size is the number of most-recent results to keep.
	Size int
	// MinSamples is the minimum number of results required before a verdict
	// is returned; below this threshold the primary result is passed through.
	MinSamples int
	// FailureThreshold is the fraction [0,1] of failures that causes the
	// window to report NotServing.
	FailureThreshold float64
}

// DefaultWindowConfig returns sensible defaults.
func DefaultWindowConfig() WindowConfig {
	return WindowConfig{
		Size:             10,
		MinSamples:       3,
		FailureThreshold: 0.5,
	}
}

type windowState struct {
	mu      sync.Mutex
	buf     []bool // true == success
	head    int
	count   int
	cfg     WindowConfig
}

func newWindowState(cfg WindowConfig) *windowState {
	return &windowState{buf: make([]bool, cfg.Size), cfg: cfg}
}

func (w *windowState) record(success bool) StatusResult {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf[w.head%w.cfg.Size] = success
	w.head++
	if w.count < w.cfg.Size {
		w.count++
	}
	if w.count < w.cfg.MinSamples {
		if success {
			return StatusResult{Status: StatusServing}
		}
		return StatusResult{Status: StatusNotServing}
	}
	var failures int
	for i := 0; i < w.count; i++ {
		if !w.buf[i] {
			failures++
		}
	}
	rate := float64(failures) / float64(w.count)
	if rate >= w.cfg.FailureThreshold {
		return StatusResult{Status: StatusNotServing}
	}
	return StatusResult{Status: StatusServing}
}

// WithWindow wraps fn so that its results are evaluated against a sliding
// window of recent outcomes. When the failure rate exceeds FailureThreshold
// the wrapper returns NotServing regardless of the current call result.
func WithWindow(cfg WindowConfig, fn HealthCheckFn) HealthCheckFn {
	state := newWindowState(cfg)
	return func(ctx context.Context, service string) StatusResult {
		res := fn(ctx, service)
		return state.record(res.Error == nil && res.Status == StatusServing)
	}
}

// WithWindowDefault calls WithWindow with DefaultWindowConfig.
func WithWindowDefault(fn HealthCheckFn) HealthCheckFn {
	return WithWindow(DefaultWindowConfig(), fn)
}

// WindowTicker periodically decays the window by injecting a synthetic
// "unknown" sample, preventing stale data from dominating forever.
func WindowTicker(ctx context.Context, interval time.Duration, fn HealthCheckFn, service string) {
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				_ = fn(ctx, service)
			}
		}
	}()
}
