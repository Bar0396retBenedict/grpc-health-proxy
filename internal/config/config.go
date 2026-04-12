package config

import (
	"flag"
	"fmt"
	"time"
)

// Config holds all configuration for the proxy server.
type Config struct {
	// GRPCAddr is the address of the upstream gRPC server to health-check.
	GRPCAddr string

	// HTTPAddr is the address the HTTP server will listen on.
	HTTPAddr string

	// ServiceName is the gRPC service name to check (empty string checks overall server health).
	ServiceName string

	// DialTimeout is the timeout for establishing the gRPC connection.
	DialTimeout time.Duration

	// CheckInterval is how often to poll the gRPC health endpoint.
	CheckInterval time.Duration

	// UseTLS indicates whether to use TLS when connecting to the gRPC server.
	UseTLS bool

	// TLSCACert is the path to the CA certificate for TLS verification.
	TLSCACert string
}

// FromFlags parses configuration from command-line flags.
func FromFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.GRPCAddr, "grpc-addr", "localhost:50051", "Address of the upstream gRPC server")
	flag.StringVar(&cfg.HTTPAddr, "http-addr", ":8080", "Address for the HTTP health endpoint to listen on")
	flag.StringVar(&cfg.ServiceName, "service", "", "gRPC service name to check (empty = server-level check)")
	flag.DurationVar(&cfg.DialTimeout, "dial-timeout", 5*time.Second, "Timeout for gRPC connection dial")
	flag.DurationVar(&cfg.CheckInterval, "check-interval", 10*time.Second, "Interval between gRPC health checks")
	flag.BoolVar(&cfg.UseTLS, "tls", false, "Use TLS when connecting to the gRPC server")
	flag.StringVar(&cfg.TLSCACert, "tls-ca-cert", "", "Path to CA certificate for TLS (empty = system pool)")

	flag.Parse()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.GRPCAddr == "" {
		return fmt.Errorf("grpc-addr must not be empty")
	}
	if c.HTTPAddr == "" {
		return fmt.Errorf("http-addr must not be empty")
	}
	if c.DialTimeout <= 0 {
		return fmt.Errorf("dial-timeout must be positive")
	}
	if c.CheckInterval <= 0 {
		return fmt.Errorf("check-interval must be positive")
	}
	return nil
}
