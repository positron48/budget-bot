package ui

import (
	"strings"
	"testing"

	"budget-bot/internal/domain"
)

func TestMessageFormatter_Formatters(t *testing.T) {
	mf := NewMessageFormatter()
	if mf.FormatMoney(12345, "RUB") != "123.45 RUB" { t.Fatalf("money format") }
	st := &domain.Stats{Period: "2025-01", TotalIncome: 10000, TotalExpense: 5000, Currency: "USD"}
	text := mf.FormatStats(st)
	if text == "" || !strings.Contains(text, "Статистика 2025-01") { t.Fatalf("stats format: %s", text) }
	cats := []*domain.Category{{ID: "1", Name: "Food", Emoji: "🍽️"}}
	list := mf.FormatCategoriesList(cats)
	if list == "" { t.Fatalf("categories format empty") }
	line := mf.FormatTransactionLine("-", 100, "USD", "coffee")
	if line == "" { t.Fatalf("line format empty") }
}
