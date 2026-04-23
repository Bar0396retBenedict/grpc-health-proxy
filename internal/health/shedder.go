package health

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// ErrShed is returned when a request is shed due to load.
var ErrShed = errors.New("health: request shed due to load")

// DefaultShedderConfig returns a ShedderConfig with sensible defaults.
func DefaultShedderConfig() ShedderConfig {
	return ShedderConfig{
		MaxConcurrent: 10,
		Cooldown:      500 * time.Millisecond,
	}
}

// ShedderConfig controls load-shedding behaviour.
type ShedderConfig struct {
	// MaxConcurrent is the maximum number of in-flight health checks allowed.
	MaxConcurrent int64
	// Cooldown is the minimum time between shed events before allowing through
	// again (used to avoid thundering-herd on recovery).
	Cooldown time.Duration
}

type shedder struct {
	cfg       ShedderConfig
	inflight  atomic.Int64
	lastShed  atomic.Int64 // UnixNano of last shed event
}

// WithShedder wraps a HealthCheckFn with a load-shedder that rejects calls
// when the number of concurrent in-flight checks exceeds cfg.MaxConcurrent.
func WithShedder(cfg ShedderConfig, fn HealthCheckFn) HealthCheckFn {
	s := &shedder{cfg: cfg}
	return func(ctx context.Context, service string) StatusResult {
		current := s.inflight.Add(1)
		defer s.inflight.Add(-1)

		if current > cfg.MaxConcurrent {
			now := time.Now().UnixNano()
			s.lastShed.Store(now)
			return StatusResult{Status: StatusUnknown, Err: ErrShed}
		}

		// If we are within the cooldown window after a shed, be conservative
		// and return the last known-unknown rather than hammering the backend.
		last := s.lastShed.Load()
		if last != 0 {
			sinceLastShed := time.Duration(time.Now().UnixNano() - last)
			if sinceLastShed < cfg.Cooldown {
				return StatusResult{Status: StatusUnknown, Err: ErrShed}
			}
		}

		return fn(ctx, service)
	}
}
