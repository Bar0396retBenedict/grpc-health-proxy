package health_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/yourorg/grpc-health-proxy/internal/health"
)

func TestCache_DefaultUnknown(t *testing.T) {
	c := health.NewCache()
	got := c.Get()
	if got.Status != health.StatusUnknown {
		t.Errorf("expected UNKNOWN, got %s", got.Status)
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := health.NewCache()
	c.Set(health.StatusHealthy, nil)

	got := c.Get()
	if got.Status != health.StatusHealthy {
		t.Errorf("expected HEALTHY, got %s", got.Status)
	}
	if got.Err != nil {
		t.Errorf("expected nil error, got %v", got.Err)
	}
	if got.CheckedAt.IsZero() {
		t.Error("CheckedAt should not be zero")
	}
}

func TestCache_SetWithError(t *testing.T) {
	c := health.NewCache()
	sentinel := errors.New("backend unavailable")
	c.Set(health.StatusUnhealthy, sentinel)

	got := c.Get()
	if got.Status != health.StatusUnhealthy {
		t.Errorf("expected UNHEALTHY, got %s", got.Status)
	}
	if !errors.Is(got.Err, sentinel) {
		t.Errorf("expected sentinel error, got %v", got.Err)
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := health.NewCache()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.Set(health.StatusHealthy, nil)
		}()
		go func() {
			defer wg.Done()
			_ = c.Get()
		}()
	}
	wg.Wait()
}
