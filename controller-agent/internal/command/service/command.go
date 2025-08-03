package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/myczh-1/lazy-ctrl-agent/internal/command/entity"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/repository"
)

// CommandService provides business logic for command operations
type CommandService struct {
	repo repository.CommandRepository
}

// NewCommandService creates a new CommandService
func NewCommandService(repo repository.CommandRepository) *CommandService {
	return &CommandService{
		repo: repo,
	}
}

// CreateCommand creates a new command with validation
func (s *CommandService) CreateCommand(ctx context.Context, id, name, command string) (*entity.Command, error) {
	if id == "" {
		return nil, fmt.Errorf("command ID is required")
	}
	
	if command == "" {
		return nil, fmt.Errorf("command string is required")
	}
	
	// Check if command already exists
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check command existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("command with ID %s already exists", id)
	}
	
	// Create new command entity
	cmd := entity.NewCommand(id, name, command)
	
	// Save to repository
	if err := s.repo.Create(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to create command: %w", err)
	}
	
	return cmd, nil
}

// GetCommand retrieves a command by ID
func (s *CommandService) GetCommand(ctx context.Context, id string) (*entity.Command, error) {
	if id == "" {
		return nil, fmt.Errorf("command ID is required")
	}
	
	cmd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get command: %w", err)
	}
	
	return cmd, nil
}

// GetAllCommands retrieves all commands
func (s *CommandService) GetAllCommands(ctx context.Context) ([]*entity.Command, error) {
	commands, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all commands: %w", err)
	}
	
	return commands, nil
}

// UpdateCommand updates an existing command
func (s *CommandService) UpdateCommand(ctx context.Context, id, name, description, command string) (*entity.Command, error) {
	if id == "" {
		return nil, fmt.Errorf("command ID is required")
	}
	
	// Get existing command
	cmd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get command: %w", err)
	}
	
	// Update command
	cmd.Update(name, description, command)
	
	// Save updated command
	if err := s.repo.Update(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to update command: %w", err)
	}
	
	return cmd, nil
}

// UpdateCommandWithFields updates a command with multiple fields
func (s *CommandService) UpdateCommandWithFields(ctx context.Context, id string, updates map[string]interface{}) (*entity.Command, error) {
	if id == "" {
		return nil, fmt.Errorf("command ID is required")
	}
	
	// Get existing command
	cmd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get command: %w", err)
	}
	
	// Update command fields
	cmd.UpdateFields(updates)
	
	// Handle security updates separately
	if securityData, ok := updates["security"]; ok {
		if securityMap, ok := securityData.(map[string]interface{}); ok {
			requirePin, _ := securityMap["requirePin"].(bool)
			whitelist, _ := securityMap["whitelist"].(bool)
			adminOnly, _ := securityMap["adminOnly"].(bool)
			cmd.SetSecurity(requirePin, whitelist, adminOnly)
		}
	}
	
	// Handle home layout updates separately
	if homeLayoutData, ok := updates["homeLayout"]; ok {
		if homeLayoutMap, ok := homeLayoutData.(map[string]interface{}); ok {
			showOnHome, _ := homeLayoutMap["showOnHome"].(bool)
			color, _ := homeLayoutMap["color"].(string)
			priority, _ := homeLayoutMap["priority"].(int)
			
			var position *entity.PositionConfig
			if posData, ok := homeLayoutMap["defaultPosition"]; ok {
				if posMap, ok := posData.(map[string]interface{}); ok {
					x, _ := posMap["x"].(int)
					y, _ := posMap["y"].(int)
					w, _ := posMap["w"].(int)
					h, _ := posMap["h"].(int)
					position = &entity.PositionConfig{X: x, Y: y, Width: w, Height: h}
				}
			}
			
			cmd.SetHomeLayout(showOnHome, position, color, priority)
		}
	}
	
	// Save updated command
	if err := s.repo.Update(ctx, cmd); err != nil {
		return nil, fmt.Errorf("failed to update command: %w", err)
	}
	
	return cmd, nil
}

// DeleteCommand deletes a command by ID
func (s *CommandService) DeleteCommand(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("command ID is required")
	}
	
	// Check if command exists
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check command existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("command with ID %s not found", id)
	}
	
	// Delete command
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete command: %w", err)
	}
	
	return nil
}

// GetPlatformCommand returns the command string for current platform
func (s *CommandService) GetPlatformCommand(ctx context.Context, id string) (string, error) {
	cmd, err := s.GetCommand(ctx, id)
	if err != nil {
		return "", err
	}
	
	// Check if command is available on current platform
	if !cmd.IsAvailableOnPlatform() {
		return "", fmt.Errorf("command %s not available on platform %s", id, runtime.GOOS)
	}
	
	return cmd.Command, nil
}

// ValidateCommand validates if a command can be executed
func (s *CommandService) ValidateCommand(ctx context.Context, id string, allowedCommands []string, enableWhitelist bool) error {
	cmd, err := s.GetCommand(ctx, id)
	if err != nil {
		return err
	}
	
	// Check command whitelist
	if !cmd.IsWhitelisted() {
		return fmt.Errorf("command not whitelisted: %s", id)
	}
	
	// Check global whitelist
	if enableWhitelist {
		if !s.isCommandAllowed(id, allowedCommands) {
			return fmt.Errorf("command not allowed: %s", id)
		}
	}
	
	// Check platform availability
	if !cmd.IsAvailableOnPlatform() {
		return fmt.Errorf("command not available on current platform: %s", id)
	}
	
	return nil
}

// GetHomepageCommands retrieves commands for homepage display
func (s *CommandService) GetHomepageCommands(ctx context.Context) ([]*entity.Command, error) {
	commands, err := s.repo.GetHomepageCommands(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get homepage commands: %w", err)
	}
	
	return commands, nil
}

// GetCommandsByCategory retrieves commands by category
func (s *CommandService) GetCommandsByCategory(ctx context.Context, category string) ([]*entity.Command, error) {
	commands, err := s.repo.GetByCategory(ctx, category)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands by category: %w", err)
	}
	
	return commands, nil
}

// GetCommandsByUser retrieves commands for a specific user
func (s *CommandService) GetCommandsByUser(ctx context.Context, userID string) ([]*entity.Command, error) {
	commands, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands by user: %w", err)
	}
	
	return commands, nil
}

// ReloadCommands reloads command configuration
func (s *CommandService) ReloadCommands(ctx context.Context) error {
	if err := s.repo.Reload(ctx); err != nil {
		return fmt.Errorf("failed to reload commands: %w", err)
	}
	
	return nil
}

// GetCommandInfo returns detailed command information
func (s *CommandService) GetCommandInfo(ctx context.Context, id string) (map[string]interface{}, error) {
	cmd, err := s.GetCommand(ctx, id)
	if err != nil {
		return nil, err
	}
	
	info := map[string]interface{}{
		"id":          cmd.ID,
		"name":        cmd.Name,
		"description": cmd.Description,
		"category":    cmd.Category,
		"icon":        cmd.Icon,
		"command":     cmd.Command,
		"platform":    cmd.Platform,
		"timeout":     cmd.GetTimeout(),
		"requiresPin": cmd.RequiresPin(),
		"whitelisted": cmd.IsWhitelisted(),
		"available":   cmd.IsAvailableOnPlatform(),
		"createdAt":   cmd.CreatedAt.Format(time.RFC3339),
		"updatedAt":   cmd.UpdatedAt.Format(time.RFC3339),
	}
	
	// Add homepage layout information
	if cmd.ShowOnHomepage() {
		x, y, width, height := cmd.GetHomepagePosition()
		info["showOnHomepage"] = true
		info["homepagePosition"] = map[string]interface{}{
			"x":      x,
			"y":      y,
			"width":  width,
			"height": height,
		}
		info["homepageColor"] = cmd.GetHomepageColor()
		info["homepagePriority"] = cmd.GetHomepagePriority()
	} else {
		info["showOnHomepage"] = false
	}
	
	return info, nil
}

// isCommandAllowed checks if command is in allowed list
func (s *CommandService) isCommandAllowed(id string, allowedCommands []string) bool {
	if len(allowedCommands) == 0 {
		return true // Empty whitelist means allow all
	}
	
	for _, allowed := range allowedCommands {
		if allowed == id {
			return true
		}
	}
	return false
}