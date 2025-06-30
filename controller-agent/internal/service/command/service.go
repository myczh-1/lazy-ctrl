package command

import (
	"fmt"
	"runtime"

	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/myczh-1/lazy-ctrl-agent/internal/model"
)

type Service struct {
	config *config.Config
}

func NewService(config *config.Config) *Service {
	return &Service{
		config: config,
	}
}

func (s *Service) GetCommand(id string) (*model.Command, bool) {
	commands := config.GetCommandsV2()
	for _, cmd := range commands {
		if cmd.ID == id {
			return &cmd, true
		}
	}
	return nil, false
}

func (s *Service) GetAllCommands() []model.Command {
	return config.GetCommandsV2()
}

func (s *Service) GetPlatformCommand(cmd *model.Command) (string, bool) {
	platformData, ok := cmd.Platforms[runtime.GOOS]
	if !ok {
		return "", false
	}

	switch v := platformData.(type) {
	case string:
		return v, true
	case map[string]interface{}:
		if command, ok := v["command"].(string); ok {
			return command, true
		}
	}
	return "", false
}

func (s *Service) ValidateCommand(id string) error {
	cmd, ok := s.GetCommand(id)
	if !ok {
		return fmt.Errorf("command not found: %s", id)
	}

	// 检查命令白名单
	if !cmd.IsWhitelisted() {
		return fmt.Errorf("command not whitelisted: %s", id)
	}

	// 检查全局白名单
	if s.config.Security.EnableWhitelist {
		if !s.isCommandAllowed(id) {
			return fmt.Errorf("command not allowed: %s", id)
		}
	}

	return nil
}

func (s *Service) isCommandAllowed(id string) bool {
	if len(s.config.Security.AllowedCommands) == 0 {
		return true // 空白名单表示允许所有命令
	}

	for _, allowed := range s.config.Security.AllowedCommands {
		if allowed == id {
			return true
		}
	}
	return false
}

func (s *Service) ReloadCommands() error {
	return config.ReloadCommands()
}

// GetCommandTimeout 获取命令超时时间
func (s *Service) GetCommandTimeout(id string) int {
	if cmd, ok := s.GetCommand(id); ok {
		return cmd.GetTimeout()
	}
	return 10000 // 默认超时
}

// RequiresPin 检查命令是否需要PIN验证
func (s *Service) RequiresPin(id string) bool {
	if cmd, ok := s.GetCommand(id); ok {
		return cmd.RequiresPin()
	}
	return s.config.Security.PinRequired // 回退到全局配置
}

// GetCommandInfo 获取命令详细信息
func (s *Service) GetCommandInfo(id string) map[string]interface{} {
	info := make(map[string]interface{})
	info["id"] = id
	info["version"] = config.GetCommandsVersion()
	
	if cmd, ok := s.GetCommand(id); ok {
		info["name"] = cmd.Name
		info["description"] = cmd.Description
		info["category"] = cmd.Category
		info["icon"] = cmd.Icon
		info["timeout"] = cmd.GetTimeout()
		info["requiresPin"] = cmd.RequiresPin()
		info["whitelisted"] = cmd.IsWhitelisted()
		info["available"] = cmd.IsWhitelisted() // 前端需要的available字段
		
		// 添加首页相关信息
		info["showOnHomepage"] = cmd.ShowOnHomepage()
		if cmd.Homepage != nil {
			x, y, width, height := cmd.GetHomepagePosition()
			info["homepagePosition"] = map[string]interface{}{
				"x": x,
				"y": y, 
				"width": width,
				"height": height,
			}
			info["homepageColor"] = cmd.GetHomepageColor()
		}
	}
	
	return info
}