package repository

import (
	"context"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/entity"
)

// CommandRepository defines the contract for command persistence
type CommandRepository interface {
	// Create creates a new command
	Create(ctx context.Context, command *entity.Command) error
	
	// GetByID retrieves a command by its ID
	GetByID(ctx context.Context, id string) (*entity.Command, error)
	
	// GetAll retrieves all commands
	GetAll(ctx context.Context) ([]*entity.Command, error)
	
	// Update updates an existing command
	Update(ctx context.Context, command *entity.Command) error
	
	// Delete deletes a command by ID
	Delete(ctx context.Context, id string) error
	
	// GetByUserID retrieves commands for a specific user
	GetByUserID(ctx context.Context, userID string) ([]*entity.Command, error)
	
	// GetByCategory retrieves commands by category
	GetByCategory(ctx context.Context, category string) ([]*entity.Command, error)
	
	// GetHomepageCommands retrieves commands that should be displayed on homepage
	GetHomepageCommands(ctx context.Context) ([]*entity.Command, error)
	
	// Exists checks if a command with the given ID exists
	Exists(ctx context.Context, id string) (bool, error)
	
	// Reload reloads the command configuration from storage
	Reload(ctx context.Context) error
}