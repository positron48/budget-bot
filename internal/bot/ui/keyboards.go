package ui

import (
    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "budget-bot/internal/domain"
)

func CreateCategoryKeyboard(categories []*domain.Category) tgbotapi.InlineKeyboardMarkup {
    var rows [][]tgbotapi.InlineKeyboardButton
    for _, c := range categories {
        btn := tgbotapi.NewInlineKeyboardButtonData(c.Emoji+" "+c.Name, "cat:"+c.ID)
        rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
    }
    return tgbotapi.NewInlineKeyboardMarkup(rows...)
}


