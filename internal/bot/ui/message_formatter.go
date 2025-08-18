package ui

import (
	"fmt"

	"budget-bot/internal/domain"
)

type MessageFormatter struct{}

func NewMessageFormatter() *MessageFormatter { return &MessageFormatter{} }

func (mf *MessageFormatter) FormatMoney(amountMinor int64, currency string) string {
	return fmt.Sprintf("%.2f %s", float64(amountMinor)/100.0, currency)
}

func (mf *MessageFormatter) FormatStats(stats *domain.Stats) string {
	return fmt.Sprintf(
		"Статистика %s\nДоход: %s\nРасход: %s",
		stats.Period,
		mf.FormatMoney(stats.TotalIncome, stats.Currency),
		mf.FormatMoney(stats.TotalExpense, stats.Currency),
	)
}

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


