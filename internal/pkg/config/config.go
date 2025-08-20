// Package config loads and holds application configuration from files and env.
package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config is the root application configuration loaded from YAML/env.
type Config struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Server   ServerConfig   `mapstructure:"server"`
}

// TelegramConfig holds Telegram Bot API settings.
type TelegramConfig struct {
	// Token is a Telegram Bot API token
	Token string `mapstructure:"token"`
	// APIBaseURL allows pointing to a local Telegram emulator
	APIBaseURL string `mapstructure:"api_base_url"`
	// Debug enables verbose logging for bot API
	Debug bool `mapstructure:"debug"`
	// UpdatesTimeout long polling timeout in seconds
	UpdatesTimeout int `mapstructure:"updates_timeout"`
	// WebhookEnable toggles webhook mode
	WebhookEnable bool `mapstructure:"webhook_enable"`
	// WebhookURL public URL for webhook
	WebhookURL string `mapstructure:"webhook_url"`
	// WebhookPath path to serve webhook on
	WebhookPath string `mapstructure:"webhook_path"`
}

// GRPCConfig holds backend gRPC settings.
type GRPCConfig struct {
	// Address of backend gRPC server
	Address string `mapstructure:"address"`
	// Insecure skips TLS for local development
	Insecure bool `mapstructure:"insecure"`
}

// DatabaseConfig contains DB connection settings.
type DatabaseConfig struct {
	// Driver is a database driver name, e.g., sqlite
	Driver string `mapstructure:"driver"`
	// DSN is a connection string
	DSN string `mapstructure:"dsn"`
}

// LoggingConfig holds logger settings.
type LoggingConfig struct {
	// Level is a zap log level (debug, info, warn, error)
	Level string `mapstructure:"level"`
}

// MetricsConfig holds metrics service settings.
type MetricsConfig struct {
	// Enabled toggles Prometheus metrics endpoint
	Enabled bool `mapstructure:"enabled"`
	// Address is an HTTP listen address for metrics
	Address string `mapstructure:"address"`
}

// ServerConfig holds http server settings.
type ServerConfig struct {
	// Address is an HTTP listen address for health/metrics
	Address string `mapstructure:"address"`
}

// Load loads configuration from configs/config.yaml and environment variables.
func Load() (*Config, error) {
	v := viper.New()

	// Load environment variables from .env if present
	_ = godotenv.Load()

	// Defaults
	v.SetDefault("telegram.debug", true)
	v.SetDefault("telegram.updates_timeout", 30)
	v.SetDefault("telegram.webhook_enable", false)
	v.SetDefault("telegram.webhook_path", "/tg")
	v.SetDefault("grpc.address", "127.0.0.1:8080")
	v.SetDefault("grpc.insecure", true)
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "file:./data/bot.sqlite?_foreign_keys=on")
	v.SetDefault("logging.level", "debug")
	v.SetDefault("metrics.enabled", false)
	v.SetDefault("metrics.address", ":9090")
	v.SetDefault("server.address", ":8088")

	// Files
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")

	// Env
	v.SetEnvPrefix("BOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Map some convenient env names
	_ = v.BindEnv("telegram.token", "TELEGRAM_BOT_TOKEN")
	_ = v.BindEnv("telegram.api_base_url", "TELEGRAM_API_BASE_URL")
	_ = v.BindEnv("telegram.debug", "TELEGRAM_DEBUG")
	_ = v.BindEnv("telegram.updates_timeout", "TELEGRAM_UPDATES_TIMEOUT")
	_ = v.BindEnv("telegram.webhook_enable", "TELEGRAM_WEBHOOK_ENABLE")
	_ = v.BindEnv("telegram.webhook_url", "TELEGRAM_WEBHOOK_URL")
	_ = v.BindEnv("telegram.webhook_path", "TELEGRAM_WEBHOOK_PATH")

	_ = v.BindEnv("grpc.address", "GRPC_SERVER_ADDRESS")
	_ = v.BindEnv("grpc.insecure", "GRPC_INSECURE")

	_ = v.BindEnv("database.driver", "DATABASE_DRIVER")
	_ = v.BindEnv("database.dsn", "DATABASE_DSN")

	_ = v.BindEnv("logging.level", "LOG_LEVEL")

	_ = v.BindEnv("metrics.enabled", "METRICS_ENABLED")
	_ = v.BindEnv("metrics.address", "METRICS_ADDRESS")
	_ = v.BindEnv("server.address", "SERVER_ADDRESS")

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


