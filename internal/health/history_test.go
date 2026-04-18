package health

import (
	"testing"

	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

var (
	serving2    = grpc_health_v1.HealthCheckResponse_SERVING
	notServing2 = grpc_health_v1.HealthCheckResponse_NOT_SERVING
)

func TestHistory_EmptyLast(t *testing.T) {
	h := NewHistory(5)
	_, ok := h.Last()
	if ok {
		t.Fatal("expected false for empty history")
	}
}

func TestHistory_RecordAndLast(t *testing.T) {
	h := NewHistory(5)
	h.Record(serving2)
	e, ok := h.Last()
	if !ok {
		t.Fatal("expected event")
	}
	if e.Status != serving2 {
		t.Fatalf("expected SERVING, got %v", e.Status)
	}
}

func TestHistory_EventsOrder(t *testing.T) {
	h := NewHistory(5)
	h.Record(serving2)
	h.Record(notServing2)
	events := h.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Status != serving2 {
		t.Errorf("first event should be SERVING")
	}
	if events[1].Status != notServing2 {
		t.Errorf("second event should be NOT_SERVING")
	}
}

func TestHistory_CapsAtMaxSize(t *testing.T) {
	h := NewHistory(3)
	for i := 0; i < 6; i++ {
		h.Record(serving2)
	}
	if len(h.Events()) != 3 {
		t.Fatalf("expected 3 events after capping, got %d", len(h.Events()))
	}
}

func TestHistory_DefaultMaxSize(t *testing.T) {
	h := NewHistory(0)
	for i := 0; i < 15; i++ {
		h.Record(serving2)
	}
	if len(h.Events()) != 10 {
		t.Fatalf("expected default cap of 10, got %d", len(h.Events()))
	}
}

func TestHistory_EventsCopy(t *testing.T) {
	h := NewHistory(5)
	h.Record(serving2)
	events := h.Events()
	events[0].Status = notServing2
	_, _ = h.Last()
	e, _ := h.Last()
	if e.Status != serving2 {
		t.Error("Events() should return a copy, not a reference")
	}
}
