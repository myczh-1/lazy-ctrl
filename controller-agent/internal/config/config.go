package config

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
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
	Enabled bool   `mapstructure:"enabled"`
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
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
	ConfigPath   string                 `mapstructure:"config_path"`
	Commands     map[string]interface{} `mapstructure:"commands"`
	HotReload    bool                   `mapstructure:"hot_reload"`
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
		viper.AddConfigPath("./configs")
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

	viper.SetConfigFile(commandsPath)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read commands config: %w", err)
	}

	commands := make(map[string]interface{})
	if err := viper.Unmarshal(&commands); err != nil {
		return fmt.Errorf("failed to unmarshal commands: %w", err)
	}

	globalConfig.Commands.Commands = commands
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

func ReloadCommands() error {
	return loadCommands()
}