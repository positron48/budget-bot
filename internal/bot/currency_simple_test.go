package bot

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockFXClient is a simple mock for testing
type MockFXClient struct{}

func (m *MockFXClient) GetRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time, accessToken string) (float64, error) {
	// Simple mock implementation
	if fromCurrency == "USD" && toCurrency == "RUB" {
		return 75.5, nil
	}
	return 1.0, nil
}

func TestCurrencyConverter_ConvertToBaseCurrency_SameCurrency(t *testing.T) {
	ctx := context.Background()
	fxClient := &MockFXClient{}
	logger := zap.NewNop()
	converter := NewCurrencyConverter(fxClient, logger)

	amount := int64(1000)
	currency := "USD"
	date := time.Now()

	result, err := converter.ConvertToBaseCurrency(ctx, amount, currency, currency, date, "RUB")

	assert.NoError(t, err)
	assert.Equal(t, amount, result)
}

func TestCurrencyConverter_ConvertToBaseCurrency_ZeroAmount(t *testing.T) {
	ctx := context.Background()
	fxClient := &MockFXClient{}
	logger := zap.NewNop()
	converter := NewCurrencyConverter(fxClient, logger)

	amount := int64(0)
	fromCurrency := "USD"
	toCurrency := "RUB"
	date := time.Now()

	result, err := converter.ConvertToBaseCurrency(ctx, amount, fromCurrency, toCurrency, date, "RUB")

	assert.NoError(t, err)
	assert.Equal(t, int64(0), result)
}

func TestCurrencyConverter_cacheKey(t *testing.T) {
	converter := &CurrencyConverter{}

	fromCurrency := "USD"
	toCurrency := "RUB"
	date := time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC)

	key := converter.cacheKey(fromCurrency, toCurrency, date)

	expectedKey := "USD|RUB|2023-12-25"
	assert.Equal(t, expectedKey, key)
}

func TestCurrencyConverter_cacheSet(t *testing.T) {
	converter := &CurrencyConverter{
		cache: &fxCache{data: make(map[string]cachedRate)},
	}

	key := "USD|RUB|2023-12-25"
	rate := 75.5

	converter.cacheSet(key, rate)

	cachedRate, exists := converter.cacheGet(key)
	assert.True(t, exists)
	assert.Equal(t, rate, cachedRate)
}

func TestCurrencyConverter_cacheGet(t *testing.T) {
	converter := &CurrencyConverter{
		cache: &fxCache{data: make(map[string]cachedRate)},
	}

	key := "USD|RUB|2023-12-25"
	expectedRate := 75.5

	// Set cache
	converter.cacheSet(key, expectedRate)

	// Get from cache
	rate, exists := converter.cacheGet(key)

	assert.True(t, exists)
	assert.Equal(t, expectedRate, rate)
}

func TestCurrencyConverter_cacheGet_NotExists(t *testing.T) {
	converter := &CurrencyConverter{
		cache: &fxCache{data: make(map[string]cachedRate)},
	}

	key := "USD|RUB|2023-12-25"

	// Get from empty cache
	rate, exists := converter.cacheGet(key)

	assert.False(t, exists)
	assert.Equal(t, 0.0, rate)
}

func TestCurrencyParser_ValidateCurrency_ValidCurrencies(t *testing.T) {
	parser := NewCurrencyParser()

	validCurrencies := []string{"USD", "EUR", "RUB", "GBP", "JPY"}

	for _, currency := range validCurrencies {
		result := parser.ValidateCurrency(currency)
		assert.True(t, result, "Currency %s should be valid", currency)
	}
}

func TestCurrencyParser_ValidateCurrency_InvalidCurrencies(t *testing.T) {
	parser := NewCurrencyParser()

	invalidCurrencies := []string{"", "US", "USDD", "123", "ABC", "USD ", " USD", "CNY", "CAD", "AUD", "CHF", "SEK"}

	for _, currency := range invalidCurrencies {
		result := parser.ValidateCurrency(currency)
		assert.False(t, result, "Currency '%s' should be invalid", currency)
	}
}

func TestCurrencyParser_ParseCurrency_ValidInput(t *testing.T) {
	parser := NewCurrencyParser()

	testCases := []struct {
		input    string
		expected string
	}{
		{"100 USD", "USD"},
		{"50.25 EUR", "EUR"},
		{"1000 RUB", "RUB"},
		{"$100", "USD"},
		{"€50", "EUR"},
		{"₽1000", "RUB"},
		{"100", ""}, // No currency specified
		{"", ""},    // Empty input
	}

	for _, tc := range testCases {
		code, _, _ := parser.ParseCurrency(tc.input)
		assert.Equal(t, tc.expected, code, "Input: '%s'", tc.input)
	}
}

func TestCurrencyParser_ParseCurrency_ComplexInput(t *testing.T) {
	parser := NewCurrencyParser()

	testCases := []struct {
		input    string
		expected string
	}{
		{"Spent 100 USD on groceries", "USD"},
		{"Received 50.75 EUR from client", "EUR"},
		{"Payment: 1000 RUB", "RUB"},
		{"$100.50 for dinner", "USD"},
		{"€25.99 subscription", "EUR"},
		{"₽5000 salary", "RUB"},
		{"100 USD and 50 EUR", "USD"}, // First match
		{"No currency here", ""},
		{"123.45", ""},
	}

	for _, tc := range testCases {
		code, _, _ := parser.ParseCurrency(tc.input)
		assert.Equal(t, tc.expected, code, "Input: '%s'", tc.input)
	}
}
