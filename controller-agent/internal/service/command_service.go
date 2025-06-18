package service

import (
	"context"

	"lazy-ctrl/controller-agent/pkg/config"
	"lazy-ctrl/controller-agent/pkg/executor"
)

type CommandService struct {
	config   *config.Config
	executor *executor.Executor
}

func NewCommandService(cfg *config.Config) *CommandService {
	return &CommandService{
		config:   cfg,
		executor: executor.NewExecutor(cfg),
	}
}

func (s *CommandService) ExecuteCommand(ctx context.Context, commandID string, args []string) *executor.ExecutionResult {
	return s.executor.ExecuteCommand(ctx, commandID, args)
}

func (s *CommandService) ListCommands() []config.Command {
	return s.config.Commands
}

func (s *CommandService) GetCommand(id string) *config.Command {
	return s.config.GetCommandByID(id)
}