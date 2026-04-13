package health

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

var errFake = errors.New("fake error")

func TestWithRetry_SucceedsOnFirstAttempt(t *testing.T) {
	calls := 0
	err := WithRetry(context.Background(), DefaultRetryConfig(), func(_ context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_RetriesOnFailure(t *testing.T) {
	var calls int32
	cfg := RetryConfig{MaxAttempts: 3, Delay: time.Millisecond}
	err := WithRetry(context.Background(), cfg, func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		if atomic.LoadInt32(&calls) < 3 {
			return errFake
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error after retries, got %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_ReturnsLastError(t *testing.T) {
	cfg := RetryConfig{MaxAttempts: 3, Delay: time.Millisecond}
	err := WithRetry(context.Background(), cfg, func(_ context.Context) error {
		return errFake
	})
	if !errors.Is(err, errFake) {
		t.Fatalf("expected errFake, got %v", err)
	}
}

func TestWithRetry_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	err := WithRetry(ctx, DefaultRetryConfig(), func(_ context.Context) error {
		calls++
		return errFake
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected 0 calls for pre-cancelled context, got %d", calls)
	}
}

func TestWithRetry_CancelDuringDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := RetryConfig{MaxAttempts: 5, Delay: 500 * time.Millisecond}

	var calls int32
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := WithRetry(ctx, cfg, func(_ context.Context) error {
		atomic.AddInt32(&calls, 1)
		return errFake
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if atomic.LoadInt32(&calls) > 2 {
		t.Fatalf("expected at most 2 calls before cancel, got %d", calls)
	}
}
