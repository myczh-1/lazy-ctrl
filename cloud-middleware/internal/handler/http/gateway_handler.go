package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/myczh-1/lazy-ctrl-cloud/internal/service"
	controllerPb "github.com/myczh-1/lazy-ctrl-agent/proto"
)

// GatewayHandler handles HTTP requests for the gateway service
type GatewayHandler struct {
	gatewayService *service.GatewayService
	deviceService  *service.DeviceService
}

// NewGatewayHandler creates a new gateway handler
func NewGatewayHandler(gatewayService *service.GatewayService, deviceService *service.DeviceService) *GatewayHandler {
	return &GatewayHandler{
		gatewayService: gatewayService,
		deviceService:  deviceService,
	}
}

// ExecuteCommandRequest represents the request body for command execution
type ExecuteCommandRequest struct {
	DeviceID  string `json:"device_id" binding:"required"`
	CommandID string `json:"command_id" binding:"required"`
	Timeout   int32  `json:"timeout,omitempty"` // optional timeout in seconds
}

// ExecuteCommandResponse represents the response for command execution
type ExecuteCommandResponse struct {
	Success         bool   `json:"success"`
	Output          string `json:"output,omitempty"`
	Error           string `json:"error,omitempty"`
	ExitCode        int32  `json:"exit_code"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
}

// DeviceConnectionRequest represents the request to connect a device
type DeviceConnectionRequest struct {
	DeviceID string `json:"device_id" binding:"required"`
	Address  string `json:"address" binding:"required"` // IP:Port
}

// DeviceListResponse represents the list of connected devices
type DeviceListResponse struct {
	Devices []string `json:"devices"`
}

// DeviceStatusResponse represents device status information
type DeviceStatusResponse struct {
	DeviceID    string    `json:"device_id"`
	Address     string    `json:"address"`
	IsHealthy   bool      `json:"is_healthy"`
	LastPing    time.Time `json:"last_ping"`
	ConnectedAt time.Time `json:"connected_at"`
}

// CommandInfo represents command information from device
type CommandInfo struct {
	ID                string `json:"id"`
	Description       string `json:"description"`
	PlatformSupported bool   `json:"platform_supported"`
	PlatformCommand   string `json:"platform_command"`
}

// CommandListResponse represents the list of commands from device
type CommandListResponse struct {
	Commands []CommandInfo `json:"commands"`
}

// ExecuteCommand executes a command on a remote device
// @Summary Execute command on device
// @Description Execute a command on a remote device through gRPC
// @Tags Gateway
// @Accept json
// @Produce json
// @Param request body ExecuteCommandRequest true "Command execution request"
// @Success 200 {object} ExecuteCommandResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/gateway/execute [post]
func (h *GatewayHandler) ExecuteCommand(c *gin.Context) {
	var req ExecuteCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default timeout if not specified
	if req.Timeout == 0 {
		req.Timeout = 30 // 30 seconds default
	}

	// Execute command through gateway service
	resp, err := h.gatewayService.ExecuteCommand(req.DeviceID, req.CommandID, req.Timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := ExecuteCommandResponse{
		Success:         resp.Success,
		Output:          resp.Output,
		Error:           resp.Error,
		ExitCode:        resp.ExitCode,
		ExecutionTimeMs: resp.ExecutionTimeMs,
	}

	c.JSON(http.StatusOK, response)
}

// ListCommands lists all available commands on a device
// @Summary List device commands
// @Description Get list of all commands available on a device
// @Tags Gateway
// @Accept json
// @Produce json
// @Param device_id query string true "Device ID"
// @Success 200 {object} CommandListResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/gateway/commands [get]
func (h *GatewayHandler) ListCommands(c *gin.Context) {
	deviceID := c.Query("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	resp, err := h.gatewayService.ListCommands(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert proto response to HTTP response
	commands := make([]CommandInfo, len(resp.Commands))
	for i, cmd := range resp.Commands {
		commands[i] = CommandInfo{
			ID:                cmd.Id,
			Description:       cmd.Description,
			PlatformSupported: cmd.PlatformSupported,
			PlatformCommand:   cmd.PlatformCommand,
		}
	}

	response := CommandListResponse{
		Commands: commands,
	}

	c.JSON(http.StatusOK, response)
}

// ReloadConfig reloads configuration on a device
// @Summary Reload device configuration
// @Description Reload configuration and commands on a device
// @Tags Gateway
// @Accept json
// @Produce json
// @Param device_id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/gateway/devices/{device_id}/reload [post]
func (h *GatewayHandler) ReloadConfig(c *gin.Context) {
	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	resp, err := h.gatewayService.ReloadConfig(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         resp.Success,
		"message":         resp.Message,
		"commands_loaded": resp.CommandsLoaded,
	})
}

// HealthCheck performs health check on a device
// @Summary Device health check
// @Description Check if a device is healthy and responding
// @Tags Gateway
// @Accept json
// @Produce json
// @Param device_id path string true "Device ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/gateway/devices/{device_id}/health [get]
func (h *GatewayHandler) HealthCheck(c *gin.Context) {
	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	// Get client and perform health check
	client, err := h.gatewayService.GetDeviceClient(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Perform health check via gRPC
	req := &controllerPb.HealthCheckRequest{}
	resp, err := client.HealthCheck(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         resp.Status,
		"version":        resp.Version,
		"uptime_seconds": resp.UptimeSeconds,
	})
}

// ConnectDevice establishes a connection to a device
// @Summary Connect to device
// @Description Establish gRPC connection to a device
// @Tags Gateway
// @Accept json
// @Produce json
// @Param request body DeviceConnectionRequest true "Device connection request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/gateway/devices/connect [post]
func (h *GatewayHandler) ConnectDevice(c *gin.Context) {
	var req DeviceConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.gatewayService.AddDevice(req.DeviceID, req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Device connected successfully",
		"device_id": req.DeviceID,
		"address":   req.Address,
	})
}

// DisconnectDevice removes a device connection
// @Summary Disconnect device
// @Description Remove gRPC connection to a device
// @Tags Gateway
// @Accept json
// @Produce json
// @Param device_id path string true "Device ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/gateway/devices/{device_id}/disconnect [delete]
func (h *GatewayHandler) DisconnectDevice(c *gin.Context) {
	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	err := h.gatewayService.RemoveDevice(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Device disconnected successfully",
		"device_id": deviceID,
	})
}

// ListConnectedDevices lists all connected devices
// @Summary List connected devices
// @Description Get list of all devices currently connected to the gateway
// @Tags Gateway
// @Accept json
// @Produce json
// @Success 200 {object} DeviceListResponse
// @Router /api/v1/gateway/devices [get]
func (h *GatewayHandler) ListConnectedDevices(c *gin.Context) {
	devices := h.gatewayService.ListConnectedDevices()

	response := DeviceListResponse{
		Devices: devices,
	}

	c.JSON(http.StatusOK, response)
}

// GetDeviceStatus gets the status of a specific device
// @Summary Get device status
// @Description Get detailed status information for a connected device
// @Tags Gateway
// @Accept json
// @Produce json
// @Param device_id path string true "Device ID"
// @Success 200 {object} DeviceStatusResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/gateway/devices/{device_id}/status [get]
func (h *GatewayHandler) GetDeviceStatus(c *gin.Context) {
	deviceID := c.Param("device_id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	conn, err := h.gatewayService.GetDeviceStatus(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := DeviceStatusResponse{
		DeviceID:    conn.DeviceID,
		Address:     conn.Address,
		IsHealthy:   conn.IsHealthy,
		LastPing:    conn.LastPing,
		ConnectedAt: conn.ConnectedAt,
	}

	c.JSON(http.StatusOK, response)
}