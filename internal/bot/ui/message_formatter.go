// Package ui contains helpers to format messages and build keyboards.
package ui

import (
	"fmt"

	"budget-bot/internal/domain"
)

// MessageFormatter provides helpers to format bot messages.
type MessageFormatter struct{}

// NewMessageFormatter constructs a MessageFormatter.
func NewMessageFormatter() *MessageFormatter { return &MessageFormatter{} }

// FormatMoney renders money from minor units with currency code, e.g., 1234 RUB.
func (mf *MessageFormatter) FormatMoney(amountMinor int64, currency string) string {
	return fmt.Sprintf("%.2f %s", float64(amountMinor)/100.0, currency)
}

// FormatStats renders a compact stats summary.
func (mf *MessageFormatter) FormatStats(stats *domain.Stats) string {
	return fmt.Sprintf(
		"Статистика %s\nДоход: %s\nРасход: %s",
		stats.Period,
		mf.FormatMoney(stats.TotalIncome, stats.Currency),
		mf.FormatMoney(stats.TotalExpense, stats.Currency),
	)
}

// FormatCategoriesList renders a simple categories list.
func (mf *MessageFormatter) FormatCategoriesList(categories []*domain.Category) string {
	if len(categories) == 0 {
		return "Категорий нет"
	}
	result := "Категории:\n"
	for _, c := range categories {
		result += fmt.Sprintf("%s %s (%s)\n", c.Emoji, c.Name, c.ID)
	}
	return result
}

// FormatTransactionLine renders a one-line transaction entry.
func (mf *MessageFormatter) FormatTransactionLine(sign string, amountMinor int64, currency, comment string) string {
	return fmt.Sprintf("%s %.2f %s %s", sign, float64(amountMinor)/100.0, currency, comment)
}


