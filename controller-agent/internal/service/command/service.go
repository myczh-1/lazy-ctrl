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

func (s *Service) GetCommand(id string) (interface{}, bool) {
	if config.IsV2Commands() {
		// 新格式命令
		commands := config.GetCommandsV2()
		for _, cmd := range commands {
			if cmd.ID == id {
				return cmd, true
			}
		}
		return nil, false
	}
	// 旧格式命令
	commands := config.GetCommands()
	cmd, ok := commands[id]
	return cmd, ok
}

// GetCommandV2 获取新格式命令
func (s *Service) GetCommandV2(id string) (*model.Command, bool) {
	if !config.IsV2Commands() {
		return nil, false
	}
	commands := config.GetCommandsV2()
	for i, cmd := range commands {
		if cmd.ID == id {
			return &commands[i], true
		}
	}
	return nil, false
}

func (s *Service) GetAllCommands() map[string]interface{} {
	return config.GetCommands()
}

func (s *Service) GetPlatformCommand(cmd interface{}) (string, bool) {
	switch v := cmd.(type) {
	case model.Command:
		// 新格式命令
		return s.getPlatformCommandV2(&v)
	case *model.Command:
		// 新格式命令指针
		return s.getPlatformCommandV2(v)
	case string:
		// 简单字符串命令，直接返回
		return v, true
	case map[string]interface{}:
		// 旧格式平台特定命令
		if platformCmd, ok := v[runtime.GOOS]; ok {
			if cmdStr, ok := platformCmd.(string); ok {
				return cmdStr, true
			}
		}
	}
	return "", false
}

// getPlatformCommandV2 从新格式命令中获取平台特定命令
func (s *Service) getPlatformCommandV2(cmd *model.Command) (string, bool) {
	platformData, ok := cmd.Platforms[runtime.GOOS]
	if !ok {
		return "", false
	}

	switch v := platformData.(type) {
	case string:
		// 简单字符串格式 (兼容旧版)
		return v, true
	case map[string]interface{}:
		// 结构化格式
		if command, ok := v["command"].(string); ok {
			return command, true
		}
		// 检查是否是复杂命令序列
		if _, hasCommands := v["commands"]; hasCommands {
			// TODO: 将来实现复杂命令序列支持
			return "", false
		}
	}
	return "", false
}

func (s *Service) ValidateCommand(id string) error {
	// 检查命令是否存在
	_, ok := s.GetCommand(id)
	if !ok {
		return fmt.Errorf("command not found: %s", id)
	}

	// 新格式命令的额外验证
	if config.IsV2Commands() {
		if cmd, ok := s.GetCommandV2(id); ok {
			// 检查命令白名单
			if !cmd.IsWhitelisted() {
				return fmt.Errorf("command not whitelisted: %s", id)
			}
		}
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
	if config.IsV2Commands() {
		if cmd, ok := s.GetCommandV2(id); ok {
			return cmd.GetTimeout()
		}
	}
	return 10000 // 默认超时
}

// RequiresPin 检查命令是否需要PIN验证
func (s *Service) RequiresPin(id string) bool {
	if config.IsV2Commands() {
		if cmd, ok := s.GetCommandV2(id); ok {
			return cmd.RequiresPin()
		}
	}
	return s.config.Security.PinRequired // 回退到全局配置
}

// GetCommandInfo 获取命令详细信息
func (s *Service) GetCommandInfo(id string) map[string]interface{} {
	info := make(map[string]interface{})
	info["id"] = id
	info["version"] = config.GetCommandsVersion()
	
	if config.IsV2Commands() {
		if cmd, ok := s.GetCommandV2(id); ok {
			info["name"] = cmd.Name
			info["description"] = cmd.Description
			info["category"] = cmd.Category
			info["icon"] = cmd.Icon
			info["timeout"] = cmd.GetTimeout()
			info["requiresPin"] = cmd.RequiresPin()
			info["whitelisted"] = cmd.IsWhitelisted()
		}
	} else {
		info["name"] = id
		info["timeout"] = 10000
		info["requiresPin"] = s.config.Security.PinRequired
	}
	
	return info
}