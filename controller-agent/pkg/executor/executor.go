package executor

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"lazy-ctrl/controller-agent/pkg/config"
)

type ExecutionResult struct {
	Success    bool   `json:"success"`
	Output     string `json:"output"`
	Error      string `json:"error"`
	ExitCode   int    `json:"exit_code"`
	ExecutedAt string `json:"executed_at"`
}

type Executor struct {
	config *config.Config
}

func NewExecutor(cfg *config.Config) *Executor {
	return &Executor{
		config: cfg,
	}
}

func (e *Executor) ExecuteCommand(ctx context.Context, commandID string, args []string) *ExecutionResult {
	command := e.config.GetCommandByID(commandID)
	if command == nil {
		return &ExecutionResult{
			Success:    false,
			Error:      "Command not found: " + commandID,
			ExecutedAt: time.Now().Format(time.RFC3339),
		}
	}

	if !e.isPathAllowed(command.ScriptPath) {
		return &ExecutionResult{
			Success:    false,
			Error:      "Script path not allowed: " + command.ScriptPath,
			ExecutedAt: time.Now().Format(time.RFC3339),
		}
	}

	allArgs := append(command.Args, args...)
	cmd := exec.CommandContext(ctx, command.ScriptPath, allArgs...)

	if command.WorkDir != "" {
		cmd.Dir = command.WorkDir
	}

	if len(command.Env) > 0 {
		for k, v := range command.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &ExecutionResult{
		Success:    err == nil,
		Output:     stdout.String(),
		Error:      stderr.String(),
		ExecutedAt: time.Now().Format(time.RFC3339),
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		if result.Error == "" {
			result.Error = err.Error()
		}
	}

	return result
}

func (e *Executor) isPathAllowed(path string) bool {
	if len(e.config.Security.AllowedPaths) == 0 {
		return true
	}

	for _, allowedPath := range e.config.Security.AllowedPaths {
		if strings.HasPrefix(path, allowedPath) {
			return true
		}
	}

	return false
}