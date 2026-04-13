package metrics

import (
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
	return "health_check_total " + itoa(c.HealthCheckTotal.Load()) + "\n" +
		"health_check_success " + itoa(c.HealthCheckSuccess.Load()) + "\n" +
		"health_check_failure " + itoa(c.HealthCheckFailure.Load()) + "\n" +
		"http_request_total " + itoa(c.HTTPRequestTotal.Load()) + "\n"
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
