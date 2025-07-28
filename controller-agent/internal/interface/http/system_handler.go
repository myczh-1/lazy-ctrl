package http

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/service"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/security"
)

// SystemHandler handles HTTP requests for system operations
type SystemHandler struct {
	commandService  *service.CommandService
	securityService *security.Service
}

// NewSystemHandler creates a new system handler
func NewSystemHandler(
	commandService *service.CommandService,
	securityService *security.Service,
) *SystemHandler {
	return &SystemHandler{
		commandService:  commandService,
		securityService: securityService,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	System    SystemInfo        `json:"system"`
	Services  map[string]string `json:"services"`
}

// SystemInfo represents system information
type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	GoVersion    string `json:"goVersion"`
	NumCPU       int    `json:"numCPU"`
	NumGoroutine int    `json:"numGoroutine"`
}

// AuthRequest represents the authentication request
type AuthRequest struct {
	Pin string `json:"pin" binding:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ReloadResponse represents the reload operation response
type ReloadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents error response format
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// @Summary Health check
// @Description Get the health status of the application
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *SystemHandler) HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "2.0.0", // This should be injected from build
		System: SystemInfo{
			OS:           runtime.GOOS,
			Architecture: runtime.GOARCH,
			GoVersion:    runtime.Version(),
			NumCPU:       runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		Services: map[string]string{
			"command_service":  "healthy",
			"executor_service": "healthy",
			"security_service": "healthy",
		},
	}
	
	c.JSON(http.StatusOK, response)
}

// @Summary Verify PIN
// @Description Verify PIN for authentication
// @Tags authentication
// @Accept json
// @Produce json
// @Param auth body AuthRequest true "Authentication request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/verify [post]
func (h *SystemHandler) VerifyPin(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}
	
	if h.securityService.ValidatePin(req.Pin) {
		c.JSON(http.StatusOK, AuthResponse{
			Success: true,
			Message: "PIN verified successfully",
		})
	} else {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid PIN",
		})
	}
}

// @Summary Reload commands
// @Description Reload command configuration from file
// @Tags system
// @Produce json
// @Success 200 {object} ReloadResponse
// @Failure 500 {object} ErrorResponse
// @Router /reload [post]
func (h *SystemHandler) ReloadCommands(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := h.commandService.ReloadCommands(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to reload commands",
			Message: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, ReloadResponse{
		Success: true,
		Message: "Commands reloaded successfully",
	})
}

// @Summary Get system version
// @Description Get the current version of the application
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /version [get]
func (h *SystemHandler) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"version":   "2.0.0",
		"buildTime": "2024-01-01T00:00:00Z", // This should be injected from build
		"gitCommit": "unknown",               // This should be injected from build
	})
}

// @Summary Get system status
// @Description Get detailed system status and metrics
// @Tags system
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /status [get]
func (h *SystemHandler) GetStatus(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Get command count
	commands, err := h.commandService.GetAllCommands(ctx)
	commandCount := 0
	if err == nil {
		commandCount = len(commands)
	}
	
	// Get homepage command count
	homepageCommands, err := h.commandService.GetHomepageCommands(ctx)
	homepageCount := 0
	if err == nil {
		homepageCount = len(homepageCommands)
	}
	
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	status := map[string]interface{}{
		"uptime":    time.Since(time.Now()).String(), // This should track actual uptime
		"timestamp": time.Now().Format(time.RFC3339),
		"system": map[string]interface{}{
			"os":           runtime.GOOS,
			"architecture": runtime.GOARCH,
			"goVersion":    runtime.Version(),
			"numCPU":       runtime.NumCPU(),
			"numGoroutine": runtime.NumGoroutine(),
		},
		"memory": map[string]interface{}{
			"allocated":     memStats.Alloc,
			"totalAlloc":    memStats.TotalAlloc,
			"sys":           memStats.Sys,
			"numGC":         memStats.NumGC,
		},
		"commands": map[string]interface{}{
			"total":    commandCount,
			"homepage": homepageCount,
		},
		"services": map[string]string{
			"command_service":  "running",
			"executor_service": "running",
			"security_service": "running",
		},
	}
	
	c.JSON(http.StatusOK, status)
}