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
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RequirePin bool `json:"requirePin"`
	Whitelist  bool `json:"whitelist"`
	AdminOnly  bool `json:"adminOnly,omitempty"`
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

// LegacyCommandConfig 旧版命令配置 (向后兼容)
type LegacyCommandConfig map[string]interface{}

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