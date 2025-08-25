package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew_WithDifferentLevels(t *testing.T) {
	tests := []struct {
		name        string
		level       string
		expectError bool
	}{
		{
			name:        "debug level",
			level:       "debug",
			expectError: false,
		},
		{
			name:        "info level",
			level:       "info",
			expectError: false,
		},
		{
			name:        "warn level",
			level:       "warn",
			expectError: false,
		},
		{
			name:        "error level",
			level:       "error",
			expectError: false,
		},
		{
			name:        "DEBUG uppercase",
			level:       "DEBUG",
			expectError: false,
		},
		{
			name:        "INFO uppercase",
			level:       "INFO",
			expectError: false,
		},
		{
			name:        "empty level",
			level:       "",
			expectError: false,
		},
		{
			name:        "invalid level",
			level:       "invalid",
			expectError: false, // Function doesn't return error for invalid levels
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				
				// Test that logger is functional
				logger.Info("test message")
				logger.Debug("test debug message")
				logger.Warn("test warning message")
				logger.Error("test error message")
			}
		})
	}
}

func TestNew_LoggerFunctionality(t *testing.T) {
	logger, err := New("debug")
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	// Test basic logging functionality
	logger.Info("info message")
	logger.Debug("debug message")
	logger.Warn("warning message")
	logger.Error("error message")

	// Test with fields
	logger.Info("message with fields", zap.String("key", "value"))
	logger.Error("error with fields", zap.Int("code", 500))
}

func TestNew_DevelopmentVsProduction(t *testing.T) {
	// Test development config
	devLogger, err := New("debug")
	assert.NoError(t, err)
	assert.NotNil(t, devLogger)

	// Test production config
	prodLogger, err := New("info")
	assert.NoError(t, err)
	assert.NotNil(t, prodLogger)

	// Both loggers should be functional
	devLogger.Info("development logger test")
	prodLogger.Info("production logger test")
}
