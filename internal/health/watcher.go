package health

import (
	"context"
	"log"
	"sync"
)

// Watcher monitors cache state changes and notifies registered listeners.
type Watcher struct {
	cache     *Cache
	listeners []chan Status
	mu        sync.Mutex
}

// NewWatcher creates a Watcher that observes the given Cache.
func NewWatcher(cache *Cache) *Watcher {
	return &Watcher{
		cache: cache,
	}
}

// Subscribe returns a channel that receives Status updates for the given service.
// The channel is buffered with capacity 1 to avoid blocking the watcher loop.
func (w *Watcher) Subscribe() <-chan Status {
	ch := make(chan Status, 1)
	w.mu.Lock()
	w.listeners = append(w.listeners, ch)
	w.mu.Unlock()
	return ch
}

// Unsubscribe removes the given channel from the listener list and closes it.
// It is safe to call Unsubscribe concurrently with Watch.
func (w *Watcher) Unsubscribe(sub <-chan Status) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, ch := range w.listeners {
		if ch == sub {
			w.listeners = append(w.listeners[:i], w.listeners[i+1:]...)
			close(ch)
			return
		}
	}
}

// Watch polls the cache for the named service at each tick from the poller
// and broadcasts status changes to all subscribers until ctx is cancelled.
func (w *Watcher) Watch(ctx context.Context, service string, updates <-chan Status) {
	var last Status = StatusUnknown
	for {
		select {
		case <-ctx.Done():
			w.closeAll()
			return
		case s, ok := <-updates:
			if !ok {
				w.closeAll()
				return
			}
			if s != last {
				log.Printf("[watcher] service=%q status changed: %v -> %v", service, last, s)
				last = s
				w.broadcast(s)
			}
		}
	}
}

func (w *Watcher) broadcast(s Status) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, ch := range w.listeners {
		select {
		case ch <- s:
		default:
			// drop if listener is not consuming
		}
	}
}

func (w *Watcher) closeAll() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, ch := range w.listeners {
		close(ch)
	}
	w.listeners = nil
}
