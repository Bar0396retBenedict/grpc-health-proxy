package health

import (
	"errors"
	"testing"
	"time"
)

func TestReporter_ReportServing(t *testing.T) {
	cache := NewCache()
	cache.Set("svc", StatusServing, nil)
	r := NewReporter(cache, NewHistory(5))

	rep := r.Report("svc")
	if rep.Service != "svc" {
		t.Errorf("expected svc, got %s", rep.Service)
	}
	if rep.Status != StatusServing {
		t.Errorf("expected Serving, got %s", rep.Status)
	}
	if rep.Err != nil {
		t.Errorf("unexpected error: %v", rep.Err)
	}
	if rep.CheckedAt.IsZero() {
		t.Error("CheckedAt should not be zero")
	}
}

func TestReporter_ReportWithError(t *testing.T) {
	cache := NewCache()
	errDown := errors.New("down")
	cache.Set("svc", StatusNotServing, errDown)
	r := NewReporter(cache, NewHistory(5))

	rep := r.Report("svc")
	if rep.Status != StatusNotServing {
		t.Errorf("expected NotServing, got %s", rep.Status)
	}
	if rep.Err != errDown {
		t.Errorf("expected errDown, got %v", rep.Err)
	}
}

func TestReporter_ReportUnknownService(t *testing.T) {
	cache := NewCache()
	r := NewReporter(cache, NewHistory(5))

	rep := r.Report("missing")
	if rep.Status != StatusUnknown {
		t.Errorf("expected Unknown, got %s", rep.Status)
	}
}

func TestReporter_RecentEvents(t *testing.T) {
	cache := NewCache()
	h := NewHistory(10)
	for i := 0; i < 5; i++ {
		h.Record("svc", NewEvent("svc", StatusServing, nil, time.Now()))
	}
	r := NewReporter(cache, h)

	events := r.RecentEvents("svc", 3)
	if len(events) != 3 {
		t.Errorf("expected 3 events, got %d", len(events))
	}
}

func TestReporter_RecentEvents_FewerThanN(t *testing.T) {
	cache := NewCache()
	h := NewHistory(10)
	h.Record("svc", NewEvent("svc", StatusServing, nil, time.Now()))
	r := NewReporter(cache, h)

	events := r.RecentEvents("svc", 10)
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
}
