package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/sirupsen/logrus"
	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/service"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/security"
	pb "github.com/myczh-1/lazy-ctrl-agent/proto"
)

// Server represents the gRPC server
type Server struct {
	pb.UnimplementedControllerServiceServer
	config          *config.Config
	logger          *logrus.Logger
	commandService  *service.CommandService
	executorService *executor.Service
	securityService *security.Service
	grpcServer      *grpc.Server
	startTime       time.Time
}

// NewServer creates a new gRPC server instance
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
		startTime:       time.Now(),
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.Server.GRPC.Host, s.config.Server.GRPC.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
	)

	pb.RegisterControllerServiceServer(s.grpcServer, s)

	s.logger.WithField("addr", listen.Addr().String()).Info("Starting gRPC server")

	if err := s.grpcServer.Serve(listen); err != nil {
		return fmt.Errorf("gRPC server failed: %w", err)
	}

	return nil
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server")
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// unaryInterceptor provides common middleware for all gRPC calls
func (s *Server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	// Extract client info
	peer, _ := peer.FromContext(ctx)
	clientIP := "unknown"
	if peer != nil {
		clientIP = peer.Addr.String()
	}

	// Rate limiting
	if err := s.securityService.CheckRateLimit(clientIP); err != nil {
		s.logger.WithFields(logrus.Fields{
			"method":    info.FullMethod,
			"client_ip": clientIP,
			"error":     err.Error(),
		}).Warn("gRPC rate limit exceeded")
		return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
	}

	// Call the handler
	resp, err := handler(ctx, req)

	// Log the request
	s.logger.WithFields(logrus.Fields{
		"method":    info.FullMethod,
		"client_ip": clientIP,
		"duration":  time.Since(start),
		"error":     err,
	}).Info("gRPC request completed")

	return resp, err
}

// ExecuteCommand executes a command via gRPC
func (s *Server) ExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest) (*pb.ExecuteCommandResponse, error) {
	if req.CommandId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "command ID is required")
	}

	// Get command
	cmd, err := s.commandService.GetCommand(ctx, req.CommandId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "command not found: %s", req.CommandId)
	}

	// Get platform command
	platformCommand, err := s.commandService.GetPlatformCommand(ctx, req.CommandId)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "command not available: %s", err.Error())
	}

	// Execute with timeout
	timeout := time.Duration(cmd.GetTimeout()) * time.Millisecond
	if req.TimeoutSeconds > 0 {
		timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}
	
	executeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()
	result, err := s.executorService.Execute(executeCtx, platformCommand)
	executionTime := time.Since(startTime)
	
	if err != nil {
		return &pb.ExecuteCommandResponse{
			Success:         false,
			Output:          "",
			Error:           err.Error(),
			ExitCode:        -1,
			ExecutionTimeMs: executionTime.Milliseconds(),
		}, nil
	}

	return &pb.ExecuteCommandResponse{
		Success:         result.Success,
		Output:          result.Output,
		Error:           result.Error,
		ExitCode:        int32(result.ExitCode),
		ExecutionTimeMs: executionTime.Milliseconds(),
	}, nil
}

// ListCommands returns all available commands
func (s *Server) ListCommands(ctx context.Context, req *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	commands, err := s.commandService.GetAllCommands(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get commands: %s", err.Error())
	}

	pbCommands := make([]*pb.CommandInfo, len(commands))
	for i, cmd := range commands {
		pbCommands[i] = &pb.CommandInfo{
			Id:                cmd.ID,
			Description:       cmd.Description,
			PlatformSupported: cmd.IsAvailableOnPlatform(),
			PlatformCommand:   cmd.Command,
		}
	}

	return &pb.ListCommandsResponse{
		Commands: pbCommands,
	}, nil
}

// ReloadConfig reloads the configuration
func (s *Server) ReloadConfig(ctx context.Context, req *pb.ReloadConfigRequest) (*pb.ReloadConfigResponse, error) {
	err := s.commandService.ReloadCommands(ctx)
	if err != nil {
		return &pb.ReloadConfigResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	commands, _ := s.commandService.GetAllCommands(ctx)
	return &pb.ReloadConfigResponse{
		Success:        true,
		Message:        "Configuration reloaded successfully",
		CommandsLoaded: int32(len(commands)),
	}, nil
}

// HealthCheck performs health check
func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{
		Status:        "SERVING",
		Version:       "2.0.0",
		UptimeSeconds: int64(time.Since(s.startTime).Seconds()),
	}, nil
}