package command

import (
	"fmt"
	"runtime"

	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
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
	commands := config.GetCommands()
	cmd, ok := commands[id]
	return cmd, ok
}

func (s *Service) GetAllCommands() map[string]interface{} {
	return config.GetCommands()
}

func (s *Service) GetPlatformCommand(cmd interface{}) (string, bool) {
	switch v := cmd.(type) {
	case string:
		// 简单字符串命令，直接返回
		return v, true
	case map[string]interface{}:
		// 平台特定命令，根据当前平台选择
		if platformCmd, ok := v[runtime.GOOS]; ok {
			if cmdStr, ok := platformCmd.(string); ok {
				return cmdStr, true
			}
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

	// 检查白名单
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