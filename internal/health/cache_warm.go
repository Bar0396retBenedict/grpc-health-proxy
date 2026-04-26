package health

import (
	"context"
	"log"
	"sync"
	"time"
)

// WarmupConfig controls how the cache is pre-populated before the proxy
// starts serving traffic.
type WarmupConfig struct {
	// Timeout is the maximum time to wait for all services to return an
	// initial status before giving up and allowing traffic anyway.
	Timeout time.Duration
	// Services is the list of gRPC service names to warm up.
	Services []string
}

// DefaultWarmupConfig returns sensible defaults for cache warm-up.
func DefaultWarmupConfig() WarmupConfig {
	return WarmupConfig{
		Timeout: 5 * time.Second,
	}
}

// WarmCache calls fn for every service in cfg.Services concurrently and
// stores the result in cache. It returns once all checks complete or
// cfg.Timeout elapses, whichever comes first. Missing results are left as
// StatusUnknown.
func WarmCache(ctx context.Context, cfg WarmupConfig, cache *Cache, fn HealthCheckFn) error {
	if len(cfg.Services) == 0 {
		return nil
	}

	warmCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	var wg sync.WaitGroup
	for _, svc := range cfg.Services {
		wg.Add(1)
		go func(service string) {
			defer wg.Done()
			result := fn(warmCtx, service)
			cache.Set(service, result.Status, result.Err)
			if result.Err != nil {
				log.Printf("[warmup] service=%s err=%v", service, result.Err)
			} else {
				log.Printf("[warmup] service=%s status=%s", service, result.Status)
			}
		}(svc)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-warmCtx.Done():
		return warmCtx.Err()
	}
}
