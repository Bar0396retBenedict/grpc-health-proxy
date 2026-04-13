package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourorg/grpc-health-proxy/internal/config"
	"github. err := config.FromFlags(flag.CommandLine, os.Args[1:])
	if err != nil {
		log.Fatalf("invalid configurationn	poller := health.NewPoller(cfg, checker, cache)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go poller.Run(ctx)

	mux := server.NewServeMux(cfg, cache)
	srv := server.New(cfg, mux)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-quit
	log.Println("received shutdown signal")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("server exited cleanly")
}
