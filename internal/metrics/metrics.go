package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

// Counters holds atomic counters for proxy metrics.
type Counters struct {
	HealthCheckTotal   atomic.Int64
	HealthCheckSuccess atomic.Int64
	HealthCheckFailure atomic.Int64
	HTTPRequestTotal   atomic.Int64
}

// Global is the default metrics instance used by the proxy.
var Global = &Counters{}

// RecordHealthCheck increments the total and success/failure counters.
func (c *Counters) RecordHealthCheck(success bool) {
	c.HealthCheckTotal.Add(1)
	if success {
		c.HealthCheckSuccess.Add(1)
	} else {
		c.HealthCheckFailure.Add(1)
	}
}

// RecordHTTPRequest increments the HTTP request counter.
func (c *Counters) RecordHTTPRequest() {
	c.HTTPRequestTotal.Add(1)
}

// Reset zeroes all counters (useful in tests).
func (c *Counters) Reset() {
	c.HealthCheckTotal.Store(0)
	c.HealthCheckSuccess.Store(0)
	c.HealthCheckFailure.Store(0)
	c.HTTPRequestTotal.Store(0)
}

// Handler returns an HTTP handler that exposes counters as plain text.
func (c *Counters) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(c.Format()))
	})
}

// Format returns a human-readable snapshot of the counters.
func (c *Counters) Format() string {
	return fmt.Sprintf(
		"health_check_total %d\nhealth_check_success %d\nhealth_check_failure %d\nhttp_request_total %d\n",
		c.HealthCheckTotal.Load(),
		c.HealthCheckSuccess.Load(),
		c.HealthCheckFailure.Load(),
		c.HTTPRequestTotal.Load(),
	)
}
