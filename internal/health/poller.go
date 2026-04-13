package health

import (
	"context"
	"log/slog"
	"time"
)

// Poller periodically runs a health check and writes results to a Cache.
type Poller struct {
	checker     *Checker
	cache       *Cache
	interval    time.Duration
	serviceName string
	logger      *slog.Logger
}

// NewPoller creates a Poller that checks the given service on the provided interval.
func NewPoller(checker *Checker, cache *Cache, interval time.Duration, serviceName string, logger *slog.Logger) *Poller {
	if logger == nil {
		logger = slog.Default()
	}
	return &Poller{
		checker:     checker,
		cache:       cache,
		interval:    interval,
		serviceName: serviceName,
		logger:      logger,
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (p *Poller) Run(ctx context.Context) {
	p.poll(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.poll(ctx)
		case <-ctx.Done():
			p.logger.Info("health poller stopped")
			return
		}
	}
}

func (p *Poller) poll(ctx context.Context) {
	status, err := p.checker.Check(ctx, p.serviceName)
	p.cache.Set(status, err)

	if err != nil {
		p.logger.Warn("health check failed",
			slog.String("service", p.serviceName),
			slog.String("error", err.Error()),
		)
		return
	}
	p.logger.Debug("health check completed",
		slog.String("service", p.serviceName),
		slog.String("status", status.String()),
	)
}
