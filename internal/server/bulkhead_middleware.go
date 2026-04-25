package server

import (
	"net/http"
	"sync/atomic"
)

// BulkheadMiddleware limits the number of concurrently handled HTTP requests.
// Requests that arrive when the limit is already reached receive 503 immediately.
func BulkheadMiddleware(maxConcurrent int64) func(http.Handler) http.Handler {
	var inflight atomic.Int64

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			current := inflight.Add(1)
			defer inflight.Add(-1)

			if current > maxConcurrent {
				http.Error(w, "service unavailable: too many concurrent requests", http.StatusServiceUnavailable)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
