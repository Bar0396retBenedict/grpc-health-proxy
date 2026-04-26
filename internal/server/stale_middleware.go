package server

import (
	"net/http"
	"strconv"
	"time"
)

// StalenessHeader is the response header that indicates the age of a stale
// cached health result in seconds, when the proxy served a cached value.
const StalenessHeader = "X-Health-Cache-Age"

// StaleHeaderMiddleware injects an X-Health-Cache-Age header into responses
// that were served from the stale cache. It detects staleness by inspecting
// the X-Served-From-Cache request context header set by upstream handlers.
//
// Usage: wrap a health handler that sets "X-Served-From-Cache: <unix-ts>" on
// the request before delegating, then this middleware converts the timestamp
// into a human-readable age header on the response.
func StaleHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		servedAt := r.Header.Get("X-Served-From-Cache")
		if servedAt == "" {
			return
		}

		ts, err := strconv.ParseInt(servedAt, 10, 64)
		if err != nil {
			return
		}

		age := time.Since(time.Unix(ts, 0))
		if age < 0 {
			age = 0
		}
		w.Header().Set(StalenessHeader, strconv.FormatInt(int64(age.Seconds()), 10))
	})
}

// MarkServedFromCache stamps the request with the Unix timestamp at which the
// cached result was recorded. Handlers that serve stale data should call this
// before writing their response so that StaleHeaderMiddleware can annotate the
// outbound response.
func MarkServedFromCache(r *http.Request, cachedAt time.Time) {
	r.Header.Set("X-Served-From-Cache", strconv.FormatInt(cachedAt.Unix(), 10))
}
