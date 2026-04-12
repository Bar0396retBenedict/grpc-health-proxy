package config

import (
	"testing"
	"time"
)

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		GRPCAddr:      "localhost:50051",
		HTTPAddr:      ":8080",
		DialTimeout:   5 * time.Second,
		CheckInterval: 10 * time.Second,
	}

	if err := cfg.validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_EmptyGRPCAddr(t *testing.T) {
	cfg := &Config{
		GRPCAddr:      "",
		HTTPAddr:      ":8080",
		DialTimeout:   5 * time.Second,
		CheckInterval: 10 * time.Second,
	}

	if err := cfg.validate(); err == nil {
		t.Fatal("expected error for empty grpc-addr, got nil")
	}
}

func TestValidate_EmptyHTTPAddr(t *testing.T) {
	cfg := &Config{
		GRPCAddr:      "localhost:50051",
		HTTPAddr:      "",
		DialTimeout:   5 * time.Second,
		CheckInterval: 10 * time.Second,
	}

	if err := cfg.validate(); err == nil {
		t.Fatal("expected error for empty http-addr, got nil")
	}
}

func TestValidate_ZeroDialTimeout(t *testing.T) {
	cfg := &Config{
		GRPCAddr:      "localhost:50051",
		HTTPAddr:      ":8080",
		DialTimeout:   0,
		CheckInterval: 10 * time.Second,
	}

	if err := cfg.validate(); err == nil {
		t.Fatal("expected error for zero dial-timeout, got nil")
	}
}

func TestValidate_ZeroCheckInterval(t *testing.T) {
	cfg := &Config{
		GRPCAddr:      "localhost:50051",
		HTTPAddr:      ":8080",
		DialTimeout:   5 * time.Second,
		CheckInterval: 0,
	}

	if err := cfg.validate(); err == nil {
		t.Fatal("expected error for zero check-interval, got nil")
	}
}

func TestValidate_WithServiceName(t *testing.T) {
	cfg := &Config{
		GRPCAddr:      "myservice:9090",
		HTTPAddr:      ":9000",
		ServiceName:   "my.package.MyService",
		DialTimeout:   3 * time.Second,
		CheckInterval: 5 * time.Second,
		UseTLS:        true,
		TLSCACert:     "/etc/certs/ca.pem",
	}

	if err := cfg.validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}
