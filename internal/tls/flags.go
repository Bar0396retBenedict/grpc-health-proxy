package tls

import "flag"

// RegisterFlags registers TLS-related CLI flags into the given FlagSet
// and returns a pointer to the populated Config.
func RegisterFlags(fs *flag.FlagSet) *Config {
	cfg := &Config{}
	fs.StringVar(&cfg.CACertFile, "tls-ca-cert", "",
		"Path to the CA certificate used to verify the gRPC server's certificate")
	fs.StringVar(&cfg.ClientCertFile, "tls-client-cert", "",
		"Path to the client certificate for mTLS authentication")
	fs.StringVar(&cfg.ClientKeyFile, "tls-client-key", "",
		"Path to the client private key for mTLS authentication")
	fs.StringVar(&cfg.ServerName, "tls-server-name", "",
		"Override the server name used for TLS SNI and certificate verification")
	fs.BoolVar(&cfg.Insecure, "tls-insecure", false,
		"Disable TLS certificate verification (not recommended in production)")
	return cfg
}
