package server

import (
	"context"
	"net"

	"lazy-ctrl/controller-agent/internal/service"
	pb "lazy-ctrl/controller-agent/api/proto/controller"

	"google.golang.org/grpc"
	"go.uber.org/zap"
)

type GRPCServer struct {
	server         *grpc.Server
	commandService *service.CommandService
	logger         *zap.Logger
}

func NewGRPCServer(commandService *service.CommandService) *GRPCServer {
	logger, _ := zap.NewProduction()
	
	return &GRPCServer{
		server:         grpc.NewServer(),
		commandService: commandService,
		logger:         logger,
	}
}

func (s *GRPCServer) ExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest) (*pb.ExecuteCommandResponse, error) {
	s.logger.Info("Executing command", zap.String("command_id", req.CommandId))
	
	result := s.commandService.ExecuteCommand(ctx, req.CommandId, req.Args)
	
	return &pb.ExecuteCommandResponse{
		Success:    result.Success,
		Output:     result.Output,
		Error:      result.Error,
		ExitCode:   int32(result.ExitCode),
		ExecutedAt: result.ExecutedAt,
	}, nil
}

func (s *GRPCServer) ListCommands(ctx context.Context, req *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	commands := s.commandService.ListCommands()
	
	var pbCommands []*pb.CommandInfo
	for _, cmd := range commands {
		pbCommands = append(pbCommands, &pb.CommandInfo{
			Id:          cmd.ID,
			Name:        cmd.Name,
			Description: cmd.Description,
		})
	}
	
	return &pb.ListCommandsResponse{
		Commands: pbCommands,
	}, nil
}

func (s *GRPCServer) Start(addr string) error {
	pb.RegisterControllerServiceServer(s.server, s)
	
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	
	s.logger.Info("Starting gRPC server", zap.String("address", addr))
	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop() {
	s.server.GracefulStop()
}