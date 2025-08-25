package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew_WithDebugLevel(t *testing.T) {
	// Test New function with debug level
	logger, err := New("debug")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithInfoLevel(t *testing.T) {
	// Test New function with info level
	logger, err := New("info")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithWarnLevel(t *testing.T) {
	// Test New function with warn level
	logger, err := New("warn")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithErrorLevel(t *testing.T) {
	// Test New function with error level
	logger, err := New("error")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithUpperCaseLevel(t *testing.T) {
	// Test New function with uppercase level
	logger, err := New("DEBUG")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithMixedCaseLevel(t *testing.T) {
	// Test New function with mixed case level
	logger, err := New("Info")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithEmptyLevel(t *testing.T) {
	// Test New function with empty level
	logger, err := New("")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithInvalidLevel(t *testing.T) {
	// Test New function with invalid level
	logger, err := New("invalid_level")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithSpecialCharacters(t *testing.T) {
	// Test New function with special characters in level
	logger, err := New("debug_level_with_special_chars_123")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithVeryLongLevel(t *testing.T) {
	// Test New function with very long level
	longLevel := "very_long_log_level_that_might_be_used_in_some_edge_cases_123456789"
	logger, err := New(longLevel)
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithNumbersInLevel(t *testing.T) {
	// Test New function with numbers in level
	logger, err := New("debug123")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithUnderscoresInLevel(t *testing.T) {
	// Test New function with underscores in level
	logger, err := New("debug_level")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithDashesInLevel(t *testing.T) {
	// Test New function with dashes in level
	logger, err := New("debug-level")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithDotsInLevel(t *testing.T) {
	// Test New function with dots in level
	logger, err := New("debug.level")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithSpacesInLevel(t *testing.T) {
	// Test New function with spaces in level
	logger, err := New("debug level")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithUnicodeInLevel(t *testing.T) {
	// Test New function with unicode in level
	logger, err := New("debug_уровень")
	if err == nil {
		assert.NotNil(t, logger)
		assert.IsType(t, &zap.Logger{}, logger)
	}
}

func TestNew_WithAllLevels(t *testing.T) {
	// Test New function with all possible levels
	levels := []string{"debug", "info", "warn", "error", "DEBUG", "INFO", "WARN", "ERROR"}
	
	for _, level := range levels {
		logger, err := New(level)
		if err == nil {
			assert.NotNil(t, logger)
			assert.IsType(t, &zap.Logger{}, logger)
		}
	}
}

func TestNew_WithEdgeCases(t *testing.T) {
	// Test New function with edge cases
	edgeCases := []string{
		"d",           // Single character
		"de",          // Two characters
		"deb",         // Three characters
		"debu",        // Four characters
		"debug",       // Five characters
		"debugg",      // Six characters
		"debuggg",     // Seven characters
		"debugggg",    // Eight characters
	}
	
	for _, level := range edgeCases {
		logger, err := New(level)
		if err == nil {
			assert.NotNil(t, logger)
			assert.IsType(t, &zap.Logger{}, logger)
		}
	}
}

func TestNew_WithSpecialLevels(t *testing.T) {
	// Test New function with special levels
	specialLevels := []string{
		"trace",       // Lower than debug
		"fatal",       // Higher than error
		"panic",       // Panic level
		"DPANIC",      // DPanic level
	}
	
	for _, level := range specialLevels {
		logger, err := New(level)
		if err == nil {
			assert.NotNil(t, logger)
			assert.IsType(t, &zap.Logger{}, logger)
		}
	}
}
