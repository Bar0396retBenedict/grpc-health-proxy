package health

import "google.golang.org/grpc/health/grpc_health_v1"

// Status represents the health state of a service.
type Status int

const (
	StatusUnknown  Status = iota
	StatusServing
	StatusNotServing
)

// FromProto converts a gRPC health status to our internal Status.
func FromProto(s grpc_health_v1.HealthCheckResponse_ServingStatus) Status {
	switch s {
	case grpc_health_v1.HealthCheckResponse_SERVING:
		return StatusServing
	case grpc_health_v1.HealthCheckResponse_NOT_SERVING:
		return StatusNotServing
	default:
		return StatusUnknown
	}
}

// String returns a human-readable representation of the status.
func (s Status) String() string {
	switch s {
	case StatusServing:
		return "SERVING"
	case StatusNotServing:
		return "NOT_SERVING"
	default:
		return "UNKNOWN"
	}
}

// IsServing returns true when the status indicates healthy.
func (s Status) IsServing() bool {
	return s == StatusServing
}
