package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWithCircuitBreaker_AllowsWhenClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	calls := 0
	fn := WithCircuitBreaker(cb, func(ctx context.Context) error {
		calls++
		return nil
	})

	err := fn(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestWithCircuitBreaker_BlocksWhenOpen(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()
	cfg.FailureThreshold = 2
	cfg.OpenDuration = 10 * time.Second
	cb := NewCircuitBreaker(cfg)

	failFn := func(ctx context.Context) error {
		return errors.New("fail")
	}
	protected := WithCircuitBreaker(cb, failFn)

	// Trigger enough failures to open the circuit.
	_ = protected(context.Background())
	_ = protected(context.Background())

	calls := 0
	blocked := WithCircuitBreaker(cb, func(ctx context.Context) error {
		calls++
		return nil
	})

	err := blocked(context.Background())
	if err == nil {
		t.Fatal("expected circuit breaker error, got nil")
	}
	if calls != 0 {
		t.Fatalf("expected 0 calls when circuit open, got %d", calls)
	}
}

func TestWithCircuitBreaker_RecordsSuccess(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	fn := WithCircuitBreaker(cb, func(ctx context.Context) error {
		return nil
	})

	for i := 0; i < 5; i++ {
		if err := fn(context.Background()); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}
}

func TestWithRetryAndCircuitBreaker_RetriesAndRecords(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	attempts := 0
	cfg := RetryConfig{MaxAttempts: 3, Delay: 0}

	fn := WithRetryAndCircuitBreaker(cb, cfg, func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("transient")
		}
		return nil
	})

	err := fn(context.Background())
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
