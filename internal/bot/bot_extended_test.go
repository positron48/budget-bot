package bot

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewAuthManager_Extended(t *testing.T) {
	// Test NewAuthManager with different parameters
	mockAuthClient := &MockAuthClient{}
	mockSessionRepo := &MockSessionRepository{}
	logger := zap.NewNop()
	
	am1 := NewAuthManager(mockAuthClient, mockSessionRepo, logger)
	assert.NotNil(t, am1)
	assert.Equal(t, mockAuthClient, am1.authClient)
	assert.Equal(t, mockSessionRepo, am1.sessionRepo)
	assert.Equal(t, logger, am1.logger)
	
	// Test with nil logger
	am2 := NewAuthManager(mockAuthClient, mockSessionRepo, nil)
	assert.NotNil(t, am2)
	assert.Equal(t, mockAuthClient, am2.authClient)
	assert.Equal(t, mockSessionRepo, am2.sessionRepo)
	assert.Nil(t, am2.logger)
}

func TestNewCategoryNameMapper_Extended(t *testing.T) {
	// Test NewCategoryNameMapper with different parameters
	mockCategoryClient := &MockCategoryClient{}
	
	cnm := NewCategoryNameMapper(mockCategoryClient)
	assert.NotNil(t, cnm)
}

func TestNewCurrencyConverter_Extended(t *testing.T) {
	// Test NewCurrencyConverter with different parameters
	mockFxClient := &MockFXClient{}
	logger := zap.NewNop()
	
	cc1 := NewCurrencyConverter(mockFxClient, logger)
	assert.NotNil(t, cc1)
	assert.Equal(t, mockFxClient, cc1.fxClient)
	assert.Equal(t, logger, cc1.logger)
	assert.NotNil(t, cc1.cache)
	
	// Test with nil fxClient
	cc2 := NewCurrencyConverter(nil, logger)
	assert.NotNil(t, cc2)
	assert.Nil(t, cc2.fxClient)
	assert.Equal(t, logger, cc2.logger)
	assert.NotNil(t, cc2.cache)
	
	// Test with nil logger
	cc3 := NewCurrencyConverter(mockFxClient, nil)
	assert.NotNil(t, cc3)
	assert.Equal(t, mockFxClient, cc3.fxClient)
	assert.NotNil(t, cc3.logger) // Should create a nop logger
	assert.NotNil(t, cc3.cache)
}

func TestNewCurrencyParser_Extended(t *testing.T) {
	// Test NewCurrencyParser
	cp := NewCurrencyParser()
	assert.NotNil(t, cp)
}

func TestNewMessageParser_Extended(t *testing.T) {
	// Test NewMessageParser
	mp := NewMessageParser()
	assert.NotNil(t, mp)
}
