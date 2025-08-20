// Package ui contains helpers to build Telegram reply/inline keyboards.
package ui

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "budget-bot/internal/domain"
    grpcclient "budget-bot/internal/grpc"
)

// CreateCategoryKeyboard builds an inline keyboard for categories.
func CreateCategoryKeyboard(categories []*domain.Category) tgbotapi.InlineKeyboardMarkup {
    var rows [][]tgbotapi.InlineKeyboardButton
    for _, c := range categories {
        btn := tgbotapi.NewInlineKeyboardButtonData(c.Emoji+" "+c.Name, "cat:"+c.ID)
        rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
    }
    return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// CreateConfirmationKeyboard builds Yes/No inline keyboard.
func CreateConfirmationKeyboard() tgbotapi.InlineKeyboardMarkup {
    yes := tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить", "confirm:yes")
    no := tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "confirm:no")
    return tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(yes, no),
    )
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
        tgbotapi.NewKeyboardButton("/switch_tenant"),
    )
    kb := tgbotapi.NewReplyKeyboard(row1, row2)
    kb.ResizeKeyboard = true
    kb.Selective = true
    return kb
}


