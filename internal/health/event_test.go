package health

import (
	"errors"
	"testing"
	"time"
)

func TestNewEvent_SetsFields(t *testing.T) {
	before := time.Now()
	e := NewEvent("svc", StatusServing, EventKindPoll, nil)
	after := time.Now()

	if e.Service != "svc" {
		t.Errorf("Service = %q, want svc", e.Service)
	}
	if e.Status != StatusServing {
		t.Errorf("Status = %v, want StatusServing", e.Status)
	}
	if e.Kind != EventKindPoll {
		t.Errorf("Kind = %v, want EventKindPoll", e.Kind)
	}
	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Error("Timestamp out of expected range")
	}
}

func TestEvent_IsHealthy_True(t *testing.T) {
	e := NewEvent("svc", StatusServing, EventKindPoll, nil)
	if !e.IsHealthy() {
		t.Error("expected IsHealthy() true")
	}
}

func TestEvent_IsHealthy_FalseOnNotServing(t *testing.T) {
	e := NewEvent("svc", StatusNotServing, EventKindPoll, nil)
	if e.IsHealthy() {
		t.Error("expected IsHealthy() false for NOT_SERVING")
	}
}

func TestEvent_IsHealthy_FalseOnError(t *testing.T) {
	e := NewEvent("svc", StatusServing, EventKindPoll, errors.New("oops"))
	if e.IsHealthy() {
		t.Error("expected IsHealthy() false when err != nil")
	}
}

func TestEvent_IsHealthy_FalseOnUnknown(t *testing.T) {
	e := NewEvent("svc", StatusUnknown, EventKindWatcher, nil)
	if e.IsHealthy() {
		t.Error("expected IsHealthy() false for UNKNOWN")
	}
}
