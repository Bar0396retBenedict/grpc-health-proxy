package health

import "sync"

// Deduplicator suppresses redundant status-change notifications by only
// forwarding a new status when it differs from the previously seen one.
// It is safe for concurrent use.
type Deduplicator struct {
	mu   sync.Mutex
	last map[string]Status
}

// NewDeduplicator returns an initialised Deduplicator.
func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		last: make(map[string]Status),
	}
}

// Changed reports whether the given status for service differs from the last
// recorded status. If it does differ (or if no status has been recorded yet)
// the new status is stored and true is returned. Otherwise false is returned
// and the stored status is left unchanged.
func (d *Deduplicator) Changed(service string, s Status) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	prev, seen := d.last[service]
	if seen && prev == s {
		return false
	}

	d.last[service] = s
	return true
}

// Reset clears the recorded status for service so that the next call to
// Changed for that service always returns true.
func (d *Deduplicator) Reset(service string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.last, service)
}

// ResetAll clears all recorded statuses.
func (d *Deduplicator) ResetAll() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.last = make(map[string]Status)
}
