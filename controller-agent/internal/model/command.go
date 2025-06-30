package model

// CommandConfig 命令配置文件根结构
type CommandConfig struct {
	Version  string    `json:"version"`
	Commands []Command `json:"commands"`
}

// Command 单个命令配置
type Command struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Category    string                 `json:"category,omitempty"`
	Icon        string                 `json:"icon,omitempty"`
	Platforms   map[string]interface{} `json:"platforms"`
	Security    *SecurityConfig        `json:"security,omitempty"`
	Timeout     int                    `json:"timeout,omitempty"`
	UserID      string                 `json:"userId,omitempty"`
	DeviceID    string                 `json:"deviceId,omitempty"`
	Homepage    *HomepageConfig        `json:"homepage,omitempty"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RequirePin bool `json:"requirePin"`
	Whitelist  bool `json:"whitelist"`
	AdminOnly  bool `json:"adminOnly,omitempty"`
}

// HomepageConfig 首页展示配置
type HomepageConfig struct {
	ShowOnHomepage bool `json:"showOnHomepage"`
	X              int  `json:"x,omitempty"`
	Y              int  `json:"y,omitempty"`
	Width          int  `json:"width,omitempty"`
	Height         int  `json:"height,omitempty"`
	Color          string `json:"color,omitempty"`
}

// PlatformCommand 平台特定命令配置
type PlatformCommand struct {
	Command  string        `json:"command,omitempty"`
	Commands []CommandStep `json:"commands,omitempty"`
	Type     string        `json:"type,omitempty"`
}

// CommandStep 命令步骤 (用于序列命令)
type CommandStep struct {
	Type     string `json:"type"`
	Cmd      string `json:"cmd,omitempty"`
	Duration int    `json:"duration,omitempty"`
}


// GetTimeout 获取命令超时时间，如果未设置则返回默认值
func (c *Command) GetTimeout() int {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 10000 // 默认10秒
}

// IsWhitelisted 检查命令是否在白名单中
func (c *Command) IsWhitelisted() bool {
	if c.Security == nil {
		return true // 默认允许
	}
	return c.Security.Whitelist
}

// RequiresPin 检查命令是否需要PIN验证
func (c *Command) RequiresPin() bool {
	if c.Security == nil {
		return false
	}
	return c.Security.RequirePin
}

// RequiresAdmin 检查命令是否需要管理员权限
func (c *Command) RequiresAdmin() bool {
	if c.Security == nil {
		return false
	}
	return c.Security.AdminOnly
}

// ShowOnHomepage 检查命令是否在首页显示
func (c *Command) ShowOnHomepage() bool {
	if c.Homepage == nil {
		return false
	}
	return c.Homepage.ShowOnHomepage
}

// GetHomepagePosition 获取首页位置信息
func (c *Command) GetHomepagePosition() (x, y, width, height int) {
	if c.Homepage == nil {
		return 0, 0, 1, 1 // 默认大小
	}
	x = c.Homepage.X
	y = c.Homepage.Y
	width = c.Homepage.Width
	height = c.Homepage.Height
	
	// 设置默认值
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}
	
	return x, y, width, height
}

// GetHomepageColor 获取首页卡片颜色
func (c *Command) GetHomepageColor() string {
	if c.Homepage == nil || c.Homepage.Color == "" {
		return ""
	}
	return c.Homepage.Color
}