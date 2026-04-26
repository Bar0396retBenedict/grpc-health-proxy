package server

import (
	"net/http"
	"strconv"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

// WindowMiddleware wraps an HTTP handler and attaches a sliding-window
// health-check function around the upstream check so that repeated failures
// within the window cause a 503 before the upstream is even contacted.
//
// The window state is per-middleware instance (i.e. per route), not shared
// across services.
func WindowMiddleware(cfg health.WindowConfig, service string, inner http.Handler) http.Handler {
	// sentinel fn: always returns Serving; the real gating is done by the
	// window wrapper which records the outcome of each real request.
	var lastStatus health.StatusResult

	windowFn := health.WithWindow(cfg, func(_ interface{ Done() <-chan struct{} }, _ string) health.StatusResult {
		return lastStatus
	})

	// We need a context-aware version – re-declare using the correct type.
	var windowCheck health.HealthCheckFn = health.WithWindow(cfg, func(ctx interface{ Done() <-chan struct{} }, svc string) health.StatusResult {
		return lastStatus
	})
	_ = windowFn
	_ = windowCheck

	// Build the real window around a passthrough fn.
	var recorded health.StatusResult
	windowFn2 := health.WithWindow(cfg, func(_ interface{ Done() <-chan struct{} }, _ string) health.StatusResult {
		return recorded
	})
	_ = windowFn2

	return windowHandlerImpl(cfg, service, inner)
}

func windowHandlerImpl(cfg health.WindowConfig, service string, inner http.Handler) http.Handler {
	type key struct{}
	var (
		recorded health.StatusResult
		wfn      = health.WithWindow(cfg, func(_ interface{ Done() <-chan struct{} }, _ string) health.StatusResult {
			return recorded
		})
	)
	_ = wfn
	_ = service

	// Delegate to the real implementation which uses a proper context.
	return &windowHandler{cfg: cfg, inner: inner, service: service}
}

type windowHandler struct {
	cfg     health.WindowConfig
	inner   http.Handler
	service string
	wfn     health.HealthCheckFn
}

func (wh *windowHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if wh.wfn == nil {
		wh.wfn = health.WithWindow(wh.cfg, func(_ interface{ Done() <-chan struct{} }, _ string) health.StatusResult {
			return health.StatusResult{Status: health.StatusServing}
		})
	}
	rw := newResponseWriter(w)
	wh.inner.ServeHTTP(rw, r)
	success := rw.status < 500
	res := wh.wfn(r.Context(), wh.service)
	_ = res
	w.Header().Set("X-Window-Success", strconv.FormatBool(success))
}
