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
	ConfigPath     string                       `mapstructure:"config_path"`
	Commands       map[string]interface{}       `mapstructure:"commands"`
	CommandsV2     []model.Command              `mapstructure:"-"`
	LegacyCommands model.LegacyCommandConfig    `mapstructure:"-"`
	HotReload      bool                         `mapstructure:"hot_reload"`
	Version        string                       `mapstructure:"-"`
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

	// 尝试解析为新格式
	var newConfig model.CommandConfig
	if err := json.Unmarshal(data, &newConfig); err == nil && newConfig.Version != "" {
		// 新格式
		globalConfig.Commands.CommandsV2 = newConfig.Commands
		globalConfig.Commands.Version = newConfig.Version
		fmt.Printf("Loaded commands config v%s with %d commands\n", newConfig.Version, len(newConfig.Commands))
		return nil
	}

	// 尝试解析为旧格式 (向后兼容)
	var legacyCommands model.LegacyCommandConfig
	if err := json.Unmarshal(data, &legacyCommands); err != nil {
		return fmt.Errorf("failed to parse commands config: %w", err)
	}

	globalConfig.Commands.LegacyCommands = legacyCommands
	globalConfig.Commands.Commands = legacyCommands
	globalConfig.Commands.Version = "1.0"
	fmt.Printf("Loaded legacy commands config with %d commands\n", len(legacyCommands))
	return nil
}

func Get() *Config {
	return globalConfig
}

func GetCommands() map[string]interface{} {
	if globalConfig == nil {
		return make(map[string]interface{})
	}
	return globalConfig.Commands.Commands
}

// GetCommandsV2 获取新格式的命令配置
func GetCommandsV2() []model.Command {
	if globalConfig == nil {
		return []model.Command{}
	}
	return globalConfig.Commands.CommandsV2
}

// GetLegacyCommands 获取旧格式的命令配置
func GetLegacyCommands() model.LegacyCommandConfig {
	if globalConfig == nil {
		return make(model.LegacyCommandConfig)
	}
	return globalConfig.Commands.LegacyCommands
}

// GetCommandsVersion 获取命令配置版本
func GetCommandsVersion() string {
	if globalConfig == nil {
		return "1.0"
	}
	return globalConfig.Commands.Version
}

// IsV2Commands 检查是否使用新版命令格式
func IsV2Commands() bool {
	return GetCommandsVersion() != "1.0"
}

func ReloadCommands() error {
	return loadCommands()
}