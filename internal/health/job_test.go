package health

import (
	"context"
	"errors"
	"testing"
)

func TestJob_RunUpdatesCache(t *testing.T) {
	cache := NewCache()
	el := NewEventLog(10)

	job := NewJob("svc", func(ctx context.Context) StatusResult {
		return StatusResult{Status: StatusServing}
	}, cache, el)

	job.Run(context.Background())

	status, err := cache.Get("svc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != StatusServing {
		t.Errorf("expected Serving, got %v", status)
	}
}

func TestJob_RunRecordsEvent(t *testing.T) {
	cache := NewCache()
	el := NewEventLog(10)

	job := NewJob("svc", func(ctx context.Context) StatusResult {
		return StatusResult{Status: StatusNotServing}
	}, cache, el)

	job.Run(context.Background())

	events := el.All()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Service != "svc" {
		t.Errorf("expected service svc, got %s", events[0].Service)
	}
}

func TestJob_RunWithError(t *testing.T) {
	cache := NewCache()
	el := NewEventLog(10)
	expectedErr := errors.New("dial failed")

	job := NewJob("svc", func(ctx context.Context) StatusResult {
		return StatusResult{Status: StatusUnknown, Err: expectedErr}
	}, cache, el)

	job.Run(context.Background())

	_, err := cache.Get("svc")
	if err == nil {
		t.Error("expected error in cache")
	}
}
