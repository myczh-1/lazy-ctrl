package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/model"
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

	// 检查PIN验证
	if h.commandService.RequiresPin(commandID) {
		pin := c.GetHeader("X-Pin")
		if !h.securityService.ValidatePin(pin) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "PIN required or invalid",
			})
			return
		}
	}

	// 执行命令
	result, err := h.executeCommand(c, commandID, cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// executeCommand 执行命令
func (h *ExecuteHandler) executeCommand(c *gin.Context, commandID string, cmd *model.Command) (*executor.ExecutionResult, error) {
	// 获取超时时间
	timeout := h.getTimeout(c, commandID)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 获取平台特定命令
	platformCmd, ok := h.commandService.GetPlatformCommand(cmd)
	if !ok {
		// 检查是否是复杂命令序列
		if steps := h.extractCommandSteps(cmd); len(steps) > 0 {
			return h.executorService.ExecuteSequence(ctx, steps)
		}
		return nil, fmt.Errorf("command not supported on this platform")
	}

	return h.executorService.Execute(ctx, platformCmd)
}


// extractCommandSteps 从命令中提取命令步骤
func (h *ExecuteHandler) extractCommandSteps(cmd *model.Command) []model.CommandStep {
	// 基于新的设计，单个命令条目不包含复杂步骤
	// 如果需要复杂步骤，应该由前端生成对应的命令
	// 这里暂时返回空，保持简单设计
	return nil
}

// getTimeout 获取超时时间
func (h *ExecuteHandler) getTimeout(c *gin.Context, commandID string) time.Duration {
	// 优先使用URL参数
	if timeoutStr := c.Query("timeout"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			return time.Duration(timeout) * time.Second
		}
	}

	// 使用命令特定超时
	timeout := h.commandService.GetCommandTimeout(commandID)
	return time.Duration(timeout) * time.Millisecond

	// 默认超时
	return 30 * time.Second
}