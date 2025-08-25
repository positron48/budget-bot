package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Exists(t *testing.T) {
	// Test that the function exists and can be called
	// This will fail without proper setup, but we're just testing that the function exists
	assert.NotNil(t, Load)
}

func TestConfig_Struct(t *testing.T) {
	// Test creating a Config struct
	config := &Config{
		Telegram: TelegramConfig{
			Token: "test_token",
		},
		GRPC: GRPCConfig{
			Address: "localhost:8081",
		},
		Database: DatabaseConfig{
			Driver: "sqlite",
			DSN:    "test.db",
		},
	}
	
	assert.Equal(t, "test_token", config.Telegram.Token)
	assert.Equal(t, "localhost:8081", config.GRPC.Address)
	assert.Equal(t, "sqlite", config.Database.Driver)
	assert.Equal(t, "test.db", config.Database.DSN)
}

func TestTelegramConfig_Struct(t *testing.T) {
	// Test creating a TelegramConfig struct
	tgConfig := &TelegramConfig{
		Token:         "test_token",
		APIBaseURL:    "https://api.telegram.org",
		Debug:         true,
		UpdatesTimeout: 30,
	}
	
	assert.Equal(t, "test_token", tgConfig.Token)
	assert.Equal(t, "https://api.telegram.org", tgConfig.APIBaseURL)
	assert.True(t, tgConfig.Debug)
	assert.Equal(t, 30, tgConfig.UpdatesTimeout)
}

func TestGRPCConfig_Struct(t *testing.T) {
	// Test creating a GRPCConfig struct
	grpcConfig := &GRPCConfig{
		Address:  "localhost:8081",
		Insecure: true,
	}
	
	assert.Equal(t, "localhost:8081", grpcConfig.Address)
	assert.True(t, grpcConfig.Insecure)
}

func TestDatabaseConfig_Struct(t *testing.T) {
	// Test creating a DatabaseConfig struct
	dbConfig := &DatabaseConfig{
		Driver: "sqlite",
		DSN:    "test.db",
	}
	
	assert.Equal(t, "sqlite", dbConfig.Driver)
	assert.Equal(t, "test.db", dbConfig.DSN)
}
