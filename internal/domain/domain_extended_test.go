package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMoney_Extended(t *testing.T) {
	// Test NewMoney with different values
	money1 := NewMoney(10000, "USD")
	money2 := NewMoney(5000, "EUR")
	money3 := NewMoney(0, "RUB")
	
	assert.Equal(t, int64(10000), money1.AmountMinor)
	assert.Equal(t, "USD", money1.CurrencyCode)
	
	assert.Equal(t, int64(5000), money2.AmountMinor)
	assert.Equal(t, "EUR", money2.CurrencyCode)
	
	assert.Equal(t, int64(0), money3.AmountMinor)
	assert.Equal(t, "RUB", money3.CurrencyCode)
}

func TestNewMoney_WithLargeAmount(t *testing.T) {
	// Test NewMoney with large amount
	largeAmount := int64(999999999)
	money := NewMoney(largeAmount, "USD")
	
	assert.Equal(t, largeAmount, money.AmountMinor)
	assert.Equal(t, "USD", money.CurrencyCode)
}

func TestNewMoney_WithNegativeAmount(t *testing.T) {
	// Test NewMoney with negative amount
	negativeAmount := int64(-1000)
	money := NewMoney(negativeAmount, "USD")
	
	assert.Equal(t, negativeAmount, money.AmountMinor)
	assert.Equal(t, "USD", money.CurrencyCode)
}

func TestCategory_Extended(t *testing.T) {
	// Test Category with different values
	category1 := &Category{
		ID:    "cat1",
		Name:  "Groceries",
		Emoji: "🛒",
	}
	
	category2 := &Category{
		ID:    "cat2",
		Name:  "Transport",
		Emoji: "🚗",
	}
	
	assert.Equal(t, "cat1", category1.ID)
	assert.Equal(t, "Groceries", category1.Name)
	assert.Equal(t, "🛒", category1.Emoji)
	
	assert.Equal(t, "cat2", category2.ID)
	assert.Equal(t, "Transport", category2.Name)
	assert.Equal(t, "🚗", category2.Emoji)
}

func TestCategory_WithSpecialCharacters(t *testing.T) {
	// Test Category with special characters in name
	category := &Category{
		ID:    "cat_special",
		Name:  "Café & Restaurant",
		Emoji: "🍽️",
	}
	
	assert.Equal(t, "cat_special", category.ID)
	assert.Equal(t, "Café & Restaurant", category.Name)
	assert.Equal(t, "🍽️", category.Emoji)
}

func TestCategoryTotal_Extended(t *testing.T) {
	// Test CategoryTotal with different values
	total1 := &CategoryTotal{
		CategoryID: "cat1",
		Name:       "Groceries",
		SumMinor:   10000,
		Currency:   "USD",
	}
	
	total2 := &CategoryTotal{
		CategoryID: "cat2",
		Name:       "Transport",
		SumMinor:   5000,
		Currency:   "USD",
	}
	
	assert.Equal(t, "cat1", total1.CategoryID)
	assert.Equal(t, "Groceries", total1.Name)
	assert.Equal(t, int64(10000), total1.SumMinor)
	assert.Equal(t, "USD", total1.Currency)
	
	assert.Equal(t, "cat2", total2.CategoryID)
	assert.Equal(t, "Transport", total2.Name)
	assert.Equal(t, int64(5000), total2.SumMinor)
	assert.Equal(t, "USD", total2.Currency)
}

func TestStats_Extended(t *testing.T) {
	// Test Stats with different values
	stats1 := &Stats{
		Period:       "2023-12",
		TotalIncome:  20000,
		TotalExpense: 15000,
		Currency:     "USD",
	}
	
	stats2 := &Stats{
		Period:       "2023-11",
		TotalIncome:  0,
		TotalExpense: 0,
		Currency:     "EUR",
	}
	
	assert.Equal(t, "2023-12", stats1.Period)
	assert.Equal(t, int64(20000), stats1.TotalIncome)
	assert.Equal(t, int64(15000), stats1.TotalExpense)
	assert.Equal(t, "USD", stats1.Currency)
	
	assert.Equal(t, "2023-11", stats2.Period)
	assert.Equal(t, int64(0), stats2.TotalIncome)
	assert.Equal(t, int64(0), stats2.TotalExpense)
	assert.Equal(t, "EUR", stats2.Currency)
}

func TestStats_WithDifferentCurrencies(t *testing.T) {
	// Test Stats with different currencies
	statsUSD := &Stats{
		Period:       "2023-12",
		TotalIncome:  100000,
		TotalExpense: 75000,
		Currency:     "USD",
	}
	
	statsEUR := &Stats{
		Period:       "2023-12",
		TotalIncome:  85000,
		TotalExpense: 60000,
		Currency:     "EUR",
	}
	
	assert.Equal(t, "USD", statsUSD.Currency)
	assert.Equal(t, "EUR", statsEUR.Currency)
	assert.NotEqual(t, statsUSD.TotalIncome, statsEUR.TotalIncome)
}
