// Package ui contains helpers to build Telegram reply/inline keyboards.
package ui

import (
	"budget-bot/internal/domain"
	grpcclient "budget-bot/internal/grpc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// CreateCategoryKeyboard builds an inline keyboard for categories.
func CreateCategoryKeyboard(categories []*domain.Category) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range categories {
		btn := tgbotapi.NewInlineKeyboardButtonData(c.Emoji+" "+c.Name, "v1:cat_select:"+c.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// CreatePostSelectionKeyboard builds actions after category is selected.
func CreatePostSelectionKeyboard(source, opID string) tgbotapi.InlineKeyboardMarkup {
	change := tgbotapi.NewInlineKeyboardButtonData("Сменить категорию", "v1:change:"+opID)
	if source == "mapping" {
		forget := tgbotapi.NewInlineKeyboardButtonData("Забыть выбор", "v1:forget:"+opID)
		return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(forget, change))
	}
	remember := tgbotapi.NewInlineKeyboardButtonData("Запомнить выбор", "v1:remember:"+opID)
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(remember, change))
}

// CreateChangeCategoryKeyboard builds category keyboard bound to operation id.
func CreateChangeCategoryKeyboard(categories []*domain.Category, opID string) tgbotapi.InlineKeyboardMarkup {
	_ = opID
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range categories {
		btn := tgbotapi.NewInlineKeyboardButtonData(c.Emoji+" "+c.Name, "v1:cat_select:"+c.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// CreateLanguageKeyboard builds language selection keyboard.
func CreateLanguageKeyboard() tgbotapi.InlineKeyboardMarkup {
	ru := tgbotapi.NewInlineKeyboardButtonData("🇷🇺 Русский", "lang:ru")
	en := tgbotapi.NewInlineKeyboardButtonData("🇺🇸 English", "lang:en")
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(ru, en))
}

// CreateCurrencyKeyboard builds default currency selection keyboard.
func CreateCurrencyKeyboard() tgbotapi.InlineKeyboardMarkup {
	rub := tgbotapi.NewInlineKeyboardButtonData("₽ RUB", "cur:RUB")
	usd := tgbotapi.NewInlineKeyboardButtonData("$ USD", "cur:USD")
	eur := tgbotapi.NewInlineKeyboardButtonData("€ EUR", "cur:EUR")
	gbp := tgbotapi.NewInlineKeyboardButtonData("£ GBP", "cur:GBP")
	jpy := tgbotapi.NewInlineKeyboardButtonData("¥ JPY", "cur:JPY")
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(rub, usd, eur),
		tgbotapi.NewInlineKeyboardRow(gbp, jpy),
	)
}

// CreateTenantKeyboard builds a tenant selection keyboard.
func CreateTenantKeyboard(items []*grpcclient.Tenant) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range items {
		btn := tgbotapi.NewInlineKeyboardButtonData(t.Name, "tenant:"+t.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// CreateMainMenuKeyboard builds the main menu keyboard.
func CreateMainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	row1 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/stats"),
		tgbotapi.NewKeyboardButton("/recent"),
		tgbotapi.NewKeyboardButton("/top_categories"),
	)
	row2 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/categories"),
		tgbotapi.NewKeyboardButton("/profile"),
		tgbotapi.NewKeyboardButton("/help"),
	)
	kb := tgbotapi.NewReplyKeyboard(row1, row2)
	kb.ResizeKeyboard = true
	kb.Selective = true
	return kb
}

// CreateHelpKeyboard builds the main help menu keyboard.
func CreateHelpKeyboard() tgbotapi.InlineKeyboardMarkup {
	auth := tgbotapi.NewInlineKeyboardButtonData("🔐 Аутентификация", "help:auth")
	transactions := tgbotapi.NewInlineKeyboardButtonData("💰 Транзакции", "help:transactions")
	categories := tgbotapi.NewInlineKeyboardButtonData("🏷️ Категории", "help:categories")
	stats := tgbotapi.NewInlineKeyboardButtonData("📊 Статистика", "help:stats")
	settings := tgbotapi.NewInlineKeyboardButtonData("⚙️ Настройки", "help:settings")
	admin := tgbotapi.NewInlineKeyboardButtonData("👨‍💼 Админ", "help:admin")

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(auth, transactions),
		tgbotapi.NewInlineKeyboardRow(categories, stats),
		tgbotapi.NewInlineKeyboardRow(settings, admin),
	)
}

// CreateBackToHelpKeyboard builds a keyboard with back button.
func CreateBackToHelpKeyboard() tgbotapi.InlineKeyboardMarkup {
	back := tgbotapi.NewInlineKeyboardButtonData("🔙 Назад к справке", "help:")

	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(back),
	)
}
