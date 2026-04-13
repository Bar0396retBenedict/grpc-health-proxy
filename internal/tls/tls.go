package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// Config holds TLS configuration for gRPC connections.
type Config struct {
	CACertFile     string
	ClientCertFile string
	ClientKeyFile  string
	ServerName     string
	Insecure       bool
}

// BuildClientTLSConfig constructs a *tls.Config from the given Config.
// If Insecure is true, all other fields are ignored and server certificate
// verification is disabled.
func BuildClientTLSConfig(cfg Config) (*tls.Config, error) {
	if cfg.Insecure {
		return &tls.Config{InsecureSkipVerify: true}, nil //nolint:gosec
	}

	tlsCfg := &tls.Config{
		ServerName: cfg.ServerName,
		MinVersion: tls.VersionTLS12,
	}

	if cfg.CACertFile != "" {
		caCert, err := os.ReadFile(cfg.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA cert: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA cert from %s", cfg.CACertFile)
		}
		tlsCfg.RootCAs = pool
	}

	if cfg.ClientCertFile != "" && cfg.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.ClientCertFile, cfg.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading client key pair: %w", err)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	return tlsCfg, nil
}
