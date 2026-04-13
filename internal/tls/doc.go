// Package tls provides utilities for constructing TLS configurations used
// when dialing upstream gRPC servers.
//
// It supports:
//   - Plain insecure connections (for development/testing)
//   - Server-side TLS with a custom CA certificate
//   - Mutual TLS (mTLS) with client certificate and key
//   - SNI server name override via ServerName
//
// Typical usage:
//
//	cfg := tls.Config{
//	    CACertFile: "/etc/certs/ca.pem",
//	    ServerName: "my-grpc-service",
//	}
//	tlsCfg, err := tls.BuildClientTLSConfig(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	creds := credentials.NewTLS(tlsCfg)
//	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(creds))
package tls
