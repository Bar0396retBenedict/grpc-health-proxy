package health

import "time"

// EventKind classifies what triggered a health event.
type EventKind string

const (
	EventKindPoll      EventKind = "poll"
	EventKindWatcher   EventKind = "watcher"
	EventKindManual    EventKind = "manual"
)

// Event captures a single health status observation.
type Event struct {
	Service   string
	Status    Status
	Kind      EventKind
	Err       error
	Timestamp time.Time
}

// NewEvent constructs an Event with the current time.
func NewEvent(service string, status Status, kind EventKind, err error) Event {
	return Event{
		Service:   service,
		Status:    status,
		Kind:      kind,
		Err:       err,
		Timestamp: time.Now(),
	}
}

// IsHealthy returns true when the event represents a healthy state.
func (e Event) IsHealthy() bool {
	return e.Status.IsServing() && e.Err == nil
}
