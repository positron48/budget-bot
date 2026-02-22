package ui

import (
	"strings"
	"testing"

	"budget-bot/internal/domain"
	grpcclient "budget-bot/internal/grpc"
)

func TestCreateCategoryKeyboard(t *testing.T) {
	cats := []*domain.Category{{ID: "c1", Name: "Food", Emoji: "🍽️"}, {ID: "c2", Name: "Taxi", Emoji: "🚕"}}
	kb := CreateCategoryKeyboard(cats)
	if len(kb.InlineKeyboard) != 2 {
		t.Fatalf("rows: %d", len(kb.InlineKeyboard))
	}
}

func TestCreateLanguageAndCurrency(t *testing.T) {
	l := CreateLanguageKeyboard()
	if len(l.InlineKeyboard) == 0 {
		t.Fatalf("lang empty")
	}
	cur := CreateCurrencyKeyboard()
	if len(cur.InlineKeyboard) == 0 {
		t.Fatalf("currency empty")
	}
}

func TestCreateTenantKeyboard(t *testing.T) {
	items := []*grpcclient.Tenant{{ID: "t1", Name: "A"}, {ID: "t2", Name: "B"}}
	kb := CreateTenantKeyboard(items)
	if len(kb.InlineKeyboard) != 2 {
		t.Fatalf("rows: %d", len(kb.InlineKeyboard))
	}
}

func TestCreateMainMenuKeyboard(t *testing.T) {
	kb := CreateMainMenuKeyboard()
	if len(kb.Keyboard) == 0 {
		t.Fatalf("main menu empty")
	}
}

func TestCreateHelpKeyboard(t *testing.T) {
	kb := CreateHelpKeyboard("ru")
	if len(kb.InlineKeyboard) != 3 {
		t.Fatalf("Expected 3 rows, got %d", len(kb.InlineKeyboard))
	}

	// Check first row has 2 buttons
	if len(kb.InlineKeyboard[0]) != 2 {
		t.Fatalf("Expected 2 buttons in first row, got %d", len(kb.InlineKeyboard[0]))
	}

	// Check button texts
	expectedButtons := []string{"🔐 Аутентификация", "💰 Транзакции", "🏷️ Категории", "📊 Статистика", "⚙️ Настройки", "👨‍💼 Админ"}
	buttonIndex := 0

	for _, row := range kb.InlineKeyboard {
		for _, button := range row {
			if button.Text != expectedButtons[buttonIndex] {
				t.Fatalf("Expected button text %s, got %s", expectedButtons[buttonIndex], button.Text)
			}
			if button.CallbackData == nil || !strings.HasPrefix(*button.CallbackData, "help:") {
				t.Fatalf("Expected callback data to start with 'help:', got %v", button.CallbackData)
			}
			buttonIndex++
		}
	}
}

func TestCreateHelpKeyboard_CallbackData(t *testing.T) {
	kb := CreateHelpKeyboard("ru")

	expectedCallbacks := []string{
		"help:auth",
		"help:transactions",
		"help:categories",
		"help:stats",
		"help:settings",
		"help:admin",
	}

	callbackIndex := 0
	for _, row := range kb.InlineKeyboard {
		for _, button := range row {
			if button.CallbackData == nil {
				t.Fatalf("Button %s has nil callback data", button.Text)
			}
			if *button.CallbackData != expectedCallbacks[callbackIndex] {
				t.Fatalf("Expected callback data %s, got %s for button %s",
					expectedCallbacks[callbackIndex], *button.CallbackData, button.Text)
			}
			callbackIndex++
		}
	}

	if callbackIndex != len(expectedCallbacks) {
		t.Fatalf("Expected %d callbacks, got %d", len(expectedCallbacks), callbackIndex)
	}
}

func TestCreateBackToHelpKeyboard(t *testing.T) {
	kb := CreateBackToHelpKeyboard("ru")
	if len(kb.InlineKeyboard) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(kb.InlineKeyboard))
	}

	if len(kb.InlineKeyboard[0]) != 1 {
		t.Fatalf("Expected 1 button, got %d", len(kb.InlineKeyboard[0]))
	}

	button := kb.InlineKeyboard[0][0]
	if button.Text != "🔙 Назад к справке" {
		t.Fatalf("Expected button text '🔙 Назад к справке', got %s", button.Text)
	}

	if button.CallbackData == nil || *button.CallbackData != "help:" {
		t.Fatalf("Expected callback data 'help:', got %v", button.CallbackData)
	}
}
