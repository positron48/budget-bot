package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingConfig_Struct(t *testing.T) {
	// Test creating a LoggingConfig struct
	loggingConfig := &LoggingConfig{
		Level: "debug",
	}
	
	assert.Equal(t, "debug", loggingConfig.Level)
}

func TestLoggingConfig_WithDifferentLevels(t *testing.T) {
	// Test LoggingConfig with different levels
	levels := []string{"debug", "info", "warn", "error"}
	
	for _, level := range levels {
		loggingConfig := &LoggingConfig{
			Level: level,
		}
		assert.Equal(t, level, loggingConfig.Level)
	}
}

func TestMetricsConfig_Struct(t *testing.T) {
	// Test creating a MetricsConfig struct
	metricsConfig := &MetricsConfig{
		Enabled: true,
		Address: ":9090",
	}
	
	assert.True(t, metricsConfig.Enabled)
	assert.Equal(t, ":9090", metricsConfig.Address)
}

func TestMetricsConfig_WithDifferentSettings(t *testing.T) {
	// Test MetricsConfig with different settings
	metricsConfig1 := &MetricsConfig{
		Enabled: true,
		Address: ":9090",
	}
	
	metricsConfig2 := &MetricsConfig{
		Enabled: false,
		Address: ":8080",
	}
	
	assert.True(t, metricsConfig1.Enabled)
	assert.Equal(t, ":9090", metricsConfig1.Address)
	
	assert.False(t, metricsConfig2.Enabled)
	assert.Equal(t, ":8080", metricsConfig2.Address)
}

func TestServerConfig_Struct(t *testing.T) {
	// Test creating a ServerConfig struct
	serverConfig := &ServerConfig{
		Address: ":8088",
	}
	
	assert.Equal(t, ":8088", serverConfig.Address)
}

func TestServerConfig_WithDifferentAddresses(t *testing.T) {
	// Test ServerConfig with different addresses
	addresses := []string{":8088", ":9090", "localhost:8080", "0.0.0.0:8088"}
	
	for _, address := range addresses {
		serverConfig := &ServerConfig{
			Address: address,
		}
		assert.Equal(t, address, serverConfig.Address)
	}
}

func TestOAuthConfig_Struct(t *testing.T) {
	// Test creating an OAuthConfig struct
	oauthConfig := &OAuthConfig{
		WebBaseURL: "http://localhost:3000",
	}
	
	assert.Equal(t, "http://localhost:3000", oauthConfig.WebBaseURL)
}

func TestOAuthConfig_WithDifferentURLs(t *testing.T) {
	// Test OAuthConfig with different URLs
	urls := []string{
		"http://localhost:3000",
		"https://example.com",
		"https://auth.example.com/oauth",
	}
	
	for _, url := range urls {
		oauthConfig := &OAuthConfig{
			WebBaseURL: url,
		}
		assert.Equal(t, url, oauthConfig.WebBaseURL)
	}
}

func TestTelegramConfig_Extended(t *testing.T) {
	// Test TelegramConfig with different values
	tgConfig1 := &TelegramConfig{
		Token:         "test_token_1",
		APIBaseURL:    "https://api.telegram.org",
		Debug:         true,
		UpdatesTimeout: 30,
		WebhookEnable: false,
		WebhookURL:    "",
		WebhookDomain: "",
		WebhookPath:   "/tg",
	}
	
	tgConfig2 := &TelegramConfig{
		Token:         "test_token_2",
		APIBaseURL:    "https://api.telegram.org",
		Debug:         false,
		UpdatesTimeout: 60,
		WebhookEnable: true,
		WebhookURL:    "https://example.com/tg",
		WebhookDomain: "https://example.com",
		WebhookPath:   "/tg",
	}
	
	assert.Equal(t, "test_token_1", tgConfig1.Token)
	assert.True(t, tgConfig1.Debug)
	assert.Equal(t, 30, tgConfig1.UpdatesTimeout)
	assert.False(t, tgConfig1.WebhookEnable)
	
	assert.Equal(t, "test_token_2", tgConfig2.Token)
	assert.False(t, tgConfig2.Debug)
	assert.Equal(t, 60, tgConfig2.UpdatesTimeout)
	assert.True(t, tgConfig2.WebhookEnable)
	assert.Equal(t, "https://example.com/tg", tgConfig2.WebhookURL)
}

func TestGRPCConfig_Extended(t *testing.T) {
	// Test GRPCConfig with different values
	grpcConfig1 := &GRPCConfig{
		Address:  "localhost:8081",
		Insecure: true,
	}
	
	grpcConfig2 := &GRPCConfig{
		Address:  "example.com:443",
		Insecure: false,
	}
	
	assert.Equal(t, "localhost:8081", grpcConfig1.Address)
	assert.True(t, grpcConfig1.Insecure)
	
	assert.Equal(t, "example.com:443", grpcConfig2.Address)
	assert.False(t, grpcConfig2.Insecure)
}

func TestDatabaseConfig_Extended(t *testing.T) {
	// Test DatabaseConfig with different values
	dbConfig1 := &DatabaseConfig{
		Driver: "sqlite",
		DSN:    "file:./data/bot.sqlite?_foreign_keys=on",
	}
	
	dbConfig2 := &DatabaseConfig{
		Driver: "postgres",
		DSN:    "postgres://user:pass@localhost/dbname",
	}
	
	assert.Equal(t, "sqlite", dbConfig1.Driver)
	assert.Equal(t, "file:./data/bot.sqlite?_foreign_keys=on", dbConfig1.DSN)
	
	assert.Equal(t, "postgres", dbConfig2.Driver)
	assert.Equal(t, "postgres://user:pass@localhost/dbname", dbConfig2.DSN)
}

func TestConfig_WithAllFields(t *testing.T) {
	// Test Config with all fields populated
	config := &Config{
		Telegram: TelegramConfig{
			Token:         "test_token",
			APIBaseURL:    "https://api.telegram.org",
			Debug:         true,
			UpdatesTimeout: 30,
			WebhookEnable: false,
			WebhookURL:    "",
			WebhookDomain: "",
			WebhookPath:   "/tg",
		},
		GRPC: GRPCConfig{
			Address:  "localhost:8081",
			Insecure: true,
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "file:./data/bot.sqlite?_foreign_keys=on",
		},
		Server: ServerConfig{
			Address: ":8088",
		},
		Logging: LoggingConfig{
			Level: "debug",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Address: ":9090",
		},
		OAuth: OAuthConfig{
			WebBaseURL: "http://localhost:3000",
		},
	}
	
	assert.Equal(t, "test_token", config.Telegram.Token)
	assert.Equal(t, "localhost:8081", config.GRPC.Address)
	assert.Equal(t, "sqlite", config.Database.Driver)
	assert.Equal(t, ":8088", config.Server.Address)
	assert.Equal(t, "debug", config.Logging.Level)
	assert.True(t, config.Metrics.Enabled)
	assert.Equal(t, "http://localhost:3000", config.OAuth.WebBaseURL)
}

func TestConfig_WithEmptyFields(t *testing.T) {
	// Test Config with empty fields
	config := &Config{
		Telegram: TelegramConfig{},
		GRPC:     GRPCConfig{},
		Database: DatabaseConfig{},
		Server:   ServerConfig{},
		Logging:  LoggingConfig{},
		Metrics:  MetricsConfig{},
		OAuth:    OAuthConfig{},
	}
	
	assert.Equal(t, "", config.Telegram.Token)
	assert.Equal(t, "", config.GRPC.Address)
	assert.Equal(t, "", config.Database.Driver)
	assert.Equal(t, "", config.Server.Address)
	assert.Equal(t, "", config.Logging.Level)
	assert.False(t, config.Metrics.Enabled)
	assert.Equal(t, "", config.OAuth.WebBaseURL)
}

func TestTelegramConfig_WithSpecialCharacters(t *testing.T) {
	// Test TelegramConfig with special characters
	tgConfig := &TelegramConfig{
		Token:         "test_token_with_special_chars_123!@#",
		APIBaseURL:    "https://api.telegram.org/bot",
		Debug:         true,
		UpdatesTimeout: 30,
		WebhookEnable: true,
		WebhookURL:    "https://example.com/webhook/tg",
		WebhookDomain: "https://example.com",
		WebhookPath:   "/webhook/tg",
	}
	
	assert.Equal(t, "test_token_with_special_chars_123!@#", tgConfig.Token)
	assert.Equal(t, "https://api.telegram.org/bot", tgConfig.APIBaseURL)
	assert.True(t, tgConfig.Debug)
	assert.Equal(t, 30, tgConfig.UpdatesTimeout)
	assert.True(t, tgConfig.WebhookEnable)
	assert.Equal(t, "https://example.com/webhook/tg", tgConfig.WebhookURL)
	assert.Equal(t, "https://example.com", tgConfig.WebhookDomain)
	assert.Equal(t, "/webhook/tg", tgConfig.WebhookPath)
}

func TestDatabaseConfig_WithComplexDSN(t *testing.T) {
	// Test DatabaseConfig with complex DSN
	complexDSN := "postgres://user:password@localhost:5432/dbname?sslmode=require&connect_timeout=10"
	dbConfig := &DatabaseConfig{
		Driver: "postgres",
		DSN:    complexDSN,
	}
	
	assert.Equal(t, "postgres", dbConfig.Driver)
	assert.Equal(t, complexDSN, dbConfig.DSN)
}

func TestMetricsConfig_WithDifferentPorts(t *testing.T) {
	// Test MetricsConfig with different ports
	ports := []string{":9090", ":8080", ":3000", ":5000"}
	
	for _, port := range ports {
		metricsConfig := &MetricsConfig{
			Enabled: true,
			Address: port,
		}
		assert.True(t, metricsConfig.Enabled)
		assert.Equal(t, port, metricsConfig.Address)
	}
}
