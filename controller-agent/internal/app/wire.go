package app

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/infrastructure"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/service"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/security"
)

// Container holds all application dependencies
type Container struct {
	Config          *config.Config
	Logger          *logrus.Logger
	CommandService  *service.CommandService
	ExecutorService *executor.Service
	SecurityService *security.Service
}

// NewContainer creates and initializes all application dependencies
func NewContainer(configPath string) (*Container, error) {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	
	// Load configuration
	if err := config.Load(configPath); err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	
	cfg := config.Get()
	
	// Configure logger based on config
	if level, err := logrus.ParseLevel(cfg.Log.Level); err == nil {
		logger.SetLevel(level)
	}
	
	if cfg.Log.Format == "text" {
		logger.SetFormatter(&logrus.TextFormatter{})
	}
	
	// Setup log output if specified
	if cfg.Log.OutputPath != "" {
		// Note: File output setup moved to main.go to handle file closure properly
		logger.WithField("log_file", cfg.Log.OutputPath).Info("Log file configured")
	}
	
	// Initialize command repository
	commandRepo := infrastructure.NewFileCommandRepository(cfg.Commands.ConfigPath)
	
	// Initialize command repository with data
	if fileRepo, ok := commandRepo.(*infrastructure.FileCommandRepository); ok {
		if err := fileRepo.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize command repository: %w", err)
		}
	}
	
	// Initialize services
	commandService := service.NewCommandService(commandRepo)
	executorService := executor.NewService(logger)
	securityService := security.NewService(cfg, logger)
	
	container := &Container{
		Config:          cfg,
		Logger:          logger,
		CommandService:  commandService,
		ExecutorService: executorService,
		SecurityService: securityService,
	}
	
	logger.WithFields(logrus.Fields{
		"config_loaded": true,
		"services_initialized": true,
	}).Info("Application container initialized successfully")
	
	return container, nil
}

// Shutdown gracefully shuts down all components
func (c *Container) Shutdown() {
	c.Logger.Info("Shutting down application container")
	
	// Add any cleanup logic here if needed
	// For example, closing database connections, flushing logs, etc.
	
	c.Logger.Info("Application container shutdown complete")
}