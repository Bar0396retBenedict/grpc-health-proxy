package health

import (
	"context"
	"log"
)

// ShadowConfig controls how shadow (dark-launch) health checks behave.
type ShadowConfig struct {
	// Label is used in log output to identify the shadow target.
	Label string
}

// DefaultShadowConfig returns a ShadowConfig with sensible defaults.
func DefaultShadowConfig() ShadowConfig {
	return ShadowConfig{
		Label: "shadow",
	}
}

// WithShadow runs fn and shadow concurrently. The result of fn is always
// returned to the caller; the shadow result is logged but otherwise discarded.
// This is useful for dark-launching a new upstream without affecting traffic.
func WithShadow(fn HealthCheckFn, shadow HealthCheckFn, cfg ShadowConfig) HealthCheckFn {
	return func(ctx context.Context, service string) StatusResult {
		// Fire shadow in background.
		go func() {
			result := shadow(ctx, service)
			if result.Err != nil {
				log.Printf("shadow[%s] service=%s error=%v", cfg.Label, service, result.Err)
			} else {
				log.Printf("shadow[%s] service=%s status=%s", cfg.Label, service, result.Status)
			}
		}()

		return fn(ctx, service)
	}
}
