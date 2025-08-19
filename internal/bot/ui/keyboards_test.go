package ui

import (
	"testing"

	"budget-bot/internal/domain"
	grpcclient "budget-bot/internal/grpc"
)

func TestCreateCategoryKeyboard(t *testing.T) {
	cats := []*domain.Category{{ID: "c1", Name: "Food", Emoji: "🍽️"}, {ID: "c2", Name: "Taxi", Emoji: "🚕"}}
	kb := CreateCategoryKeyboard(cats)
	if len(kb.InlineKeyboard) != 2 { t.Fatalf("rows: %d", len(kb.InlineKeyboard)) }
}

func TestCreateConfirmationAndLanguageAndCurrency(t *testing.T) {
	c := CreateConfirmationKeyboard()
	if len(c.InlineKeyboard) == 0 { t.Fatalf("confirm empty") }
	l := CreateLanguageKeyboard()
	if len(l.InlineKeyboard) == 0 { t.Fatalf("lang empty") }
	cur := CreateCurrencyKeyboard()
	if len(cur.InlineKeyboard) == 0 { t.Fatalf("currency empty") }
}

func TestCreateTenantKeyboard(t *testing.T) {
	items := []*grpcclient.Tenant{{ID: "t1", Name: "A"}, {ID: "t2", Name: "B"}}
	kb := CreateTenantKeyboard(items)
	if len(kb.InlineKeyboard) != 2 { t.Fatalf("rows: %d", len(kb.InlineKeyboard)) }
}

func TestCreateMainMenuKeyboard(t *testing.T) {
	kb := CreateMainMenuKeyboard()
	if len(kb.Keyboard) == 0 { t.Fatalf("main menu empty") }
}
