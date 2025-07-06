package model

// CommandConfig 命令配置文件根结构
type CommandConfig struct {
	Version  string    `json:"version"`
	Commands []Command `json:"commands"`
}

// Command 单个命令配置
type Command struct {
	ID             string               `json:"id"`
	Name           string               `json:"name,omitempty"`
	Description    string               `json:"description,omitempty"`
	Category       string               `json:"category,omitempty"`
	Icon           string               `json:"icon,omitempty"`
	Command        string               `json:"command"`
	Platform       string               `json:"platform"`
	CommandType    string               `json:"commandType,omitempty"`
	Security       *SecurityConfig      `json:"security,omitempty"`
	Timeout        int                  `json:"timeout,omitempty"`
	UserID         string               `json:"userId,omitempty"`
	DeviceID       string               `json:"deviceId,omitempty"`
	HomeLayout     *HomeLayoutConfig    `json:"homeLayout,omitempty"`
	TemplateId     string               `json:"templateId,omitempty"`
	TemplateParams map[string]interface{} `json:"templateParams,omitempty"`
	CreatedAt      string               `json:"createdAt,omitempty"`
	UpdatedAt      string               `json:"updatedAt,omitempty"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RequirePin bool `json:"requirePin"`
	Whitelist  bool `json:"whitelist"`
	AdminOnly  bool `json:"adminOnly,omitempty"`
}

// HomeLayoutConfig 首页展示配置
type HomeLayoutConfig struct {
	ShowOnHome      bool                   `json:"showOnHome"`
	DefaultPosition *PositionConfig        `json:"defaultPosition,omitempty"`
	Color           string                 `json:"color,omitempty"`
	Priority        int                    `json:"priority,omitempty"`
}

// PositionConfig 位置配置
type PositionConfig struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"w"`
	Height int `json:"h"`
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
	if c.HomeLayout == nil {
		return false
	}
	return c.HomeLayout.ShowOnHome
}

// GetHomepagePosition 获取首页位置信息
func (c *Command) GetHomepagePosition() (x, y, width, height int) {
	if c.HomeLayout == nil || c.HomeLayout.DefaultPosition == nil {
		return 0, 0, 1, 1 // 默认大小
	}
	pos := c.HomeLayout.DefaultPosition
	x = pos.X
	y = pos.Y
	width = pos.Width
	height = pos.Height
	
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
	if c.HomeLayout == nil || c.HomeLayout.Color == "" {
		return ""
	}
	return c.HomeLayout.Color
}

// GetHomepagePriority 获取首页优先级
func (c *Command) GetHomepagePriority() int {
	if c.HomeLayout == nil {
		return 0
	}
	return c.HomeLayout.Priority
}