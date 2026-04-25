package health

import (
	"context"
	"time"
)

// HedgeConfig controls hedged request behaviour.
type HedgeConfig struct {
	// Delay is how long to wait before issuing the hedge request.
	Delay time.Duration
	// MaxHedges is the maximum number of additional requests to issue.
	MaxHedges int
}

// DefaultHedgeConfig returns a sensible default configuration.
func DefaultHedgeConfig() HedgeConfig {
	return HedgeConfig{
		Delay:     50 * time.Millisecond,
		MaxHedges: 1,
	}
}

// WithHedge wraps fn so that if it does not return within cfg.Delay a second
// (hedged) call is issued concurrently. The first successful result wins; if
// all calls fail the last error is returned.
func WithHedge(cfg HedgeConfig, fn HealthCheckFn) HealthCheckFn {
	if cfg.MaxHedges <= 0 || cfg.Delay <= 0 {
		return fn
	}
	return func(ctx context.Context, service string) StatusResult {
		type result struct {
			sr StatusResult
		}

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		resultCh := make(chan result, cfg.MaxHedges+1)

		launch := func() {
			go func() {
				resultCh <- result{fn(ctx, service)}
			}()
		}

		launch()

		timer := time.NewTimer(cfg.Delay)
		defer timer.Stop()

		hedged := 0
		var last StatusResult
		received := 0
		total := 1

		for {
			select {
			case <-ctx.Done():
				return StatusResult{Status: StatusUnknown, Err: ctx.Err()}
			case <-timer.C:
				if hedged < cfg.MaxHedges {
					launch()
					hedged++
					total++
					timer.Reset(cfg.Delay)
				}
			case r := <-resultCh:
				received++
				if r.sr.Err == nil {
					cancel()
					return r.sr
				}
				last = r.sr
				if received == total {
					return last
				}
			}
		}
	}
}
