package health

import (
	"sync"
	"time"
)

// CachedStatus holds the last known health status and when it was recorded.
type CachedStatus struct {
	Status    Status
	Err       error
	CheckedAt time.Time
}

// Cache stores the most recent health check result and provides thread-safe access.
type Cache struct {
	mu     sync.RWMutex
	latest CachedStatus
}

// NewCache initialises a Cache with an unknown status.
func NewCache() *Cache {
	return &Cache{
		latest: CachedStatus{
			Status:    StatusUnknown,
			CheckedAt: time.Now(),
		},
	}
}

// Set stores a new health result.
func (c *Cache) Set(status Status, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.latest = CachedStatus{
		Status:    status,
		Err:       err,
		CheckedAt: time.Now(),
	}
}

// Get returns the most recent cached status.
func (c *Cache) Get() CachedStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latest
}
