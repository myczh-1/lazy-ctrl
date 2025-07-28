package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/service"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/infrastructure/security"
)

// ExecuteHandler handles HTTP requests for command execution
type ExecuteHandler struct {
	commandService  *service.CommandService
	executorService *executor.Service
	securityService *security.Service
}

// NewExecuteHandler creates a new execute handler
func NewExecuteHandler(
	commandService *service.CommandService,
	executorService *executor.Service,
	securityService *security.Service,
) *ExecuteHandler {
	return &ExecuteHandler{
		commandService:  commandService,
		executorService: executorService,
		securityService: securityService,
	}
}

// ExecuteRequest represents the request payload for command execution
type ExecuteRequest struct {
	ID  string `form:"id" json:"id" binding:"required"`
	Pin string `form:"pin" json:"pin"`
}

// ExecuteResponse represents the response for command execution
type ExecuteResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
	ExitCode int   `json:"exitCode"`
	Duration int64  `json:"duration"` // Duration in milliseconds
}

// @Summary Execute a command
// @Description Execute a command by its ID
// @Tags execution
// @Accept json
// @Produce json
// @Param id query string true "Command ID"
// @Param pin query string false "PIN for authentication (if required)"
// @Success 200 {object} ExecuteResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /execute [get]
func (h *ExecuteHandler) ExecuteCommand(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request parameters",
			Message: err.Error(),
		})
		return
	}
	
	h.executeCommand(c, req)
}

// @Summary Execute a command (POST)
// @Description Execute a command by its ID using POST method
// @Tags execution
// @Accept json
// @Produce json
// @Param request body ExecuteRequest true "Execute request"
// @Success 200 {object} ExecuteResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /execute [post]
func (h *ExecuteHandler) ExecuteCommandPost(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}
	
	h.executeCommand(c, req)
}

// executeCommand performs the actual command execution
func (h *ExecuteHandler) executeCommand(c *gin.Context, req ExecuteRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	clientIP := c.ClientIP()
	
	// Rate limiting check
	if err := h.securityService.CheckRateLimit(clientIP); err != nil {
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Error:   "Rate limit exceeded",
			Message: "Too many requests, please try again later",
		})
		return
	}
	
	// Get command to check if PIN is required
	cmd, err := h.commandService.GetCommand(ctx, req.ID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "failed to get command: command not found: "+req.ID {
			status = http.StatusNotFound
		}
		c.JSON(status, ErrorResponse{
			Error:   "Command not found",
			Message: err.Error(),
		})
		return
	}
	
	// PIN verification if required
	if cmd.RequiresPin() {
		if !h.securityService.ValidatePin(req.Pin) {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Error:   "Authentication failed",
				Message: "Invalid or missing PIN",
			})
			return
		}
	}
	
	// Get platform-specific command
	platformCommand, err := h.commandService.GetPlatformCommand(ctx, req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Command not available",
			Message: err.Error(),
		})
		return
	}
	
	// Record execution start time
	startTime := time.Now()
	
	// Execute command with timeout
	executeCtx, executeCancel := context.WithTimeout(ctx, time.Duration(cmd.GetTimeout())*time.Millisecond)
	defer executeCancel()
	
	result, err := h.executorService.Execute(executeCtx, platformCommand)
	duration := time.Since(startTime).Milliseconds()
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, ExecuteResponse{
			Success:  false,
			Output:   "",
			Error:    err.Error(),
			ExitCode: -1,
			Duration: duration,
		})
		return
	}
	
	// Return successful execution result
	c.JSON(http.StatusOK, ExecuteResponse{
		Success:  result.Success,
		Output:   result.Output,
		Error:    result.Error,
		ExitCode: result.ExitCode,
		Duration: duration,
	})
}

// @Summary Get command execution info
// @Description Get information about a command without executing it
// @Tags execution
// @Produce json
// @Param id query string true "Command ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /execute/info [get]
func (h *ExecuteHandler) GetCommandInfo(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: "Command ID is required",
		})
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	info, err := h.commandService.GetCommandInfo(ctx, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "failed to get command: command not found: "+id {
			status = http.StatusNotFound
		}
		c.JSON(status, ErrorResponse{
			Error:   "Failed to get command info",
			Message: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, info)
}