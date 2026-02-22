package ui

import (
	"strings"
	"testing"

	"budget-bot/internal/domain"
)

func TestCreateChangeCategoryKeyboard_CallbackDataLength(t *testing.T) {
	cats := []*domain.Category{{ID: strings.Repeat("a", 36), Name: "Food", Emoji: "🍔"}}
	kb := CreateChangeCategoryKeyboard(cats, strings.Repeat("b", 36))
	got := kb.InlineKeyboard[0][0].CallbackData
	if got == nil {
		t.Fatalf("callback is nil")
	}
	if len(*got) > 64 {
		t.Fatalf("callback length too long: %d", len(*got))
	}
}
