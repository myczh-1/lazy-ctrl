package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        string    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Phone     string    `gorm:"uniqueIndex" json:"phone"`
	Password  string    `gorm:"not null" json:"-"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	Status    string    `gorm:"default:active" json:"status"` // active, disabled, suspended
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Settings
	Settings *UserSettings `gorm:"embedded;embeddedPrefix:settings_" json:"settings"`

	// Associations
	Devices []UserDevice `gorm:"foreignKey:UserID" json:"devices,omitempty"`
}

// UserSettings represents user configuration settings
type UserSettings struct {
	Language                     string `gorm:"default:zh-CN" json:"language"`
	Timezone                     string `gorm:"default:Asia/Shanghai" json:"timezone"`
	EmailNotifications           bool   `gorm:"default:true" json:"email_notifications"`
	PushNotifications            bool   `gorm:"default:true" json:"push_notifications"`
	TwoFactorEnabled             bool   `gorm:"default:false" json:"two_factor_enabled"`
	DeviceVerificationRequired   bool   `gorm:"default:true" json:"device_verification_required"`
	SessionTimeoutMinutes        int    `gorm:"default:60" json:"session_timeout_minutes"`
}

// TableName returns the table name for User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate will set UUID and timestamps
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate will set updated timestamp
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}