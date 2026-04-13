package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRecordHealthCheck(t *testing.T) {
	c := &Counters{}
	c.RecordHealthCheck(true)
	c.RecordHealthCheck(true)
	c.RecordHealthCheck(false)

	if got := c.HealthCheckTotal.Load(); got != 3 {
		t.Errorf("total: want 3, got %d", got)
	}
	if got := c.HealthCheckSuccess.Load(); got != 2 {
		t.Errorf("success: want 2, got %d", got)
	}
	if got := c.HealthCheckFailure.Load(); got != 1 {
		t.Errorf("failure: want 1, got %d", got)
	}
}

func TestRecordHTTPRequest(t *testing.T) {
	c := &Counters{}
	for i := 0; i < 5; i++ {
		c.RecordHTTPRequest()
	}
	if got := c.HTTPRequestTotal.Load(); got != 5 {
		t.Errorf("http_request_total: want 5, got %d", got)
	}
}

func TestReset(t *testing.T) {
	c := &Counters{}
	c.RecordHealthCheck(true)
	c.RecordHTTPRequest()
	c.Reset()

	if c.HealthCheckTotal.Load() != 0 || c.HTTPRequestTotal.Load() != 0 {
		t.Error("Reset did not zero counters")
	}
}

func TestHandler_OutputFormat(t *testing.T) {
	c := &Counters{}
	c.RecordHealthCheck(true)
	c.RecordHealthCheck(false)
	c.RecordHTTPRequest()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	c.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: want 200, got %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	s := string(body)
	for _, want := range []string{
		"health_check_total 2",
		"health_check_success 1",
		"health_check_failure 1",
		"http_request_total 1",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q; got:\n%s", want, s)
		}
	}
}

func TestConcurrentRecording(t *testing.T) {
	c := &Counters{}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.RecordHealthCheck(i%2 == 0)
			c.RecordHTTPRequest()
		}(i)
	}
	wg.Wait()

	if got := c.HealthCheckTotal.Load(); got != 100 {
		t.Errorf("concurrent total: want 100, got %d", got)
	}
	if got := c.HTTPRequestTotal.Load(); got != 100 {
		t.Errorf("concurrent http: want 100, got %d", got)
	}
}
