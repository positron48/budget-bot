package ui

import (
	"testing"

	"budget-bot/internal/domain"
	grpcclient "budget-bot/internal/grpc"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/stretchr/testify/assert"
)

func TestCreateCategoryKeyboard_WithCategories(t *testing.T) {
	// Test CreateCategoryKeyboard with categories
	categories := []*domain.Category{
		{
			ID:    "cat1",
			Name:  "Groceries",
			Emoji: "🛒",
		},
		{
			ID:    "cat2",
			Name:  "Transport",
			Emoji: "🚗",
		},
	}

	keyboard := CreateCategoryKeyboard(categories)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)
}

func TestCreateCategoryKeyboard_WithEmptyCategories(t *testing.T) {
	// Test CreateCategoryKeyboard with empty categories
	categories := []*domain.Category{}

	keyboard := CreateCategoryKeyboard(categories)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 0)
}

func TestCreateCategoryKeyboard_WithSingleCategory(t *testing.T) {
	// Test CreateCategoryKeyboard with single category
	categories := []*domain.Category{
		{
			ID:    "cat1",
			Name:  "Groceries",
			Emoji: "🛒",
		},
	}

	keyboard := CreateCategoryKeyboard(categories)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 1)
}

func TestCreateTenantKeyboard_WithTenants(t *testing.T) {
	// Test CreateTenantKeyboard with tenants
	tenants := []*grpcclient.Tenant{
		{
			ID:   "tenant1",
			Name: "Personal",
		},
		{
			ID:   "tenant2",
			Name: "Work",
		},
	}

	keyboard := CreateTenantKeyboard(tenants)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)
}

func TestCreateTenantKeyboard_WithEmptyTenants(t *testing.T) {
	// Test CreateTenantKeyboard with empty tenants
	tenants := []*grpcclient.Tenant{}

	keyboard := CreateTenantKeyboard(tenants)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 0)
}

func TestCreateTenantKeyboard_WithSingleTenant(t *testing.T) {
	// Test CreateTenantKeyboard with single tenant
	tenants := []*grpcclient.Tenant{
		{
			ID:   "tenant1",
			Name: "Personal",
		},
	}

	keyboard := CreateTenantKeyboard(tenants)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 1)
}

func TestCreateLanguageKeyboard_Structure(t *testing.T) {
	// Test CreateLanguageKeyboard structure
	keyboard := CreateLanguageKeyboard()
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 1)
	assert.Len(t, keyboard.InlineKeyboard[0], 2)

	// Check button texts
	assert.Contains(t, keyboard.InlineKeyboard[0][0].Text, "Русский")
	assert.Contains(t, keyboard.InlineKeyboard[0][1].Text, "English")
}

func TestCreateCurrencyKeyboard_Structure(t *testing.T) {
	// Test CreateCurrencyKeyboard structure
	keyboard := CreateCurrencyKeyboard()
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)
	assert.Len(t, keyboard.InlineKeyboard[0], 3)
	assert.Len(t, keyboard.InlineKeyboard[1], 2)

	// Check button texts
	assert.Contains(t, keyboard.InlineKeyboard[0][0].Text, "RUB")
	assert.Contains(t, keyboard.InlineKeyboard[0][1].Text, "USD")
	assert.Contains(t, keyboard.InlineKeyboard[0][2].Text, "EUR")
	assert.Contains(t, keyboard.InlineKeyboard[1][0].Text, "GBP")
	assert.Contains(t, keyboard.InlineKeyboard[1][1].Text, "JPY")
}

func TestCreateMainMenuKeyboard_Structure(t *testing.T) {
	// Test CreateMainMenuKeyboard structure
	keyboard := CreateMainMenuKeyboard()
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.ReplyKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.Keyboard, 2)
	assert.Len(t, keyboard.Keyboard[0], 3)
	assert.Len(t, keyboard.Keyboard[1], 3)

	// Check button texts
	assert.Equal(t, "/stats", keyboard.Keyboard[0][0].Text)
	assert.Equal(t, "/recent", keyboard.Keyboard[0][1].Text)
	assert.Equal(t, "/top_categories", keyboard.Keyboard[0][2].Text)
	assert.Equal(t, "/categories", keyboard.Keyboard[1][0].Text)
	assert.Equal(t, "/profile", keyboard.Keyboard[1][1].Text)
	assert.Equal(t, "/help", keyboard.Keyboard[1][2].Text)
}

func TestCreateHelpKeyboard_Structure(t *testing.T) {
	// Test CreateHelpKeyboard structure
	keyboard := CreateHelpKeyboard("ru")
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 3)
	assert.Len(t, keyboard.InlineKeyboard[0], 2)
	assert.Len(t, keyboard.InlineKeyboard[1], 2)
	assert.Len(t, keyboard.InlineKeyboard[2], 2)

	// Check button texts
	assert.Contains(t, keyboard.InlineKeyboard[0][0].Text, "Аутентификация")
	assert.Contains(t, keyboard.InlineKeyboard[0][1].Text, "Транзакции")
	assert.Contains(t, keyboard.InlineKeyboard[1][0].Text, "Категории")
	assert.Contains(t, keyboard.InlineKeyboard[1][1].Text, "Статистика")
	assert.Contains(t, keyboard.InlineKeyboard[2][0].Text, "Настройки")
	assert.Contains(t, keyboard.InlineKeyboard[2][1].Text, "Админ")
}

func TestCreateBackToHelpKeyboard_Structure(t *testing.T) {
	// Test CreateBackToHelpKeyboard structure
	keyboard := CreateBackToHelpKeyboard("ru")
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 1)
	assert.Len(t, keyboard.InlineKeyboard[0], 1)

	// Check button text
	assert.Contains(t, keyboard.InlineKeyboard[0][0].Text, "Назад к справке")
}

func TestCreateCategoryKeyboard_WithManyCategories(t *testing.T) {
	// Test CreateCategoryKeyboard with many categories
	categories := []*domain.Category{
		{ID: "cat1", Name: "Groceries", Emoji: "🛒"},
		{ID: "cat2", Name: "Transport", Emoji: "🚗"},
		{ID: "cat3", Name: "Entertainment", Emoji: "🎬"},
		{ID: "cat4", Name: "Healthcare", Emoji: "🏥"},
		{ID: "cat5", Name: "Education", Emoji: "📚"},
		{ID: "cat6", Name: "Shopping", Emoji: "🛍️"},
	}

	keyboard := CreateCategoryKeyboard(categories)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 6)
}

func TestCreateCategoryKeyboard_WithSpecialCharacters(t *testing.T) {
	// Test CreateCategoryKeyboard with special characters
	categories := []*domain.Category{
		{
			ID:    "cat_special",
			Name:  "Café & Restaurant",
			Emoji: "🍽️",
		},
		{
			ID:    "cat_unicode",
			Name:  "Тестовая категория",
			Emoji: "🧪",
		},
	}

	keyboard := CreateCategoryKeyboard(categories)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)
}

func TestCreateTenantKeyboard_WithManyTenants(t *testing.T) {
	// Test CreateTenantKeyboard with many tenants
	tenants := []*grpcclient.Tenant{
		{ID: "tenant1", Name: "Personal"},
		{ID: "tenant2", Name: "Work"},
		{ID: "tenant3", Name: "Family"},
		{ID: "tenant4", Name: "Business"},
		{ID: "tenant5", Name: "Side Project"},
	}

	keyboard := CreateTenantKeyboard(tenants)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 5)
}

func TestCreateTenantKeyboard_WithSpecialCharacters(t *testing.T) {
	// Test CreateTenantKeyboard with special characters
	tenants := []*grpcclient.Tenant{
		{
			ID:   "tenant_special",
			Name: "Café & Restaurant Business",
		},
		{
			ID:   "tenant_unicode",
			Name: "Тестовый тенант",
		},
	}

	keyboard := CreateTenantKeyboard(tenants)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)
}

func TestCreateCategoryKeyboard_WithEmptyNames(t *testing.T) {
	// Test CreateCategoryKeyboard with empty names
	categories := []*domain.Category{
		{
			ID:    "cat1",
			Name:  "",
			Emoji: "🛒",
		},
		{
			ID:    "cat2",
			Name:  "Transport",
			Emoji: "",
		},
	}

	keyboard := CreateCategoryKeyboard(categories)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)
}

func TestCreateTenantKeyboard_WithEmptyNames(t *testing.T) {
	// Test CreateTenantKeyboard with empty names
	tenants := []*grpcclient.Tenant{
		{
			ID:   "tenant1",
			Name: "",
		},
		{
			ID:   "tenant2",
			Name: "Work",
		},
	}

	keyboard := CreateTenantKeyboard(tenants)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 2)
}

func TestCreateCategoryKeyboard_WithNilCategories(t *testing.T) {
	// Test CreateCategoryKeyboard with nil categories
	var categories []*domain.Category

	keyboard := CreateCategoryKeyboard(categories)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 0)
}

func TestCreateTenantKeyboard_WithNilTenants(t *testing.T) {
	// Test CreateTenantKeyboard with nil tenants
	var tenants []*grpcclient.Tenant

	keyboard := CreateTenantKeyboard(tenants)
	assert.NotNil(t, keyboard)
	assert.IsType(t, tgbotapi.InlineKeyboardMarkup{}, keyboard)
	assert.Len(t, keyboard.InlineKeyboard, 0)
}
