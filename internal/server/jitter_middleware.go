package server

import (
	"math/rand"
	"net/http"
	"time"
)

// JitterMiddleware adds a small random delay to each HTTP response. This is
// useful in scenarios where many clients poll the health endpoint simultaneously
// and you want to spread their retry cadence after a restart.
//
// maxJitter is the upper bound of the uniform random delay. A zero or negative
// value disables the middleware entirely (the next handler is called directly).
func JitterMiddleware(maxJitter time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if maxJitter <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// #nosec G404 — non-cryptographic randomness is intentional here.
			delay := time.Duration(rand.Int63n(int64(maxJitter)))
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				http.Error(w, "request cancelled", http.StatusServiceUnavailable)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
