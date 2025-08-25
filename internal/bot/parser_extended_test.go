package bot

import (
	"testing"

	"budget-bot/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMessageParser_Validate_ValidInput(t *testing.T) {
	parser := NewMessageParser()

	validTransactions := []*ParsedTransaction{
		{
			Amount:   domain.NewMoney(10000, "USD"),
			Currency: "USD",
		},
		{
			Amount:   domain.NewMoney(5025, "EUR"),
			Currency: "EUR",
		},
		{
			Amount:   domain.NewMoney(100000, "RUB"),
			Currency: "RUB",
		},
	}

	for _, transaction := range validTransactions {
		errors := parser.Validate(transaction)
		assert.Empty(t, errors, "Transaction should be valid: %+v", transaction)
	}
}

func TestMessageParser_Validate_InvalidInput(t *testing.T) {
	parser := NewMessageParser()

	invalidTransactions := []*ParsedTransaction{
		nil, // nil transaction
		{
			Amount:   nil, // no amount
			Currency: "USD",
		},
		{
			Amount:   domain.NewMoney(0, "USD"), // zero amount
			Currency: "USD",
		},
		{
			Amount:   domain.NewMoney(10000, "INVALID"), // invalid currency
			Currency: "INVALID",
		},
	}

	for _, transaction := range invalidTransactions {
		errors := parser.Validate(transaction)
		assert.NotEmpty(t, errors, "Transaction should be invalid: %+v", transaction)
	}
}

func TestMessageParser_Validate_EdgeCases(t *testing.T) {
	parser := NewMessageParser()

	edgeCases := []struct {
		transaction *ParsedTransaction
		shouldBeValid bool
	}{
		{
			transaction: &ParsedTransaction{
				Amount:   domain.NewMoney(1, "USD"), // minimal amount
				Currency: "USD",
			},
			shouldBeValid: true,
		},
		{
			transaction: &ParsedTransaction{
				Amount:   domain.NewMoney(99999999, "USD"), // large amount
				Currency: "USD",
			},
			shouldBeValid: true,
		},
		{
			transaction: &ParsedTransaction{
				Amount:   domain.NewMoney(10000, "USD"),
				Currency: "", // empty currency
			},
			shouldBeValid: true, // empty currency is valid
		},
		{
			transaction: &ParsedTransaction{
				Amount:   domain.NewMoney(10000, "CNY"), // unsupported currency
				Currency: "CNY",
			},
			shouldBeValid: false,
		},
	}

	for _, tc := range edgeCases {
		errors := parser.Validate(tc.transaction)
		if tc.shouldBeValid {
			assert.Empty(t, errors, "Transaction should be valid: %+v", tc.transaction)
		} else {
			assert.NotEmpty(t, errors, "Transaction should be invalid: %+v", tc.transaction)
		}
	}
}

func TestMessageParser_Validate_CurrencyValidation(t *testing.T) {
	parser := NewMessageParser()

	currencyTests := []struct {
		currency      string
		shouldBeValid bool
	}{
		{"USD", true},
		{"EUR", true},
		{"RUB", true},
		{"GBP", true},
		{"JPY", true},
		{"", true}, // empty currency is valid
		{"CNY", false}, // not supported
		{"CAD", false}, // not supported
		{"INVALID", false},
	}

	for _, tc := range currencyTests {
		transaction := &ParsedTransaction{
			Amount:   domain.NewMoney(10000, tc.currency),
			Currency: tc.currency,
		}
		errors := parser.Validate(transaction)
		if tc.shouldBeValid {
			assert.Empty(t, errors, "Currency should be valid: %s", tc.currency)
		} else {
			assert.NotEmpty(t, errors, "Currency should be invalid: %s", tc.currency)
		}
	}
}

func TestMessageParser_Validate_AmountValidation(t *testing.T) {
	parser := NewMessageParser()

	amountTests := []struct {
		amountMinor   int64
		shouldBeValid bool
	}{
		{1, true},           // minimal amount
		{100, true},         // 1.00
		{10000, true},       // 100.00
		{99999999, true},    // large amount
		{0, false},          // zero amount
		{-100, true},        // negative amount is valid
	}

	for _, tc := range amountTests {
		transaction := &ParsedTransaction{
			Amount:   domain.NewMoney(tc.amountMinor, "USD"),
			Currency: "USD",
		}
		errors := parser.Validate(transaction)
		if tc.shouldBeValid {
			assert.Empty(t, errors, "Amount should be valid: %d", tc.amountMinor)
		} else {
			assert.NotEmpty(t, errors, "Amount should be invalid: %d", tc.amountMinor)
		}
	}
}
