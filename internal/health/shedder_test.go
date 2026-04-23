package health

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestWithShedder_AllowsUnderLimit(t *testing.T) {
	cfg := ShedderConfig{MaxConcurrent: 5, Cooldown: 50 * time.Millisecond}
	calls := 0
	fn := WithShedder(cfg, func(_ context.Context, _ string) StatusResult {
		calls++
		return StatusResult{Status: StatusServing}
	})

	result := fn(context.Background(), "svc")
	if result.Err != nil {
		t.Fatalf("expected no error, got %v", result.Err)
	}
	if result.Status != StatusServing {
		t.Fatalf("expected Serving, got %v", result.Status)
	}
	if calls != 1 {
		t.Fatalf("expected inner fn called once, got %d", calls)
	}
}

func TestWithShedder_ShedsOverLimit(t *testing.T) {
	const max = 3
	cfg := ShedderConfig{MaxConcurrent: max, Cooldown: 10 * time.Millisecond}

	ready := make(chan struct{})
	block := make(chan struct{})

	slow := func(_ context.Context, _ string) StatusResult {
		ready <- struct{}{}
		<-block
		return StatusResult{Status: StatusServing}
	}

	fn := WithShedder(cfg, slow)

	var wg sync.WaitGroup
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(context.Background(), "svc")
		}()
		<-ready // ensure goroutine is inside fn before continuing
	}

	// Now all max slots are occupied; the next call should be shed.
	result := fn(context.Background(), "svc")
	if result.Err != ErrShed {
		t.Fatalf("expected ErrShed, got %v", result.Err)
	}
	if result.Status != StatusUnknown {
		t.Fatalf("expected Unknown, got %v", result.Status)
	}

	close(block)
	wg.Wait()
}

func TestWithShedder_CooldownBlocksAfterShed(t *testing.T) {
	cfg := ShedderConfig{MaxConcurrent: 1, Cooldown: 200 * time.Millisecond}

	block := make(chan struct{})
	fn := WithShedder(cfg, func(_ context.Context, _ string) StatusResult {
		<-block
		return StatusResult{Status: StatusServing}
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn(context.Background(), "svc")
	}()

	time.Sleep(10 * time.Millisecond) // let goroutine enter fn

	// Trigger a shed to record lastShed timestamp.
	shed := fn(context.Background(), "svc")
	if shed.Err != ErrShed {
		t.Fatalf("expected ErrShed on first shed, got %v", shed.Err)
	}

	close(block)
	wg.Wait()

	// Even though inflight is now 0, we are within cooldown — should still shed.
	result := fn(context.Background(), "svc")
	if result.Err != ErrShed {
		t.Fatalf("expected ErrShed during cooldown, got %v", result.Err)
	}
}

func TestWithShedder_AllowsAfterCooldown(t *testing.T) {
	cfg := ShedderConfig{MaxConcurrent: 1, Cooldown: 20 * time.Millisecond}

	block := make(chan struct{})
	fn := WithShedder(cfg, func(_ context.Context, _ string) StatusResult {
		select {
		case <-block:
		default:
		}
		return StatusResult{Status: StatusServing}
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn(context.Background(), "svc")
	}()
	time.Sleep(5 * time.Millisecond)

	// Shed to set lastShed.
	fn(context.Background(), "svc")
	close(block)
	wg.Wait()

	// Wait for cooldown to expire.
	time.Sleep(30 * time.Millisecond)

	result := fn(context.Background(), "svc")
	if result.Err != nil {
		t.Fatalf("expected no error after cooldown, got %v", result.Err)
	}
	if result.Status != StatusServing {
		t.Fatalf("expected Serving after cooldown, got %v", result.Status)
	}
}
