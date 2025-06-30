package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/server/http/handlers"
	"github.com/myczh-1/lazy-ctrl-agent/internal/server/http/middleware"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/command"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/security"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

type Server struct {
	config          *config.Config
	logger          *logrus.Logger
	commandService  *command.Service
	executorService *executor.Service
	securityService *security.Service
	router          *gin.Engine
	
	// Handlers
	executeHandler *handlers.ExecuteHandler
	commandHandler *handlers.CommandHandler
	systemHandler  *handlers.SystemHandler
}

func NewServer(
	config *config.Config,
	logger *logrus.Logger,
	commandService *command.Service,
	executorService *executor.Service,
	securityService *security.Service,
) *Server {
	if config.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	server := &Server{
		config:          config,
		logger:          logger,
		commandService:  commandService,
		executorService: executorService,
		securityService: securityService,
		router:          router,
		
		// Initialize handlers
		executeHandler: handlers.NewExecuteHandler(commandService, executorService, securityService),
		commandHandler: handlers.NewCommandHandler(commandService),
		systemHandler:  handlers.NewSystemHandler(),
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// 添加中间件
	s.router.Use(middleware.CORSMiddleware()) // 首先添加CORS支持
	s.router.Use(middleware.LoggingMiddleware(s.logger))
	s.router.Use(middleware.SecurityMiddleware(s.config, s.securityService))

	// 静态文件服务（如果配置了）
	if s.config.Server.HTTP.StaticPath != "" && s.config.Server.HTTP.StaticDir != "" {
		s.router.Static(s.config.Server.HTTP.StaticPath, s.config.Server.HTTP.StaticDir)
		s.logger.WithFields(logrus.Fields{
			"path": s.config.Server.HTTP.StaticPath,
			"dir":  s.config.Server.HTTP.StaticDir,
		}).Info("Static file server enabled")
	}

	// Swagger文档路由
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.logger.Info("Swagger documentation available at: /swagger/index.html")

	// API路由
	api := s.router.Group("/api/v1")
	{
		// 执行相关接口
		api.GET("/execute", s.executeHandler.HandleExecute)
		
		// 命令管理接口
		api.GET("/commands", s.commandHandler.HandleListCommands)
		api.POST("/commands", s.commandHandler.CreateCommand)
		api.PUT("/commands/:id", s.commandHandler.UpdateCommand)
		api.DELETE("/commands/:id", s.commandHandler.DeleteCommand)
		api.POST("/reload", s.commandHandler.HandleReloadConfig)
		
		// 系统接口
		api.GET("/health", s.systemHandler.HandleHealth)
	}

	// 兼容旧版本路由
	s.router.GET("/execute", s.executeHandler.HandleExecute)
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.HTTP.Host, s.config.Server.HTTP.Port)
	s.logger.WithField("addr", addr).Info("Starting HTTP server")
	return s.router.Run(addr)
}