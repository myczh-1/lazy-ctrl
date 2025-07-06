package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/service/command"
)

type CommandHandler struct {
	commandService *command.Service
}

func NewCommandHandler(commandService *command.Service) *CommandHandler {
	return &CommandHandler{
		commandService: commandService,
	}
}

// HandleListCommands lists all available commands
//
//	@Summary		List all commands
//	@Description	Get a list of all registered commands
//	@Tags			commands
//	@Accept			json
//	@Produce		json
//	@Param			X-Pin	header		string	false	"PIN for authentication (if required)"
//	@Success		200		{object}	gin.H{commands=map[string]interface{}}
//	@Failure		401		{object}	gin.H	"Unauthorized"
//	@Failure		429		{object}	gin.H	"Too Many Requests"
//	@Security		PinAuth
//	@Router			/commands [get]
func (h *CommandHandler) HandleListCommands(c *gin.Context) {
	commands := config.GetCommands()
	commandList := make([]gin.H, 0, len(commands))
	
	for _, cmd := range commands {
		commandInfo := gin.H{
			"id":          cmd.ID,
			"name":        cmd.Name,
			"description": cmd.Description,
			"category":    cmd.Category,
			"icon":        cmd.Icon,
			"timeout":     cmd.GetTimeout(),
			"requiresPin": cmd.RequiresPin(),
			"whitelisted": cmd.IsWhitelisted(),
			"available":   cmd.IsWhitelisted(),
		}
		
		// 添加首页相关信息
		commandInfo["showOnHomepage"] = cmd.ShowOnHomepage()
		if cmd.HomeLayout != nil {
			x, y, width, height := cmd.GetHomepagePosition()
			commandInfo["homepagePosition"] = gin.H{
				"x": x,
				"y": y,
				"width": width,
				"height": height,
			}
			commandInfo["homepageColor"] = cmd.GetHomepageColor()
			commandInfo["homepagePriority"] = cmd.GetHomepagePriority()
		}
		
		// 获取当前平台的命令详情
		if platformCmd, ok := h.commandService.GetPlatformCommand(&cmd); ok {
			commandInfo["available"] = true
			commandInfo["command"] = platformCmd
		} else {
			commandInfo["available"] = false
		}
		
		commandList = append(commandList, commandInfo)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"version":  config.GetCommandsVersion(),
		"commands": commandList,
	})
}

// HandleReloadConfig reloads the command configuration
//
//	@Summary		Reload commands configuration
//	@Description	Reload commands from the configuration file
//	@Tags			commands
//	@Accept			json
//	@Produce		json
//	@Param			X-Pin	header		string	false	"PIN for authentication (if required)"
//	@Success		200		{object}	gin.H{message=string}
//	@Failure		401		{object}	gin.H	"Unauthorized"
//	@Failure		429		{object}	gin.H	"Too Many Requests"
//	@Failure		500		{object}	gin.H	"Internal Server Error"
//	@Security		PinAuth
//	@Router			/reload [post]
func (h *CommandHandler) HandleReloadConfig(c *gin.Context) {
	if err := h.commandService.ReloadCommands(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration reloaded successfully",
	})
}

// CreateCommand creates a new command
//
//	@Summary		Create a new command
//	@Description	Create a new command with the specified configuration
//	@Tags			commands
//	@Accept			json
//	@Produce		json
//	@Param			X-Pin		header		string	false	"PIN for authentication (if required)"
//	@Param			command		body		object	true	"Command configuration"
//	@Success		201			{object}	gin.H{message=string,id=string}
//	@Failure		400			{object}	gin.H	"Bad Request"
//	@Failure		401			{object}	gin.H	"Unauthorized"
//	@Failure		429			{object}	gin.H	"Too Many Requests"
//	@Failure		500			{object}	gin.H	"Internal Server Error"
//	@Security		PinAuth
//	@Router			/commands [post]
func (h *CommandHandler) CreateCommand(c *gin.Context) {
	// TODO: 创建命令的实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

// UpdateCommand updates an existing command
//
//	@Summary		Update a command
//	@Description	Update an existing command configuration
//	@Tags			commands
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string	true	"Command ID"
//	@Param			X-Pin		header		string	false	"PIN for authentication (if required)"
//	@Param			command		body		object	true	"Updated command configuration"
//	@Success		200			{object}	gin.H{message=string}
//	@Failure		400			{object}	gin.H	"Bad Request"
//	@Failure		401			{object}	gin.H	"Unauthorized"
//	@Failure		404			{object}	gin.H	"Not Found"
//	@Failure		429			{object}	gin.H	"Too Many Requests"
//	@Failure		500			{object}	gin.H	"Internal Server Error"
//	@Security		PinAuth
//	@Router			/commands/{id} [put]
func (h *CommandHandler) UpdateCommand(c *gin.Context) {
	// TODO: 修改命令的实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}

// DeleteCommand deletes an existing command
//
//	@Summary		Delete a command
//	@Description	Delete an existing command by its ID
//	@Tags			commands
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Command ID"
//	@Param			X-Pin	header		string	false	"PIN for authentication (if required)"
//	@Success		200		{object}	gin.H{message=string}
//	@Failure		401		{object}	gin.H	"Unauthorized"
//	@Failure		404		{object}	gin.H	"Not Found"
//	@Failure		429		{object}	gin.H	"Too Many Requests"
//	@Failure		500		{object}	gin.H	"Internal Server Error"
//	@Security		PinAuth
//	@Router			/commands/{id} [delete]
func (h *CommandHandler) DeleteCommand(c *gin.Context) {
	// TODO: 删除命令的实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented yet",
	})
}