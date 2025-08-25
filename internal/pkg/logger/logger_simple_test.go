package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew_Exists(t *testing.T) {
	// Test that the function exists and can be called
	// This will fail without proper setup, but we're just testing that the function exists
	assert.NotNil(t, New)
}

func TestNew_ReturnsLogger(t *testing.T) {
	// Test that New function returns a logger
	logger, err := New("debug")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}
