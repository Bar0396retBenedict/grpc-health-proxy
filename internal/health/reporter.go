package health

import "time"

// Report holds a snapshot of health state for a service.
type Report struct {
	Service   string
	Status    Status
	Err       error
	CheckedAt time.Time
}

// Reporter builds Reports from cached state and event history.
type Reporter struct {
	cache   *Cache
	history *History
}

// NewReporter creates a Reporter backed by the given cache and history.
func NewReporter(cache *Cache, history *History) *Reporter {
	return &Reporter{cache: cache, history: history}
}

// Report returns the current health report for a service.
func (r *Reporter) Report(service string) Report {
	status, err := r.cache.Get(service)
	return Report{
		Service:   service,
		Status:    status,
		Err:       err,
		CheckedAt: time.Now(),
	}
}

// RecentEvents returns the last n events for a service.
func (r *Reporter) RecentEvents(service string, n int) []Event {
	all := r.history.Events(service)
	if n >= len(all) {
		return all
	}
	return all[len(all)-n:]
}
