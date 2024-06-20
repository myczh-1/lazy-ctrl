package grpcserver

import (
	"context"
	"fmt"
	"runtime"
	"time"

	pb "github.com/myczh-1/lazy-ctrl-agent/proto"
	"github.com/myczh-1/lazy-ctrl-agent/internal/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/handler"
)

type ControllerServer struct {
	pb.UnimplementedControllerServiceServer
	startTime time.Time
	version   string
}

func NewControllerServer() *ControllerServer {
	return &ControllerServer{
		startTime: time.Now(),
		version:   "1.0.0",
	}
}

func (s *ControllerServer) ExecuteCommand(ctx context.Context, req *pb.ExecuteCommandRequest) (*pb.ExecuteCommandResponse, error) {
	startTime := time.Now()
	
	fmt.Printf("gRPC received command ID: %s\n", req.CommandId)
	
	// 获取命令
	cmd, ok := handler.GetCommand(req.CommandId)
	if !ok {
		return &pb.ExecuteCommandResponse{
			Success:         false,
			Error:          fmt.Sprintf("Command not found: %s", req.CommandId),
			ExitCode:       1,
			ExecutionTimeMs: time.Since(startTime).Milliseconds(),
		}, nil
	}
	
	// 获取平台特定命令
	platformCmd, ok := handler.GetCommandForPlatform(cmd)
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
	output, err := executor.RunCommandWithContext(ctx, platformCmd)
	success := err == nil
	exitCode := int32(0)
	errorMsg := ""
	
	if err != nil {
		errorMsg = err.Error()
		exitCode = 1
	}
	
	return &pb.ExecuteCommandResponse{
		Success:         success,
		Output:         output,
		Error:          errorMsg,
		ExitCode:       exitCode,
		ExecutionTimeMs: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *ControllerServer) ListCommands(ctx context.Context, req *pb.ListCommandsRequest) (*pb.ListCommandsResponse, error) {
	commands := handler.GetAllCommands()
	var commandInfos []*pb.CommandInfo
	
	for id, cmd := range commands {
		platformCmd, supported := handler.GetCommandForPlatform(cmd)
		
		commandInfos = append(commandInfos, &pb.CommandInfo{
			Id:               id,
			Description:      fmt.Sprintf("Command: %s", id),
			PlatformSupported: supported,
			PlatformCommand:  platformCmd,
		})
	}
	
	return &pb.ListCommandsResponse{
		Commands: commandInfos,
	}, nil
}

func (s *ControllerServer) ReloadConfig(ctx context.Context, req *pb.ReloadConfigRequest) (*pb.ReloadConfigResponse, error) {
	err := handler.LoadCommands("config/commands.json")
	if err != nil {
		return &pb.ReloadConfigResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to reload config: %v", err),
		}, nil
	}
	
	commandCount := len(handler.GetAllCommands())
	
	return &pb.ReloadConfigResponse{
		Success:        true,
		Message:        "Configuration reloaded successfully",
		CommandsLoaded: int32(commandCount),
	}, nil
}

func (s *ControllerServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	uptime := time.Since(s.startTime).Seconds()
	
	return &pb.HealthCheckResponse{
		Status:        "SERVING",
		Version:       s.version,
		UptimeSeconds: int64(uptime),
	}, nil
}