package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/myczh-1/lazy-ctrl-agent/internal/model"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Security SecurityConfig `mapstructure:"security"`
	Commands CommandsConfig `mapstructure:"commands"`
	MQTT     MQTTConfig     `mapstructure:"mqtt"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
	GRPC GRPCConfig `mapstructure:"grpc"`
}

type HTTPConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	StaticPath string `mapstructure:"static_path"`
	StaticDir  string `mapstructure:"static_dir"`
}

type GRPCConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
}

type SecurityConfig struct {
	EnableWhitelist   bool     `mapstructure:"enable_whitelist"`
	PinRequired       bool     `mapstructure:"pin_required"`
	Pin               string   `mapstructure:"pin"`
	RateLimitEnabled  bool     `mapstructure:"rate_limit_enabled"`
	RateLimitPerMin   int      `mapstructure:"rate_limit_per_min"`
	AllowedCommands   []string `mapstructure:"allowed_commands"`
}

type CommandsConfig struct {
	ConfigPath string          `mapstructure:"config_path"`
	Commands   []model.Command `mapstructure:"-"`
	HotReload  bool            `mapstructure:"hot_reload"`
	Version    string          `mapstructure:"-"`
}

type MQTTConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Broker     string `mapstructure:"broker"`
	Port       int    `mapstructure:"port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	ClientID   string `mapstructure:"client_id"`
	TopicBase  string `mapstructure:"topic_base"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

var globalConfig *Config

func Load(configPath string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.AddConfigPath("./config")
		viper.AddConfigPath(".")
	}

	// 设置默认值
	setDefaults()

	// 读取环境变量
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置
			fmt.Println("Config file not found, using defaults")
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	globalConfig = &Config{}
	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	// 加载命令配置
	if err := loadCommands(); err != nil {
		return fmt.Errorf("error loading commands: %w", err)
	}

	return nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.http.enabled", true)
	viper.SetDefault("server.http.host", "0.0.0.0")
	viper.SetDefault("server.http.port", 7070)
	viper.SetDefault("server.http.static_path", "")
	viper.SetDefault("server.http.static_dir", "")
	viper.SetDefault("server.grpc.enabled", true)
	viper.SetDefault("server.grpc.host", "0.0.0.0")
	viper.SetDefault("server.grpc.port", 7071)

	// Security defaults
	viper.SetDefault("security.enable_whitelist", true)
	viper.SetDefault("security.pin_required", false)
	viper.SetDefault("security.rate_limit_enabled", true)
	viper.SetDefault("security.rate_limit_per_min", 60)

	// Commands defaults
	viper.SetDefault("commands.config_path", "config/commands.json")
	viper.SetDefault("commands.hot_reload", true)

	// MQTT defaults
	viper.SetDefault("mqtt.enabled", false)
	viper.SetDefault("mqtt.port", 1883)
	viper.SetDefault("mqtt.client_id", "lazy-ctrl-agent")
	viper.SetDefault("mqtt.topic_base", "lazy-ctrl")

	// Log defaults
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output_path", "")
}

func loadCommands() error {
	commandsPath := globalConfig.Commands.ConfigPath
	if !filepath.IsAbs(commandsPath) {
		commandsPath = filepath.Join(".", commandsPath)
	}

	// 读取JSON文件
	data, err := ioutil.ReadFile(commandsPath)
	if err != nil {
		return fmt.Errorf("failed to read commands file: %w", err)
	}

	// 解析命令配置
	var config model.CommandConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse commands config: %w", err)
	}

	if config.Version == "" {
		return fmt.Errorf("missing version in commands config")
	}

	globalConfig.Commands.Commands = config.Commands
	globalConfig.Commands.Version = config.Version
	fmt.Printf("Loaded commands config v%s with %d commands\n", config.Version, len(config.Commands))
	return nil
}

func Get() *Config {
	return globalConfig
}

func GetCommands() []model.Command {
	if globalConfig == nil {
		return []model.Command{}
	}
	return globalConfig.Commands.Commands
}

// GetCommandsV2 为了向后兼容保留这个方法
func GetCommandsV2() []model.Command {
	return GetCommands()
}

// GetCommandsVersion 获取命令配置版本
func GetCommandsVersion() string {
	if globalConfig == nil {
		return "3.0"
	}
	return globalConfig.Commands.Version
}

func ReloadCommands() error {
	return loadCommands()
}

// AddCommand 添加新命令
func AddCommand(cmd *Command) error {
	if globalConfig == nil {
		return fmt.Errorf("config not loaded")
	}
	
	// 检查命令是否已存在
	for _, existingCmd := range globalConfig.Commands.Commands {
		if existingCmd.ID == cmd.ID {
			return fmt.Errorf("command with ID %s already exists", cmd.ID)
		}
	}
	
	// 转换为model.Command
	newCmd := model.Command{
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
		TemplateParams: cmd.TemplateParams,
		CreatedAt:      cmd.CreatedAt,
		UpdatedAt:      cmd.UpdatedAt,
	}
	
	// 转换安全配置
	if cmd.Security != nil {
		newCmd.Security = &model.SecurityConfig{
			RequirePin: cmd.Security.RequirePin,
			Whitelist:  cmd.Security.Whitelist,
			AdminOnly:  cmd.Security.AdminOnly,
		}
	}
	
	// 转换首页布局配置
	if cmd.HomeLayout != nil {
		newCmd.HomeLayout = &model.HomeLayoutConfig{
			ShowOnHome: cmd.HomeLayout.ShowOnHome,
			Color:      cmd.HomeLayout.Color,
			Priority:   cmd.HomeLayout.Priority,
		}
		if cmd.HomeLayout.DefaultPosition != nil {
			newCmd.HomeLayout.DefaultPosition = &model.PositionConfig{
				X:      cmd.HomeLayout.DefaultPosition.X,
				Y:      cmd.HomeLayout.DefaultPosition.Y,
				Width:  cmd.HomeLayout.DefaultPosition.Width,
				Height: cmd.HomeLayout.DefaultPosition.Height,
			}
		}
	}
	
	// 添加到命令列表
	globalConfig.Commands.Commands = append(globalConfig.Commands.Commands, newCmd)
	
	// 保存到文件
	return saveCommands()
}

// UpdateCommand 更新命令
func UpdateCommand(cmd *Command) error {
	if globalConfig == nil {
		return fmt.Errorf("config not loaded")
	}
	
	// 查找并更新命令
	for i, existingCmd := range globalConfig.Commands.Commands {
		if existingCmd.ID == cmd.ID {
			// 更新命令信息
			updatedCmd := model.Command{
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
				TemplateParams: cmd.TemplateParams,
				CreatedAt:      existingCmd.CreatedAt, // 保持原创建时间
				UpdatedAt:      cmd.UpdatedAt,
			}
			
			// 转换安全配置
			if cmd.Security != nil {
				updatedCmd.Security = &model.SecurityConfig{
					RequirePin: cmd.Security.RequirePin,
					Whitelist:  cmd.Security.Whitelist,
					AdminOnly:  cmd.Security.AdminOnly,
				}
			}
			
			// 转换首页布局配置
			if cmd.HomeLayout != nil {
				updatedCmd.HomeLayout = &model.HomeLayoutConfig{
					ShowOnHome: cmd.HomeLayout.ShowOnHome,
					Color:      cmd.HomeLayout.Color,
					Priority:   cmd.HomeLayout.Priority,
				}
				if cmd.HomeLayout.DefaultPosition != nil {
					updatedCmd.HomeLayout.DefaultPosition = &model.PositionConfig{
						X:      cmd.HomeLayout.DefaultPosition.X,
						Y:      cmd.HomeLayout.DefaultPosition.Y,
						Width:  cmd.HomeLayout.DefaultPosition.Width,
						Height: cmd.HomeLayout.DefaultPosition.Height,
					}
				}
			}
			
			globalConfig.Commands.Commands[i] = updatedCmd
			return saveCommands()
		}
	}
	
	return fmt.Errorf("command not found: %s", cmd.ID)
}

// DeleteCommand 删除命令
func DeleteCommand(id string) error {
	if globalConfig == nil {
		return fmt.Errorf("config not loaded")
	}
	
	// 查找并删除命令
	for i, cmd := range globalConfig.Commands.Commands {
		if cmd.ID == id {
			// 从切片中删除
			globalConfig.Commands.Commands = append(
				globalConfig.Commands.Commands[:i],
				globalConfig.Commands.Commands[i+1:]...,
			)
			return saveCommands()
		}
	}
	
	return fmt.Errorf("command not found: %s", id)
}

// saveCommands 保存命令配置到文件
func saveCommands() error {
	commandConfig := model.CommandConfig{
		Version:  globalConfig.Commands.Version,
		Commands: globalConfig.Commands.Commands,
	}
	
	data, err := json.MarshalIndent(commandConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}
	
	commandsPath := globalConfig.Commands.ConfigPath
	if !filepath.IsAbs(commandsPath) {
		commandsPath = filepath.Join(".", commandsPath)
	}
	
	if err := ioutil.WriteFile(commandsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write commands file: %w", err)
	}
	
	return nil
}

// Command 结构体用于API接口
type Command struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name,omitempty"`
	Description    string                 `json:"description,omitempty"`
	Category       string                 `json:"category,omitempty"`
	Icon           string                 `json:"icon,omitempty"`
	Command        string                 `json:"command"`
	Platform       string                 `json:"platform"`
	CommandType    string                 `json:"commandType,omitempty"`
	Security       *model.SecurityConfig  `json:"security,omitempty"`
	Timeout        int                    `json:"timeout,omitempty"`
	UserID         string                 `json:"userId,omitempty"`
	DeviceID       string                 `json:"deviceId,omitempty"`
	HomeLayout     *model.HomeLayoutConfig `json:"homeLayout,omitempty"`
	TemplateId     string                 `json:"templateId,omitempty"`
	TemplateParams map[string]interface{} `json:"templateParams,omitempty"`
	CreatedAt      string                 `json:"createdAt,omitempty"`
	UpdatedAt      string                 `json:"updatedAt,omitempty"`
}