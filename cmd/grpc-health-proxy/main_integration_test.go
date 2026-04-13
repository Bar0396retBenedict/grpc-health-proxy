//go:build integration

package main_test

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

// TestMain_Integration starts the binary and checks that /livez responds.
// Run with: go test -tags=integration ./cmd/grpc-health-proxy/...
func TestMain_Integration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", ".",
		"-http-addr", "127.0.0.1:18080",
		"-grpc-addr", "127.0.0.1:50099",
		"-dial-timeout", "1s",
		"-check-interval", "2s",
	)

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start binary: %v", err)
	}
	defer cmd.Process.Kill() //nolint:errcheck

	// Allow the server to start.
	time.Sleep(500 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:18080/livez"))
	if err != nil {
		t.Fatalf("failed to reach /livez: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 from /livez, got %d", resp.StatusCode)
	}
}
