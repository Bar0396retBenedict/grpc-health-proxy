package health

import (
	"context"
	"sync"
	"time"
)

// SchedulerConfig holds configuration for the Scheduler.
type SchedulerConfig struct {
	Interval time.Duration
	Jitter   time.Duration
}

// DefaultSchedulerConfig returns a sensible default.
func DefaultSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{
		Interval: 10 * time.Second,
		Jitter:   1 * time.Second,
	}
}

// Scheduler runs a set of named jobs at a fixed interval.
type Scheduler struct {
	cfg  SchedulerConfig
	jobs map[string]func(ctx context.Context)
	mu   sync.Mutex
}

// NewScheduler creates a new Scheduler with the given config.
func NewScheduler(cfg SchedulerConfig) *Scheduler {
	return &Scheduler{
		cfg:  cfg,
		jobs: make(map[string]func(ctx context.Context)),
	}
}

// Register adds a named job to the scheduler.
func (s *Scheduler) Register(name string, fn func(ctx context.Context)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[name] = fn
}

// Run starts all registered jobs and blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	s.mu.Lock()
	snap := make(map[string]func(ctx context.Context), len(s.jobs))
	for k, v := range s.jobs {
		snap[k] = v
	}
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, fn := range snap {
		wg.Add(1)
		go func(job func(ctx context.Context)) {
			defer wg.Done()
			ticker := time.NewTicker(s.cfg.Interval)
			defer ticker.Stop()
			job(ctx)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					job(ctx)
				}
			}
		}(fn)
	}
	wg.Wait()
}
