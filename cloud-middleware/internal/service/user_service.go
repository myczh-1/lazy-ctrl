package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/myczh-1/lazy-ctrl-cloud/internal/auth"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/config"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/model"
	"github.com/myczh-1/lazy-ctrl-cloud/internal/repository"
)

// UserService defines the interface for user business logic
type UserService interface {
	// Authentication
	Login(username, password string) (*LoginResult, error)
	RefreshToken(refreshToken string) (*auth.TokenPair, error)
	Logout(userID, refreshToken string) error

	// User Management (Admin only)
	CreateUser(req *CreateUserRequest) (*model.User, error)
	GetUser(userID string) (*model.User, error)
	UpdateUser(userID string, req *UpdateUserRequest) (*model.User, error)
	DeleteUser(userID string) error
	ListUsers(offset, limit int) ([]*model.User, int64, error)
	SetUserRole(userID, role string) error

	// Profile Management
	GetProfile(userID string) (*model.User, error)
	UpdateProfile(userID string, req *UpdateProfileRequest) (*model.User, error)
	ChangePassword(userID, oldPassword, newPassword string) error

	// System
	InitializeSystem() error
	IsAdmin(userID string) (bool, error)
}

// LoginResult represents login response
type LoginResult struct {
	User   *model.User      `json:"user"`
	Tokens *auth.TokenPair  `json:"tokens"`
}

// CreateUserRequest represents create user request
type CreateUserRequest struct {
	Username string                `json:"username" binding:"required"`
	Email    string                `json:"email" binding:"required"`
	Phone    string                `json:"phone"`
	Password string                `json:"password" binding:"required"`
	Nickname string                `json:"nickname"`
	Role     string                `json:"role"`
	Settings *model.UserSettings   `json:"settings"`
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
	Email    string                `json:"email"`
	Phone    string                `json:"phone"`
	Nickname string                `json:"nickname"`
	Role     string                `json:"role"`
	Status   string                `json:"status"`
	Settings *model.UserSettings   `json:"settings"`
}

// UpdateProfileRequest represents update profile request
type UpdateProfileRequest struct {
	Nickname  string                `json:"nickname"`
	AvatarURL string                `json:"avatar_url"`
	Settings  *model.UserSettings   `json:"settings"`
}

// userService implements UserService interface
type userService struct {
	userRepo   repository.UserRepository
	jwtService *auth.JWTService
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, jwtConfig config.JWTConfig) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtService: auth.NewJWTService(jwtConfig),
	}
}

// Login authenticates user and returns tokens
func (s *userService) Login(username, password string) (*LoginResult, error) {
	// Validate credentials
	user, err := s.userRepo.ValidateCredentials(username, password)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Remove password from response
	user.Password = ""

	return &LoginResult{
		User:   user,
		Tokens: tokens,
	}, nil
}

// RefreshToken generates new tokens from refresh token
func (s *userService) RefreshToken(refreshToken string) (*auth.TokenPair, error) {
	return s.jwtService.RefreshToken(refreshToken)
}

// Logout invalidates user tokens (placeholder - in production would need token blacklist)
func (s *userService) Logout(userID, refreshToken string) error {
	// In a production system, you would typically:
	// 1. Add the refresh token to a blacklist
	// 2. Store blacklisted tokens in Redis with expiration
	// For now, we'll just validate the token
	_, err := s.jwtService.ValidateToken(refreshToken)
	return err
}

// CreateUser creates a new user (admin only)
func (s *userService) CreateUser(req *CreateUserRequest) (*model.User, error) {
	// Validate input
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Check if username or email already exists
	if s.userRepo.IsUsernameExists(req.Username) {
		return nil, errors.New("username already exists")
	}
	if s.userRepo.IsEmailExists(req.Email) {
		return nil, errors.New("email already exists")
	}

	// Set default role if not specified
	role := req.Role
	if role == "" {
		role = "user"
	}

	// Set default settings if not provided
	settings := req.Settings
	if settings == nil {
		settings = &model.UserSettings{
			Language:                   "zh-CN",
			Timezone:                   "Asia/Shanghai",
			EmailNotifications:         true,
			PushNotifications:          true,
			TwoFactorEnabled:           false,
			DeviceVerificationRequired: true,
			SessionTimeoutMinutes:      60,
		}
	}

	// Create user
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
		Nickname:  req.Nickname,
		Role:      role,
		Status:    "active",
		Settings:  settings,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Remove password from response
	user.Password = ""
	return user, nil
}

// GetUser retrieves user by ID
func (s *userService) GetUser(userID string) (*model.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	
	// Remove password from response
	user.Password = ""
	return user, nil
}

// UpdateUser updates user information (admin only)
func (s *userService) UpdateUser(userID string, req *UpdateUserRequest) (*model.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Email != "" {
		if err := s.validateEmail(req.Email); err != nil {
			return nil, err
		}
		// Check if email is already used by another user
		if existingUser, _ := s.userRepo.GetByEmail(req.Email); existingUser != nil && existingUser.ID != userID {
			return nil, errors.New("email already exists")
		}
		user.Email = req.Email
	}
	
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	
	if req.Role != "" {
		if req.Role != "admin" && req.Role != "user" {
			return nil, errors.New("invalid role")
		}
		user.Role = req.Role
	}
	
	if req.Status != "" {
		if req.Status != "active" && req.Status != "disabled" && req.Status != "suspended" {
			return nil, errors.New("invalid status")
		}
		user.Status = req.Status
	}
	
	if req.Settings != nil {
		user.Settings = req.Settings
	}

	// Update user
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Remove password from response
	user.Password = ""
	return user, nil
}

// DeleteUser soft deletes a user (admin only)
func (s *userService) DeleteUser(userID string) error {
	return s.userRepo.Delete(userID)
}

// ListUsers returns paginated list of users
func (s *userService) ListUsers(offset, limit int) ([]*model.User, int64, error) {
	users, total, err := s.userRepo.List(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	// Remove passwords from response
	for _, user := range users {
		user.Password = ""
	}

	return users, total, nil
}

// SetUserRole sets user role (admin only)
func (s *userService) SetUserRole(userID, role string) error {
	return s.userRepo.SetUserRole(userID, role)
}

// GetProfile gets user profile (own profile)
func (s *userService) GetProfile(userID string) (*model.User, error) {
	return s.GetUser(userID)
}

// UpdateProfile updates user profile (own profile)
func (s *userService) UpdateProfile(userID string, req *UpdateProfileRequest) (*model.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	// Update allowed fields
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	
	if req.Settings != nil {
		user.Settings = req.Settings
	}

	// Update user
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Remove password from response
	user.Password = ""
	return user, nil
}

// ChangePassword changes user password
func (s *userService) ChangePassword(userID, oldPassword, newPassword string) error {
	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	return s.userRepo.ChangePassword(userID, oldPassword, newPassword)
}

// InitializeSystem initializes the system with default admin user
func (s *userService) InitializeSystem() error {
	return s.userRepo.CreateDefaultAdmin()
}

// IsAdmin checks if user is admin
func (s *userService) IsAdmin(userID string) (bool, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return false, err
	}
	
	return user.Role == "admin", nil
}

// Validation helpers

func (s *userService) validateCreateUserRequest(req *CreateUserRequest) error {
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}

	if err := s.validateEmail(req.Email); err != nil {
		return err
	}

	if err := s.validatePassword(req.Password); err != nil {
		return err
	}

	if req.Role != "" && req.Role != "admin" && req.Role != "user" {
		return errors.New("invalid role")
	}

	return nil
}

func (s *userService) validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

func (s *userService) validatePassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return errors.New("password is required")
	}

	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	// In production, you might want stronger password requirements
	// For now, we'll keep it simple

	return nil
}