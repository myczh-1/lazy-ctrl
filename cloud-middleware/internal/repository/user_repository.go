package repository

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/myczh-1/lazy-ctrl-cloud/internal/model"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// User CRUD operations
	Create(user *model.User) error
	GetByID(id string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id string) error
	List(offset, limit int) ([]*model.User, int64, error)

	// Authentication operations
	ValidateCredentials(username, password string) (*model.User, error)
	ChangePassword(userID, oldPassword, newPassword string) error

	// Admin operations
	CreateDefaultAdmin() error
	SetUserRole(userID, role string) error
	IsUsernameExists(username string) bool
	IsEmailExists(email string) bool
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userRepository) Create(user *model.User) error {
	// Hash password before saving
	if user.Password != "" {
		hashedPassword, err := r.hashPassword(user.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(id string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(user *model.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(id string) error {
	if err := r.db.Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// List retrieves users with pagination
func (r *userRepository) List(offset, limit int) ([]*model.User, int64, error) {
	var users []*model.User
	var total int64

	// Count total records
	if err := r.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// ValidateCredentials validates user credentials and returns user if valid
func (r *userRepository) ValidateCredentials(username, password string) (*model.User, error) {
	var user model.User
	
	// Try to find user by username, email, or phone
	if err := r.db.Where("username = ? OR email = ? OR phone = ?", username, username, username).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user is active
	if user.Status != "active" {
		return nil, errors.New("user account is disabled")
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

// ChangePassword changes user password
func (r *userRepository) ChangePassword(userID, oldPassword, newPassword string) error {
	// Get user
	user, err := r.GetByID(userID)
	if err != nil {
		return err
	}

	// Validate old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := r.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := r.db.Model(user).Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// CreateDefaultAdmin creates a default admin user if no admin exists
func (r *userRepository) CreateDefaultAdmin() error {
	// Check if any admin user exists
	var count int64
	if err := r.db.Model(&model.User{}).Where("status = 'active'").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing users: %w", err)
	}

	// If users exist, don't create default admin
	if count > 0 {
		return nil
	}

	// Create default admin user
	admin := &model.User{
		Username: "admin",
		Email:    "admin@lazy-ctrl.local",
		Password: "admin123", // Will be hashed in Create method
		Nickname: "系统管理员",
		Role:     "admin",
		Status:   "active",
		Settings: &model.UserSettings{
			Language:                   "zh-CN",
			Timezone:                   "Asia/Shanghai",
			EmailNotifications:         false,
			PushNotifications:          false,
			TwoFactorEnabled:           false,
			DeviceVerificationRequired: true,
			SessionTimeoutMinutes:      60,
		},
	}

	if err := r.Create(admin); err != nil {
		return fmt.Errorf("failed to create default admin: %w", err)
	}

	return nil
}

// SetUserRole sets user role (admin or user)
func (r *userRepository) SetUserRole(userID, role string) error {
	if role != "admin" && role != "user" {
		return errors.New("invalid role: must be 'admin' or 'user'")
	}

	if err := r.db.Model(&model.User{}).Where("id = ?", userID).Update("role", role).Error; err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	return nil
}

// IsUsernameExists checks if username already exists
func (r *userRepository) IsUsernameExists(username string) bool {
	var count int64
	r.db.Model(&model.User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

// IsEmailExists checks if email already exists
func (r *userRepository) IsEmailExists(email string) bool {
	var count int64
	r.db.Model(&model.User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

// hashPassword hashes a password using bcrypt
func (r *userRepository) hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}