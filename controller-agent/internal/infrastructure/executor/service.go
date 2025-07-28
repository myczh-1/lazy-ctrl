package executor

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type Service struct {
	logger *logrus.Logger
}

type ExecutionResult struct {
	Success       bool          `json:"success"`
	Output        string        `json:"output"`
	Error         string        `json:"error"`
	ExitCode      int           `json:"exit_code"`
	ExecutionTime time.Duration `json:"execution_time"`
}

func NewService(logger *logrus.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) Execute(ctx context.Context, command string) (*ExecutionResult, error) {
	startTime := time.Now()
	
	s.logger.WithFields(logrus.Fields{
		"command":  command,
		"platform": runtime.GOOS,
	}).Info("Executing command")

	cmd := s.prepareCommand(ctx, command)
	
	output, err := cmd.CombinedOutput()
	executionTime := time.Since(startTime)
	
	result := &ExecutionResult{
		Success:       err == nil,
		Output:        string(output),
		ExecutionTime: executionTime,
	}

	if err != nil {
		result.Error = err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		
		s.logger.WithFields(logrus.Fields{
			"command":        command,
			"error":          err.Error(),
			"output":         string(output),
			"execution_time": executionTime,
			"exit_code":      result.ExitCode,
		}).Error("Command execution failed")
	} else {
		s.logger.WithFields(logrus.Fields{
			"command":        command,
			"execution_time": executionTime,
		}).Info("Command executed successfully")
	}

	return result, nil
}

func (s *Service) ExecuteWithTimeout(command string, timeout time.Duration) (*ExecutionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	return s.Execute(ctx, command)
}

func (s *Service) prepareCommand(ctx context.Context, command string) *exec.Cmd {
	var cmd *exec.Cmd
	
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(command, "powershell") {
			// 解析PowerShell命令参数
			parts := strings.SplitN(command, " ", 3)
			if len(parts) >= 3 && parts[1] == "-c" {
				// 去掉外层引号
				script := strings.Trim(parts[2], "\"")
				cmd = exec.CommandContext(ctx, "powershell", "-Command", script)
			} else {
				cmd = exec.CommandContext(ctx, "cmd", "/C", command)
			}
		} else {
			cmd = exec.CommandContext(ctx, "cmd", "/C", command)
		}
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}
	
	return cmd
}

func (s *Service) ValidateCommand(command string) error {
	if strings.TrimSpace(command) == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// 基本安全检查 - 防止危险命令
	dangerousPatterns := []string{
		"rm -rf /",
		"del /s /q C:\\",
		"format c:",
		"mkfs.",
		"fdisk",
	}

	lowerCmd := strings.ToLower(command)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerCmd, strings.ToLower(pattern)) {
			return fmt.Errorf("potentially dangerous command detected: %s", pattern)
		}
	}

	return nil
}

