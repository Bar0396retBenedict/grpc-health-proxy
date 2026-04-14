package health

import (
	"sync"
	"testing"
)

func TestDeduplicator_FirstCallAlwaysChanged(t *testing.T) {
	d := NewDeduplicator()
	if !d.Changed("svc", StatusHealthy) {
		t.Fatal("expected Changed=true for first observation")
	}
}

func TestDeduplicator_SameStatusNotChanged(t *testing.T) {
	d := NewDeduplicator()
	d.Changed("svc", StatusHealthy)
	if d.Changed("svc", StatusHealthy) {
		t.Fatal("expected Changed=false for identical consecutive status")
	}
}

func TestDeduplicator_DifferentStatusChanged(t *testing.T) {
	d := NewDeduplicator()
	d.Changed("svc", StatusHealthy)
	if !d.Changed("svc", StatusUnhealthy) {
		t.Fatal("expected Changed=true when status transitions")
	}
}

func TestDeduplicator_IndependentServices(t *testing.T) {
	d := NewDeduplicator()
	d.Changed("a", StatusHealthy)
	d.Changed("b", StatusUnhealthy)

	// Same status for each service — should not be changed.
	if d.Changed("a", StatusHealthy) {
		t.Fatal("service a: expected Changed=false")
	}
	if d.Changed("b", StatusUnhealthy) {
		t.Fatal("service b: expected Changed=false")
	}

	// Transition only service a.
	if !d.Changed("a", StatusUnhealthy) {
		t.Fatal("service a: expected Changed=true after transition")
	}
	if d.Changed("b", StatusUnhealthy) {
		t.Fatal("service b: expected Changed=false, unaffected by service a")
	}
}

func TestDeduplicator_ResetMakesNextCallChanged(t *testing.T) {
	d := NewDeduplicator()
	d.Changed("svc", StatusHealthy)
	d.Reset("svc")
	if !d.Changed("svc", StatusHealthy) {
		t.Fatal("expected Changed=true after Reset")
	}
}

func TestDeduplicator_ResetAll(t *testing.T) {
	d := NewDeduplicator()
	d.Changed("a", StatusHealthy)
	d.Changed("b", StatusUnhealthy)
	d.ResetAll()

	for _, svc := range []string{"a", "b"} {
		if !d.Changed(svc, StatusHealthy) {
			t.Fatalf("service %s: expected Changed=true after ResetAll", svc)
		}
	}
}

func TestDeduplicator_ConcurrentAccess(t *testing.T) {
	d := NewDeduplicator()
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			svc := "svc"
			for j := 0; j < 50; j++ {
				s := StatusHealthy
				if j%2 == 0 {
					s = StatusUnhealthy
				}
				d.Changed(svc, s)
			}
		}(i)
	}
	wg.Wait() // should not race
}
