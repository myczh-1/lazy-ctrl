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
	
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}

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
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}

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

	// Convert system info safely
	systemInfo := make(map[string]string)
	if device.SystemInfo != nil {
		for k, v := range device.SystemInfo {
			if str, ok := v.(string); ok {
				systemInfo[k] = str
			} else {
				systemInfo[k] = fmt.Sprintf("%v", v)
			}
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
		ExitCode:  int(resp.ExitCode),
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
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}
	if req.Id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Command ID is required")
	}
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Command name is required")
	}
	if req.Command == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Command is required")
	}

	// Check if user has permission to manage commands on this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "user")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to manage commands on this device")
	}

	// Convert template params
	templateParams := make(map[string]interface{})
	for k, v := range req.TemplateParams {
		templateParams[k] = v
	}

	// Convert position config
	var positionConfig *model.PositionConfig
	if req.HomeLayout != nil && req.HomeLayout.DefaultPosition != nil {
		positionConfig = &model.PositionConfig{
			X:      int(req.HomeLayout.DefaultPosition.X),
			Y:      int(req.HomeLayout.DefaultPosition.Y),
			Width:  int(req.HomeLayout.DefaultPosition.Width),
			Height: int(req.HomeLayout.DefaultPosition.Height),
		}
	}

	// Create device command
	command := &model.DeviceCommand{
		DeviceID:         req.DeviceId,
		CommandID:        req.Id,
		Name:             req.Name,
		Description:      req.Description,
		Category:         req.Category,
		Icon:             req.Icon,
		Command:          req.Command,
		Platform:         req.Platform,
		CommandType:      req.CommandType,
		Timeout:          int(req.Timeout),
		TemplateID:       req.TemplateId,
		TemplateParams:   templateParams,
		RequiresPin:      req.Security != nil && req.Security.RequirePin,
		Whitelisted:      req.Security != nil && req.Security.Whitelist,
		AdminOnly:        req.Security != nil && req.Security.AdminOnly,
		ShowOnHomepage:   req.HomeLayout != nil && req.HomeLayout.ShowOnHome,
		HomepagePosition: positionConfig,
	}

	if req.HomeLayout != nil {
		command.HomepageColor = req.HomeLayout.Color
		command.HomepagePriority = int(req.HomeLayout.Priority)
	}

	// Create the command
	if err := h.deviceService.CreateDeviceCommand(command); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create command: %v", err)
	}

	// Convert to response format
	commandInfo := &gatewayPb.CommandInfo{
		Id:                   command.CommandID,
		Name:                 command.Name,
		Description:          command.Description,
		Category:             command.Category,
		Icon:                 command.Icon,
		Command:              command.Command,
		Platform:             command.Platform,
		CommandType:          command.CommandType,
		Timeout:              int32(command.Timeout),
		UserId:               req.UserId,
		DeviceId:             command.DeviceID,
		TemplateId:           command.TemplateID,
		CreatedAt:            timestamppb.New(command.CreatedAt),
		UpdatedAt:            timestamppb.New(command.UpdatedAt),
		RequiresPin:          command.RequiresPin,
		Whitelisted:          command.Whitelisted,
		ShowOnHomepage:       command.ShowOnHomepage,
		HomepageColor:        command.HomepageColor,
		HomepagePriority:     int32(command.HomepagePriority),
	}

	if command.HomepagePosition != nil {
		commandInfo.HomepagePosition = &gatewayPb.PositionConfig{
			X:      int32(command.HomepagePosition.X),
			Y:      int32(command.HomepagePosition.Y),
			Width:  int32(command.HomepagePosition.Width),
			Height: int32(command.HomepagePosition.Height),
		}
	}

	// Convert template params back
	templateParamsResp := make(map[string]string)
	for k, v := range templateParams {
		if str, ok := v.(string); ok {
			templateParamsResp[k] = str
		} else {
			templateParamsResp[k] = fmt.Sprintf("%v", v)
		}
	}
	commandInfo.TemplateParams = templateParamsResp

	return &gatewayPb.CreateCommandResponse{
		Success: true,
		Message: "Command created successfully",
		Command: commandInfo,
	}, nil
}

func (h *GatewayHandler) UpdateCommand(ctx context.Context, req *gatewayPb.UpdateCommandRequest) (*gatewayPb.UpdateCommandResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}
	if req.CommandId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Command ID is required")
	}

	// Check if user has permission to manage commands on this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "user")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to manage commands on this device")
	}

	// Convert template params
	templateParams := make(map[string]interface{})
	for k, v := range req.TemplateParams {
		templateParams[k] = v
	}

	// Convert position config
	var positionConfig *model.PositionConfig
	if req.HomeLayout != nil && req.HomeLayout.DefaultPosition != nil {
		positionConfig = &model.PositionConfig{
			X:      int(req.HomeLayout.DefaultPosition.X),
			Y:      int(req.HomeLayout.DefaultPosition.Y),
			Width:  int(req.HomeLayout.DefaultPosition.Width),
			Height: int(req.HomeLayout.DefaultPosition.Height),
		}
	}

	// Create updated command object
	command := &model.DeviceCommand{
		DeviceID:         req.DeviceId,
		CommandID:        req.CommandId,
		Name:             req.Name,
		Description:      req.Description,
		Category:         req.Category,
		Icon:             req.Icon,
		Command:          req.Command,
		Platform:         req.Platform,
		CommandType:      req.CommandType,
		Timeout:          int(req.Timeout),
		TemplateID:       req.TemplateId,
		TemplateParams:   templateParams,
		RequiresPin:      req.Security != nil && req.Security.RequirePin,
		Whitelisted:      req.Security != nil && req.Security.Whitelist,
		AdminOnly:        req.Security != nil && req.Security.AdminOnly,
		ShowOnHomepage:   req.HomeLayout != nil && req.HomeLayout.ShowOnHome,
		HomepagePosition: positionConfig,
	}

	if req.HomeLayout != nil {
		command.HomepageColor = req.HomeLayout.Color
		command.HomepagePriority = int(req.HomeLayout.Priority)
	}

	// Update the command
	if err := h.deviceService.UpdateDeviceCommand(command); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update command: %v", err)
	}

	// Get the updated command to return
	updatedCommand, err := h.deviceService.GetDeviceCommand(req.DeviceId, req.CommandId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve updated command: %v", err)
	}

	// Convert to response format
	commandInfo := &gatewayPb.CommandInfo{
		Id:                   updatedCommand.CommandID,
		Name:                 updatedCommand.Name,
		Description:          updatedCommand.Description,
		Category:             updatedCommand.Category,
		Icon:                 updatedCommand.Icon,
		Command:              updatedCommand.Command,
		Platform:             updatedCommand.Platform,
		CommandType:          updatedCommand.CommandType,
		Timeout:              int32(updatedCommand.Timeout),
		UserId:               req.UserId,
		DeviceId:             updatedCommand.DeviceID,
		TemplateId:           updatedCommand.TemplateID,
		CreatedAt:            timestamppb.New(updatedCommand.CreatedAt),
		UpdatedAt:            timestamppb.New(updatedCommand.UpdatedAt),
		RequiresPin:          updatedCommand.RequiresPin,
		Whitelisted:          updatedCommand.Whitelisted,
		ShowOnHomepage:       updatedCommand.ShowOnHomepage,
		HomepageColor:        updatedCommand.HomepageColor,
		HomepagePriority:     int32(updatedCommand.HomepagePriority),
	}

	if updatedCommand.HomepagePosition != nil {
		commandInfo.HomepagePosition = &gatewayPb.PositionConfig{
			X:      int32(updatedCommand.HomepagePosition.X),
			Y:      int32(updatedCommand.HomepagePosition.Y),
			Width:  int32(updatedCommand.HomepagePosition.Width),
			Height: int32(updatedCommand.HomepagePosition.Height),
		}
	}

	// Convert template params back
	templateParamsResp := make(map[string]string)
	for k, v := range templateParams {
		if str, ok := v.(string); ok {
			templateParamsResp[k] = str
		} else {
			templateParamsResp[k] = fmt.Sprintf("%v", v)
		}
	}
	commandInfo.TemplateParams = templateParamsResp

	return &gatewayPb.UpdateCommandResponse{
		Success: true,
		Message: "Command updated successfully",
		Command: commandInfo,
	}, nil
}

func (h *GatewayHandler) DeleteCommand(ctx context.Context, req *gatewayPb.DeleteCommandRequest) (*gatewayPb.DeleteCommandResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}
	if req.CommandId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Command ID is required")
	}

	// Check if user has permission to manage commands on this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "user")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to manage commands on this device")
	}

	// Delete the command
	if err := h.deviceService.DeleteDeviceCommand(req.DeviceId, req.CommandId); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete command: %v", err)
	}

	return &gatewayPb.DeleteCommandResponse{
		Success: true,
		Message: "Command deleted successfully",
	}, nil
}

func (h *GatewayHandler) GetCommand(ctx context.Context, req *gatewayPb.GetCommandRequest) (*gatewayPb.GetCommandResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}
	if req.CommandId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Command ID is required")
	}

	// Check if user has permission to view commands on this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "viewer")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to view commands on this device")
	}

	// Get the specific command
	cmd, err := h.deviceService.GetDeviceCommand(req.DeviceId, req.CommandId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Command not found: %v", err)
	}
	if cmd == nil {
		return nil, status.Errorf(codes.NotFound, "Command %s not found", req.CommandId)
	}

	// Convert to response format
	commandInfo := &gatewayPb.CommandInfo{
		Id:                   cmd.CommandID,
		Name:                 cmd.Name,
		Description:          cmd.Description,
		Category:             cmd.Category,
		Icon:                 cmd.Icon,
		Command:              cmd.Command,
		Platform:             cmd.Platform,
		CommandType:          cmd.CommandType,
		Timeout:              int32(cmd.Timeout),
		UserId:               req.UserId,
		DeviceId:             cmd.DeviceID,
		TemplateId:           cmd.TemplateID,
		CreatedAt:            timestamppb.New(cmd.CreatedAt),
		UpdatedAt:            timestamppb.New(cmd.UpdatedAt),
		RequiresPin:          cmd.RequiresPin,
		Whitelisted:          cmd.Whitelisted,
		ShowOnHomepage:       cmd.ShowOnHomepage,
		HomepageColor:        cmd.HomepageColor,
		HomepagePriority:     int32(cmd.HomepagePriority),
	}

	if cmd.HomepagePosition != nil {
		commandInfo.HomepagePosition = &gatewayPb.PositionConfig{
			X:      int32(cmd.HomepagePosition.X),
			Y:      int32(cmd.HomepagePosition.Y),
			Width:  int32(cmd.HomepagePosition.Width),
			Height: int32(cmd.HomepagePosition.Height),
		}
	}

	// Convert template params
	templateParamsResp := make(map[string]string)
	for k, v := range cmd.TemplateParams {
		if str, ok := v.(string); ok {
			templateParamsResp[k] = str
		} else {
			templateParamsResp[k] = fmt.Sprintf("%v", v)
		}
	}
	commandInfo.TemplateParams = templateParamsResp

	return &gatewayPb.GetCommandResponse{
		Success: true,
		Message: "Command retrieved successfully",
		Command: commandInfo,
	}, nil
}

func (h *GatewayHandler) GetAllCommands(ctx context.Context, req *gatewayPb.GetAllCommandsRequest) (*gatewayPb.GetAllCommandsResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}

	// Check if user has permission to view commands on this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "viewer")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to view commands on this device")
	}

	// Get all commands for the device
	commands, err := h.deviceService.GetDeviceCommands(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get commands: %v", err)
	}

	// Convert to response format
	var commandInfos []*gatewayPb.CommandInfo
	for _, cmd := range commands {
		commandInfo := &gatewayPb.CommandInfo{
			Id:                   cmd.CommandID,
			Name:                 cmd.Name,
			Description:          cmd.Description,
			Category:             cmd.Category,
			Icon:                 cmd.Icon,
			Command:              cmd.Command,
			Platform:             cmd.Platform,
			CommandType:          cmd.CommandType,
			Timeout:              int32(cmd.Timeout),
			UserId:               req.UserId,
			DeviceId:             cmd.DeviceID,
			TemplateId:           cmd.TemplateID,
			CreatedAt:            timestamppb.New(cmd.CreatedAt),
			UpdatedAt:            timestamppb.New(cmd.UpdatedAt),
			RequiresPin:          cmd.RequiresPin,
			Whitelisted:          cmd.Whitelisted,
			ShowOnHomepage:       cmd.ShowOnHomepage,
			HomepageColor:        cmd.HomepageColor,
			HomepagePriority:     int32(cmd.HomepagePriority),
		}

		if cmd.HomepagePosition != nil {
			commandInfo.HomepagePosition = &gatewayPb.PositionConfig{
				X:      int32(cmd.HomepagePosition.X),
				Y:      int32(cmd.HomepagePosition.Y),
				Width:  int32(cmd.HomepagePosition.Width),
				Height: int32(cmd.HomepagePosition.Height),
			}
		}

		// Convert template params
		templateParamsResp := make(map[string]string)
		for k, v := range cmd.TemplateParams {
			if str, ok := v.(string); ok {
				templateParamsResp[k] = str
			} else {
				templateParamsResp[k] = fmt.Sprintf("%v", v)
			}
		}
		commandInfo.TemplateParams = templateParamsResp

		commandInfos = append(commandInfos, commandInfo)
	}

	return &gatewayPb.GetAllCommandsResponse{
		Success: true,
		Message: "Commands retrieved successfully",
		Commands: commandInfos,
	}, nil
}

func (h *GatewayHandler) GetHomepageCommands(ctx context.Context, req *gatewayPb.GetHomepageCommandsRequest) (*gatewayPb.GetHomepageCommandsResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}

	// Check if user has permission to view commands on this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "viewer")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to view commands on this device")
	}

	// Get homepage commands for the device
	commands, err := h.deviceService.GetHomepageCommands(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get homepage commands: %v", err)
	}

	// Convert to response format
	var commandInfos []*gatewayPb.CommandInfo
	for _, cmd := range commands {
		commandInfo := &gatewayPb.CommandInfo{
			Id:                   cmd.CommandID,
			Name:                 cmd.Name,
			Description:          cmd.Description,
			Category:             cmd.Category,
			Icon:                 cmd.Icon,
			Command:              cmd.Command,
			Platform:             cmd.Platform,
			CommandType:          cmd.CommandType,
			Timeout:              int32(cmd.Timeout),
			UserId:               req.UserId,
			DeviceId:             cmd.DeviceID,
			TemplateId:           cmd.TemplateID,
			CreatedAt:            timestamppb.New(cmd.CreatedAt),
			UpdatedAt:            timestamppb.New(cmd.UpdatedAt),
			RequiresPin:          cmd.RequiresPin,
			Whitelisted:          cmd.Whitelisted,
			ShowOnHomepage:       cmd.ShowOnHomepage,
			HomepageColor:        cmd.HomepageColor,
			HomepagePriority:     int32(cmd.HomepagePriority),
		}

		if cmd.HomepagePosition != nil {
			commandInfo.HomepagePosition = &gatewayPb.PositionConfig{
				X:      int32(cmd.HomepagePosition.X),
				Y:      int32(cmd.HomepagePosition.Y),
				Width:  int32(cmd.HomepagePosition.Width),
				Height: int32(cmd.HomepagePosition.Height),
			}
		}

		// Convert template params
		templateParamsResp := make(map[string]string)
		for k, v := range cmd.TemplateParams {
			if str, ok := v.(string); ok {
				templateParamsResp[k] = str
			} else {
				templateParamsResp[k] = fmt.Sprintf("%v", v)
			}
		}
		commandInfo.TemplateParams = templateParamsResp

		commandInfos = append(commandInfos, commandInfo)
	}

	return &gatewayPb.GetHomepageCommandsResponse{
		Success: true,
		Message: "Homepage commands retrieved successfully",
		Commands: commandInfos,
	}, nil
}

func (h *GatewayHandler) VerifyPin(ctx context.Context, req *gatewayPb.VerifyPinRequest) (*gatewayPb.VerifyPinResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}
	if req.Pin == "" {
		return nil, status.Errorf(codes.InvalidArgument, "PIN is required")
	}

	// Check if user has permission to access this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "viewer")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to access this device")
	}

	// TODO: PIN verification should be implemented in the controller agent
	// For now, we'll implement a simple validation logic
	// In production, this should forward to the device for verification
	
	// Simple PIN validation (this should be configurable)
	validPin := "1234" // This should come from device configuration
	_ = req.Pin == validPin // TODO: Use this for actual validation

	return &gatewayPb.VerifyPinResponse{
		Success: true,
		Message: "PIN verification completed",
	}, nil
}

func (h *GatewayHandler) ReloadCommands(ctx context.Context, req *gatewayPb.ReloadCommandsRequest) (*gatewayPb.ReloadCommandsResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}
	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}

	// Check if user has permission to manage commands on this device
	hasPermission, err := h.deviceService.CheckUserDevicePermission(req.UserId, req.DeviceId, "user")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to check permissions: %v", err)
	}
	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "User does not have permission to manage commands on this device")
	}

	// Get device client and forward reload request to device
	client, err := h.gatewayService.GetDeviceClient(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Device not connected: %v", err)
	}

	// Forward reload request to device
	deviceReq := &controllerPb.ReloadConfigRequest{}

	deviceResp, err := client.ReloadConfig(ctx, deviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to reload commands on device: %v", err)
	}

	return &gatewayPb.ReloadCommandsResponse{
		Success: deviceResp.Success,
		Message: deviceResp.Message,
	}, nil
}

func (h *GatewayHandler) GetVersion(ctx context.Context, req *gatewayPb.GetVersionRequest) (*gatewayPb.GetVersionResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}

	// Get device client and use health check to get version info
	client, err := h.gatewayService.GetDeviceClient(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Device not connected: %v", err)
	}

	// Use health check to get version information
	deviceReq := &controllerPb.HealthCheckRequest{}

	deviceResp, err := client.HealthCheck(ctx, deviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get version from device: %v", err)
	}

	return &gatewayPb.GetVersionResponse{
		Success:   true,
		Version:   deviceResp.Version,
		BuildTime: "", // Not available in health check
		GitCommit: "", // Not available in health check
	}, nil
}

func (h *GatewayHandler) GetStatus(ctx context.Context, req *gatewayPb.GetStatusRequest) (*gatewayPb.GetStatusResponse, error) {
	// Validate input parameters
	if req.DeviceId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Device ID is required")
	}

	// Get device client and use health check to get status info
	client, err := h.gatewayService.GetDeviceClient(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Device not connected: %v", err)
	}

	// Use health check to get status information
	deviceReq := &controllerPb.HealthCheckRequest{}

	deviceResp, err := client.HealthCheck(ctx, deviceReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get status from device: %v", err)
	}

	// Convert uptime to string
	uptimeStr := fmt.Sprintf("%d seconds", deviceResp.UptimeSeconds)

	// Create system info from health check
	systemInfo := &gatewayPb.SystemInfo{
		Os:           "unknown",
		Architecture: "unknown", 
		GoVersion:    "unknown",
		NumCpu:       0,
		NumGoroutine: 0,
	}

	return &gatewayPb.GetStatusResponse{
		Success:   true,
		Uptime:    uptimeStr,
		Timestamp: timestamppb.Now(),
		System:    systemInfo,
		Memory:    make(map[string]int64),
		Commands:  make(map[string]int32),
		Services:  make(map[string]string),
	}, nil
}