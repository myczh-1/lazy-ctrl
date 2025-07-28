package config

import (
	"fmt"
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
	ConfigPath string `mapstructure:"config_path"`
	HotReload  bool   `mapstructure:"hot_reload"`
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
		viper.AddConfigPath(".")
	}

	// Set defaults
	setDefaults()

	// Read environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, use defaults
			fmt.Println("Config file not found, using defaults")
		} else {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	globalConfig = &Config{}
	if err := viper.Unmarshal(globalConfig); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
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
	viper.SetDefault("commands.config_path", "configs/commands.json")
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

func Get() *Config {
	return globalConfig
}