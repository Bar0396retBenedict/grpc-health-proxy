package health

import (
	"testing"
	"time"
)

func TestDefaultBackoffConfig(t *testing.T) {
	cfg := DefaultBackoffConfig()
	if cfg.InitialInterval != 500*time.Millisecond {
		t.Errorf("expected InitialInterval 500ms, got %v", cfg.InitialInterval)
	}
	if cfg.MaxInterval != 30*time.Second {
		t.Errorf("expected MaxInterval 30s, got %v", cfg.MaxInterval)
	}
	if cfg.Multiplier != 1.5 {
		t.Errorf("expected Multiplier 1.5, got %v", cfg.Multiplier)
	}
	if cfg.MaxElapsed != 0 {
		t.Errorf("expected MaxElapsed 0, got %v", cfg.MaxElapsed)
	}
}

func TestBackoff_FirstAttempt(t *testing.T) {
	cfg := DefaultBackoffConfig()
	d := cfg.Backoff(0)
	if d != 500*time.Millisecond {
		t.Errorf("expected 500ms for attempt 0, got %v", d)
	}
}

func TestBackoff_GrowsWithAttempt(t *testing.T) {
	cfg := DefaultBackoffConfig()
	prev := cfg.Backoff(0)
	for i := 1; i <= 5; i++ {
		curr := cfg.Backoff(i)
		if curr <= prev {
			t.Errorf("expected backoff to grow: attempt %d gave %v <= %v", i, curr, prev)
		}
		prev = curr
	}
}

func TestBackoff_CappedAtMaxInterval(t *testing.T) {
	cfg := DefaultBackoffConfig()
	// After many attempts the delay must not exceed MaxInterval.
	for i := 0; i < 50; i++ {
		d := cfg.Backoff(i)
		if d > cfg.MaxInterval {
			t.Errorf("attempt %d: backoff %v exceeded MaxInterval %v", i, d, cfg.MaxInterval)
		}
	}
}

func TestBackoff_NegativeAttemptTreatedAsZero(t *testing.T) {
	cfg := DefaultBackoffConfig()
	if cfg.Backoff(-3) != cfg.Backoff(0) {
		t.Error("negative attempt should produce same result as attempt 0")
	}
}

func TestExceededMaxElapsed_ZeroNeverExceeds(t *testing.T) {
	cfg := DefaultBackoffConfig() // MaxElapsed == 0
	if cfg.ExceededMaxElapsed(24 * time.Hour) {
		t.Error("expected ExceededMaxElapsed to be false when MaxElapsed is 0")
	}
}

func TestExceededMaxElapsed_ExceedsWhenOverLimit(t *testing.T) {
	cfg := DefaultBackoffConfig()
	cfg.MaxElapsed = 5 * time.Second
	if cfg.ExceededMaxElapsed(4 * time.Second) {
		t.Error("expected false when elapsed < MaxElapsed")
	}
	if !cfg.ExceededMaxElapsed(5 * time.Second) {
		t.Error("expected true when elapsed == MaxElapsed")
	}
	if !cfg.ExceededMaxElapsed(10 * time.Second) {
		t.Error("expected true when elapsed > MaxElapsed")
	}
}
