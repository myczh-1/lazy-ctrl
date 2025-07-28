package entity

import (
	"runtime"
	"time"
)

// Command represents a command entity in the domain
type Command struct {
	ID             string
	Name           string
	Description    string
	Category       string
	Icon           string
	Command        string
	Platform       string
	CommandType    string
	Security       *SecurityConfig
	Timeout        int
	UserID         string
	DeviceID       string
	HomeLayout     *HomeLayoutConfig
	TemplateId     string
	TemplateParams map[string]interface{}
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SecurityConfig represents security configuration for a command
type SecurityConfig struct {
	RequirePin bool
	Whitelist  bool
	AdminOnly  bool
}

// HomeLayoutConfig represents homepage layout configuration
type HomeLayoutConfig struct {
	ShowOnHome      bool
	DefaultPosition *PositionConfig
	Color           string
	Priority        int
}

// PositionConfig represents position configuration for UI layout
type PositionConfig struct {
	X      int
	Y      int
	Width  int
	Height int
}

// CommandStep represents a step in a sequential command
type CommandStep struct {
	Type     string
	Cmd      string
	Duration int
}

// NewCommand creates a new command with default values
func NewCommand(id, name, command string) *Command {
	now := time.Now()
	return &Command{
		ID:        id,
		Name:      name,
		Command:   command,
		Platform:  runtime.GOOS,
		Timeout:   10000, // 10 seconds default
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GetTimeout returns the command timeout, with default fallback
func (c *Command) GetTimeout() int {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 10000 // Default 10 seconds
}

// IsWhitelisted checks if the command is whitelisted
func (c *Command) IsWhitelisted() bool {
	if c.Security == nil {
		return true // Default allow
	}
	return c.Security.Whitelist
}

// RequiresPin checks if the command requires PIN verification
func (c *Command) RequiresPin() bool {
	if c.Security == nil {
		return false
	}
	return c.Security.RequirePin
}

// RequiresAdmin checks if the command requires admin privileges
func (c *Command) RequiresAdmin() bool {
	if c.Security == nil {
		return false
	}
	return c.Security.AdminOnly
}

// ShowOnHomepage checks if the command should be displayed on homepage
func (c *Command) ShowOnHomepage() bool {
	if c.HomeLayout == nil {
		return false
	}
	return c.HomeLayout.ShowOnHome
}

// GetHomepagePosition returns homepage position information
func (c *Command) GetHomepagePosition() (x, y, width, height int) {
	if c.HomeLayout == nil || c.HomeLayout.DefaultPosition == nil {
		return 0, 0, 1, 1 // Default size
	}
	pos := c.HomeLayout.DefaultPosition
	x = pos.X
	y = pos.Y
	width = pos.Width
	height = pos.Height
	
	// Set default values
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	
	return x, y, width, height
}

// GetHomepageColor returns homepage card color
func (c *Command) GetHomepageColor() string {
	if c.HomeLayout == nil || c.HomeLayout.Color == "" {
		return ""
	}
	return c.HomeLayout.Color
}

// GetHomepagePriority returns homepage priority
func (c *Command) GetHomepagePriority() int {
	if c.HomeLayout == nil {
		return 0
	}
	return c.HomeLayout.Priority
}

// IsAvailableOnPlatform checks if command is available on current platform
func (c *Command) IsAvailableOnPlatform() bool {
	return c.Platform == runtime.GOOS
}

// Update updates the command with new values and sets UpdatedAt
func (c *Command) Update(name, description, command string) {
	if name != "" {
		c.Name = name
	}
	if description != "" {
		c.Description = description
	}
	if command != "" {
		c.Command = command
	}
	c.UpdatedAt = time.Now()
}

// SetSecurity sets security configuration for the command
func (c *Command) SetSecurity(requirePin, whitelist, adminOnly bool) {
	c.Security = &SecurityConfig{
		RequirePin: requirePin,
		Whitelist:  whitelist,
		AdminOnly:  adminOnly,
	}
}

// SetHomeLayout sets homepage layout configuration
func (c *Command) SetHomeLayout(showOnHome bool, position *PositionConfig, color string, priority int) {
	c.HomeLayout = &HomeLayoutConfig{
		ShowOnHome:      showOnHome,
		DefaultPosition: position,
		Color:           color,
		Priority:        priority,
	}
}