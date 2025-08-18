package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config is the root application configuration loaded from YAML/env.
type Config struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

type TelegramConfig struct {
	Token        string `mapstructure:"token"`
	APIBaseURL   string `mapstructure:"api_base_url"`
	Debug        bool   `mapstructure:"debug"`
	UpdatesTimeout int  `mapstructure:"updates_timeout"`
}

type GRPCConfig struct {
	Address  string `mapstructure:"address"`
	Insecure bool   `mapstructure:"insecure"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Address string `mapstructure:"address"`
}

// Load loads configuration from configs/config.yaml and environment variables.
func Load() (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("telegram.debug", true)
	v.SetDefault("telegram.updates_timeout", 30)
	v.SetDefault("grpc.address", "127.0.0.1:8080")
	v.SetDefault("grpc.insecure", true)
	v.SetDefault("database.driver", "sqlite3")
	v.SetDefault("database.dsn", "file:./data/bot.sqlite3?_foreign_keys=on")
	v.SetDefault("logging.level", "debug")
	v.SetDefault("metrics.enabled", false)
	v.SetDefault("metrics.address", ":9090")

	// Files
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	// Env
	v.SetEnvPrefix("BOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Map some convenient env names
	v.BindEnv("telegram.token", "TELEGRAM_BOT_TOKEN")
	v.BindEnv("telegram.api_base_url", "TELEGRAM_API_BASE_URL")
	v.BindEnv("grpc.address", "GRPC_SERVER_ADDRESS")
	v.BindEnv("database.dsn", "DATABASE_DSN")
	v.BindEnv("logging.level", "LOG_LEVEL")

	// Read file if present
	if err := v.ReadInConfig(); err != nil {
		// Non-fatal: allow running with only envs/defaults
		// But return a clearer error if it's an unexpected issue.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed unmarshalling config: %w", err)
	}

	return &cfg, nil
}


