package server

import (
	"net/http"

	"github.com/your-org/grpc-health-proxy/internal/health"
)

// RegisterProbeRoutes registers /readyz and /livez probe endpoints on mux.
//
// /readyz uses a Probe with threshold semantics so that transient blips do not
// immediately flip the readiness gate.
//
// /livez is a lightweight one-shot check with no threshold tracking.
func RegisterProbeRoutes(
	mux *http.ServeMux,
	serviceName string,
	checkFn health.HealthCheckFn,
	probeCfg health.ProbeConfig,
) {
	probe := health.NewProbe(probeCfg, checkFn)
	probeHandler := NewProbeHandler(probe, serviceName)

	mux.Handle("/readyz", probeHandler)
	mux.Handle("/livez", ReadyzHandler(checkFn, serviceName))
}
