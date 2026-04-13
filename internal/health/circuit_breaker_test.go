package health

import (
	"testing"
	"time"
)

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	if !cb.Allow() {
		t.Fatal("expected circuit breaker to allow requests initially")
	}
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected state Closed, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cfg := CircuitBreakerConfig{FailureThreshold: 3, SuccessThreshold: 2, OpenDuration: 10 * time.Second}
	cb := NewCircuitBreaker(cfg)

	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}

	if cb.CurrentState() != StateOpen {
		t.Fatalf("expected state Open after threshold, got %v", cb.CurrentState())
	}
	if cb.Allow() {
		t.Fatal("expected circuit breaker to block requests when open")
	}
}

func TestCircuitBreaker_HalfOpenAfterDuration(t *testing.T) {
	cfg := CircuitBreakerConfig{FailureThreshold: 1, SuccessThreshold: 1, OpenDuration: 50 * time.Millisecond}
	cb := NewCircuitBreaker(cfg)
	cb.RecordFailure()

	if cb.CurrentState() != StateOpen {
		t.Fatal("expected Open state")
	}

	time.Sleep(60 * time.Millisecond)

	if !cb.Allow() {
		t.Fatal("expected circuit breaker to allow after open duration")
	}
	if cb.CurrentState() != StateHalfOpen {
		t.Fatalf("expected HalfOpen state, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_ClosesAfterSuccessThreshold(t *testing.T) {
	cfg := CircuitBreakerConfig{FailureThreshold: 1, SuccessThreshold: 2, OpenDuration: 50 * time.Millisecond}
	cb := NewCircuitBreaker(cfg)
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	cb.Allow() // transitions to HalfOpen

	cb.RecordSuccess()
	if cb.CurrentState() != StateHalfOpen {
		t.Fatal("expected still HalfOpen after one success")
	}
	cb.RecordSuccess()
	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected Closed after success threshold, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_ReopensOnFailureInHalfOpen(t *testing.T) {
	cfg := CircuitBreakerConfig{FailureThreshold: 1, SuccessThreshold: 2, OpenDuration: 50 * time.Millisecond}
	cb := NewCircuitBreaker(cfg)
	cb.RecordFailure()
	time.Sleep(60 * time.Millisecond)
	cb.Allow() // transitions to HalfOpen

	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatalf("expected Open after failure in HalfOpen, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_SuccessResetFailureCount(t *testing.T) {
	cfg := CircuitBreakerConfig{FailureThreshold: 3, SuccessThreshold: 1, OpenDuration: 10 * time.Second}
	cb := NewCircuitBreaker(cfg)

	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordSuccess() // should reset failure count
	cb.RecordFailure()
	cb.RecordFailure()

	if cb.CurrentState() != StateClosed {
		t.Fatalf("expected Closed, got %v", cb.CurrentState())
	}
}
