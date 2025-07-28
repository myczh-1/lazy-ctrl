package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/myczh-1/lazy-ctrl-agent/internal/interface/http"
	"github.com/myczh-1/lazy-ctrl-agent/internal/interface/grpc"
	"github.com/myczh-1/lazy-ctrl-agent/internal/interface/mqtt"
)

// Server interface defines the contract for all server types
type Server interface {
	Start() error
}

// Stopper interface defines graceful shutdown capability
type Stopper interface {
	Stop()
}

// Application represents the main application
type Application struct {
	container *Container
	servers   []Server
}

// NewApplication creates a new application instance
func NewApplication(configPath string) (*Application, error) {
	container, err := NewContainer(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	
	app := &Application{
		container: container,
		servers:   make([]Server, 0),
	}
	
	// Initialize servers based on configuration
	if err := app.initializeServers(); err != nil {
		return nil, fmt.Errorf("failed to initialize servers: %w", err)
	}
	
	return app, nil
}

// initializeServers creates and configures all enabled servers
func (a *Application) initializeServers() error {
	cfg := a.container.Config
	logger := a.container.Logger
	
	// Initialize HTTP server if enabled
	if cfg.Server.HTTP.Enabled {
		httpServer := http.NewServer(
			cfg,
			logger,
			a.container.CommandService,
			a.container.ExecutorService,
			a.container.SecurityService,
		)
		a.servers = append(a.servers, httpServer)
		logger.WithField("port", cfg.Server.HTTP.Port).Info("HTTP server enabled")
	}
	
	// Initialize gRPC server if enabled
	if cfg.Server.GRPC.Enabled {
		grpcServer := grpc.NewServer(
			cfg,
			logger,
			a.container.CommandService,
			a.container.ExecutorService,
			a.container.SecurityService,
		)
		a.servers = append(a.servers, grpcServer)
		logger.WithField("port", cfg.Server.GRPC.Port).Info("gRPC server enabled")
	}
	
	// Initialize MQTT client if enabled
	if cfg.MQTT.Enabled {
		mqttClient := mqtt.NewClient(
			cfg,
			logger,
			a.container.CommandService,
			a.container.ExecutorService,
			a.container.SecurityService,
		)
		a.servers = append(a.servers, mqttClient)
		logger.WithField("broker", cfg.MQTT.Broker).Info("MQTT client enabled")
	}
	
	if len(a.servers) == 0 {
		return fmt.Errorf("no servers enabled in configuration")
	}
	
	return nil
}

// Run starts the application and blocks until shutdown
func (a *Application) Run() error {
	logger := a.container.Logger
	logger.Info("Starting application...")
	
	// Start background tasks
	a.startBackgroundTasks()
	
	// Start all servers
	_, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	var wg sync.WaitGroup
	errChan := make(chan error, len(a.servers))
	
	// Start all servers concurrently
	for _, server := range a.servers {
		wg.Add(1)
		go func(s Server) {
			defer wg.Done()
			if err := s.Start(); err != nil {
				errChan <- fmt.Errorf("server start failed: %w", err)
			}
		}(server)
	}
	
	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	select {
	case err := <-errChan:
		logger.WithError(err).Error("Server error occurred")
		cancel()
		return err
	case sig := <-sigChan:
		logger.WithField("signal", sig).Info("Received shutdown signal")
		cancel()
	}
	
	// Perform graceful shutdown
	return a.shutdown(&wg)
}

// startBackgroundTasks starts background maintenance tasks
func (a *Application) startBackgroundTasks() {
	logger := a.container.Logger
	securityService := a.container.SecurityService
	
	// Start rate limiter cleanup task
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		
		logger.Debug("Starting rate limiter cleanup task")
		for range ticker.C {
			securityService.CleanupRateLimiter()
		}
	}()
	
	// Add other background tasks here as needed
	// For example: metrics collection, health checks, etc.
}

// shutdown performs graceful shutdown of all components
func (a *Application) shutdown(wg *sync.WaitGroup) error {
	logger := a.container.Logger
	logger.Info("Shutting down application...")
	
	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	
	// Channel to signal shutdown completion
	shutdownChan := make(chan struct{})
	
	go func() {
		// Stop all servers that support graceful shutdown
		for _, server := range a.servers {
			if stopper, ok := server.(Stopper); ok {
				stopper.Stop()
			}
		}
		
		// Wait for all server goroutines to complete
		wg.Wait()
		
		// Shutdown application container
		a.container.Shutdown()
		
		close(shutdownChan)
	}()
	
	// Wait for graceful shutdown or timeout
	select {
	case <-shutdownChan:
		logger.Info("Graceful shutdown completed successfully")
		return nil
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing exit")
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// GetContainer returns the application container for testing purposes
func (a *Application) GetContainer() *Container {
	return a.container
}