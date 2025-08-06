package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/middleware"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/model"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/service"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// LoginRequest represents login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response
type LoginResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	User         *UserResponse `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents refresh token response
type RefreshTokenResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

// LogoutRequest represents logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateUserRequest represents create user request (admin only)
type CreateUserRequest struct {
	Username string                      `json:"username" binding:"required"`
	Email    string                      `json:"email" binding:"required"`
	Phone    string                      `json:"phone"`
	Password string                      `json:"password" binding:"required"`
	Nickname string                      `json:"nickname"`
	Role     string                      `json:"role"`
}

// UpdateUserRequest represents update user request (admin only)
type UpdateUserRequest struct {
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

// UpdateProfileRequest represents update profile request
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// ChangePasswordRequest represents change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UserResponse represents user in response
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    []*UserResponse `json:"data"`
	Total   int64          `json:"total"`
	Page    int            `json:"page"`
	Limit   int            `json:"limit"`
}

// StandardResponse represents standard API response
type StandardResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	result, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		Success:      true,
		Message:      "Login successful",
		User:         h.toUserResponse(result.User),
		AccessToken:  result.Tokens.AccessToken,
		RefreshToken: result.Tokens.RefreshToken,
		ExpiresAt:    result.Tokens.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// RefreshToken handles token refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	tokens, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Message: "Invalid refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, RefreshTokenResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Logout handles user logout
func (h *UserHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	if err := h.userService.Logout(userID, req.RefreshToken); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// GetProfile gets user profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    h.toUserResponse(user),
	})
}

// UpdateProfile updates user profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	serviceReq := &service.UpdateProfileRequest{
		Nickname:  req.Nickname,
		AvatarURL: req.AvatarURL,
	}

	user, err := h.userService.UpdateProfile(userID, serviceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data:    h.toUserResponse(user),
	})
}

// ChangePassword changes user password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return
	}

	if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// Admin operations

// CreateUser creates a new user (admin only)
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Check if user is admin
	if !h.checkAdminPermission(c) {
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	serviceReq := &service.CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		Nickname: req.Nickname,
		Role:     req.Role,
	}

	user, err := h.userService.CreateUser(serviceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, StandardResponse{
		Success: true,
		Message: "User created successfully",
		Data:    h.toUserResponse(user),
	})
}

// GetUser gets user by ID (admin only)
func (h *UserHandler) GetUser(c *gin.Context) {
	// Check if user is admin
	if !h.checkAdminPermission(c) {
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "User ID is required",
		})
		return
	}

	user, err := h.userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data:    h.toUserResponse(user),
	})
}

// UpdateUser updates user (admin only)
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Check if user is admin
	if !h.checkAdminPermission(c) {
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "User ID is required",
		})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "Invalid request format: " + err.Error(),
		})
		return
	}

	serviceReq := &service.UpdateUserRequest{
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
		Role:     req.Role,
		Status:   req.Status,
	}

	user, err := h.userService.UpdateUser(userID, serviceReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: "User updated successfully",
		Data:    h.toUserResponse(user),
	})
}

// DeleteUser deletes user (admin only)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Check if user is admin
	if !h.checkAdminPermission(c) {
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: "User ID is required",
		})
		return
	}

	if err := h.userService.DeleteUser(userID); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ListUsers lists users with pagination (admin only)
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Check if user is admin
	if !h.checkAdminPermission(c) {
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := (page - 1) * limit

	users, total, err := h.userService.ListUsers(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Convert to response format
	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = h.toUserResponse(user)
	}

	c.JSON(http.StatusOK, UserListResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    userResponses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	})
}

// Helper functions

// checkAdminPermission checks if current user is admin
func (h *UserHandler) checkAdminPermission(c *gin.Context) bool {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Message: "User not authenticated",
		})
		return false
	}

	isAdmin, err := h.userService.IsAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Message: "Failed to check user permissions",
		})
		return false
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, StandardResponse{
			Success: false,
			Message: "Admin permission required",
		})
		return false
	}

	return true
}

// toUserResponse converts user model to response format
func (h *UserHandler) toUserResponse(user *model.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}