package app

import (
	"context"
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/myczh-1/lazy-ctrl-cloud/internal/config"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/database"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/handler/http"
	grpchandler "github.com/myczh-1/lazy-ctrl-cloud/internal/handler/grpc"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/middleware"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/repository"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/service"
	gatewayPb "github.com/myczh-1/lazy-ctrl-cloud/proto"
)

// Application represents the main application
type Application struct {
	config     *config.Config
	db         *gorm.DB
	grpcServer *grpc.Server
	
	// Services
	userService    service.UserService
	deviceService  *service.DeviceService
	gatewayService *service.GatewayService
	
	// HTTP handlers
	userHandler    *http.UserHandler
	gatewayHandler *http.GatewayHandler
	
	// gRPC handlers
	grpcGatewayHandler *grpchandler.GatewayHandler
}

// NewApplication creates a new application instance
func NewApplication(cfg *config.Config) (*Application, error) {
	app := &Application{
		config: cfg,
	}
	
	// Initialize database
	if err := app.initDatabase(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	// Initialize services
	if err := app.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}
	
	// Initialize handlers
	if err := app.initHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize handlers: %w", err)
	}
	
	return app, nil
}

// initDatabase initializes the database connection
func (a *Application) initDatabase() error {
	db, err := database.Connect(a.config.Database)
	if err != nil {
		return err
	}
	
	a.db = db
	return nil
}

// initServices initializes all services
func (a *Application) initServices() error {
	// Initialize repositories
	userRepo := repository.NewUserRepository(a.db)
	deviceRepo := repository.NewDeviceRepository(a.db)
	
	// Initialize services
	a.userService = service.NewUserService(userRepo, a.config.JWT)
	a.deviceService = service.NewDeviceService(deviceRepo)
	a.gatewayService = service.NewGatewayService()
	
	// Initialize default admin user
	if err := a.userService.InitializeSystem(); err != nil {
		return fmt.Errorf("failed to initialize system: %w", err)
	}
	
	return nil
}

// initHandlers initializes all handlers
func (a *Application) initHandlers() error {
	// HTTP handlers
	a.userHandler = http.NewUserHandler(a.userService)
	a.gatewayHandler = http.NewGatewayHandler(a.gatewayService, a.deviceService)
	
	// gRPC handlers
	a.grpcGatewayHandler = grpchandler.NewGatewayHandler(a.gatewayService, a.deviceService)
	
	return nil
}

// Router returns the HTTP router
func (a *Application) Router() *gin.Engine {
	router := gin.New()
	
	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	
	// API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", a.userHandler.Login)
			auth.POST("/refresh", a.userHandler.RefreshToken)
			auth.POST("/logout", middleware.AuthRequired(), a.userHandler.Logout)
		}
		
		// User profile routes
		user := v1.Group("/user", middleware.AuthRequired())
		{
			user.GET("/profile", a.userHandler.GetProfile)
			user.PUT("/profile", a.userHandler.UpdateProfile)
			user.POST("/change-password", a.userHandler.ChangePassword)
		}
		
		// Admin routes for user management
		admin := v1.Group("/admin", middleware.AuthRequired())
		{
			admin.POST("/users", a.userHandler.CreateUser)
			admin.GET("/users", a.userHandler.ListUsers)
			admin.GET("/users/:user_id", a.userHandler.GetUser)
			admin.PUT("/users/:user_id", a.userHandler.UpdateUser)
			admin.DELETE("/users/:user_id", a.userHandler.DeleteUser)
		}
		
		// Device routes (placeholder)
		// device := v1.Group("/device", middleware.AuthRequired())
		// {
		//     device.POST("/bind", a.deviceHandler.BindDevice)
		//     device.DELETE("/:device_id", a.deviceHandler.UnbindDevice)
		//     device.GET("/list", a.deviceHandler.GetUserDevices)
		//     device.PUT("/:device_id", a.deviceHandler.UpdateDeviceInfo)
		// }
		
		// Gateway routes
		gateway := v1.Group("/gateway", middleware.AuthRequired())
		{
			// Command execution
			gateway.POST("/execute", a.gatewayHandler.ExecuteCommand)
			gateway.GET("/commands", a.gatewayHandler.ListCommands)
			
			// Device management
			gateway.POST("/devices/connect", a.gatewayHandler.ConnectDevice)
			gateway.DELETE("/devices/:device_id/disconnect", a.gatewayHandler.DisconnectDevice)
			gateway.GET("/devices", a.gatewayHandler.ListConnectedDevices)
			gateway.GET("/devices/:device_id/status", a.gatewayHandler.GetDeviceStatus)
			gateway.GET("/devices/:device_id/health", a.gatewayHandler.HealthCheck)
			gateway.POST("/devices/:device_id/reload", a.gatewayHandler.ReloadConfig)
		}
	}
	
	return router
}

// StartGRPCServer starts the gRPC server
func (a *Application) StartGRPCServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", a.config.GRPC.Port, err)
	}
	
	a.grpcServer = grpc.NewServer()
	
	// Register gRPC services
	gatewayPb.RegisterGatewayServiceServer(a.grpcServer, a.grpcGatewayHandler)
	// userPb.RegisterUserServiceServer(a.grpcServer, a.grpcUserHandler)
	
	return a.grpcServer.Serve(lis)
}

// Stop gracefully stops the application
func (a *Application) Stop(ctx context.Context) error {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	
	// Stop gateway service
	if a.gatewayService != nil {
		a.gatewayService.Stop()
	}
	
	// Close database connection
	if a.db != nil {
		sqlDB, err := a.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
	
	return nil
}