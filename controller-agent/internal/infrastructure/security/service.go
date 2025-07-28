package security

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/myczh-1/lazy-ctrl-agent/internal/config"
	"github.com/sirupsen/logrus"
)

type Service struct {
	config      *config.Config
	logger      *logrus.Logger
	rateLimiter map[string]*rateLimitEntry
	mutex       sync.RWMutex
}

type rateLimitEntry struct {
	count     int
	resetTime time.Time
}

func NewService(config *config.Config, logger *logrus.Logger) *Service {
	return &Service{
		config:      config,
		logger:      logger,
		rateLimiter: make(map[string]*rateLimitEntry),
	}
}

func (s *Service) ValidatePin(providedPin string) bool {
	if !s.config.Security.PinRequired {
		return true
	}

	if s.config.Security.Pin == "" {
		s.logger.Warn("PIN validation enabled but no PIN configured")
		return false
	}

	// 使用SHA256哈希比较PIN
	hasher := sha256.New()
	hasher.Write([]byte(providedPin))
	providedHash := fmt.Sprintf("%x", hasher.Sum(nil))

	hasher.Reset()
	hasher.Write([]byte(s.config.Security.Pin))
	configHash := fmt.Sprintf("%x", hasher.Sum(nil))

	return providedHash == configHash
}

func (s *Service) CheckRateLimit(clientID string) error {
	if !s.config.Security.RateLimitEnabled {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	entry, exists := s.rateLimiter[clientID]

	if !exists || now.After(entry.resetTime) {
		// 新客户端或者重置时间已过
		s.rateLimiter[clientID] = &rateLimitEntry{
			count:     1,
			resetTime: now.Add(time.Minute),
		}
		return nil
	}

	if entry.count >= s.config.Security.RateLimitPerMin {
		s.logger.WithFields(logrus.Fields{
			"client_id": clientID,
			"count":     entry.count,
			"limit":     s.config.Security.RateLimitPerMin,
		}).Warn("Rate limit exceeded")
		
		return fmt.Errorf("rate limit exceeded: %d requests per minute", s.config.Security.RateLimitPerMin)
	}

	entry.count++
	return nil
}

func (s *Service) ValidateCommandAccess(commandID string) error {
	if !s.config.Security.EnableWhitelist {
		return nil
	}

	if len(s.config.Security.AllowedCommands) == 0 {
		return nil // 空白名单表示允许所有命令
	}

	for _, allowed := range s.config.Security.AllowedCommands {
		if allowed == commandID {
			return nil
		}
	}

	s.logger.WithFields(logrus.Fields{
		"command_id":      commandID,
		"allowed_commands": s.config.Security.AllowedCommands,
	}).Warn("Command not in whitelist")

	return fmt.Errorf("command not allowed: %s", commandID)
}

func (s *Service) GetClientID(remoteAddr string, userAgent string) string {
	// 生成客户端标识，用于限流
	identifier := fmt.Sprintf("%s:%s", remoteAddr, userAgent)
	hasher := sha256.New()
	hasher.Write([]byte(identifier))
	return fmt.Sprintf("%x", hasher.Sum(nil))[:16]
}

func (s *Service) CleanupRateLimiter() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for clientID, entry := range s.rateLimiter {
		if now.After(entry.resetTime) {
			delete(s.rateLimiter, clientID)
		}
	}
}