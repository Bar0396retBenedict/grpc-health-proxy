package tls_test

import (
	"crypto/tls"
	"os"
	"testing"

	internaltls "github.com/yourorg/grpc-health-proxy/internal/tls"
)

func TestBuildClientTLSConfig_Insecure(t *testing.T) {
	cfg := internaltls.Config{Insecure: true}
	tlsCfg, err := internaltls.BuildClientTLSConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tlsCfg.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be true")
	}
}

func TestBuildClientTLSConfig_DefaultSecure(t *testing.T) {
	cfg := internaltls.Config{}
	tlsCfg, err := internaltls.BuildClientTLSConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tlsCfg.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify to be false")
	}
	if tlsCfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinVersion TLS 1.2, got %v", tlsCfg.MinVersion)
	}
}

func TestBuildClientTLSConfig_ServerName(t *testing.T) {
	cfg := internaltls.Config{ServerName: "example.com"}
	tlsCfg, err := internaltls.BuildClientTLSConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tlsCfg.ServerName != "example.com" {
		t.Errorf("expected ServerName 'example.com', got %q", tlsCfg.ServerName)
	}
}

func TestBuildClientTLSConfig_MissingCACert(t *testing.T) {
	cfg := internaltls.Config{CACertFile: "/nonexistent/ca.crt"}
	_, err := internaltls.BuildClientTLSConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing CA cert file")
	}
}

func TestBuildClientTLSConfig_InvalidCACert(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "ca-*.pem")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	_, _ = f.WriteString("not a valid pem")
	_ = f.Close()

	cfg := internaltls.Config{CACertFile: f.Name()}
	_, err = internaltls.BuildClientTLSConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid CA cert PEM")
	}
}

func TestBuildClientTLSConfig_MissingClientKeyPair(t *testing.T) {
	cfg := internaltls.Config{
		ClientCertFile: "/nonexistent/client.crt",
		ClientKeyFile:  "/nonexistent/client.key",
	}
	_, err := internaltls.BuildClientTLSConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing client key pair")
	}
}
