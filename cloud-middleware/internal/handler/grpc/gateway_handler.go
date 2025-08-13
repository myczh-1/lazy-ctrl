package grpc

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	gatewayPb "github.com/myczh-1/lazy-ctrl-cloud/proto"
	controllerPb "github.com/myczh-1/lazy-ctrl-agent/proto"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/service"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/model"
)

// GatewayHandler implements the GatewayService gRPC server
type GatewayHandler struct {
	gatewayPb.UnimplementedGatewayServiceServer
	gatewayService *service.GatewayService
	deviceService  *service.DeviceService
}

// NewGatewayHandler creates a new gateway gRPC handler
func NewGatewayHandler(gatewayService *service.GatewayService, deviceService *service.DeviceService) *GatewayHandler {
	return &GatewayHandler{
		gatewayService: gatewayService,
		deviceService:  deviceService,
	}
}

// RegisterDevice registers a new device with the cloud gateway
func (h *GatewayHandler) RegisterDevice(ctx context.Context, req *gatewayPb.RegisterDeviceRequest) (*gatewayPb.RegisterDeviceResponse, error) {
	log.Printf("RegisterDevice called for device %s by user %s", req.DeviceId, req.UserId)

	// Convert metadata map
	systemInfo := make(map[string]interface{})
	for k, v := range req.Metadata {
		systemInfo[k] = v
	}

	// Register device in database
	device, err := h.deviceService.RegisterDevice(
		req.UserId,
		req.DeviceId,
		req.DeviceName,
		"desktop", // default device type
		req.Platform,
		req.Version,
		systemInfo,
	)
	if err != nil {
		log.Printf("Failed to register device %s: %v", req.DeviceId, err)
		return &gatewayPb.RegisterDeviceResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to register device: %v", err),
		}, nil
	}

	// Generate access token (simplified - in production use proper JWT)
	accessToken := fmt.Sprintf("device_token_%s_%d", req.DeviceId, device.CreatedAt.Unix())

	return &gatewayPb.RegisterDeviceResponse{
		Success:     true,
		Message:     "Device registered successfully",
		AccessToken: accessToken,
	}, nil
}

// GetDeviceStatus returns the status of a specific device
func (h *GatewayHandler) GetDeviceStatus(ctx context.Context, req *gatewayPb.GetDeviceStatusRequest) (*gatewayPb.GetDeviceStatusResponse, error) {
	// Check if user has permission to access this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "viewer")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to access this device")
	}

	// Get device from database
	device, err := h.deviceService.GetDeviceByID(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Device not found: %v", err)
	}

	// Convert system info
	systemInfo := make(map[string]string)
	for k, v := range device.SystemInfo {
		if str, ok := v.(string); ok {
			systemInfo[k] = str
		} else {
			systemInfo[k] = fmt.Sprintf("%v", v)
		}
	}

	deviceStatus := &gatewayPb.DeviceStatus{
		DeviceId:   device.ID,
		DeviceName: device.DeviceName,
		Platform:   device.Platform,
		Online:     device.Online,
		LastSeen:   timestamppb.New(device.LastSeen),
		Version:    device.AgentVersion,
		SystemInfo: systemInfo,
	}

	return &gatewayPb.GetDeviceStatusResponse{
		Success: true,
		Message: "Device status retrieved successfully",
		Device:  deviceStatus,
	}, nil
}

// ListUserDevices returns all devices associated with a user
func (h *GatewayHandler) ListUserDevices(ctx context.Context, req *gatewayPb.ListUserDevicesRequest) (*gatewayPb.ListUserDevicesResponse, error) {
	devices, err := h.deviceService.GetUserDevices(req.UserId, false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get user devices: %v", err)
	}

	var deviceStatuses []*gatewayPb.DeviceStatus
	for _, device := range devices {
		// Convert system info
		systemInfo := make(map[string]string)
		for k, v := range device.SystemInfo {
			if str, ok := v.(string); ok {
				systemInfo[k] = str
			} else {
				systemInfo[k] = fmt.Sprintf("%v", v)
			}
		}

		deviceStatus := &gatewayPb.DeviceStatus{
			DeviceId:   device.ID,
			DeviceName: device.DeviceName,
			Platform:   device.Platform,
			Online:     device.Online,
			LastSeen:   timestamppb.New(device.LastSeen),
			Version:    device.AgentVersion,
			SystemInfo: systemInfo,
		}
		deviceStatuses = append(deviceStatuses, deviceStatus)
	}

	return &gatewayPb.ListUserDevicesResponse{
		Success: true,
		Message: "User devices retrieved successfully",
		Devices: deviceStatuses,
	}, nil
}

// ExecuteCommand executes a command on a remote device
func (h *GatewayHandler) ExecuteCommand(ctx context.Context, req *gatewayPb.ExecuteCommandRequest) (*gatewayPb.ExecuteCommandResponse, error) {
	// Check if user has permission to execute commands on this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "user")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to execute commands on this device")
	}

	// Execute command through gateway service
	resp, err := h.gatewayService.ExecuteCommand(req.DeviceId, req.CommandId, 30) // 30 second timeout
	if err != nil {
		log.Printf("Failed to execute command %s on device %s: %v", req.CommandId, req.DeviceId, err)
		return &gatewayPb.ExecuteCommandResponse{
			Success:  false,
			Error:    fmt.Sprintf("Failed to execute command: %v", err),
			ExitCode: -1,
		}, nil
	}

	// Log execution
	executionLog := &model.ExecutionLog{
		UserID:    req.UserId,
		DeviceID:  req.DeviceId,
		CommandID: req.CommandId,
		Success:   resp.Success,
		Output:    resp.Output,
		Error:     resp.Error,
		ExitCode:  resp.ExitCode,
		Duration:  resp.ExecutionTimeMs,
	}

	// Save execution log (ignore errors for now)
	// TODO: Implement execution log repository and save
	_ = executionLog

	return &gatewayPb.ExecuteCommandResponse{
		Success:  resp.Success,
		Output:   resp.Output,
		Error:    resp.Error,
		ExitCode: resp.ExitCode,
		Duration: resp.ExecutionTimeMs,
	}, nil
}

// HealthCheck performs a health check on a device
func (h *GatewayHandler) HealthCheck(ctx context.Context, req *gatewayPb.HealthCheckRequest) (*gatewayPb.HealthCheckResponse, error) {
	// Get device client and perform health check
	client, err := h.gatewayService.GetDeviceClient(req.DeviceId)
	if err != nil {
		return &gatewayPb.HealthCheckResponse{
			Success: false,
			Status:  "NOT_SERVING",
		}, nil
	}

	// Forward health check to device
	resp, err := client.HealthCheck(ctx, &controllerPb.HealthCheckRequest{})
	if err != nil {
		return &gatewayPb.HealthCheckResponse{
			Success: false,
			Status:  "NOT_SERVING",
		}, nil
	}

	// Update device status in database
	_ = h.deviceService.UpdateDeviceLastSeen(req.DeviceId)

	return &gatewayPb.HealthCheckResponse{
		Success:   true,
		Status:    resp.Status,
		Timestamp: timestamppb.Now(),
		Version:   resp.Version,
		System: &gatewayPb.SystemInfo{
			Os:           "unknown", // These would come from the actual response
			Architecture: "unknown",
			GoVersion:    "unknown",
			NumCpu:       0,
			NumGoroutine: 0,
		},
		Services: make(map[string]string),
	}, nil
}

// GetCommandInfo gets information about a specific command
func (h *GatewayHandler) GetCommandInfo(ctx context.Context, req *gatewayPb.GetCommandInfoRequest) (*gatewayPb.GetCommandInfoResponse, error) {
	// Check permissions
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "viewer")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to access this device")
	}

	// Get command list from device
	resp, err := h.gatewayService.ListCommands(req.DeviceId)
	if err != nil {
		return &gatewayPb.GetCommandInfoResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to get command info: %v", err),
		}, nil
	}

	// Find the specific command
	var commandInfo map[string]string
	for _, cmd := range resp.Commands {
		if cmd.Id == req.CommandId {
			commandInfo = map[string]string{
				"id":                 cmd.Id,
				"description":        cmd.Description,
				"platform_supported": fmt.Sprintf("%v", cmd.PlatformSupported),
				"platform_command":   cmd.PlatformCommand,
			}
			break
		}
	}

	if commandInfo == nil {
		return &gatewayPb.GetCommandInfoResponse{
			Success: false,
			Message: "Command not found",
		}, nil
	}

	return &gatewayPb.GetCommandInfoResponse{
		Success: true,
		Message: "Command info retrieved successfully",
		Info:    commandInfo,
	}, nil
}

// Placeholder implementations for other methods to satisfy the interface

func (h *GatewayHandler) CreateCommand(ctx context.Context, req *gatewayPb.CreateCommandRequest) (*gatewayPb.CreateCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "CreateCommand not implemented yet")
}

func (h *GatewayHandler) UpdateCommand(ctx context.Context, req *gatewayPb.UpdateCommandRequest) (*gatewayPb.UpdateCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "UpdateCommand not implemented yet")
}

func (h *GatewayHandler) DeleteCommand(ctx context.Context, req *gatewayPb.DeleteCommandRequest) (*gatewayPb.DeleteCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "DeleteCommand not implemented yet")
}

func (h *GatewayHandler) GetCommand(ctx context.Context, req *gatewayPb.GetCommandRequest) (*gatewayPb.GetCommandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetCommand not implemented yet")
}

func (h *GatewayHandler) GetAllCommands(ctx context.Context, req *gatewayPb.GetAllCommandsRequest) (*gatewayPb.GetAllCommandsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetAllCommands not implemented yet")
}

func (h *GatewayHandler) GetHomepageCommands(ctx context.Context, req *gatewayPb.GetHomepageCommandsRequest) (*gatewayPb.GetHomepageCommandsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetHomepageCommands not implemented yet")
}

func (h *GatewayHandler) VerifyPin(ctx context.Context, req *gatewayPb.VerifyPinRequest) (*gatewayPb.VerifyPinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "VerifyPin not implemented yet")
}

func (h *GatewayHandler) ReloadCommands(ctx context.Context, req *gatewayPb.ReloadCommandsRequest) (*gatewayPb.ReloadCommandsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "ReloadCommands not implemented yet")
}

func (h *GatewayHandler) GetVersion(ctx context.Context, req *gatewayPb.GetVersionRequest) (*gatewayPb.GetVersionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetVersion not implemented yet")
}

func (h *GatewayHandler) GetStatus(ctx context.Context, req *gatewayPb.GetStatusRequest) (*gatewayPb.GetStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "GetStatus not implemented yet")
}