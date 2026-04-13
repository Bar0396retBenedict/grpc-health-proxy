package metrics

import "net/http"

// RequestCountingMiddleware wraps an http.Handler and increments the HTTP
// request counter on every incoming request.
func RequestCountingMiddleware(c *Counters, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.RecordHTTPRequest()
		next.ServeHTTP(w, r)
	})
}
