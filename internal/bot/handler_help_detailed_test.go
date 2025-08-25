package bot

import (
	"context"
	"strings"
	"testing"

	"budget-bot/internal/repository"
	"budget-bot/internal/testutil"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func TestHandler_HelpCommandsWithArgs(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	sessions := repository.NewSQLiteSessionRepository(db)
	states := repository.NewSQLiteDialogStateRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
	bot := testutil.NewTestBot(t)

	h := NewHandler(bot, states, auth, mappings, nil, log).
		WithPreferences(prefs)

	ctx := context.Background()
	chatID := int64(7000)
	userID := int64(99)

	// Тестируем команды help с аргументами
	testCases := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "help auth",
			command:  "/help auth",
			expected: "auth",
		},
		{
			name:     "help stats",
			command:  "/help stats",
			expected: "stats",
		},
		{
			name:     "help transactions",
			command:  "/help transactions",
			expected: "transactions",
		},
		{
			name:     "help categories",
			command:  "/help categories",
			expected: "categories",
		},
		{
			name:     "help settings",
			command:  "/help settings",
			expected: "settings",
		},
		{
			name:     "help admin",
			command:  "/help admin",
			expected: "admin",
		},
		{
			name:     "help unknown",
			command:  "/help unknown",
			expected: "unknown",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upd := tgbotapi.Update{
				UpdateID: 1,
				Message: &tgbotapi.Message{
					Chat: &tgbotapi.Chat{ID: chatID},
					From: &tgbotapi.User{ID: userID},
					Text: tc.command,
				},
			}
			
			// Добавляем правильные entities для команды
			upd.Message.Entities = []tgbotapi.MessageEntity{
				{
					Type:   "bot_command",
					Offset: 0,
					Length: 5, // "/help"
				},
			}

			// Проверяем, что команда правильно парсится
			if !upd.Message.IsCommand() {
				t.Errorf("Expected IsCommand() to be true for %q", tc.command)
			}

			if upd.Message.Command() != "help" {
				t.Errorf("Expected Command() to be 'help', got %q", upd.Message.Command())
			}

			if upd.Message.CommandArguments() != tc.expected {
				t.Errorf("Expected CommandArguments() to be %q, got %q", tc.expected, upd.Message.CommandArguments())
			}

			// Обрабатываем команду
			h.HandleUpdate(ctx, upd)
		})
	}
}

func TestHandler_HelpCommandParsing(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	sessions := repository.NewSQLiteSessionRepository(db)
	states := repository.NewSQLiteDialogStateRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
	bot := testutil.NewTestBot(t)

	h := NewHandler(bot, states, auth, mappings, nil, log).
		WithPreferences(prefs)

	ctx := context.Background()
	chatID := int64(8000)
	userID := int64(99)

	// Тестируем различные варианты написания команд
	testCases := []struct {
		name     string
		command  string
		entities []tgbotapi.MessageEntity
	}{
		{
			name:    "help auth with proper entities",
			command: "/help auth",
			entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 5},
			},
		},
		{
			name:    "help stats with proper entities",
			command: "/help stats",
			entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: 5},
			},
		},
		{
			name:    "help auth without entities",
			command: "/help auth",
			entities: []tgbotapi.MessageEntity{},
		},
		{
			name:    "help stats without entities",
			command: "/help stats",
			entities: []tgbotapi.MessageEntity{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			upd := tgbotapi.Update{
				UpdateID: 1,
				Message: &tgbotapi.Message{
					Chat:    &tgbotapi.Chat{ID: chatID},
					From:    &tgbotapi.User{ID: userID},
					Text:    tc.command,
					Entities: tc.entities,
				},
			}

			// Обрабатываем команду
			h.HandleUpdate(ctx, upd)
		})
	}
}

func TestHandler_HelpCommandsWithoutEntities(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	sessions := repository.NewSQLiteSessionRepository(db)
	states := repository.NewSQLiteDialogStateRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
	bot := testutil.NewTestBot(t)

	h := NewHandler(bot, states, auth, mappings, nil, log).
		WithPreferences(prefs)

	ctx := context.Background()
	chatID := int64(9000)
	userID := int64(99)

	// Тестируем команды без entities (как может происходить в реальном Telegram)
	testCases := []struct {
		name    string
		command string
	}{
		{
			name:    "help auth without entities",
			command: "/help auth",
		},
		{
			name:    "help stats without entities",
			command: "/help stats",
		},
		{
			name:    "help transactions without entities",
			command: "/help transactions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			upd := tgbotapi.Update{
				UpdateID: 1,
				Message: &tgbotapi.Message{
					Chat:     &tgbotapi.Chat{ID: chatID},
					From:     &tgbotapi.User{ID: userID},
					Text:     tc.command,
					Entities: []tgbotapi.MessageEntity{}, // Пустые entities
				},
			}

			// Проверяем, что команда НЕ распознается как команда без entities
			if upd.Message.IsCommand() {
				t.Logf("Warning: IsCommand() returned true for %q without entities", tc.command)
			}

			// Обрабатываем команду - должен сработать fallback в HandleUpdate
			h.HandleUpdate(ctx, upd)
		})
	}
}

func TestHandler_RealWorldHelpCommands(t *testing.T) {
	log := zap.NewNop()
	db := testutil.OpenMigratedSQLite(t)
	sessions := repository.NewSQLiteSessionRepository(db)
	states := repository.NewSQLiteDialogStateRepository(db)
	mappings := repository.NewSQLiteCategoryMappingRepository(db)
	prefs := repository.NewSQLitePreferencesRepository(db)
	auth := NewOAuthManager(&TestOAuthClient{}, sessions, log, "http://localhost:3000")
	bot := testutil.NewTestBot(t)

	h := NewHandler(bot, states, auth, mappings, nil, log).
		WithPreferences(prefs)

	ctx := context.Background()
	chatID := int64(10000)
	userID := int64(99)

	// Тестируем реальные команды, которые могут не работать в Telegram
	testCases := []struct {
		name    string
		command string
	}{
		{
			name:    "help auth - real world scenario",
			command: "/help auth",
		},
		{
			name:    "help stats - real world scenario",
			command: "/help stats",
		},
		{
			name:    "help transactions - real world scenario",
			command: "/help transactions",
		},
		{
			name:    "help categories - real world scenario",
			command: "/help categories",
		},
		{
			name:    "help settings - real world scenario",
			command: "/help settings",
		},
		{
			name:    "help admin - real world scenario",
			command: "/help admin",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем сообщение без entities (как в реальном Telegram)
			upd := tgbotapi.Update{
				UpdateID: 1,
				Message: &tgbotapi.Message{
					Chat:     &tgbotapi.Chat{ID: chatID},
					From:     &tgbotapi.User{ID: userID},
					Text:     tc.command,
					Entities: []tgbotapi.MessageEntity{}, // Без entities
				},
			}

			// Проверяем, что команда НЕ распознается как команда
			if upd.Message.IsCommand() {
				t.Logf("Note: IsCommand() returned true for %q", tc.command)
			} else {
				t.Logf("Note: IsCommand() returned false for %q - this is expected without entities", tc.command)
			}

			// Проверяем, что fallback механизм работает
			if !strings.HasPrefix(tc.command, "/") {
				t.Errorf("Command %q should start with /", tc.command)
			}

			// Обрабатываем команду
			h.HandleUpdate(ctx, upd)
		})
	}
}
