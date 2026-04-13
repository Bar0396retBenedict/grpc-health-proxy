package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/yourorg/grpc-health-proxy/internal/config"
)

// Server wraps an HTTP server with graceful shutdown support.
type Server struct {
	httpServer *http.Server
	cfg        *config.Config
}

// New creates a new Server with the given config and handler mux.
func New(cfg *config.Config, mux http.Handler) *Server {
	return &Server{
		cfg: cfg,
		httpServer: &http.Server{
			Addr:         cfg.HTTPAddr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
	}
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	fmt.Printf("HTTP server listening on %s\n", s.cfg.HTTPAddr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}
