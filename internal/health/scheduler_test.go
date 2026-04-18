package health

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_RunsJobAtInterval(t *testing.T) {
	cfg := SchedulerConfig{Interval: 20 * time.Millisecond, Jitter: 0}
	s := NewScheduler(cfg)

	var count int64
	s.Register("test", func(ctx context.Context) {
		atomic.AddInt64(&count, 1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Millisecond)
	defer cancel()
	s.Run(ctx)

	got := atomic.LoadInt64(&count)
	if got < 3 {
		t.Errorf("expected at least 3 invocations, got %d", got)
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	cfg := SchedulerConfig{Interval: 10 * time.Millisecond, Jitter: 0}
	s := NewScheduler(cfg)

	var count int64
	s.Register("job", func(ctx context.Context) {
		atomic.AddInt64(&count, 1)
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s.Run(ctx)

	got := atomic.LoadInt64(&count)
	if got > 2 {
		t.Errorf("expected at most 2 invocations after immediate cancel, got %d", got)
	}
}

func TestScheduler_MultipleJobs(t *testing.T) {
	cfg := SchedulerConfig{Interval: 20 * time.Millisecond, Jitter: 0}
	s := NewScheduler(cfg)

	var a, b int64
	s.Register("a", func(ctx context.Context) { atomic.AddInt64(&a, 1) })
	s.Register("b", func(ctx context.Context) { atomic.AddInt64(&b, 1) })

	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()
	s.Run(ctx)

	if atomic.LoadInt64(&a) < 2 {
		t.Error("job a ran too few times")
	}
	if atomic.LoadInt64(&b) < 2 {
		t.Error("job b ran too few times")
	}
}

func TestScheduler_DefaultConfig(t *testing.T) {
	cfg := DefaultSchedulerConfig()
	if cfg.Interval <= 0 {
		t.Error("expected positive interval")
	}
	if cfg.Jitter < 0 {
		t.Error("expected non-negative jitter")
	}
}
