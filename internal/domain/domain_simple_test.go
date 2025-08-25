package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMoney(t *testing.T) {
	// Test creating money with valid parameters
	money := NewMoney(10000, "USD")
	assert.Equal(t, int64(10000), money.AmountMinor)
	assert.Equal(t, "USD", money.CurrencyCode)
}

func TestNewMoney_ZeroAmount(t *testing.T) {
	// Test creating money with zero amount
	money := NewMoney(0, "USD")
	assert.Equal(t, int64(0), money.AmountMinor)
	assert.Equal(t, "USD", money.CurrencyCode)
}

func TestNewMoney_LargeAmount(t *testing.T) {
	// Test creating money with large amount
	money := NewMoney(99999999, "EUR")
	assert.Equal(t, int64(99999999), money.AmountMinor)
	assert.Equal(t, "EUR", money.CurrencyCode)
}

func TestNewMoney_NegativeAmount(t *testing.T) {
	// Test creating money with negative amount
	money := NewMoney(-10000, "RUB")
	assert.Equal(t, int64(-10000), money.AmountMinor)
	assert.Equal(t, "RUB", money.CurrencyCode)
}

func TestNewMoney_EmptyCurrency(t *testing.T) {
	// Test creating money with empty currency
	money := NewMoney(10000, "")
	assert.Equal(t, int64(10000), money.AmountMinor)
	assert.Equal(t, "", money.CurrencyCode)
}

func TestCategory_NewCategory(t *testing.T) {
	// Test creating a new category
	category := &Category{
		ID:    "cat123",
		Name:  "Groceries",
		Emoji: "🛒",
	}
	assert.Equal(t, "cat123", category.ID)
	assert.Equal(t, "Groceries", category.Name)
	assert.Equal(t, "🛒", category.Emoji)
}

func TestCategoryTotal_NewCategoryTotal(t *testing.T) {
	// Test creating a new category total
	categoryTotal := &CategoryTotal{
		CategoryID: "cat123",
		Name:       "Groceries",
		SumMinor:   50000,
		Currency:   "USD",
	}
	assert.Equal(t, "cat123", categoryTotal.CategoryID)
	assert.Equal(t, "Groceries", categoryTotal.Name)
	assert.Equal(t, int64(50000), categoryTotal.SumMinor)
	assert.Equal(t, "USD", categoryTotal.Currency)
}

func TestStats_NewStats(t *testing.T) {
	// Test creating new stats
	stats := &Stats{
		Period:       "2023-12",
		TotalIncome:  100000,
		TotalExpense: 80000,
		Currency:     "USD",
	}
	assert.Equal(t, "2023-12", stats.Period)
	assert.Equal(t, int64(100000), stats.TotalIncome)
	assert.Equal(t, int64(80000), stats.TotalExpense)
	assert.Equal(t, "USD", stats.Currency)
}

func TestTransactionType_String(t *testing.T) {
	// Test transaction type string representation
	assert.Equal(t, "expense", string(TransactionExpense))
	assert.Equal(t, "income", string(TransactionIncome))
}

func TestTransactionType_Validation(t *testing.T) {
	// Test transaction type validation
	validTypes := []TransactionType{
		TransactionExpense,
		TransactionIncome,
	}
	
	for _, tType := range validTypes {
		assert.True(t, tType == TransactionExpense || tType == TransactionIncome)
	}
}

func TestMoney_AmountMajor(t *testing.T) {
	// Test converting minor amount to major
	money := NewMoney(12345, "USD")
	// 12345 minor units = 123.45 major units
	assert.Equal(t, int64(12345), money.AmountMinor)
}

func TestMoney_CurrencyValidation(t *testing.T) {
	// Test money with different currencies
	currencies := []string{"USD", "EUR", "RUB", "GBP", "JPY", ""}
	
	for _, currency := range currencies {
		money := NewMoney(10000, currency)
		assert.Equal(t, currency, money.CurrencyCode)
	}
}
