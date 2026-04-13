package health

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc/health/grpc_health_v1"
)

// Poller periodically checks the health of a gRPC service and updates a Cache.
type Poller struct {
	checker  *Checker
	cache    *Cache
	interval time.Duration
}

// NewPoller creates a new Poller with the given checker, cache, and poll interval.
func NewPoller(checker *Checker, cache *Cache, interval time.Duration) *Poller {
	return &Poller{
		checker:  checker,
		cache:    cache,
		interval: interval,
	}
}

// Start begins polling the gRPC health endpoint for the given service name.
// It blocks until the context is cancelled, then returns.
func (p *Poller) Start(ctx context.Context, service string) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	// Run an immediate check before waiting for the first tick.
	p.poll(ctx, service)

	for {
		select {
		case <-ctx.Done():
			log.Printf("poller: stopping for service %q", service)
			return
		case <-ticker.C:
			p.poll(ctx, service)
		}
	}
}

func (p *Poller) poll(ctx context.Context, service string) {
	status, err := p.checker.Check(ctx, service)
	if err != nil {
		log.Printf("poller: health check error for service %q: %v", service, err)
		p.cache.Set(service, grpc_health_v1.HealthCheckResponse_UNKNOWN, err)
		return
	}
	p.cache.Set(service, status, nil)
	log.Printf("poller: service %q status=%v", service, status)
}
