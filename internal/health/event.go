package health

import (
	"fmt"
	"time"
)

// EventKind classifies what triggered a health event.
type EventKind string

const (
	EventKindPoll    EventKind = "poll"
	EventKindWatcher EventKind = "watcher"
	EventKindManual  EventKind = "manual"
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

// String returns a human-readable summary of the event.
func (e Event) String() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] service=%q status=%s kind=%s err=%v",
			e.Timestamp.Format(time.RFC3339), e.Service, e.Status, e.Kind, e.Err)
	}
	return fmt.Sprintf("[%s] service=%q status=%s kind=%s",
		e.Timestamp.Format(time.RFC3339), e.Service, e.Status, e.Kind)
}
