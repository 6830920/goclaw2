package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
)

type Config struct {
	Zhipu    ZhipuConfig    `mapstructure:"zhipu"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Memory   MemoryConfig   `mapstructure:"memory"`
	Gateway  GatewayConfig  `mapstructure:"gateway"`
}

type ZhipuConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	BaseURL     string  `mapstructure:"base_url"`
	Model       string  `mapstructure:"model"`
	Temperature float64 `mapstructure:"temperature"`
	MaxTokens   int     `mapstructure:"max_tokens"`
}

type AgentConfig struct {
	MaxHistory int `mapstructure:"max_history"`
}

type MemoryConfig struct {
	Type       string `mapstructure:"type"`
	FilePath   string `mapstructure:"file_path"`
	Workspace  string `mapstructure:"workspace"`  // Workspace 目录
}

type GatewayConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
}

var globalConfig *Config

// Load initializes the configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Read from config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Try to find config file in default locations
		v.SetConfigName(".goclaw")
		v.SetConfigType("yaml")
		v.AddConfigPath("$HOME")
		v.AddConfigPath(".")
		v.ReadInConfig()
	}

	// Environment variable overrides
	v.SetEnvPrefix("GOCLAW")
	v.AutomaticEnv()

	// Bind environment variables
	v.BindEnv("zhipu.api_key", "ZHIPU_API_KEY", "GOCLAW_ZHIPU_API_KEY")
	v.BindEnv("zhipu.model", "ZHIPU_MODEL", "GOCLAW_ZHIPU_MODEL")
	v.BindEnv("zhipu.temperature", "ZHIPU_TEMPERATURE", "GOCLAW_ZHIPU_TEMPERATURE")
	v.BindEnv("zhipu.max_tokens", "ZHIPU_MAX_TOKENS", "GOCLAW_ZHIPU_MAX_TOKENS")
	v.BindEnv("gateway.port", "GOCLAW_GATEWAY_PORT")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand ~ in workspace path
	cfg.Memory.Workspace = expandPath(cfg.Memory.Workspace)

	// Validate
	if err := validate(&cfg); err != nil {
		return nil, err
	}

	globalConfig = &cfg
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("zhipu.base_url", "https://open.bigmodel.cn/api/paas/v4")
	v.SetDefault("zhipu.model", "glm-4-flash")
	v.SetDefault("zhipu.temperature", 0.7)
	v.SetDefault("zhipu.max_tokens", 4096)
	v.SetDefault("agent.max_history", 50)
	v.SetDefault("memory.type", "sqlite")
	v.SetDefault("memory.file_path", "./goclaw.db")
	v.SetDefault("memory.workspace", "~/.goclaw/workspace")
	v.SetDefault("gateway.enabled", false)
	v.SetDefault("gateway.port", 8080)
	v.SetDefault("gateway.host", "localhost")
}

func validate(cfg *Config) error {
	if cfg.Zhipu.APIKey == "" {
		return fmt.Errorf("zhipu api_key is required (set ZHIPU_API_KEY environment variable)")
	}
	return nil
}

// Get returns the global configuration
func Get() *Config {
	return globalConfig
}

// GetEnvOrDefault retrieves an environment variable or returns a default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvIntOrDefault retrieves an environment variable as an integer or returns a default value
func GetEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func expandPath(path string) string {
	if len(path) == 0 {
		return path
	}
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}
