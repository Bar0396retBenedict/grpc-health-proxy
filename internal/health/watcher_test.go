package health

import (
	"context"
	"testing"
	"time"
)

func TestWatcher_ReceivesStatusChange(t *testing.T) {
	cache := NewCache()
	w := NewWatcher(cache)

	updates := make(chan Status, 4)
	sub := w.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Watch(ctx, "svc", updates)

	updates <- StatusHealthy

	select {
	case s := <-sub:
		if s != StatusHealthy {
			t.Fatalf("expected Healthy, got %v", s)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for status")
	}
}

func TestWatcher_NoBroadcastOnSameStatus(t *testing.T) {
	cache := NewCache()
	w := NewWatcher(cache)

	updates := make(chan Status, 4)
	sub := w.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Watch(ctx, "svc", updates)

	updates <- StatusHealthy
	updates <- StatusHealthy // duplicate — should not be broadcast

	// Drain first
	select {
	case <-sub:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first status")
	}

	// Second should not arrive
	select {
	case s := <-sub:
		t.Fatalf("unexpected second broadcast: %v", s)
	case <-time.After(100 * time.Millisecond):
		// expected
	}
}

func TestWatcher_ClosesOnContextCancel(t *testing.T) {
	cache := NewCache()
	w := NewWatcher(cache)

	updates := make(chan Status)
	sub := w.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		w.Watch(ctx, "svc", updates)
		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Watch did not return after context cancel")
	}

	// channel should be closed
	select {
	case _, ok := <-sub:
		if ok {
			t.Fatal("expected channel to be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("channel was not closed")
	}
}

func TestWatcher_MultipleSubscribers(t *testing.T) {
	cache := NewCache()
	w := NewWatcher(cache)

	updates := make(chan Status, 4)
	sub1 := w.Subscribe()
	sub2 := w.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Watch(ctx, "svc", updates)

	updates <- StatusUnhealthy

	for i, sub := range []<-chan Status{sub1, sub2} {
		select {
		case s := <-sub:
			if s != StatusUnhealthy {
				t.Fatalf("subscriber %d: expected Unhealthy, got %v", i, s)
			}
		case <-time.After(time.Second):
			t.Fatalf("subscriber %d: timed out", i)
		}
	}
}
