package health

import (
	"context"
	"fmt"
	"sync"
)

// JobRegistry manages a set of Jobs and integrates with a Scheduler.
type JobRegistry struct {
	mu        sync.Mutex
	jobs      map[string]*Job
	scheduler *Scheduler
}

// NewJobRegistry creates a JobRegistry backed by the given Scheduler.
func NewJobRegistry(s *Scheduler) *JobRegistry {
	return &JobRegistry{
		jobs:      make(map[string]*Job),
		scheduler: s,
	}
}

// Add registers a job and schedules it.
func (r *JobRegistry) Add(job *Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.jobs[job.service]; exists {
		return fmt.Errorf("job for service %q already registered", job.service)
	}
	r.jobs[job.service] = job
	r.scheduler.Register(job.service, job.Run)
	return nil
}

// Remove unregisters a job by service name.
func (r *JobRegistry) Remove(service string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.jobs, service)
}

// Services returns the list of registered service names.
func (r *JobRegistry) Services() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	names := make([]string, 0, len(r.jobs))
	for k := range r.jobs {
		names = append(names, k)
	}
	return names
}

// Run starts the scheduler.
func (r *JobRegistry) Run(ctx context.Context) {
	r.scheduler.Run(ctx)
}
