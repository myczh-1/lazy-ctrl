package model

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Device represents a controlled device
type Device struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	DeviceName   string    `gorm:"not null" json:"device_name"`
	DeviceType   string    `gorm:"not null" json:"device_type"` // desktop, laptop, server
	Platform     string    `gorm:"not null" json:"platform"`   // windows, linux, macos
	IPAddress    string    `json:"ip_address"`
	MACAddress   string    `json:"mac_address"`
	AgentVersion string    `json:"agent_version"`
	Online       bool      `gorm:"default:false" json:"online"`
	LastSeen     time.Time `json:"last_seen"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// System information stored as JSON
	SystemInfo SystemInfo `gorm:"type:text" json:"system_info"`

	// Device settings
	Settings *DeviceSettings `gorm:"embedded;embeddedPrefix:settings_" json:"settings"`

	// Associations
	Users    []UserDevice    `gorm:"foreignKey:DeviceID" json:"users,omitempty"`
	Commands []DeviceCommand `gorm:"foreignKey:DeviceID" json:"commands,omitempty"`
}

// SystemInfo represents device system information
type SystemInfo map[string]interface{}

// Value implements driver.Valuer interface for GORM
func (s SystemInfo) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner interface for GORM
func (s *SystemInfo) Scan(value interface{}) error {
	if value == nil {
		*s = make(SystemInfo)
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, s)
}

// DeviceSettings represents device configuration
type DeviceSettings struct {
	AutoStart                bool     `gorm:"default:false" json:"auto_start"`
	AllowRemoteShutdown      bool     `gorm:"default:false" json:"allow_remote_shutdown"`
	RequirePinForExecution   bool     `gorm:"default:true" json:"require_pin_for_execution"`
	AllowedCommands          string   `gorm:"type:text" json:"allowed_commands"`
	SecurityLevel            string   `gorm:"default:medium" json:"security_level"` // low, medium, high
}

// UserDevice represents the many-to-many relationship between users and devices
type UserDevice struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	UserID   string `gorm:"not null;index" json:"user_id"`
	DeviceID string `gorm:"not null;index" json:"device_id"`
	Role     string `gorm:"not null;default:user" json:"role"` // owner, admin, user, viewer
	Status   string `gorm:"not null;default:active" json:"status"` // active, disabled
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Foreign keys
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Device Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
}

// DeviceCommand represents commands configured on a device
type DeviceCommand struct {
	ID             string                 `gorm:"primaryKey" json:"id"`
	DeviceID       string                 `gorm:"not null;index" json:"device_id"`
	CommandID      string                 `gorm:"not null" json:"command_id"` // Original command ID from device
	Name           string                 `gorm:"not null" json:"name"`
	Description    string                 `json:"description"`
	Category       string                 `json:"category"`
	Icon           string                 `json:"icon"`
	Command        string                 `gorm:"not null" json:"command"`
	Platform       string                 `json:"platform"`
	CommandType    string                 `json:"command_type"`
	Timeout        int                    `gorm:"default:30000" json:"timeout"` // milliseconds
	TemplateID     string                 `json:"template_id"`
	TemplateParams map[string]interface{} `gorm:"type:text" json:"template_params"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      gorm.DeletedAt         `gorm:"index" json:"-"`

	// Security settings
	RequiresPin bool `gorm:"default:false" json:"requires_pin"`
	Whitelisted bool `gorm:"default:false" json:"whitelisted"`
	AdminOnly   bool `gorm:"default:false" json:"admin_only"`

	// Homepage layout settings
	ShowOnHomepage   bool             `gorm:"default:false" json:"show_on_homepage"`
	HomepageColor    string           `json:"homepage_color"`
	HomepagePriority int              `gorm:"default:0" json:"homepage_priority"`
	HomepagePosition *PositionConfig  `gorm:"type:text" json:"homepage_position"`

	// Foreign key
	Device Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
}

// PositionConfig represents position configuration
type PositionConfig struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Value implements driver.Valuer interface for GORM
func (p PositionConfig) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements sql.Scanner interface for GORM
func (p *PositionConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	return json.Unmarshal(bytes, p)
}

// ExecutionLog represents command execution history
type ExecutionLog struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"not null;index" json:"user_id"`
	DeviceID  string    `gorm:"not null;index" json:"device_id"`
	CommandID string    `gorm:"not null;index" json:"command_id"`
	Success   bool      `json:"success"`
	Output    string    `json:"output"`
	Error     string    `json:"error"`
	ExitCode  int       `json:"exit_code"`
	Duration  int64     `json:"duration"` // milliseconds
	ClientIP  string    `json:"client_ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`

	// Foreign keys
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Device Device `gorm:"foreignKey:DeviceID" json:"device,omitempty"`
}

// TableName methods
func (Device) TableName() string {
	return "devices"
}

func (UserDevice) TableName() string {
	return "user_devices"
}

func (DeviceCommand) TableName() string {
	return "device_commands"
}

func (ExecutionLog) TableName() string {
	return "execution_logs"
}

// BeforeCreate hooks
func (d *Device) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = generateUUID()
	}
	d.CreatedAt = time.Now()
	d.UpdatedAt = time.Now()
	return nil
}

func (ud *UserDevice) BeforeCreate(tx *gorm.DB) error {
	ud.CreatedAt = time.Now()
	ud.UpdatedAt = time.Now()
	return nil
}

func (dc *DeviceCommand) BeforeCreate(tx *gorm.DB) error {
	if dc.ID == "" {
		dc.ID = generateUUID()
	}
	dc.CreatedAt = time.Now()
	dc.UpdatedAt = time.Now()
	return nil
}

func (el *ExecutionLog) BeforeCreate(tx *gorm.DB) error {
	if el.ID == "" {
		el.ID = generateUUID()
	}
	el.CreatedAt = time.Now()
	return nil
}

// BeforeUpdate hooks
func (d *Device) BeforeUpdate(tx *gorm.DB) error {
	d.UpdatedAt = time.Now()
	return nil
}

func (ud *UserDevice) BeforeUpdate(tx *gorm.DB) error {
	ud.UpdatedAt = time.Now()
	return nil
}

func (dc *DeviceCommand) BeforeUpdate(tx *gorm.DB) error {
	dc.UpdatedAt = time.Now()
	return nil
}

// generateUUID generates a simple UUID-like string
func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}