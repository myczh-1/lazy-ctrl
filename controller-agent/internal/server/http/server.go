package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/command"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/security"
	"github.com/sirupsen/logrus"
)

type Server struct {
	config          *config.Config
	logger          *logrus.Logger
	commandService  *command.Service
	executorService *executor.Service
	securityService *security.Service
	router          *gin.Engine
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
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// 添加日志中间件
	s.router.Use(s.loggingMiddleware())
	
	// 添加安全中间件
	s.router.Use(s.securityMiddleware())

	// API路由
	api := s.router.Group("/api/v1")
	{
		api.GET("/execute", s.handleExecute)
		api.GET("/commands", s.handleListCommands)
		api.POST("/reload", s.handleReloadConfig)
		api.GET("/health", s.handleHealth)
	}

	// 兼容旧版本路由
	s.router.GET("/execute", s.handleExecute)
}

func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(params gin.LogFormatterParams) string {
		s.logger.WithFields(logrus.Fields{
			"method":     params.Method,
			"path":       params.Path,
			"status":     params.StatusCode,
			"latency":    params.Latency,
			"client_ip":  params.ClientIP,
			"user_agent": params.Request.UserAgent(),
		}).Info("HTTP request")
		return ""
	})
}

func (s *Server) securityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成客户端ID用于限流
		clientID := s.securityService.GetClientID(c.ClientIP(), c.Request.UserAgent())
		
		// 检查限流
		if err := s.securityService.CheckRateLimit(clientID); err != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		// PIN验证（如果需要）
		if s.config.Security.PinRequired {
			pin := c.GetHeader("X-Pin")
			if pin == "" {
				pin = c.Query("pin")
			}
			
			if !s.securityService.ValidatePin(pin) {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid or missing PIN",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func (s *Server) handleExecute(c *gin.Context) {
	commandID := c.Query("id")
	if commandID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing command ID parameter",
		})
		return
	}

	// 验证命令访问权限
	if err := s.securityService.ValidateCommandAccess(commandID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 获取命令
	cmd, ok := s.commandService.GetCommand(commandID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Command not found: %s", commandID),
		})
		return
	}

	// 获取平台特定命令
	platformCmd, ok := s.commandService.GetPlatformCommand(cmd)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Command not supported on this platform",
		})
		return
	}

	// 执行命令
	var result *executor.ExecutionResult
	var err error

	timeoutStr := c.Query("timeout")
	if timeoutStr != "" {
		if timeout, parseErr := strconv.Atoi(timeoutStr); parseErr == nil {
			result, err = s.executorService.ExecuteWithTimeout(platformCmd, time.Duration(timeout)*time.Second)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid timeout parameter",
			})
			return
		}
	} else {
		result, err = s.executorService.ExecuteWithTimeout(platformCmd, 30*time.Second)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) handleListCommands(c *gin.Context) {
	commands := s.commandService.GetAllCommands()
	c.JSON(http.StatusOK, gin.H{
		"commands": commands,
	})
}

func (s *Server) handleReloadConfig(c *gin.Context) {
	if err := s.commandService.ReloadCommands(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration reloaded successfully",
	})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now().Unix(),
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.HTTP.Host, s.config.Server.HTTP.Port)
	s.logger.WithField("addr", addr).Info("Starting HTTP server")
	return s.router.Run(addr)
}