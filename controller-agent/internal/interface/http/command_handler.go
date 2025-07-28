package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/entity"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/service"
)

// CommandHandler handles HTTP requests for command operations
type CommandHandler struct {
	commandService *service.CommandService
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(commandService *service.CommandService) *CommandHandler {
	return &CommandHandler{
		commandService: commandService,
	}
}

// CreateCommandRequest represents the request payload for creating a command
type CreateCommandRequest struct {
	ID             string                 `json:"id" binding:"required"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	Icon           string                 `json:"icon"`
	Command        string                 `json:"command" binding:"required"`
	Platform       string                 `json:"platform"`
	CommandType    string                 `json:"commandType"`
	Timeout        int                    `json:"timeout"`
	UserID         string                 `json:"userId"`
	DeviceID       string                 `json:"deviceId"`
	TemplateId     string                 `json:"templateId"`
	TemplateParams map[string]interface{} `json:"templateParams"`
	Security       *SecurityRequest       `json:"security"`
	HomeLayout     *HomeLayoutRequest     `json:"homeLayout"`
}

// UpdateCommandRequest represents the request payload for updating a command
type UpdateCommandRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Command     string `json:"command"`
}

// SecurityRequest represents security configuration in request
type SecurityRequest struct {
	RequirePin bool `json:"requirePin"`
	Whitelist  bool `json:"whitelist"`
	AdminOnly  bool `json:"adminOnly"`
}

// HomeLayoutRequest represents home layout configuration in request
type HomeLayoutRequest struct {
	ShowOnHome      bool                `json:"showOnHome"`
	DefaultPosition *PositionRequest    `json:"defaultPosition"`
	Color           string              `json:"color"`
	Priority        int                 `json:"priority"`
}

// PositionRequest represents position configuration in request
type PositionRequest struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"w"`
	Height int `json:"h"`
}

// CommandResponse represents the response format for command operations
type CommandResponse struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	Icon           string                 `json:"icon"`
	Command        string                 `json:"command"`
	Platform       string                 `json:"platform"`
	CommandType    string                 `json:"commandType"`
	Timeout        int                    `json:"timeout"`
	UserID         string                 `json:"userId"`
	DeviceID       string                 `json:"deviceId"`
	TemplateId     string                 `json:"templateId"`
	TemplateParams map[string]interface{} `json:"templateParams"`
	CreatedAt      string                 `json:"createdAt"`
	UpdatedAt      string                 `json:"updatedAt"`
	RequiresPin    bool                   `json:"requiresPin"`
	Whitelisted    bool                   `json:"whitelisted"`
	Available      bool                   `json:"available"`
	ShowOnHomepage bool                   `json:"showOnHomepage"`
	HomepageColor  string                 `json:"homepageColor,omitempty"`
	HomepagePriority int                  `json:"homepagePriority,omitempty"`
	HomepagePosition *PositionResponse    `json:"homepagePosition,omitempty"`
}

// PositionResponse represents position in response
type PositionResponse struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// @Summary Create a new command
// @Description Create a new command with the provided configuration
// @Tags commands
// @Accept json
// @Produce json
// @Param command body CreateCommandRequest true "Command configuration"
// @Success 201 {object} CommandResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /commands [post]
func (h *CommandHandler) CreateCommand(c *gin.Context) {
	var req CreateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Create command using service
	cmd, err := h.commandService.CreateCommand(ctx, req.ID, req.Name, req.Command)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "command with ID "+req.ID+" already exists" {
			status = http.StatusConflict
		}
		c.JSON(status, ErrorResponse{
			Error:   "Failed to create command",
			Message: err.Error(),
		})
		return
	}
	
	// Update additional fields if provided
	h.updateCommandFields(cmd, &req)
	
	// Convert to response format
	response := h.commandToResponse(cmd)
	c.JSON(http.StatusCreated, response)
}

// @Summary Get all commands
// @Description Retrieve all available commands
// @Tags commands
// @Produce json
// @Success 200 {array} CommandResponse
// @Failure 500 {object} ErrorResponse
// @Router /commands [get]
func (h *CommandHandler) GetAllCommands(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	commands, err := h.commandService.GetAllCommands(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve commands",
			Message: err.Error(),
		})
		return
	}
	
	// Convert to response format
	responses := make([]CommandResponse, len(commands))
	for i, cmd := range commands {
		responses[i] = h.commandToResponse(cmd)
	}
	
	c.JSON(http.StatusOK, responses)
}

// @Summary Get command by ID
// @Description Retrieve a specific command by its ID
// @Tags commands
// @Produce json
// @Param id path string true "Command ID"
// @Success 200 {object} CommandResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /commands/{id} [get]
func (h *CommandHandler) GetCommand(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: "Command ID is required",
		})
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd, err := h.commandService.GetCommand(ctx, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "failed to get command: command not found: "+id {
			status = http.StatusNotFound
		}
		c.JSON(status, ErrorResponse{
			Error:   "Failed to retrieve command",
			Message: err.Error(),
		})
		return
	}
	
	response := h.commandToResponse(cmd)
	c.JSON(http.StatusOK, response)
}

// @Summary Update command
// @Description Update an existing command
// @Tags commands
// @Accept json
// @Produce json
// @Param id path string true "Command ID"
// @Param command body UpdateCommandRequest true "Command updates"
// @Success 200 {object} CommandResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /commands/{id} [put]
func (h *CommandHandler) UpdateCommand(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: "Command ID is required",
		})
		return
	}
	
	var req UpdateCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cmd, err := h.commandService.UpdateCommand(ctx, id, req.Name, req.Description, req.Command)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "failed to get command: command not found: "+id {
			status = http.StatusNotFound
		}
		c.JSON(status, ErrorResponse{
			Error:   "Failed to update command",
			Message: err.Error(),
		})
		return
	}
	
	response := h.commandToResponse(cmd)
	c.JSON(http.StatusOK, response)
}

// @Summary Delete command
// @Description Delete a command by ID
// @Tags commands
// @Produce json
// @Param id path string true "Command ID"
// @Success 204 "Command deleted successfully"
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /commands/{id} [delete]
func (h *CommandHandler) DeleteCommand(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: "Command ID is required",
		})
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := h.commandService.DeleteCommand(ctx, id)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "command with ID "+id+" not found" {
			status = http.StatusNotFound
		}
		c.JSON(status, ErrorResponse{
			Error:   "Failed to delete command",
			Message: err.Error(),
		})
		return
	}
	
	c.Status(http.StatusNoContent)
}

// @Summary Get homepage commands
// @Description Retrieve commands configured for homepage display
// @Tags commands
// @Produce json
// @Success 200 {array} CommandResponse
// @Failure 500 {object} ErrorResponse
// @Router /commands/homepage [get]
func (h *CommandHandler) GetHomepageCommands(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	commands, err := h.commandService.GetHomepageCommands(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to retrieve homepage commands",
			Message: err.Error(),
		})
		return
	}
	
	responses := make([]CommandResponse, len(commands))
	for i, cmd := range commands {
		responses[i] = h.commandToResponse(cmd)
	}
	
	c.JSON(http.StatusOK, responses)
}

// updateCommandFields updates additional command fields from request
func (h *CommandHandler) updateCommandFields(cmd *entity.Command, req *CreateCommandRequest) {
	if req.Description != "" {
		cmd.Description = req.Description
	}
	if req.Category != "" {
		cmd.Category = req.Category
	}
	if req.Icon != "" {
		cmd.Icon = req.Icon
	}
	if req.Platform != "" {
		cmd.Platform = req.Platform
	}
	if req.CommandType != "" {
		cmd.CommandType = req.CommandType
	}
	if req.Timeout > 0 {
		cmd.Timeout = req.Timeout
	}
	if req.UserID != "" {
		cmd.UserID = req.UserID
	}
	if req.DeviceID != "" {
		cmd.DeviceID = req.DeviceID
	}
	if req.TemplateId != "" {
		cmd.TemplateId = req.TemplateId
	}
	if req.TemplateParams != nil {
		cmd.TemplateParams = req.TemplateParams
	}
	
	// Set security configuration
	if req.Security != nil {
		cmd.SetSecurity(req.Security.RequirePin, req.Security.Whitelist, req.Security.AdminOnly)
	}
	
	// Set home layout configuration
	if req.HomeLayout != nil {
		var position *entity.PositionConfig
		if req.HomeLayout.DefaultPosition != nil {
			position = &entity.PositionConfig{
				X:      req.HomeLayout.DefaultPosition.X,
				Y:      req.HomeLayout.DefaultPosition.Y,
				Width:  req.HomeLayout.DefaultPosition.Width,
				Height: req.HomeLayout.DefaultPosition.Height,
			}
		}
		cmd.SetHomeLayout(req.HomeLayout.ShowOnHome, position, req.HomeLayout.Color, req.HomeLayout.Priority)
	}
}

// commandToResponse converts command entity to response format
func (h *CommandHandler) commandToResponse(cmd *entity.Command) CommandResponse {
	response := CommandResponse{
		ID:             cmd.ID,
		Name:           cmd.Name,
		Description:    cmd.Description,
		Category:       cmd.Category,
		Icon:           cmd.Icon,
		Command:        cmd.Command,
		Platform:       cmd.Platform,
		CommandType:    cmd.CommandType,
		Timeout:        cmd.GetTimeout(),
		UserID:         cmd.UserID,
		DeviceID:       cmd.DeviceID,
		TemplateId:     cmd.TemplateId,
		TemplateParams: cmd.TemplateParams,
		CreatedAt:      cmd.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      cmd.UpdatedAt.Format(time.RFC3339),
		RequiresPin:    cmd.RequiresPin(),
		Whitelisted:    cmd.IsWhitelisted(),
		Available:      cmd.IsAvailableOnPlatform(),
		ShowOnHomepage: cmd.ShowOnHomepage(),
		HomepageColor:  cmd.GetHomepageColor(),
		HomepagePriority: cmd.GetHomepagePriority(),
	}
	
	// Add homepage position if available
	if cmd.ShowOnHomepage() {
		x, y, width, height := cmd.GetHomepagePosition()
		response.HomepagePosition = &PositionResponse{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		}
	}
	
	return response
}