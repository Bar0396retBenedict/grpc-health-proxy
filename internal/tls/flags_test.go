package tls_test

import (
	"flag"
	"testing"

	internaltls "github.com/yourorg/grpc-health-proxy/internal/tls"
)

func TestRegisterFlags_Defaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := internaltls.RegisterFlags(fs)

	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if cfg.CACertFile != "" {
		t.Errorf("expected empty CACertFile, got %q", cfg.CACertFile)
	}
	if cfg.ClientCertFile != "" {
		t.Errorf("expected empty ClientCertFile, got %q", cfg.ClientCertFile)
	}
	if cfg.ClientKeyFile != "" {
		t.Errorf("expected empty ClientKeyFile, got %q", cfg.ClientKeyFile)
	}
	if cfg.ServerName != "" {
		t.Errorf("expected empty ServerName, got %q", cfg.ServerName)
	}
	if cfg.Insecure {
		t.Error("expected Insecure to be false by default")
	}
}

func TestRegisterFlags_ParsesValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := internaltls.RegisterFlags(fs)

	args := []string{
		"-tls-ca-cert", "/etc/certs/ca.pem",
		"-tls-client-cert", "/etc/certs/client.pem",
		"-tls-client-key", "/etc/certs/client.key",
		"-tls-server-name", "my-service",
		"-tls-insecure",
	}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if cfg.CACertFile != "/etc/certs/ca.pem" {
		t.Errorf("CACertFile mismatch: %q", cfg.CACertFile)
	}
	if cfg.ClientCertFile != "/etc/certs/client.pem" {
		t.Errorf("ClientCertFile mismatch: %q", cfg.ClientCertFile)
	}
	if cfg.ClientKeyFile != "/etc/certs/client.key" {
		t.Errorf("ClientKeyFile mismatch: %q", cfg.ClientKeyFile)
	}
	if cfg.ServerName != "my-service" {
		t.Errorf("ServerName mismatch: %q", cfg.ServerName)
	}
	if !cfg.Insecure {
		t.Error("expected Insecure to be true")
	}
}
