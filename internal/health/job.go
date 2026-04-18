package health

import (
	"context"
	"log"
)

// Job wraps a HealthCheckFn and writes results to a Cache and EventLog.
type Job struct {
	service string
	check   HealthCheckFn
	cache   *Cache
	log     *EventLog
}

// NewJob creates a Job for the given service.
func NewJob(service string, check HealthCheckFn, cache *Cache, el *EventLog) *Job {
	return &Job{
		service: service,
		check:   check,
		cache:   cache,
		log:     el,
	}
}

// Run executes the health check and updates the cache and event log.
func (j *Job) Run(ctx context.Context) {
	result := j.check(ctx)
	j.cache.Set(j.service, result.Status, result.Err)

	evt := NewEvent(j.service, result.Status, result.Err)
	j.log.Record(evt)

	if result.Err != nil {
		log.Printf("[health] service=%s status=%s err=%v", j.service, result.Status, result.Err)
	} else {
		log.Printf("[health] service=%s status=%s", j.service, result.Status)
	}
}
