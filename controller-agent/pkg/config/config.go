package config

import (
	"encoding/json"
	"os"
)

type Command struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	ScriptPath  string            `json:"script_path"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	WorkDir     string            `json:"work_dir,omitempty"`
}

type SecurityConfig struct {
	AllowedPaths []string `json:"allowed_paths"`
	Whitelist    []string `json:"whitelist"`
	RequireAuth  bool     `json:"require_auth"`
}

type Config struct {
	Commands []Command      `json:"commands"`
	Security SecurityConfig `json:"security"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) SaveConfig(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (c *Config) GetCommandByID(id string) *Command {
	for _, cmd := range c.Commands {
		if cmd.ID == id {
			return &cmd
		}
	}
	return nil
}