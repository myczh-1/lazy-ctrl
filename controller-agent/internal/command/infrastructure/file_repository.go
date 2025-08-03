package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"

	"github.com/myczh-1/lazy-ctrl-agent/internal/command/entity"
	"github.com/myczh-1/lazy-ctrl-agent/internal/command/repository"
)

// FileCommandRepository implements CommandRepository using file storage
type FileCommandRepository struct {
	configPath string
	commands   map[string]*entity.Command
	version    string
	mu         sync.RWMutex // 保护并发访问
}

// CommandConfig represents the JSON structure of command configuration file
type CommandConfig struct {
	Version  string           `json:"version"`
	Commands []*entity.Command `json:"commands"`
}

// NewFileCommandRepository creates a new file-based command repository
func NewFileCommandRepository(configPath string) repository.CommandRepository {
	return &FileCommandRepository{
		configPath: configPath,
		commands:   make(map[string]*entity.Command),
		version:    "3.0",
	}
}

// Create creates a new command
func (r *FileCommandRepository) Create(ctx context.Context, command *entity.Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if command already exists
	if _, exists := r.commands[command.ID]; exists {
		return fmt.Errorf("command with ID %s already exists", command.ID)
	}
	
	// Add command to memory
	r.commands[command.ID] = command
	
	// Save to file
	return r.saveToFile()
}

// GetByID retrieves a command by its ID
func (r *FileCommandRepository) GetByID(ctx context.Context, id string) (*entity.Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	command, exists := r.commands[id]
	if !exists {
		return nil, fmt.Errorf("command not found: %s", id)
	}
	
	// Return a copy to prevent external modification
	return r.copyCommand(command), nil
}

// GetAll retrieves all commands
func (r *FileCommandRepository) GetAll(ctx context.Context) ([]*entity.Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	commands := make([]*entity.Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, r.copyCommand(cmd))
	}
	return commands, nil
}

// Update updates an existing command
func (r *FileCommandRepository) Update(ctx context.Context, command *entity.Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if command exists
	if _, exists := r.commands[command.ID]; !exists {
		return fmt.Errorf("command not found: %s", command.ID)
	}
	
	// Update command in memory
	r.commands[command.ID] = command
	
	// Save to file
	return r.saveToFile()
}

// Delete deletes a command by ID
func (r *FileCommandRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Check if command exists
	if _, exists := r.commands[id]; !exists {
		return fmt.Errorf("command not found: %s", id)
	}
	
	// Delete from memory
	delete(r.commands, id)
	
	// Save to file
	return r.saveToFile()
}

// GetByUserID retrieves commands for a specific user
func (r *FileCommandRepository) GetByUserID(ctx context.Context, userID string) ([]*entity.Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var commands []*entity.Command
	for _, cmd := range r.commands {
		if cmd.UserID == userID {
			commands = append(commands, r.copyCommand(cmd))
		}
	}
	return commands, nil
}

// GetByCategory retrieves commands by category
func (r *FileCommandRepository) GetByCategory(ctx context.Context, category string) ([]*entity.Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var commands []*entity.Command
	for _, cmd := range r.commands {
		if cmd.Category == category {
			commands = append(commands, r.copyCommand(cmd))
		}
	}
	return commands, nil
}

// GetHomepageCommands retrieves commands that should be displayed on homepage
func (r *FileCommandRepository) GetHomepageCommands(ctx context.Context) ([]*entity.Command, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var commands []*entity.Command
	for _, cmd := range r.commands {
		if cmd.ShowOnHomepage() {
			commands = append(commands, r.copyCommand(cmd))
		}
	}
	return commands, nil
}

// Exists checks if a command with the given ID exists
func (r *FileCommandRepository) Exists(ctx context.Context, id string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.commands[id]
	return exists, nil
}

// Reload reloads the command configuration from storage
func (r *FileCommandRepository) Reload(ctx context.Context) error {
	return r.loadFromFile()
}

// Initialize loads commands from file on startup
func (r *FileCommandRepository) Initialize() error {
	return r.loadFromFile()
}

// loadFromFile loads commands from the configuration file
func (r *FileCommandRepository) loadFromFile() error {
	// Resolve absolute path
	configPath := r.configPath
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(".", configPath)
	}
	
	// Read JSON file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read commands file: %w", err)
	}
	
	// Parse command configuration
	var config struct {
		Version  string `json:"version"`
		Commands []struct {
			ID             string                 `json:"id"`
			Name           string                 `json:"name,omitempty"`
			Description    string                 `json:"description,omitempty"`
			Category       string                 `json:"category,omitempty"`
			Icon           string                 `json:"icon,omitempty"`
			Command        string                 `json:"command"`
			Platform       string                 `json:"platform"`
			CommandType    string                 `json:"commandType,omitempty"`
			Security       *entity.SecurityConfig `json:"security,omitempty"`
			Timeout        int                    `json:"timeout,omitempty"`
			UserID         string                 `json:"userId,omitempty"`
			DeviceID       string                 `json:"deviceId,omitempty"`
			HomeLayout     *entity.HomeLayoutConfig `json:"homeLayout,omitempty"`
			TemplateId     string                 `json:"templateId,omitempty"`
			TemplateParams map[string]interface{} `json:"templateParams,omitempty"`
			CreatedAt      string                 `json:"createdAt,omitempty"`
			UpdatedAt      string                 `json:"updatedAt,omitempty"`
		} `json:"commands"`
	}
	
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse commands config: %w", err)
	}
	
	if config.Version == "" {
		return fmt.Errorf("missing version in commands config")
	}
	
	// Convert to entity commands
	r.mu.Lock()
	r.commands = make(map[string]*entity.Command)
	r.version = config.Version
	
	for _, cmdData := range config.Commands {
		cmd := &entity.Command{
			ID:             cmdData.ID,
			Name:           cmdData.Name,
			Description:    cmdData.Description,
			Category:       cmdData.Category,
			Icon:           cmdData.Icon,
			Command:        cmdData.Command,
			Platform:       cmdData.Platform,
			CommandType:    cmdData.CommandType,
			Security:       cmdData.Security,
			Timeout:        cmdData.Timeout,
			UserID:         cmdData.UserID,
			DeviceID:       cmdData.DeviceID,
			HomeLayout:     cmdData.HomeLayout,
			TemplateId:     cmdData.TemplateId,
			TemplateParams: cmdData.TemplateParams,
		}
		
		// Parse timestamps
		if cmdData.CreatedAt != "" {
			if t, err := time.Parse(time.RFC3339, cmdData.CreatedAt); err == nil {
				cmd.CreatedAt = t
			} else {
				cmd.CreatedAt = time.Now()
			}
		} else {
			cmd.CreatedAt = time.Now()
		}
		
		if cmdData.UpdatedAt != "" {
			if t, err := time.Parse(time.RFC3339, cmdData.UpdatedAt); err == nil {
				cmd.UpdatedAt = t
			} else {
				cmd.UpdatedAt = time.Now()
			}
		} else {
			cmd.UpdatedAt = time.Now()
		}
		
		r.commands[cmd.ID] = cmd
	}
	r.mu.Unlock()
	
	return nil
}

// saveToFile saves current commands to the configuration file
func (r *FileCommandRepository) saveToFile() error {
	// Convert entities to JSON structure
	commands := make([]map[string]interface{}, 0, len(r.commands))
	
	for _, cmd := range r.commands {
		cmdData := map[string]interface{}{
			"id":          cmd.ID,
			"name":        cmd.Name,
			"description": cmd.Description,
			"category":    cmd.Category,
			"icon":        cmd.Icon,
			"command":     cmd.Command,
			"platform":    cmd.Platform,
			"createdAt":   cmd.CreatedAt.Format(time.RFC3339),
			"updatedAt":   cmd.UpdatedAt.Format(time.RFC3339),
		}
		
		// Add optional fields
		if cmd.CommandType != "" {
			cmdData["commandType"] = cmd.CommandType
		}
		if cmd.Timeout > 0 {
			cmdData["timeout"] = cmd.Timeout
		}
		if cmd.UserID != "" {
			cmdData["userId"] = cmd.UserID
		}
		if cmd.DeviceID != "" {
			cmdData["deviceId"] = cmd.DeviceID
		}
		if cmd.TemplateId != "" {
			cmdData["templateId"] = cmd.TemplateId
		}
		if cmd.TemplateParams != nil {
			cmdData["templateParams"] = cmd.TemplateParams
		}
		if cmd.Security != nil {
			cmdData["security"] = cmd.Security
		}
		if cmd.HomeLayout != nil {
			cmdData["homeLayout"] = cmd.HomeLayout
		}
		
		commands = append(commands, cmdData)
	}
	
	// Create config structure
	config := map[string]interface{}{
		"version":  r.version,
		"commands": commands,
	}
	
	// Marshal to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}
	
	// Resolve absolute path
	configPath := r.configPath
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(".", configPath)
	}
	
	// Write to file
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write commands file: %w", err)
	}
	
	return nil
}

// copyCommand creates a deep copy of a command entity
func (r *FileCommandRepository) copyCommand(cmd *entity.Command) *entity.Command {
	newCmd := &entity.Command{
		ID:             cmd.ID,
		Name:           cmd.Name,
		Description:    cmd.Description,
		Category:       cmd.Category,
		Icon:           cmd.Icon,
		Command:        cmd.Command,
		Platform:       cmd.Platform,
		CommandType:    cmd.CommandType,
		Timeout:        cmd.Timeout,
		UserID:         cmd.UserID,
		DeviceID:       cmd.DeviceID,
		TemplateId:     cmd.TemplateId,
		CreatedAt:      cmd.CreatedAt,
		UpdatedAt:      cmd.UpdatedAt,
	}
	
	// Deep copy Security
	if cmd.Security != nil {
		newCmd.Security = &entity.SecurityConfig{
			RequirePin: cmd.Security.RequirePin,
			Whitelist:  cmd.Security.Whitelist,
			AdminOnly:  cmd.Security.AdminOnly,
		}
	}
	
	// Deep copy HomeLayout
	if cmd.HomeLayout != nil {
		newCmd.HomeLayout = &entity.HomeLayoutConfig{
			ShowOnHome: cmd.HomeLayout.ShowOnHome,
			Color:      cmd.HomeLayout.Color,
			Priority:   cmd.HomeLayout.Priority,
		}
		if cmd.HomeLayout.DefaultPosition != nil {
			newCmd.HomeLayout.DefaultPosition = &entity.PositionConfig{
				X:      cmd.HomeLayout.DefaultPosition.X,
				Y:      cmd.HomeLayout.DefaultPosition.Y,
				Width:  cmd.HomeLayout.DefaultPosition.Width,
				Height: cmd.HomeLayout.DefaultPosition.Height,
			}
		}
	}
	
	// Deep copy TemplateParams
	if cmd.TemplateParams != nil {
		newCmd.TemplateParams = make(map[string]interface{})
		for k, v := range cmd.TemplateParams {
			newCmd.TemplateParams[k] = v
		}
	}
	
	return newCmd
}