package grpc

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/command"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/security"
	pb "github.com/myczh-1/lazy-ctrl-agent/proto"
	"github.com/sirupsen/logrus"
)

type Server struct {
	pb.UnimplementedControllerServiceServer
	config          *config.Config
	logger          *logrus.Logger
	commandService  *command.Service
	executorService *executor.Service
	securityService *security.Service
	grpcServer      *grpc.Server
	startTime       time.Time
}

func NewServer(
	config *config.Config,
	logger *logrus.Logger,
	commandService *command.Service,
	executorService *executor.Service,
	securityService *security.Service,
) *Server {
	server := &Server{
		config:          config,
		logger:          logger,
		commandService:  commandService,
		executorService: executorService,
		securityService: securityService,
		startTime:       time.Now(),
	}

	// 创建gRPC服务器，添加中间件
	server.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(server.authInterceptor),
	)

	pb.RegisterControllerServiceServer(server.grpcServer, server)
	return server
}

func (s *Server) authInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 记录请求日志
	peer, _ := peer.FromContext(ctx)
	s.logger.WithFields(logrus.Fields{
		"method":     info.FullMethod,
		"remote_addr": peer.Addr.String(),
	}).Info("gRPC request received")

	// 检查限流
	clientID := s.securityService.GetClientID(peer.Addr.String(), "grpc")
	if err := s.securityService.CheckRateLimit(clientID); err != nil {
		s.logger.WithField("client_id", clientID).Warn("Rate limit exceeded")
		return nil, status.Error(codes.ResourceExhausted, err.Error())
	}

	// PIN验证
	if s.config.Security.PinRequired {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "Missing metadata")
		}

		pins := md.Get("pin")
		if len(pins) == 0 {
			return nil, status.Error(codes.Unauthenticated, "Missing PIN")
		}

		if !s.securityService.ValidatePin(pins[0]) {
			return nil, status.Error(codes.Unauthenticated, "Invalid PIN")
		}
	}

	return handler(ctx, req)
}

func (s *Server) ExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest) (*pb.ExecuteCommandResponse, error) {
	startTime := time.Now()
	
	s.logger.WithField("command_id", req.CommandId).Info("Executing command via gRPC")

	// 验证命令访问权限
	if err := s.securityService.ValidateCommandAccess(req.CommandId); err != nil {
		return &pb.ExecuteCommandResponse{
			Success:         false,
			Error:          err.Error(),
			ExitCode:       1,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 获取命令
	cmd, ok := s.commandService.GetCommand(req.CommandId)
	if !ok {
		return &pb.ExecuteCommandResponse{
			Success:         false,
			Error:          fmt.Sprintf("Command not found: %s", req.CommandId),
			ExitCode:       1,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 获取平台特定命令
	platformCmd, ok := s.commandService.GetPlatformCommand(cmd)
	if !ok {
		return &pb.ExecuteCommandResponse{
			Success:         false,
			Error:          fmt.Sprintf("Command not supported on platform: %s", runtime.GOOS),
			ExitCode:       1,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	// 设置超时
	if req.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	// 执行命令
	result, err := s.executorService.Execute(ctx, platformCmd)
	if err != nil {
		return &pb.ExecuteCommandResponse{
			Success:         false,
			Error:          err.Error(),
			ExitCode:       1,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, nil
	}

	return &pb.ExecuteCommandResponse{
		Success:         result.Success,
		Output:         result.Output,
		Error:          result.Error,
		ExitCode:       int32(result.ExitCode),
		ExecutionTimeMs: result.ExecutionTime.Milliseconds(),
	}, nil
}

func (s *Server) ListCommands(ctx context.Context, req *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	commands := s.commandService.GetAllCommands()
	var commandInfos []*pb.CommandInfo

	for _, cmd := range commands {
		platformCmd, supported := s.commandService.GetPlatformCommand(&cmd)
		
		commandInfos = append(commandInfos, &pb.CommandInfo{
			Id:               cmd.ID,
			Description:      fmt.Sprintf("Command: %s", cmd.ID),
			PlatformSupported: supported,
			PlatformCommand:  platformCmd,
		})
	}

	return &pb.ListCommandsResponse{
		Commands: commandInfos,
	}, nil
}

func (s *Server) ReloadConfig(ctx context.Context, req *pb.ReloadConfigRequest) (*pb.ReloadConfigResponse, error) {
	err := s.commandService.ReloadCommands()
	if err != nil {
		return &pb.ReloadConfigResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to reload config: %v", err),
		}, nil
	}

	commandCount := len(s.commandService.GetAllCommands())

	return &pb.ReloadConfigResponse{
		Success:        true,
		Message:        "Configuration reloaded successfully",
		CommandsLoaded: int32(commandCount),
	}, nil
}

func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	uptime := time.Since(s.startTime).Seconds()

	return &pb.HealthCheckResponse{
		Status:        "SERVING",
		Version:       "2.0.0",
		UptimeSeconds: int64(uptime),
	}, nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.GRPC.Host, s.config.Server.GRPC.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.logger.WithField("addr", addr).Info("Starting gRPC server")
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server")
	s.grpcServer.GracefulStop()
}