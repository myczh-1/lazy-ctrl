package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/myczh-1/lazy-ctrl-agent/internal/command/service"
	"github.com/myczh-1/lazy-ctrl-agent/internal/common"
	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-ag
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/security"
	"github.com/myczh-1/lazy-ctrl-agent/internal/common"
)

// Server represents the HTTP server
type Server struct {
	config          *config.Config
	logger          *logrus.Logger
	commandService  *service.CommandService
	executorService *executor.Service
	securityService *security.Service
	engine          *gin.Engine
	server          *http.Server
}

// NewServer creates a new HTTP server instance
func NewServer(
	cfg *config.Config,
	logger *logrus.Logger,
	commandService *service.CommandService,
	executorService *executor.Service,
	securityService *security.Service,
) *Server {
	return &Server{
		config:          cfg,
		logger:          logger,
		commandService:  commandService,
		executorService: executorService,
		securityService: securityService,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.setupEngine()
	s.setupRoutes()

	
	s.logger.WithFields(logrus.Fields{
		"port": s.config.Server.HTTP.Port,
		"host": s.config.Server.HTTP.Host,

	
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed to start: %w", err)

	
	return nil
}

// Stop stops the HTTP server gracefully
func (s *Server) Stop() {

	
	ctx, cancel := context.WithTimeout(context.Background(), common.DefaultShutdownTimeout)

	
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown HTTP server gracefully")
	} else {
		s.logger.Info("HTTP server stopped gracefully")
	}
}

// setupEngine configures the Gin engine
func (s *Server) setupEngine() {
	// Set Gin mode based on log level
	if s.logger.Level == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)

	

	
	// Add middleware
	s.engine.Use(s.requestIDMiddleware())
	s.engine.Use(s.loggingMiddleware())
	s.engine.Use(s.corsMiddleware())
	s.engine.Use(s.recoveryMiddleware())
	s.engine.Use(s.securityMiddleware())
	s.engine.Use(utils.ResponseFormatterMiddleware())
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	// Create handlers
	commandHandler := NewCommandHandler(s.commandService)
	executeHandler := NewExecuteHandler(s.commandService, s.executorService, s.securityService)

	
	// API v1 routes
	v1 := s.engine.Group("/api/v1")
	{
		// System routes
		v1.GET("/health", systemHandler.HealthCheck)
		v1.GET("/version", systemHandler.GetVersion)
		v1.GET("/status", systemHandler.GetStatus)

		
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/verify", systemHandler.VerifyPin)

		
		// Command routes
		commands := v1.Group("/commands")
		{
			commands.POST("", commandHandler.CreateCommand)
			commands.GET("", commandHandler.GetAllCommands)
			commands.GET("/homepage", commandHandler.GetHomepageCommands)
			commands.GET("/:id", commandHandler.GetCommand)
			commands.PUT("/:id", commandHandler.UpdateCommand)
			commands.DELETE("/:id", commandHandler.DeleteCommand)

		
		// Execution routes
		v1.GET("/execute", executeHandler.ExecuteCommand)
		v1.POST("/execute", executeHandler.ExecuteCommandPost)
		v1.GET("/execute/info", executeHandler.GetCommandInfo)

	
	// Serve static files if configured
	if s.config.Server.HTTP.StaticPath != "" && s.config.Server.HTTP.StaticDir != "" {
		s.engine.Static(s.config.Server.HTTP.StaticPath, s.config.Server.HTTP.StaticDir)
		s.logger.WithFields(logrus.Fields{
			"path": s.config.Server.HTTP.StaticPath,
			"dir":  s.config.Server.HTTP.StaticDir,
		}).Info("Serving static files")

	
	// Default route for API documentation or health check
	s.engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   common.AppName,
			"version":   common.AppVersion,
			"status":    common.StatusHealthy,
			"timestamp": utils.GetCurrentTimestamp(),
		})
	})
}

// setupServer configures the HTTP server
func (s *Server) setupServer() {
	s.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", s.config.Server.HTTP.Host, s.config.Server.HTTP.Port),
		Handler:        s.engine,
		ReadTimeout:    common.DefaultHTTPTimeout,
		WriteTimeout:   common.DefaultHTTPTimeout,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}
}

// Middleware implementations

// requestIDMiddleware adds request ID to each request
func (s *Server) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := utils.GetRequestID(c)
		utils.SetRequestID(c, requestID)
		c.Next()
	}
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			s.logger.WithFields(logrus.Fields{
				"method":     param.Method,
				"path":       param.Path,
				"status":     param.StatusCode,
				"latency":    param.Latency,
				"client_ip":  param.ClientIP,
				"user_agent": param.Request.UserAgent(),
				"request_id": param.Keys[string(common.ContextKeyRequestID)],
			}).Info("HTTP request")
			return ""
		},
		Output: s.logger.Writer(),
	})
}

// corsMiddleware handles CORS headers
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		
		// Allow all origins for development
		// In production, this should be configured more restrictively
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Pin, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")

		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return

		
		s.logger.WithField("origin", origin).Debug("CORS request processed")
		c.Next()
	}
}

// recoveryMiddleware handles panics
func (s *Server) recoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		s.logger.WithFields(logrus.Fields{
			"panic":      recovered,
			"request_id": utils.GetRequestID(c),
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,

		
		utils.InternalError(c, "Internal server error occurred")
	})
}

// securityMiddleware adds security headers
func (s *Server) securityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		
		// Set content type for API endpoints
		if c.Request.URL.Path != "/" && c.Request.URL.Path != "/health" {
			c.Header(common.HeaderContentType, common.ContentTypeJSON)

		
		// Extract and store user information
		utils.GetUserIP(c)
		utils.GetUserAgent(c)

		
		c.Next()
	}
n
}