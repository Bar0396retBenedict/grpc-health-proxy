package health

import (
	"context"
	"log"
)

// FallbackConfig controls behaviour of the fallback wrapper.
type FallbackConfig struct {
	// Label is used in log messages to identify this fallback.
	Label string
	// FallbackOnUnknown treats Unknown status as a failure and triggers the fallback.
	FallbackOnUnknown bool
}

// DefaultFallbackConfig returns a FallbackConfig with sensible defaults.
func DefaultFallbackConfig() FallbackConfig {
	return FallbackConfig{
		Label:             "fallback",
		FallbackOnUnknown: false,
	}
}

// WithFallback wraps primary with a fallback HealthcheckFn. If primary returns
// an error (or Unknown when FallbackOnUnknown is set), fallback is invoked and
// its result is returned instead. The primary error is logged but not surfaced.
func WithFallback(cfg FallbackConfig, primary, fallback HealthcheckFn) HealthcheckFn {
	return func(ctx context.Context, service string) StatusResult {
		result := primary(ctx, service)
		if result.Error != nil {
			log.Printf("[%s] primary check failed for %q: %v; invoking fallback", cfg.Label, service, result.Error)
			return fallback(ctx, service)
		}
		if cfg.FallbackOnUnknown && result.Status == StatusUnknown {
			log.Printf("[%s] primary returned unknown for %q; invoking fallback", cfg.Label, service)
			return fallback(ctx, service)
		}
		return result
	}
}
