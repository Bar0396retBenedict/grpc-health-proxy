package server

import (
	"net/http"
	"sync"
	"time"
)

// CoalesceMiddleware deduplicates concurrent identical HTTP requests within a
// short window. All requests that arrive while a response is being prepared
// receive the same buffered response, reducing upstream load.
//
// Only GET requests are coalesced; other methods pass through unchanged.
func CoalesceMiddleware(window time.Duration, next http.Handler) http.Handler {
	if window <= 0 {
		return next
	}

	type flight struct {
		done   chan struct{}
		status int
		body   []byte
		header http.Header
	}

	var (
		mu      sync.Mutex
		flights = map[string]*flight{}
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		key := r.URL.Path

		mu.Lock()
		if f, ok := flights[key]; ok {
			mu.Unlock()
			<-f.done
			for k, vals := range f.header {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(f.status)
			_, _ = w.Write(f.body)
			return
		}

		f := &flight{done: make(chan struct{})}
		flights[key] = f
		mu.Unlock()

		time.Sleep(window)

		rec := newResponseWriter(w)
		next.ServeHTTP(rec, r)

		f.status = rec.status
		f.body = rec.body
		f.header = rec.Header()

		mu.Lock()
		delete(flights, key)
		mu.Unlock()

		close(f.done)
	})
}
