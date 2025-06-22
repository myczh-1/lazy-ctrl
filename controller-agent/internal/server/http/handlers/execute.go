package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/command"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/executor"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/security"
)

type ExecuteHandler struct {
	commandService  *command.Service
	executorService *executor.Service
	securityService *security.Service
}

func NewExecuteHandler(
	commandService *command.Service,
	executorService *executor.Service,
	securityService *security.Service,
) *ExecuteHandler {
	return &ExecuteHandler{
		commandService:  commandService,
		executorService: executorService,
		securityService: securityService,
	}
}

// HandleExecute executes a command by ID
//
//	@Summary		Execute a command
//	@Description	Execute a registered command by its ID with optional timeout
//	@Tags			execute
//	@Accept			json
//	@Produce		json
//	@Param			id			query		string	true	"Command ID"
//	@Param			timeout		query		int		false	"Timeout in seconds (default: 30)"
//	@Param			X-Pin		header		string	false	"PIN for authentication (if required)"
//	@Success		200			{object}	executor.ExecutionResult
//	@Failure		400			{object}	gin.H	"Bad Request"
//	@Failure		401			{object}	gin.H	"Unauthorized"
//	@Failure		403			{object}	gin.H	"Forbidden"
//	@Failure		404			{object}	gin.H	"Not Found"
//	@Failure		429			{object}	gin.H	"Too Many Requests"
//	@Failure		500			{object}	gin.H	"Internal Server Error"
//	@Security		PinAuth
//	@Router			/execute [get]
func (h *ExecuteHandler) HandleExecute(c *gin.Context) {
	commandID := c.Query("id")
	if commandID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing command ID parameter",
		})
		return
	}

	// 验证命令访问权限
	if err := h.securityService.ValidateCommandAccess(commandID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 获取命令
	cmd, ok := h.commandService.GetCommand(commandID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Command not found: %s", commandID),
		})
		return
	}

	// 获取平台特定命令
	platformCmd, ok := h.commandService.GetPlatformCommand(cmd)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Command not supported on this platform",
		})
		return
	}

	// 执行命令
	var result *executor.ExecutionResult
	var err error

	timeoutStr := c.Query("timeout")
	if timeoutStr != "" {
		if timeout, parseErr := strconv.Atoi(timeoutStr); parseErr == nil {
			result, err = h.executorService.ExecuteWithTimeout(platformCmd, time.Duration(timeout)*time.Second)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid timeout parameter",
			})
			return
		}
	} else {
		result, err = h.executorService.ExecuteWithTimeout(platformCmd, 30*time.Second)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}