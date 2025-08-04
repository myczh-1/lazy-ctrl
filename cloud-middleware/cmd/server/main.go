package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/myczh-1/lazy-ctrl-cloud/internal/app"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/config"
)

var (
	configPath = flag.String("config", "configs/config.yaml", "Path to config file")
	version    = flag.Bool("version", false, "Show version")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("lazy-ctrl-cloud v1.0.0\n")
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create application
	application, err := app.NewApplication(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: application.Router(),
	}

	go func() {
		log.Printf("Starting HTTP server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start gRPC server
	go func() {
		log.Printf("Starting gRPC server on port %d", cfg.GRPC.Port)
		if err := application.StartGRPCServer(); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	if err := application.Stop(ctx); err != nil {
		log.Printf("Application stop error: %v", err)
	}

	log.Println("Servers exited")
}