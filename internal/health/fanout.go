package health

import (
	"context"
	"sync"
)

// FanoutConfig controls how fanout health checks behave.
type FanoutConfig struct {
	// Label is used for logging/tracing purposes.
	Label string
}

// DefaultFanoutConfig returns a sensible default configuration.
func DefaultFanoutConfig() FanoutConfig {
	return FanoutConfig{
		Label: "fanout",
	}
}

// WithFanout runs fn against each of the provided targets concurrently and
// returns the first successful (SERVING) result. If all targets fail or return
// non-serving status the last result is returned.
//
// This is useful when you want to probe multiple equivalent backends and
// succeed as soon as any one of them is healthy.
func WithFanout(targets []string, fn func(ctx context.Context, target string) StatusResult, cfg FanoutConfig) HealthCheckFn {
	return func(ctx context.Context) StatusResult {
		if len(targets) == 0 {
			return StatusResult{Status: StatusUnknown}
		}

		type result struct {
			sr     StatusResult
			target string
		}

		resultCh := make(chan result, len(targets))
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var wg sync.WaitGroup
		for _, t := range targets {
			wg.Add(1)
			go func(target string) {
				defer wg.Done()
				sr := fn(ctx, target)
				resultCh <- result{sr: sr, target: target}
			}(t)
		}

		go func() {
			wg.Wait()
			close(resultCh)
		}()

		var last StatusResult
		for r := range resultCh {
			last = r.sr
			if r.sr.IsServing() {
				cancel()
				// drain remaining results so goroutines can exit
				go func() {
					for range resultCh { //nolint:revive
					}
				}()
				return r.sr
			}
		}
		return last
	}
}
