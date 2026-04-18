package health

import (
	"testing"

	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestFromProto_Serving(t *testing.T) {
	s := FromProto(grpc_health_v1.HealthCheckResponse_SERVING)
	if s != StatusServing {
		t.Fatalf("expected StatusServing, got %v", s)
	}
}

func TestFromProto_NotServing(t *testing.T) {
	s := FromProto(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	if s != StatusNotServing {
		t.Fatalf("expected StatusNotServing, got %v", s)
	}
}

func TestFromProto_Unknown(t *testing.T) {
	s := FromProto(grpc_health_v1.HealthCheckResponse_UNKNOWN)
	if s != StatusUnknown {
		t.Fatalf("expected StatusUnknown, got %v", s)
	}
}

func TestStatus_String(t *testing.T) {
	cases := []struct {
		s    Status
		want string
	}{
		{StatusServing, "SERVING"},
		{StatusNotServing, "NOT_SERVING"},
		{StatusUnknown, "UNKNOWN"},
	}
	for _, c := range cases {
		if got := c.s.String(); got != c.want {
			t.Errorf("Status(%d).String() = %q, want %q", c.s, got, c.want)
		}
	}
}

func TestStatus_IsServing(t *testing.T) {
	if !StatusServing.IsServing() {
		t.Error("StatusServing.IsServing() should be true")
	}
	if StatusNotServing.IsServing() {
		t.Error("StatusNotServing.IsServing() should be false")
	}
	if StatusUnknown.IsServing() {
		t.Error("StatusUnknown.IsServing() should be false")
	}
}
