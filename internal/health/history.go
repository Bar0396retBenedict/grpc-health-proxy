package health

import (
	"sync"
	"time"

	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

// StatusEvent records a single health status transition.
type StatusEvent struct {
	Status    grpc_health_v1.HealthCheckResponse_ServingStatus
	Timestamp time.Time
}

// History tracks the last N status events for a service.
type History struct {
	mu      sync.Mutex
	events  []StatusEvent
	maxSize int
}

// NewHistory creates a History that retains up to maxSize events.
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 10
	}
	return &History{maxSize: maxSize}
}

// Record appends a new status event.
func (h *History) Record(status grpc_health_v1.HealthCheckResponse_ServingStatus) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, StatusEvent{Status: status, Timestamp: time.Now()})
	if len(h.events) > h.maxSize {
		h.events = h.events[len(h.events)-h.maxSize:]
	}
}

// Events returns a copy of the recorded events, oldest first.
func (h *History) Events() []StatusEvent {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]StatusEvent, len(h.events))
	copy(out, h.events)
	return out
}

// Last returns the most recent event and true, or zero value and false if empty.
func (h *History) Last() (StatusEvent, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.events) == 0 {
		return StatusEvent{}, false
	}
	return h.events[len(h.events)-1], true
}
