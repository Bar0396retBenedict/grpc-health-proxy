package health

import "sync"

// EventLog stores recent health events per service.
type EventLog struct {
	mu      sync.Mutex
	events  []Event
	maxSize int
}

// NewEventLog creates an EventLog capped at maxSize entries.
func NewEventLog(maxSize int) *EventLog {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &EventLog{maxSize: maxSize}
}

// Append adds an event to the log, evicting the oldest if full.
func (l *EventLog) Append(e Event) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.events) >= l.maxSize {
		l.events = l.events[1:]
	}
	l.events = append(l.events, e)
}

// All returns a snapshot of all stored events.
func (l *EventLog) All() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

// Latest returns the most recent event, and false if the log is empty.
func (l *EventLog) Latest() (Event, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.events) == 0 {
		return Event{}, false
	}
	return l.events[len(l.events)-1], true
}
