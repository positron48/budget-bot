package bot

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewCurrencyConverter_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, NewCurrencyConverter)
}

func TestNewCurrencyConverter_ReturnsConverter(t *testing.T) {
	// Test that NewCurrencyConverter returns a converter
	logger := zap.NewNop()
	converter := NewCurrencyConverter(nil, logger)
	assert.NotNil(t, converter)
	assert.IsType(t, &CurrencyConverter{}, converter)
}

func TestCurrencyConverter_Struct(t *testing.T) {
	// Test creating a CurrencyConverter struct
	converter := &CurrencyConverter{
		fxClient: nil,
		logger:   zap.NewNop(),
		cache:    &fxCache{data: make(map[string]cachedRate)},
	}
	
	assert.NotNil(t, converter)
	assert.Nil(t, converter.fxClient)
	assert.NotNil(t, converter.logger)
	assert.NotNil(t, converter.cache)
}

func TestFxCache_Struct(t *testing.T) {
	// Test creating an fxCache struct
	cache := &fxCache{
		data: make(map[string]cachedRate),
	}
	
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.data)
}

func TestCachedRate_Struct(t *testing.T) {
	// Test creating a cachedRate struct
	now := time.Now()
	rate := cachedRate{
		rate:     1.5,
		storedAt: now,
	}
	
	assert.Equal(t, 1.5, rate.rate)
	assert.Equal(t, now, rate.storedAt)
}
