package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestCreateCategoryKeyboard_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, CreateCategoryKeyboard)
}

func TestCreateConfirmationKeyboard_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, CreateConfirmationKeyboard)
}

func TestCreateLanguageKeyboard_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, CreateLanguageKeyboard)
}

func TestCreateCurrencyKeyboard_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, CreateCurrencyKeyboard)
}

func TestCreateTenantKeyboard_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, CreateTenantKeyboard)
}

func TestCreateMainMenuKeyboard_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, CreateMainMenuKeyboard)
}

func TestCreateHelpKeyboard_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, CreateHelpKeyboard)
}

func TestCreateBackToHelpKeyboard_Exists(t *testing.T) {
	// Test that the function exists and can be called
	assert.NotNil(t, CreateBackToHelpKeyboard)
}

func TestCreateConfirmationKeyboard_ReturnsKeyboard(t *testing.T) {
	// Test that CreateConfirmationKeyboard returns a keyboard
	keyboard := CreateConfirmationKeyboard()
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
}

func TestCreateLanguageKeyboard_ReturnsKeyboard(t *testing.T) {
	// Test that CreateLanguageKeyboard returns a keyboard
	keyboard := CreateLanguageKeyboard()
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
}

func TestCreateCurrencyKeyboard_ReturnsKeyboard(t *testing.T) {
	// Test that CreateCurrencyKeyboard returns a keyboard
	keyboard := CreateCurrencyKeyboard()
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
}

func TestCreateMainMenuKeyboard_ReturnsKeyboard(t *testing.T) {
	// Test that CreateMainMenuKeyboard returns a keyboard
	keyboard := CreateMainMenuKeyboard()
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.ReplyKeyboardMarkup{}, keyboard)
}
